package chaining

import (
	"context"
	"fmt"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/chaining"
	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
	"github.com/sirupsen/logrus"
)

// ChainExecutor implements the ChainService interface
type ChainExecutor struct {
	stegoService stego.StegoService
	mediaService media.Service
}

// NewChainExecutor creates a new chain executor
func NewChainExecutor(stegoService stego.StegoService, mediaService media.Service) *ChainExecutor {
	return &ChainExecutor{
		stegoService: stegoService,
		mediaService: mediaService,
	}
}

// ExecuteChain executes a technique chain on the payload
func (e *ChainExecutor) ExecuteChain(ctx context.Context, chain *chaining.Chain, payload []byte, carriers [][]byte) (*chaining.ChainExecutionResult, error) {
	logger.Log.WithFields(logrus.Fields{
		"chain_id": chain.ID,
		"mode":     chain.Mode,
		"links":    len(chain.Links),
	}).Info("executing technique chain")

	startTime := time.Now()

	// Validate chain
	if err := chain.Validate(); err != nil {
		return nil, fmt.Errorf("chain validation failed: %w", err)
	}

	var result *chaining.ChainExecutionResult
	var err error

	switch chain.Mode {
	case chaining.ChainModeSequential:
		result, err = e.executeSequential(ctx, chain, payload, carriers)
	case chaining.ChainModeLayered:
		result, err = e.executeLayered(ctx, chain, payload, carriers)
	case chaining.ChainModeSplit:
		result, err = e.executeSplit(ctx, chain, payload, carriers)
	default:
		return nil, fmt.Errorf("unsupported chain mode: %s", chain.Mode)
	}

	if err != nil {
		logger.Log.WithError(err).Error("chain execution failed")
		return nil, err
	}

	result.TotalDuration = time.Since(startTime)
	logger.Log.WithFields(logrus.Fields{
		"chain_id": chain.ID,
		"duration": result.TotalDuration,
		"success":  result.Success,
	}).Info("chain execution completed")

	return result, nil
}

// executeSequential applies techniques one after another
func (e *ChainExecutor) executeSequential(ctx context.Context, chain *chaining.Chain, payload []byte, carriers [][]byte) (*chaining.ChainExecutionResult, error) {
	result := &chaining.ChainExecutionResult{
		ChainID:     chain.ID,
		Mode:        chain.Mode,
		StepResults: make([]chaining.StepResult, 0, len(chain.Links)),
		Success:     true,
	}

	currentData := payload
	currentCarrier := carriers[0]

	// Apply each technique in sequence
	for i, link := range chain.Links {
		stepStart := time.Now()
		stepResult := chaining.StepResult{
			LinkOrder: i,
			Technique: link.Technique,
			InputSize: int64(len(currentData)),
		}

		logger.Log.WithFields(logrus.Fields{
			"step":      i,
			"technique": link.Technique,
			"data_size": len(currentData),
		}).Debug("executing chain step")

		// Embed data using current technique
		container, err := e.stegoService.Embed(ctx, currentCarrier, currentData, link.Technique)
		if err != nil {
			stepResult.Success = false
			stepResult.ErrorMessage = err.Error()
			result.StepResults = append(result.StepResults, stepResult)
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("step %d failed: %v", i, err)
			return result, err
		}

		// Extract the stego media from the container
		embedResult := container.CoverMedia

		stepResult.OutputSize = int64(len(embedResult))
		stepResult.Duration = time.Since(stepStart)
		stepResult.Success = true
		result.StepResults = append(result.StepResults, stepResult)

		// Output becomes input for next technique (if not last)
		if i < len(chain.Links)-1 {
			currentData = embedResult
			if i+1 < len(carriers) {
				currentCarrier = carriers[i+1]
			}
		} else {
			result.FinalOutput = embedResult
		}
	}

	return result, nil
}

// executeLayered embeds different data portions using different techniques
func (e *ChainExecutor) executeLayered(ctx context.Context, chain *chaining.Chain, payload []byte, carriers [][]byte) (*chaining.ChainExecutionResult, error) {
	result := &chaining.ChainExecutionResult{
		ChainID:     chain.ID,
		Mode:        chain.Mode,
		StepResults: make([]chaining.StepResult, 0, len(chain.Links)),
		Success:     true,
	}

	if len(carriers) == 0 {
		return nil, fmt.Errorf("layered mode requires at least one carrier")
	}

	carrier := carriers[0]
	offset := 0

	// Embed each portion with its technique
	for i, link := range chain.Links {
		stepStart := time.Now()

		// Calculate data portion based on weight
		portionSize := int(float64(len(payload)) * link.Weight)
		if offset+portionSize > len(payload) {
			portionSize = len(payload) - offset
		}

		portion := payload[offset : offset+portionSize]

		stepResult := chaining.StepResult{
			LinkOrder: i,
			Technique: link.Technique,
			InputSize: int64(len(portion)),
		}

		logger.Log.WithFields(logrus.Fields{
			"step":      i,
			"technique": link.Technique,
			"weight":    link.Weight,
			"portion":   len(portion),
		}).Debug("executing layered chain step")

		// Embed portion
		container, err := e.stegoService.Embed(ctx, carrier, portion, link.Technique)
		if err != nil {
			stepResult.Success = false
			stepResult.ErrorMessage = err.Error()
			result.StepResults = append(result.StepResults, stepResult)
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("step %d failed: %v", i, err)
			return result, err
		}

		// Extract the stego media from the container
		embedResult := container.CoverMedia

		stepResult.OutputSize = int64(len(embedResult))
		stepResult.Duration = time.Since(stepStart)
		stepResult.Success = true
		result.StepResults = append(result.StepResults, stepResult)

		// Update carrier with embedded data
		carrier = embedResult
		offset += portionSize
	}

	result.FinalOutput = carrier
	return result, nil
}

// executeSplit distributes data across multiple carriers
func (e *ChainExecutor) executeSplit(ctx context.Context, chain *chaining.Chain, payload []byte, carriers [][]byte) (*chaining.ChainExecutionResult, error) {
	result := &chaining.ChainExecutionResult{
		ChainID:     chain.ID,
		Mode:        chain.Mode,
		StepResults: make([]chaining.StepResult, 0, len(chain.Links)),
		Success:     true,
	}

	if len(carriers) < len(chain.Links) {
		return nil, fmt.Errorf("split mode requires at least %d carriers, got %d", len(chain.Links), len(carriers))
	}

	// Calculate shard size
	shardSize := len(payload) / len(chain.Links)
	offset := 0

	// Embed each shard in a different carrier
	var combinedOutput []byte

	for i, link := range chain.Links {
		stepStart := time.Now()

		// Calculate shard size (last shard gets remainder)
		currentShardSize := shardSize
		if i == len(chain.Links)-1 {
			currentShardSize = len(payload) - offset
		}

		shard := payload[offset : offset+currentShardSize]

		stepResult := chaining.StepResult{
			LinkOrder: i,
			Technique: link.Technique,
			InputSize: int64(len(shard)),
		}

		logger.Log.WithFields(logrus.Fields{
			"step":      i,
			"technique": link.Technique,
			"shard":     len(shard),
			"carrier":   i,
		}).Debug("executing split chain step")

		// Embed shard in carrier
		container, err := e.stegoService.Embed(ctx, carriers[i], shard, link.Technique)
		if err != nil {
			stepResult.Success = false
			stepResult.ErrorMessage = err.Error()
			result.StepResults = append(result.StepResults, stepResult)
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("step %d failed: %v", i, err)
			return result, err
		}

		// Extract the stego media from the container
		embedResult := container.CoverMedia

		stepResult.OutputSize = int64(len(embedResult))
		stepResult.Duration = time.Since(stepStart)
		stepResult.Success = true
		stepResult.IntermediateKey = fmt.Sprintf("shard_%d", i)
		result.StepResults = append(result.StepResults, stepResult)

		combinedOutput = append(combinedOutput, embedResult...)
		offset += currentShardSize
	}

	result.FinalOutput = combinedOutput
	return result, nil
}

// ReverseChain extracts data by reversing the chain
func (e *ChainExecutor) ReverseChain(ctx context.Context, chain *chaining.Chain, stegoMedia [][]byte) ([]byte, error) {
	logger.Log.WithFields(logrus.Fields{
		"chain_id": chain.ID,
		"mode":     chain.Mode,
	}).Info("reversing technique chain")

	switch chain.Mode {
	case chaining.ChainModeSequential:
		return e.reverseSequential(ctx, chain, stegoMedia)
	case chaining.ChainModeLayered:
		return e.reverseLayered(ctx, chain, stegoMedia)
	case chaining.ChainModeSplit:
		return e.reverseSplit(ctx, chain, stegoMedia)
	default:
		return nil, fmt.Errorf("unsupported chain mode: %s", chain.Mode)
	}
}

// reverseSequential extracts data by applying techniques in reverse
func (e *ChainExecutor) reverseSequential(ctx context.Context, chain *chaining.Chain, stegoMedia [][]byte) ([]byte, error) {
	currentData := stegoMedia[len(stegoMedia)-1]

	// Apply techniques in reverse order
	for i := len(chain.Links) - 1; i >= 0; i-- {
		link := chain.Links[i]

		logger.Log.WithFields(logrus.Fields{
			"step":      i,
			"technique": link.Technique,
		}).Debug("extracting chain step")

		extracted, err := e.stegoService.Extract(ctx, currentData, link.Technique)
		if err != nil {
			return nil, fmt.Errorf("extraction failed at step %d: %w", i, err)
		}

		currentData = extracted
	}

	return currentData, nil
}

// reverseLayered extracts layered data
func (e *ChainExecutor) reverseLayered(ctx context.Context, chain *chaining.Chain, stegoMedia [][]byte) ([]byte, error) {
	if len(stegoMedia) == 0 {
		return nil, fmt.Errorf("no stego media provided")
	}

	carrier := stegoMedia[0]
	var reconstructed []byte

	// Extract each portion
	for i, link := range chain.Links {
		logger.Log.WithFields(logrus.Fields{
			"step":      i,
			"technique": link.Technique,
			"weight":    link.Weight,
		}).Debug("extracting layered chain step")

		extracted, err := e.stegoService.Extract(ctx, carrier, link.Technique)
		if err != nil {
			return nil, fmt.Errorf("extraction failed at step %d: %w", i, err)
		}

		reconstructed = append(reconstructed, extracted...)
	}

	return reconstructed, nil
}

// reverseSplit reconstructs data from multiple carriers
func (e *ChainExecutor) reverseSplit(ctx context.Context, chain *chaining.Chain, stegoMedia [][]byte) ([]byte, error) {
	if len(stegoMedia) < len(chain.Links) {
		return nil, fmt.Errorf("insufficient stego media: need %d, got %d", len(chain.Links), len(stegoMedia))
	}

	var reconstructed []byte

	// Extract each shard
	for i, link := range chain.Links {
		logger.Log.WithFields(logrus.Fields{
			"step":      i,
			"technique": link.Technique,
		}).Debug("extracting split chain shard")

		extracted, err := e.stegoService.Extract(ctx, stegoMedia[i], link.Technique)
		if err != nil {
			return nil, fmt.Errorf("extraction failed at shard %d: %w", i, err)
		}

		reconstructed = append(reconstructed, extracted...)
	}

	return reconstructed, nil
}

// EstimateChainCapacity calculates total capacity for a chain
func (e *ChainExecutor) EstimateChainCapacity(ctx context.Context, chain *chaining.Chain, carriers [][]byte) (int64, error) {
	// Capacity depends on mode and weakest link
	switch chain.Mode {
	case chaining.ChainModeSequential:
		// Capacity is limited by the smallest capacity in the chain
		return e.estimateSequentialCapacity(ctx, chain, carriers)
	case chaining.ChainModeLayered:
		// Capacity is sum of all technique capacities (weighted)
		return e.estimateLayeredCapacity(ctx, chain, carriers)
	case chaining.ChainModeSplit:
		// Capacity is sum of all carrier capacities
		return e.estimateSplitCapacity(ctx, chain, carriers)
	default:
		return 0, fmt.Errorf("unsupported chain mode: %s", chain.Mode)
	}
}

func (e *ChainExecutor) estimateSequentialCapacity(ctx context.Context, chain *chaining.Chain, carriers [][]byte) (int64, error) {
	// For sequential, capacity is limited by the smallest step
	minCapacity := int64(^uint(0) >> 1) // Max int64

	for i, link := range chain.Links {
		var carrier []byte
		if i < len(carriers) {
			carrier = carriers[i]
		}

		capacity, err := e.stegoService.CalculateCapacity(ctx, carrier, link.Technique)
		if err != nil {
			return 0, fmt.Errorf("capacity calculation failed for step %d: %w", i, err)
		}

		if capacity < minCapacity {
			minCapacity = capacity
		}
	}

	return minCapacity, nil
}

func (e *ChainExecutor) estimateLayeredCapacity(ctx context.Context, chain *chaining.Chain, carriers [][]byte) (int64, error) {
	if len(carriers) == 0 {
		return 0, fmt.Errorf("no carriers provided")
	}

	carrier := carriers[0]
	totalCapacity := int64(0)

	for _, link := range chain.Links {
		capacity, err := e.stegoService.CalculateCapacity(ctx, carrier, link.Technique)
		if err != nil {
			return 0, fmt.Errorf("capacity calculation failed: %w", err)
		}

		// Weight the capacity
		weightedCapacity := int64(float64(capacity) * link.Weight)
		totalCapacity += weightedCapacity
	}

	return totalCapacity, nil
}

func (e *ChainExecutor) estimateSplitCapacity(ctx context.Context, chain *chaining.Chain, carriers [][]byte) (int64, error) {
	if len(carriers) < len(chain.Links) {
		return 0, fmt.Errorf("insufficient carriers")
	}

	totalCapacity := int64(0)

	for i, link := range chain.Links {
		capacity, err := e.stegoService.CalculateCapacity(ctx, carriers[i], link.Technique)
		if err != nil {
			return 0, fmt.Errorf("capacity calculation failed for carrier %d: %w", i, err)
		}

		totalCapacity += capacity
	}

	return totalCapacity, nil
}

// ValidateChain checks if a chain configuration is valid
func (e *ChainExecutor) ValidateChain(ctx context.Context, chain *chaining.Chain) error {
	return chain.Validate()
}

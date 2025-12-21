package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/greysquirr3l/shadowforge/internal/domain/chaining"
	chaininfra "github.com/greysquirr3l/shadowforge/internal/infrastructure/chaining"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
	"github.com/sirupsen/logrus"
)

// CreateChainHandler handles chain creation commands
type CreateChainHandler struct {
	executor *chaininfra.ChainExecutor
}

// NewCreateChainHandler creates a new handler
func NewCreateChainHandler(executor *chaininfra.ChainExecutor) *CreateChainHandler {
	return &CreateChainHandler{
		executor: executor,
	}
}

// Handle executes the create chain command
func (h *CreateChainHandler) Handle(ctx context.Context, cmd CreateChainCommand) (*CreateChainResult, error) {
	logger.Log.WithFields(logrus.Fields{
		"mode":  cmd.Mode,
		"links": len(cmd.Links),
	}).Info("creating technique chain")

	// Convert DTOs to domain objects
	links := make([]chaining.ChainLink, len(cmd.Links))
	for i, linkDTO := range cmd.Links {
		links[i] = chaining.ChainLink{
			TechniqueID:   linkDTO.TechniqueID,
			Technique:     chaining.StegoTechnique(linkDTO.Technique),
			Configuration: linkDTO.Configuration,
			Order:         linkDTO.Order,
			Weight:        linkDTO.Weight,
		}
	}

	// Create chain
	chain := &chaining.Chain{
		ID:          uuid.New().String(),
		Mode:        cmd.Mode,
		Links:       links,
		CreatedAt:   time.Now(),
		Description: cmd.Description,
	}

	// Validate
	if err := h.executor.ValidateChain(ctx, chain); err != nil {
		return &CreateChainResult{
			Success: false,
			Error:   fmt.Sprintf("validation failed: %v", err),
		}, err
	}

	logger.Log.WithField("chain_id", chain.ID).Info("chain created successfully")

	return &CreateChainResult{
		ChainID:   chain.ID,
		Success:   true,
		CreatedAt: chain.CreatedAt,
	}, nil
}

// ExecuteChainHandler handles chain execution commands
type ExecuteChainHandler struct {
	executor *chaininfra.ChainExecutor
}

// NewExecuteChainHandler creates a new handler
func NewExecuteChainHandler(executor *chaininfra.ChainExecutor) *ExecuteChainHandler {
	return &ExecuteChainHandler{
		executor: executor,
	}
}

// Handle executes the chain execution command
func (h *ExecuteChainHandler) Handle(ctx context.Context, cmd ExecuteChainCommand) (*ExecuteChainResult, error) {
	logger.Log.WithFields(logrus.Fields{
		"chain_id": cmd.ChainID,
		"carriers": len(cmd.CarrierPaths),
	}).Info("executing technique chain")

	// Load payload
	payload, err := os.ReadFile(cmd.PayloadPath)
	if err != nil {
		return &ExecuteChainResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read payload: %v", err),
		}, err
	}

	// Load carriers
	carriers := make([][]byte, len(cmd.CarrierPaths))
	for i, path := range cmd.CarrierPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return &ExecuteChainResult{
				Success: false,
				Error:   fmt.Sprintf("failed to read carrier %d: %v", i, err),
			}, err
		}
		carriers[i] = data
	}

	// TODO: Load actual chain configuration
	// For now, create a sample chain
	chain := &chaining.Chain{
		ID:   cmd.ChainID,
		Mode: chaining.ChainModeSequential,
		Links: []chaining.ChainLink{
			{Order: 0, Technique: "lsb"},
			{Order: 1, Technique: "dct"},
		},
	}

	// Execute chain
	result, err := h.executor.ExecuteChain(ctx, chain, payload, carriers)
	if err != nil {
		return &ExecuteChainResult{
			ChainID: cmd.ChainID,
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Save output
	if err := os.WriteFile(cmd.OutputPath, result.FinalOutput, 0644); err != nil {
		return &ExecuteChainResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write output: %v", err),
		}, err
	}

	// Build summary
	steps := make([]StepSummary, len(result.StepResults))
	for i, step := range result.StepResults {
		steps[i] = StepSummary{
			Order:     step.LinkOrder,
			Technique: string(step.Technique),
			Duration:  step.Duration,
			Success:   step.Success,
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"chain_id": cmd.ChainID,
		"duration": result.TotalDuration,
		"steps":    len(steps),
	}).Info("chain execution completed")

	return &ExecuteChainResult{
		ChainID:       cmd.ChainID,
		Success:       true,
		OutputPaths:   []string{cmd.OutputPath},
		StepsSummary:  steps,
		TotalDuration: result.TotalDuration,
	}, nil
}

// ReverseChainHandler handles chain reversal commands
type ReverseChainHandler struct {
	executor *chaininfra.ChainExecutor
}

// NewReverseChainHandler creates a new handler
func NewReverseChainHandler(executor *chaininfra.ChainExecutor) *ReverseChainHandler {
	return &ReverseChainHandler{
		executor: executor,
	}
}

// Handle executes the chain reversal command
func (h *ReverseChainHandler) Handle(ctx context.Context, cmd ReverseChainCommand) (*ReverseChainResult, error) {
	logger.Log.WithFields(logrus.Fields{
		"chain_id":    cmd.ChainID,
		"stego_media": len(cmd.StegoMediaPaths),
	}).Info("reversing technique chain")

	// Load stego media
	stegoMedia := make([][]byte, len(cmd.StegoMediaPaths))
	for i, path := range cmd.StegoMediaPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return &ReverseChainResult{
				Success: false,
				Error:   fmt.Sprintf("failed to read stego media %d: %v", i, err),
			}, err
		}
		stegoMedia[i] = data
	}

	// TODO: Load actual chain configuration
	chain := &chaining.Chain{
		ID:   cmd.ChainID,
		Mode: chaining.ChainModeSequential,
		Links: []chaining.ChainLink{
			{Order: 0, Technique: "lsb"},
			{Order: 1, Technique: "dct"},
		},
	}

	// Reverse chain
	extracted, err := h.executor.ReverseChain(ctx, chain, stegoMedia)
	if err != nil {
		return &ReverseChainResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Save output
	if err := os.WriteFile(cmd.OutputPath, extracted, 0644); err != nil {
		return &ReverseChainResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write output: %v", err),
		}, err
	}

	logger.Log.WithFields(logrus.Fields{
		"chain_id":       cmd.ChainID,
		"extracted_size": len(extracted),
	}).Info("chain reversal completed")

	return &ReverseChainResult{
		Success:       true,
		ExtractedPath: cmd.OutputPath,
		ExtractedSize: int64(len(extracted)),
	}, nil
}

package chaining

import (
	"context"
	"fmt"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// ConditionalBranch represents a decision point in chain execution
type ConditionalBranch struct {
	ID          string          `json:"id"`
	Condition   BranchCondition `json:"condition"`
	TrueChain   *Chain          `json:"true_chain"`
	FalseChain  *Chain          `json:"false_chain"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

// BranchCondition defines when to take a branch
type BranchCondition struct {
	Type      ConditionType     `json:"type"`
	Parameter string            `json:"parameter"`
	Operator  ConditionOperator `json:"operator"`
	Value     interface{}       `json:"value"`
}

// ConditionType represents the type of condition
type ConditionType string

const (
	ConditionCapacity       ConditionType = "capacity"
	ConditionDetectability  ConditionType = "detectability"
	ConditionMediaFormat    ConditionType = "media_format"
	ConditionPayloadSize    ConditionType = "payload_size"
	ConditionCarrierQuality ConditionType = "carrier_quality"
)

// ConditionOperator defines comparison operators
type ConditionOperator string

const (
	OperatorGreaterThan ConditionOperator = "gt"
	OperatorLessThan    ConditionOperator = "lt"
	OperatorEquals      ConditionOperator = "eq"
	OperatorNotEquals   ConditionOperator = "ne"
	OperatorContains    ConditionOperator = "contains"
)

// Evaluate evaluates the branch condition
func (c *BranchCondition) Evaluate(ctx context.Context, carrier []byte, payloadSize int64) (bool, error) {
	switch c.Type {
	case ConditionPayloadSize:
		return c.evaluateNumeric(float64(payloadSize))
	case ConditionMediaFormat:
		// TODO: Implement format detection
		return false, fmt.Errorf("media format condition not yet implemented")
	case ConditionCapacity:
		// TODO: Implement capacity check
		return false, fmt.Errorf("capacity condition not yet implemented")
	default:
		return false, fmt.Errorf("unknown condition type: %s", c.Type)
	}
}

func (c *BranchCondition) evaluateNumeric(actual float64) (bool, error) {
	expected, ok := c.Value.(float64)
	if !ok {
		return false, fmt.Errorf("expected numeric value for condition")
	}

	switch c.Operator {
	case OperatorGreaterThan:
		return actual > expected, nil
	case OperatorLessThan:
		return actual < expected, nil
	case OperatorEquals:
		return actual == expected, nil
	case OperatorNotEquals:
		return actual != expected, nil
	default:
		return false, fmt.Errorf("invalid operator for numeric comparison: %s", c.Operator)
	}
}

// AdaptiveWeightCalculator dynamically adjusts weights based on carrier properties
type AdaptiveWeightCalculator struct {
	MinWeight     float64
	MaxWeight     float64
	CapacityBased bool
	QualityBased  bool
}

// CalculateWeights computes optimal weights for layered chaining
func (a *AdaptiveWeightCalculator) CalculateWeights(
	ctx context.Context,
	carriers [][]byte,
	techniques []StegoTechnique,
	capacities []int64,
	qualities []float64,
) ([]float64, error) {
	if len(carriers) == 0 || len(techniques) == 0 {
		return nil, fmt.Errorf("carriers and techniques required")
	}

	weights := make([]float64, len(techniques))

	if a.CapacityBased {
		// Weight by capacity proportion
		totalCapacity := int64(0)
		for _, cap := range capacities {
			totalCapacity += cap
		}

		for i, cap := range capacities {
			if totalCapacity > 0 {
				weights[i] = float64(cap) / float64(totalCapacity)
			} else {
				weights[i] = 1.0 / float64(len(techniques))
			}
		}
	} else if a.QualityBased {
		// Weight by quality scores
		totalQuality := 0.0
		for _, q := range qualities {
			totalQuality += q
		}

		for i, q := range qualities {
			if totalQuality > 0 {
				weights[i] = q / totalQuality
			} else {
				weights[i] = 1.0 / float64(len(techniques))
			}
		}
	} else {
		// Equal weights
		weight := 1.0 / float64(len(techniques))
		for i := range weights {
			weights[i] = weight
		}
	}

	// Normalize and clamp weights
	return a.normalizeWeights(weights), nil
}

func (a *AdaptiveWeightCalculator) normalizeWeights(weights []float64) []float64 {
	sum := 0.0
	for _, w := range weights {
		sum += w
	}

	normalized := make([]float64, len(weights))
	for i, w := range weights {
		normalized[i] = w / sum

		// Clamp to min/max
		if normalized[i] < a.MinWeight {
			normalized[i] = a.MinWeight
		}
		if normalized[i] > a.MaxWeight {
			normalized[i] = a.MaxWeight
		}
	}

	// Re-normalize after clamping
	sum = 0.0
	for _, w := range normalized {
		sum += w
	}
	for i := range normalized {
		normalized[i] /= sum
	}

	return normalized
}

// ParallelExecutor enables concurrent chain execution
type ParallelExecutor struct {
	MaxWorkers    int
	ResultChannel chan *StepResult
	ErrorChannel  chan error
}

// ExecuteParallel runs chain steps concurrently
func (p *ParallelExecutor) ExecuteParallel(
	ctx context.Context,
	chain *Chain,
	payload []byte,
	carriers [][]byte,
	stegoService stego.StegoService,
) (*ChainExecutionResult, error) {
	if chain.Mode != ChainModeSplit {
		return nil, fmt.Errorf("parallel execution only supports split mode")
	}

	result := &ChainExecutionResult{
		ChainID:     chain.ID,
		Mode:        chain.Mode,
		StepResults: make([]StepResult, 0, len(chain.Links)),
		Success:     true,
	}

	start := time.Now()

	// Calculate shard size
	shardSize := len(payload) / len(chain.Links)
	if len(payload)%len(chain.Links) != 0 {
		shardSize++
	}

	// Worker pool
	jobs := make(chan parallelJob, len(chain.Links))
	results := make(chan parallelResult, len(chain.Links))

	// Start workers
	workerCount := p.MaxWorkers
	if workerCount <= 0 {
		workerCount = 4 // Default
	}
	if workerCount > len(chain.Links) {
		workerCount = len(chain.Links)
	}

	for w := 0; w < workerCount; w++ {
		go p.worker(ctx, jobs, results, stegoService)
	}

	// Submit jobs
	for i, link := range chain.Links {
		start := i * shardSize
		end := start + shardSize
		if end > len(payload) {
			end = len(payload)
		}

		job := parallelJob{
			index:   i,
			link:    link,
			shard:   payload[start:end],
			carrier: carriers[i],
		}
		jobs <- job
	}
	close(jobs)

	// Collect results
	stepResults := make([]parallelResult, len(chain.Links))
	for i := 0; i < len(chain.Links); i++ {
		res := <-results
		stepResults[res.index] = res
		if res.err != nil {
			result.Success = false
		}
	}

	// Build final result
	combinedOutput := make([]byte, 0, len(payload))
	for i, res := range stepResults {
		stepResult := StepResult{
			LinkOrder:  i,
			Technique:  res.link.Technique,
			InputSize:  int64(len(res.shard)),
			OutputSize: int64(len(res.output)),
			Duration:   res.duration,
			Success:    res.err == nil,
		}
		if res.err != nil {
			stepResult.ErrorMessage = res.err.Error()
		}
		result.StepResults = append(result.StepResults, stepResult)
		combinedOutput = append(combinedOutput, res.output...)
	}

	result.FinalOutput = combinedOutput
	result.TotalDuration = time.Since(start)

	return result, nil
}

type parallelJob struct {
	index   int
	link    ChainLink
	shard   []byte
	carrier []byte
}

type parallelResult struct {
	index    int
	link     ChainLink
	shard    []byte
	output   []byte
	duration time.Duration
	err      error
}

func (p *ParallelExecutor) worker(
	ctx context.Context,
	jobs <-chan parallelJob,
	results chan<- parallelResult,
	stegoService stego.StegoService,
) {
	for job := range jobs {
		start := time.Now()

		container, err := stegoService.Embed(ctx, job.carrier, job.shard, job.link.Technique)

		var output []byte
		if err == nil && container != nil {
			output = container.CoverMedia
		}

		results <- parallelResult{
			index:    job.index,
			link:     job.link,
			shard:    job.shard,
			output:   output,
			duration: time.Since(start),
			err:      err,
		}
	}
}

// KofNRecovery implements Reed-Solomon recovery across chain shards
type KofNRecovery struct {
	DataShards   int
	ParityShards int
	MinShards    int // K value
}

// EncodeWithRecovery adds Reed-Solomon protection to chain shards
func (k *KofNRecovery) EncodeWithRecovery(payload []byte) ([][]byte, error) {
	if k.DataShards <= 0 || k.ParityShards <= 0 {
		return nil, fmt.Errorf("invalid shard configuration")
	}

	totalShards := k.DataShards + k.ParityShards
	shardSize := (len(payload) + k.DataShards - 1) / k.DataShards

	// Pad payload to multiple of data shards
	paddedSize := shardSize * k.DataShards
	padded := make([]byte, paddedSize)
	copy(padded, payload)

	// Create data shards
	shards := make([][]byte, totalShards)
	for i := 0; i < k.DataShards; i++ {
		start := i * shardSize
		end := start + shardSize
		shards[i] = padded[start:end]
	}

	// Create parity shards (placeholder - would use Reed-Solomon library)
	for i := k.DataShards; i < totalShards; i++ {
		shards[i] = make([]byte, shardSize)
		// TODO: Implement actual Reed-Solomon parity calculation
	}

	return shards, nil
}

// RecoverFromShards reconstructs payload from K-of-N shards
func (k *KofNRecovery) RecoverFromShards(shards [][]byte, payloadSize int) ([]byte, error) {
	if len(shards) < k.MinShards {
		return nil, fmt.Errorf("insufficient shards: have %d, need %d", len(shards), k.MinShards)
	}

	// TODO: Implement actual Reed-Solomon recovery
	// For now, simple concatenation of data shards
	recovered := make([]byte, 0, payloadSize)
	for i := 0; i < k.DataShards && i < len(shards); i++ {
		if shards[i] != nil {
			recovered = append(recovered, shards[i]...)
		}
	}

	if len(recovered) > payloadSize {
		recovered = recovered[:payloadSize]
	}

	return recovered, nil
}

// ChainTemplate represents a reusable chain configuration
type ChainTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Mode        ChainMode         `json:"mode"`
	Links       []ChainLink       `json:"links"`
	Tags        []string          `json:"tags"`
	UseCase     string            `json:"use_case"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}

// TemplateLibrary manages predefined chain templates
type TemplateLibrary struct {
	templates map[string]*ChainTemplate
}

// NewTemplateLibrary creates a library with default templates
func NewTemplateLibrary() *TemplateLibrary {
	lib := &TemplateLibrary{
		templates: make(map[string]*ChainTemplate),
	}

	// Add default templates
	lib.addDefaultTemplates()
	return lib
}

func (t *TemplateLibrary) addDefaultTemplates() {
	// High Security Sequential
	t.templates["high-security-sequential"] = &ChainTemplate{
		ID:          "high-security-sequential",
		Name:        "High Security Sequential",
		Description: "Three-technique sequential chain for maximum security",
		Mode:        ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Order: 0},
			{TechniqueID: "dct", Technique: "dct", Order: 1},
			{TechniqueID: "phase", Technique: "phase", Order: 2},
		},
		Tags:    []string{"security", "sequential", "multi-layer"},
		UseCase: "Maximum security for highly sensitive data",
	}

	// Balanced Layered
	t.templates["balanced-layered"] = &ChainTemplate{
		ID:          "balanced-layered",
		Name:        "Balanced Layered",
		Description: "Two-technique layered chain with balanced weights",
		Mode:        ChainModeLayered,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Weight: 0.5},
			{TechniqueID: "dct", Technique: "dct", Weight: 0.5},
		},
		Tags:    []string{"balanced", "layered", "dual-technique"},
		UseCase: "Balanced capacity and security in single carrier",
	}

	// Wide Distribution Split
	t.templates["wide-distribution-split"] = &ChainTemplate{
		ID:          "wide-distribution-split",
		Name:        "Wide Distribution Split",
		Description: "Four-carrier split for maximum distribution",
		Mode:        ChainModeSplit,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb"},
			{TechniqueID: "dct", Technique: "dct"},
			{TechniqueID: "phase", Technique: "phase"},
			{TechniqueID: "zerowidth", Technique: "zerowidth"},
		},
		Tags:    []string{"distribution", "split", "multi-carrier"},
		UseCase: "Distribute payload across multiple media types",
	}

	// Quick Embed
	t.templates["quick-embed"] = &ChainTemplate{
		ID:          "quick-embed",
		Name:        "Quick Embed",
		Description: "Simple LSB-DCT sequential for fast embedding",
		Mode:        ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Order: 0},
			{TechniqueID: "dct", Technique: "dct", Order: 1},
		},
		Tags:    []string{"quick", "sequential", "basic"},
		UseCase: "Fast embedding with moderate security",
	}
}

// GetTemplate retrieves a template by ID
func (t *TemplateLibrary) GetTemplate(id string) (*ChainTemplate, error) {
	template, exists := t.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// ListTemplates returns all available templates
func (t *TemplateLibrary) ListTemplates() []*ChainTemplate {
	templates := make([]*ChainTemplate, 0, len(t.templates))
	for _, template := range t.templates {
		templates = append(templates, template)
	}
	return templates
}

// CreateChainFromTemplate instantiates a chain from a template
func (t *TemplateLibrary) CreateChainFromTemplate(templateID string) (*Chain, error) {
	template, err := t.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	chain := &Chain{
		ID:          fmt.Sprintf("%s-%d", templateID, time.Now().Unix()),
		Mode:        template.Mode,
		Links:       make([]ChainLink, len(template.Links)),
		Description: template.Description,
		CreatedAt:   time.Now(),
	}

	copy(chain.Links, template.Links)

	return chain, nil
}

// AddCustomTemplate adds a user-defined template
func (t *TemplateLibrary) AddCustomTemplate(template *ChainTemplate) error {
	if template.ID == "" {
		return fmt.Errorf("template ID required")
	}
	if _, exists := t.templates[template.ID]; exists {
		return fmt.Errorf("template already exists: %s", template.ID)
	}

	template.CreatedAt = time.Now()
	t.templates[template.ID] = template
	return nil
}

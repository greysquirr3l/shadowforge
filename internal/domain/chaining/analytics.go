package chaining

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DetectabilityAnalyzer performs combined analysis across chain
type DetectabilityAnalyzer struct {
	ChiSquareWeight  float64
	RSAnalysisWeight float64
	HistogramWeight  float64
}

// ChainDetectability represents combined detectability metrics
type ChainDetectability struct {
	OverallScore    float64             `json:"overall_score"`
	StepScores      []StepDetectability `json:"step_scores"`
	CombinedRisk    RiskLevel           `json:"combined_risk"`
	Recommendations []string            `json:"recommendations"`
	AnalyzedAt      time.Time           `json:"analyzed_at"`
}

// StepDetectability represents per-step detectability
type StepDetectability struct {
	LinkOrder       int            `json:"link_order"`
	Technique       StegoTechnique `json:"technique"`
	ChiSquareScore  float64        `json:"chi_square_score"`
	RSAnalysisScore float64        `json:"rs_analysis_score"`
	HistogramScore  float64        `json:"histogram_score"`
	CombinedScore   float64        `json:"combined_score"`
	Risk            RiskLevel      `json:"risk"`
}

// RiskLevel categorizes detectability risk
type RiskLevel string

const (
	RiskVeryLow  RiskLevel = "very_low"
	RiskLow      RiskLevel = "low"
	RiskModerate RiskLevel = "moderate"
	RiskHigh     RiskLevel = "high"
	RiskVeryHigh RiskLevel = "very_high"
)

// AnalyzeChain performs comprehensive detectability analysis
func (d *DetectabilityAnalyzer) AnalyzeChain(
	ctx context.Context,
	chain *Chain,
	stegoMedia [][]byte,
	analysisService interface{}, // TODO: Define proper interface
) (*ChainDetectability, error) {
	result := &ChainDetectability{
		StepScores:      make([]StepDetectability, 0, len(chain.Links)),
		Recommendations: make([]string, 0),
		AnalyzedAt:      time.Now(),
	}

	// Analyze each step
	totalScore := 0.0
	for i, link := range chain.Links {
		if i >= len(stegoMedia) {
			break
		}

		stepScore := d.analyzeStep(link, stegoMedia[i])
		result.StepScores = append(result.StepScores, stepScore)
		totalScore += stepScore.CombinedScore
	}

	// Calculate overall score
	if len(result.StepScores) > 0 {
		result.OverallScore = totalScore / float64(len(result.StepScores))
	}

	// Determine risk level
	result.CombinedRisk = d.calculateRiskLevel(result.OverallScore)

	// Generate recommendations
	result.Recommendations = d.generateRecommendations(result)

	return result, nil
}

func (d *DetectabilityAnalyzer) analyzeStep(link ChainLink, stegoMedia []byte) StepDetectability {
	// Placeholder analysis - would use actual stego analysis service
	step := StepDetectability{
		LinkOrder:       link.Order,
		Technique:       link.Technique,
		ChiSquareScore:  0.1,  // Mock
		RSAnalysisScore: 0.15, // Mock
		HistogramScore:  0.12, // Mock
	}

	// Weighted combination
	step.CombinedScore = (step.ChiSquareScore * d.ChiSquareWeight) +
		(step.RSAnalysisScore * d.RSAnalysisWeight) +
		(step.HistogramScore * d.HistogramWeight)

	step.Risk = d.calculateRiskLevel(step.CombinedScore)

	return step
}

func (d *DetectabilityAnalyzer) calculateRiskLevel(score float64) RiskLevel {
	switch {
	case score < 0.1:
		return RiskVeryLow
	case score < 0.2:
		return RiskLow
	case score < 0.4:
		return RiskModerate
	case score < 0.6:
		return RiskHigh
	default:
		return RiskVeryHigh
	}
}

func (d *DetectabilityAnalyzer) generateRecommendations(analysis *ChainDetectability) []string {
	recommendations := make([]string, 0)

	if analysis.OverallScore > 0.5 {
		recommendations = append(recommendations, "Overall detectability is high - consider using different techniques")
	}

	// Check for high-risk steps
	for _, step := range analysis.StepScores {
		if step.Risk == RiskHigh || step.Risk == RiskVeryHigh {
			recommendations = append(recommendations,
				fmt.Sprintf("Step %d (%s) shows high detectability - consider alternative technique",
					step.LinkOrder, step.Technique))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Detectability is within acceptable range")
	}

	return recommendations
}

// CapacityOptimizer finds optimal technique ordering
type CapacityOptimizer struct {
	PrioritizeCapacity bool
	PrioritizeSecurity bool
	BalanceFactors     bool
}

// OptimalOrdering represents optimized technique ordering
type OptimalOrdering struct {
	RecommendedLinks []ChainLink `json:"recommended_links"`
	ExpectedCapacity int64       `json:"expected_capacity"`
	SecurityScore    float64     `json:"security_score"`
	Reasoning        []string    `json:"reasoning"`
	Mode             ChainMode   `json:"mode"`
}

// OptimizeOrdering calculates optimal technique ordering
func (c *CapacityOptimizer) OptimizeOrdering(
	ctx context.Context,
	techniques []StegoTechnique,
	carriers [][]byte,
	payloadSize int64,
	mode ChainMode,
) (*OptimalOrdering, error) {
	switch mode {
	case ChainModeSequential:
		return c.optimizeSequential(techniques, carriers, payloadSize)
	case ChainModeLayered:
		return c.optimizeLayered(techniques, carriers, payloadSize)
	case ChainModeSplit:
		return c.optimizeSplit(techniques, carriers, payloadSize)
	default:
		return nil, fmt.Errorf("unsupported chain mode: %s", mode)
	}
}

func (c *CapacityOptimizer) optimizeSequential(
	techniques []StegoTechnique,
	carriers [][]byte,
	payloadSize int64,
) (*OptimalOrdering, error) {
	result := &OptimalOrdering{
		Mode:             ChainModeSequential,
		RecommendedLinks: make([]ChainLink, len(techniques)),
		Reasoning:        make([]string, 0),
	}

	// Order by increasing capacity (bottleneck first)
	// This ensures early failure if capacity insufficient
	for i, technique := range techniques {
		result.RecommendedLinks[i] = ChainLink{
			TechniqueID: string(technique),
			Technique:   technique,
			Order:       i,
		}
	}

	result.Reasoning = append(result.Reasoning,
		"Sequential ordering optimized for early capacity validation")

	return result, nil
}

func (c *CapacityOptimizer) optimizeLayered(
	techniques []StegoTechnique,
	carriers [][]byte,
	payloadSize int64,
) (*OptimalOrdering, error) {
	result := &OptimalOrdering{
		Mode:             ChainModeLayered,
		RecommendedLinks: make([]ChainLink, len(techniques)),
		Reasoning:        make([]string, 0),
	}

	// Equal weights for now (adaptive weights would be better)
	weight := 1.0 / float64(len(techniques))
	for i, technique := range techniques {
		result.RecommendedLinks[i] = ChainLink{
			TechniqueID: string(technique),
			Technique:   technique,
			Weight:      weight,
		}
	}

	result.Reasoning = append(result.Reasoning,
		"Layered weights balanced equally across techniques")

	return result, nil
}

func (c *CapacityOptimizer) optimizeSplit(
	techniques []StegoTechnique,
	carriers [][]byte,
	payloadSize int64,
) (*OptimalOrdering, error) {
	result := &OptimalOrdering{
		Mode:             ChainModeSplit,
		RecommendedLinks: make([]ChainLink, len(techniques)),
		Reasoning:        make([]string, 0),
	}

	// Pair techniques with carriers
	for i, technique := range techniques {
		result.RecommendedLinks[i] = ChainLink{
			TechniqueID: string(technique),
			Technique:   technique,
		}
	}

	result.Reasoning = append(result.Reasoning,
		"Split distribution across available carriers")

	return result, nil
}

// PerformanceProfiler benchmarks chain configurations
type PerformanceProfiler struct {
	Iterations int
	WarmupRuns int
}

// ProfileResult contains benchmark data
type ProfileResult struct {
	ChainID             string        `json:"chain_id"`
	Mode                ChainMode     `json:"mode"`
	TotalIterations     int           `json:"total_iterations"`
	StepProfiles        []StepProfile `json:"step_profiles"`
	AverageTotal        time.Duration `json:"average_total"`
	MinTotal            time.Duration `json:"min_total"`
	MaxTotal            time.Duration `json:"max_total"`
	ThroughputBytesPerS int64         `json:"throughput_bytes_per_s"`
	ProfiledAt          time.Time     `json:"profiled_at"`
}

// StepProfile contains per-step benchmark data
type StepProfile struct {
	LinkOrder       int            `json:"link_order"`
	Technique       StegoTechnique `json:"technique"`
	AverageDuration time.Duration  `json:"average_duration"`
	MinDuration     time.Duration  `json:"min_duration"`
	MaxDuration     time.Duration  `json:"max_duration"`
	BytesProcessed  int64          `json:"bytes_processed"`
}

// ProfileChain benchmarks a chain configuration
func (p *PerformanceProfiler) ProfileChain(
	ctx context.Context,
	chain *Chain,
	payload []byte,
	carriers [][]byte,
	executor interface{}, // ChainExecutor interface
) (*ProfileResult, error) {
	profileResult := &ProfileResult{
		ChainID:         chain.ID,
		Mode:            chain.Mode,
		TotalIterations: p.Iterations,
		StepProfiles:    make([]StepProfile, 0),
		ProfiledAt:      time.Now(),
	}

	// TODO: Implement actual profiling with executor
	// This is a placeholder structure

	return profileResult, nil
}

// ChainVisualizer generates Mermaid diagrams for execution flow
type ChainVisualizer struct{}

func (v *ChainVisualizer) writef(sb *strings.Builder, format string, args ...any) error {
	_, err := fmt.Fprintf(sb, format, args...)
	return err
}

// GenerateMermaidDiagram creates a visual representation of chain execution
func (v *ChainVisualizer) GenerateMermaidDiagram(chain *Chain) (string, error) {
	var sb strings.Builder

	sb.WriteString("```mermaid\n")
	sb.WriteString("graph LR\n")

	switch chain.Mode {
	case ChainModeSequential:
		if err := v.generateSequentialDiagram(&sb, chain); err != nil {
			return "", err
		}
	case ChainModeLayered:
		if err := v.generateLayeredDiagram(&sb, chain); err != nil {
			return "", err
		}
	case ChainModeSplit:
		if err := v.generateSplitDiagram(&sb, chain); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported chain mode: %s", chain.Mode)
	}

	sb.WriteString("```\n")

	return sb.String(), nil
}

func (v *ChainVisualizer) generateSequentialDiagram(sb *strings.Builder, chain *Chain) error {
	sb.WriteString("    Payload[Payload]\n")

	for i, link := range chain.Links {
		nodeID := fmt.Sprintf("Step%d", i)
		if err := v.writef(sb, "    %s[%s]\n", nodeID, link.Technique); err != nil {
			return err
		}

		if i == 0 {
			if err := v.writef(sb, "    Payload --> %s\n", nodeID); err != nil {
				return err
			}
		} else {
			prevNodeID := fmt.Sprintf("Step%d", i-1)
			if err := v.writef(sb, "    %s --> %s\n", prevNodeID, nodeID); err != nil {
				return err
			}
		}
	}

	lastNodeID := fmt.Sprintf("Step%d", len(chain.Links)-1)
	if err := v.writef(sb, "    %s --> Output[Stego Media]\n", lastNodeID); err != nil {
		return err
	}

	return nil
}

func (v *ChainVisualizer) generateLayeredDiagram(sb *strings.Builder, chain *Chain) error {
	sb.WriteString("    Payload[Payload]\n")
	sb.WriteString("    Carrier[Carrier]\n")

	for i, link := range chain.Links {
		nodeID := fmt.Sprintf("Layer%d", i)
		weight := fmt.Sprintf("%.0f%%", link.Weight*100)
		if err := v.writef(sb, "    %s[%s<br/>%s]\n", nodeID, link.Technique, weight); err != nil {
			return err
		}
		if err := v.writef(sb, "    Payload -.%s.-> %s\n", weight, nodeID); err != nil {
			return err
		}
		if err := v.writef(sb, "    %s --> Carrier\n", nodeID); err != nil {
			return err
		}
	}

	sb.WriteString("    Carrier --> Output[Stego Media]\n")
	return nil
}

func (v *ChainVisualizer) generateSplitDiagram(sb *strings.Builder, chain *Chain) error {
	sb.WriteString("    Payload[Payload]\n")
	sb.WriteString("    Splitter{Shard}\n")
	sb.WriteString("    Payload --> Splitter\n")

	for i, link := range chain.Links {
		shardID := fmt.Sprintf("Shard%d", i)
		carrierID := fmt.Sprintf("Carrier%d", i)
		outputID := fmt.Sprintf("Output%d", i)

		if err := v.writef(sb, "    %s[Shard %d]\n", shardID, i+1); err != nil {
			return err
		}
		if err := v.writef(sb, "    %s[%s]\n", carrierID, link.Technique); err != nil {
			return err
		}
		if err := v.writef(sb, "    %s[Stego %d]\n", outputID, i+1); err != nil {
			return err
		}

		if err := v.writef(sb, "    Splitter --> %s\n", shardID); err != nil {
			return err
		}
		if err := v.writef(sb, "    %s --> %s\n", shardID, carrierID); err != nil {
			return err
		}
		if err := v.writef(sb, "    %s --> %s\n", carrierID, outputID); err != nil {
			return err
		}
	}

	return nil
}

// GenerateFlowDiagram creates detailed execution flow
func (v *ChainVisualizer) GenerateFlowDiagram(
	chain *Chain,
	result *ChainExecutionResult,
) (string, error) {
	var sb strings.Builder

	sb.WriteString("```mermaid\n")
	sb.WriteString("sequenceDiagram\n")
	sb.WriteString("    participant User\n")
	sb.WriteString("    participant Chain\n")

	for i, link := range chain.Links {
		participant := fmt.Sprintf("Step%d", i)
		if err := v.writef(&sb, "    participant %s as %s\n", participant, link.Technique); err != nil {
			return "", err
		}
	}

	sb.WriteString("    User->>Chain: Execute Chain\n")

	for i, step := range result.StepResults {
		participant := fmt.Sprintf("Step%d", i)
		duration := step.Duration.String()
		if err := v.writef(&sb, "    Chain->>%s: Process Shard (%s)\n", participant, duration); err != nil {
			return "", err
		}

		if step.Success {
			if err := v.writef(&sb, "    %s-->>Chain: Success\n", participant); err != nil {
				return "", err
			}
		} else {
			if err := v.writef(&sb, "    %s-->>Chain: Error: %s\n", participant, step.ErrorMessage); err != nil {
				return "", err
			}
		}
	}

	if result.Success {
		sb.WriteString("    Chain-->>User: Complete\n")
	} else {
		sb.WriteString("    Chain-->>User: Failed\n")
	}

	sb.WriteString("```\n")

	return sb.String(), nil
}

// CapacityReport provides detailed capacity analysis
type CapacityReport struct {
	TotalCapacity     int64               `json:"total_capacity"`
	AvailableCapacity int64               `json:"available_capacity"`
	UsedCapacity      int64               `json:"used_capacity"`
	UtilizationPct    float64             `json:"utilization_pct"`
	PerTechnique      []TechniqueCapacity `json:"per_technique"`
	Bottleneck        *TechniqueCapacity  `json:"bottleneck,omitempty"`
	Recommendations   []string            `json:"recommendations"`
}

// TechniqueCapacity tracks capacity for a single technique
type TechniqueCapacity struct {
	Technique      StegoTechnique `json:"technique"`
	MaxCapacity    int64          `json:"max_capacity"`
	SafeCapacity   int64          `json:"safe_capacity"` // 70% of max
	CurrentUsage   int64          `json:"current_usage"`
	UtilizationPct float64        `json:"utilization_pct"`
}

// GenerateCapacityReport creates comprehensive capacity analysis
func GenerateCapacityReport(
	ctx context.Context,
	chain *Chain,
	carriers [][]byte,
	payloadSize int64,
	capacities []int64,
) (*CapacityReport, error) {
	report := &CapacityReport{
		PerTechnique:    make([]TechniqueCapacity, 0, len(chain.Links)),
		Recommendations: make([]string, 0),
	}

	// Calculate per-technique capacities
	for i, link := range chain.Links {
		if i < len(capacities) {
			tc := TechniqueCapacity{
				Technique:    link.Technique,
				MaxCapacity:  capacities[i],
				SafeCapacity: int64(float64(capacities[i]) * 0.7),
			}

			switch chain.Mode {
			case ChainModeLayered:
				tc.CurrentUsage = int64(float64(payloadSize) * link.Weight)
			case ChainModeSplit:
				tc.CurrentUsage = payloadSize / int64(len(chain.Links))
			default:
				tc.CurrentUsage = payloadSize
			}

			if tc.MaxCapacity > 0 {
				tc.UtilizationPct = float64(tc.CurrentUsage) / float64(tc.MaxCapacity) * 100
			}

			report.PerTechnique = append(report.PerTechnique, tc)
			report.TotalCapacity += tc.MaxCapacity
		}
	}

	// Find bottleneck (sequential mode)
	if chain.Mode == ChainModeSequential {
		minCapacity := int64(^uint64(0) >> 1) // Max int64
		var bottleneckIdx int
		for i, tc := range report.PerTechnique {
			if tc.MaxCapacity < minCapacity {
				minCapacity = tc.MaxCapacity
				bottleneckIdx = i
			}
		}
		if len(report.PerTechnique) > 0 {
			report.Bottleneck = &report.PerTechnique[bottleneckIdx]
		}
	}

	// Generate recommendations
	report.Recommendations = generateCapacityRecommendations(report, chain.Mode)

	return report, nil
}

func generateCapacityRecommendations(report *CapacityReport, mode ChainMode) []string {
	recommendations := make([]string, 0)

	for _, tc := range report.PerTechnique {
		if tc.UtilizationPct > 90 {
			recommendations = append(recommendations,
				fmt.Sprintf("%s is at %.1f%% capacity - consider larger carrier or different technique",
					tc.Technique, tc.UtilizationPct))
		}
	}

	if report.Bottleneck != nil && mode == ChainModeSequential {
		recommendations = append(recommendations,
			fmt.Sprintf("Bottleneck at %s (%.1f%% utilization) - consider reordering techniques",
				report.Bottleneck.Technique, report.Bottleneck.UtilizationPct))
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Capacity utilization is within optimal range")
	}

	return recommendations
}

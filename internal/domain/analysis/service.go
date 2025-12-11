// Package analysis provides security analysis service interfaces.
package analysis

import "context"

// Service defines domain operations for security analysis.
type Service interface {
	// AnalyzeAsset performs security analysis on a target asset.
	AnalyzeAsset(ctx context.Context, targetAssetID string, analysisType AnalysisType) (*SecurityAnalysis, error)

	// RunChiSquareTest performs chi-square statistical analysis.
	RunChiSquareTest(ctx context.Context, data []byte) (float64, error)

	// RunRSAnalysis performs Regular-Singular analysis.
	RunRSAnalysis(ctx context.Context, data []byte) (float64, error)

	// CalculateEntropy calculates Shannon entropy of data.
	CalculateEntropy(ctx context.Context, data []byte) (float64, error)

	// CalculateDetectionProbability calculates overall detection probability.
	CalculateDetectionProbability(ctx context.Context, analysis *SecurityAnalysis) (float64, error)

	// AssessThreat creates a threat assessment for a specific vector.
	AssessThreat(ctx context.Context, analysis *SecurityAnalysis, vector ThreatVector) (*ThreatAssessment, error)

	// GenerateReport generates a comprehensive analysis report.
	GenerateReport(ctx context.Context, analysis *SecurityAnalysis) ([]byte, error)
}

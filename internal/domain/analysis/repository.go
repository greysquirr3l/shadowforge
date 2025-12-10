// Package analysis provides security analysis repository interfaces.
package analysis

import "context"

// Repository defines persistence operations for security analysis.
type Repository interface {
// SaveAnalysis persists a security analysis.
SaveAnalysis(ctx context.Context, analysis *SecurityAnalysis) error

// GetAnalysis retrieves an analysis by ID.
GetAnalysis(ctx context.Context, id AnalysisID) (*SecurityAnalysis, error)

// ListAnalyses retrieves analyses with pagination.
ListAnalyses(ctx context.Context, targetAssetID string, limit, offset int) ([]*SecurityAnalysis, error)

// DeleteAnalysis removes an analysis.
DeleteAnalysis(ctx context.Context, id AnalysisID) error

// SaveThreatAssessment persists a threat assessment.
SaveThreatAssessment(ctx context.Context, assessment *ThreatAssessment) error

// GetThreatAssessments retrieves threat assessments for an analysis.
GetThreatAssessments(ctx context.Context, analysisID AnalysisID) ([]*ThreatAssessment, error)

// GetCriticalThreats retrieves critical threats for a target asset.
GetCriticalThreats(ctx context.Context, targetAssetID string) ([]*ThreatAssessment, error)
}

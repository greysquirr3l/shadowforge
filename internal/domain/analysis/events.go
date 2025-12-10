// Package analysis provides security analysis domain events.
package analysis

import "time"

// AnalysisStarted is emitted when a security analysis begins.
type AnalysisStarted struct {
ID            AnalysisID
TargetAssetID string
AnalysisType  AnalysisType
StartedAt     time.Time
}

// AnalysisCompleted is emitted when a security analysis completes successfully.
type AnalysisCompleted struct {
ID                   AnalysisID
ThreatLevel          ThreatLevel
DetectionProbability float64
ChiSquareScore       float64
RSAnalysisScore      float64
EntropyScore         float64
Duration             time.Duration
CompletedAt          time.Time
}

// ThreatDetected is emitted when a security threat is detected.
type ThreatDetected struct {
AssessmentID AssessmentID
AnalysisID   AnalysisID
ThreatVector ThreatVector
Severity     Severity
Confidence   float64
DetectedAt   time.Time
}

// CriticalThreatDetected is emitted when a critical threat is detected.
type CriticalThreatDetected struct {
AssessmentID            AssessmentID
AnalysisID              AnalysisID
ThreatVector            ThreatVector
Severity                Severity
RequiresImmediateAction bool
DetectedAt              time.Time
}

// AnalysisExpired is emitted when analysis results expire.
type AnalysisExpired struct {
ID                 AnalysisID
OriginalThreatLevel ThreatLevel
ExpiredAt          time.Time
}

// ReportGenerated is emitted when an analysis report is generated.
type ReportGenerated struct {
ID           AnalysisID
ReportFormat string
ReportSize   int64
GeneratedAt  time.Time
}

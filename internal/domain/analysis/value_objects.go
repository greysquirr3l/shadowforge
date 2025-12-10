// Package analysis provides security analysis value objects.
package analysis

import (
	"github.com/google/uuid"
)

// AnalysisID represents a unique identifier for security analysis.
type AnalysisID struct {
	value string
}

// NewAnalysisID creates a new analysis ID.
func NewAnalysisID() (AnalysisID, error) {
	id := uuid.New().String()
	return AnalysisID{value: id}, nil
}

// NewAnalysisIDFromString creates an analysis ID from string.
func NewAnalysisIDFromString(s string) (AnalysisID, error) {
	if s == "" {
		return AnalysisID{}, ErrInvalidAnalysisID
	}
	if _, err := uuid.Parse(s); err != nil {
		return AnalysisID{}, ErrInvalidAnalysisID
	}
	return AnalysisID{value: s}, nil
}

// String returns string representation.
func (id AnalysisID) String() string {
	return id.value
}

// IsZero checks if ID is zero value.
func (id AnalysisID) IsZero() bool {
	return id.value == ""
}

// AssessmentID represents a unique identifier for threat assessment.
type AssessmentID struct {
	value string
}

// NewAssessmentID creates a new assessment ID.
func NewAssessmentID() (AssessmentID, error) {
	id := uuid.New().String()
	return AssessmentID{value: id}, nil
}

// NewAssessmentIDFromString creates an assessment ID from string.
func NewAssessmentIDFromString(s string) (AssessmentID, error) {
	if s == "" {
		return AssessmentID{}, ErrInvalidAssessmentID
	}
	if _, err := uuid.Parse(s); err != nil {
		return AssessmentID{}, ErrInvalidAssessmentID
	}
	return AssessmentID{value: s}, nil
}

// String returns string representation.
func (id AssessmentID) String() string {
	return id.value
}

// IsZero checks if ID is zero value.
func (id AssessmentID) IsZero() bool {
	return id.value == ""
}

// AnalysisType represents the type of security analysis performed.
type AnalysisType string

const (
	// AnalysisTypeChiSquare represents chi-square statistical analysis.
	AnalysisTypeChiSquare AnalysisType = "chi_square"

	// AnalysisTypeRS represents RS (Regular-Singular) analysis.
	AnalysisTypeRS AnalysisType = "rs_analysis"

	// AnalysisTypeEntropy represents entropy-based analysis.
	AnalysisTypeEntropy AnalysisType = "entropy"

	// AnalysisTypeVisual represents visual pattern analysis.
	AnalysisTypeVisual AnalysisType = "visual"

	// AnalysisTypeHistogram represents histogram analysis.
	AnalysisTypeHistogram AnalysisType = "histogram"

	// AnalysisTypeFull represents comprehensive full analysis.
	AnalysisTypeFull AnalysisType = "full"
)

// IsValid checks if analysis type is valid.
func (at AnalysisType) IsValid() bool {
	switch at {
	case AnalysisTypeChiSquare, AnalysisTypeRS, AnalysisTypeEntropy,
		AnalysisTypeVisual, AnalysisTypeHistogram, AnalysisTypeFull:
		return true
	default:
		return false
	}
}

// String returns string representation.
func (at AnalysisType) String() string {
	return string(at)
}

// ThreatLevel represents the overall threat level from analysis.
type ThreatLevel string

const (
	// ThreatLevelNone indicates no threat detected.
	ThreatLevelNone ThreatLevel = "none"

	// ThreatLevelLow indicates low threat level.
	ThreatLevelLow ThreatLevel = "low"

	// ThreatLevelMedium indicates medium threat level.
	ThreatLevelMedium ThreatLevel = "medium"

	// ThreatLevelHigh indicates high threat level.
	ThreatLevelHigh ThreatLevel = "high"

	// ThreatLevelCritical indicates critical threat level.
	ThreatLevelCritical ThreatLevel = "critical"
)

// IsValid checks if threat level is valid.
func (tl ThreatLevel) IsValid() bool {
	switch tl {
	case ThreatLevelNone, ThreatLevelLow, ThreatLevelMedium,
		ThreatLevelHigh, ThreatLevelCritical:
		return true
	default:
		return false
	}
}

// String returns string representation.
func (tl ThreatLevel) String() string {
	return string(tl)
}

// RequiresAction returns true if threat level requires action.
func (tl ThreatLevel) RequiresAction() bool {
	return tl == ThreatLevelHigh || tl == ThreatLevelCritical
}

// ThreatVector represents a specific threat attack vector.
type ThreatVector string

const (
	// VectorChiSquare represents chi-square anomaly detection.
	VectorChiSquare ThreatVector = "chi_square_anomaly"

	// VectorRSSignature represents RS analysis signature detection.
	VectorRSSignature ThreatVector = "rs_signature"

	// VectorEntropyAnomaly represents entropy anomaly detection.
	VectorEntropyAnomaly ThreatVector = "entropy_anomaly"

	// VectorVisualPattern represents visual pattern detection.
	VectorVisualPattern ThreatVector = "visual_pattern"

	// VectorHistogramShift represents histogram shift detection.
	VectorHistogramShift ThreatVector = "histogram_shift"

	// VectorLSBPatterns represents LSB pattern detection.
	VectorLSBPatterns ThreatVector = "lsb_patterns"
)

// IsValid checks if threat vector is valid.
func (tv ThreatVector) IsValid() bool {
	switch tv {
	case VectorChiSquare, VectorRSSignature, VectorEntropyAnomaly,
		VectorVisualPattern, VectorHistogramShift, VectorLSBPatterns:
		return true
	default:
		return false
	}
}

// String returns string representation.
func (tv ThreatVector) String() string {
	return string(tv)
}

// Severity represents threat severity level.
type Severity string

const (
	// SeverityInfo represents informational severity.
	SeverityInfo Severity = "info"

	// SeverityLow represents low severity.
	SeverityLow Severity = "low"

	// SeverityMedium represents medium severity.
	SeverityMedium Severity = "medium"

	// SeverityHigh represents high severity.
	SeverityHigh Severity = "high"

	// SeverityCritical represents critical severity.
	SeverityCritical Severity = "critical"
)

// IsValid checks if severity is valid.
func (s Severity) IsValid() bool {
	switch s {
	case SeverityInfo, SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	default:
		return false
	}
}

// String returns string representation.
func (s Severity) String() string {
	return string(s)
}

// IsGreaterThan compares severity levels.
func (s Severity) IsGreaterThan(other Severity) bool {
	severityOrder := map[Severity]int{
		SeverityInfo:     0,
		SeverityLow:      1,
		SeverityMedium:   2,
		SeverityHigh:     3,
		SeverityCritical: 4,
	}
	return severityOrder[s] > severityOrder[other]
}

// AnalysisStatus represents the status of security analysis.
type AnalysisStatus string

const (
	// StatusPending indicates analysis is pending.
	StatusPending AnalysisStatus = "pending"

	// StatusRunning indicates analysis is in progress.
	StatusRunning AnalysisStatus = "running"

	// StatusComplete indicates analysis completed successfully.
	StatusComplete AnalysisStatus = "complete"

	// StatusFailed indicates analysis failed.
	StatusFailed AnalysisStatus = "failed"
)

// IsValid checks if status is valid.
func (as AnalysisStatus) IsValid() bool {
	switch as {
	case StatusPending, StatusRunning, StatusComplete, StatusFailed:
		return true
	default:
		return false
	}
}

// String returns string representation.
func (as AnalysisStatus) String() string {
	return string(as)
}

// IsFinal returns true if status is terminal.
func (as AnalysisStatus) IsFinal() bool {
	return as == StatusComplete || as == StatusFailed
}

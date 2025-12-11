// Package analysis provides security analysis domain entities.
package analysis

import (
	"time"
)

// SecurityAnalysis is the root aggregate for security analysis operations.
type SecurityAnalysis struct {
	ID                   AnalysisID
	TargetAssetID        string // References media.AssetID
	AnalysisType         AnalysisType
	ChiSquareScore       float64
	RSAnalysisScore      float64
	EntropyScore         float64
	ThreatLevel          ThreatLevel
	DetectionProbability float64
	Recommendations      []string
	PerformedAt          time.Time
	CompletedAt          time.Time
	Status               AnalysisStatus
}

// NewSecurityAnalysis creates a new security analysis.
func NewSecurityAnalysis(targetAssetID string, analysisType AnalysisType) (*SecurityAnalysis, error) {
	if targetAssetID == "" {
		return nil, ErrEmptyTargetAsset
	}

	id, err := NewAnalysisID()
	if err != nil {
		return nil, err
	}

	return &SecurityAnalysis{
		ID:              id,
		TargetAssetID:   targetAssetID,
		AnalysisType:    analysisType,
		Status:          StatusPending,
		PerformedAt:     time.Now(),
		Recommendations: []string{},
	}, nil
}

// SetChiSquareScore sets the chi-square test score.
func (s *SecurityAnalysis) SetChiSquareScore(score float64) error {
	if score < 0 || score > 1 {
		return ErrInvalidScore
	}
	s.ChiSquareScore = score
	return nil
}

// SetRSAnalysisScore sets the RS analysis score.
func (s *SecurityAnalysis) SetRSAnalysisScore(score float64) error {
	if score < 0 || score > 1 {
		return ErrInvalidScore
	}
	s.RSAnalysisScore = score
	return nil
}

// SetEntropyScore sets the entropy score.
func (s *SecurityAnalysis) SetEntropyScore(score float64) error {
	if score < 0 || score > 1 {
		return ErrInvalidScore
	}
	s.EntropyScore = score
	return nil
}

// CalculateThreatLevel calculates the overall threat level based on scores.
func (s *SecurityAnalysis) CalculateThreatLevel() {
	avgScore := (s.ChiSquareScore + s.RSAnalysisScore + s.EntropyScore) / 3.0

	switch {
	case avgScore < 0.25:
		s.ThreatLevel = ThreatLevelLow
	case avgScore < 0.50:
		s.ThreatLevel = ThreatLevelMedium
	case avgScore < 0.75:
		s.ThreatLevel = ThreatLevelHigh
	default:
		s.ThreatLevel = ThreatLevelCritical
	}
}

// SetDetectionProbability sets the probability of detection.
func (s *SecurityAnalysis) SetDetectionProbability(probability float64) error {
	if probability < 0 || probability > 1 {
		return ErrInvalidProbability
	}
	s.DetectionProbability = probability
	return nil
}

// AddRecommendation adds a security recommendation.
func (s *SecurityAnalysis) AddRecommendation(recommendation string) {
	if recommendation != "" {
		s.Recommendations = append(s.Recommendations, recommendation)
	}
}

// Complete marks the analysis as complete.
func (s *SecurityAnalysis) Complete() {
	s.Status = StatusComplete
	s.CompletedAt = time.Now()
}

// Fail marks the analysis as failed.
func (s *SecurityAnalysis) Fail() {
	s.Status = StatusFailed
	s.CompletedAt = time.Now()
}

// Validate validates the security analysis.
func (s *SecurityAnalysis) Validate() error {
	if s.ID.IsZero() {
		return ErrInvalidAnalysisID
	}
	if s.TargetAssetID == "" {
		return ErrEmptyTargetAsset
	}
	if !s.AnalysisType.IsValid() {
		return ErrInvalidAnalysisType
	}
	return nil
}

// ThreatAssessment represents a detailed threat assessment for a specific vector.
type ThreatAssessment struct {
	ID           AssessmentID
	AnalysisID   AnalysisID
	ThreatVector ThreatVector
	Severity     Severity
	Description  string
	Mitigation   string
	AssessedAt   time.Time
}

// NewThreatAssessment creates a new threat assessment.
func NewThreatAssessment(analysisID AnalysisID, vector ThreatVector, severity Severity) (*ThreatAssessment, error) {
	if analysisID.IsZero() {
		return nil, ErrInvalidAnalysisID
	}
	if !vector.IsValid() {
		return nil, ErrInvalidThreatVector
	}
	if !severity.IsValid() {
		return nil, ErrInvalidSeverity
	}

	id, err := NewAssessmentID()
	if err != nil {
		return nil, err
	}

	return &ThreatAssessment{
		ID:           id,
		AnalysisID:   analysisID,
		ThreatVector: vector,
		Severity:     severity,
		AssessedAt:   time.Now(),
	}, nil
}

// SetDescription sets the threat description.
func (t *ThreatAssessment) SetDescription(description string) {
	t.Description = description
}

// SetMitigation sets the mitigation strategy.
func (t *ThreatAssessment) SetMitigation(mitigation string) {
	t.Mitigation = mitigation
}

// Validate validates the threat assessment.
func (t *ThreatAssessment) Validate() error {
	if t.ID.IsZero() {
		return ErrInvalidAssessmentID
	}
	if t.AnalysisID.IsZero() {
		return ErrInvalidAnalysisID
	}
	if !t.ThreatVector.IsValid() {
		return ErrInvalidThreatVector
	}
	if !t.Severity.IsValid() {
		return ErrInvalidSeverity
	}
	return nil
}

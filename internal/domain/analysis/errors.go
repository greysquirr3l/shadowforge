// Package analysis provides security analysis errors.
package analysis

import "errors"

var (
// ErrInvalidAnalysisID is returned when analysis ID is invalid.
ErrInvalidAnalysisID = errors.New("invalid analysis ID")

// ErrEmptyTargetAsset is returned when target asset ID is empty.
ErrEmptyTargetAsset = errors.New("target asset ID cannot be empty")

// ErrInvalidAnalysisType is returned when analysis type is invalid.
ErrInvalidAnalysisType = errors.New("invalid analysis type")

// ErrInvalidScore is returned when score is out of valid range.
ErrInvalidScore = errors.New("score must be between 0.0 and 1.0")

// ErrInvalidProbability is returned when probability is out of valid range.
ErrInvalidProbability = errors.New("probability must be between 0.0 and 1.0")

// ErrInvalidAssessmentID is returned when assessment ID is invalid.
ErrInvalidAssessmentID = errors.New("invalid assessment ID")

// ErrInvalidThreatVector is returned when threat vector is invalid.
ErrInvalidThreatVector = errors.New("invalid threat vector")

// ErrInvalidSeverity is returned when severity is invalid.
ErrInvalidSeverity = errors.New("invalid severity")

// ErrEmptyRecommendation is returned when recommendation is empty.
ErrEmptyRecommendation = errors.New("recommendation cannot be empty")

// ErrEmptyDescription is returned when description is empty.
ErrEmptyDescription = errors.New("description cannot be empty")

// ErrEmptyMitigation is returned when mitigation is empty.
ErrEmptyMitigation = errors.New("mitigation cannot be empty")

// ErrAnalysisNotComplete is returned when analysis is not in complete state.
ErrAnalysisNotComplete = errors.New("analysis is not complete")

// ErrAnalysisAlreadyComplete is returned when analysis is already complete.
ErrAnalysisAlreadyComplete = errors.New("analysis is already complete")
)

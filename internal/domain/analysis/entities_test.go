package analysis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================
// SecurityAnalysis Entity Tests
// =====================================================

func TestSecurityAnalysis_NewSecurityAnalysis_Success(t *testing.T) {
	tests := []struct {
		name          string
		targetAssetID string
		analysisType  AnalysisType
	}{
		{"chi_square_analysis", "asset-123", AnalysisTypeChiSquare},
		{"rs_analysis", "asset-456", AnalysisTypeRS},
		{"entropy_analysis", "asset-789", AnalysisTypeEntropy},
		{"full_analysis", "asset-abc", AnalysisTypeFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := NewSecurityAnalysis(tt.targetAssetID, tt.analysisType)

			require.NoError(t, err)
			require.NotNil(t, analysis)
			assert.False(t, analysis.ID.IsZero())
			assert.Equal(t, tt.targetAssetID, analysis.TargetAssetID)
			assert.Equal(t, tt.analysisType, analysis.AnalysisType)
			assert.Equal(t, StatusPending, analysis.Status)
			assert.WithinDuration(t, time.Now(), analysis.PerformedAt, time.Second)
			assert.Empty(t, analysis.Recommendations)
		})
	}
}

func TestSecurityAnalysis_NewSecurityAnalysis_EmptyTargetAsset(t *testing.T) {
	analysis, err := NewSecurityAnalysis("", AnalysisTypeChiSquare)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyTargetAsset)
	assert.Nil(t, analysis)
}

func TestSecurityAnalysis_SetChiSquareScore_Success(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"zero_score", 0.0},
		{"low_score", 0.25},
		{"medium_score", 0.5},
		{"high_score", 0.75},
		{"max_score", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeChiSquare)
			err := analysis.SetChiSquareScore(tt.score)

			require.NoError(t, err)
			assert.Equal(t, tt.score, analysis.ChiSquareScore)
		})
	}
}

func TestSecurityAnalysis_SetChiSquareScore_InvalidScore(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"negative_score", -0.1},
		{"above_range", 1.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeChiSquare)
			err := analysis.SetChiSquareScore(tt.score)

			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidScore)
		})
	}
}

func TestSecurityAnalysis_CalculateThreatLevel(t *testing.T) {
	tests := []struct {
		name            string
		chiSquareScore  float64
		rsAnalysisScore float64
		entropyScore    float64
		expectedLevel   ThreatLevel
	}{
		{"all_zero_low", 0.0, 0.0, 0.0, ThreatLevelLow},
		{"low_boundary", 0.2, 0.2, 0.2, ThreatLevelLow},
		{"medium_boundary", 0.3, 0.4, 0.4, ThreatLevelMedium},
		{"high_boundary", 0.6, 0.6, 0.6, ThreatLevelHigh},
		{"critical", 0.8, 0.9, 0.85, ThreatLevelCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeFull)
			_ = analysis.SetChiSquareScore(tt.chiSquareScore)
			_ = analysis.SetRSAnalysisScore(tt.rsAnalysisScore)
			_ = analysis.SetEntropyScore(tt.entropyScore)

			analysis.CalculateThreatLevel()

			assert.Equal(t, tt.expectedLevel, analysis.ThreatLevel)
		})
	}
}

func TestSecurityAnalysis_AddRecommendation(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeFull)

	analysis.AddRecommendation("Use lower embedding rate")
	analysis.AddRecommendation("")
	analysis.AddRecommendation("Increase noise")

	assert.Len(t, analysis.Recommendations, 2)
}

func TestSecurityAnalysis_Complete(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeFull)

	analysis.Complete()

	assert.Equal(t, StatusComplete, analysis.Status)
	assert.False(t, analysis.CompletedAt.IsZero())
}

func TestSecurityAnalysis_Validate_Success(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeChiSquare)

	err := analysis.Validate()

	assert.NoError(t, err)
}

// =====================================================
// ThreatAssessment Entity Tests
// =====================================================

func TestThreatAssessment_NewThreatAssessment_Success(t *testing.T) {
	analysisID, _ := NewAnalysisID()

	assessment, err := NewThreatAssessment(analysisID, VectorChiSquare, SeverityCritical)

	require.NoError(t, err)
	require.NotNil(t, assessment)
	assert.False(t, assessment.ID.IsZero())
	assert.Equal(t, analysisID, assessment.AnalysisID)
}

func TestThreatAssessment_NewThreatAssessment_InvalidAnalysisID(t *testing.T) {
	assessment, err := NewThreatAssessment(AnalysisID{}, VectorChiSquare, SeverityHigh)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAnalysisID)
	assert.Nil(t, assessment)
}

func TestThreatAssessment_Validate_Success(t *testing.T) {
	analysisID, _ := NewAnalysisID()
	assessment, _ := NewThreatAssessment(analysisID, VectorChiSquare, SeverityHigh)

	err := assessment.Validate()

	assert.NoError(t, err)
}

func TestSecurityAnalysis_SetRSAnalysisScore_Success(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeRS)

	err := analysis.SetRSAnalysisScore(0.42)

	require.NoError(t, err)
	assert.Equal(t, 0.42, analysis.RSAnalysisScore)
}

func TestSecurityAnalysis_SetRSAnalysisScore_InvalidScore(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeRS)

	err := analysis.SetRSAnalysisScore(1.5)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidScore)
}

func TestSecurityAnalysis_SetEntropyScore_Success(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeEntropy)

	err := analysis.SetEntropyScore(0.68)

	require.NoError(t, err)
	assert.Equal(t, 0.68, analysis.EntropyScore)
}

func TestSecurityAnalysis_SetEntropyScore_InvalidScore(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeEntropy)

	err := analysis.SetEntropyScore(-0.2)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidScore)
}

func TestSecurityAnalysis_SetDetectionProbability_Success(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeFull)

	err := analysis.SetDetectionProbability(0.15)

	require.NoError(t, err)
	assert.Equal(t, 0.15, analysis.DetectionProbability)
}

func TestSecurityAnalysis_SetDetectionProbability_Invalid(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeFull)

	err := analysis.SetDetectionProbability(2.0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidProbability)
}

func TestSecurityAnalysis_Fail(t *testing.T) {
	analysis, _ := NewSecurityAnalysis("asset-123", AnalysisTypeFull)

	analysis.Fail()

	assert.Equal(t, StatusFailed, analysis.Status)
	assert.False(t, analysis.CompletedAt.IsZero())
}

func TestSecurityAnalysis_Validate_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupAnalysis func() *SecurityAnalysis
		expectedErr   error
	}{
		{
			name: "zero_analysis_id",
			setupAnalysis: func() *SecurityAnalysis {
				a, _ := NewSecurityAnalysis("asset-123", AnalysisTypeChiSquare)
				a.ID = AnalysisID{}
				return a
			},
			expectedErr: ErrInvalidAnalysisID,
		},
		{
			name: "empty_target_asset",
			setupAnalysis: func() *SecurityAnalysis {
				a, _ := NewSecurityAnalysis("asset-123", AnalysisTypeChiSquare)
				a.TargetAssetID = ""
				return a
			},
			expectedErr: ErrEmptyTargetAsset,
		},
		{
			name: "invalid_analysis_type",
			setupAnalysis: func() *SecurityAnalysis {
				a, _ := NewSecurityAnalysis("asset-123", AnalysisTypeChiSquare)
				a.AnalysisType = "invalid_type"
				return a
			},
			expectedErr: ErrInvalidAnalysisType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := tt.setupAnalysis()

			err := analysis.Validate()

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestThreatAssessment_SetDescription(t *testing.T) {
	analysisID, _ := NewAnalysisID()
	assessment, _ := NewThreatAssessment(analysisID, VectorEntropyAnomaly, SeverityHigh)

	assessment.SetDescription("Detected unusual entropy patterns in image data")

	assert.Equal(t, "Detected unusual entropy patterns in image data", assessment.Description)
}

func TestThreatAssessment_SetMitigation(t *testing.T) {
	analysisID, _ := NewAnalysisID()
	assessment, _ := NewThreatAssessment(analysisID, VectorVisualPattern, SeverityMedium)

	assessment.SetMitigation("Apply additional noise to visual patterns")

	assert.Equal(t, "Apply additional noise to visual patterns", assessment.Mitigation)
}

func TestThreatAssessment_Validate_Errors(t *testing.T) {
	tests := []struct {
		name            string
		setupAssessment func() *ThreatAssessment
		expectedErr     error
	}{
		{
			name: "zero_assessment_id",
			setupAssessment: func() *ThreatAssessment {
				analysisID, _ := NewAnalysisID()
				a, _ := NewThreatAssessment(analysisID, VectorChiSquare, SeverityHigh)
				a.ID = AssessmentID{}
				return a
			},
			expectedErr: ErrInvalidAssessmentID,
		},
		{
			name: "zero_analysis_id",
			setupAssessment: func() *ThreatAssessment {
				analysisID, _ := NewAnalysisID()
				a, _ := NewThreatAssessment(analysisID, VectorChiSquare, SeverityHigh)
				a.AnalysisID = AnalysisID{}
				return a
			},
			expectedErr: ErrInvalidAnalysisID,
		},
		{
			name: "invalid_threat_vector",
			setupAssessment: func() *ThreatAssessment {
				analysisID, _ := NewAnalysisID()
				a, _ := NewThreatAssessment(analysisID, VectorChiSquare, SeverityHigh)
				a.ThreatVector = "invalid_vector"
				return a
			},
			expectedErr: ErrInvalidThreatVector,
		},
		{
			name: "invalid_severity",
			setupAssessment: func() *ThreatAssessment {
				analysisID, _ := NewAnalysisID()
				a, _ := NewThreatAssessment(analysisID, VectorChiSquare, SeverityHigh)
				a.Severity = "invalid_severity"
				return a
			},
			expectedErr: ErrInvalidSeverity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessment := tt.setupAssessment()

			err := assessment.Validate()

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestThreatAssessment_NewThreatAssessment_InvalidVector(t *testing.T) {
	analysisID, _ := NewAnalysisID()

	assessment, err := NewThreatAssessment(analysisID, "invalid_vector", SeverityHigh)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidThreatVector)
	assert.Nil(t, assessment)
}

func TestThreatAssessment_NewThreatAssessment_InvalidSeverity(t *testing.T) {
	analysisID, _ := NewAnalysisID()

	assessment, err := NewThreatAssessment(analysisID, VectorChiSquare, "invalid_severity")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSeverity)
	assert.Nil(t, assessment)
}

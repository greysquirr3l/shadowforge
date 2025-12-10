package analysis

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================
// AnalysisID Tests
// =====================================================

func TestAnalysisID_NewAnalysisID_Success(t *testing.T) {
	id, err := NewAnalysisID()

	require.NoError(t, err)
	assert.False(t, id.IsZero())
	assert.NotEmpty(t, id.String())

	// Verify it's a valid UUID
	_, err = uuid.Parse(id.String())
	assert.NoError(t, err)
}

func TestAnalysisID_NewAnalysisIDFromString_Success(t *testing.T) {
	validUUID := uuid.New().String()

	id, err := NewAnalysisIDFromString(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, id.String())
	assert.False(t, id.IsZero())
}

func TestAnalysisID_NewAnalysisIDFromString_EmptyString(t *testing.T) {
	id, err := NewAnalysisIDFromString("")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAnalysisID)
	assert.True(t, id.IsZero())
}

func TestAnalysisID_NewAnalysisIDFromString_InvalidUUID(t *testing.T) {
	id, err := NewAnalysisIDFromString("not-a-valid-uuid")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAnalysisID)
	assert.True(t, id.IsZero())
}

func TestAnalysisID_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		id       AnalysisID
		expected bool
	}{
		{"zero_value", AnalysisID{}, true},
		{"empty_string", AnalysisID{value: ""}, true},
		{"with_value", AnalysisID{value: uuid.New().String()}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.IsZero())
		})
	}
}

// =====================================================
// AssessmentID Tests
// =====================================================

func TestAssessmentID_NewAssessmentID_Success(t *testing.T) {
	id, err := NewAssessmentID()

	require.NoError(t, err)
	assert.False(t, id.IsZero())
	assert.NotEmpty(t, id.String())

	// Verify it's a valid UUID
	_, err = uuid.Parse(id.String())
	assert.NoError(t, err)
}

func TestAssessmentID_NewAssessmentIDFromString_Success(t *testing.T) {
	validUUID := uuid.New().String()

	id, err := NewAssessmentIDFromString(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, id.String())
	assert.False(t, id.IsZero())
}

func TestAssessmentID_NewAssessmentIDFromString_EmptyString(t *testing.T) {
	id, err := NewAssessmentIDFromString("")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAssessmentID)
	assert.True(t, id.IsZero())
}

func TestAssessmentID_NewAssessmentIDFromString_InvalidUUID(t *testing.T) {
	id, err := NewAssessmentIDFromString("not-a-valid-uuid")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAssessmentID)
	assert.True(t, id.IsZero())
}

func TestAssessmentID_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		id       AssessmentID
		expected bool
	}{
		{"zero_value", AssessmentID{}, true},
		{"empty_string", AssessmentID{value: ""}, true},
		{"with_value", AssessmentID{value: uuid.New().String()}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.IsZero())
		})
	}
}

// =====================================================
// AnalysisType Tests
// =====================================================

func TestAnalysisType_IsValid(t *testing.T) {
	tests := []struct {
		name         string
		analysisType AnalysisType
		expected     bool
	}{
		{"chi_square", AnalysisTypeChiSquare, true},
		{"rs_analysis", AnalysisTypeRS, true},
		{"entropy", AnalysisTypeEntropy, true},
		{"visual", AnalysisTypeVisual, true},
		{"histogram", AnalysisTypeHistogram, true},
		{"full", AnalysisTypeFull, true},
		{"invalid", AnalysisType("invalid"), false},
		{"empty", AnalysisType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.analysisType.IsValid())
		})
	}
}

func TestAnalysisType_String(t *testing.T) {
	tests := []struct {
		name         string
		analysisType AnalysisType
		expected     string
	}{
		{"chi_square", AnalysisTypeChiSquare, "chi_square"},
		{"rs_analysis", AnalysisTypeRS, "rs_analysis"},
		{"entropy", AnalysisTypeEntropy, "entropy"},
		{"visual", AnalysisTypeVisual, "visual"},
		{"histogram", AnalysisTypeHistogram, "histogram"},
		{"full", AnalysisTypeFull, "full"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.analysisType.String())
		})
	}
}

// =====================================================
// ThreatLevel Tests
// =====================================================

func TestThreatLevel_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		threatLevel ThreatLevel
		expected    bool
	}{
		{"none", ThreatLevelNone, true},
		{"low", ThreatLevelLow, true},
		{"medium", ThreatLevelMedium, true},
		{"high", ThreatLevelHigh, true},
		{"critical", ThreatLevelCritical, true},
		{"invalid", ThreatLevel("invalid"), false},
		{"empty", ThreatLevel(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.threatLevel.IsValid())
		})
	}
}

func TestThreatLevel_RequiresAction(t *testing.T) {
	tests := []struct {
		name        string
		threatLevel ThreatLevel
		expected    bool
	}{
		{"none_no_action", ThreatLevelNone, false},
		{"low_no_action", ThreatLevelLow, false},
		{"medium_no_action", ThreatLevelMedium, false},
		{"high_requires_action", ThreatLevelHigh, true},
		{"critical_requires_action", ThreatLevelCritical, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.threatLevel.RequiresAction())
		})
	}
}

func TestThreatLevel_String(t *testing.T) {
	tests := []struct {
		name        string
		threatLevel ThreatLevel
		expected    string
	}{
		{"none", ThreatLevelNone, "none"},
		{"low", ThreatLevelLow, "low"},
		{"medium", ThreatLevelMedium, "medium"},
		{"high", ThreatLevelHigh, "high"},
		{"critical", ThreatLevelCritical, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.threatLevel.String())
		})
	}
}

// =====================================================
// ThreatVector Tests
// =====================================================

func TestThreatVector_IsValid(t *testing.T) {
	tests := []struct {
		name         string
		threatVector ThreatVector
		expected     bool
	}{
		{"chi_square", VectorChiSquare, true},
		{"rs_signature", VectorRSSignature, true},
		{"entropy_anomaly", VectorEntropyAnomaly, true},
		{"visual_pattern", VectorVisualPattern, true},
		{"histogram_shift", VectorHistogramShift, true},
		{"lsb_patterns", VectorLSBPatterns, true},
		{"invalid", ThreatVector("invalid"), false},
		{"empty", ThreatVector(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.threatVector.IsValid())
		})
	}
}

func TestThreatVector_String(t *testing.T) {
	tests := []struct {
		name         string
		threatVector ThreatVector
		expected     string
	}{
		{"chi_square", VectorChiSquare, "chi_square_anomaly"},
		{"rs_signature", VectorRSSignature, "rs_signature"},
		{"entropy_anomaly", VectorEntropyAnomaly, "entropy_anomaly"},
		{"visual_pattern", VectorVisualPattern, "visual_pattern"},
		{"histogram_shift", VectorHistogramShift, "histogram_shift"},
		{"lsb_patterns", VectorLSBPatterns, "lsb_patterns"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.threatVector.String())
		})
	}
}

// =====================================================
// Severity Tests
// =====================================================

func TestSeverity_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		expected bool
	}{
		{"info", SeverityInfo, true},
		{"low", SeverityLow, true},
		{"medium", SeverityMedium, true},
		{"high", SeverityHigh, true},
		{"critical", SeverityCritical, true},
		{"invalid", Severity("invalid"), false},
		{"empty", Severity(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.severity.IsValid())
		})
	}
}

func TestSeverity_IsGreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		other    Severity
		expected bool
	}{
		{"info_not_greater_than_info", SeverityInfo, SeverityInfo, false},
		{"low_greater_than_info", SeverityLow, SeverityInfo, true},
		{"low_not_greater_than_low", SeverityLow, SeverityLow, false},
		{"medium_greater_than_low", SeverityMedium, SeverityLow, true},
		{"medium_not_greater_than_medium", SeverityMedium, SeverityMedium, false},
		{"high_greater_than_medium", SeverityHigh, SeverityMedium, true},
		{"critical_greater_than_high", SeverityCritical, SeverityHigh, true},
		{"critical_greater_than_all", SeverityCritical, SeverityInfo, true},
		{"info_not_greater_than_critical", SeverityInfo, SeverityCritical, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.severity.IsGreaterThan(tt.other))
		})
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		expected string
	}{
		{"info", SeverityInfo, "info"},
		{"low", SeverityLow, "low"},
		{"medium", SeverityMedium, "medium"},
		{"high", SeverityHigh, "high"},
		{"critical", SeverityCritical, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.severity.String())
		})
	}
}

// =====================================================
// AnalysisStatus Tests
// =====================================================

func TestAnalysisStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   AnalysisStatus
		expected bool
	}{
		{"pending", StatusPending, true},
		{"running", StatusRunning, true},
		{"complete", StatusComplete, true},
		{"failed", StatusFailed, true},
		{"invalid", AnalysisStatus("invalid"), false},
		{"empty", AnalysisStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestAnalysisStatus_IsFinal(t *testing.T) {
	tests := []struct {
		name     string
		status   AnalysisStatus
		expected bool
	}{
		{"pending_not_final", StatusPending, false},
		{"running_not_final", StatusRunning, false},
		{"complete_is_final", StatusComplete, true},
		{"failed_is_final", StatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsFinal())
		})
	}
}

func TestAnalysisStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   AnalysisStatus
		expected string
	}{
		{"pending", StatusPending, "pending"},
		{"running", StatusRunning, "running"},
		{"complete", StatusComplete, "complete"},
		{"failed", StatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

package chaining

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChain_Validate_Sequential tests sequential chain validation
func TestChain_Validate_Sequential(t *testing.T) {
	tests := []struct {
		name    string
		chain   *Chain
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_sequential_chain",
			chain: &Chain{
				ID:   "test-seq-1",
				Mode: ChainModeSequential,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb", Order: 0},
					{TechniqueID: "dct", Technique: "dct", Order: 1},
					{TechniqueID: "phase", Technique: "phase", Order: 2},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid_non_consecutive_order",
			chain: &Chain{
				ID:   "test-seq-2",
				Mode: ChainModeSequential,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb", Order: 0},
					{TechniqueID: "dct", Technique: "dct", Order: 2}, // Gap!
				},
			},
			wantErr: true,
			errMsg:  "consecutive",
		},
		{
			name: "missing_technique",
			chain: &Chain{
				ID:   "test-seq-3",
				Mode: ChainModeSequential,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "", Order: 0},
					{TechniqueID: "dct", Technique: "dct", Order: 1},
				},
			},
			wantErr: true,
			errMsg:  "missing technique",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.chain.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestChain_Validate_Layered tests layered chain validation
func TestChain_Validate_Layered(t *testing.T) {
	tests := []struct {
		name    string
		chain   *Chain
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_layered_chain",
			chain: &Chain{
				ID:   "test-layered-1",
				Mode: ChainModeLayered,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb", Weight: 0.6},
					{TechniqueID: "dct", Technique: "dct", Weight: 0.4},
				},
			},
			wantErr: false,
		},
		{
			name: "weights_sum_to_one",
			chain: &Chain{
				ID:   "test-layered-2",
				Mode: ChainModeLayered,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb", Weight: 0.5},
					{TechniqueID: "dct", Technique: "dct", Weight: 0.3},
					{TechniqueID: "phase", Technique: "phase", Weight: 0.2},
				},
			},
			wantErr: false,
		},
		{
			name: "weights_dont_sum_to_one",
			chain: &Chain{
				ID:   "test-layered-3",
				Mode: ChainModeLayered,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb", Weight: 0.6},
					{TechniqueID: "dct", Technique: "dct", Weight: 0.6},
				},
			},
			wantErr: true,
			errMsg:  "sum to 1.0",
		},
		{
			name: "invalid_weight_zero",
			chain: &Chain{
				ID:   "test-layered-4",
				Mode: ChainModeLayered,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb", Weight: 0.0},
					{TechniqueID: "dct", Technique: "dct", Weight: 1.0},
				},
			},
			wantErr: true,
			errMsg:  "invalid weight",
		},
		{
			name: "invalid_weight_above_one",
			chain: &Chain{
				ID:   "test-layered-5",
				Mode: ChainModeLayered,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb", Weight: 1.5},
					{TechniqueID: "dct", Technique: "dct", Weight: 0.5},
				},
			},
			wantErr: true,
			errMsg:  "invalid weight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.chain.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestChain_Validate_Split tests split chain validation
func TestChain_Validate_Split(t *testing.T) {
	tests := []struct {
		name    string
		chain   *Chain
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_split_chain",
			chain: &Chain{
				ID:   "test-split-1",
				Mode: ChainModeSplit,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb"},
					{TechniqueID: "dct", Technique: "dct"},
					{TechniqueID: "phase", Technique: "phase"},
				},
			},
			wantErr: false,
		},
		{
			name: "only_one_technique",
			chain: &Chain{
				ID:   "test-split-2",
				Mode: ChainModeSplit,
				Links: []ChainLink{
					{TechniqueID: "lsb", Technique: "lsb"},
				},
			},
			wantErr: true,
			errMsg:  "at least 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.chain.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestChain_GetTotalLinks tests link counting
func TestChain_GetTotalLinks(t *testing.T) {
	chain := &Chain{
		ID:   "test-1",
		Mode: ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Order: 0},
			{TechniqueID: "dct", Technique: "dct", Order: 1},
			{TechniqueID: "phase", Technique: "phase", Order: 2},
		},
	}

	assert.Equal(t, 3, chain.GetTotalLinks())
}

// TestChain_GetTechniqueByOrder tests technique lookup by order
func TestChain_GetTechniqueByOrder(t *testing.T) {
	chain := &Chain{
		ID:   "test-1",
		Mode: ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Order: 0},
			{TechniqueID: "dct", Technique: "dct", Order: 1},
			{TechniqueID: "phase", Technique: "phase", Order: 2},
		},
	}

	t.Run("existing_order", func(t *testing.T) {
		link, err := chain.GetTechniqueByOrder(1)
		require.NoError(t, err)
		assert.Equal(t, "dct", string(link.Technique))
	})

	t.Run("non_existing_order", func(t *testing.T) {
		link, err := chain.GetTechniqueByOrder(99)
		assert.Error(t, err)
		assert.Nil(t, link)
	})
}

// TestChainExecutionResult_Success tests execution result tracking
func TestChainExecutionResult_Success(t *testing.T) {
	result := &ChainExecutionResult{
		ChainID:       "test-1",
		Mode:          ChainModeSequential,
		StepResults:   make([]StepResult, 0),
		FinalOutput:   []byte("output"),
		TotalDuration: 100 * time.Millisecond,
		Success:       true,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "test-1", result.ChainID)
	assert.Equal(t, ChainModeSequential, result.Mode)
	assert.NotEmpty(t, result.FinalOutput)
}

// TestStepResult_Tracking tests step result tracking
func TestStepResult_Tracking(t *testing.T) {
	step := StepResult{
		LinkOrder:    0,
		Technique:    "lsb",
		InputSize:    1024,
		OutputSize:   2048,
		Duration:     50 * time.Millisecond,
		Success:      true,
		ErrorMessage: "",
	}

	assert.Equal(t, 0, step.LinkOrder)
	assert.Equal(t, StegoTechnique("lsb"), step.Technique)
	assert.Equal(t, int64(1024), step.InputSize)
	assert.Equal(t, int64(2048), step.OutputSize)
	assert.True(t, step.Success)
}

// TestChain_Validate_EmptyID tests validation with empty ID
func TestChain_Validate_EmptyID(t *testing.T) {
	chain := &Chain{
		ID:   "",
		Mode: ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Order: 0},
		},
	}

	err := chain.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID is required")
}

// TestChain_Validate_TooFewLinks tests validation with insufficient links
func TestChain_Validate_TooFewLinks(t *testing.T) {
	chain := &Chain{
		ID:    "test-1",
		Mode:  ChainModeSequential,
		Links: []ChainLink{}, // Empty
	}

	err := chain.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2")
}

// TestChain_Validate_InvalidMode tests validation with invalid mode
func TestChain_Validate_InvalidMode(t *testing.T) {
	chain := &Chain{
		ID:   "test-1",
		Mode: ChainMode("invalid"),
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Order: 0},
			{TechniqueID: "dct", Technique: "dct", Order: 1},
		},
	}

	err := chain.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid chain mode")
}

// Benchmark tests

// BenchmarkChain_Validate_Sequential benchmarks sequential chain validation
func BenchmarkChain_Validate_Sequential(b *testing.B) {
	chain := &Chain{
		ID:   "bench-seq",
		Mode: ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Order: 0},
			{TechniqueID: "dct", Technique: "dct", Order: 1},
			{TechniqueID: "phase", Technique: "phase", Order: 2},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chain.Validate()
	}
}

// BenchmarkChain_Validate_Layered benchmarks layered chain validation
func BenchmarkChain_Validate_Layered(b *testing.B) {
	chain := &Chain{
		ID:   "bench-layered",
		Mode: ChainModeLayered,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Weight: 0.5},
			{TechniqueID: "dct", Technique: "dct", Weight: 0.3},
			{TechniqueID: "phase", Technique: "phase", Weight: 0.2},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chain.Validate()
	}
}

// BenchmarkChain_GetTechniqueByOrder benchmarks technique lookup
func BenchmarkChain_GetTechniqueByOrder(b *testing.B) {
	chain := &Chain{
		ID:   "bench-lookup",
		Mode: ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "lsb", Technique: "lsb", Order: 0},
			{TechniqueID: "dct", Technique: "dct", Order: 1},
			{TechniqueID: "phase", Technique: "phase", Order: 2},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chain.GetTechniqueByOrder(1)
	}
}

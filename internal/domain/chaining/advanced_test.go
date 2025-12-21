package chaining

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// Test Conditional Branching

func TestBranchCondition_Evaluate(t *testing.T) {
	tests := []struct {
		name        string
		condition   BranchCondition
		payloadSize int64
		expected    bool
	}{
		{
			name: "size_greater_than_true",
			condition: BranchCondition{
				Type:     ConditionPayloadSize,
				Operator: OperatorGreaterThan,
				Value:    float64(10000),
			},
			payloadSize: 15000,
			expected:    true,
		},
		{
			name: "size_greater_than_false",
			condition: BranchCondition{
				Type:     ConditionPayloadSize,
				Operator: OperatorGreaterThan,
				Value:    float64(10000),
			},
			payloadSize: 5000,
			expected:    false,
		},
		{
			name: "size_less_than_true",
			condition: BranchCondition{
				Type:     ConditionPayloadSize,
				Operator: OperatorLessThan,
				Value:    float64(10000),
			},
			payloadSize: 5000,
			expected:    true,
		},
		{
			name: "size_less_than_false",
			condition: BranchCondition{
				Type:     ConditionPayloadSize,
				Operator: OperatorLessThan,
				Value:    float64(10000),
			},
			payloadSize: 15000,
			expected:    false,
		},
		{
			name: "size_equals_true",
			condition: BranchCondition{
				Type:     ConditionPayloadSize,
				Operator: OperatorEquals,
				Value:    float64(10000),
			},
			payloadSize: 10000,
			expected:    true,
		},
		{
			name: "size_equals_false",
			condition: BranchCondition{
				Type:     ConditionPayloadSize,
				Operator: OperatorEquals,
				Value:    float64(10000),
			},
			payloadSize: 10001,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			carrier := []byte("test carrier data")

			result, err := tt.condition.Evaluate(ctx, carrier, tt.payloadSize)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionalBranch_Create(t *testing.T) {
	trueChain := &Chain{
		ID:   "true_chain",
		Mode: ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "lsb_large", Technique: stego.LSB, Order: 1},
		},
	}

	falseChain := &Chain{
		ID:   "false_chain",
		Mode: ChainModeSequential,
		Links: []ChainLink{
			{TechniqueID: "dct_small", Technique: stego.DCT, Order: 1},
		},
	}

	condition := BranchCondition{
		Type:      ConditionPayloadSize,
		Operator:  OperatorGreaterThan,
		Value:     float64(10000),
		Parameter: "size",
	}

	branch := &ConditionalBranch{
		ID:          "test_branch",
		Condition:   condition,
		TrueChain:   trueChain,
		FalseChain:  falseChain,
		Description: "Test conditional branch",
		CreatedAt:   time.Now(),
	}

	assert.NotNil(t, branch)
	assert.Equal(t, "test_branch", branch.ID)
	assert.NotNil(t, branch.TrueChain)
	assert.NotNil(t, branch.FalseChain)
	assert.Equal(t, ConditionPayloadSize, branch.Condition.Type)
}

// Test Adaptive Weights

func TestAdaptiveWeightCalculator_CalculateWeights_EqualWeights(t *testing.T) {
	calc := &AdaptiveWeightCalculator{
		MinWeight:     0.1,
		MaxWeight:     0.9,
		CapacityBased: false,
		QualityBased:  false,
	}

	ctx := context.Background()
	carriers := [][]byte{
		[]byte("carrier1"),
		[]byte("carrier2"),
		[]byte("carrier3"),
	}
	techniques := []StegoTechnique{stego.LSB, stego.DCT, stego.PhaseEncoding}
	capacities := []int64{1000, 2000, 3000}
	qualities := []float64{0.8, 0.9, 0.7}

	weights, err := calc.CalculateWeights(ctx, carriers, techniques, capacities, qualities)
	require.NoError(t, err)
	require.Len(t, weights, 3)

	// Equal weights should be approximately 1/3 each
	expectedWeight := 1.0 / 3.0
	for i, w := range weights {
		assert.InDelta(t, expectedWeight, w, 0.01, "Weight %d should be approximately equal", i)
	}

	// Weights should sum to 1.0
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	assert.InDelta(t, 1.0, sum, 0.01, "Weights should sum to 1.0")
}

func TestAdaptiveWeightCalculator_CalculateWeights_CapacityBased(t *testing.T) {
	calc := &AdaptiveWeightCalculator{
		MinWeight:     0.1,
		MaxWeight:     0.9,
		CapacityBased: true,
		QualityBased:  false,
	}

	ctx := context.Background()
	carriers := [][]byte{
		[]byte("carrier1"),
		[]byte("carrier2"),
		[]byte("carrier3"),
	}
	techniques := []StegoTechnique{stego.LSB, stego.DCT, stego.PhaseEncoding}
	capacities := []int64{1000, 2000, 3000} // Total 6000
	qualities := []float64{0.8, 0.9, 0.7}

	weights, err := calc.CalculateWeights(ctx, carriers, techniques, capacities, qualities)
	require.NoError(t, err)
	require.Len(t, weights, 3)

	// Weights should be proportional to capacity
	// carrier1: 1000/6000 ≈ 0.167
	// carrier2: 2000/6000 ≈ 0.333
	// carrier3: 3000/6000 ≈ 0.500
	assert.InDelta(t, 0.167, weights[0], 0.02, "Capacity weight 1")
	assert.InDelta(t, 0.333, weights[1], 0.02, "Capacity weight 2")
	assert.InDelta(t, 0.500, weights[2], 0.02, "Capacity weight 3")

	// Verify sum equals 1.0
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	assert.InDelta(t, 1.0, sum, 0.01, "Weights should sum to 1.0")
}

func TestAdaptiveWeightCalculator_CalculateWeights_QualityBased(t *testing.T) {
	calc := &AdaptiveWeightCalculator{
		MinWeight:     0.1,
		MaxWeight:     0.9,
		CapacityBased: false,
		QualityBased:  true,
	}

	ctx := context.Background()
	carriers := [][]byte{
		[]byte("carrier1"),
		[]byte("carrier2"),
		[]byte("carrier3"),
	}
	techniques := []StegoTechnique{stego.LSB, stego.DCT, stego.PhaseEncoding}
	capacities := []int64{1000, 2000, 3000}
	qualities := []float64{0.3, 0.5, 0.7} // Total 1.5

	weights, err := calc.CalculateWeights(ctx, carriers, techniques, capacities, qualities)
	require.NoError(t, err)
	require.Len(t, weights, 3)

	// Weights should be proportional to quality
	// carrier1: 0.3/1.5 = 0.2
	// carrier2: 0.5/1.5 ≈ 0.333
	// carrier3: 0.7/1.5 ≈ 0.467
	assert.InDelta(t, 0.2, weights[0], 0.02, "Quality weight 1")
	assert.InDelta(t, 0.333, weights[1], 0.02, "Quality weight 2")
	assert.InDelta(t, 0.467, weights[2], 0.02, "Quality weight 3")

	// Verify sum equals 1.0
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	assert.InDelta(t, 1.0, sum, 0.01, "Weights should sum to 1.0")
}

func TestAdaptiveWeightCalculator_WeightClamping(t *testing.T) {
	// Test min/max weight clamping behavior
	calc := &AdaptiveWeightCalculator{
		MinWeight:     0.2,
		MaxWeight:     0.6,
		CapacityBased: true,
		QualityBased:  false,
	}

	ctx := context.Background()
	carriers := [][]byte{
		[]byte("carrier1"),
		[]byte("carrier2"),
		[]byte("carrier3"),
	}
	techniques := []StegoTechnique{stego.LSB, stego.DCT, stego.PhaseEncoding}
	capacities := []int64{100, 5000, 10000} // Very uneven distribution
	qualities := []float64{0.8, 0.9, 0.7}

	weights, err := calc.CalculateWeights(ctx, carriers, techniques, capacities, qualities)
	require.NoError(t, err)
	require.Len(t, weights, 3)

	// After clamping and renormalization, weights should sum to 1.0
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	assert.InDelta(t, 1.0, sum, 0.001, "Weights should sum to 1.0 after clamping")

	// No weight should exceed max (even after renorm, due to re-normalization logic)
	for i, w := range weights {
		assert.LessOrEqual(t, w, calc.MaxWeight+0.01, "Weight %d above maximum", i)
	}

	// The largest capacity (10000) should get a significant weight
	assert.Greater(t, weights[2], weights[0], "Largest capacity should get higher weight")
	assert.Greater(t, weights[2], weights[1], "Largest capacity should get higher weight")
}

// Test Parallel Execution

func TestParallelExecutor_Create(t *testing.T) {
	executor := &ParallelExecutor{
		MaxWorkers:    4,
		ResultChannel: make(chan *StepResult, 10),
		ErrorChannel:  make(chan error, 10),
	}

	assert.NotNil(t, executor)
	assert.Equal(t, 4, executor.MaxWorkers)
	assert.NotNil(t, executor.ResultChannel)
	assert.NotNil(t, executor.ErrorChannel)
}

// Note: Full parallel execution testing requires mock stego/media services
// which we're avoiding in unit tests. See E2E tests for integration testing.

// Benchmarks

func BenchmarkBranchCondition_Evaluate(b *testing.B) {
	condition := BranchCondition{
		Type:     ConditionPayloadSize,
		Operator: OperatorGreaterThan,
		Value:    float64(10000),
	}

	ctx := context.Background()
	carrier := []byte("test carrier data")
	payloadSize := int64(15000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = condition.Evaluate(ctx, carrier, payloadSize)
	}
}

func BenchmarkAdaptiveWeights_CalculateCapacityBased(b *testing.B) {
	calc := &AdaptiveWeightCalculator{
		MinWeight:     0.1,
		MaxWeight:     0.9,
		CapacityBased: true,
	}

	ctx := context.Background()
	carriers := [][]byte{
		[]byte("carrier1"),
		[]byte("carrier2"),
		[]byte("carrier3"),
	}
	techniques := []StegoTechnique{stego.LSB, stego.DCT, stego.PhaseEncoding}
	capacities := []int64{1000, 2000, 3000}
	qualities := []float64{0.8, 0.9, 0.7}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calc.CalculateWeights(ctx, carriers, techniques, capacities, qualities)
	}
}

func BenchmarkAdaptiveWeights_CalculateQualityBased(b *testing.B) {
	calc := &AdaptiveWeightCalculator{
		MinWeight:    0.1,
		MaxWeight:    0.9,
		QualityBased: true,
	}

	ctx := context.Background()
	carriers := [][]byte{
		[]byte("carrier1"),
		[]byte("carrier2"),
		[]byte("carrier3"),
	}
	techniques := []StegoTechnique{stego.LSB, stego.DCT, stego.PhaseEncoding}
	capacities := []int64{1000, 2000, 3000}
	qualities := []float64{0.8, 0.9, 0.7}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calc.CalculateWeights(ctx, carriers, techniques, capacities, qualities)
	}
}

package errorcorrection_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
)

// ============================================================================
// MessageID Tests
// ============================================================================

func TestMessageID_NewMessageID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantErr  bool
		errCheck error
	}{
		{
			name:    "valid_message_id",
			id:      "msg-12345",
			wantErr: false,
		},
		{
			name:    "valid_uuid",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:     "empty_id",
			id:       "",
			wantErr:  true,
			errCheck: errorcorrection.ErrInvalidMessageID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			msgID, err := errorcorrection.NewMessageID(tt.id)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
				assert.True(t, msgID.IsZero())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.id, msgID.String())
				assert.False(t, msgID.IsZero())
			}
		})
	}
}

func TestMessageID_GenerateMessageID(t *testing.T) {
	t.Run("generates_unique_ids", func(t *testing.T) {
		// Act
		id1 := errorcorrection.GenerateMessageID()
		id2 := errorcorrection.GenerateMessageID()

		// Assert
		assert.NotEqual(t, id1.String(), id2.String())
		assert.False(t, id1.IsZero())
		assert.False(t, id2.IsZero())
		assert.NotEmpty(t, id1.String())
		assert.NotEmpty(t, id2.String())

		// UUID format validation (8-4-4-4-12)
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, id1.String())
	})
}

func TestMessageID_String(t *testing.T) {
	t.Run("returns_string_representation", func(t *testing.T) {
		// Arrange
		expectedID := "test-message-id-123"
		msgID, err := errorcorrection.NewMessageID(expectedID)
		require.NoError(t, err)

		// Act
		result := msgID.String()

		// Assert
		assert.Equal(t, expectedID, result)
	})
}

func TestMessageID_IsZero(t *testing.T) {
	tests := []struct {
		name      string
		messageID func() errorcorrection.MessageID
		wantZero  bool
	}{
		{
			name: "non_zero_message_id",
			messageID: func() errorcorrection.MessageID {
				id, _ := errorcorrection.NewMessageID("valid-id")
				return id
			},
			wantZero: false,
		},
		{
			name: "generated_message_id",
			messageID: func() errorcorrection.MessageID {
				return errorcorrection.GenerateMessageID()
			},
			wantZero: false,
		},
		{
			name: "zero_value_message_id",
			messageID: func() errorcorrection.MessageID {
				return errorcorrection.MessageID{}
			},
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			msgID := tt.messageID()

			// Act
			isZero := msgID.IsZero()

			// Assert
			assert.Equal(t, tt.wantZero, isZero)
		})
	}
}

func TestMessageID_Equals(t *testing.T) {
	tests := []struct {
		name   string
		id1    errorcorrection.MessageID
		id2    errorcorrection.MessageID
		equals bool
	}{
		{
			name:   "equal_message_ids",
			id1:    errorcorrection.GenerateMessageID(),
			id2:    errorcorrection.GenerateMessageID(),
			equals: false, // Different generated IDs
		},
		{
			name: "same_string_value",
			id1: func() errorcorrection.MessageID {
				id, _ := errorcorrection.NewMessageID("same-id")
				return id
			}(),
			id2: func() errorcorrection.MessageID {
				id, _ := errorcorrection.NewMessageID("same-id")
				return id
			}(),
			equals: true,
		},
		{
			name: "different_string_values",
			id1: func() errorcorrection.MessageID {
				id, _ := errorcorrection.NewMessageID("id-1")
				return id
			}(),
			id2: func() errorcorrection.MessageID {
				id, _ := errorcorrection.NewMessageID("id-2")
				return id
			}(),
			equals: false,
		},
		{
			name:   "both_zero",
			id1:    errorcorrection.MessageID{},
			id2:    errorcorrection.MessageID{},
			equals: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.id1.Equals(tt.id2)

			// Assert
			assert.Equal(t, tt.equals, result)
		})
	}
}

// ============================================================================
// ShardID Tests
// ============================================================================

func TestShardID_NewShardID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantErr  bool
		errCheck error
	}{
		{
			name:    "valid_shard_id",
			id:      "shard-12345",
			wantErr: false,
		},
		{
			name:    "valid_uuid",
			id:      "650e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:     "empty_id",
			id:       "",
			wantErr:  true,
			errCheck: errorcorrection.ErrInvalidShardID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			shardID, err := errorcorrection.NewShardID(tt.id)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
				assert.True(t, shardID.IsZero())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.id, shardID.String())
				assert.False(t, shardID.IsZero())
			}
		})
	}
}

func TestShardID_GenerateShardID(t *testing.T) {
	t.Run("generates_unique_ids", func(t *testing.T) {
		// Act
		id1 := errorcorrection.GenerateShardID()
		id2 := errorcorrection.GenerateShardID()

		// Assert
		assert.NotEqual(t, id1.String(), id2.String())
		assert.False(t, id1.IsZero())
		assert.False(t, id2.IsZero())
		assert.NotEmpty(t, id1.String())
		assert.NotEmpty(t, id2.String())

		// UUID format validation
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, id1.String())
	})
}

func TestShardID_String(t *testing.T) {
	t.Run("returns_string_representation", func(t *testing.T) {
		// Arrange
		expectedID := "test-shard-id-456"
		shardID, err := errorcorrection.NewShardID(expectedID)
		require.NoError(t, err)

		// Act
		result := shardID.String()

		// Assert
		assert.Equal(t, expectedID, result)
	})
}

func TestShardID_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		shardID  func() errorcorrection.ShardID
		wantZero bool
	}{
		{
			name: "non_zero_shard_id",
			shardID: func() errorcorrection.ShardID {
				id, _ := errorcorrection.NewShardID("valid-id")
				return id
			},
			wantZero: false,
		},
		{
			name: "generated_shard_id",
			shardID: func() errorcorrection.ShardID {
				return errorcorrection.GenerateShardID()
			},
			wantZero: false,
		},
		{
			name: "zero_value_shard_id",
			shardID: func() errorcorrection.ShardID {
				return errorcorrection.ShardID{}
			},
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			shardID := tt.shardID()

			// Act
			isZero := shardID.IsZero()

			// Assert
			assert.Equal(t, tt.wantZero, isZero)
		})
	}
}

func TestShardID_Equals(t *testing.T) {
	tests := []struct {
		name   string
		id1    errorcorrection.ShardID
		id2    errorcorrection.ShardID
		equals bool
	}{
		{
			name:   "different_generated_ids",
			id1:    errorcorrection.GenerateShardID(),
			id2:    errorcorrection.GenerateShardID(),
			equals: false,
		},
		{
			name: "same_string_value",
			id1: func() errorcorrection.ShardID {
				id, _ := errorcorrection.NewShardID("same-id")
				return id
			}(),
			id2: func() errorcorrection.ShardID {
				id, _ := errorcorrection.NewShardID("same-id")
				return id
			}(),
			equals: true,
		},
		{
			name: "different_string_values",
			id1: func() errorcorrection.ShardID {
				id, _ := errorcorrection.NewShardID("id-1")
				return id
			}(),
			id2: func() errorcorrection.ShardID {
				id, _ := errorcorrection.NewShardID("id-2")
				return id
			}(),
			equals: false,
		},
		{
			name:   "both_zero",
			id1:    errorcorrection.ShardID{},
			id2:    errorcorrection.ShardID{},
			equals: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.id1.Equals(tt.id2)

			// Assert
			assert.Equal(t, tt.equals, result)
		})
	}
}

// ============================================================================
// RedundancyLevel Tests
// ============================================================================

func TestRedundancyLevel_NewRedundancyLevel(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		wantErr    bool
		errCheck   error
	}{
		{
			name:       "valid_low_17_percent",
			percentage: 0.17,
			wantErr:    false,
		},
		{
			name:       "valid_balanced_33_percent",
			percentage: 0.33,
			wantErr:    false,
		},
		{
			name:       "valid_conservative_50_percent",
			percentage: 0.50,
			wantErr:    false,
		},
		{
			name:       "valid_high_75_percent",
			percentage: 0.75,
			wantErr:    false,
		},
		{
			name:       "valid_maximum_100_percent",
			percentage: 1.0,
			wantErr:    false,
		},
		{
			name:       "invalid_zero",
			percentage: 0.0,
			wantErr:    true,
			errCheck:   errorcorrection.ErrInvalidRedundancy,
		},
		{
			name:       "invalid_negative",
			percentage: -0.25,
			wantErr:    true,
			errCheck:   errorcorrection.ErrInvalidRedundancy,
		},
		{
			name:       "invalid_over_100_percent",
			percentage: 1.5,
			wantErr:    true,
			errCheck:   errorcorrection.ErrInvalidRedundancy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			level, err := errorcorrection.NewRedundancyLevel(tt.percentage)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.percentage, level.Percentage())
			}
		})
	}
}

func TestRedundancyLevel_PredefinedLevels(t *testing.T) {
	tests := []struct {
		name            string
		level           errorcorrection.RedundancyLevel
		expectedPercent float64
		expectedString  string
	}{
		{
			name:            "low_redundancy_17_percent",
			level:           errorcorrection.RedundancyLow,
			expectedPercent: 0.17,
			expectedString:  "17%",
		},
		{
			name:            "balanced_redundancy_33_percent",
			level:           errorcorrection.RedundancyBalanced,
			expectedPercent: 0.33,
			expectedString:  "33%",
		},
		{
			name:            "conservative_redundancy_50_percent",
			level:           errorcorrection.RedundancyConservative,
			expectedPercent: 0.50,
			expectedString:  "50%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act & Assert
			assert.Equal(t, tt.expectedPercent, tt.level.Percentage())
			assert.Equal(t, tt.expectedString, tt.level.String())
		})
	}
}

func TestRedundancyLevel_ParityShardCount(t *testing.T) {
	tests := []struct {
		name       string
		level      errorcorrection.RedundancyLevel
		dataShards int
		wantParity int
	}{
		{
			name:       "low_17_percent_with_15_data_shards",
			level:      errorcorrection.RedundancyLow,
			dataShards: 15,
			wantParity: 2, // 15 * 0.17 = 2.55 → 2
		},
		{
			name:       "balanced_33_percent_with_10_data_shards",
			level:      errorcorrection.RedundancyBalanced,
			dataShards: 10,
			wantParity: 3, // 10 * 0.33 = 3.3 → 3
		},
		{
			name:       "conservative_50_percent_with_10_data_shards",
			level:      errorcorrection.RedundancyConservative,
			dataShards: 10,
			wantParity: 5, // 10 * 0.50 = 5.0 → 5
		},
		{
			name:       "balanced_with_20_data_shards",
			level:      errorcorrection.RedundancyBalanced,
			dataShards: 20,
			wantParity: 6, // 20 * 0.33 = 6.6 → 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			parityCount := tt.level.ParityShardCount(tt.dataShards)

			// Assert
			assert.Equal(t, tt.wantParity, parityCount)
		})
	}
}

func TestRedundancyLevel_String(t *testing.T) {
	tests := []struct {
		name           string
		percentage     float64
		expectedString string
	}{
		{
			name:           "17_percent",
			percentage:     0.17,
			expectedString: "17%",
		},
		{
			name:           "33_percent",
			percentage:     0.33,
			expectedString: "33%",
		},
		{
			name:           "50_percent",
			percentage:     0.50,
			expectedString: "50%",
		},
		{
			name:           "75_percent",
			percentage:     0.75,
			expectedString: "75%",
		},
		{
			name:           "100_percent",
			percentage:     1.0,
			expectedString: "100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			level, err := errorcorrection.NewRedundancyLevel(tt.percentage)
			require.NoError(t, err)

			// Act
			result := level.String()

			// Assert
			assert.Equal(t, tt.expectedString, result)
		})
	}
}

func TestCalculateRedundancy(t *testing.T) {
	tests := []struct {
		name         string
		dataShards   int
		parityShards int
		wantPercent  float64
	}{
		{
			name:         "conservative_10_10",
			dataShards:   10,
			parityShards: 10,
			wantPercent:  1.0, // 100%
		},
		{
			name:         "balanced_10_5",
			dataShards:   10,
			parityShards: 5,
			wantPercent:  0.5, // 50%
		},
		{
			name:         "balanced_15_5",
			dataShards:   15,
			parityShards: 5,
			wantPercent:  0.3333333333333333, // 33.33%
		},
		{
			name:         "aggressive_20_3",
			dataShards:   20,
			parityShards: 3,
			wantPercent:  0.15, // 15%
		},
		{
			name:         "zero_data_shards",
			dataShards:   0,
			parityShards: 5,
			wantPercent:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			level := errorcorrection.CalculateRedundancy(tt.dataShards, tt.parityShards)

			// Assert
			assert.InDelta(t, tt.wantPercent, level.Percentage(), 0.0001)
		})
	}
}

// ============================================================================
// ShardConfiguration Tests
// ============================================================================

func TestShardConfiguration_NewShardConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		dataShards   int
		parityShards int
		wantErr      bool
		errCheck     error
	}{
		{
			name:         "valid_conservative_10_10",
			dataShards:   10,
			parityShards: 10,
			wantErr:      false,
		},
		{
			name:         "valid_balanced_15_5",
			dataShards:   15,
			parityShards: 5,
			wantErr:      false,
		},
		{
			name:         "valid_aggressive_20_3",
			dataShards:   20,
			parityShards: 3,
			wantErr:      false,
		},
		{
			name:         "zero_data_shards",
			dataShards:   0,
			parityShards: 5,
			wantErr:      true,
			errCheck:     errorcorrection.ErrInvalidShardCount,
		},
		{
			name:         "zero_parity_shards",
			dataShards:   10,
			parityShards: 0,
			wantErr:      true,
			errCheck:     errorcorrection.ErrInvalidShardCount,
		},
		{
			name:         "negative_data_shards",
			dataShards:   -5,
			parityShards: 5,
			wantErr:      true,
			errCheck:     errorcorrection.ErrInvalidShardCount,
		},
		{
			name:         "negative_parity_shards",
			dataShards:   10,
			parityShards: -3,
			wantErr:      true,
			errCheck:     errorcorrection.ErrInvalidShardCount,
		},
		{
			name:         "exceeds_max_total_shards",
			dataShards:   200,
			parityShards: 100,
			wantErr:      true,
			errCheck:     errorcorrection.ErrTooManyShards,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			config, err := errorcorrection.NewShardConfiguration(tt.dataShards, tt.parityShards)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, tt.dataShards, config.DataShards)
				assert.Equal(t, tt.parityShards, config.ParityShards)
				assert.NotZero(t, config.Redundancy.Percentage())
			}
		})
	}
}

func TestShardConfiguration_TotalShards(t *testing.T) {
	tests := []struct {
		name         string
		dataShards   int
		parityShards int
		wantTotal    int
	}{
		{
			name:         "conservative_10_10",
			dataShards:   10,
			parityShards: 10,
			wantTotal:    20,
		},
		{
			name:         "balanced_15_5",
			dataShards:   15,
			parityShards: 5,
			wantTotal:    20,
		},
		{
			name:         "aggressive_20_3",
			dataShards:   20,
			parityShards: 3,
			wantTotal:    23,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config, err := errorcorrection.NewShardConfiguration(tt.dataShards, tt.parityShards)
			require.NoError(t, err)

			// Act
			total := config.TotalShards()

			// Assert
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestShardConfiguration_MinimumRequired(t *testing.T) {
	tests := []struct {
		name         string
		dataShards   int
		parityShards int
		wantMinimum  int
	}{
		{
			name:         "conservative_10_10_needs_10",
			dataShards:   10,
			parityShards: 10,
			wantMinimum:  10,
		},
		{
			name:         "balanced_15_5_needs_15",
			dataShards:   15,
			parityShards: 5,
			wantMinimum:  15,
		},
		{
			name:         "aggressive_20_3_needs_20",
			dataShards:   20,
			parityShards: 3,
			wantMinimum:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config, err := errorcorrection.NewShardConfiguration(tt.dataShards, tt.parityShards)
			require.NoError(t, err)

			// Act
			minimum := config.MinimumRequired()

			// Assert
			assert.Equal(t, tt.wantMinimum, minimum)
			assert.Equal(t, tt.dataShards, minimum, "minimum should equal data shards")
		})
	}
}

func TestShardConfiguration_MaximumFailures(t *testing.T) {
	tests := []struct {
		name          string
		dataShards    int
		parityShards  int
		wantMaxLosses int
	}{
		{
			name:          "conservative_10_10_tolerates_10_losses",
			dataShards:    10,
			parityShards:  10,
			wantMaxLosses: 10,
		},
		{
			name:          "balanced_15_5_tolerates_5_losses",
			dataShards:    15,
			parityShards:  5,
			wantMaxLosses: 5,
		},
		{
			name:          "aggressive_20_3_tolerates_3_losses",
			dataShards:    20,
			parityShards:  3,
			wantMaxLosses: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config, err := errorcorrection.NewShardConfiguration(tt.dataShards, tt.parityShards)
			require.NoError(t, err)

			// Act
			maxLosses := config.MaximumFailures()

			// Assert
			assert.Equal(t, tt.wantMaxLosses, maxLosses)
			assert.Equal(t, tt.parityShards, maxLosses, "max failures should equal parity shards")
		})
	}
}

func TestShardConfiguration_ComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name               string
		dataShards         int
		parityShards       int
		expectedTotal      int
		expectedMinimum    int
		expectedMaxLosses  int
		expectedRedundancy string
	}{
		{
			name:               "minimal_1_1",
			dataShards:         1,
			parityShards:       1,
			expectedTotal:      2,
			expectedMinimum:    1,
			expectedMaxLosses:  1,
			expectedRedundancy: "100%",
		},
		{
			name:               "conservative_5_5",
			dataShards:         5,
			parityShards:       5,
			expectedTotal:      10,
			expectedMinimum:    5,
			expectedMaxLosses:  5,
			expectedRedundancy: "100%",
		},
		{
			name:               "balanced_10_5",
			dataShards:         10,
			parityShards:       5,
			expectedTotal:      15,
			expectedMinimum:    10,
			expectedMaxLosses:  5,
			expectedRedundancy: "50%",
		},
		{
			name:               "typical_production_20_10",
			dataShards:         20,
			parityShards:       10,
			expectedTotal:      30,
			expectedMinimum:    20,
			expectedMaxLosses:  10,
			expectedRedundancy: "50%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config, err := errorcorrection.NewShardConfiguration(tt.dataShards, tt.parityShards)
			require.NoError(t, err)

			// Act & Assert
			assert.Equal(t, tt.expectedTotal, config.TotalShards())
			assert.Equal(t, tt.expectedMinimum, config.MinimumRequired())
			assert.Equal(t, tt.expectedMaxLosses, config.MaximumFailures())

			// Redundancy string should start with expected percentage
			assert.True(t, strings.HasPrefix(config.Redundancy.String(), tt.expectedRedundancy),
				"Expected redundancy to start with %s, got %s", tt.expectedRedundancy, config.Redundancy.String())
		})
	}
}

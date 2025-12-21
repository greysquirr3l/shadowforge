package errorcorrection_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
)

// ============================================================================
// ProtectedMessage Tests
// ============================================================================

func TestProtectedMessage_NewProtectedMessage(t *testing.T) {
	tests := []struct {
		name         string
		messageID    errorcorrection.MessageID
		data         []byte
		dataShards   int
		parityShards int
		wantErr      bool
		errCheck     error
	}{
		{
			name:         "valid_protected_message",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte("test data for Reed-Solomon encoding"),
			dataShards:   10,
			parityShards: 5,
			wantErr:      false,
		},
		{
			name:         "empty_data",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte{},
			dataShards:   10,
			parityShards: 5,
			wantErr:      true,
			errCheck:     errorcorrection.ErrEmptyData,
		},
		{
			name:         "zero_data_shards",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte("test data"),
			dataShards:   0,
			parityShards: 5,
			wantErr:      true,
			errCheck:     errorcorrection.ErrInvalidShardCount,
		},
		{
			name:         "zero_parity_shards",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte("test data"),
			dataShards:   10,
			parityShards: 0,
			wantErr:      true,
			errCheck:     errorcorrection.ErrInvalidShardCount,
		},
		{
			name:         "negative_data_shards",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte("test data"),
			dataShards:   -5,
			parityShards: 5,
			wantErr:      true,
			errCheck:     errorcorrection.ErrInvalidShardCount,
		},
		{
			name:         "negative_parity_shards",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte("test data"),
			dataShards:   10,
			parityShards: -3,
			wantErr:      true,
			errCheck:     errorcorrection.ErrInvalidShardCount,
		},
		{
			name:         "conservative_redundancy_50_percent",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte("conservative protected message"),
			dataShards:   10,
			parityShards: 10,
			wantErr:      false,
		},
		{
			name:         "balanced_redundancy_33_percent",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte("balanced protected message"),
			dataShards:   15,
			parityShards: 5,
			wantErr:      false,
		},
		{
			name:         "aggressive_redundancy_17_percent",
			messageID:    errorcorrection.GenerateMessageID(),
			data:         []byte("aggressive protected message"),
			dataShards:   20,
			parityShards: 3,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			msg, err := errorcorrection.NewProtectedMessage(tt.messageID, tt.data, tt.dataShards, tt.parityShards)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
				assert.Nil(t, msg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, msg)
				assert.True(t, msg.ID.Equals(tt.messageID))
				assert.Equal(t, tt.data, msg.OriginalData)
				assert.Equal(t, tt.dataShards, msg.DataShards)
				assert.Equal(t, tt.parityShards, msg.ParityShards)
				assert.NotZero(t, msg.EncodedAt)
				assert.WithinDuration(t, time.Now(), msg.EncodedAt, time.Second)
				assert.Len(t, msg.Shards, 0)
			}
		})
	}
}

func TestProtectedMessage_AddShard(t *testing.T) {
	tests := []struct {
		name      string
		setupMsg  func() *errorcorrection.ProtectedMessage
		shard     func(msgID errorcorrection.MessageID) *errorcorrection.Shard
		wantErr   bool
		errCheck  error
		wantCount int
	}{
		{
			name: "add_valid_shard",
			setupMsg: func() *errorcorrection.ProtectedMessage {
				msgID := errorcorrection.GenerateMessageID()
				msg, _ := errorcorrection.NewProtectedMessage(msgID, []byte("test"), 10, 5)
				return msg
			},
			shard: func(msgID errorcorrection.MessageID) *errorcorrection.Shard {
				shardID := errorcorrection.GenerateShardID()
				shard, _ := errorcorrection.NewShard(shardID, msgID, 0, []byte("shard data"), false)
				return shard
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "add_nil_shard",
			setupMsg: func() *errorcorrection.ProtectedMessage {
				msgID := errorcorrection.GenerateMessageID()
				msg, _ := errorcorrection.NewProtectedMessage(msgID, []byte("test"), 10, 5)
				return msg
			},
			shard: func(msgID errorcorrection.MessageID) *errorcorrection.Shard {
				return nil
			},
			wantErr:   true,
			errCheck:  errorcorrection.ErrNilShard,
			wantCount: 0,
		},
		{
			name: "add_shard_with_mismatched_message_id",
			setupMsg: func() *errorcorrection.ProtectedMessage {
				msgID := errorcorrection.GenerateMessageID()
				msg, _ := errorcorrection.NewProtectedMessage(msgID, []byte("test"), 10, 5)
				return msg
			},
			shard: func(msgID errorcorrection.MessageID) *errorcorrection.Shard {
				// Create shard with DIFFERENT message ID
				differentMsgID := errorcorrection.GenerateMessageID()
				shardID := errorcorrection.GenerateShardID()
				shard, _ := errorcorrection.NewShard(shardID, differentMsgID, 0, []byte("shard data"), false)
				return shard
			},
			wantErr:   true,
			errCheck:  errorcorrection.ErrShardMessageMismatch,
			wantCount: 0,
		},
		{
			name: "add_multiple_shards",
			setupMsg: func() *errorcorrection.ProtectedMessage {
				msgID := errorcorrection.GenerateMessageID()
				msg, _ := errorcorrection.NewProtectedMessage(msgID, []byte("test"), 10, 5)
				// Add first shard
				shardID1 := errorcorrection.GenerateShardID()
				shard1, _ := errorcorrection.NewShard(shardID1, msgID, 0, []byte("shard 1"), false)
				_ = msg.AddShard(shard1)
				// Add second shard
				shardID2 := errorcorrection.GenerateShardID()
				shard2, _ := errorcorrection.NewShard(shardID2, msgID, 1, []byte("shard 2"), false)
				_ = msg.AddShard(shard2)
				return msg
			},
			shard: func(msgID errorcorrection.MessageID) *errorcorrection.Shard {
				// Add third shard
				shardID3 := errorcorrection.GenerateShardID()
				shard3, _ := errorcorrection.NewShard(shardID3, msgID, 2, []byte("shard 3"), false)
				return shard3
			},
			wantErr:   false,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			msg := tt.setupMsg()
			msgID := msg.ID
			shard := tt.shard(msgID)

			// Act
			err := msg.AddShard(shard)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, msg.Shards, tt.wantCount)
		})
	}
}

func TestProtectedMessage_TotalShards(t *testing.T) {
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
			name:         "balanced_10_5",
			dataShards:   10,
			parityShards: 5,
			wantTotal:    15,
		},
		{
			name:         "aggressive_15_3",
			dataShards:   15,
			parityShards: 3,
			wantTotal:    18,
		},
		{
			name:         "minimal_1_1",
			dataShards:   1,
			parityShards: 1,
			wantTotal:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			msgID := errorcorrection.GenerateMessageID()
			msg, err := errorcorrection.NewProtectedMessage(msgID, []byte("test"), tt.dataShards, tt.parityShards)
			require.NoError(t, err)

			// Act
			total := msg.TotalShards()

			// Assert
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestProtectedMessage_MinimumShardsNeeded(t *testing.T) {
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
			msgID := errorcorrection.GenerateMessageID()
			msg, err := errorcorrection.NewProtectedMessage(msgID, []byte("test"), tt.dataShards, tt.parityShards)
			require.NoError(t, err)

			// Act
			minimum := msg.MinimumShardsNeeded()

			// Assert
			assert.Equal(t, tt.wantMinimum, minimum)
			assert.Equal(t, tt.dataShards, minimum, "minimum should equal data shards")
		})
	}
}

func TestProtectedMessage_CanRecover(t *testing.T) {
	tests := []struct {
		name         string
		dataShards   int
		parityShards int
		addShards    int
		canRecover   bool
	}{
		{
			name:         "insufficient_shards_0_of_10",
			dataShards:   10,
			parityShards: 5,
			addShards:    0,
			canRecover:   false,
		},
		{
			name:         "insufficient_shards_9_of_10",
			dataShards:   10,
			parityShards: 5,
			addShards:    9,
			canRecover:   false,
		},
		{
			name:         "exactly_minimum_shards_10_of_10",
			dataShards:   10,
			parityShards: 5,
			addShards:    10,
			canRecover:   true,
		},
		{
			name:         "more_than_minimum_11_of_10",
			dataShards:   10,
			parityShards: 5,
			addShards:    11,
			canRecover:   true,
		},
		{
			name:         "all_shards_15_of_10",
			dataShards:   10,
			parityShards: 5,
			addShards:    15,
			canRecover:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			msgID := errorcorrection.GenerateMessageID()
			msg, err := errorcorrection.NewProtectedMessage(msgID, []byte("test"), tt.dataShards, tt.parityShards)
			require.NoError(t, err)

			// Add specified number of shards
			for i := 0; i < tt.addShards; i++ {
				shardID := errorcorrection.GenerateShardID()
				shard, err := errorcorrection.NewShard(shardID, msgID, i, []byte("shard data"), false)
				require.NoError(t, err)
				err = msg.AddShard(shard)
				require.NoError(t, err)
			}

			// Act
			canRecover := msg.CanRecover()

			// Assert
			assert.Equal(t, tt.canRecover, canRecover)
		})
	}
}

func TestProtectedMessage_Validate(t *testing.T) {
	tests := []struct {
		name     string
		setupMsg func() *errorcorrection.ProtectedMessage
		wantErr  bool
		errCheck error
	}{
		{
			name: "valid_message",
			setupMsg: func() *errorcorrection.ProtectedMessage {
				msgID := errorcorrection.GenerateMessageID()
				msg, _ := errorcorrection.NewProtectedMessage(msgID, []byte("test"), 10, 5)
				return msg
			},
			wantErr: false,
		},
		{
			name: "zero_message_id",
			setupMsg: func() *errorcorrection.ProtectedMessage {
				msg, _ := errorcorrection.NewProtectedMessage(errorcorrection.MessageID{}, []byte("test"), 10, 5)
				return msg
			},
			wantErr:  true,
			errCheck: errorcorrection.ErrInvalidMessageID,
		},
		{
			name: "invalid_data_shards",
			setupMsg: func() *errorcorrection.ProtectedMessage {
				msgID := errorcorrection.GenerateMessageID()
				msg := &errorcorrection.ProtectedMessage{
					ID:           msgID,
					OriginalData: []byte("test"),
					DataShards:   0,
					ParityShards: 5,
				}
				return msg
			},
			wantErr:  true,
			errCheck: errorcorrection.ErrInvalidShardCount,
		},
		{
			name: "invalid_parity_shards",
			setupMsg: func() *errorcorrection.ProtectedMessage {
				msgID := errorcorrection.GenerateMessageID()
				msg := &errorcorrection.ProtectedMessage{
					ID:           msgID,
					OriginalData: []byte("test"),
					DataShards:   10,
					ParityShards: 0,
				}
				return msg
			},
			wantErr:  true,
			errCheck: errorcorrection.ErrInvalidShardCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			msg := tt.setupMsg()

			// Act
			err := msg.Validate()

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Shard Tests
// ============================================================================

func TestShard_NewShard(t *testing.T) {
	tests := []struct {
		name      string
		shardID   errorcorrection.ShardID
		messageID errorcorrection.MessageID
		index     int
		data      []byte
		isParity  bool
		wantErr   bool
		errCheck  error
	}{
		{
			name:      "valid_data_shard",
			shardID:   errorcorrection.GenerateShardID(),
			messageID: errorcorrection.GenerateMessageID(),
			index:     0,
			data:      []byte("shard data content"),
			isParity:  false,
			wantErr:   false,
		},
		{
			name:      "valid_parity_shard",
			shardID:   errorcorrection.GenerateShardID(),
			messageID: errorcorrection.GenerateMessageID(),
			index:     10,
			data:      []byte("parity shard content"),
			isParity:  true,
			wantErr:   false,
		},
		{
			name:      "zero_shard_id",
			shardID:   errorcorrection.ShardID{},
			messageID: errorcorrection.GenerateMessageID(),
			index:     0,
			data:      []byte("data"),
			isParity:  false,
			wantErr:   true,
			errCheck:  errorcorrection.ErrInvalidShardID,
		},
		{
			name:      "zero_message_id",
			shardID:   errorcorrection.GenerateShardID(),
			messageID: errorcorrection.MessageID{},
			index:     0,
			data:      []byte("data"),
			isParity:  false,
			wantErr:   true,
			errCheck:  errorcorrection.ErrInvalidMessageID,
		},
		{
			name:      "negative_index",
			shardID:   errorcorrection.GenerateShardID(),
			messageID: errorcorrection.GenerateMessageID(),
			index:     -1,
			data:      []byte("data"),
			isParity:  false,
			wantErr:   true,
			errCheck:  errorcorrection.ErrInvalidShardIndex,
		},
		{
			name:      "empty_data",
			shardID:   errorcorrection.GenerateShardID(),
			messageID: errorcorrection.GenerateMessageID(),
			index:     0,
			data:      []byte{},
			isParity:  false,
			wantErr:   true,
			errCheck:  errorcorrection.ErrEmptyShardData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			shard, err := errorcorrection.NewShard(tt.shardID, tt.messageID, tt.index, tt.data, tt.isParity)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
				assert.Nil(t, shard)
			} else {
				require.NoError(t, err)
				require.NotNil(t, shard)
				assert.True(t, shard.ID.Equals(tt.shardID))
				assert.True(t, shard.MessageID.Equals(tt.messageID))
				assert.Equal(t, tt.index, shard.Index)
				assert.Equal(t, tt.data, shard.Data)
				assert.Equal(t, tt.isParity, shard.IsParity)
				assert.NotZero(t, shard.CreatedAt)
				assert.WithinDuration(t, time.Now(), shard.CreatedAt, time.Second)
			}
		})
	}
}

func TestShard_SetChecksum(t *testing.T) {
	t.Run("set_checksum", func(t *testing.T) {
		// Arrange
		shardID := errorcorrection.GenerateShardID()
		messageID := errorcorrection.GenerateMessageID()
		shard, err := errorcorrection.NewShard(shardID, messageID, 0, []byte("test data"), false)
		require.NoError(t, err)

		checksum := []byte{0x12, 0x34, 0x56, 0x78}

		// Act
		shard.SetChecksum(checksum)

		// Assert
		assert.Equal(t, checksum, shard.Checksum)
	})
}

func TestShard_VerifyChecksum(t *testing.T) {
	tests := []struct {
		name       string
		stored     []byte
		calculated []byte
		wantValid  bool
	}{
		{
			name:       "matching_checksums",
			stored:     []byte{0x12, 0x34, 0x56, 0x78},
			calculated: []byte{0x12, 0x34, 0x56, 0x78},
			wantValid:  true,
		},
		{
			name:       "mismatched_checksums",
			stored:     []byte{0x12, 0x34, 0x56, 0x78},
			calculated: []byte{0x12, 0x34, 0x56, 0xFF},
			wantValid:  false,
		},
		{
			name:       "empty_stored_checksum",
			stored:     []byte{},
			calculated: []byte{0x12, 0x34, 0x56, 0x78},
			wantValid:  false,
		},
		{
			name:       "empty_calculated_checksum",
			stored:     []byte{0x12, 0x34, 0x56, 0x78},
			calculated: []byte{},
			wantValid:  false,
		},
		{
			name:       "both_empty",
			stored:     []byte{},
			calculated: []byte{},
			wantValid:  false,
		},
		{
			name:       "different_lengths",
			stored:     []byte{0x12, 0x34},
			calculated: []byte{0x12, 0x34, 0x56, 0x78},
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			shardID := errorcorrection.GenerateShardID()
			messageID := errorcorrection.GenerateMessageID()
			shard, err := errorcorrection.NewShard(shardID, messageID, 0, []byte("test"), false)
			require.NoError(t, err)
			shard.SetChecksum(tt.stored)

			// Act
			valid := shard.VerifyChecksum(tt.calculated)

			// Assert
			assert.Equal(t, tt.wantValid, valid)
		})
	}
}

func TestShard_VerifyChecksum_ConstantTime(t *testing.T) {
	t.Run("constant_time_comparison", func(t *testing.T) {
		// Arrange: Create shard with checksum
		shardID := errorcorrection.GenerateShardID()
		messageID := errorcorrection.GenerateMessageID()
		shard, err := errorcorrection.NewShard(shardID, messageID, 0, []byte("test data"), false)
		require.NoError(t, err)

		storedChecksum := []byte{0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD, 0xEF, 0x01}
		shard.SetChecksum(storedChecksum)

		matchingChecksum := []byte{0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD, 0xEF, 0x01}
		mismatchedChecksum := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		// Act: Measure timing for matching comparison
		// Use higher iterations to reduce system timing noise impact
		iterations := 50000
		start := time.Now()
		for i := 0; i < iterations; i++ {
			shard.VerifyChecksum(matchingChecksum)
		}
		matchingDuration := time.Since(start)

		// Act: Measure timing for mismatched comparison
		start = time.Now()
		for i := 0; i < iterations; i++ {
			shard.VerifyChecksum(mismatchedChecksum)
		}
		mismatchedDuration := time.Since(start)

		// Assert: Timing should be similar (within 25% tolerance)
		// Note: System-level timing noise can cause variations; the important
		// part is that the underlying implementation uses constant-time operations
		// (XOR loop without early exit). This test verifies no order-of-magnitude
		// difference exists, which would indicate a timing attack vulnerability.
		ratio := float64(matchingDuration) / float64(mismatchedDuration)
		assert.InDelta(t, 1.0, ratio, 0.25,
			"Timing difference too large: %v vs %v (ratio: %.2f)\n"+
				"Note: Some variance is expected due to system noise",
			matchingDuration, mismatchedDuration, ratio)
	})
}

func TestShard_Validate(t *testing.T) {
	tests := []struct {
		name       string
		setupShard func() *errorcorrection.Shard
		wantErr    bool
		errCheck   error
	}{
		{
			name: "valid_shard",
			setupShard: func() *errorcorrection.Shard {
				shardID := errorcorrection.GenerateShardID()
				messageID := errorcorrection.GenerateMessageID()
				shard, _ := errorcorrection.NewShard(shardID, messageID, 0, []byte("test"), false)
				return shard
			},
			wantErr: false,
		},
		{
			name: "zero_shard_id",
			setupShard: func() *errorcorrection.Shard {
				messageID := errorcorrection.GenerateMessageID()
				shard := &errorcorrection.Shard{
					ID:        errorcorrection.ShardID{},
					MessageID: messageID,
					Index:     0,
					Data:      []byte("test"),
				}
				return shard
			},
			wantErr:  true,
			errCheck: errorcorrection.ErrInvalidShardID,
		},
		{
			name: "zero_message_id",
			setupShard: func() *errorcorrection.Shard {
				shardID := errorcorrection.GenerateShardID()
				shard := &errorcorrection.Shard{
					ID:        shardID,
					MessageID: errorcorrection.MessageID{},
					Index:     0,
					Data:      []byte("test"),
				}
				return shard
			},
			wantErr:  true,
			errCheck: errorcorrection.ErrInvalidMessageID,
		},
		{
			name: "negative_index",
			setupShard: func() *errorcorrection.Shard {
				shardID := errorcorrection.GenerateShardID()
				messageID := errorcorrection.GenerateMessageID()
				shard := &errorcorrection.Shard{
					ID:        shardID,
					MessageID: messageID,
					Index:     -1,
					Data:      []byte("test"),
				}
				return shard
			},
			wantErr:  true,
			errCheck: errorcorrection.ErrInvalidShardIndex,
		},
		{
			name: "empty_data",
			setupShard: func() *errorcorrection.Shard {
				shardID := errorcorrection.GenerateShardID()
				messageID := errorcorrection.GenerateMessageID()
				shard := &errorcorrection.Shard{
					ID:        shardID,
					MessageID: messageID,
					Index:     0,
					Data:      []byte{},
				}
				return shard
			},
			wantErr:  true,
			errCheck: errorcorrection.ErrEmptyShardData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			shard := tt.setupShard()

			// Act
			err := shard.Validate()

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					assert.ErrorIs(t, err, tt.errCheck)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

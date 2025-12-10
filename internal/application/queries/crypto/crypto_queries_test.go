package crypto_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
	"github.com/greysquirr3l/shadowforge/internal/application/queries/crypto"
	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

var errKeyPairNotFound = errors.New("key pair not found")

// ============================================================================
// MOCK REPOSITORY
// ============================================================================

// MockRepository is a mock implementation of domain_crypto.Repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetKeyPair(ctx context.Context, keyID string) (*domain_crypto.KeyPair, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain_crypto.KeyPair), args.Error(1)
}

func (m *MockRepository) ListKeyPairs(ctx context.Context, algorithm *domain_crypto.PQCAlgorithm, limit, offset int) ([]*domain_crypto.KeyPair, int, error) {
	args := m.Called(ctx, algorithm, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain_crypto.KeyPair), args.Int(1), args.Error(2)
}

func (m *MockRepository) SaveKeyPair(ctx context.Context, keyPair *domain_crypto.KeyPair) error {
	args := m.Called(ctx, keyPair)
	return args.Error(0)
}

func (m *MockRepository) DeleteKeyPair(ctx context.Context, keyID string) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

// ============================================================================
// GET KEY INFO QUERY TESTS
// ============================================================================

func TestGetKeyInfoQuery_Validate(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.KeyInfoRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_request",
			request: dto.KeyInfoRequest{
				KeyID:     "test-key-id-123",
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr: false,
		},
		{
			name: "empty_key_id",
			request: dto.KeyInfoRequest{
				KeyID:     "",
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr:     true,
			errContains: "key ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := crypto.NewGetKeyInfoQuery(tt.request)
			err := query.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetKeyInfoQueryHandler_Handle(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		request     dto.KeyInfoRequest
		setupMock   func(*MockRepository)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *dto.KeyInfoResponse)
	}{
		{
			name: "success_key_found",
			request: dto.KeyInfoRequest{
				KeyID: "test-key-123",
			},
			setupMock: func(repo *MockRepository) {
				keyPair := &domain_crypto.KeyPair{
					ID:        "test-key-123",
					Algorithm: domain_crypto.Kyber1024,
					PublicKey: []byte("public-key-data"),
					CreatedAt: now.Add(-24 * time.Hour),
					ExpiresAt: &[]time.Time{now.Add(24 * time.Hour)}[0],
				}
				repo.On("GetKeyPair", ctx, "test-key-123").Return(keyPair, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.KeyInfoResponse) {
				assert.Equal(t, "test-key-123", resp.KeyID)
				assert.Equal(t, domain_crypto.Kyber1024, resp.Algorithm)
				assert.NotEmpty(t, resp.PublicKey)
				assert.NotNil(t, resp.ExpiresAt)
			},
		},
		{
			name: "success_expired_key",
			request: dto.KeyInfoRequest{
				KeyID: "expired-key",
			},
			setupMock: func(repo *MockRepository) {
				keyPair := &domain_crypto.KeyPair{
					ID:        "expired-key",
					Algorithm: domain_crypto.Dilithium3,
					PublicKey: []byte("public-key-data"),
					CreatedAt: now.Add(-48 * time.Hour),
					ExpiresAt: &[]time.Time{now.Add(-1 * time.Hour)}[0], // Expired 1 hour ago
				}
				repo.On("GetKeyPair", ctx, "expired-key").Return(keyPair, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.KeyInfoResponse) {
				assert.Equal(t, "expired-key", resp.KeyID)
				assert.Equal(t, domain_crypto.Dilithium3, resp.Algorithm)
			},
		},
		{
			name: "success_no_expiration",
			request: dto.KeyInfoRequest{
				KeyID: "no-expiration-key",
			},
			setupMock: func(repo *MockRepository) {
				keyPair := &domain_crypto.KeyPair{
					ID:        "no-expiration-key",
					Algorithm: domain_crypto.Kyber1024,
					PublicKey: []byte("public-key-data"),
					CreatedAt: now,
					ExpiresAt: nil, // No expiration
				}
				repo.On("GetKeyPair", ctx, "no-expiration-key").Return(keyPair, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.KeyInfoResponse) {
				assert.Equal(t, "no-expiration-key", resp.KeyID)
				assert.Nil(t, resp.ExpiresAt)
			},
		},
		{
			name: "key_not_found",
			request: dto.KeyInfoRequest{
				KeyID: "non-existent",
			},
			setupMock: func(repo *MockRepository) {
				repo.On("GetKeyPair", ctx, "non-existent").Return(nil, errKeyPairNotFound)
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "repository_error",
			request: dto.KeyInfoRequest{
				KeyID: "error-key",
			},
			setupMock: func(repo *MockRepository) {
				repo.On("GetKeyPair", ctx, "error-key").Return(nil, assert.AnError)
			},
			wantErr:     true,
			errContains: "assert.AnError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			handler := crypto.NewGetKeyInfoQueryHandler(mockRepo)
			query := crypto.NewGetKeyInfoQuery(tt.request)

			// Act
			result, err := handler.Handle(ctx, query)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)

				if tt.validate != nil {
					tt.validate(t, &result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// LIST KEYS QUERY TESTS
// ============================================================================
func TestListKeysQuery_Validate(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.ListKeysRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_request_no_filter",
			request: dto.ListKeysRequest{
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "valid_request_with_filter",
			request: dto.ListKeysRequest{
				Algorithm: &domain_crypto.Kyber1024,
				Limit:     20,
				Offset:    10,
			},
			wantErr: false,
		},
		{
			name: "zero_limit",
			request: dto.ListKeysRequest{
				Limit:  0,
				Offset: 0,
			},
			wantErr:     true,
			errContains: "limit must be greater than 0",
		},
		{
			name: "negative_limit",
			request: dto.ListKeysRequest{
				Limit:  -5,
				Offset: 0,
			},
			wantErr:     true,
			errContains: "limit must be greater than 0",
		},
		{
			name: "negative_offset",
			request: dto.ListKeysRequest{
				Limit:  10,
				Offset: -1,
			},
			wantErr:     true,
			errContains: "offset must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := crypto.NewListKeysQuery(tt.request)
			err := query.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListKeysQueryHandler_Handle(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		request     dto.ListKeysRequest
		setupMock   func(*MockRepository)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *dto.ListKeysResponse)
	}{
		{
			name: "success_no_filter",
			request: dto.ListKeysRequest{
				Limit:  10,
				Offset: 0,
			},
			setupMock: func(repo *MockRepository) {
				keyPairs := []*domain_crypto.KeyPair{
					{
						ID:        "key-1",
						Algorithm: domain_crypto.Kyber1024,
						CreatedAt: now.Add(-24 * time.Hour),
						ExpiresAt: &[]time.Time{now.Add(24 * time.Hour)}[0],
					},
					{
						ID:        "key-2",
						Algorithm: domain_crypto.Dilithium3,
						CreatedAt: now.Add(-48 * time.Hour),
						ExpiresAt: nil,
					},
				}
				repo.On("ListKeyPairs", ctx, (*domain_crypto.PQCAlgorithm)(nil), 10, 0).
					Return(keyPairs, 2, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.ListKeysResponse) {
				assert.Len(t, resp.Keys, 2)
				assert.Equal(t, 2, resp.Count)
				assert.Equal(t, 10, resp.Limit)
				assert.Equal(t, 0, resp.Offset)

				// First key
				assert.Equal(t, "key-1", resp.Keys[0].KeyID)
				assert.Equal(t, domain_crypto.Kyber1024, resp.Keys[0].Algorithm)
				assert.False(t, resp.Keys[0].IsExpired)

				// Second key
				assert.Equal(t, "key-2", resp.Keys[1].KeyID)
				assert.Equal(t, domain_crypto.Dilithium3, resp.Keys[1].Algorithm)
				assert.False(t, resp.Keys[1].IsExpired)
			},
		},
		{
			name: "success_with_algorithm_filter",
			request: dto.ListKeysRequest{
				Algorithm: &domain_crypto.Kyber1024,
				Limit:     5,
				Offset:    0,
			},
			setupMock: func(repo *MockRepository) {
				keyPairs := []*domain_crypto.KeyPair{
					{
						ID:        "kyber-key-1",
						Algorithm: domain_crypto.Kyber1024,
						CreatedAt: now,
						ExpiresAt: nil,
					},
				}
				repo.On("ListKeyPairs", ctx, &domain_crypto.Kyber1024, 5, 0).
					Return(keyPairs, 1, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.ListKeysResponse) {
				assert.Len(t, resp.Keys, 1)
				assert.Equal(t, 1, resp.Count)
				assert.Equal(t, domain_crypto.Kyber1024, resp.Keys[0].Algorithm)
			},
		},
		{
			name: "success_expired_key_detection",
			request: dto.ListKeysRequest{
				Limit:  10,
				Offset: 0,
			},
			setupMock: func(repo *MockRepository) {
				keyPairs := []*domain_crypto.KeyPair{
					{
						ID:        "expired-key",
						Algorithm: domain_crypto.Kyber1024,
						CreatedAt: now.Add(-72 * time.Hour),
						ExpiresAt: &[]time.Time{now.Add(-1 * time.Hour)}[0], // Expired
					},
				}
				repo.On("ListKeyPairs", ctx, (*domain_crypto.PQCAlgorithm)(nil), 10, 0).
					Return(keyPairs, 1, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.ListKeysResponse) {
				assert.Len(t, resp.Keys, 1)
				assert.True(t, resp.Keys[0].IsExpired)
			},
		},
		{
			name: "success_pagination",
			request: dto.ListKeysRequest{
				Limit:  2,
				Offset: 5,
			},
			setupMock: func(repo *MockRepository) {
				keyPairs := []*domain_crypto.KeyPair{
					{
						ID:        "key-6",
						Algorithm: domain_crypto.Kyber1024,
						CreatedAt: now,
						ExpiresAt: nil,
					},
					{
						ID:        "key-7",
						Algorithm: domain_crypto.Dilithium3,
						CreatedAt: now,
						ExpiresAt: nil,
					},
				}
				repo.On("ListKeyPairs", ctx, (*domain_crypto.PQCAlgorithm)(nil), 2, 5).
					Return(keyPairs, 10, nil) // Total count is 10
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.ListKeysResponse) {
				assert.Len(t, resp.Keys, 2)
				assert.Equal(t, 10, resp.Count)
				assert.Equal(t, 2, resp.Limit)
				assert.Equal(t, 5, resp.Offset)
			},
		},
		{
			name: "success_empty_result",
			request: dto.ListKeysRequest{
				Algorithm: &domain_crypto.Kyber1024,
				Limit:     10,
				Offset:    0,
			},
			setupMock: func(repo *MockRepository) {
				repo.On("ListKeyPairs", ctx, &domain_crypto.Kyber1024, 10, 0).
					Return([]*domain_crypto.KeyPair{}, 0, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.ListKeysResponse) {
				assert.Empty(t, resp.Keys)
				assert.Equal(t, 0, resp.Count)
			},
		},
		{
			name: "repository_error",
			request: dto.ListKeysRequest{
				Limit:  10,
				Offset: 0,
			},
			setupMock: func(repo *MockRepository) {
				repo.On("ListKeyPairs", ctx, (*domain_crypto.PQCAlgorithm)(nil), 10, 0).
					Return(nil, 0, assert.AnError)
			},
			wantErr:     true,
			errContains: "assert.AnError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			handler := crypto.NewListKeysQueryHandler(mockRepo)
			query := crypto.NewListKeysQuery(tt.request)

			// Act
			result, err := handler.Handle(ctx, query)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				response, ok := result.(dto.ListKeysResponse)
				require.True(t, ok, "result should be dto.ListKeysResponse")

				if tt.validate != nil {
					tt.validate(t, &response)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// GET CAPACITY QUERY TESTS
// ============================================================================

func TestGetCapacityQuery_Validate(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.GetCapacityRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_request_kyber",
			request: dto.GetCapacityRequest{
				MediaSize: 1024 * 1024, // 1 MB
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr: false,
		},
		{
			name: "valid_request_dilithium",
			request: dto.GetCapacityRequest{
				MediaSize: 5 * 1024 * 1024, // 5 MB
				Algorithm: domain_crypto.Dilithium3,
			},
			wantErr: false,
		},
		{
			name: "zero_media_size",
			request: dto.GetCapacityRequest{
				MediaSize: 0,
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr:     true,
			errContains: "media size must be greater than 0",
		},
		{
			name: "negative_media_size",
			request: dto.GetCapacityRequest{
				MediaSize: -1000,
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr:     true,
			errContains: "media size must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := crypto.NewGetCapacityQuery(tt.request)
			err := query.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetCapacityQueryHandler_Handle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		request     dto.GetCapacityRequest
		wantErr     bool
		errContains string
		validate    func(*testing.T, *dto.GetCapacityResponse)
	}{
		{
			name: "success_kyber1024_overhead",
			request: dto.GetCapacityRequest{
				MediaSize: 10000,
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.GetCapacityResponse) {
				expectedOverhead := int64(1628) // Kyber overhead
				safetyMargin := int64(64)
				expectedMax := int64(10000) - expectedOverhead - safetyMargin

				assert.Equal(t, expectedMax, resp.MaxPayloadSize)
				assert.Equal(t, domain_crypto.Kyber1024, resp.Algorithm)
				assert.Equal(t, expectedOverhead, resp.OverheadBytes)
				assert.WithinDuration(t, time.Now(), resp.CalculatedAt, 2*time.Second)
			},
		},
		{
			name: "success_dilithium3_overhead",
			request: dto.GetCapacityRequest{
				MediaSize: 20000,
				Algorithm: domain_crypto.Dilithium3,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.GetCapacityResponse) {
				expectedOverhead := int64(5245) // Dilithium overhead
				safetyMargin := int64(64)
				expectedMax := int64(20000) - expectedOverhead - safetyMargin

				assert.Equal(t, expectedMax, resp.MaxPayloadSize)
				assert.Equal(t, domain_crypto.Dilithium3, resp.Algorithm)
				assert.Equal(t, expectedOverhead, resp.OverheadBytes)
			},
		},
		{
			name: "small_media_size_underflow_protection",
			request: dto.GetCapacityRequest{
				MediaSize: 1000, // Small size
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.GetCapacityResponse) {
				// 1000 - 1628 - 64 = negative, should be 0
				assert.Equal(t, int64(0), resp.MaxPayloadSize)
				assert.Equal(t, int64(1628), resp.OverheadBytes)
			},
		},
		{
			name: "large_media_size",
			request: dto.GetCapacityRequest{
				MediaSize: 100 * 1024 * 1024, // 100 MB
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.GetCapacityResponse) {
				expectedOverhead := int64(1628)
				safetyMargin := int64(64)
				expectedMax := int64(100*1024*1024) - expectedOverhead - safetyMargin

				assert.Equal(t, expectedMax, resp.MaxPayloadSize)
				assert.Greater(t, resp.MaxPayloadSize, int64(100*1024*1024-2000))
			},
		},
		{
			name: "exact_overhead_size",
			request: dto.GetCapacityRequest{
				MediaSize: 1692, // Exactly Kyber overhead + safety margin
				Algorithm: domain_crypto.Kyber1024,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *dto.GetCapacityResponse) {
				// 1692 - 1628 - 64 = 0
				assert.Equal(t, int64(0), resp.MaxPayloadSize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := crypto.NewGetCapacityQueryHandler()
			query := crypto.NewGetCapacityQuery(tt.request)

			// Act
			result, err := handler.Handle(ctx, query)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				response, ok := result.(dto.GetCapacityResponse)
				require.True(t, ok, "result should be dto.GetCapacityResponse")

				if tt.validate != nil {
					tt.validate(t, &response)
				}
			}
		})
	}
}

// ============================================================================
// QUERY NAME TESTS
// ============================================================================

func TestQueryNames(t *testing.T) {
	tests := []struct {
		name     string
		query    interface{ QueryName() string }
		wantName string
	}{
		{
			name:     "get_key_info_query_name",
			query:    crypto.NewGetKeyInfoQuery(dto.KeyInfoRequest{KeyID: "test"}),
			wantName: "crypto.GetKeyInfo",
		},
		{
			name:     "list_keys_query_name",
			query:    crypto.NewListKeysQuery(dto.ListKeysRequest{Limit: 10}),
			wantName: "crypto.ListKeys",
		},
		{
			name:     "get_capacity_query_name",
			query:    crypto.NewGetCapacityQuery(dto.GetCapacityRequest{MediaSize: 1000, Algorithm: domain_crypto.Kyber1024}),
			wantName: "crypto.GetCapacity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.query.QueryName())
		})
	}
}

# CQRS Query Patterns - Shadowforge

> **Reference Implementation**: Crypto Domain Queries (90.1% coverage)

This document outlines the established patterns for implementing CQRS queries across all Shadowforge bounded contexts.

---

## Table of Contents

1. [Query Structure](#query-structure)
2. [Query Handler Structure](#query-handler-structure)
3. [DTO Definitions](#dto-definitions)
4. [Repository Integration](#repository-integration)
5. [Testing Patterns](#testing-patterns)
6. [Implementation Checklist](#implementation-checklist)

---

## Query Structure

### Required Components

Every query must implement the `queries.Query` interface:

```go
type Query interface {
    QueryName() string
    Validate() error
    Request() interface{}
}
```

### Example Implementation

```go
// key_info_query.go
package crypto

import (
    "github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
    "github.com/greysquirr3l/shadowforge/internal/application/queries"
)

type GetKeyInfoQuery struct {
    request dto.KeyInfoRequest
}

// NewGetKeyInfoQuery creates a new query instance
func NewGetKeyInfoQuery(keyID string, algorithm domain_crypto.PQCAlgorithm) *GetKeyInfoQuery {
    return &GetKeyInfoQuery{
        request: dto.KeyInfoRequest{
            KeyID:     keyID,
            Algorithm: algorithm,
        },
    }
}

// QueryName returns the unique identifier for this query
func (q *GetKeyInfoQuery) QueryName() string {
    return "crypto.GetKeyInfo"
}

// Validate ensures the query has all required parameters
func (q *GetKeyInfoQuery) Validate() error {
    if q.request.KeyID == "" {
        return fmt.Errorf("key ID is required")
    }
    return nil
}

// Request returns the query request DTO
func (q *GetKeyInfoQuery) Request() interface{} {
    return q.request
}
```

### Key Patterns

✅ **QueryName Format**: `<domain>.<Operation>` (e.g., "crypto.GetKeyInfo", "stego.GetCapacity")

✅ **Validation**: Check all required fields, return descriptive errors

✅ **Constructor**: Provide `New*Query()` function accepting primitive types

❌ **Anti-Pattern**: Don't expose internal request struct directly

---

## Query Handler Structure

### Handler Interface

All handlers must implement:

```go
type QueryHandler interface {
    Handle(ctx context.Context, query interface{}) (interface{}, error)
}
```

### Example Implementation

```go
type GetKeyInfoQueryHandler struct {
    repository domain_crypto.Repository
    logger     *slog.Logger
}

func NewGetKeyInfoQueryHandler(
    repository domain_crypto.Repository,
    logger *slog.Logger,
) *GetKeyInfoQueryHandler {
    return &GetKeyInfoQueryHandler{
        repository: repository,
        logger:     logger,
    }
}

func (h *GetKeyInfoQueryHandler) Handle(
    ctx context.Context,
    q queries.Query,
) (crypto.KeyInfoResponse, error) {
    // Type assert to specific query type
    query, ok := q.(*GetKeyInfoQuery)
    if !ok {
        return crypto.KeyInfoResponse{}, fmt.Errorf(
            "invalid query type: expected *GetKeyInfoQuery, got %T", q)
    }

    // Validate query
    if err := query.Validate(); err != nil {
        return crypto.KeyInfoResponse{}, fmt.Errorf("validation failed: %w", err)
    }

    req := query.Request().(dto.KeyInfoRequest)

    // Fetch from repository
    keyPair, err := h.repository.GetKeyPair(ctx, req.KeyID)
    if err != nil {
        h.logger.Error("failed to get key pair",
            slog.String("key_id", req.KeyID),
            slog.String("error", err.Error()))
        return crypto.KeyInfoResponse{}, fmt.Errorf("failed to get key pair: %w", err)
    }

    // Calculate derived fields (e.g., IsExpired)
    var expiresAt *time.Time
    var isExpired bool
    if keyPair.ExpiresAt != nil {
        expiresAt = keyPair.ExpiresAt
        isExpired = time.Now().After(*keyPair.ExpiresAt)
    }

    // Build response DTO
    response := crypto.KeyInfoResponse{
        KeyID:      keyPair.ID,
        Algorithm:  keyPair.Algorithm,
        CreatedAt:  keyPair.CreatedAt,
        ExpiresAt:  expiresAt,
        PublicKey:  keyPair.PublicKey,
        PrivateKey: keyPair.PrivateKey,
    }

    return response, nil
}
```

### Return Type Pattern

⚠️ **CRITICAL**: Handlers return **struct values**, not pointers:

```go
// ✅ CORRECT - Return struct value
func (h *Handler) Handle(ctx context.Context, q interface{}) (interface{}, error) {
    response := dto.SomeResponse{
        Field1: value1,
        Field2: value2,
    }
    return response, nil  // Return struct, not &response
}

// ❌ INCORRECT - Don't return pointers unless absolutely necessary
return &response, nil  // Avoid this pattern
```

**Rationale**: Returning struct values via `interface{}` prevents unnecessary heap allocations and simplifies type assertions in tests.

### Key Patterns

✅ **Type Assertion**: Validate query type at start of handler

✅ **Validation**: Call `query.Validate()` even if already validated

✅ **Logging**: Log errors with structured context

✅ **Error Wrapping**: Use `fmt.Errorf()` with `%w` for error chains

✅ **Response Construction**: Build complete DTO before returning

❌ **Anti-Pattern**: Don't skip validation in handler

❌ **Anti-Pattern**: Don't return nil for successful queries (return zero value struct)

---

## DTO Definitions

### Request DTOs

```go
// queries.go in dto/<domain>/ package
package crypto

import (
    domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// KeyInfoRequest represents a request to get key information
type KeyInfoRequest struct {
    KeyID     string                      `json:"key_id"`
    Algorithm domain_crypto.PQCAlgorithm `json:"algorithm"`
}

// ListKeysRequest represents a request to list key pairs
type ListKeysRequest struct {
    Algorithm *domain_crypto.PQCAlgorithm `json:"algorithm,omitempty"` // Optional filter
    Limit     int                         `json:"limit"`
    Offset    int                         `json:"offset"`
}
```

### Response DTOs

```go
// KeyInfoResponse contains detailed key information
type KeyInfoResponse struct {
    KeyID      string                      `json:"key_id"`
    Algorithm  domain_crypto.PQCAlgorithm `json:"algorithm"`
    CreatedAt  time.Time                   `json:"created_at"`
    ExpiresAt  *time.Time                  `json:"expires_at,omitempty"` // Nullable
    PublicKey  []byte                      `json:"public_key"`
    PrivateKey []byte                      `json:"private_key"`
}

// ListKeysResponse contains paginated key listing
type ListKeysResponse struct {
    Keys   []KeySummary `json:"keys"`
    Count  int          `json:"count"`   // Total matching keys
    Limit  int          `json:"limit"`   // Requested page size
    Offset int          `json:"offset"`  // Starting position
}

// KeySummary provides summary information for list views
type KeySummary struct {
    KeyID     string                      `json:"key_id"`
    Algorithm domain_crypto.PQCAlgorithm `json:"algorithm"`
    CreatedAt time.Time                   `json:"created_at"`
    ExpiresAt *time.Time                  `json:"expires_at,omitempty"`
    IsExpired bool                        `json:"is_expired"`
}
```

### Key Patterns

✅ **JSON Tags**: Always include `json` tags for API serialization

✅ **Nullable Fields**: Use pointers for optional fields (`*time.Time`, `*string`)

✅ **Enums**: Use domain types (`domain_crypto.PQCAlgorithm`) not strings

✅ **Summary Types**: Create lightweight types for list responses

✅ **Derived Fields**: Include calculated fields (e.g., `IsExpired`) in responses

❌ **Anti-Pattern**: Don't duplicate domain entities in DTOs (transform instead)

---

## Repository Integration

### Read-Only Pattern

Queries **NEVER** modify state. Use read-only repository methods:

```go
type Repository interface {
    GetKeyPair(ctx context.Context, keyID string) (*KeyPair, error)
    ListKeyPairs(ctx context.Context, algorithm *PQCAlgorithm, limit, offset int) ([]*KeyPair, int, error)
    // Commands would use: SaveKeyPair, DeleteKeyPair, etc.
}
```

### List/Pagination Pattern

```go
func (h *ListKeysQueryHandler) Handle(ctx context.Context, q interface{}) (interface{}, error) {
    // ... validation ...

    req := query.Request().(dto.ListKeysRequest)

    // Repository returns items + total count
    keyPairs, totalCount, err := h.repository.ListKeyPairs(
        ctx,
        req.Algorithm,  // Optional filter
        req.Limit,
        req.Offset,
    )
    if err != nil {
        return dto.ListKeysResponse{}, fmt.Errorf("failed to list keys: %w", err)
    }

    // Transform domain entities to DTOs
    keys := make([]dto.KeySummary, len(keyPairs))
    for i, kp := range keyPairs {
        keys[i] = dto.KeySummary{
            KeyID:     kp.ID,
            Algorithm: kp.Algorithm,
            CreatedAt: kp.CreatedAt,
            ExpiresAt: kp.ExpiresAt,
            IsExpired: kp.ExpiresAt != nil && time.Now().After(*kp.ExpiresAt),
        }
    }

    return dto.ListKeysResponse{
        Keys:   keys,
        Count:  totalCount,  // Total items, not page size
        Limit:  req.Limit,
        Offset: req.Offset,
    }, nil
}
```

### Key Patterns

✅ **Total Count**: Return total matching items for pagination UI

✅ **Filtering**: Support optional filters with nullable parameters

✅ **Transformation**: Convert domain entities to DTOs in handler

✅ **Derived Fields**: Calculate runtime values (IsExpired, etc.)

❌ **Anti-Pattern**: Don't return domain entities directly in responses

---

## Testing Patterns

### MockRepository Pattern

```go
// crypto_queries_test.go
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

func (m *MockRepository) ListKeyPairs(
    ctx context.Context,
    algorithm *domain_crypto.PQCAlgorithm,
    limit, offset int,
) ([]*domain_crypto.KeyPair, int, error) {
    args := m.Called(ctx, algorithm, limit, offset)
    if args.Get(0) == nil {
        return nil, 0, args.Error(2)
    }
    return args.Get(0).([]*domain_crypto.KeyPair), args.Int(1), args.Error(2)
}
```

### Table-Driven Validation Tests

```go
func TestGetKeyInfoQuery_Validate(t *testing.T) {
    tests := []struct {
        name        string
        keyID       string
        algorithm   domain_crypto.PQCAlgorithm
        expectError bool
        errorMsg    string
    }{
        {
            name:        "valid_request",
            keyID:       "test-key-123",
            algorithm:   domain_crypto.Kyber1024,
            expectError: false,
        },
        {
            name:        "empty_key_id",
            keyID:       "",
            algorithm:   domain_crypto.Kyber1024,
            expectError: true,
            errorMsg:    "key ID is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            query := NewGetKeyInfoQuery(tt.keyID, tt.algorithm)
            err := query.Validate()

            if tt.expectError {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Handler Tests with Type Assertions

⚠️ **CRITICAL**: Handle struct returns correctly:

```go
func TestGetKeyInfoQueryHandler_Handle(t *testing.T) {
    tests := []struct {
        name      string
        keyID     string
        setupMock func(*MockRepository)
        expectErr bool
        validate  func(*testing.T, *crypto.KeyInfoResponse)
    }{
        {
            name:  "success_key_found",
            keyID: "test-key-123",
            setupMock: func(repo *MockRepository) {
                keyPair := &domain_crypto.KeyPair{
                    ID:         "test-key-123",
                    Algorithm:  domain_crypto.Kyber1024,
                    PublicKey:  []byte("public"),
                    PrivateKey: []byte("private"),
                    CreatedAt:  time.Now(),
                    ExpiresAt:  nil,
                }
                repo.On("GetKeyPair", mock.Anything, "test-key-123").
                    Return(keyPair, nil)
            },
            expectErr: false,
            validate: func(t *testing.T, resp *crypto.KeyInfoResponse) {
                assert.Equal(t, "test-key-123", resp.KeyID)
                assert.Equal(t, domain_crypto.Kyber1024, resp.Algorithm)
                assert.NotEmpty(t, resp.PublicKey)
                assert.NotEmpty(t, resp.PrivateKey)
                assert.Nil(t, resp.ExpiresAt)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockRepository)
            if tt.setupMock != nil {
                tt.setupMock(mockRepo)
            }

            handler := NewGetKeyInfoQueryHandler(mockRepo, slog.Default())
            query := NewGetKeyInfoQuery(tt.keyID, domain_crypto.Kyber1024)

            result, err := handler.Handle(context.Background(), query)

            if tt.expectErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                require.NotNil(t, result)

                // ✅ CORRECT - Handler returns struct directly
                response := result.(crypto.KeyInfoResponse)

                // ❌ INCORRECT - Don't use pointer assertion
                // response, ok := result.(*crypto.KeyInfoResponse)

                if tt.validate != nil {
                    tt.validate(t, &response)
                }
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

### List Query Tests with Struct Returns

```go
func TestListKeysQueryHandler_Handle(t *testing.T) {
    tests := []struct {
        name      string
        // ... test fields ...
        validate  func(*testing.T, *dto.ListKeysResponse)
    }{
        {
            name: "success_with_pagination",
            setupMock: func(repo *MockRepository) {
                // ... mock setup ...
            },
            validate: func(t *testing.T, resp *dto.ListKeysResponse) {
                assert.Equal(t, 10, resp.Count)   // Total count
                assert.Equal(t, 2, resp.Limit)    // Page size
                assert.Equal(t, 5, resp.Offset)   // Start position
                assert.Len(t, resp.Keys, 2)       // Items returned
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... handler setup ...

            result, err := handler.Handle(ctx, query)

            require.NoError(t, err)
            require.NotNil(t, result)

            // ✅ CORRECT - Return is struct value, not pointer
            response, ok := result.(dto.ListKeysResponse)
            require.True(t, ok, "result should be dto.ListKeysResponse")

            if tt.validate != nil {
                tt.validate(t, &response)  // Pass pointer to validation
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Coverage Requirements

- **Validation Tests**: Test all edge cases (empty, zero, negative, invalid)
- **Handler Success Tests**: Test normal flows, empty results, pagination
- **Handler Error Tests**: Test repository failures, not found, invalid input
- **Target Coverage**: **80%+ minimum**, **90%+ ideal**

### Key Patterns

✅ **MockRepository**: Use testify/mock for repository

✅ **Table-Driven**: Test multiple scenarios in one function

✅ **Validation Functions**: Use `validate` callback for complex assertions

✅ **Struct Returns**: Assert `result.(dto.Response)` not `result.(*dto.Response)`

✅ **Race Detection**: Run with `-race` flag

❌ **Anti-Pattern**: Don't skip error path tests

❌ **Anti-Pattern**: Don't use pointer assertions for struct returns

---

## Implementation Checklist

### Per-Query Checklist

Use this checklist for each query you implement:

#### 1. Define DTOs (`dto/<domain>/queries.go`)

- [ ] Create request DTO with all required fields
- [ ] Create response DTO with all return fields
- [ ] Add JSON tags for API serialization
- [ ] Use domain types for enums (not strings)
- [ ] Use pointers for nullable fields (`*time.Time`)
- [ ] Create summary types for list responses

#### 2. Implement Query Struct (`<query_name>_query.go`)

- [ ] Define query struct with private request field
- [ ] Implement `QueryName()` string method
- [ ] Implement `Validate()` error method
- [ ] Implement `Request()` interface{} method
- [ ] Create `New*Query()` constructor function
- [ ] Use format `<domain>.<Operation>` for query name

#### 3. Implement Query Handler

- [ ] Define handler struct with repository and logger
- [ ] Create `New*QueryHandler()` constructor
- [ ] Implement `Handle(ctx, interface{}) (interface{}, error)`
- [ ] Type assert to specific query type
- [ ] Call `query.Validate()`
- [ ] Fetch data from repository (read-only)
- [ ] Transform domain entities to DTOs
- [ ] Calculate derived fields (IsExpired, etc.)
- [ ] Return struct value, not pointer
- [ ] Add structured logging for errors

#### 4. Create Test File (`<domain>_queries_test.go`)

- [ ] Create MockRepository implementation
- [ ] Implement all repository methods (GetX, ListX)
- [ ] Write validation tests (5+ test cases)
  - [ ] Valid request
  - [ ] Empty required fields
  - [ ] Zero numeric values
  - [ ] Negative numeric values
  - [ ] Invalid combinations
- [ ] Write handler success tests (3+ scenarios)
  - [ ] Success with data found
  - [ ] Success with optional fields
  - [ ] Success with empty result
  - [ ] Success with pagination
- [ ] Write handler error tests (2+ scenarios)
  - [ ] Repository error
  - [ ] Not found error
- [ ] Write QueryName test
- [ ] Use table-driven test pattern
- [ ] Create validation callback functions

#### 5. Verify Implementation

- [ ] Run `go build ./...` - verify compilation
- [ ] Run `go test -v -race ./...` - all tests pass
- [ ] Run `go test -cover ./...` - check coverage
- [ ] Verify 80%+ coverage achieved
- [ ] Run `golangci-lint run` - no linting errors
- [ ] Check error messages are descriptive
- [ ] Verify struct return types (not pointers)

#### 6. Commit Query Completion

- [ ] Stage files: `git add internal/application/queries/<domain>/`
- [ ] Stage DTOs: `git add internal/application/dto/<domain>/queries.go`
- [ ] Write descriptive commit message
- [ ] Include coverage stats in commit message
- [ ] Reference CQRS pattern compliance

---

## Reference Implementation

**Location**: `internal/application/queries/crypto/`

**Files**:
- `key_info_query.go` - Simple key lookup (108 lines)
- `list_keys_query.go` - Paginated listing (158 lines)
- `capacity_query.go` - Calculation query (111 lines)
- `crypto_queries_test.go` - Comprehensive tests (710 lines, 27 test cases)

**Coverage**: 90.1% with race detection

**Test Statistics**:
- 11 validation tests (edge cases)
- 16 handler tests (success + error paths)
- MockRepository with 4 methods
- Zero race conditions

Use crypto queries as the pattern for implementing queries in the remaining 7 domains.

---

## Common Pitfalls

### Return Type Confusion

```go
// ❌ WRONG - Returning pointer
func (h *Handler) Handle(ctx, q) (interface{}, error) {
    response := dto.Response{...}
    return &response, nil  // DON'T DO THIS
}

// ✅ CORRECT - Returning struct value
func (h *Handler) Handle(ctx, q) (interface{}, error) {
    response := dto.Response{...}
    return response, nil  // Return struct directly
}
```

### Type Assertion in Tests

```go
// ❌ WRONG - Pointer assertion
result, err := handler.Handle(ctx, query)
response, ok := result.(*dto.Response)  // Fails because result is struct

// ✅ CORRECT - Struct assertion
result, err := handler.Handle(ctx, query)
response, ok := result.(dto.Response)   // Works correctly
```

### Validation in Handler

```go
// ❌ WRONG - Skipping validation
func (h *Handler) Handle(ctx, q) (interface{}, error) {
    query := q.(*MyQuery)
    // Skip validation because already called elsewhere
    // DANGEROUS - always validate!
}

// ✅ CORRECT - Always validate
func (h *Handler) Handle(ctx, q) (interface{}, error) {
    query := q.(*MyQuery)
    if err := query.Validate(); err != nil {
        return MyResponse{}, fmt.Errorf("validation failed: %w", err)
    }
    // ... proceed safely ...
}
```

---

## Next Steps

After completing crypto queries (✅ DONE):

1. **Stego Domain Queries** (HIGH PRIORITY):
   - GetEmbedCapacityQuery
   - GetMediaInfoQuery
   - ListEmbeddedDataQuery

2. **Media Domain Queries**:
   - GetMediaCapacityQuery
   - ValidateMediaQuery
   - ListSupportedFormatsQuery

3. **ErrorCorrection Domain Queries**:
   - GetShardInfoQuery
   - ListShardsQuery
   - GetRedundancyQuery

4. **Distribution Domain Queries**:
   - GetManifestQuery
   - ListDistributionsQuery
   - GetShardStatusQuery

5. **Reconstruction Domain Queries**:
   - GetRecoveryStatusQuery
   - ListAvailableShardsQuery

6. **Archive Domain Queries**:
   - GetArchiveInfoQuery
   - ListArchiveContentsQuery

7. **Analysis Domain Queries**:
   - GetDetectabilityScoreQuery
   - GetStatisticalAnalysisQuery

Follow this document for consistent CQRS implementation across all domains.

---

**Last Updated**: December 2025
**Version**: 1.0.0
**Coverage**: Based on crypto queries (90.1% coverage)

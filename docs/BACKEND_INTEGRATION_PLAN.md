# Backend Integration Plan

**Date:** December 12, 2025
**Status:** Interface errors fixed, implementation errors remain
**Goal:** Enable CLI binary build by completing backend handler implementation

---

## Summary

The CLI simulation testing is **COMPLETE and SUCCESSFUL** (all 7 test scenarios passed). However, the CLI binary cannot be built due to incomplete backend handler implementations in `stego_handlers.go`.

**Recent Progress:**
- ✅ Fixed interface naming errors (ECService → ErrorCorrectionService, MediaService → Service)
- ✅ All interface types now correctly reference domain services
- ⚠️ Implementation errors remain (~10+ method call mismatches)

---

## Current Build Errors

**File:** `internal/application/commands/stego_handlers.go`

### Error Categories:

1. **Method Not Found Errors (3)**:
   - `h.mediaService.ProcessMedia undefined` (line 73)
   - `stegoContainer.GetEmbeddedData undefined` (line 131)
   - `quality.GetScore undefined` (line 160)

2. **Function Signature Mismatches (3)**:
   - `h.cryptoService.Encrypt` - wrong args (line 92)
     - Have: `(ctx, []byte, string)`
     - Want: `(ctx, []byte, []byte, PQCAlgorithm)`
   - `h.ecService.Decode` - wrong args (line 247)
     - Have: `(ctx, []byte)`
     - Want: `(ctx, []*Shard, *ShardConfiguration)`
   - `h.cryptoService.Decrypt` - wrong type (line 255)
     - Have: `[]byte`
     - Want: `*CryptoPayload`

3. **Type Conversion Errors (3)**:
   - Cannot use `*CryptoPayload` as `[]byte` (line 96)
   - Cannot use `float64` as `*ShardConfiguration` (line 104)
   - Cannot use `*ProtectedMessage` as `[]byte` (line 108)

4. **Unused Variable (1)**:
   - `coverAsset` declared but not used (line 73)

**Total:** ~10 errors preventing binary build

---

## Domain Service Interfaces

### crypto.CryptoService

```go
type CryptoService interface {
    // Encrypt encrypts data using post-quantum cryptography.
    Encrypt(ctx context.Context, data []byte, publicKey []byte, algorithm PQCAlgorithm) (*CryptoPayload, error)

    // Decrypt decrypts a crypto payload.
    Decrypt(ctx context.Context, payload *CryptoPayload, privateKey []byte) ([]byte, error)

    // Sign creates a digital signature.
    Sign(ctx context.Context, data []byte, privateKey []byte) (*Signature, error)

    // Verify verifies a digital signature.
    Verify(ctx context.Context, data []byte, signature *Signature, publicKey []byte) (bool, error)
}
```

### errorcorrection.ErrorCorrectionService

```go
type ErrorCorrectionService interface {
    // Encode encodes data into shards using Reed-Solomon.
    Encode(ctx context.Context, data []byte, config *ShardConfiguration) (*ProtectedMessage, error)

    // Decode reconstructs original data from shards.
    Decode(ctx context.Context, shards []*Shard, config *ShardConfiguration) ([]byte, error)

    // VerifyShards checks shard integrity.
    VerifyShards(ctx context.Context, shards []*Shard) error
}
```

### media.Service

```go
type Service interface {
    // LoadMedia loads media data and creates a MediaAsset.
    LoadMedia(ctx context.Context, data []byte, format MediaFormat) (*MediaAsset, error)

    // DetectFormat automatically detects media format.
    DetectFormat(ctx context.Context, data []byte) (MediaFormat, error)

    // CalculateCapacity calculates embedding capacity.
    CalculateCapacity(ctx context.Context, asset *MediaAsset) (*CapacityInfo, error)

    // SanitizeMetadata removes all metadata.
    SanitizeMetadata(ctx context.Context, asset *MediaAsset) error

    // ValidateMedia validates media integrity.
    ValidateMedia(ctx context.Context, asset *MediaAsset) error

    // AnalyzeQuality analyzes quality score.
    AnalyzeQuality(ctx context.Context, asset *MediaAsset) (float64, error)
}
```

### stego.StegoService

```go
type StegoService interface {
    // Embed embeds data into cover media.
    Embed(ctx context.Context, data []byte, cover *media.MediaAsset, technique StegoTechnique) (*StegoContainer, error)

    // Extract extracts hidden data from stego media.
    Extract(ctx context.Context, stego *media.MediaAsset, technique StegoTechnique) ([]byte, error)

    // AnalyzeCapacity analyzes embedding capacity.
    AnalyzeCapacity(ctx context.Context, cover *media.MediaAsset, technique StegoTechnique) (*CapacityInfo, error)
}
```

---

## Required Fixes

### 1. Fix EmbedHandler.Handle() Method

**Location:** `stego_handlers.go:73-170`

**Issues:**
- Line 73: `ProcessMedia` doesn't exist → use `LoadMedia`
- Line 92: `Encrypt` signature mismatch → provide publicKey and algorithm
- Line 96: Type conversion → use `encryptedPayload.Data` to get []byte
- Line 104: Type conversion → create proper `ShardConfiguration` object
- Line 108: Type conversion → need to extract shards as []byte
- Line 131: `GetEmbeddedData` doesn't exist → need proper method
- Line 160: `GetScore` doesn't exist → quality is likely already float64

**Fix Strategy:**
1. Replace `ProcessMedia` with `LoadMedia(ctx, coverData, format)`
2. Add key generation/loading for crypto operations
3. Create `ShardConfiguration` from redundancy float
4. Extract shard data properly from `ProtectedMessage`
5. Use correct stego container methods
6. Fix quality score extraction

### 2. Fix ExtractHandler.Handle() Method

**Location:** `stego_handlers.go:200-280`

**Issues:**
- Line 247: `Decode` signature mismatch → provide []*Shard and config
- Line 255: Type mismatch → payload is []byte, need *CryptoPayload

**Fix Strategy:**
1. Extract shards from embedded data
2. Create `ShardConfiguration` for decoding
3. Wrap decrypted data in `CryptoPayload` structure
4. Use proper type conversions

### 3. Add Missing Domain Methods

**Required additions:**
- `StegoContainer.Data()` or proper extraction method
- Quality score should be simple float64, not object with GetScore()

---

## Implementation Priority

### Phase 1: Critical Path (Enable Build) ⚠️ HIGH PRIORITY
1. Fix all type conversions in handlers
2. Add proper ShardConfiguration creation
3. Fix crypto service call signatures
4. Remove or use `coverAsset` variable
5. **Goal:** Achieve clean `go build`

### Phase 2: Domain Completeness
1. Implement missing domain service methods
2. Add proper error handling
3. Complete stego service integration
4. **Goal:** Handlers fully functional (no stubs)

### Phase 3: CLI Integration
1. Wire up real services in CLI initialization
2. Replace simulation with backend calls
3. Test end-to-end with real files
4. **Goal:** Working CLI binary with real steganography

---

## Next Steps

### Immediate (Today)
1. ✅ **Fix interface naming** (COMPLETED)
2. 🎯 **Fix EmbedHandler type conversions** (NEXT)
3. 🎯 **Fix ExtractHandler type conversions**
4. 🎯 **Achieve clean build**

### Short-term (This Week)
1. Complete handler implementations
2. Wire up services in CLI
3. Test with real files
4. Document integration process

### Medium-term (Next Week)
1. Replace CLI simulation with real backend
2. Performance testing
3. Error handling improvements
4. Integration tests

---

## Success Criteria

- ✅ Interface naming fixed (ErrorCorrectionService, media.Service)
- [ ] `go build -o bin/shadowforge ./cmd/cli/` succeeds
- [ ] All handler methods compile without errors
- [ ] CLI can run basic embed/extract operations
- [ ] Integration tests pass
- [ ] No panic or runtime errors
- [ ] Memory management verified (no leaks)

---

## Notes

- **CLI Simulation:** Fully functional, all 7 tests passed
- **Backend Status:** Interface errors fixed, implementation errors remain
- **Blocker:** Type conversions and method signatures preventing build
- **Strategy:** Fix type issues first, then implement missing methods
- **Timeline:** 1-2 days for clean build, 3-5 days for full integration

---

**Last Updated:** December 12, 2025
**Next Review:** After achieving clean build

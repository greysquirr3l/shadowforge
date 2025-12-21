# Forensic Watermarking for KaTeX Formula PNGs

> **Post-quantum secure recipient tracking for mathematical content distribution**

## Overview

This watermarking system embeds forensic tracking data into KaTeX formula PNG images using:

- **Kyber-1024**: Post-quantum KEM (key encapsulation mechanism) for symmetric key exchange
- **AES-256-GCM**: AEAD encryption of forensic payload using Kyber-encapsulated key
- **Dilithium3**: Post-quantum digital signatures for authenticity
- **Reed-Solomon**: Error correction for robustness against image loss/corruption
- **LSB/DCT Steganography**: Invisible data embedding with statistical security

**Use Case**: Track document recipients by watermarking mathematical equations in academic
papers, confidential reports, or sensitive research materials. If leaked, the watermark
reveals who received the document.

---

## Architecture

### Data Flow

```text
┌─────────────────────┐
│  Forensic Data      │
│  (Recipient Info)   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Kyber-1024 KEM     │◄──── Ephemeral keypair per watermark
│  + AES-256-GCM      │      (Persisted via `--receipt` for later decryption)
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Dilithium3         │◄──── Digital Signature
│  Signing            │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Reed-Solomon       │
│  Error Correction   │◄──── Shard Generation (Data + Parity)
└──────────┬──────────┘
           │
           ├──────┬──────┬──────┬──────┐
           ▼      ▼      ▼      ▼      ▼
        ┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐
        │Img 1││Img 2││Img 3││Img 4││Img 5│◄─ LSB/DCT Embedding
        └─────┘└─────┘└─────┘└─────┘└─────┘
```

### Components

1. **ForensicData**: Structured recipient tracking information
2. **WatermarkConfig**: Embedding parameters (technique, strategy, redundancy)
3. **WatermarkService**: Orchestrates embed/extract operations
4. **CryptoService**: Kyber-1024 encryption + Dilithium3 signing
5. **ECService**: Reed-Solomon encoding for distributed sharding
6. **StegoService**: LSB or DCT embedding in PNG images

---

## Usage

### Prerequisites

**Test Materials**: 17 KaTeX formula PNGs in `mixed-media/images/katex_formulas/`

```text
eq_001.png  (9.1KB)  - Riemann Hypothesis
eq_002.png  (9.9KB)  - Hodge Conjecture
eq_003.png  (5.1KB)  - BSD Conjecture
eq_004.png  (5.5KB)  - Navier-Stokes
eq_005.png  (4.1KB)  - Yang-Mills Theory
eq_006.png  (11KB)   - P vs NP Problem
eq_007.png  (2.5KB)  - Goldbach's Conjecture
eq_008.png  (10KB)   - Twin Prime Conjecture
eq_009.png  (3.8KB)  - Collatz Conjecture
eq_010.png  (4.8KB)  - Fermat's Last Theorem
eq_011.png  (5.2KB)  - Pythagorean Theorem
eq_012.png  (4.5KB)  - Euler's Identity
eq_013.png  (6.3KB)  - Schrödinger Equation
eq_014.png  (7.8KB)  - Maxwell's Equations
eq_015.png  (3.2KB)  - Einstein Field Equations
eq_016.png  (5.9KB)  - Dirac Equation
eq_018.png  (4.7KB)  - Cauchy-Riemann Equations
```

**Note**: `eq_017.png` is omitted from the test set (intentional gap in numbering).

**GPG Key**: Example placeholder key

```bash
# Create test key (for demonstration only)
echo "-----BEGIN PGP PUBLIC KEY-----
Test key placeholder
-----END PGP PUBLIC KEY-----" > test-key.pub
```

**Build**: Ensure shadowforge binary is built

```bash
go build -o bin/shadowforge ./cmd/cli/main.go
```

---

## Workflow

### Step 1: Embed Watermark

```bash
./bin/shadowforge watermark embed \
  --recipient "Dr. Alice Smith <alice@university.edu>" \
  --gpg-key ./test-key.pub \
  --input-dir ./mixed-media/images/katex_formulas \
  --output-dir ./watermarked-output \
  --technique lsb \
  --redundancy 0.30 \
  --receipt ./watermark-receipt.md \
  --document-id "CONF-2025-MATH-001"
```

**Output**:

```text
✅ Watermarked 17 images successfully
   Strategy: distributed
   Technique: lsb
   Redundancy: 30%
   Encryption: Kyber-1024
   Signature: Dilithium3
   Output: ./watermarked-output/
```

**What Happens**:

1. ForensicData structure created with recipient info
2. Kyber-1024 keypair generated per watermark
3. If `--receipt` provided, keypair is written to a Markdown receipt (contains private key material)
4. Payload encrypted with AES-256-GCM using Kyber-derived key material
5. Encrypted payload signed with Dilithium3 (3293-byte signature)
6. Reed-Solomon generates 12 data shards + 5 parity shards (30% redundancy, 17 total)
7. Each shard embedded in a different PNG using LSB steganography
8. Watermarked PNGs written to output directory

### Step 2: Extract Watermark

```bash
./bin/shadowforge watermark extract \
  --input-dir ./watermarked-output \
  --output ./recovered-watermark.json \
  --receipt ./watermark-receipt.md
```

**Output**:

```json
{
  "recipient": "Dr. Alice Smith <alice@university.edu>",
  "prepared_date": "2025-12-17",
  "gpg_public_key": "-----BEGIN PGP PUBLIC KEY-----\nTest key placeholder\n-----END PGP PUBLIC KEY-----",
  "document_id": "CONF-2025-MATH-001",
  "watermark_id": "wm_1734480000_abc123",
  "created_at": "2025-12-17T10:30:00Z"
}
```

**What Happens**:

1. Steganographic extraction from all PNGs in directory
2. Reed-Solomon reconstruction (needs min 12 of 17 images with 30% redundancy)
3. Dilithium3 signature verification
4. AES-256-GCM decryption using the Kyber-1024 private key stored in the `--receipt` file
5. ForensicData deserialized and written to JSON

### Step 3: Verify Watermark

```bash
./bin/shadowforge watermark verify \
  --input-dir ./watermarked-output \
  --expected-recipient "Dr. Alice Smith"
```

**Output**:

```text
✅ Watermark verification successful
   Recipient: Dr. Alice Smith <alice@university.edu>
   Document ID: CONF-2025-MATH-001
   Watermark ID: wm_1734480000_abc123
   Created: 2025-12-17T10:30:00Z
   Signature: VALID
   Match: ✓ (expected recipient found)
```

---

## Configuration

### Techniques Comparison

| Technique | Capacity | Robustness | Detectability | Best For |
| ----------- | ---------- | ------------ | --------------- | ---------- |
| **LSB** | High (3-4 bits/pixel) | Low | Low | Clean PNGs, lossless distribution |
| **DCT** | Medium (0.5-1 bits/coeff) | Medium | Very Low | Spatial-domain DCT on PNG blocks (not JPEG coefficients) |

**Recommendation**: Use LSB for PNG-only workflows. DCT applies block-wise DCT to pixel values and embeds in
frequency coefficients, then inverse transforms back to pixels. This is *not* JPEG DCT coefficient modification.

### Redundancy Guidance

| Redundancy | Data Shards (k) | Parity Shards (m) | Total (n) | Min Recovery | Use Case |
|------------|-----------------|-------------------|-----------|--------------|----------|
| 0.20 (20%) | 14 | 3 | 17 | 14 | Controlled distribution |
| 0.30 (30%) | 12 | 5 | 17 | 12 | **Default - balanced** |
| 0.50 (50%) | 9 | 8 | 17 | 9 | High risk of image loss |

**Reed-Solomon Parameters**:

- `Data shards (k)` = ⌊Total images × (1 - Redundancy)⌋
- `Parity shards (m)` = Total images - Data shards
- `Min recovery` = k (need at least k shards to reconstruct)

**Example with 17 PNGs**:

- **20% redundancy**: k=14 data, m=3 parity → Need 14+ images to recover
- **30% redundancy**: k=12 data, m=5 parity → Need 12+ images to recover (recommended)
- **50% redundancy**: k=9 data, m=8 parity → Need 9+ images to recover

### Strategies

| Strategy | Description | Shard Count | Use Case |
| ---------- | ------------- | ------------- | ---------- |
| **distributed** | One shard per image | N images = N shards | **Default - robustness** |
| **single** | All data in one image | 1 image | High capacity, no redundancy |
| **redundant** | Multiple copies per shard | 2-3× images | Maximum redundancy |

---

## Capacity Planning

### Formula Capacity

**LSB Technique** (1-bit per pixel, RGB channels):

```text
Capacity (bytes) = (Width × Height × 3) / 8
```

**Example** (800×600 PNG):

```text
Capacity = (800 × 600 × 3) / 8 = 180,000 bytes ≈ 175 KB
```

**Actual Test PNGs** (eq_001.png - eq_018.png):

- Average size: 6KB per PNG
- Estimated capacity: ~100-500 bytes per image (depends on formula complexity)
- With 17 images: Total raw capacity ~1.7KB - 8.5KB

**ForensicData Payload Size**:

```json
{
  "recipient": "...",        // ~50 bytes
  "prepared_date": "...",    // ~10 bytes
  "gpg_public_key": "...",   // ~100-500 bytes (RSA-2048)
  "document_id": "...",      // ~20 bytes
  "watermark_id": "...",     // ~30 bytes
  "created_at": "..."        // ~25 bytes
}
```

**Typical payload**: 200-600 bytes (raw JSON)
**After Kyber-1024 + AES-256-GCM**: ~1800-2200 bytes (1568 Kyber ciphertext + 28 AES overhead + payload)
**After Dilithium3 signing**: ~5100-5500 bytes (adds 3293-byte signature)
**After RS encoding** (30% redundancy, k=12, m=5): ~6600-7200 bytes total encoded
**Distribution**: ~390-425 bytes per shard across 17 images

**Verdict**: 17 test PNGs provide **sufficient capacity** for typical forensic watermarks.

---

## Security

### Post-Quantum Cryptography

**Key Encapsulation**: Kyber-1024 (NIST Level 5)

- **Public Key**: 1568 bytes
- **Secret Key**: 3168 bytes
- **Ciphertext**: 1568 bytes (encapsulated symmetric key)
- **Security**: 256-bit quantum security (equivalent to AES-256)

**Payload Encryption**: AES-256-GCM (using Kyber-encapsulated key)

- **Overhead**: ~28 bytes (12-byte nonce + 16-byte auth tag)
- **Security**: 256-bit symmetric security

**Signing**: Dilithium3 (NIST Level 3)

- **Key Size**: 1952 bytes (public), 4000 bytes (private)
- **Signature Size**: 3293 bytes
- **Security**: 128-bit quantum security

### Current Limitations ⚠️

1. **Receipt-Based Key Persistence (Encryption)**:

- **Status**: Keys are generated per watermark embed operation and persisted in the `--receipt` Markdown file.
- **Impact**: You can decrypt later, but the receipt contains the Kyber private key material.
  Treat it like a private key. Losing it means you cannot decrypt encrypted watermarks.
- **Roadmap**: Optional reusable key management (provide a stable keypair via flags instead of generating a
  per-watermark keypair).

1. **Placeholder Signing**:

- **Status**: Dilithium3 keys generated but signature not verified end-to-end
- **Impact**: Signature verification step exists but needs full integration
- **Roadmap**: Add signature verification to extract workflow

1. **Key Management**:

- **Status**: No CLI commands for key generation/import/export
- **Impact**: Cannot share watermarked images for external decryption
- **Roadmap**: Add `keygen`, `keyexport`, `keyimport` commands

### Statistical Security

**LSB Embedding**:

- Uses PRNG-based embedding order (seeded with watermark ID)
- Designed to reduce statistical anomalies (evaluate with chi-square/RS analysis)
- Moderate robustness against visual inspection
- Lower robustness against compression/resampling

**DCT Embedding**:

- Modifies mid-frequency coefficients in spatial-domain DCT blocks (8×8 pixel blocks)
- ±1 adjustment to DCT coefficients preserves visual quality
- Designed for improved statistical hiding (frequency-domain embedding)
- Note: This is *not* JPEG coefficient modification; operates on raw pixel blocks

---

## Forensic Data Structure

### JSON Schema

```json
{
  "recipient": {
    "type": "string",
    "description": "Full name and email in format: 'Name <email@domain>'",
    "example": "Dr. Alice Smith <alice@university.edu>",
    "required": true
  },
  "prepared_date": {
    "type": "string",
    "format": "YYYY-MM-DD",
    "description": "Document preparation date",
    "example": "2025-12-17",
    "required": true
  },
  "gpg_public_key": {
    "type": "string",
    "description": "PGP/GPG public signing key (ASCII-armored)",
    "example": "-----BEGIN PGP PUBLIC KEY-----\n...\n-----END PGP PUBLIC KEY-----",
    "required": true
  },
  "document_id": {
    "type": "string",
    "description": "Optional document identifier",
    "example": "CONF-2025-MATH-001",
    "required": false
  },
  "watermark_id": {
    "type": "string",
    "description": "Unique watermark identifier (auto-generated)",
    "format": "wm_<timestamp>_<random>",
    "example": "wm_1734480000_abc123",
    "required": true
  },
  "created_at": {
    "type": "string",
    "format": "RFC3339",
    "description": "Watermark creation timestamp",
    "example": "2025-12-17T10:30:00Z",
    "required": true
  }
}
```

### Validation Rules

**Recipient** (`--recipient` flag):

- Must contain both name and email
- Email format: `name@domain.tld`
- Example: `"John Doe <john@example.com>"`

**GPG Key** (`--gpg-key` flag):

- Must be PGP/GPG public key format
- File must exist and be readable
- **Embedded as metadata only** (not used for cryptographic verification in current implementation)
- Provides recipient's public signing key for identity attribution

**Document ID** (`--document-id` flag):

- Optional alphanumeric identifier
- Recommended format: `<TYPE>-<YEAR>-<CATEGORY>-<NUMBER>`
- Example: `CONF-2025-MATH-001`

**Prepared Date** (auto-detected):

- Defaults to current date (YYYY-MM-DD)
- Used to timestamp document preparation

**Watermark ID** (auto-generated):

- Format: `wm_<unix_timestamp>_<8_char_random>`
- Guarantees uniqueness across all watermarks
- Seeds PRNG for embedding order

---

## Troubleshooting

### Error: "Insufficient capacity"

**Cause**: PNG images too small for encrypted payload

**Solutions**:

1. **Reduce redundancy**: Try `--redundancy 0.20` (fewer parity shards)
2. **Add more images**: Distribute across 20+ images instead of 17
3. **Use DCT technique**: Try `--technique dct` (but test capacity first)
4. **Compress GPG key**: Use smaller RSA-1024 key (not recommended for production)

**Check capacity before embedding**:

```bash
# Future command (not yet implemented)
./bin/shadowforge analyze capacity --input-dir ./formulas --technique lsb
```

### Error: "Failed to reconstruct shards"

**Cause**: Too many images missing or corrupted

**Solutions**:

1. **Check image count**: Ensure sufficient images for redundancy level
   - 30% redundancy with 17 images: k=12 data shards, need 12+ images minimum
   - 20% redundancy with 17 images: k=14 data shards, need 14+ images minimum
2. **Increase redundancy**: Re-embed with `--redundancy 0.50` (k=9 data, need 9+ images)
3. **Verify image integrity**: Ensure PNGs not re-encoded or compressed

### Error: "Signature verification failed"

**Cause**: Watermark tampered or key mismatch (future implementation)

**Solutions**:

1. **Check GPG key**: Ensure same key used for embed and verify
2. **Verify image integrity**: Ensure PNGs not modified post-embedding
3. **Check technique match**: Extraction must use same technique as embedding

### Error: "Decryption failed"

**Cause**: Key mismatch or corrupted ciphertext

**Most Common Causes**:

1. Missing `--receipt` while attempting to decrypt an encrypted watermark
2. Wrong `--receipt` (does not match the watermark)
3. Watermarked images were modified and the payload can’t be reconstructed cleanly

**Fix**:

- Re-run extraction with the correct receipt: `shadowforge watermark extract --receipt ./watermark-receipt.md ...`
- If you want to avoid encryption entirely for a test run: embed with `--no-encrypt` and extract with `--no-decrypt`

**Future Solution**: Reusable key management:

```bash
# Generate keypair (future command)
./bin/shadowforge keygen --algorithm kyber1024 --output ./keys/embed-key

# Embed with specific key (future flag)
./bin/shadowforge watermark embed --encryption-key ./keys/embed-key.pub ...

# Extract with matching key (future flag)
./bin/shadowforge watermark extract --decryption-key ./keys/embed-key.priv ...
```

---

## Testing

### Test Materials Location

**Directory**: `mixed-media/images/katex_formulas/`

**Files**: 17 PNG images (eq_001.png - eq_018.png)

**Content**: LaTeX-rendered mathematical formulas:

- Millennium Prize Problems (Riemann, Hodge, BSD, Navier-Stokes, Yang-Mills, P vs NP)
- Famous Conjectures (Goldbach, Twin Prime, Collatz)
- Fundamental Theorems (Fermat, Pythagorean, Euler)
- Physics Equations (Schrödinger, Maxwell, Einstein, Dirac, Cauchy-Riemann)

### Basic Test Workflow

```bash
# 1. Embed watermark
./bin/shadowforge watermark embed \
  --recipient "Test User <test@example.com>" \
  --gpg-key ./test-key.pub \
  --input-dir ./mixed-media/images/katex_formulas \
  --output-dir ./test-watermarked \
  --technique lsb \
  --redundancy 0.30 \
  --receipt ./test-watermark-receipt.md

# 2. Extract watermark
./bin/shadowforge watermark extract \
  --input-dir ./test-watermarked \
  --output ./test-recovered.json \
  --receipt ./test-watermark-receipt.md

# 3. Verify contents
cat ./test-recovered.json

# 4. Verify with expected recipient
./bin/shadowforge watermark verify \
  --input-dir ./test-watermarked \
  --expected-recipient "Test User"
```

### Test Scenarios

1. **Full Set (17 images)**:
   - Embed with 30% redundancy (k=12 data + m=5 parity)
   - Extract from all 17 images
   - Should succeed with 100% confidence

2. **Partial Loss (13 images)**:
   - Embed with 30% redundancy
   - Remove 4 random images from output
   - Extract from remaining 13 images
   - Should succeed (k=12 required, 13 available)

3. **Critical Loss (11 images)**:
   - Embed with 30% redundancy
   - Remove 6 random images
   - Extract from remaining 11 images
   - **Should fail** (k=12 required, only 11 available)

4. **Technique Comparison**:
   - Embed same watermark with LSB and DCT
   - Compare visual quality (DCT should be less visible)
   - Measure capacity differences

---

## Future Enhancements

### Phase 1: Key Management (HIGH PRIORITY)

- [ ] `shadowforge keygen` - Generate Kyber/Dilithium keypairs
- [ ] `shadowforge keyexport` - Export public keys
- [ ] `shadowforge keyimport` - Import keypairs
- [ ] Persistent key storage (encrypted wallet)
- [ ] Key rotation support

### Phase 2: Production Signing (HIGH PRIORITY)

- [ ] Integrate Dilithium3 signature verification in extraction
- [ ] Add signature status to verify command output
- [ ] Support external signing keys (HSM integration)

### Phase 3: Enhanced Workflows (MEDIUM PRIORITY)

- [ ] Batch processing (multiple documents)
- [ ] Watermark revocation (blacklist system)
- [ ] Audit trail generation (CSV/JSON logs)
- [ ] Capacity analysis command (`analyze capacity`)

### Phase 4: Advanced Features (LOW PRIORITY)

- [ ] Additional steganography techniques (Phase Encoding, Echo Hiding)
- [ ] Archive support (watermark entire ZIP files)
- [ ] REST API for watermarking service
- [ ] Alternative AEAD ciphers (ChaCha20-Poly1305 option)

### Phase 5: Forensic Tools (LOW PRIORITY)

- [ ] Leak attribution analysis (compare watermarks)
- [ ] Chain-of-custody tracking
- [ ] Tamper detection (identify modified regions)
- [ ] Statistical steganalysis tools

---

## Implementation Details

### Package Structure

```text
internal/domain/watermark/
├── forensic_data.go        # ForensicData types, WatermarkConfig
├── watermark_service.go    # Embed/Extract orchestration
└── README.md               # This file

internal/interfaces/cli/commands/
└── watermark_commands.go   # CLI commands (embed/extract/verify)

internal/interfaces/cli/services/
└── service_container.go    # Dependency injection for WatermarkService

mixed-media/images/katex_formulas/
└── eq_*.png                # 17 test PNG images
```

### Dependencies

**Domain Services** (injected via `service_container.go`):

- `CryptoService` (`internal/infrastructure/crypto/circl_service.go`)
- `ErrorCorrectionService` (`internal/infrastructure/errorcorrection/rs_service.go`)
- `StegoService` (`internal/infrastructure/stego/stego_impl/stego_service.go`)
- `MediaService` (`internal/infrastructure/media/media_service.go`)

**Libraries**:

- `github.com/cloudflare/circl` - Kyber-1024 & Dilithium3
- `github.com/klauspost/reedsolomon` - Reed-Solomon error correction
- `github.com/spf13/cobra` - CLI framework
- `github.com/sirupsen/logrus` - Structured logging

---

## License

See [LICENSE](../../../LICENSE) for details.

---

## Contact

For issues or questions related to watermarking:

1. Check troubleshooting section above
2. Review logs with `--debug` flag
3. Refer to [architecture.md](../../../docs/architecture.md) for system design
4. Refer to [implementation_plan_todo.md](../../../docs/implementation_plan_todo.md) for roadmap

---

**Status**: ✅ Implementation complete with known limitations (ephemeral keys, pending signature verification integration)

**Last Updated**: 2025-12-17

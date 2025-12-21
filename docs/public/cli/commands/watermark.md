# Watermark Commands

Add invisible digital watermarks to media files for ownership verification and integrity checking.

## Overview

Watermarking embeds a unique identifier or signature into media that:
- Persists even if the file is compressed or modified
- Is imperceptible to human perception
- Can be used to prove ownership or authenticity
- Survives common image manipulations

Unlike steganography (which hides secret data), watermarks are typically:
- Known to exist but difficult to remove
- Used for copyright and authentication
- Resilient to attacks and transformations
- Legally binding in ownership disputes

## Basic Watermarking

### `shadowforge watermark add`

Add a watermark to image or audio files.

**Usage:**
```bash
shadowforge watermark add [flags]
```

**Flags:**
```
  --input FILE              Input media file (required)
  --output FILE             Output watermarked file (required)
  --mark STRING             Watermark text/ID (required)
  --type TYPE               Watermark type (text, signature, hash, uuid)
  --method METHOD           Embedding method (lsb, dct, spread-spectrum)
  --strength LEVEL          Watermark strength (1-100, default: 50)
  --key FILE                Encryption key for watermark (optional)
  --invisible               Make watermark imperceptible (default: true)
  --robust                  Add redundancy for robustness (default: true)
  --format FORMAT           Output media format (auto, png, jpg, wav)
  --quality PERCENT         Output quality (1-100, default: 90)
  --verify                  Verify watermark was embedded (default: true)
  -v, --verbose             Enable verbose output
  --output-format FORMAT    Output format (text, json, yaml)
```

**Watermark Types:**

- `text` - Human-readable text watermark
  - Example: "Copyright 2025 John Doe"
  - Length: up to 100 characters
  - Best for: Simple identification

- `signature` - Digital signature watermark
  - Uses private key to sign watermark
  - Cryptographically verifiable
  - Best for: Legal authentication

- `hash` - Hash-based watermark
  - Watermark is hash of content + secret key
  - Proves file authenticity
  - Best for: Integrity verification

- `uuid` - Unique identifier watermark
  - Auto-generated or provided UUID
  - Tracks individual file copies
  - Best for: Copy tracking

**Embedding Methods:**

- `lsb` - Least significant bit (visible media only)
  - Fast, simple
  - Visible with statistical analysis
  - Best for: Quick watermarking

- `dct` - Discrete cosine transform (JPEG only)
  - More robust than LSB
  - Survives JPEG recompression
  - Best for: JPEG watermarking

- `spread-spectrum` - Spread spectrum (audio only)
  - Highly robust against attacks
  - Imperceptible to human hearing
  - Best for: Audio watermarking

**Examples:**

Add text watermark to image:
```bash
shadowforge watermark add \
  --input photo.jpg \
  --output watermarked.jpg \
  --mark "Copyright 2025 Acme Corp" \
  --type text \
  --strength 50
```

Add signature watermark with private key:
```bash
shadowforge watermark add \
  --input document.png \
  --output authenticated.png \
  --mark "Legal Document" \
  --type signature \
  --key private.key \
  --strength 75 \
  --robust
```

Add UUID watermark for copy tracking:
```bash
shadowforge watermark add \
  --input product.png \
  --output copy-001.png \
  --mark "550e8400-e29b-41d4-a716-446655440000" \
  --type uuid \
  --invisible \
  --verify
```

Add watermark to audio file:
```bash
shadowforge watermark add \
  --input music.wav \
  --output watermarked.wav \
  --mark "MyArtist" \
  --type text \
  --method spread-spectrum \
  --strength 60
```

High-strength watermark with JSON output:
```bash
shadowforge watermark add \
  --input image.png \
  --output secure.png \
  --mark "Classified Document" \
  --type signature \
  --key sign.key \
  --strength 100 \
  --robust \
  --verify \
  --output-format json \
  --verbose
```

## Watermark Extraction

### `shadowforge watermark extract`

Extract and verify watermarks from watermarked files.

**Usage:**
```bash
shadowforge watermark extract [flags]
```

**Flags:**
```
  --input FILE              Watermarked media file (required)
  --type TYPE               Expected watermark type (text, signature, hash, uuid)
  --key FILE                Verification key (optional, for signature watermarks)
  --tolerance INT           Extraction tolerance (0-100, default: 10)
  --method METHOD           Extraction method (auto, lsb, dct, spread-spectrum)
  --verbose                 Show extraction details
  --output-format FORMAT    Output format (text, json, yaml)
```

**Examples:**

Extract text watermark:
```bash
shadowforge watermark extract \
  --input watermarked.jpg \
  --type text \
  --verbose
```

Extract and verify signature watermark:
```bash
shadowforge watermark extract \
  --input authenticated.png \
  --type signature \
  --key public.key \
  --output-format json
```

Extract UUID watermark:
```bash
shadowforge watermark extract \
  --input copy-001.png \
  --type uuid
```

Extract from audio with detailed output:
```bash
shadowforge watermark extract \
  --input watermarked.wav \
  --type text \
  --method spread-spectrum \
  --verbose
```

## Watermark Verification

### `shadowforge watermark verify`

Verify watermark presence and integrity without extracting.

**Usage:**
```bash
shadowforge watermark verify [flags]
```

**Flags:**
```
  --input FILE              Watermarked media file (required)
  --expected STRING         Expected watermark value (optional)
  --key FILE                Verification key (optional)
  --type TYPE               Watermark type (optional, for validation)
  --tolerance INT           Verification tolerance (default: 10)
  --report                  Generate detailed report (default: true)
  --output-format FORMAT    Output format (text, json, yaml)
```

**Output:**

For `text` or `uuid` watermarks:
```json
{
  "present": true,
  "match": true,
  "confidence": 95,
  "watermark": "Copyright 2025",
  "method": "lsb",
  "strength": 50
}
```

For signature watermarks:
```json
{
  "present": true,
  "signature_valid": true,
  "confidence": 99,
  "signer": "organization@example.com",
  "timestamp": "2025-12-21T10:30:00Z",
  "method": "dct"
}
```

**Examples:**

Quick verification:
```bash
shadowforge watermark verify --input watermarked.jpg
```

Verify expected watermark:
```bash
shadowforge watermark verify \
  --input document.png \
  --expected "Copyright 2025 Acme" \
  --type text
```

Verify signature with report:
```bash
shadowforge watermark verify \
  --input authenticated.png \
  --key public.key \
  --type signature \
  --report \
  --output-format json
```

## Batch Watermarking

### `shadowforge watermark batch-add`

Add the same watermark to multiple files.

**Usage:**
```bash
shadowforge watermark batch-add [flags]
```

**Flags:**
```
  --input FILES             Input files (comma-separated or glob)
  --output-dir DIR          Output directory (required)
  --mark STRING             Watermark text/ID (required)
  --type TYPE               Watermark type (default: text)
  --method METHOD           Embedding method (default: auto)
  --strength LEVEL          Watermark strength (default: 50)
  --preserve-structure      Keep original directory structure
  --parallel                Use parallel processing (default: true)
  --workers INT             Number of workers (default: CPU count)
  --skip-errors             Continue on errors (default: false)
  --verbose                 Verbose output
```

**Examples:**

Watermark all PNG files:
```bash
shadowforge watermark batch-add \
  --input "*.png" \
  --output-dir watermarked/ \
  --mark "Copyright 2025" \
  --type text
```

Watermark with parallel processing:
```bash
shadowforge watermark batch-add \
  --input photos/*.jpg \
  --output-dir watermarked/ \
  --mark "batch-001" \
  --type uuid \
  --workers 4 \
  --parallel
```

Watermark with structure preservation:
```bash
shadowforge watermark batch-add \
  --input "media/**/*.png" \
  --output-dir watermarked/ \
  --mark "Licensed" \
  --preserve-structure \
  --strength 75
```

## Batch Verification

### `shadowforge watermark batch-verify`

Verify watermarks in multiple files.

**Usage:**
```bash
shadowforge watermark batch-verify [flags]
```

**Flags:**
```
  --input FILES             Input files (comma-separated or glob)
  --expected STRING         Expected watermark (optional)
  --type TYPE               Watermark type (optional)
  --generate-report         Create verification report (default: true)
  --report-file FILE        Report output file
  --parallel                Use parallel processing (default: true)
  --output-format FORMAT    Output format (text, json, yaml)
```

**Examples:**

Verify all watermarked files:
```bash
shadowforge watermark batch-verify \
  --input watermarked/*.png \
  --generate-report \
  --report-file verification-report.json
```

Verify with expected watermark:
```bash
shadowforge watermark batch-verify \
  --input "*.jpg" \
  --expected "Copyright 2025" \
  --type text \
  --parallel
```

## Use Cases

### Copyright Protection
```bash
# Add copyright watermark to all product images
shadowforge watermark batch-add \
  --input "product-images/*.jpg" \
  --output-dir protected/ \
  --mark "© 2025 Company Name" \
  --type text \
  --strength 60

# Later, verify ownership
shadowforge watermark verify --input protected/image1.jpg
```

### Copy Tracking
```bash
# Assign unique ID to each licensed copy
for i in {1..100}; do
  uuid=$(uuidgen)
  shadowforge watermark add \
    --input template.png \
    --output "copy-$i.png" \
    --mark "$uuid" \
    --type uuid
done

# Track which copy was leaked
shadowforge watermark extract --input leaked-copy.png
```

### Document Authentication
```bash
# Sign document with private key
shadowforge watermark add \
  --input document.pdf.png \
  --output authenticated.png \
  --mark "Legal Department" \
  --type signature \
  --key /secure/private.key \
  --strength 100

# Verify authenticity with public key
shadowforge watermark verify \
  --input authenticated.png \
  --key /public/public.key \
  --type signature
```

### Audio Attribution
```bash
# Watermark music with artist name
shadowforge watermark add \
  --input original-song.wav \
  --output distributed.wav \
  --mark "Artist Name" \
  --type text \
  --method spread-spectrum \
  --strength 70

# Verify artist watermark
shadowforge watermark extract \
  --input distributed.wav \
  --type text
```

## Watermark Robustness

Watermarks can survive various media transformations:

| Transformation | LSB | DCT | Spread-Spectrum |
|---|---|---|---|
| Compression (90% quality) | ❌ | ✅ | ✅ |
| Cropping (25% removal) | ❌ | ⚠️ | ✅ |
| Rotation (small angles) | ❌ | ⚠️ | ✅ |
| Filtering (blur, sharpen) | ❌ | ⚠️ | ✅ |
| Format conversion | ❌ | ⚠️ | ✅ |
| Noise addition (Gaussian) | ⚠️ | ✅ | ✅ |

**Legend:**
- ✅ Survives transformation
- ⚠️ May survive, depends on extent
- ❌ Typically destroyed

For maximum robustness, use:
- `--method spread-spectrum` for audio
- `--method dct` for JPEG images
- `--strength 75-100` with `--robust` flag

## Security Considerations

### Watermark Keys
```bash
# Generate secure watermark key pair
openssl genrsa -out watermark-private.pem 2048
openssl rsa -in watermark-private.pem -pubout -out watermark-public.pem

# Use with signature watermarks
shadowforge watermark add \
  --input media.png \
  --output authenticated.png \
  --mark "Document ID" \
  --type signature \
  --key watermark-private.pem
```

### Watermark Secrecy
- **Text watermarks**: Mark is human-readable but method is hidden
- **Signature watermarks**: Both mark and method are protected by key
- **Hash watermarks**: Require knowledge of secret key to forge
- **UUID watermarks**: Unique per copy, good for tracking

### Watermark Attacks
Potential attacks and mitigations:

| Attack | Description | Mitigation |
|---|---|---|
| Removal | Strip watermark | Use robust method, high strength |
| Forgery | Add fake watermark | Use signature watermarks with keys |
| Spoofing | Claim false watermark | Use timestamp, digital signatures |
| Collusion | Compare multiple copies | Use unique UUID per copy |

For sensitive applications, combine watermarking with:
- Cryptographic signatures (see `watermark add --type signature`)
- Blockchain timestamping (future feature)
- Regular verification audits

## Performance

Typical performance on modern hardware:

| Operation | Image (1MB) | Audio (10MB) |
|---|---|---|
| Add watermark (LSB) | ~50ms | ~200ms |
| Add watermark (DCT) | ~100ms | N/A |
| Add watermark (Spread-spectrum) | N/A | ~500ms |
| Extract watermark | ~30ms | ~150ms |
| Verify watermark | ~20ms | ~100ms |
| Batch (100 files) | ~30s | ~120s |

Use `--parallel --workers 4` to process multiple files concurrently and reduce total time.

## Troubleshooting

### "Watermark not detected"
**Cause**: File was heavily compressed or damaged
**Solution**: Use `--tolerance 20` or higher when extracting

### "Extraction confidence low"
**Cause**: Watermark signal weak or corrupted
**Solution**: Re-embed with `--strength 75+` and `--robust`

### "Signature verification failed"
**Cause**: Wrong key or file was modified
**Solution**: Verify using correct public key, check file integrity

### "Method not supported for this file type"
**Cause**: Tried to use audio method on image
**Solution**: Use appropriate method (see flag options)

### "Output file larger than input"
**Cause**: Watermarking increases file size
**Solution**: Use compression: `--quality 85` for JPEG

## See Also

- [Embed Command](embed.md) - For data hiding
- [Extract Command](extract.md) - For data recovery
- [Security Best Practices](../security.md)
- [FAQ](../faq.md) - Frequently asked questions

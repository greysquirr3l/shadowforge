# Analysis Commands

Analyze media capacity, verify stego integrity, and plan embedding operations.

## Capacity Analysis

### `shadowforge analyze capacity`

Determine how much data can be safely embedded in carrier media.

**Usage:**
```bash
shadowforge analyze capacity [flags]
```

**Flags:**
```
  --input FILE              Media file to analyze (required)
  --payload-size INT        Optional: payload size in bytes
  --technique TECH          Specific technique (default: all)
  --quality INT             Quality level 1-100 (default: 90)
  --json                    JSON output format
  -v, --verbose             Verbose output
```

**Output:**
Returns capacity information for each supported technique:
- Raw capacity (maximum)
- Safe capacity (recommended)
- Capacity utilization estimate
- Technique recommendations

**Examples:**

Analyze a PNG image:
```bash
shadowforge analyze capacity --input photo.png

# Output:
# Media: photo.png (PNG, 3024x4032 pixels, 3.2 MB)
#
# Capacity by technique:
#   LSB:         847 KB (safe: 508 KB) - ✓ Excellent
#   DCT:         N/A (PNG not DCT-based)
#   Palette:     245 KB (safe: 147 KB) - ○ Good
#   Zero-Width:  15 KB - Text only
#
# Recommendation: LSB for maximum capacity, Palette for lower detection risk
```

Check if image can hold a specific payload:
```bash
shadowforge analyze capacity \
  --input image.png \
  --payload-size 100000

# Output:
# Payload size: 100 KB
# Available capacity: 847 KB
# ✓ Sufficient capacity for all techniques
```

Analyze multiple media files:
```bash
shadowforge analyze capacity \
  --input photo1.jpg,photo2.png,audio.wav \
  --payload-size 50000

# Output:
# Results for 3 files:
# photo1.jpg: 256 KB safe capacity
# photo2.png: 847 KB safe capacity ← Recommended
# audio.wav:  512 KB safe capacity
```

## Media Validation

### `shadowforge validate`

Check if a file contains hidden data and verify its integrity.

**Usage:**
```bash
shadowforge validate [flags]
```

**Flags:**
```
  --input FILE              Stego media file (required)
  --detailed                Show detailed analysis
  --expected-size INT       Expected payload size (optional)
  --technique TECH          Specific technique to validate
  --json                    JSON output format
  -v, --verbose             Verbose output
```

**Examples:**

Quick validation:
```bash
shadowforge validate --input stego.png

# Output:
# ✓ Valid stego media detected
# Technique: LSB
# Payload size: 1024 bytes
# Integrity: ✓ Verified
```

Detailed analysis:
```bash
shadowforge validate --input stego.png --detailed

# Output:
# ✓ Valid stego media
# Technique: LSB
# Estimated payload: 1024 bytes
# Carrier media: PNG (3024x4032)
# Quality metrics:
#   - Statistical distribution: Normal
#   - Error rate: 0% - ✓ Clean
#   - Detectability score: Low - ✓ Safe
```

Validate against expected size:
```bash
shadowforge validate \
  --input stego.png \
  --expected-size 1024

# Output:
# ✓ Media contains ~1024 bytes
# Status: Valid
```

## Format Information

### `shadowforge formats`

List all supported media formats and their capabilities.

**Usage:**
```bash
shadowforge formats [flags]
```

**Flags:**
```
  --category CAT            Filter by category (images, audio, text, archive)
  --technique TECH          Filter by technique
  --json                    JSON output
  -v, --verbose             Show detailed information
```

**Examples:**

List all supported formats:
```bash
shadowforge formats

# Output:
# Supported Media Formats
#
# Images:
#   PNG  - ✓ LSB, ✓ Palette
#   JPEG - ✓ DCT
#   BMP  - ✓ LSB
#   GIF  - ✓ Palette
#
# Audio:
#   WAV  - ✓ LSB, ✓ Phase, ✓ Echo
#
# Text:
#   TXT  - ✓ Zero-Width
#
# Archive:
#   ZIP, TAR, TAR.GZ
```

Show image format details:
```bash
shadowforge formats --category images --verbose

# Output:
# PNG
#   Extensions: .png
#   Techniques: LSB, Palette
#   Min size: 10x10 pixels
#   Capacity: ~1 bit/pixel (LSB)
#   Quality: Lossless
#   Recommended: Excellent for LSB
#
# JPEG
#   Extensions: .jpg, .jpeg
#   Techniques: DCT
#   Min size: 8x8 pixels (DCT block)
#   Capacity: ~0.5 bits/coefficient
#   Quality: Lossy
#   Recommended: Good for DCT
#   Note: Re-compression corrupts hidden data
```

## Detectability Analysis

### `shadowforge analyze detectability`

Analyze how detectable hidden data is in media.

**Usage:**
```bash
shadowforge analyze detectability [flags]
```

**Flags:**
```
  --input FILE              Stego media file (required)
  --technique TECH          Technique used (auto-detect or specify)
  --reference FILE          Original carrier for comparison
  --detailed                Show detailed analysis
  --json                    JSON output
```

**Examples:**

Analyze detectability:
```bash
shadowforge analyze detectability --input stego.png

# Output:
# Detectability Analysis
# Technique: LSB
# Risk level: LOW ✓
#
# Metrics:
#   Chi-square test: 0.95 (normal variation)
#   Statistical analysis: Passed
#   Visual inspection: Identical to original
#
# Verdict: Hidden data not detectable with standard tools
```

Compare against original:
```bash
shadowforge analyze detectability \
  --input stego.png \
  --reference original.png \
  --detailed

# Output:
# Comparison Analysis
# Embedding technique: LSB
#
# Visual difference: 0% - Identical
# Color histogram: 99.8% match - ✓ Good
# Frequency analysis: 99.7% match - ✓ Good
# Statistical properties: Similar - ✓ Good
#
# Detectability: Very low with standard analysis
```

## Security Assessment

### `shadowforge analyze security`

Comprehensive security assessment of embedded data.

**Usage:**
```bash
shadowforge analyze security [flags]
```

**Flags:**
```
  --input FILE              Stego or cover media file
  --with-extraction         Attempt extraction for verification
  --password PASSWORD       Password if needed for extraction
  --detailed                Detailed report
```

**Examples:**

Security assessment:
```bash
shadowforge analyze security --input stego.png

# Output:
# Security Assessment Report
#
# Embedding:
#   ✓ Encryption: Not verified (no extraction)
#   ✓ Technique: LSB
#   ✓ Detectability: Low
#
# Recommendations:
#   • Data is well-hidden with LSB
#   • Consider additional encryption for sensitive data
#   • Avoid re-compression of carrier
```

---

See [Getting Started](../getting-started.md) for workflow examples.

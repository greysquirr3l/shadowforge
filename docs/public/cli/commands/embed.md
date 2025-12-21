# Embed Commands

Embed secret data in carrier media using various steganographic techniques.

## Basic Embedding

### `shadowforge embed`

Embed a secret in a single carrier image (basic one-to-one embedding).

**Usage:**
```bash
shadowforge embed [flags]
```

**Flags:**
```
  --input FILE              Secret payload file (required)
  --cover FILE              Carrier media file (required)
  --output FILE             Output stego file (required)
  --technique TECH          Steganography technique (default: auto-detect)
  --password PASSWORD       Encrypt with password (optional)
  --preserve-metadata       Don't sanitize metadata (default: false)
  --quality PERCENT         Target quality level (1-100, default: 90)
  --verify                  Verify embedding success (default: false)
  -v, --verbose             Enable verbose output
  --output-format FORMAT    Output format (text, json, yaml)
```

**Supported Techniques:**
- `lsb` - Least significant bit (images)
- `dct` - Discrete cosine transform (JPEG)
- `palette` - Palette manipulation (GIF/PNG)
- `phase` - Phase encoding (audio)
- `echo` - Echo hiding (audio)
- `lsb-audio` - LSB for audio
- `zero-width` - Zero-width characters (text)
- `auto` - Detect automatically (default)

**Examples:**

Basic embedding with automatic technique selection:
```bash
shadowforge embed --input secret.pdf --cover image.png --output stego.png
```

Explicit technique selection:
```bash
shadowforge embed \
  --input secret.txt \
  --cover photo.jpg \
  --output stego.jpg \
  --technique dct
```

With password protection:
```bash
shadowforge embed \
  --input confidential.doc \
  --cover background.png \
  --output protected.png \
  --password "my-secure-passphrase"
```

Verbose output with JSON result:
```bash
shadowforge embed \
  --input secret.txt \
  --cover image.png \
  --output stego.png \
  --verbose \
  --output-format json \
  --verify
```

## Distributed Embedding

### `shadowforge embed-distributed`

Distribute a single secret across multiple carrier images with optional redundancy (one-to-many pattern).

**Usage:**
```bash
shadowforge embed-distributed [flags]
```

**Flags:**
```
  --input FILE              Secret payload (required)
  --cover FILES             Carrier images (comma-separated, required)
  --output-dir DIR          Output directory (required)
  --threshold INT           Minimum shards needed to recover (default: K=data-shards)
  --redundancy PERCENT      Redundancy level (0-50%, default: 30%)
  --shard-technique TECH    Technique for each shard (default: auto)
  --manifest               Generate manifest file (default: true)
  --parallel               Use parallel processing (default: true)
  --workers INT            Number of workers (default: CPU count)
  -v, --verbose            Enable verbose output
```

**How It Works:**
1. Your secret is encrypted
2. Error correction codes add redundancy
3. Data is split into shards
4. Each shard is embedded in a different carrier
5. A manifest file tracks which shard is where

**Recovery:** You need at least `--threshold` carriers to recover the secret. With redundancy, losing some carriers doesn't prevent recovery.

**Examples:**

Basic distribution across 5 images:
```bash
shadowforge embed-distributed \
  --input important-data.zip \
  --cover img1.png,img2.png,img3.png,img4.png,img5.png \
  --output-dir distributed-output/
```

With custom threshold (recoverable from any 3 of 5):
```bash
shadowforge embed-distributed \
  --input critical-data.zip \
  --cover carrier1.png,carrier2.png,carrier3.png,carrier4.png,carrier5.png \
  --output-dir distribution/ \
  --threshold 3 \
  --redundancy 40%
```

High redundancy for maximum resilience:
```bash
shadowforge embed-distributed \
  --input secret.doc \
  --cover *.png \
  --output-dir resilient/ \
  --redundancy 50% \
  --threshold 10
```

## Batch Embedding

### `shadowforge embed-batch`

Embed multiple secrets in a single carrier (many-to-one pattern).

**Usage:**
```bash
shadowforge embed-batch [flags]
```

**Flags:**
```
  --input FILES             Multiple payload files (comma-separated)
  --cover FILE              Single carrier image (required)
  --output FILE             Output stego file (required)
  --compress                Compress payloads before embedding (default: false)
  --encrypt                 Encrypt each payload (default: false)
  --index                   Create index file for selective extraction
  --technique TECH          Embedding technique (default: auto)
  -v, --verbose             Enable verbose output
```

**Index File:**
When `--index` is used, a JSON file is created mapping payload names to offsets.

**Examples:**

Embed 3 documents in one image:
```bash
shadowforge embed-batch \
  --input doc1.pdf,doc2.pdf,doc3.pdf \
  --cover container.png \
  --output stego.png \
  --index
```

With compression:
```bash
shadowforge embed-batch \
  --input file1.txt,file2.txt,file3.txt,file4.txt \
  --cover image.png \
  --output compressed-stego.png \
  --compress \
  --encrypt
```

## Matrix Distribution

### `shadowforge embed-matrix`

Distribute multiple secrets across multiple carriers (many-to-many pattern).

**Usage:**
```bash
shadowforge embed-matrix [flags]
```

**Flags:**
```
  --input FILES             Multiple payloads (comma-separated)
  --cover FILES             Multiple carriers (comma-separated)
  --output-dir DIR          Output directory (required)
  --allocation MODE         Distribution mode (default: optimized)
                            Options: round-robin, random, optimized, balanced
  --threshold INT           Shards per payload needed to recover
  --redundancy PERCENT      Error correction level (0-50%)
  --manifest               Create manifest (default: true)
  --parallel               Use parallel processing (default: true)
  -v, --verbose            Enable verbose output
```

**Allocation Modes:**
- `round-robin` - Sequential distribution
- `random` - Randomized allocation
- `optimized` - Capacity-aware optimization
- `balanced` - Even distribution

**Examples:**

Distribute 3 secrets across 6 carriers:
```bash
shadowforge embed-matrix \
  --input secret1.pdf,secret2.pdf,secret3.pdf \
  --cover img1.png,img2.png,img3.png,img4.png,img5.png,img6.png \
  --output-dir matrix-output/ \
  --allocation optimized
```

With high redundancy:
```bash
shadowforge embed-matrix \
  --input doc1.doc,doc2.doc \
  --cover carrier1.png,carrier2.png,carrier3.png,carrier4.png \
  --output-dir distribution/ \
  --allocation balanced \
  --redundancy 40% \
  --threshold 3
```

## Common Options

All embed commands support:

### Password Protection
```bash
# Interactive password prompt
shadowforge embed --input secret.txt --cover image.png --output stego.png --password

# Password in flag (not recommended - visible in shell history)
shadowforge embed --input secret.txt --cover image.png --output stego.png --password "passphrase"
```

### Metadata Handling
```bash
# Preserve original image metadata
shadowforge embed \
  --input secret.txt \
  --cover photo.png \
  --output stego.png \
  --preserve-metadata

# Default: sanitize metadata for security
shadowforge embed --input secret.txt --cover photo.png --output stego.png
```

### Output Formats
```bash
# Human-readable text (default)
shadowforge embed --input secret.txt --cover image.png --output stego.png --output-format text

# JSON output for scripting
shadowforge embed --input secret.txt --cover image.png --output stego.png --output-format json

# YAML output
shadowforge embed --input secret.txt --cover image.png --output stego.png --output-format yaml
```

## Troubleshooting

**Error: "Insufficient capacity"**
- Use a larger carrier image
- Use a smaller payload
- Check capacity: `shadowforge analyze capacity --input image.png --payload-size <bytes>`

**Error: "Unsupported media format"**
- Check supported formats: `shadowforge formats`
- Convert media to supported format

**Error: "Technique not suitable for media"**
- Let Shadowforge auto-detect: `--technique auto`
- Or choose a different technique

---

See [Extract Commands](./extract.md) to recover embedded secrets.

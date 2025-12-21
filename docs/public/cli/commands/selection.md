# Selection Commands

Intelligent cover media selection and analysis for steganography.

## Overview

Selection commands help you find, analyze, and automatically select optimal cover media for your embedding operations.

**Key Features:**
- Scan directories for suitable media
- Analyze capacity of available files
- Auto-select optimal carriers for payload size
- Get suggestions for missing capacity
- Support for glob patterns and recursive scanning

## Scanning Media

### `shadowforge scan`

Scan directories to inventory available cover media.

**Usage:**
```bash
shadowforge scan [flags]
```

**Flags:**
```
  --input-dir DIR          Directory to scan (required)
  --recursive              Recursively scan subdirectories
  --min-capacity SIZE      Minimum capacity threshold (e.g., 100KB)
  --format-filter TYPES    Filter by format (png, jpg, wav, etc.)
  --output-format FORMAT   Output format (text, json, yaml)
```

**Examples:**

Scan Pictures directory:
```bash
shadowforge scan --input-dir ~/Pictures
```

Scan recursively with minimum capacity filter:
```bash
shadowforge scan \
  --input-dir ~/Documents \
  --recursive \
  --min-capacity 100KB
```

Scan for specific formats only:
```bash
shadowforge scan \
  --input-dir ~/media \
  --format-filter "png,jpg" \
  --recursive
```

**Output Example:**
```
Scanning directory: ~/Pictures

Found 12 media files:

IMAGES:
  vacation.png (3024x4032)
    Capacity: 847 KB (safe: 508 KB)
    Techniques: LSB, DCT, Palette

  family.jpg (2560x1920)
    Capacity: 591 KB (safe: 354 KB)
    Techniques: LSB, DCT

  sunset.png (1920x1440)
    Capacity: 423 KB (safe: 254 KB)
    Techniques: LSB, DCT, Palette

AUDIO:
  music.wav (3:45 duration, 44.1kHz)
    Capacity: 1.2 MB (safe: 900 KB)
    Techniques: LSB, Phase, Echo

TEXT:
  document.txt (5 pages)
    Capacity: 500 bytes (safe: 300 bytes)
    Techniques: Zero-Width
```

## Selecting Optimal Media

### `shadowforge select`

Automatically select optimal cover media for a payload.

**Usage:**
```bash
shadowforge select [flags]
```

**Flags:**
```
  --input-dir DIR          Directory with candidate media (required)
  --payload-size SIZE      Size of payload to embed (required)
  --strategy STRATEGY      Selection strategy: greedy, optimal, diverse
  --max-files NUM          Maximum number of files to use
  --detectability LEVEL    Target detectability (low, medium, high)
  --output-format FORMAT   Output format (text, json, yaml)
```

**Examples:**

Select media for 50KB payload:
```bash
shadowforge select \
  --input-dir ~/Pictures \
  --payload-size 50KB
```

Select with optimal strategy prioritizing diversity:
```bash
shadowforge select \
  --input-dir ~/Documents/media \
  --payload-size 100KB \
  --strategy diverse \
  --max-files 5
```

Select prioritizing low detectability:
```bash
shadowforge select \
  --input-dir ~/media \
  --payload-size 500KB \
  --detectability low
```

**Output Example:**
```
Selection Plan for 50KB Payload

Selected Media (3 files):

1. vacation.png
   Raw Capacity: 847 KB
   Safe Capacity: 508 KB
   Utilization: 9.8% (low risk)
   Quality Score: 95/100

2. family.jpg
   Raw Capacity: 591 KB
   Safe Capacity: 354 KB
   Utilization: 14.1% (low-medium risk)
   Quality Score: 88/100

3. sunset.png
   Raw Capacity: 423 KB
   Safe Capacity: 254 KB
   Utilization: 19.7% (medium risk)
   Quality Score: 92/100

Summary:
  Total Selected: 50 KB out of available 1.1 MB
  Utilization Rate: 4.5% (very safe)
  Average Quality: 91.7/100
  Media Diversity: HIGH (images only)
  Estimated Detectability: LOW
```

## Getting Suggestions

### `shadowforge suggest`

Get suggestions for addressing capacity shortfalls.

**Usage:**
```bash
shadowforge suggest [flags]
```

**Flags:**
```
  --required-capacity SIZE   Capacity needed (required)
  --available-capacity SIZE  Current available capacity (required)
  --media-types TYPES        Available media types (image, audio, text)
  --output-format FORMAT     Output format (text, json, yaml)
```

**Examples:**

Get suggestions for 500KB payload with 300KB available:
```bash
shadowforge suggest \
  --required-capacity 500KB \
  --available-capacity 300KB \
  --media-types "image,audio"
```

Get detailed suggestions as JSON:
```bash
shadowforge suggest \
  --required-capacity 1MB \
  --available-capacity 400KB \
  --media-types "image,audio,text" \
  --output-format json
```

**Output Example:**
```
Capacity Shortage: 200 KB needed

MEDIA SUGGESTIONS:

Option 1: Add Images (Recommended)
  Add 2 PNG files at 1920x1080 (~200KB each)
  → Provides ~400KB additional capacity
  → Technique: LSB or DCT
  → Risk: Low detectability

  Sources:
    • Unsplash.com - High-quality stock photos
    • Pixabay.com - License-free images
    • Pexels.com - Large image library

Option 2: Add Audio
  Add 1 WAV file, 3+ minutes duration (~500KB)
  → Provides ~500KB additional capacity
  → Technique: Phase Encoding or Echo Hiding
  → Risk: Medium (audio analysis may detect)

Option 3: Reduce Reed-Solomon Redundancy
  Current: 30% redundancy (30 parity shards)
  Reduce to: 20% redundancy (20 parity shards)
  → Gains ~100KB capacity
  ⚠️  WARNING: Lower resilience (K-of-N threshold)

Option 4: Split Payload
  Split 500KB into 2 × 250KB operations
  → Use 300KB + wait for additional media
  → Two separate embedding operations
```

## Comparing Media

### `shadowforge select --dry-run`

Preview selection plan without committing.

**Usage:**
```bash
shadowforge select \
  --input-dir ~/media \
  --payload-size 100KB \
  --dry-run
```

Shows detailed plan with:
- Selected files and utilization
- Capacity breakdown
- Quality scores
- Detectability estimates
- Technique recommendations

## Real-World Workflows

### Workflow 1: Find Best Image for 50KB

```bash
# Scan available images
shadowforge scan --input-dir ~/Pictures --format-filter png

# Auto-select best candidates
shadowforge select \
  --input-dir ~/Pictures \
  --payload-size 50KB \
  --detectability low

# Use selected image
shadowforge embed \
  --input secret.txt \
  --cover vacation.png \
  --output stego.png
```

### Workflow 2: Plan Distribution

```bash
# Check what media we have
shadowforge scan --input-dir ~/media --recursive

# Plan for 500KB distributed across 1:N
shadowforge select \
  --input-dir ~/media \
  --payload-size 500KB \
  --strategy diverse \
  --max-files 5

# Execute distributed embedding
shadowforge embed-distributed \
  --input critical-data.zip \
  --cover-files vacation.png,family.jpg,sunset.png,music.wav,document.txt \
  --threshold 3 \
  --output-dir distributed/
```

### Workflow 3: Handle Capacity Shortfall

```bash
# Plan embedding but face shortage
shadowforge select \
  --input-dir ~/Pictures \
  --payload-size 500KB \
  # Error: Only 300KB available

# Get suggestions
shadowforge suggest \
  --required-capacity 500KB \
  --available-capacity 300KB \
  --media-types "image,audio"

# Option A: Add more media
# Option B: Reduce redundancy
# Option C: Split payload
```

## Selection Strategies

| Strategy | Behavior | Use Case |
|----------|----------|----------|
| `greedy` | Fast, simple selection | Quick operations |
| `optimal` | Best utilization | Efficiency matters |
| `diverse` | Mix media types | Detectability important |

## Capacity Calculation

Capacity = Raw Media Capacity × Technique Efficiency × Quality Factor

Example for PNG image:
- Raw capacity: 850 KB
- LSB technique: 100% efficient
- 95% quality requirement: 850 × 1.0 × 0.95 = **807.5 KB**

---

See also: [Getting Started](../getting-started.md) | [Best Practices](../best-practices.md) | [Embed Commands](./embed.md)

# Technique Chaining Commands

Combine multiple steganography techniques to increase security and capacity.

## Overview

Technique chaining allows you to:
- Apply multiple steganography techniques sequentially
- Distribute data across multiple carriers with different techniques
- Layer different techniques on the same carrier
- Maximize capacity and security through combination

**Three Chaining Modes:**

1. **Sequential** - Apply techniques one after another (technique1 → technique2)
2. **Layered** - Embed different data portions with different techniques in same carrier
3. **Split** - Distribute data across carriers with different techniques per carrier

## Creating Chains

### `shadowforge chain create`

Define a new technique chain configuration.

**Usage:**
```bash
shadowforge chain create [flags]
```

**Flags:**
```
  --name NAME              Chain name (required)
  --techniques LIST        Comma-separated techniques (lsb, dct, phase, etc.)
  --mode MODE              Chaining mode: sequential, layered, split
  --weights WEIGHTS        Technique weights (comma-separated, 0.0-1.0)
  --description TEXT       Chain description
```

**Examples:**

Sequential chaining (LSB then DCT):
```bash
shadowforge chain create \
  --name "lsb-then-dct" \
  --techniques "lsb,dct" \
  --mode sequential
```

Layered chaining with custom weights:
```bash
shadowforge chain create \
  --name "layered-defense" \
  --techniques "lsb,dct,phase" \
  --mode layered \
  --weights "0.5,0.3,0.2" \
  --description "LSB primary, DCT secondary, Phase tertiary"
```

Split across carriers:
```bash
shadowforge chain create \
  --name "distributed-techniques" \
  --techniques "lsb,dct,phase,echo" \
  --mode split \
  --weights "0.25,0.25,0.25,0.25"
```

## Executing Chains

### `shadowforge chain execute`

Embed data using a defined chain configuration.

**Usage:**
```bash
shadowforge chain execute [flags]
```

**Flags:**
```
  --chain NAME             Chain name (required)
  --input FILE             Input payload (required)
  --cover-files FILES      Comma-separated cover media files
  --cover-dir DIR          Directory containing cover media
  --output FILE            Output file (sequential/layered modes)
  --output-dir DIR         Output directory (split mode)
  --verify                 Verify embedding success
  --dry-run                Show plan without executing
  --output-format FORMAT   Output format (text, json, yaml)
```

**Examples:**

Sequential chaining:
```bash
shadowforge chain execute \
  --chain "lsb-then-dct" \
  --input secret.txt \
  --cover image.jpg \
  --output embedded.jpg
```

Layered chaining (all data in one carrier):
```bash
shadowforge chain execute \
  --chain "layered-defense" \
  --input secret.pdf \
  --cover carrier.png \
  --output secured.png \
  --verify
```

Split across multiple carriers:
```bash
shadowforge chain execute \
  --chain "distributed-techniques" \
  --input critical-data.zip \
  --cover-files "img1.png,img2.jpg,img3.png,img4.jpg" \
  --output-dir ./split-output
```

Dry run (plan without executing):
```bash
shadowforge chain execute \
  --chain "layered-defense" \
  --input secret.pdf \
  --cover-dir ~/Pictures \
  --dry-run
```

## Extracting Chained Data

### `shadowforge chain extract`

Recover data embedded with technique chaining.

**Usage:**
```bash
shadowforge chain extract [flags]
```

**Flags:**
```
  --chain NAME             Chain name (required)
  --input FILE             Stego file (single or reference file)
  --input-files FILES      Multiple stego files (split mode)
  --input-dir DIR          Directory containing stego files
  --output FILE            Output file
  --output-dir DIR         Output directory
  --verify                 Verify extraction integrity
  --output-format FORMAT   Output format (text, json, yaml)
```

**Examples:**

Sequential extraction:
```bash
shadowforge chain extract \
  --chain "lsb-then-dct" \
  --input embedded.jpg \
  --output recovered.txt \
  --verify
```

Layered extraction:
```bash
shadowforge chain extract \
  --chain "layered-defense" \
  --input secured.png \
  --output recovered.pdf
```

Split extraction from multiple files:
```bash
shadowforge chain extract \
  --chain "distributed-techniques" \
  --input-dir ./split-output \
  --output recovered-data.zip
```

## Listing Chains

### `shadowforge chain list`

Show all available chain configurations.

**Usage:**
```bash
shadowforge chain list [flags]
```

**Flags:**
```
  --detailed               Show detailed chain information
  --output-format FORMAT   Output format (text, json, yaml)
```

**Example Output:**
```
Available Chains:

1. lsb-then-dct (sequential)
   Techniques: LSB → DCT
   Weights: 0.50, 0.50
   Status: Ready

2. layered-defense (layered)
   Techniques: LSB, DCT, Phase
   Weights: 0.50, 0.30, 0.20
   Description: LSB primary, DCT secondary, Phase tertiary
   Status: Ready

3. distributed-techniques (split)
   Techniques: LSB, DCT, Phase, Echo
   Weights: 0.25, 0.25, 0.25, 0.25
   Status: Ready
```

## Chaining Modes Explained

### Sequential Mode

Apply techniques one after another:

```
Input → [LSB Embed] → Intermediate → [DCT Embed] → Output
```

**Properties:**
- Second technique operates on result of first
- Compounding security
- Single output file
- Higher detectability risk

**Use Case:**
- Maximum security through layering
- Obfuscation of embedding method

### Layered Mode

Embed different data portions with different techniques in same carrier:

```
Input → Split into portions → [Technique 1: Part A] ┐
                              [Technique 2: Part B] ├→ Single Carrier
                              [Technique 3: Part C] ┘
```

**Properties:**
- Single carrier with distributed data
- Different techniques for different parts
- Weights control distribution
- Better balance of capacity and security

**Use Case:**
- Limited carriers but high security needed
- Complex data with varying sensitivity levels

### Split Mode

Distribute data across carriers with different techniques per carrier:

```
Input → [LSB in Img1] → Carrier 1
     → [DCT in Img2] → Carrier 2
     → [Phase in Aud] → Carrier 3
     → [Echo in Aud2] → Carrier 4
```

**Properties:**
- Multiple output files
- Each carrier uses different technique
- Distributed redundancy
- Optimal for resistant distribution

**Use Case:**
- Distributed storage (1:N pattern)
- Multi-carrier resilience
- Maximum flexibility

## Real-World Examples

### Example 1: High-Security Document

```bash
# Create layered chain
shadowforge chain create \
  --name "document-security" \
  --techniques "lsb,dct,phase" \
  --mode layered \
  --weights "0.4,0.35,0.25"

# Embed document
shadowforge chain execute \
  --chain "document-security" \
  --input confidential.pdf \
  --cover photograph.png \
  --output secure-photo.png

# Later, extract
shadowforge chain extract \
  --chain "document-security" \
  --input secure-photo.png \
  --output recovered.pdf \
  --verify
```

### Example 2: Resilient Distribution

```bash
# Create split chain
shadowforge chain create \
  --name "resilient" \
  --techniques "lsb,dct,lsb,dct" \
  --mode split \
  --weights "0.25,0.25,0.25,0.25"

# Distribute across 4 carriers
shadowforge chain execute \
  --chain "resilient" \
  --input critical-data.zip \
  --cover-files "img1.png,img2.jpg,img3.png,img4.jpg" \
  --output-dir ./distributed

# Recovery from any combination
shadowforge chain extract \
  --chain "resilient" \
  --input-dir ./distributed \
  --output recovered-data.zip
```

## Capacity Planning

Different modes have different capacity trade-offs:

| Mode | Capacity | Carriers | Recovery |
|------|----------|----------|----------|
| Sequential | Medium | 1 | Yes (if last carrier OK) |
| Layered | Medium | 1 | Partial (weight-dependent) |
| Split | High | Many | Yes (threshold-dependent) |

---

See also: [Getting Started](../getting-started.md) | [Best Practices](../best-practices.md) | [Watermark Commands](./watermark.md)

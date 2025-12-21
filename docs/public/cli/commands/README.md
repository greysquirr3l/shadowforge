# Commands Reference

Complete reference for all Shadowforge CLI commands.

## Command Categories

### Embedding & Extraction
- **[embed](./embed.md)** - Embed secrets in carrier media
- **[extract](./extract.md)** - Recover secrets from stego media

### Analysis
- **[analyze](./analyze.md)** - Analyze capacity and security
- **[validate](./analyze.md#media-validation)** - Verify stego media integrity

### Archive Management
- **[archive](./archive.md)** - Create and manage archives
- **[archive create](./archive.md)** - Create compressed archives
- **[archive extract](./archive.md)** - Extract archive contents

### Advanced Features
- **[watermark](./watermark.md)** - Forensic watermarking
- **[chain](./chain.md)** - Technique chaining
- **[select](./selection.md)** - Intelligent cover media selection
- **[scan](./selection.md)** - Analyze available media

### Utilities
- **[keygen](./utility.md)** - Generate encryption keys
- **[formats](./utility.md)** - List supported media formats
- **[version](./utility.md)** - Display version information
- **[help](./utility.md)** - Show help information

## Global Flags

All commands support these global flags:

```bash
--help, -h           Show help information
--verbose, -v        Enable verbose output
--log-level LEVEL    Set logging level (debug, info, warn, error)
--config FILE        Load configuration from file
```

## Output Formats

Commands can output in multiple formats:

```bash
--output-format json    Output in JSON format
--output-format text    Output in human-readable text (default)
--output-format yaml    Output in YAML format
```

## File Specifications

### Input Files
Can be specified as:
- Single file: `--input file.txt`
- Multiple files (comma-separated): `--input file1.txt,file2.txt,file3.txt`
- Glob pattern: `--input "*.png"` (quote patterns containing wildcards)
- Directory: `--input-dir /path/to/directory`

### Output
- Single output: `--output file.txt`
- Output directory: `--output-dir /path/to/directory`

## Common Patterns

### Quiet Mode
```bash
shadowforge embed --input secret.txt --cover image.png --output stego.png --quiet
```

### JSON Output
```bash
shadowforge analyze capacity --input image.png --output-format json
```

### Verbose Debugging
```bash
shadowforge embed --input secret.txt --cover image.png --output stego.png -vv --log-level debug
```

## Command Index

| Command | Purpose | Complexity |
|---------|---------|-----------|
| `embed` | Basic embedding | Beginner |
| `extract` | Basic extraction | Beginner |
| `analyze capacity` | Check carrier capacity | Beginner |
| `archive create` | Bundle stego files | Beginner |
| `formats` | List supported types | Beginner |
| `validate` | Verify stego integrity | Intermediate |
| `select` | Auto-select carriers | Intermediate |
| `chain` | Multi-technique embedding | Advanced |
| `watermark` | Forensic tracking | Advanced |

## Quick Examples

**Basic embedding:**
```bash
shadowforge embed --input secret.txt --cover image.png --output stego.png
```

**Extract with integrity verification:**
```bash
shadowforge extract --input stego.png --output secret.txt --verify
```

**Distributed resilient embedding:**
```bash
shadowforge embed \
  --input data.zip \
  --cover img1.png,img2.png,img3.png,img4.png,img5.png \
  --pattern 1:N \
  --threshold 3 \
  --output-dir distributed/
```

**Archive stego files:**
```bash
shadowforge archive create --input-dir ./stego-output --format tar.gz --output stegos.tar.gz
```

**Scan media for capacity:**
```bash
shadowforge scan ~/Pictures --min-capacity 100KB
```

**Technique chaining:**
```bash
shadowforge chain create --name "multi-technique" --techniques lsb,dct --weights 0.5,0.5
shadowforge chain execute --chain "multi-technique" --input secret.txt --cover image.png --output stego.png
```

---

**Want details on a specific command?** See the individual command pages:

- [Embed Commands](./embed.md)
- [Extract Commands](./extract.md)
- [Analysis Commands](./analyze.md)
- [Archive Commands](./archive.md)
- [Watermark Commands](./watermark.md)
- [Chain Commands](./chain.md)
- [Selection Commands](./selection.md)
- [Utility Commands](./utility.md)

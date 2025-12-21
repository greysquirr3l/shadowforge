# Archive Commands

Create, manage, and extract archives of stego media files.

## Overview

Archive commands bundle multiple stego files into compressed packages for easier distribution and storage.

Supported formats:
- **ZIP** - Widely compatible, optional AES-256 password protection
- **TAR** - UNIX standard, preserves permissions
- **TAR.GZ** - TAR with gzip compression

## Creating Archives

### `shadowforge archive create`

Create a new archive from stego media files.

**Usage:**
```bash
shadowforge archive create [flags]
```

**Flags:**
```
  --input FILES             Input files (comma-separated or glob pattern)
  --input-dir DIR          Input directory to archive
  --format FORMAT          Archive format: zip, tar, tar.gz (default: tar.gz)
  --output FILE            Output archive path (required)
  --password               Enable password protection (ZIP only)
  --compress LEVEL         Compression level: none, fastest, default, best
  --preserve-structure     Preserve directory structure in archive
  --output-format FORMAT   Output format (text, json, yaml)
```

**Examples:**

Create TAR.GZ archive:
```bash
shadowforge archive create \
  --input-dir ./stego-output \
  --format tar.gz \
  --output stego-bundle.tar.gz
```

Create password-protected ZIP:
```bash
shadowforge archive create \
  --input "stego1.png,stego2.png,stego3.png" \
  --format zip \
  --output stego-files.zip \
  --password
```

Create with best compression:
```bash
shadowforge archive create \
  --input-dir ./results \
  --format tar.gz \
  --output archive.tar.gz \
  --compress best
```

## Extracting Archives

### `shadowforge archive extract`

Extract contents from stego archives.

**Usage:**
```bash
shadowforge archive extract [flags]
```

**Flags:**
```
  --input FILE             Archive file (required)
  --output-dir DIR         Output directory for extracted files
  --password               Prompt for password (if encrypted)
  --verify                 Verify extracted files
  --output-format FORMAT   Output format (text, json, yaml)
```

**Examples:**

Extract TAR.GZ archive:
```bash
shadowforge archive extract \
  --input stego-bundle.tar.gz \
  --output-dir ./recovered
```

Extract password-protected ZIP:
```bash
shadowforge archive extract \
  --input stego-files.zip \
  --output-dir ./recovered \
  --password
```

Extract and verify:
```bash
shadowforge archive extract \
  --input archive.tar.gz \
  --output-dir ./recovered \
  --verify
```

## Listing Archive Contents

### `shadowforge archive list`

List contents of an archive without extracting.

**Usage:**
```bash
shadowforge archive list [flags]
```

**Flags:**
```
  --input FILE             Archive file (required)
  --detailed               Show detailed file information
  --output-format FORMAT   Output format (text, json, yaml)
```

**Examples:**

List archive contents:
```bash
shadowforge archive list --input stego-bundle.tar.gz
```

List with details:
```bash
shadowforge archive list \
  --input stego-files.zip \
  --detailed \
  --output-format json
```

## Real-World Workflows

### Scenario 1: Distribute Resilient Data

```bash
# Embed data across 5 images
shadowforge embed \
  --input critical-data.zip \
  --cover-files img1.png,img2.png,img3.png,img4.png,img5.png \
  --pattern 1:N \
  --threshold 3 \
  --output-dir distributed/

# Bundle into single archive
shadowforge archive create \
  --input-dir distributed/ \
  --format tar.gz \
  --output critical-resilient.tar.gz

# Share the archive (any 3 of 5 images can recover data)
```

### Scenario 2: Secure Document Bundle

```bash
# Embed multiple documents
shadowforge embed \
  --input "doc1.pdf,doc2.pdf,doc3.pdf" \
  --cover image.png \
  --pattern N:1 \
  --output stego.png

# Create password-protected archive
shadowforge archive create \
  --input stego.png \
  --format zip \
  --output secure-docs.zip \
  --password

# Share: recipient extracts archive with password, then extracts documents
```

## Archive Format Comparison

| Format | Pros | Cons | Use Case |
|--------|------|------|----------|
| ZIP | Wide compatibility, password protection available | Limited compression | Quick sharing |
| TAR | Preserves permissions, large file support | No compression by default | UNIX systems |
| TAR.GZ | Good compression, preserves permissions | Less compatible with Windows | Distribution, archival |

## Security Considerations

- **ZIP Encryption**: Uses AES-256 encryption
- **Password Strength**: Use strong, random passwords
- **Transmission**: Send archive and password through different channels
- **Verification**: Always verify extracted files match originals

---

See also: [Getting Started](../getting-started.md) | [Best Practices](../best-practices.md)

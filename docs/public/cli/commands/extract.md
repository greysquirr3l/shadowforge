# Extract Commands

Recover secret data that has been embedded in carrier media.

## Basic Extraction

### `shadowforge extract`

Extract a secret from a single stego image.

**Usage:**
```bash
shadowforge extract [flags]
```

**Flags:**
```
  --input FILE              Stego media file (required)
  --output FILE             Output recovered file (required)
  --password PASSWORD       Password for decryption (if encrypted)
  --verify                  Verify integrity after extraction (default: true)
  --technique TECH          Override technique detection
  --force                   Overwrite output if exists (default: false)
  -v, --verbose             Enable verbose output
  --output-format FORMAT    Output format (text, json, yaml)
```

**Examples:**

Basic extraction:
```bash
shadowforge extract --input stego.png --output recovered.txt
```

With password:
```bash
shadowforge extract \
  --input protected.png \
  --output secret.pdf \
  --password "my-passphrase"
```

Verbose with JSON output:
```bash
shadowforge extract \
  --input stego.png \
  --output recovered.txt \
  --verbose \
  --output-format json \
  --verify
```

Override technique detection:
```bash
shadowforge extract \
  --input stego.png \
  --output recovered.txt \
  --technique dct
```

## Distributed Extraction

### `shadowforge extract-distributed`

Recover a secret distributed across multiple carriers (one-to-many recovery).

**Usage:**
```bash
shadowforge extract-distributed [flags]
```

**Flags:**
```
  --input FILES             Stego image files (comma-separated or directory)
  --input-dir DIR           Directory containing stego images
  --output FILE             Output recovered file (required)
  --manifest FILE           Optional manifest file
  --threshold INT           Minimum images needed (auto-detected)
  --password PASSWORD       Decryption password (if applicable)
  --force                   Overwrite output if exists
  --verify                  Verify recovered data (default: true)
  -v, --verbose             Enable verbose output
```

**How It Works:**
1. Scans input images for embedded shards
2. Detects how many shards are available
3. If enough shards are available (≥ threshold), recovers the secret
4. Verifies recovered data integrity

**Examples:**

Basic recovery from multiple images:
```bash
shadowforge extract-distributed \
  --input stego1.png,stego2.png,stego3.png,stego4.png,stego5.png \
  --output recovered-data.zip
```

Recover from directory:
```bash
shadowforge extract-distributed \
  --input-dir ./stego-files \
  --output important.pdf
```

With manifest file:
```bash
shadowforge extract-distributed \
  --input-dir ./distributed \
  --manifest distribution.manifest \
  --output recovered.zip \
  --verbose
```

Recover from partial set (only 2 of 5 images):
```bash
# If threshold was 2 and you have 5 images, 2 is sufficient
shadowforge extract-distributed \
  --input available-image1.png,available-image2.png \
  --output recovered.data
```

## Batch Extraction

### `shadowforge extract-batch`

Extract multiple secrets from a single stego image (many-to-one extraction).

**Usage:**
```bash
shadowforge extract-batch [flags]
```

**Flags:**
```
  --input FILE              Stego media with multiple payloads
  --output-dir DIR          Directory for extracted files (required)
  --index FILE              Index file mapping payloads (optional)
  --extract NAMES           Extract only specific files (comma-separated)
  --list                    List all embedded files without extracting
  --password PASSWORD       Decryption password (if applicable)
  --verify                  Verify each extraction (default: true)
  --force                   Overwrite existing files
  -v, --verbose             Enable verbose output
```

**Examples:**

Extract all payloads:
```bash
shadowforge extract-batch \
  --input stego.png \
  --output-dir extracted/
```

List embedded files without extracting:
```bash
shadowforge extract-batch \
  --input stego.png \
  --list
```

Output:
```
✓ Stego media contains 3 embedded files:
  1. document.pdf (185 KB)
  2. spreadsheet.xlsx (45 KB)
  3. notes.txt (2 KB)
```

Extract only specific files:
```bash
shadowforge extract-batch \
  --input stego.png \
  --output-dir extracted/ \
  --extract document.pdf,notes.txt
```

## Matrix Extraction

### `shadowforge extract-matrix`

Extract multiple secrets from multiple carriers (many-to-many extraction).

**Usage:**
```bash
shadowforge extract-matrix [flags]
```

**Flags:**
```
  --input-dir DIR           Directory with stego images
  --input FILES             Stego files (comma-separated)
  --output-dir DIR          Output directory (required)
  --manifest FILE           Manifest file from embedding
  --password PASSWORD       Decryption password
  --extract NAMES           Extract specific payloads (optional)
  --threshold INT           Override minimum shards needed
  --list                    List available payloads
  --verify                  Verify recoveries (default: true)
  -v, --verbose             Enable verbose output
```

**Examples:**

Extract from matrix distribution:
```bash
shadowforge extract-matrix \
  --input-dir ./matrix-stego \
  --output-dir ./recovered \
  --manifest matrix.manifest
```

List payloads:
```bash
shadowforge extract-matrix \
  --input-dir ./stego-files \
  --list

# Output:
# Available payloads:
# - payload1.pdf (needs 3 of 6 carriers)
# - payload2.doc (needs 3 of 6 carriers)
# - payload3.zip (needs 3 of 6 carriers)
```

Extract specific payloads:
```bash
shadowforge extract-matrix \
  --input-dir ./stego-files \
  --output-dir ./extracted \
  --extract payload1.pdf,payload3.zip
```

## Integrity Verification

### Built-in Verification

All extract commands support `--verify` (enabled by default):

```bash
# Automatically verifies extracted data
shadowforge extract --input stego.png --output secret.txt --verify

# Disable verification for speed (not recommended)
shadowforge extract --input stego.png --output secret.txt --no-verify
```

**What gets verified:**
- Payload structure integrity
- CRC/checksum validation
- Digital signature verification (if present)
- Data completeness

### `shadowforge validate`

Independently verify stego media without extracting:

```bash
# Check if a stego image is valid
shadowforge validate --input stego.png

# Check multiple files
shadowforge validate --input stego1.png,stego2.png,stego3.png

# Verbose output
shadowforge validate --input stego.png --verbose
```

## Error Recovery

If extraction fails:

**"Corrupted data" error:**
- Ensure you're using the correct password (if encrypted)
- Try specifying explicit technique: `--technique dct`
- Use `shadowforge validate` to check media integrity

**"Insufficient shards" error (distributed):**
- You need at least `--threshold` images
- Check manifest to see how many are required
- Some carrier images may be corrupted

**"Image format error":**
- Ensure input is a valid image file
- Try converting to a different format
- Check file was not re-compressed after embedding

## Performance Tips

Extract from multiple files in parallel:
```bash
# Faster processing
shadowforge extract-distributed \
  --input-dir ./stego-files \
  --output recovered.zip \
  --parallel
```

Skip verification for trusted data:
```bash
# Faster but less safe
shadowforge extract --input stego.png --output secret.txt --no-verify
```

---

See [Embed Commands](./embed.md) to hide secrets in the first place.

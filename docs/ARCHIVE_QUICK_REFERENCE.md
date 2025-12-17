# Archive Creation Quick Reference

Create compressed archives of your stego files with ease.

## Quick Start

```bash
# Create a ZIP archive from multiple files
shadowforge archive create -i stego1.png,stego2.png,stego3.png -o stegos.zip

# Create a compressed TAR.GZ archive from a directory
shadowforge archive create --input-dir ./stego-output -o archive.tar.gz -f tar.gz
```

## Command Syntax

```bash
shadowforge archive create [flags]
```

## Common Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--input` | `-i` | Input files (comma-separated) or glob pattern | - |
| `--input-dir` | - | Input directory to archive | - |
| `--output` | `-o` | Output archive file (**required**) | - |
| `--format` | `-f` | Archive format (zip, tar, tar.gz) | `zip` |
| `--compression` | `-c` | Compression level (none, fastest, default, best) | `default` |
| `--password` | `-p` | Password for encryption (ZIP only) | - |
| `--encrypt` | - | Enable encryption (prompts for password) | `false` |
| `--verbose` | `-v` | Verbose output | `false` |

## Archive Formats

### ZIP (Default)

- **Best for**: Cross-platform compatibility, Windows users
- **Compression**: Deflate (default)
- **Extensions**: `.zip`
- **Password protection**: 🔮 Coming soon (requires third-party library)

```bash
shadowforge archive create -i *.png -o stegos.zip
```

### TAR

- **Best for**: Unix/Linux systems, preserving permissions
- **Compression**: None (fast but larger files)
- **Extensions**: `.tar`

```bash
shadowforge archive create -i *.png -o stegos.tar -f tar
```

### TAR.GZ

- **Best for**: Maximum compression, distribution over networks
- **Compression**: Gzip (efficient)
- **Extensions**: `.tar.gz`, `.tgz`

```bash
shadowforge archive create -i *.png -o stegos.tar.gz -f tar.gz -c best
```

## Compression Levels

| Level | Speed | Size | Use Case |
|-------|-------|------|----------|
| `none` | Fastest | Largest | Quick archival, already compressed files |
| `fastest` | Fast | Large | Time-sensitive operations |
| `default` | Balanced | Medium | **Recommended for most use cases** |
| `best` | Slow | Smallest | Distribution, long-term storage |

```bash
# Example: Maximum compression for distribution
shadowforge archive create \
  --input-dir ./stego-files \
  -o distribution.tar.gz \
  -f tar.gz \
  -c best
```

## Input Methods

### Individual Files

Specify files separated by commas:

```bash
shadowforge archive create \
  -i stego1.png,stego2.png,stego3.wav \
  -o mixed-media.zip
```

### Glob Patterns

Use shell wildcards to match multiple files:

```bash
# All PNG files in current directory
shadowforge archive create -i *.png -o images.zip

# All stego files starting with "stego-"
shadowforge archive create -i stego-*.* -o collection.zip

# Specific pattern
shadowforge archive create -i stego-image-*.png -o stego-images.zip
```

### Directory Archival

Archive entire directories recursively:

```bash
shadowforge archive create \
  --input-dir ./stego-output \
  -o complete-archive.tar.gz \
  -f tar.gz
```

This preserves the directory structure inside the archive.

## Real-World Examples

### 1. Archive Stego Images for Email

```bash
# Create a small ZIP for email attachment (< 25MB)
shadowforge archive create \
  -i stego-*.png \
  -o stego-collection.zip \
  -c best
```

### 2. Backup Stego Files with Best Compression

```bash
# Create highly compressed backup
shadowforge archive create \
  --input-dir ./all-stego-files \
  -o stego-backup-$(date +%Y%m%d).tar.gz \
  -f tar.gz \
  -c best
```

### 3. Quick Archive for Local Transfer

```bash
# Fast archival with no compression
shadowforge archive create \
  --input-dir ./stego-temp \
  -o transfer.tar \
  -f tar \
  -c none
```

### 4. Archive with Pattern Matching

```bash
# Archive only JPEG stego files
shadowforge archive create \
  -i stego-*.jpg \
  -o jpeg-stegos.zip

# Archive multiple file types
shadowforge archive create \
  -i "stego-*.{png,jpg,wav}" \
  -o multimedia-stegos.zip
```

### 5. Verbose Archival for Verification

```bash
# See detailed file-by-file progress
shadowforge archive create \
  --input-dir ./important-stegos \
  -o verified-archive.tar.gz \
  -f tar.gz \
  -v
```

## Password Protection (Future Feature)

### With Password Prompting

```bash
# Will prompt securely for password (future)
shadowforge archive create \
  -i sensitive-*.png \
  -o secure.zip \
  --encrypt
```

### With Password Provided

```bash
# Provide password directly (not recommended - visible in history)
shadowforge archive create \
  -i sensitive-*.png \
  -o secure.zip \
  -p mypassword
```

**Current Status**: Password-protected ZIP requires additional library (`github.com/yeka/zip`).

**Workarounds**:

1. **Use `zip` command after creation**:
   ```bash
   shadowforge archive create -i *.png -o stegos.zip
   zip -e secure.zip stegos.zip
   rm stegos.zip
   ```

2. **Use GPG with TAR.GZ**:
   ```bash
   shadowforge archive create -i *.png -o stegos.tar.gz -f tar.gz
   gpg -c stegos.tar.gz  # Creates stegos.tar.gz.gpg
   ```

3. **Use 7zip**:
   ```bash
   shadowforge archive create -i *.png -o stegos.zip
   7z a -p -mhe=on secure.7z stegos.zip
   ```

**Note**: Your stego files already contain Kyber-1024 encrypted payloads, so archive-level encryption is optional defense-in-depth.

## Tips & Best Practices

### 1. Choose the Right Format

- **ZIP**: Best for sharing with Windows users or when you need cross-platform compatibility
- **TAR.GZ**: Best for Unix/Linux environments and maximum compression
- **TAR**: Best for quick archival when files are already compressed

### 2. Compression Trade-offs

- **Already Compressed Data**: Stego PNG/JPEG files won't compress much further
  - Use `--compression none` or `fastest` for speed
- **Text Files**: Use `--compression best` for significant size reduction
- **Mixed Content**: Use `--compression default` (balanced)

### 3. Glob Patterns

- Always quote patterns with special characters: `-i "stego-*.{png,jpg}"`
- Test pattern first: `ls stego-*.png` before archiving
- Use absolute paths for reliability: `-i /full/path/to/stego-*.png`

### 4. Verify Archives

After creation, verify your archives:

```bash
# Verify ZIP
unzip -l output.zip

# Verify TAR.GZ
tar -tzf output.tar.gz

# Extract to test
mkdir test-extract
tar -xzf output.tar.gz -C test-extract
```

### 5. Large Archives

For very large archives (> 1GB):

- Use TAR.GZ with `best` compression
- Consider splitting files across multiple archives
- Monitor disk space (archive size ≈ 30-50% of original for mixed media)

## Troubleshooting

### Error: "no files found to archive"

- Check that input files exist: `ls -la <pattern>`
- Verify file paths are correct (use absolute paths)
- Ensure glob patterns are quoted

### Error: "cannot specify both --input and --input-dir"

- Use only ONE input method at a time
- For multiple directories, archive them separately or create parent archive

### Error: "password protection is only supported for ZIP format"

- Password protection only works with ZIP format
- Remove `--password` or `--encrypt` flags for TAR/TAR.GZ
- Use external tools (GPG) for encrypting TAR.GZ archives

### Error: "failed to read password: inappropriate ioctl for device"

- Password prompting only works in interactive terminals
- Don't pipe input to the command
- Provide password with `-p` flag if using in scripts (not recommended)

### Warning: "Failed to read file, skipping"

- File may have been deleted or moved
- Check file permissions: `ls -l <file>`
- File may be locked by another process

## Output Information

After successful creation, Shadowforge displays:

```
✅ Archive created successfully!
Output: /path/to/archive.zip
Format: zip
Files: 15
Total Size: 12.3 MB
Archive Size: 11.8 MB
Compression: 95.9%
```

**Compression percentage** shows archive size relative to original:
- **100%** = No compression (same size)
- **<100%** = Compressed smaller
- **>100%** = Larger (due to archive overhead on small/already compressed files)

## Security Notes

### What's Encrypted?

1. **Payload Data**: ✅ Always encrypted with Kyber-1024 post-quantum cryptography
2. **Stego Container**: ❌ Not encrypted (but payload inside is)
3. **Archive**: 🔮 Optional future feature (ZIP passwords)

### Why Archive-Level Encryption?

- **Defense in Depth**: Multiple layers of protection
- **Metadata Protection**: Hides file names and structure
- **Compliance**: May be required for certain regulations

### Current Security Status

Your data is already **quantum-resistant encrypted** at the payload level:
- Kyber-1024 for encryption (NIST-approved PQC)
- Dilithium3 for signatures (NIST-approved PQC)
- Reed-Solomon error correction

Archive encryption is **optional** additional protection.

## Related Commands

- `shadowforge embed` - Embed data into files before archiving
- `shadowforge extract` - Extract data from stego files in archive
- `shadowforge analyze capacity` - Calculate embedding capacity before creating stegos

## Full Example Workflow

Complete workflow from embedding to archival:

```bash
# 1. Embed secret data in multiple images
shadowforge embed 1-n \
  --secret secret.txt \
  --covers cover1.png,cover2.png,cover3.png \
  --output-dir ./stego-output

# 2. Archive the resulting stego files
shadowforge archive create \
  --input-dir ./stego-output \
  -o stego-collection-$(date +%Y%m%d).tar.gz \
  -f tar.gz \
  -c best \
  -v

# 3. Verify the archive
tar -tzf stego-collection-*.tar.gz

# 4. (Optional) Add password protection
gpg -c stego-collection-*.tar.gz

# 5. Distribute the encrypted archive
# Share stego-collection-YYYYMMDD.tar.gz.gpg
```

## Get Help

```bash
# General archive help
shadowforge archive --help

# Create command help
shadowforge archive create --help

# Global help
shadowforge --help
```

---

**Feature Status**: ✅ Production Ready (v0.5.0+)
**Password Protection**: 🔮 Coming Soon (requires third-party library)
**Supported Platforms**: macOS, Linux, Windows

*Your stego files are already quantum-resistant encrypted. Archive creation is for organization and distribution convenience.*

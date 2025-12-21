# Utility Commands

Helper commands for key management, format information, and system utilities.

## Key Generation

### `shadowforge keygen`

Generate new encryption key pairs for post-quantum cryptography.

**Usage:**
```bash
shadowforge keygen [flags]
```

**Flags:**
```
  --algorithm ALGO         Key algorithm: kyber1024, dilithium3, both (default)
  --output-dir DIR         Output directory for keys
  --public-only            Generate public key only (for recipients)
  --password               Encrypt private key with password
  --format FORMAT          Output format: pem, json
```

**Examples:**

Generate both Kyber and Dilithium keys:
```bash
shadowforge keygen
```

Generate only public keys for distribution:
```bash
shadowforge keygen --public-only --output-dir ./public-keys
```

Generate with password protection:
```bash
shadowforge keygen \
  --algorithm kyber1024 \
  --output-dir ./secure-keys \
  --password
```

**Output Files:**
- `kyber_public.key` - Kyber public key (for recipients)
- `kyber_private.key` - Kyber private key (keep secret)
- `dilithium_public.key` - Dilithium public key
- `dilithium_private.key` - Dilithium private key

### `shadowforge keyexport`

Export keys in various formats for sharing or backup.

**Usage:**
```bash
shadowforge keyexport [flags]
```

**Flags:**
```
  --input FILE             Input key file (required)
  --output FILE            Output file
  --format FORMAT          Output format: pem, json, text
  --armor                  ASCII-armor the key (PEM format)
```

**Examples:**

Export public key as ASCII-armored PEM:
```bash
shadowforge keyexport \
  --input kyber_public.key \
  --format pem \
  --armor \
  --output kyber_public.pem
```

Export as JSON for programmatic use:
```bash
shadowforge keyexport \
  --input dilithium_public.key \
  --format json \
  --output dilithium_public.json
```

### `shadowforge keyimport`

Import keys from external sources or backups.

**Usage:**
```bash
shadowforge keyimport [flags]
```

**Flags:**
```
  --input FILE             Input key file (required)
  --output FILE            Output key file
  --type TYPE              Key type: public, private
  --algorithm ALGO         Algorithm: kyber1024, dilithium3
  --password               Decrypt password-protected key
```

**Examples:**

Import public key for verification:
```bash
shadowforge keyimport \
  --input recipient_kyber_public.pem \
  --type public \
  --algorithm kyber1024 \
  --output imported_public.key
```

Import encrypted private key:
```bash
shadowforge keyimport \
  --input backup_private.key.enc \
  --password \
  --output restored_private.key
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
  --type TYPE              Filter by media type: image, audio, text, archive
  --technique TECHNIQUE    Filter by supported technique
  --verbose, -v            Show detailed format information
  --output-format FORMAT   Output format (text, json, yaml)
```

**Examples:**

List all supported formats:
```bash
shadowforge formats
```

List only image formats:
```bash
shadowforge formats --type image
```

Show details for LSB-compatible formats:
```bash
shadowforge formats --technique lsb --verbose
```

Show JSON output for programmatic use:
```bash
shadowforge formats --output-format json
```

**Output Example:**
```
Supported Media Formats

IMAGES:
  PNG
    Max Resolution: Unlimited
    Bit Depth: 1-48 bits
    Techniques: LSB, DCT (as JPEG), Palette
    Capacity: Variable (2-3 KB per megapixel)
    Status: ✓ Fully Supported

  JPEG
    Max Resolution: Unlimited
    Bit Depth: 8/24 bits
    Techniques: DCT
    Capacity: 1-2 KB per megapixel
    Status: ✓ Fully Supported (note: re-compression corrupts data)

  BMP
    Max Resolution: Unlimited
    Bit Depth: 1-32 bits
    Techniques: LSB
    Capacity: 2-3 KB per megapixel
    Status: ✓ Fully Supported

  GIF
    Max Resolution: Unlimited
    Colors: 256 palette colors
    Techniques: Palette, LSB
    Capacity: 0.5-1 KB per megapixel
    Status: ✓ Fully Supported

AUDIO:
  WAV
    Sample Rates: 8 kHz - 48 kHz
    Bit Depth: 8, 16, 24, 32 bits
    Channels: Mono, Stereo, Multi-channel
    Techniques: LSB, Phase, Echo
    Capacity: 1-5 KB per minute
    Status: ✓ Fully Supported

  FLAC
    Read-only support
    Status: ⚠️  Read-only (extraction only)

TEXT:
  Plain Text (.txt)
    Encoding: UTF-8, UTF-16, ASCII
    Techniques: Zero-Width
    Capacity: ~500 bytes per page
    Status: ✓ Fully Supported

  Markdown (.md)
    Encoding: UTF-8
    Techniques: Zero-Width
    Capacity: ~500 bytes per page
    Status: ✓ Fully Supported

ARCHIVES:
  ZIP
    Compression: Store, Deflate
    Encryption: AES-256 (optional)
    Status: ✓ Fully Supported

  TAR
    Compression: None
    Status: ✓ Fully Supported

  TAR.GZ
    Compression: Gzip
    Status: ✓ Fully Supported
```

## System Information

### `shadowforge version`

Display version and build information.

**Usage:**
```bash
shadowforge version [flags]
```

**Flags:**
```
  --json               Output version as JSON
  --check-updates      Check for newer versions online
```

**Examples:**

Show version:
```bash
shadowforge version
```

Check for updates:
```bash
shadowforge version --check-updates
```

Output as JSON:
```bash
shadowforge version --json
```

**Output Example:**
```
Shadowforge CLI v1.0.0

Build Information:
  Commit:    a1b2c3d4e5f6g7h8i9j0
  Build Date: 2025-12-21
  Go Version: 1.21.0
  OS/Arch:   darwin/arm64

Features:
  ✓ Post-Quantum Cryptography (Kyber-1024, Dilithium3)
  ✓ Reed-Solomon Error Correction
  ✓ 7 Steganography Techniques
  ✓ 4 Distribution Patterns
  ✓ Technique Chaining
  ✓ Forensic Watermarking

License: Apache 2.0
```

### `shadowforge help`

Show help for commands.

**Usage:**
```bash
shadowforge help [COMMAND]
```

**Examples:**

Show general help:
```bash
shadowforge help
```

Show help for embed command:
```bash
shadowforge help embed
```

Show help for all commands:
```bash
shadowforge help --all
```

## System Diagnostics

### `shadowforge --version`

Quick version check (alias for `version`).

```bash
shadowforge --version
# Output: shadowforge version 1.0.0
```

### `shadowforge --help`

Show complete help (alias for `help`).

```bash
shadowforge --help
# Shows all available commands and global flags
```

## Configuration

### Environment Variables

Control behavior via environment variables:

```bash
# Logging
export SHADOWFORGE_LOG_LEVEL=debug    # debug, info, warn, error
export SHADOWFORGE_LOG_MODE=cli       # cli (colorized) or api (JSON)

# Performance
export SHADOWFORGE_WORKERS=4          # Number of parallel workers
export SHADOWFORGE_BUFFER_SIZE=32MB   # Buffer size for large files

# Security
export SHADOWFORGE_SECURE_TEMP=true   # Secure temporary file handling
export SHADOWFORGE_ZERO_MEMORY=true   # Mandatory memory zeroing
```

### Configuration File

Create `~/.shadowforge/config.yaml` for persistent settings:

```yaml
logging:
  level: info
  mode: cli
  format: text

security:
  secure_temp: true
  zero_memory: true

performance:
  workers: 4
  buffer_size: 32MB

default_options:
  output_format: text
  compression: best
```

## Real-World Examples

### Example 1: Initial Setup

```bash
# Generate key pair
shadowforge keygen --password

# List all formats
shadowforge formats

# Check version
shadowforge version
```

### Example 2: Check Compatibility

```bash
# Which formats support LSB?
shadowforge formats --technique lsb

# Detailed info on JPEG
shadowforge formats --type image --verbose | grep -A 10 JPEG
```

### Example 3: Export Public Key for Others

```bash
# Generate keys
shadowforge keygen

# Export public keys only
shadowforge keyexport \
  --input kyber_public.key \
  --format pem \
  --armor \
  --output ~/share/kyber_public.pem

shadowforge keyexport \
  --input dilithium_public.key \
  --format pem \
  --armor \
  --output ~/share/dilithium_public.pem

# Share the .pem files (safe to distribute)
```

---

See also: [Getting Started](../getting-started.md) | [Best Practices](../best-practices.md)

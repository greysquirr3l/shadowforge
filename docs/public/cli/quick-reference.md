# Quick Reference

## Most Common Commands

```bash
# Embed a secret in an image
shadowforge embed secret.txt image.png -o stego.png

# Extract the secret from stego image
shadowforge extract stego.png -o recovered.txt

# Check capacity before embedding
shadowforge analyze capacity image.png

# Generate a new encryption key pair
shadowforge keygen -o my_keys.json

# List all supported formats
shadowforge formats

# Show help for any command
shadowforge embed --help
```

## Distribution Patterns

```bash
# One-to-one (simplest)
shadowforge embed secret.txt image.png -o stego.png

# One-to-many (split across multiple images)
shadowforge embed-distributed secret.txt image1.png image2.png image3.png

# Many-to-one (bundle multiple files)
shadowforge embed-batch secret1.txt secret2.txt image.png

# Many-to-many (complex distribution)
shadowforge embed-matrix secret1.txt secret2.txt image1.png image2.png
```

## Workflow Examples

### Safe Sharing: One-to-Many

```bash
# Split across 5 images, need any 3 to recover
shadowforge embed-distributed secret.txt \
  beach.png sunset.png forest.png mountain.png city.png \
  --threshold 3

# Later, recover using any 3
shadowforge extract-distributed beach.png sunset.png forest.png -o secret.txt
```

### Bundling: Multiple Files

```bash
# Embed multiple files in one image
shadowforge embed-batch \
  contracts.pdf financial.xlsx notes.txt \
  carrier.png

# Later, extract all
shadowforge extract carrier.png
```

### Archive: Secure Storage

```bash
# Create password-protected archive
shadowforge archive create \
  stego1.png stego2.png stego3.png \
  -o secure_bundle.zip \
  --encrypt

# Later, extract archive
shadowforge archive extract secure_bundle.zip
```

## Flags & Options

### Global Options
```bash
--verbose, -v         More detailed output
--quiet, -q           Minimal output
--output-format json  JSON output for scripts
```

### Embedding Options
```bash
--technique LSB       Force LSB embedding (default: auto)
--technique DCT       Force DCT (JPEG only)
--password secret     Encrypt with password
--redundancy 0.3      30% Reed-Solomon redundancy (default)
```

### Analysis Options
```bash
--technique LSB       Analyze for specific technique
--conservative        Use conservative capacity estimate
```

### Distributed Options
```bash
--threshold 3         Need 3 of N carriers to recover
--parallel true       Use parallel processing
--workers 4           Use 4 parallel workers
```

## File Type Support

| Format | Read | Write | Technique |
|--------|------|-------|-----------|
| PNG    | ✓    | ✓     | LSB, DCT, Palette |
| JPEG   | ✓    | ✓     | DCT |
| BMP    | ✓    | ✓     | LSB |
| GIF    | ✓    | ✓     | Palette |
| WAV    | ✓    | ✓     | LSB, Phase, Echo |
| TXT    | ✓    | ✓     | Zero-Width |
| ZIP    | ✓    | ✓     | N/A (Archive) |
| TAR    | ✓    | ✓     | N/A (Archive) |

## Capacity Estimates

These are approximate safe capacities (50% of theoretical max):

| Carrier | Size | Safe Capacity |
|---------|------|---------------|
| PNG 1920x1080 | 770KB | ~50KB |
| JPEG 1920x1080 | 180KB | ~13KB |
| WAV 3 min, 44.1kHz | 260KB | ~3KB (LSB) |
| Text 5000 chars | ~10KB | ~10 bytes |

Use `shadowforge analyze capacity` for exact estimates.

## Checklists

### Before Embedding
- [ ] Carrier is original, not edited
- [ ] Carrier is not already compressed
- [ ] Have backup of carrier
- [ ] Payload fits in carrier (check with `analyze`)
- [ ] Password is secure (if encrypting)
- [ ] Know which technique you're using

### After Embedding
- [ ] Stego image looks normal
- [ ] Verify can extract (test locally)
- [ ] Save manifest (for distributed)
- [ ] Secure original payload
- [ ] Consider secure deletion of temp files

### Before Distributing
- [ ] Verify recovery works (with at least threshold carriers)
- [ ] Have backup of manifest
- [ ] Confirm recipients have Shadowforge installed
- [ ] Test on their system if possible
- [ ] Establish secure channel for passwords (if used)

## Performance Tips

1. **Faster embedding:**
   - Use GPU acceleration (if available): `--gpu`
   - Parallel processing: `--parallel true --workers 8`
   - Smaller images: Reduce resolution if acceptable

2. **Smaller file sizes:**
   - Compress payload first: `gzip file`
   - Use JPEG with DCT (smaller file, same capacity)
   - Use optimized image tools: `optipng`, `jpegoptim`

3. **Better quality carriers:**
   - Use original uncompressed images
   - Higher resolution = better quality
   - Natural photos > screenshots/memes

## Security Reminders

1. **Steganography != Encryption**
   - Hide data + Encrypt it for best security
   ```bash
   gpg --symmetric secret.txt
   shadowforge embed secret.txt.gpg image.png
   ```

2. **Pattern Recognition**
   - Don't always embed in same location
   - Vary carrier formats
   - Mix with normal image sharing

3. **Plausible Deniability**
   - Use technique chaining for confusion
   - Distribute across many images
   - Can claim images are just normal photos

4. **Safe Deletion**
   ```bash
   # Secure delete (macOS/Linux)
   shred -vfz -n 3 secret.txt

   # Windows
   cipher /w:C:
   ```

## Environment Variables

```bash
# Logging
SHADOWFORGE_LOG_LEVEL=debug      # debug, info, warn, error
SHADOWFORGE_LOG_MODE=cli         # cli (colorized) or api (json)

# Performance
SHADOWFORGE_WORKERS=8            # Worker pool size
SHADOWFORGE_CACHE_SIZE=1GB       # In-memory cache

# Security
SHADOWFORGE_USE_GPU=true         # GPU acceleration
SHADOWFORGE_CONSTANT_TIME=true   # Force constant-time ops
```

## One-Liners

```bash
# Quick embed with timestamp
shadowforge embed secret.txt image.png \
  -o stego_$(date +%s).png

# Batch process images in directory
for img in *.png; do
  shadowforge analyze capacity "$img"
done

# Create backup with archive
shadowforge archive create stego*.png \
  -o backup_$(date +%Y%m%d).zip --encrypt

# Extract all from distributed
shadowforge extract-distributed distributed/*.png -o secret.txt

# Secure cleanup
find . -name "*.tmp" -exec shred -vfz {} \;
```

## Common Errors

| Error | Cause | Fix |
|-------|-------|-----|
| "insufficient capacity" | Payload too large | Use larger carrier or split payload |
| "no hidden data" | Wrong carrier/corrupted | Verify carrier file, try different technique |
| "invalid checksum" | Data corrupted | Use distributed pattern for redundancy |
| "insufficient shards" | Missing carriers | Need at least threshold number |
| "permission denied" | Binary not executable | `chmod +x $(which shadowforge)` |

## Getting Help

```bash
# General help
shadowforge help

# Command-specific help
shadowforge embed --help
shadowforge extract --help

# Version info
shadowforge version

# System diagnostics
shadowforge diagnostics

# View logs (if saved)
SHADOWFORGE_LOG_LEVEL=debug shadowforge <command> 2>&1 | tee shadowforge.log
```

---

**Need more?** See the full [documentation](./README.md) or check [FAQ](./faq.md).

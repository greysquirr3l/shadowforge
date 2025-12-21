# Troubleshooting Guide

## General Issues

### "Command not found: shadowforge"

**Problem:** The CLI isn't in your PATH.

**Solution:**
1. Verify installation: `which shadowforge`
2. If not found, reinstall using [Installation Guide](./installation.md)
3. macOS Homebrew: `brew install greysquirr3l/tap/shadowforge`
4. Manual: Add binary location to PATH: `export PATH="/path/to/bin:$PATH"`

### "Permission denied" when running shadowforge

**Problem:** Binary doesn't have execute permission.

**Solution:**
```bash
chmod +x $(which shadowforge)  # Make binary executable
```

### Application crashes without error message

**Problem:** Silent failure, no output.

**Solution:**
```bash
# Enable debug logging
SHADOWFORGE_LOG_LEVEL=debug shadowforge <command>

# Check system resources
free -h          # Memory (Linux)
sysctl hw.memsize (macOS)
```

---

## Embedding Issues

### "Insufficient capacity" error

**Problem:** Payload is too large for carrier.

**Solutions (in priority order):**
1. **Use larger carrier:**
   ```bash
   shadowforge analyze capacity image.png
   shadowforge analyze capacity larger_image.jpg
   ```

2. **Use different technique with higher capacity:**
   ```bash
   shadowforge embed secret.txt image.png --technique DCT
   # vs LSB - DCT typically has higher capacity for JPEG
   ```

3. **Reduce payload size:**
   ```bash
   # Compress payload first
   gzip secret.txt -c > secret.txt.gz
   shadowforge embed secret.txt.gz image.png
   ```

4. **Use distributed embedding:**
   ```bash
   shadowforge embed-distributed secret.txt cover1.png cover2.png cover3.png
   ```

### "Unsupported media format" error

**Problem:** Media format isn't supported.

**Solution:**
1. Check supported formats:
   ```bash
   shadowforge formats
   ```

2. Convert media to supported format:
   ```bash
   ffmpeg -i video.mp4 carrier.wav        # Convert to WAV
   ffmpeg -i image.bmp image.png          # Convert to PNG
   ```

### Embedded data doesn't survive image modifications

**Problem:** Stego image was edited (resized, rotated, cropped).

**Solutions:**
- **PNG/BMP (LSB):** Very sensitive to any modifications
  - Don't edit after embedding
  - Use original quality images

- **JPEG (DCT):** Somewhat resilient
  - But re-compression destroys data
  - Use export settings that preserve quality

- **WAV (LSB):** Sensitive to re-encoding
  - Don't convert sample rate or bit depth

**Prevention:**
```bash
# Verify image wasn't modified
md5sum image.png > image.png.md5
# Later, verify:
md5sum -c image.png.md5
```

### "Invalid carrier media" error

**Problem:** Media file is corrupted or invalid.

**Solutions:**
```bash
# Verify file integrity
file image.png           # Check file type
identify image.png       # Detailed image info (requires ImageMagick)
ffprobe -i audio.wav     # Audio info

# Repair or convert
ffmpeg -i corrupted.jpg -q:v 2 fixed.jpg
```

---

## Extraction Issues

### "No hidden data detected" error

**Problem:** Can't find embedded data in carrier.

**Possible causes and solutions:**

1. **Wrong carrier file:**
   ```bash
   # Verify you're using the right file
   ls -la *.png
   # Use the one you embedded into
   ```

2. **Data was corrupted:**
   ```bash
   # Try different techniques
   shadowforge extract stego.png --technique LSB
   shadowforge extract stego.png --technique DCT
   ```

3. **Carrier was modified:**
   ```bash
   # Re-embedding to same carrier will corrupt data
   # Use original carrier, not a copy
   ```

4. **Wrong password (if encrypted):**
   ```bash
   shadowforge extract stego.png --password your_password
   ```

### "Corrupted data - checksum failed" error

**Problem:** Extracted data failed integrity check.

**Solutions:**
1. **Verify carrier integrity:**
   ```bash
   # Check file wasn't modified
   md5sum stego.png
   # Should match what you got after embedding
   ```

2. **Try with error correction:**
   ```bash
   # Shadowforge auto-uses Reed-Solomon, but can be tuned
   shadowforge extract stego.png --redundancy high
   ```

3. **Carrier was partially corrupted:**
   - Use distributed embedding (you can lose some carriers)
   - Re-embed to reliable carriers

### "Wrong password" error

**Problem:** Incorrect decryption password.

**Solutions:**
```bash
# Ensure password is correct
shadowforge extract stego.png --password "correct password"

# If you forgot password:
# There's no recovery - steganography hides data, not protects keys
```

---

## Distributed Operations

### "Insufficient shards for recovery" error

**Problem:** Don't have enough carriers for threshold.

**Example:**
```bash
# You embedded across 10 images with threshold 7
# But only have 5 images
# Need 7+, so recovery fails
```

**Solutions:**
1. **Locate missing carriers:**
   - Check the manifest file
   - It lists which shard is in which carrier

2. **Use lower threshold next time:**
   ```bash
   shadowforge embed-distributed secret.txt \
     cover1.png cover2.png cover3.png \
     --threshold 2  # Lower threshold, fewer carriers needed
   ```

3. **Account for potential loss:**
   - With 10 carriers and threshold 7, you can lose 3
   - With 5 carriers and threshold 3, you can lose 2

### "Manifest validation failed" error

**Problem:** Manifest file is corrupted or tampered.

**Solutions:**
```bash
# If you have the original manifest:
cp backup/manifest.json manifest.json

# Or re-extract with the carriers you have
# (though you'll need to guess correct carriers)
```

### Parallel extraction is slow

**Problem:** Distributed extraction isn't fast enough.

**Solutions:**
```bash
# Increase worker count
shadowforge extract-distributed *.png \
  --workers 8          # More workers (default: CPU count)
  --parallel true      # Explicit parallel mode
```

---

## Archive Operations

### "Cannot extract archive: zip slip vulnerability detected"

**Problem:** Archive contains suspicious file paths.

**Solution:**
```bash
# This is a security feature - archive was rejected
# The archive may have been maliciously crafted
# Use trusted archives only
```

### "Archive extraction failed" error

**Problem:** Archive file is corrupted.

**Solutions:**
```bash
# Repair archive
zip -F archive.zip --out archive-fixed.zip   # ZIP repair
tar -tzf archive.tar.gz > /dev/null          # TAR verify

# Use unzipper tool for detailed diagnostics
unzip -t archive.zip
```

---

## Performance Issues

### Embedding takes very long time

**Problem:** Operation is slow.

**Possible causes:**
1. Very large image (e.g., 4K)
2. Very large payload
3. Complex distribution pattern
4. Slow storage (USB drive, network drive)

**Solutions:**
```bash
# Profile the operation
time shadowforge embed secret.txt image.png

# Use faster storage if available
# Move to local SSD instead of USB

# Check system resources
# High CPU/memory usage might be limiting
```

### Distributed embedding doesn't parallelize

**Problem:** Not using multiple CPUs.

**Solution:**
```bash
# Explicit parallel setting
shadowforge embed-distributed secret.txt *.png --parallel true

# Check CPU usage
# On macOS: Activity Monitor
# On Linux: top, htop
# On Windows: Task Manager
```

---

## Security Concerns

### "Statistical anomaly detected in carrier"

**Warning:** Carrier quality is low.

**Solutions:**
```bash
# Use a different, higher-quality image
shadowforge analyze detectability image.png
# Score shows carrier quality

# Better carrier sources:
# - Original, uncompressed photos
# - High-resolution stock photos
# - Avoid memes, screenshots, previously edited images
```

### Worried about detection

**Tips:**
1. **Use high-quality carriers:**
   - Original photos from camera
   - Stock photography
   - High-resolution images

2. **Don't embed maximum capacity:**
   - Use 50% safe capacity, not 100%
   ```bash
   shadowforge analyze capacity image.png
   # Shows safe capacity
   ```

3. **Mix with normal content:**
   - Not every image is suspicious
   - Have legitimate reasons to send images

4. **Don't re-use carriers:**
   - Use unique images each time
   - Patterns are detectable

### Data protection concerns

**Remember:**
- Steganography hides data, doesn't encrypt it
- For sensitive data, encrypt BEFORE embedding:
  ```bash
  # Encrypt first
  gpg --symmetric --cipher-algo AES256 secret.txt
  # Then embed
  shadowforge embed secret.txt.gpg image.png
  ```

---

## System-Specific Issues

### macOS: "cannot be verified" on launch

**Problem:** Gatekeeper blocks unsigned binary.

**Solution:**
```bash
# Allow execution
xattr -d com.apple.quarantine /path/to/shadowforge

# Or install via Homebrew (signed)
brew install greysquirr3l/tap/shadowforge
```

### Linux: Missing dependencies

**Problem:** Binary won't run (missing libc).

**Solution:**
```bash
# Check dependencies
ldd $(which shadowforge)

# Install missing libraries (Ubuntu):
sudo apt-get install libc6 libssl-dev
```

### Windows: Antivirus blocks execution

**Problem:** Windows Defender or antivirus flags binary.

**Solution:**
1. Build from source (antivirus trusts compiled code from source)
2. Add to antivirus whitelist
3. Contact antivirus vendor about false positive

---

## Getting Help

If your issue isn't listed:

1. **Enable debug logging:**
   ```bash
   SHADOWFORGE_LOG_LEVEL=debug SHADOWFORGE_LOG_MODE=cli shadowforge <command>
   ```

2. **Collect diagnostic info:**
   ```bash
   shadowforge version
   go version
   uname -a  # System info
   ```

3. **Check documentation:**
   - [Commands Reference](./commands/README.md)
   - [FAQ](./faq.md)

4. **Report issue:**
   - [GitHub Issues](https://github.com/greysquirr3l/shadowforge/issues)
   - Include: command, error message, debug output, system info

---

**Found a bug? Have a fix?** Contribute on [GitHub](https://github.com/greysquirr3l/shadowforge)!

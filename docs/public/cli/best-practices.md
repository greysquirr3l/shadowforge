# Best Practices

Security and operational guidance for using Shadowforge effectively.

## Security Fundamentals

### 1. Carrier Media Selection

**Choose media carefully:**
- Use high-quality, original media from trusted sources
- Avoid media that has been re-compressed or edited
- Ensure carriers are large enough for your payload

```bash
# Check capacity before embedding
shadowforge analyze capacity --input media.png --payload-size <bytes>
```

**Media quality matters:**
- **Lossless formats** (PNG, BMP, WAV) preserve hidden data perfectly
- **Lossy formats** (JPEG, MP3) can corrupt data through re-compression
- **Avoid** media that will be edited after embedding

### 2. Payload Preparation

**Encrypt sensitive data before embedding:**
```bash
# Encrypt payload first, then embed
openssl enc -aes-256-cbc -in secret.txt -out secret.enc
shadowforge embed --input secret.enc --cover media.png --output stego.png
```

**Minimize payload size:**
- Compress before embedding: `gzip secret.txt`
- Remove unnecessary metadata
- Smaller payloads = lower detectability

### 3. Secure File Handling

**Protect original files:**
```bash
# Keep originals separate from stego media
mkdir -p secure/originals secure/stego

# Embed from secure location
shadowforge embed \
  --input secure/originals/secret.txt \
  --cover secure/originals/carrier.png \
  --output secure/stego/output.png

# Securely delete temporary files
shred -vfz -n 3 /tmp/secret-copy.txt
```

**Avoid leaving traces:**
- Don't store stego files in obvious locations
- Use encrypted storage for sensitive media
- Clear shell history if commands contain secrets

```bash
# Clear sensitive commands from history
history -c
export HISTCONTROL=ignorespace  # Subsequent commands with leading space won't be saved
```

## Operational Security

### 4. Distribution Patterns

**Choose the right pattern for your use case:**

| Pattern | Use Case | Risk |
|---------|----------|------|
| 1:1 | One-time sharing | Carrier loss = total failure |
| 1:N | Critical data | Distributed, resilient |
| N:1 | Bundling | All eggs in one basket |
| N:M | Enterprise | Complex, flexible |

**For critical data, use 1:N with redundancy:**
```bash
# Create resilient distribution
shadowforge embed-distributed \
  --input critical.zip \
  --cover img1.png,img2.png,img3.png,img4.png,img5.png \
  --threshold 3 \
  --redundancy 40%

# Now you can lose 2 images and still recover
```

### 5. Manifest Files

**Always use manifests for distributed embedding:**
```bash
# Manifests track which shard is where
shadowforge embed-distributed \
  --input data.zip \
  --cover *.png \
  --output-dir distributed/ \
  --manifest

# Later, use manifest for recovery
shadowforge extract-distributed \
  --input-dir distributed/ \
  --manifest distribution.manifest \
  --output recovered.zip
```

**Protect manifests carefully:**
- Store separately from stego media
- Use encryption for sensitive distributions
- Share manifest only with intended recipients

### 6. Password Management

**For additional encryption:**
```bash
# Use strong, random passwords
openssl rand -base64 32  # Generate random password

# Embed with password
shadowforge embed \
  --input secret.txt \
  --cover image.png \
  --output stego.png \
  --password "$(openssl rand -base64 32)"

# Share password through secure channel (different from media)
```

**Never hardcode passwords:**
```bash
# ❌ DON'T: Password visible in shell history
shadowforge embed --input data --cover image.png --output stego.png --password secret

# ✓ DO: Prompt for password
shadowforge embed --input data --cover image.png --output stego.png --password
# (will prompt securely)

# ✓ DO: Use environment variable
shadowforge embed --input data --cover image.png --output stego.png --password "$STEGO_PASSWORD"
```

## Detection Avoidance

### 7. Quality and Detection Risk

**Understand trade-offs:**
- **High capacity** = Higher detectability risk
- **High quality** = Lower detectability risk
- **Small payloads** = Minimal impact on carrier

```bash
# Use lower capacity for lower detectability
shadowforge embed \
  --input secret.txt \
  --cover large-image.png \
  --output stego.png \
  --quality 95  # High quality, low capacity

# vs.

shadowforge embed \
  --input secret.txt \
  --cover image.png \
  --output stego.png \
  --quality 50  # Lower quality, higher capacity
```

### 8. Technique Selection

**Different techniques have different properties:**

| Technique | Capacity | Detectability | Best For |
|-----------|----------|----------------|----------|
| LSB | High | Low-Medium | Images, best balance |
| DCT | Medium | Very Low | JPEG, when lossy OK |
| Palette | Medium | Low | GIF, indexed color |
| Phase | Medium | Low | Audio, imperceptible |
| Zero-Width | Low | Very Low | Text, invisible |

**Strategy:**
- Use LSB for images (good capacity, low detectability)
- Use DCT for JPEG (excellent detectability properties)
- Use Phase for audio (imperceptible)
- Use Zero-Width for text (invisible)

```bash
# Analyze detectability before committing
shadowforge analyze detectability --input stego.png
```

## Backup and Recovery

### 9. Resilient Distribution

**For critical data, use threshold recovery:**
```bash
# Distribute across many carriers
shadowforge embed-distributed \
  --input critical-data.zip \
  --cover *.png \
  --threshold 5 \
  --redundancy 30%

# You can recover with any 5 images
# Even if half are lost or corrupted
```

### 10. Verification Before Distribution

**Always verify before sharing:**
```bash
# Test extraction before sending stego files
shadowforge extract --input stego.png --output test.txt --verify

# Confirm recovered content matches
diff test.txt original.txt

# Only then distribute the stego file
```

## Compliance and Legal

### 11. Jurisdiction Awareness

**Understand local laws:**
- Some jurisdictions restrict steganography
- Encryption laws vary by country
- Data privacy regulations (GDPR, etc.) may apply

**Recommendations:**
- Understand relevant laws in your jurisdiction
- Document legitimate use cases
- Use only for authorized purposes
- Respect others' data privacy rights

### 12. Audit and Documentation

**Keep records for legitimate use:**
```bash
# Log operations (metadata only, no secrets)
shadowforge embed \
  --input secret.txt \
  --cover image.png \
  --output stego.png \
  --verbose 2>&1 | tee embed-operation.log
```

**What to document:**
- Date and time of operations
- File sizes and types (no names)
- Techniques used
- Purpose (general: "data backup", "secure transfer", etc.)

**What NOT to document:**
- Secret content or names
- Recipient identity
- Specific business implications
- Operational security details

## Practical Workflows

### Scenario 1: Secure Document Transfer

```bash
# 1. Encrypt sensitive document
gpg --symmetric document.pdf

# 2. Embed in carrier
shadowforge embed \
  --input document.pdf.gpg \
  --cover background-photo.png \
  --output letter.png

# 3. Send via untrusted channel
# Send letter.png via email/chat/cloud

# 4. Recipient extracts
shadowforge extract --input letter.png --output document.pdf.gpg

# 5. Decrypt
gpg document.pdf.gpg
```

### Scenario 2: Resilient Archive

```bash
# 1. Prepare data
zip -r critical-data.zip important-files/

# 2. Distribute across multiple images
shadowforge embed-distributed \
  --input critical-data.zip \
  --cover photo1.png,photo2.png,photo3.png,photo4.png,photo5.png \
  --threshold 3 \
  --redundancy 40%

# 3. Store or distribute carriers separately
cp stego*.png /external-drive/backup1/
cp stego*.png /cloud-storage/backup2/
cp stego*.png /office-server/backup3/

# 4. Later, recover from any 3 images
shadowforge extract-distributed \
  --input /backup2/stego1.png,/backup2/stego3.png,/backup2/stego4.png \
  --output recovered.zip
```

### Scenario 3: Archived Distribution

```bash
# 1. Embed in multiple images
shadowforge embed-batch \
  --input document1.pdf,document2.pdf,document3.pdf \
  --cover container.png \
  --output stego.png

# 2. Archive for distribution
shadowforge archive create \
  --input stego.png \
  --format tar.gz \
  --output package.tar.gz

# 3. Share package
scp package.tar.gz recipient@server:~/

# 4. Recipient unpacks and extracts
tar xzf package.tar.gz
shadowforge extract-batch \
  --input stego.png \
  --output-dir extracted/
```

## Monitoring and Maintenance

### 13. System Monitoring

**Watch for suspicious activity:**
- Unexpected file access patterns
- Unusual network activity
- System resource usage spikes

**Secure your system:**
- Keep OS and tools updated
- Use firewalls and intrusion detection
- Monitor file integrity (e.g., with aide or tripwire)

### 14. Regular Testing

**Test recovery procedures:**
```bash
# Periodically test your backup recovery
shadowforge extract-distributed \
  --input test-carriers/*.png \
  --output verify-recovery.zip

# Verify content integrity
unzip verify-recovery.zip
# Check all files are intact
```

---

**Remember:** Security is a process, not a destination. Stay informed, keep systems updated, and regularly review your practices.

See [Troubleshooting](./troubleshooting.md) for common issues and solutions.

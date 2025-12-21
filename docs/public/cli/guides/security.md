# Security Hardening Guide

> **Secure Your Steganographic Operations**

Comprehensive security guidance for production Shadowforge deployments.

## Pre-Embedding Security Checklist

- [ ] Generate new key pair (never reuse keys across operations)
- [ ] Verify carrier media authenticity
- [ ] Validate payload integrity before embedding
- [ ] Check available storage space (prevent capacity overflow)
- [ ] Review detectability score (keep below 0.7)
- [ ] Test extraction with same technique before production
- [ ] Document key fingerprint for recovery
- [ ] Create secure backup of keys

## Cryptographic Security

### Key Generation Best Practices

```bash
# Generate secure Kyber-1024 key pair
shadowforge keygen --algorithm kyber1024 --key-id production-2025

# Verify key strength
shadowforge keyexport --key-id production-2025 --show-stats
```

**Key Strength Verification:**
- Kyber-1024: Post-quantum secure (NIST approved)
- 3072-bit equivalent classical strength
- Resistant to quantum attacks

### Key Storage Security

```bash
# Store private key in secure location
mkdir -p ~/.shadowforge/keys
chmod 700 ~/.shadowforge/keys

# Export key with encryption
shadowforge keyexport --key-id production-2025 \
  --format pem-encrypted \
  --password-protected \
  --output ~/.shadowforge/keys/production-2025.pem
```

**Security Recommendations:**
- Keep private keys offline when possible
- Use hardware security modules (HSM) for enterprise
- Rotate keys annually
- Maintain secure backups (encrypted, geographically distributed)
- Never share private keys

## Payload Security

### Pre-Embedding Validation

```bash
# Verify payload integrity
shadowforge validate --payload secret.txt

# Check for metadata that might leak information
file secret.txt
exiftool secret.txt

# Remove metadata before embedding
# For documents: strip EXIF data
# For archives: recreate without metadata
```

### Payload Encryption

```bash
# Double-encrypt sensitive payloads
# First: Shadowforge's built-in encryption (Kyber-1024)
# Second: External encryption layer

# External layer example (OpenSSL)
openssl enc -aes-256-cbc -in secret.txt -out secret.bin

# Then embed the encrypted file
shadowforge embed \
  --payload secret.bin \
  --cover image.png \
  --technique lsb
```

### Payload Size Awareness

```bash
# Always use safe capacity margins
# Real capacity = calculated_capacity * 0.7

shadowforge analyze capacity --technique lsb --cover image.png

# If capacity is 100KB:
# Safe to embed: 70KB (not 100KB)
# Extra margin prevents detectability
```

## Carrier Selection Security

### Avoiding Carrier Artifacts

```bash
# Analyze carrier quality BEFORE selecting
shadowforge analyze capacity \
  --cover image.png \
  --detailed \
  --show-artifacts

# Reject carriers with:
# - Unnatural color distributions
# - Compression artifacts
# - Regular patterns
```

### Carrier Diversification

```bash
# Use varied carrier types to avoid patterns
# Instead of: 20 similar PNG images
# Better: 8 PNG + 7 JPEG + 5 WAV

shadowforge select ./mixed-media \
  --payload-size 50000 \
  --diversity-mode balanced \
  --min-carriers-per-type 3
```

### Carrier Authentication

For enterprise deployments:

```bash
# Verify carrier source and integrity
# Check file hash before use
sha256sum image1.png > image1.png.sha256
# Verify later with: sha256sum -c image1.png.sha256

# Verify carrier creation date (avoid suspiciously "new" files)
stat image1.png | grep Modify
```

## Distribution Security

### Secure Multi-Carrier Distribution

```bash
# Use distributed embedding for resilience
# Even if one carrier is compromised, data is unrecoverable

# Create distribution with high redundancy
shadowforge embed-distributed \
  --payload secret.bin \
  --data-shards 10 \
  --parity-shards 6 \
  --covers cover1.png cover2.png ... cover16.png

# Distribute carriers across different physical locations
# Carrier 1-5: Location A
# Carrier 6-10: Location B
# Carrier 11-16: Location C
```

### Manifest Security

```bash
# Manifests contain shard mapping information
# Protect with same care as encryption keys

# The manifest is HMAC-protected
# Never share manifest with carriers
# Keep manifest in secure location

# Verify manifest integrity before extraction
shadowforge analyze manifest \
  --manifest-file distribution.manifest \
  --verify-signature
```

## Detection Prevention

### Statistical Analysis

```bash
# Analyze detectability BEFORE embedding
shadowforge analyze detectability \
  --technique lsb \
  --cover image.png \
  --payload-size 5000

# Acceptable scores:
# < 0.30: Very secure (stealth excellent)
# 0.30-0.50: Secure (stealth good)
# 0.50-0.70: Acceptable (stealth fair)
# > 0.70: Risky (detectability high)
```

### Statistical Stealth

```bash
# Use capacity-aware embedding
# Don't maximize capacity - preserve statistics

# Bad: Embed 100KB in 100KB capacity
# Good: Embed 70KB in 100KB capacity (30% margin)

shadowforge embed \
  --payload secret.txt \
  --cover image.png \
  --technique phase \
  --capacity-usage 0.7  # Use only 70% of available capacity
```

### Technique Selection for Stealth

| Technique | Stealth | Speed | Notes |
|-----------|---------|-------|-------|
| LSB | Good | Very Fast | Use with margin |
| DCT | Excellent | Medium | JPEG artifacts |
| Phase | Excellent | Medium | Audio only |
| Echo | Excellent | Fast | Audio only |
| Zero-Width | Perfect | Very Fast | Text only |
| Palette | Good | Very Fast | GIF/PNG only |

## Operational Security (OPSEC)

### Command Line Security

```bash
# Never use passwords on command line (visible in process list)
# Use interactive prompts instead

# Bad:
# shadowforge keygen --password mypassword

# Good:
shadowforge keygen  # Prompts securely for password
```

### File System Security

```bash
# Ensure proper permissions on Shadowforge directories
chmod 700 ~/.shadowforge
chmod 600 ~/.shadowforge/keys/*
chmod 600 ~/.shadowforge/config

# Use secure temporary storage
export TMPDIR=/dev/shm  # RAM-based, not disk
shadowforge embed --payload secret.txt --cover image.png
unset TMPDIR
```

### Log Security

```bash
# Shadowforge uses slog for logging
# By default, logs don't contain sensitive data

# In production, redirect logs securely
shadowforge --log-level info embed ... 2>/var/log/shadowforge/app.log

# Secure log files
chmod 600 /var/log/shadowforge/app.log
```

## Incident Response

### If a Carrier is Compromised

1. **Assess impact:**
   - Single carrier: Unrecoverable (with proper distribution)
   - Multiple carriers: Risk depends on K threshold
   - Extraction: Risk if manifest also compromised

2. **Mitigation:**
   ```bash
   # If using distributed embedding with N carriers, K threshold:
   # Loss of up to (N-K) carriers is acceptable

   # Example: 10 data + 5 parity = 15 total
   # Can recover with any 10 carriers
   # Loss of up to 5 carriers is safe
   ```

3. **Recovery:**
   ```bash
   # Extract from available carriers
   shadowforge extract-distributed \
     --manifest manifest.json \
     --stego-files available1.png available2.png ... \
     --output recovered.bin
   ```

### If a Key is Compromised

1. **Immediate actions:**
   - Revoke the compromised key
   - Generate new key pair
   - Re-embed sensitive data with new key

2. **Long-term:**
   - Review all data encrypted with compromised key
   - Re-encrypt and re-distribute if possible
   - Document incident

### If a Manifest is Compromised

1. **Risk assessment:**
   - Manifest only contains shard mappings
   - Does not contain keys or payload
   - Knowledge of shard layout increases detection risk

2. **Mitigation:**
   - Redistribute carriers with different shard layout
   - Consider manifest as "burned"
   - Generate new distribution with new manifest

## Network Security

### If Using Remote Storage

```bash
# Always use encrypted connections
# TLS 1.3 or higher for all communications

# Example: SFTP instead of FTP
sftp user@secure-server.com
put stego-files/* remote-location/

# Never transmit carrier files and manifest together
# Transmit carriers on one route, manifest on another
```

## Compliance Considerations

### Audit Logging

```bash
# Log all embedding/extraction operations
# Include: timestamp, technique, payload-size, carrier-list

# Example audit entry:
# 2025-12-21 14:30:45 | EMBED | technique:lsb | payload:1000 | carriers:1 | user:admin
```

### Data Retention

```bash
# Securely delete sensitive files after embedding
# Use secure deletion tools

# On macOS:
rm -P secret.txt  # Overwrite with random data before deleting

# On Linux:
shred -vfz -n 10 secret.txt  # 10 passes

# Better: Use encrypted containers
brew install ecryptfs-utils
# Or use LUKS encrypted volumes
```

## Verification Checklist

Before production deployment:

- [ ] Keys generated with Kyber-1024
- [ ] Keys backed up securely
- [ ] Payload encrypted (double encryption for critical data)
- [ ] Carriers validated for authenticity
- [ ] Detectability score < 0.70
- [ ] Distribution uses adequate redundancy
- [ ] Manifest stored separately from carriers
- [ ] Operational security procedures documented
- [ ] Team trained on security procedures
- [ ] Incident response plan established

## Additional Resources

- Post-Quantum Cryptography: [NIST PQC FAQ](https://csrc.nist.gov/projects/post-quantum-cryptography/)
- Steganography Security: Academic research on stego-steganalysis
- OPSEC Best Practices: NSA CISA Security Guidelines

---

*Security documentation for Shadowforge v1.0+*
*Last Updated: December 2025*

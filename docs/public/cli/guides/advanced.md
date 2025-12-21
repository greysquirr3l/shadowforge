# Advanced Usage Guide

> **Master Shadowforge's Advanced Capabilities**

This guide covers advanced techniques and patterns for experienced Shadowforge users.

## Advanced Embedding Strategies

### Multi-Carrier Distribution

Distribute sensitive data across multiple carriers for resilience:

```bash
# Create a distributed embedding plan
shadowforge select ./covers --payload-size 50000

# Execute distribution with specific technique
shadowforge embed-distributed \
  --payload secret.bin \
  --technique lsb \
  --data-shards 10 \
  --parity-shards 5 \
  --covers cover1.png cover2.png cover3.png cover4.png cover5.png cover6.png
```

### Technique Composition

Chain multiple techniques for enhanced security:

```bash
# Sequential chaining: LSB → DCT → Zero-Width
shadowforge chain create my-chain \
  --mode sequential \
  --payload secret.txt \
  --techniques lsb,dct,zero-width \
  --covers image1.png,image2.jpg,text.txt

# Execute the chain
shadowforge chain execute my-chain
```

### Layered Embedding

Embed different data portions with different techniques in the same carrier:

```bash
# Layered approach: sensitive data in phase encoding, metadata in LSB
shadowforge chain create layered-approach \
  --mode layered \
  --payload secret.bin \
  --weights 0.7,0.3 \
  --techniques phase,lsb \
  --cover audio.wav

shadowforge chain execute layered-approach
```

## Advanced Extraction Workflows

### Partial Recovery

Recover data when some carriers are unavailable:

```bash
# Extract from available carriers with K-of-N recovery
shadowforge extract-distributed \
  --manifest manifest.json \
  --stego-files stego1.png stego3.png stego5.png \
  --output recovered.bin
```

### Archive-Based Distribution

Protect multiple files with single distribution operation:

```bash
# Create password-protected archive first
shadowforge archive create \
  --output secret-archive.zip \
  --password "secure-password" \
  --files document1.pdf document2.docx images/photo.jpg

# Then distribute the archive
shadowforge embed-distributed \
  --payload secret-archive.zip \
  --data-shards 8 \
  --parity-shards 4 \
  --covers cover1.png cover2.png ... cover12.png
```

## Performance Optimization

### Parallel Processing

Shadowforge automatically uses all available CPU cores for:
- Multi-carrier embedding (worker pool)
- Archive processing
- Large file handling

Monitor with: `shadowforge analyze capacity --verbose`

### Memory-Efficient Handling

For large payloads (>100MB):

```bash
# Use streaming processing
shadowforge embed \
  --input large-file.bin \
  --cover cover.png \
  --technique lsb \
  --streaming true
```

## Security Best Practices for Advanced Use

### Key Management

```bash
# Generate high-security key pair
shadowforge keygen --algorithm kyber1024

# Export public key for sharing
shadowforge keyexport --key-id primary --format pem

# Import trusted keys
shadowforge keyimport --key public.pem --trust-level high
```

### Statistical Analysis

```bash
# Analyze detectability before embedding
shadowforge analyze detectability \
  --technique lsb \
  --cover image.png \
  --payload-size 5000 \
  --verbose

# Get detailed security report
shadowforge analyze capacity \
  --technique phase \
  --cover audio.wav \
  --detailed
```

### Steganographic Signature

Add metadata for integrity verification:

```bash
# Embed with signature
shadowforge embed \
  --payload secret.txt \
  --cover image.png \
  --technique dct \
  --add-signature \
  --sign-with private.key
```

## Real-World Scenarios

### Scenario 1: Secure Document Distribution

Distribute a classified document to multiple recipients:

```bash
# 1. Create encrypted archive
shadowforge archive create \
  --output classified.zip \
  --files document.pdf photos/

# 2. Generate distribution plan
shadowforge select ./covers --payload-size $(stat -f%z classified.zip)

# 3. Distribute across covers
shadowforge embed-distributed \
  --payload classified.zip \
  --data-shards 15 \
  --parity-shards 8 \
  --covers photo1.jpg photo2.jpg ... photo23.jpg

# 4. Each recipient gets subset of carriers
# Recipient A: carriers 1,2,3,5,6,8
# Recipient B: carriers 1,4,5,9,10,12
# Recipient C: carriers 2,3,7,11,13,15
```

### Scenario 2: Resilient Backup System

Create a resilient backup across multiple locations:

```bash
# 1. Create selection of available media across locations
shadowforge scan /location1 --recursive
shadowforge scan /location2 --recursive
shadowforge scan /location3 --recursive

# 2. Execute distribution with redundancy
shadowforge embed-distributed \
  --payload backup.tar.gz \
  --data-shards 12 \
  --parity-shards 8 \
  --covers /location1/*.jpg /location2/*.png /location3/*.jpg
```

### Scenario 3: Covert Communication Channel

Establish a hidden communication channel:

```bash
# Sender side
shadowforge chain create secure-channel \
  --mode sequential \
  --payload message.txt \
  --techniques phase,lsb,echo \
  --covers audio1.wav audio2.wav audio3.wav

shadowforge chain execute secure-channel

# Recipient side
shadowforge chain extract secure-channel \
  --stego-files received1.wav received2.wav received3.wav \
  --output message.txt
```

## Troubleshooting Advanced Operations

### Insufficient Capacity Issues

```bash
# Get detailed capacity report
shadowforge analyze capacity \
  --technique phase \
  --cover audio.wav \
  --detailed

# Get suggestions for resolution
shadowforge suggest --required 50000 --available 30000
```

### Extraction Failures

```bash
# Verify stego file integrity
shadowforge validate \
  --stego-file stego.png \
  --strict

# Try recovery with increased tolerance
shadowforge extract \
  --stego-file stego.png \
  --recovery-mode aggressive
```

## Performance Benchmarks

Typical performance on modern hardware:

| Operation | Technique | Carrier | Time |
|-----------|-----------|---------|------|
| Embed 1MB | LSB | PNG 10MB | 45ms |
| Embed 1MB | DCT | JPEG 500KB | 120ms |
| Embed 1MB | Phase | WAV 10MB | 180ms |
| Extract 1MB | LSB | PNG 10MB | 15ms |
| Extract 1MB | DCT | JPEG 500KB | 35ms |

## Advanced Configuration

Create a `shadowforge.conf` file in your project:

```yaml
embedding:
  default_technique: lsb
  data_shards: 10
  parity_shards: 5
  quality_target: 0.95

analysis:
  detailed_stats: true
  chi_square_threshold: 0.05

security:
  sign_all_payloads: true
  verify_signatures: true
  min_key_strength: 2048
```

Load custom configuration:

```bash
shadowforge --config shadowforge.conf embed --payload secret.txt --cover image.png
```

## Advanced Topics

### Quantum-Resistant Guarantees

Shadowforge uses post-quantum cryptography (Kyber-1024, Dilithium3), providing security against future quantum computers.

### Information Leakage Analysis

For maximum security, analyze information leakage:

```bash
shadowforge analyze detectability \
  --technique lsb \
  --cover image.png \
  --payload-size 5000 \
  --statistical-tests chi-square,rs-analysis,histogram
```

### Custom Embedding Strategies

Combine techniques strategically:

- **LSB**: Fast, large capacity, moderate security
- **DCT**: Medium capacity, image artifacts, good security
- **Phase**: Imperceptible, audio-only, excellent security
- **Zero-Width**: Invisible, text-only, metadata-suitable

Choose based on:
1. Medium type (image/audio/text)
2. Required capacity
3. Security requirements
4. Detectability tolerance

---

*Advanced usage documentation for Shadowforge v1.0+*

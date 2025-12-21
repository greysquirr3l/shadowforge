# Glossary

> **Shadowforge Terminology and Concepts**

## Core Concepts

**Steganography**
- Practice of hiding data within non-secret data in such a way that the very existence of the hidden data is concealed
- Different from cryptography (which obscures data) - steganography hides data
- Goal: Covert communication with minimal detectability

**Stego (Stego Media)**
- The output file containing both the original carrier and the hidden payload
- Appears to be innocent media (image, audio, text) to observers
- Contains encrypted payload + Reed-Solomon parity data

**Carrier / Cover Media**
- The original innocuous media file used to hide data
- Can be image (PNG, JPEG, GIF), audio (WAV), or text (TXT)
- Selected to have sufficient capacity for the payload

**Payload**
- The secret data being hidden
- Encrypted using Kyber-1024 KEM before embedding
- Encoded with Reed-Solomon error correction before steganographic embedding

## Cryptography

**Post-Quantum Cryptography (PQC)**
- Cryptographic algorithms believed to be resistant to quantum computer attacks
- Kyber-1024 and Dilithium3 are NIST-approved post-quantum standards
- Shadowforge uses these for quantum-resistant security

**Kyber-1024**
- Post-quantum key encapsulation mechanism (KEM)
- Generates ephemeral shared secrets for AES-256 encryption
- Equivalent to 3072-bit RSA in classical security strength
- NIST standard for public-key encryption

**Dilithium3**
- Post-quantum digital signature algorithm
- Used for payload signing and integrity verification
- NIST standard for digital signatures
- Resistant to future quantum attacks

**Key Pair**
- Public key: Can be shared openly, used for encryption
- Private key: Must be kept secret, used for decryption and signing

**Encapsulation**
- Process of generating a shared secret using a public key
- Used in Kyber to establish encryption key for AES-256
- Different encapsulation produces different ciphertext from same message

## Error Correction

**Reed-Solomon (RS) Encoding**
- Forward error correction code used for data resilience
- Creates parity shards allowing recovery from missing or corrupted data
- Shadowforge uses RS to enable multi-carrier distributions

**Shard**
- One piece of Reed-Solomon encoded data
- Multiple shards can reconstruct original data if K shards are available
- Total shards = data shards + parity shards

**K-of-N Threshold**
- Recovery requires K (minimum) of N (total) shards
- Example: 10-of-15 means any 10 of 15 shards can recover data
- Redundancy = (N-K) / N

**Redundancy Level**
- Percentage of additional shards for error correction
- 30% redundancy = 3 parity shards per 10 data shards
- Higher redundancy = better loss tolerance, larger data overhead

## Steganography Techniques

**LSB (Least Significant Bit)**
- Modifies the least significant bit of pixel/sample values
- Imperceptible changes to human perception
- Works with: Images (PNG, JPEG, BMP, GIF), Audio (WAV)
- Capacity: High, Security: Medium

**DCT (Discrete Cosine Transform)**
- Modifies coefficients in DCT-transformed JPEG blocks
- Middle-frequency coefficients for imperceptibility
- Works with: JPEG images
- Capacity: Medium, Security: High

**Phase Encoding**
- Modifies phase spectrum of audio using FFT
- Preserves magnitude while encoding in phase differences
- Works with: Audio (WAV)
- Capacity: Medium, Security: Excellent

**Echo Hiding**
- Encodes data using different echo delays
- Binary 0 = short delay, Binary 1 = long delay
- Works with: Audio (WAV)
- Capacity: Low, Security: Excellent

**Zero-Width Characters**
- Uses invisible Unicode characters (ZWSP, ZWJ) to encode data
- Completely invisible to human reading
- Works with: Text (TXT, documents)
- Capacity: Very Low, Security: Perfect for metadata

**Palette-Based**
- Reorders color palette or modifies palette entries
- Works with: GIF, PNG (with indexed color)
- Capacity: Medium, Security: Good

## Distribution Patterns

**One-to-One (1:1)**
- Single payload embedded in single carrier
- Basic distribution pattern
- Simplest to use, good for small data

**One-to-Many (1:N)**
- Single payload split across multiple carriers using Reed-Solomon
- Data shards + parity shards distributed
- Enables resilience: loss of parity shards is acceptable
- Good for distribution resilience

**Many-to-One (N:1)**
- Multiple payloads aggregated into single carrier
- Creates payload index for selective extraction
- Good for multiple small files

**Many-to-Many (N:M)**
- Multiple payloads distributed across multiple carriers
- Complex distribution requiring sophisticated coordination
- Good for scenarios requiring both redundancy and scalability

## Advanced Concepts

**Technique Chaining**
- Combining multiple steganography techniques
- Sequential: Apply one technique, then another
- Layered: Different portions with different techniques
- Split: Distribute across multiple carriers with different techniques

**Detectability Score**
- Numerical measure (0.0-1.0) of how detectable an embedding is
- < 0.30: Very secure (excellent stealth)
- 0.30-0.70: Acceptable security
- > 0.70: Risky (high detectability)
- Calculated using statistical analysis

**Chi-Square Analysis**
- Statistical test for detecting anomalies in data distribution
- Compares observed distribution to expected distribution
- High chi-square value indicates possible steganography

**Capacity (Embedding Capacity)**
- Amount of data that can be hidden in a carrier
- Depends on: technique, carrier type, carrier size, quality requirements
- Measured in bytes or percentage of carrier size
- Safe capacity = calculated capacity × 0.7

**Entropy**
- Measure of randomness/disorder in data
- High entropy: Data looks random, harder to analyze
- Low entropy: Data has patterns, easier to detect
- Steganography attempts to preserve entropy to avoid detection

## File Formats

**PNG (Portable Network Graphics)**
- Lossless image compression
- Suitable for LSB steganography (preserves exact pixels)
- Large file sizes for same resolution
- Ideal for high-capacity hiding

**JPEG**
- Lossy image compression
- Suitable for DCT steganography (works with DCT coefficients)
- Smaller file sizes
- Trade-off: Some loss tolerance in DCT modification

**GIF (Graphics Interchange Format)**
- Indexed color format (256 colors max)
- Suitable for palette-based steganography
- Good for animation sequences
- Limited to 256-color palette

**WAV (Waveform Audio)**
- Uncompressed audio format
- Suitable for LSB, Phase, and Echo steganography
- Large file sizes (176 KB/second at 44.1kHz stereo)
- Good capacity for audio steganography

**TXT (Text)**
- Plain text format
- Suitable for zero-width character steganography
- Very limited capacity
- Useful for metadata and side-channel communication

## Security Terms

**Detectability**
- Likelihood that hidden data presence can be detected
- Steganalysis: Attempting to detect steganography
- Shadowforge minimizes detectability through technique choice

**Steganalysis**
- Science of detecting steganography
- Uses statistical analysis to find anomalies
- Chi-square test, RS analysis, histogram analysis
- Shadowforge designs techniques resistant to steganalysis

**OPSEC (Operational Security)**
- Practices preventing information disclosure
- Never share passwords on command line
- Secure key storage and management
- Proper file deletion and access control

**Manifest**
- Metadata file describing shard distribution
- Lists which shards are in which carriers
- HMAC-protected against tampering
- Keep separately from carriers

**Deniability / Plausibility**
- Ability to deny hidden data exists
- Unencrypted stego carrier can contain plausible content only
- Not the same as encryption (encryption admits data exists)

## Operations

**Embedding**
- Process of hiding payload in carrier
- Steps: Encrypt payload → Encode with RS → Apply steganography
- Output: Stego file (innocent-looking carrier with hidden data)

**Extraction**
- Process of recovering payload from stego file
- Steps: Apply steganography extraction → Decode RS shards → Decrypt
- Output: Original payload (if no corruption beyond parity capacity)

**Analysis (Capacity)**
- Calculate how much data can hide in carrier
- Depends on technique and quality requirements
- Helps plan distribution patterns

**Analysis (Detectability)**
- Calculate detectability score
- Helps choose appropriate techniques
- Guides capacity usage (use less than max capacity)

**Distribution**
- Spreading data across multiple carriers
- Uses Reed-Solomon for resilience
- Enables one-to-many, many-to-one, many-to-many patterns

**Reconstruction**
- Recovering original data from distributed shards
- Uses Reed-Solomon decoding
- Tolerates loss up to (N-K) shards

## Abbreviations

| Abbreviation | Meaning | Context |
|--------------|---------|---------|
| PQC | Post-Quantum Cryptography | Security algorithms |
| KEM | Key Encapsulation Mechanism | Kyber operation |
| RSA | Rivest-Shamir-Adleman | Classical (vulnerable) encryption |
| AES | Advanced Encryption Standard | Symmetric encryption used by Kyber |
| LSB | Least Significant Bit | Image/audio steganography |
| DCT | Discrete Cosine Transform | JPEG steganography |
| FFT | Fast Fourier Transform | Used in phase encoding |
| RS | Reed-Solomon | Error correction coding |
| GIF | Graphics Interchange Format | Image format |
| PNG | Portable Network Graphics | Image format |
| JPEG | Joint Photographic Experts Group | Image format |
| WAV | Waveform Audio Format | Audio format |
| ZWSP | Zero-Width Space | Unicode character U+200B |
| ZWJ | Zero-Width Joiner | Unicode character U+200D |
| HMAC | Hash-based Message Authentication Code | Integrity protection |

## Measurement Units

- **Byte**: 8 bits of data
- **KB**: Kilobyte (1,024 bytes)
- **MB**: Megabyte (1,024 KB)
- **Capacity percentage**: Percentage of carrier size usable for data
- **Detectability score**: 0.0 (undetectable) to 1.0 (highly detectable)

---

*Glossary for Shadowforge v1.0+*
*Last Updated: December 2025*

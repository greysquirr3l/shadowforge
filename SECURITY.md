# Security Policy

## Supported Versions

We take security seriously. The following versions of Shadowforge are currently supported with security updates:

| Version | Supported          | Status |
| ------- | ------------------ | ------ |
| 0.7.x   | :white_check_mark: | Current stable release |
| < 0.7.0 | :x:                | No longer supported |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability in Shadowforge, please report it responsibly:

### Primary Contact

- **Email**: <s0ma@protonmail.com> (if available) or open a [private security advisory](https://github.com/greysquirr3l/shadowforge/security/advisories/new)
- **Response Time**: We aim to respond within 48 hours

### What to Include

Please include the following information in your report:

- **Description**: A clear description of the vulnerability
- **Impact**: The potential impact and severity
- **Steps to Reproduce**: Detailed steps to reproduce the issue
- **Proof of Concept**: Any PoC code or examples (if applicable)
- **Affected Versions**: Which versions are affected
- **Suggested Fix**: If you have a suggested remediation (optional)

### Security Scope

We are particularly interested in vulnerabilities related to:

- ✅ **Cryptographic Operations**: Kyber-1024, Dilithium3, key generation, encryption/decryption
- ✅ **Steganographic Techniques**: LSB, DCT, Phase, Echo, Palette, Text, LSB-Audio
- ✅ **Reed-Solomon Error Correction**: Encoding/decoding, shard recovery
- ✅ **Archive Handling**: Zip slip, path traversal, malicious archives
- ✅ **Input Validation**: Command injection, file path manipulation
- ✅ **Memory Safety**: Buffer overflows, memory leaks, sensitive data exposure
- ✅ **Side-Channel Attacks**: Timing attacks, cache attacks on cryptographic operations
- ✅ **Authentication/Authorization**: Key management, access control

### Out of Scope

The following are generally **not** considered security vulnerabilities:

- ❌ Vulnerabilities in dependencies (report to upstream projects)
- ❌ Social engineering attacks
- ❌ Denial of service through resource exhaustion (expected behavior for large files)
- ❌ Issues in unsupported versions (< 0.7.0)
- ❌ Theoretical attacks without practical exploitation

## Responsible Disclosure Policy

We follow a **90-day disclosure timeline**:

1. **Day 0**: You report the vulnerability privately
2. **Day 1-7**: We acknowledge receipt and begin investigation
3. **Day 7-30**: We develop and test a fix
4. **Day 30-60**: We prepare a security release
5. **Day 60-90**: We coordinate disclosure with you
6. **Day 90**: Public disclosure (or earlier if mutually agreed)

### Credit and Recognition

- We will credit you in the security advisory (unless you prefer to remain anonymous)
- We maintain a [Security Hall of Fame](https://github.com/greysquirr3l/shadowforge/security/advisories) for responsible disclosures
- Critical vulnerabilities may be eligible for recognition

## Security Best Practices

When using Shadowforge, follow these security guidelines:

### Cryptographic Operations

- ✅ **Always verify signatures** before trusting extracted data
- ✅ **Use strong passphrases** for key derivation (20+ characters)
- ✅ **Rotate keys regularly** for long-term deployments
- ✅ **Securely store private keys** (use hardware security modules if possible)
- ❌ **Never reuse keys** across different contexts

### Steganographic Operations

- ✅ **Use high-quality cover media** to minimize detectability
- ✅ **Monitor capacity limits** to avoid overwriting critical data
- ✅ **Validate extracted data** for corruption or tampering
- ✅ **Use Reed-Solomon encoding** for error correction (30% redundancy recommended)
- ❌ **Don't embed in compressed media** (JPEG quality > 85 recommended)

### Archive Operations

- ✅ **Validate archive contents** before extraction
- ✅ **Extract to isolated directories** to prevent path traversal
- ✅ **Scan archives for malware** before processing
- ✅ **Limit archive sizes** to prevent resource exhaustion
- ❌ **Never auto-extract untrusted archives**

### Operational Security

- ✅ **Run with least privilege** (don't use root/admin unless required)
- ✅ **Use secure channels** for key exchange (never email keys)
- ✅ **Audit all operations** in production environments
- ✅ **Keep software updated** to the latest stable version
- ❌ **Never log sensitive data** (keys, payloads, passphrases)

## Security Audits

Shadowforge undergoes regular security reviews:

- **Code Reviews**: All cryptographic code is peer-reviewed
- **Dependency Scanning**: Automated vulnerability scanning with `go list -m all | nancy sleuth`
- **Static Analysis**: `go vet` and `golangci-lint` on every commit
- **Penetration Testing**: Periodic security assessments (planned)

## Known Security Considerations

### Post-Quantum Cryptography

- **Kyber-1024**: NIST standardized (FIPS 203) - secure against quantum attacks
- **Dilithium3**: NIST standardized (FIPS 204) - quantum-resistant signatures
- **Implementation**: Uses Cloudflare CIRCL library (audited)

### Steganographic Detection

- **Statistical Analysis**: LSB/DCT techniques detectable with RS/Chi-square analysis
- **Mitigation**: Use high-quality media, stay under capacity limits, apply ±1 LSB matching
- **Recommendation**: Combine multiple techniques with chaining for enhanced security

### Reed-Solomon Error Correction

- **Redundancy Trade-off**: Higher redundancy = better recovery but larger payload
- **Recommendation**: 30% redundancy (3 parity shards per 10 data shards) for standard use
- **Critical Data**: Use 50%+ redundancy for high-reliability requirements

## Contact

- **General Security**: <s0ma@protonmail.com> (if available)
- **GitHub Security Advisories**: [Create Private Advisory](https://github.com/greysquirr3l/shadowforge/security/advisories/new)
- **Project Maintainer**: [@greysquirr3l](https://github.com/greysquirr3l)

---

**Last Updated**: December 21, 2025
**Version**: 0.7.5

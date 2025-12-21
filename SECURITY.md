# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.7.x   | :white_check_mark: |
| < 0.7   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via:

1. **Email**: Send details to the repository owner
2. **Private Security Advisory**: Use GitHub's [private vulnerability reporting](https://github.com/greysquirr3l/shadowforge/security/advisories/new)

### What to Include

Please include the following information:

- **Type of vulnerability** (e.g., timing attack, memory leak, cryptographic flaw)
- **Full description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** and attack scenarios
- **Affected versions**
- **Suggested fix** (if known)

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: < 7 days
  - High: < 14 days
  - Medium: < 30 days
  - Low: Next release cycle

## Security Best Practices

When using Shadowforge:

1. **Always verify checksums** of downloaded binaries
2. **Use strong passwords** for encrypted archives
3. **Keep dependencies updated** (check releases regularly)
4. **Run in isolated environments** when processing untrusted media
5. **Review security audit reports** before deploying

## Known Security Considerations

### Timing Attacks

- Shadowforge uses constant-time comparison for cryptographic operations
- Some non-cryptographic code may have timing variations (documented in audit)

### File Processing

- All archive extraction includes zip-slip protection
- File size limits enforced to prevent resource exhaustion
- Sanitization applied to all media inputs

### Cryptographic Guarantees

- Post-quantum security via Kyber-1024 (KEM) and Dilithium3 (signatures)
- AES-GCM for symmetric encryption
- Argon2id for key derivation
- HMAC-SHA256 for message authentication

## Security Audit

Latest security audit: [SECURITY_AUDIT_REPORT.md](./SECURITY_AUDIT_REPORT.md)
Date: December 21, 2025
Version: 0.7.6

## Hall of Fame

Contributors who have responsibly disclosed vulnerabilities will be acknowledged here (with their permission).

---

**Last Updated**: December 21, 2025

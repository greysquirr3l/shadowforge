# Security Considerations

Important security information about Shadowforge and safe usage.

## Cryptographic Security

### Post-Quantum Cryptography

Shadowforge uses NIST-approved post-quantum cryptographic algorithms:

- **Kyber-1024** - Key encapsulation mechanism (encryption)
- **Dilithium3** - Digital signature algorithm

These algorithms are designed to resist attacks from both classical and future quantum computers.

**Important:** Quantum-resistant ≠ Unbreakable. No algorithm is provably secure.

### Implementation Security

Shadowforge implements:
- Constant-time operations to prevent timing attacks
- Secure memory handling with automatic key zeroing
- Proper random number generation (cryptographically secure)
- Input validation and bounds checking

## Data Protection Model

### Ephemeral Processing

Shadowforge follows an **ephemeral processing** model:

**What we do:**
- Process data in memory only during operations
- Automatically zero all sensitive data after use
- Provide immediate results without intermediate storage

**What we don't do:**
- Store media files on disk (except where you explicitly save them)
- Maintain logs of secret data
- Cache sensitive information between operations
- Create forensic artifacts on your system

### Your Responsibility

**You must secure:**
- Original secret data (before embedding)
- Carrier media (before and after embedding)
- Stego media (after embedding)
- Passwords and encryption keys
- Backup and archive files

## Threat Model

### What Shadowforge Protects Against

✓ **Casual observation** - Hidden data is not visible
✓ **Automated detection** - Statistical analysis tools may not detect hiding
✓ **Future quantum computers** - Post-quantum cryptography resistant
✓ **Common attacks** - Timing attacks, buffer overflows, etc.

### What Shadowforge Does NOT Protect Against

✗ **Forensic analysis** - Determined forensic examiners may detect hiding
✗ **Metadata attacks** - File metadata may reveal embedding patterns
✗ **Carrier analysis** - Statistical anomalies may indicate data presence
✗ **Targeted attacks** - Adversaries with resources may find hidden data
✗ **Compromised systems** - If your system is compromised, nothing helps

### Limitations

**Steganography is not encryption:**
- Hiding data ≠ Protecting data
- Hide AND encrypt for defense in depth
- Combine with other security measures

**No perfect security:**
- All systems have vulnerabilities
- Security is a process, not a destination
- Stay informed and update regularly

## Operational Security (OPSEC)

### Secure Usage Practices

1. **Minimize exposure:**
   - Don't share stego media with untrusted parties
   - Don't embed sensitive data carelessly
   - Limit who knows you're using steganography

2. **Separate concerns:**
   - Use different passwords for different purposes
   - Store carriers separately from secrets
   - Use different techniques for different data

3. **Defense in depth:**
   - Encrypt before embedding
   - Embed in encrypted containers
   - Distribute across multiple locations
   - Use redundant backups

4. **System hardening:**
   - Keep OS and tools updated
   - Use firewall and intrusion detection
   - Monitor file access patterns
   - Secure deletion of temporary files

### What NOT to Do

**❌ Don't:**
- Hardcode passwords in scripts
- Store passwords in plain text
- Share stego media on public platforms
- Reuse media across multiple secrets
- Ignore suspicious activity
- Neglect system security

**✓ Do:**
- Use strong, random passwords
- Store passwords securely (password manager)
- Share stego media through secure channels
- Verify recipient identity
- Monitor system activity
- Keep security practices current

## Known Limitations

### Detection by Advanced Adversaries

While Shadowforge provides good protection against casual detection:

- **Sophisticated analysis** may detect statistical anomalies
- **Forensic tools** may analyze carrier modifications
- **Pattern matching** could identify embedding patterns
- **Side channels** might leak information

**Mitigation:** Use lower capacity embedding, combine multiple techniques, maintain operational security.

### Media Format Constraints

Some formats have inherent limitations:

| Format | Constraint |
|--------|-----------|
| JPEG | Re-compression corrupts hidden data |
| PNG | Better for LSB due to uncompressed format |
| WAV | Visible to audio spectrum analysis |
| GIF | Limited by indexed color palette |
| Text | Small capacity, limited techniques |

**Solution:** Choose media carefully based on your threat model.

### Threshold Recovery

One-to-many distribution requires minimum shards:
- With 30% redundancy and 5 images, you need at least 4 images
- With 50% redundancy and 5 images, you need at least 3 images
- All shards must be from the same distribution

**Solution:** Keep manifest safe, distribute carriers carefully.

## Regulatory Considerations

### Jurisdiction-Specific Issues

Steganography and encryption are regulated differently by jurisdiction:

- Some countries restrict steganography
- Some require key escrow
- Some penalize "unauthorized" encryption
- Data protection laws (GDPR, CCPA, etc.) may apply

**Your responsibility:**
- Understand laws in your jurisdiction(s)
- Ensure legitimate use cases
- Document authorization if required
- Respect data privacy regulations

**We cannot provide legal advice.** Consult local legal counsel about regulations in your area.

## Responsible Disclosure

### Reporting Security Issues

If you discover a security vulnerability:

**Do NOT:**
- Post publicly to social media
- Open public GitHub issues
- Discuss in public forums
- Share with third parties

**Do:**
1. Email security@shadowforge.io with details
2. Give 90 days for us to patch
3. Don't exploit the vulnerability
4. Help us understand the impact

We will:
- Acknowledge receipt within 48 hours
- Keep you informed of progress
- Credit you in the security advisory (optional)
- Work toward responsible disclosure

## Additional Resources

### Learning More

- [NIST Post-Quantum Cryptography](https://csrc.nist.gov/projects/post-quantum-cryptography)
- [Steganography Research Papers](https://scholar.google.com/scholar?q=steganography)
- [OWASP Security Guidelines](https://owasp.org/)

### Security Tools

- **Password Manager:** 1Password, Bitwarden, KeePass
- **Encryption:** GPG, VeraCrypt, 7-Zip
- **Secure Deletion:** shred (Linux), secure-delete, BleachBit
- **File Integrity:** AIDE, Tripwire, Osquery

### Security Practices

- Keep systems updated
- Use strong authentication
- Monitor system activity
- Regular security audits
- Employee security training (for organizations)

## Questions & Support

**Security questions?** Open a discussion on [GitHub Discussions](https://github.com/greysquirr3l/shadowforge/discussions)

**Security vulnerability?** Email security@shadowforge.io

**General support?** See [FAQ](./faq.md) or [Troubleshooting](./troubleshooting.md)

---

**Last Updated:** December 2025
**Version:** 1.0.0

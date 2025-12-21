# Shadowforge CLI Documentation

![Shadowforge Logo](../../assets/shadowforge-logo.png)

> **"Forge secrets in the shadows, shield them from quantum eyes"**

Welcome to Shadowforge, a production-grade quantum-resistant steganography tool. This documentation covers installation, usage, and best practices for the `shadowforge` command-line interface.

## ⚡ Quick Start

### Installation

**macOS with Homebrew** (coming soon):
```bash
brew install shadowforge
shadowforge --version
```

**From Source**:
```bash
go install github.com/greysquirr3l/shadowforge/cmd/cli@latest
shadowforge --version
```

### Basic Usage

Embed a secret message in an image:
```bash
shadowforge embed \
  --input secret.txt \
  --cover image.png \
  --output stego.png
```

Extract the secret:
```bash
shadowforge extract --input stego.png --output recovered.txt
```

## 📖 Documentation Structure

- **[Installation](./installation.md)** - Installation methods and system requirements
- **[Getting Started](./getting-started.md)** - Basic workflows and first-time setup
- **[Quick Reference](./quick-reference.md)** - Fast command lookup and cheatsheet
- **[Commands Reference](./commands/README.md)** - Complete command reference
  - [Embed Commands](./commands/embed.md) - Data embedding operations
  - [Extract Commands](./commands/extract.md) - Data recovery operations
  - [Analyze Commands](./commands/analyze.md) - Capacity and security analysis
  - [Archive Commands](./commands/archive.md) - Archive creation and management
  - [Watermark Commands](./commands/watermark.md) - Forensic watermarking
  - [Selection Commands](./commands/selection.md) - Intelligent cover media selection
  - [Chain Commands](./commands/chain.md) - Technique chaining
  - [Utility Commands](./commands/utility.md) - Helper utilities
- **[Best Practices](./best-practices.md)** - Security recommendations and workflows
- **[Advanced Guide](./guides/advanced.md)** - Professional-grade techniques and optimization
- **[Security Hardening](./guides/security.md)** - Production security hardening
- **[Glossary](./glossary.md)** - Terminology and concept definitions
- **[Use Cases](./use-cases.md)** - Real-world application scenarios
- **[Troubleshooting](./troubleshooting.md)** - Common issues and solutions
- **[FAQ](./faq.md)** - Frequently asked questions

## 🎯 Key Capabilities

### Data Embedding
Embed secrets in multiple carrier types:
- **Images**: PNG, BMP, JPEG, GIF
- **Audio**: WAV files
- **Text**: Plain text files
- **Archives**: ZIP, TAR, TAR.GZ

### Distribution Patterns
Choose how to distribute your secret:
- **One-to-One** (1:1) - Single carrier, single secret
- **One-to-Many** (1:N) - Distribute across multiple carriers
- **Many-to-One** (N:1) - Embed multiple secrets in one carrier
- **Many-to-Many** (N:M) - Complex distribution matrix

### Security Features
- **Post-Quantum Encryption** - Kyber-1024 key encapsulation
- **Digital Signatures** - Dilithium3 authenticity verification
- **Error Correction** - Reed-Solomon redundancy for recovery
- **Multiple Techniques** - LSB, DCT, Phase, Echo, Palette, Zero-Width

## 🔐 Security Model

Shadowforge is designed with security-first principles:

- **No persistent storage** of media by the tool itself
- **Ephemeral processing** - data exists in memory only during operations
- **Secure memory handling** - automatic key and buffer zeroing
- **Quantum-resistant cryptography** - protection against future threats

For detailed security architecture, see [Security Considerations](./security.md).

## 💡 Common Workflows

### Scenario 1: Simple Secret Sharing
Share a document secretly with a colleague:
```bash
# They send you a carrier image
# You embed your secret
shadowforge embed --input confidential.pdf --cover photo.png --output photo-with-secret.png

# Send back the stego image
# They extract:
shadowforge extract --input photo-with-secret.png --output recovered.pdf
```

### Scenario 2: Resilient Distribution
Share critical data with redundancy across multiple files:
```bash
# Distribute across 5 images, recover from any 3
shadowforge embed \
  --input critical-data.zip \
  --cover-files image1.png,image2.png,image3.png,image4.png,image5.png \
  --pattern 1:N \
  --threshold 3 \
  --output-dir distributed/

# Later, recover from any 3 images:
shadowforge extract \
  --pattern 1:N \
  --input-files image1.png,image3.png,image5.png \
  --output recovered-data.zip
```

### Scenario 3: Archive Workflows
Bundle and distribute stego files:
```bash
# Create archive of stego files
shadowforge archive create \
  --input-dir ./stego-output \
  --format tar.gz \
  --output stego-files.tar.gz

# Later, extract and recover secrets
tar xzf stego-files.tar.gz -C ./temp/
shadowforge extract-batch --input-dir ./temp --output-dir ./recovered
```

## 🚀 Next Steps

1. **First time?** Start with [Getting Started](./getting-started.md)
2. **Need help with a specific command?** Check [Commands Reference](./commands/README.md)
3. **Security concerns?** Review [Best Practices](./best-practices.md)
4. **Having trouble?** See [Troubleshooting](./troubleshooting.md)

## 📋 System Requirements

- **Operating System**: macOS 11+, Linux (glibc 2.29+), Windows 10+
- **Processor**: x86-64 or ARM64
- **Memory**: Minimum 256MB RAM (2GB+ recommended for large files)
- **Disk Space**: Minimal (depends on your files)

## ⚠️ Important Notes

> **Security is a process, not a product.** While Shadowforge implements state-of-the-art cryptography, successful use requires careful operational security:
>
> - Guard your secrets and cover media carefully
> - Use secure channels to share stego files
> - Verify recipient identity before sharing
> - Monitor for unauthorized access attempts

## 📞 Support & Feedback

- **Issues**: Report bugs on [GitHub Issues](https://github.com/greysquirr3l/shadowforge/issues)
- **Security**: Report vulnerabilities to security@shadowforge.io
- **Documentation**: Suggest improvements in [GitHub Discussions](https://github.com/greysquirr3l/shadowforge/discussions)

## 📜 License

Shadowforge is licensed under the Apache License 2.0. See [LICENSE](../../../LICENSE) for details.

---

**Version**: 1.0.0
**Last Updated**: December 2025
**Status**: Production Ready

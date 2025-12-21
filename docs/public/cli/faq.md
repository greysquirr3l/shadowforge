# Frequently Asked Questions

## Installation & Setup

**Q: Does Shadowforge work on Windows?**
A: Yes, Shadowforge works on Windows 10 and later. Pre-built binaries are available for download.

**Q: Can I install Shadowforge without Administrator privileges?**
A: Yes, you can install it in your user directory. See [Installation](./installation.md) for details.

**Q: How much disk space does Shadowforge need?**
A: The binary is ~8MB. Additional space depends on your media files.

## Basic Usage

**Q: What's the difference between embedding and extracting?**
A: Embedding hides a secret in a carrier image. Extracting recovers the secret from the stego image.

**Q: What media formats does Shadowforge support?**
A: PNG, JPEG, BMP, GIF (images), WAV (audio), TXT (text), ZIP/TAR (archives). Use `shadowforge formats` to see all.

**Q: Can I use any image as a carrier?**
A: Ideally, use high-quality original images. Avoid already-compressed or edited media.

**Q: What's the maximum size payload I can embed?**
A: It depends on carrier size and technique. Use `shadowforge analyze capacity` to check.

## Distribution Patterns

**Q: What distribution pattern should I use?**
A:
- **1:1** for simple sharing
- **1:N** for resilient distribution
- **N:1** for bundling
- **N:M** for complex scenarios

**Q: What's "threshold" in distributed embedding?**
A: The minimum number of carriers needed to recover the secret. With 5 carriers and threshold=3, you need any 3 to recover.

**Q: Can I recover from fewer carriers than the threshold?**
A: No, you need at least the threshold number of carriers.

**Q: What's the point of redundancy?**
A: Redundancy allows recovery from carrier loss. Higher redundancy = need fewer carriers, but lower capacity.

## Security Questions

**Q: Is my data safe with Shadowforge?**
A: Shadowforge provides good protection against casual detection. For sensitive data, also encrypt before embedding.

**Q: Can Shadowforge be detected?**
A: Sophisticated analysis might detect statistical anomalies, but standard tools should not.

**Q: Should I password-protect my stego files?**
A: Yes, for additional security. Use `--password` flag.

**Q: How should I store stego files?**
A: Treat them like any sensitive file - use encrypted storage, restrict access, secure deletion when no longer needed.

**Q: What if my system is compromised?**
A: Steganography alone can't protect you if your system is compromised. Use system security tools and practices.

## Troubleshooting

**Q: I get "insufficient capacity" error. What do I do?**
A:
1. Use a larger carrier image
2. Use a smaller payload
3. Choose a different technique with higher capacity
4. Use distributed embedding across multiple carriers

**Q: The recovered file doesn't match the original. Why?**
A:
1. Check you're using the correct password
2. Verify the carrier wasn't modified after embedding
3. Try specifying the technique explicitly
4. Check carrier integrity with `shadowforge validate`

**Q: Can I re-compress a JPEG after embedding?**
A: No, re-compression will likely corrupt the hidden data. Use lossless formats like PNG for safety.

**Q: How do I securely delete temporary files?**
A: Use `shred` (Linux/macOS) or `cipher /w:C:` (Windows). Or use specialized tools like BleachBit.

**Q: Do I need the manifest file to extract?**
A: No, but it helps identify which shard is where and determines the threshold. Good practice to keep it.

## Advanced Usage

**Q: Can I combine multiple steganography techniques?**
A: Yes, use `shadowforge chain` for technique chaining. Combine LSB with DCT, etc.

**Q: What's forensic watermarking?**
A: Embedding recipient identification across multiple images. Use `shadowforge watermark` for this.

**Q: Can I embed multiple files in one image?**
A: Yes, use `shadowforge embed-batch` to bundle multiple files.

**Q: Can I use Shadowforge in scripts?**
A: Yes. Use `--output-format json` for script-friendly output.

## Performance

**Q: How long does embedding take?**
A: Typically seconds to minutes depending on file sizes. Larger images take longer.

**Q: How long does extraction take?**
A: Usually faster than embedding. Similar times for distributed extraction (parallel processing).

**Q: Can I speed up distributed operations?**
A: Yes, use `--parallel` flag and `--workers N` to control parallelism.

## Legal & Compliance

**Q: Is steganography legal?**
A: It depends on your jurisdiction. Some countries restrict or regulate it. Check local laws.

**Q: Can I use Shadowforge for business?**
A: Yes, but ensure you comply with relevant regulations (encryption laws, data protection, etc.)

**Q: Do I need to disclose I'm using steganography?**
A: Check your jurisdiction's laws and your organization's policies.

**Q: What about GDPR/CCPA compliance?**
A: Shadowforge doesn't store personal data, but you're responsible for complying with data protection laws when using it.

## Getting More Help

**Can't find the answer here?**

1. Check [Troubleshooting](./troubleshooting.md) for common issues
2. Review [Getting Started](./getting-started.md) for basic workflows
3. See [Commands Reference](./commands/README.md) for detailed command info
4. Ask on [GitHub Discussions](https://github.com/greysquirr3l/shadowforge/discussions)
5. Report bugs on [GitHub Issues](https://github.com/greysquirr3l/shadowforge/issues)

---

**Still have questions?** See the [full documentation](./README.md) or reach out to the community!

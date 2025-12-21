# Getting Started

Welcome to Shadowforge! This guide will walk you through your first steganography operation.

## Your First Embedding

Let's embed a secret message in an image.

### Step 1: Prepare Your Files

Gather the files you'll need:

```bash
# Create a secret message
echo "This is a secret message!" > secret.txt

# Find or create a cover image (PNG, JPEG, or BMP)
# For testing, you can use any image, e.g., a photo from your camera
cp ~/Pictures/vacation.jpg cover.jpg
```

### Step 2: Check Media Capacity

Before embedding, verify the cover image can hold your secret:

```bash
shadowforge analyze capacity --input cover.jpg --payload-size 25
```

**Output** might look like:
```
✓ Cover media: cover.jpg (JPEG, 3024x4032)
  Available capacity: 847 KB (safe: 508 KB)
  Payload size: 25 bytes
  ✓ Sufficient capacity - Ready to embed

Recommended technique: DCT
  - Capacity: 847 KB
  - Speed: Fast (68ms embed, 9ms extract)
  - Security: Excellent for JPEG
```

### Step 3: Embed Your Secret

Now embed the secret in the cover image:

```bash
shadowforge embed \
  --input secret.txt \
  --cover cover.jpg \
  --output stego.jpg \
  --technique dct
```

**Output** might look like:
```
✓ Embedding payload in cover media...
  Technique: DCT JPEG
  Payload size: 25 bytes
  Cover media: cover.jpg (847 KB capacity)

✓ Embedding complete
  Output: stego.jpg (3.2 MB)
  Capacity used: 0.003%

The stego image looks identical to the original.
You can safely share it with others.
```

### Step 4: Share the Stego Image

The `stego.jpg` file now contains your secret message, invisibly hidden:

```bash
# You can share it normally - it looks like a regular image!
cp stego.jpg ~/Desktop/
# Send via email, upload to cloud storage, post on social media, etc.
```

### Step 5: Extract the Secret

When the recipient has the stego image, they can extract the secret:

```bash
shadowforge extract \
  --input stego.jpg \
  --output recovered.txt
```

**Output** might look like:
```
✓ Extracting payload from stego media...

✓ Extraction complete
  Output: recovered.txt
  Payload size: 25 bytes
  Integrity: ✓ Verified
```

Verify the recovered message:
```bash
cat recovered.txt
# Output: This is a secret message!
```

## Understanding Distribution Patterns

Shadowforge supports multiple ways to distribute your secret:

### One-to-One (1:1) - Simple Embedding
**What**: One secret embedded in one carrier
**Use case**: Sharing a single document securely
**Recovery**: Need 1 carrier image

```bash
# Simple: one secret, one carrier
shadowforge embed --input secret.pdf --cover image.png --output stego.png
```

### One-to-Many (1:N) - Resilient Distribution
**What**: One secret distributed across multiple carriers
**Use case**: Critical data with redundancy
**Recovery**: Need K of N carriers (configurable threshold)

```bash
# Distribute across 5 images, recoverable from any 3
shadowforge embed \
  --input critical-data.zip \
  --cover image1.png,image2.png,image3.png,image4.png,image5.png \
  --pattern 1:N \
  --threshold 3
```

**Why use this?** If some carrier images are lost or corrupted, you can still recover your secret from the remaining images.

### Many-to-One (N:1) - Batch Embedding
**What**: Multiple secrets embedded in one carrier
**Use case**: Bundling related documents
**Recovery**: Need the single carrier image

```bash
# Embed 3 documents in one image
shadowforge embed \
  --input document1.pdf,document2.pdf,document3.pdf \
  --cover image.png \
  --pattern N:1 \
  --output stego.png
```

### Many-to-Many (N:M) - Complex Distribution
**What**: Multiple secrets distributed across multiple carriers
**Use case**: Enterprise data distribution
**Recovery**: Configurable thresholds per secret

```bash
# Distribute 3 documents across 6 images
shadowforge embed \
  --input doc1.pdf,doc2.pdf,doc3.pdf \
  --cover image1.png,image2.png,image3.png,image4.png,image5.png,image6.png \
  --pattern N:M \
  --output-dir distributed/
```

## Working with Different Media Types

Shadowforge handles multiple media formats:

### Images

**PNG / BMP** (Lossless formats):
```bash
# Good candidates for steganography
shadowforge embed --input secret.txt --cover photo.png --output stego.png
```

**JPEG** (Lossy format):
```bash
# Works well, but avoid re-compression
shadowforge embed --input secret.txt --cover photo.jpg --output stego.jpg
# Note: Some JPEG editing will corrupt the hidden data
```

**GIF** (Indexed color):
```bash
# Works with palette manipulation
shadowforge embed --input secret.txt --cover animation.gif --output stego.gif
```

### Audio

**WAV** (Uncompressed):
```bash
# Embed in audio samples
shadowforge embed --input secret.txt --cover audio.wav --output stego.wav
# Hiding data in audio may be audible to trained ears
```

### Text

**Plain Text**:
```bash
# Hide data using invisible Unicode characters
shadowforge embed --input secret.txt --cover document.txt --output stego.txt
# Resulting text looks normal when displayed
```

### Archives

**ZIP, TAR, TAR.GZ**:
```bash
# Bundle stego files into an archive
shadowforge archive create \
  --input-dir ./stego-output \
  --format tar.gz \
  --output stego-bundle.tar.gz
```

## Analyzing Media

Check the capacity and suitability of media before embedding:

```bash
# Analyze a single file
shadowforge analyze capacity --input image.png

# Analyze multiple files
shadowforge analyze capacity \
  --input image1.png,image2.jpg,audio.wav \
  --payload-size 1024

# Scan a directory for suitable media
shadowforge scan ~/Pictures --min-capacity 100KB
```

## Next Steps

Now that you've learned the basics:

1. **Read about security**: [Best Practices](./best-practices.md)
2. **Explore more commands**: [Commands Reference](./commands/README.md)
3. **Use advanced features**: [Chain Commands](./commands/chain.md), [Watermarking](./commands/watermark.md)
4. **Troubleshoot issues**: [Troubleshooting](./troubleshooting.md)

## Key Concepts

### Stego Image
The result of embedding secret data in a carrier image. Appears identical to the original to the human eye.

### Capacity
The maximum amount of data that can be securely hidden in a carrier. Varies by media type and technique.

### Technique
The method used to hide data (e.g., LSB, DCT, Phase). Different techniques have different properties.

### Threshold
In resilient distribution (1:N), the minimum number of carriers needed to recover the secret.

### Payload
The secret data being hidden (documents, messages, files, etc.).

---

**Ready for more advanced features?** Check out [Technique Chaining](./commands/chain.md) and [Forensic Watermarking](./commands/watermark.md)!

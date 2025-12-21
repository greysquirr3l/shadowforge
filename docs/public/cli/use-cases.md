# Use Cases

> **Real-World Applications of Shadowforge**

This document explores practical applications and use case scenarios for Shadowforge.

## Secure Document Distribution

### Scenario: Classified Document to Multiple Recipients

**Challenge**: Distribute a sensitive document to 10 stakeholders without exposing its existence.

**Solution**:

```bash
# 1. Create password-protected archive
shadowforge archive create \
  --output classified.zip \
  --password "secure-passphrase" \
  --encryption aes-256 \
  --files document.pdf appendix.pdf evidence.zip

# 2. Analyze distribution requirements
PAYLOAD_SIZE=$(stat -f%z classified.zip)
shadowforge analyze capacity --technique lsb --cover sample.png

# 3. Select cover media
shadowforge scan ./photos --recursive
shadowforge select ./photos --payload-size $PAYLOAD_SIZE

# 4. Execute distribution
shadowforge embed-distributed \
  --payload classified.zip \
  --data-shards 8 \
  --parity-shards 5 \
  --covers photo1.jpg photo2.jpg ... photo13.jpg

# 5. Distribute carriers to recipients
# Each recipient receives 2-3 carriers from the 13 total
# Any 8 carriers enables recovery
```

**Benefits**:
- Classified content appears as innocuous photos
- Loss of up to 5 carriers doesn't compromise security
- No single person has complete information
- Plausible deniability (photos can be explained naturally)

---

## Resilient Backup System

### Scenario: Critical Data Backup Across Locations

**Challenge**: Create backups that survive location loss with undetectable storage.

**Solution**:

```bash
# 1. Create and encrypt backup
tar czf critical-backup.tar.gz /database /configs
openssl enc -aes-256-cbc -in critical-backup.tar.gz \
  -out critical-backup.bin

# 2. Distribute across geographic locations
# Location A: Office (8 carriers)
# Location B: Remote server (5 carriers)
# Location C: Off-site vault (8 carriers)

shadowforge embed-distributed \
  --payload critical-backup.bin \
  --data-shards 12 \
  --parity-shards 9 \
  --covers \
    /location-a/photo*.jpg \
    /location-b/archive*.png \
    /location-c/images/*.jpg

# 3. Each location stores carries as innocent media
# Location A: 8 family photos
# Location B: 5 website screenshots
# Location C: 8 historical archives
```

**Benefits**:
- Backup hidden in plain sight across locations
- Survives location loss (can recover with any 12 of 21 carriers)
- Undetectable storage method
- No traditional backup infrastructure needed

---

## Covert Communication Channel

### Scenario: Establish Hidden Communication Between Agents

**Challenge**: Communicate securely without evidence of contact.

**Solution**:

```bash
# Agent A (Sender) prepares message
shadowforge chain create secure-channel \
  --mode sequential \
  --payload "rendezvous location and time" \
  --techniques phase,echo,lsb \
  --covers audio1.wav audio2.wav image.png

shadowforge chain execute secure-channel

# Output: stego1.wav, stego2.wav, stego.png

# Transmit through innocuous channels:
# - Email attachments as personal audio
# - Message board as "random" image
# - Cloud storage as backups

# Agent B (Receiver) extracts message
shadowforge chain extract secure-channel \
  --stego-files received1.wav received2.wav received.png \
  --output recovered-message.txt
```

**Benefits**:
- Message hidden in plain audio/images
- Multiple transmission routes possible
- No metadata indicating communication
- Plausible explanation for shared media

---

## Privacy-Preserving Data Leaks

### Scenario: Securely Share Whistleblowing Evidence

**Challenge**: Safely transmit sensitive evidence to journalists/authorities.

**Solution**:

```bash
# Whistleblower prepares evidence
tar czf evidence.tar.gz documents/ screenshots/ logs/

# Add extra layer of encryption
openssl enc -aes-256-cbc -in evidence.tar.gz \
  -out evidence.bin \
  -pass pass:secure-passphrase

# Distribute across multiple carriers
# Use high redundancy for reliability
shadowforge embed-distributed \
  --payload evidence.bin \
  --data-shards 6 \
  --parity-shards 8 \
  --covers photo1.jpg photo2.jpg photo3.jpg photo4.jpg \
    audio1.wav audio2.wav video.mp4 landscape.png \
    screenshot.jpg document_scan.png screenshot2.jpg

# Share carriers through multiple channels:
# - Send some as email attachments
# - Upload to cloud storage
# - Share via messaging apps
# - Mail USB drive via post

# Recipient can reconstruct with any 6 carriers
# Complete loss of 8 carriers still allows recovery
```

**Benefits**:
- Evidence hidden in innocuous media
- Survives carrier loss (high parity)
- Multiple transmission methods possible
- Receiver anonymity protected

---

## Academic Research Distribution

### Scenario: Share Pre-Publication Research Securely

**Challenge**: Share research papers with collaborators before publication without leaking to competitors.

**Solution**:

```bash
# Researcher creates distribution package
shadowforge archive create \
  --output research-papers.zip \
  --password "research-passphrase" \
  --files paper1.pdf paper2.pdf data.xlsx presentation.pptx

# Distribute to 15 collaborators
# Each gets unique subset preventing correlation

shadowforge embed-distributed \
  --payload research-papers.zip \
  --data-shards 10 \
  --parity-shards 5 \
  --covers lab-photo1.png ... lab-photo15.png

# Each collaborator receives:
# Collaborator 1: carriers 1,2,3,5,7,10
# Collaborator 2: carriers 1,2,4,6,8,11
# Collaborator 3: carriers 2,3,4,6,9,12
# ... etc
```

**Benefits**:
- Research hidden as lab photos
- Each collaborator has unique subset
- Leaks traceable to specific person
- Distribution prevents unauthorized sharing

---

## Secure Key Distribution

### Scenario: Distribute Encryption Keys to Field Agents

**Challenge**: Safely distribute sensitive cryptographic keys.

**Solution**:

```bash
# Generate keys for each agent
shadowforge keygen --key-id agent-alice
shadowforge keygen --key-id agent-bob
shadowforge keygen --key-id agent-charlie

# Create key distribution package
tar czf agent-keys.tar.gz agent-*.pem

# Distribute different agents' keys to each carrier
# This prevents any single carrier loss from exposing all keys

shadowforge embed-distributed \
  --payload agent-keys.tar.gz \
  --data-shards 9 \
  --parity-shards 6 \
  --covers passport-photo1.jpg ... passport-photo15.jpg

# Distribution to field agents
# Agent Alice: gets 4 photos containing key distribution info
# Agent Bob: gets different 4 photos
# Agent Charlie: gets different 4 photos
# Central server: stores copies of all photos
```

**Benefits**:
- Keys hidden in mundane photos
- Can be transmitted through regular channels
- No metadata indicating cryptographic content
- Provides deniability

---

## Metadata Hiding

### Scenario: Embed Metadata in Published Content

**Challenge**: Include hidden metadata (watermarks, signatures) in published images.

**Solution**:

```bash
# Publishing house embeds publisher ID and publication date
shadowforge watermark embed \
  --stego-file published-image.png \
  --metadata "publisher:springerverlag,date:2025-12-21,edition:5"

# Later verification: watermark extraction
shadowforge watermark extract \
  --stego-file received-image.png \
  --output metadata.txt

# Result: metadata.txt contains:
# publisher:springerverlag
# date:2025-12-21
# edition:5
```

**Benefits**:
- Copyright information invisible to users
- Tampering detection (metadata changes = image altered)
- No visible watermarks that distract from content
- Automated detection of unauthorized modifications

---

## Audit Trail Embedding

### Scenario: Embed Audit Information in System Snapshots

**Challenge**: Prove system state at specific time without separate audit logs.

**Solution**:

```bash
# System audit embeds metadata in system screenshot
AUDIT_DATA="timestamp:2025-12-21T14:30:00Z,user:admin,\
hash:a1b2c3d4e5f6,disk_usage:87%,cpu_temp:65C"

echo $AUDIT_DATA > audit-metadata.txt

shadowforge embed \
  --payload audit-metadata.txt \
  --cover system-screenshot.png \
  --technique zero-width

# Screenshot appears normal to viewers
# But contains embedded audit trail

# Later verification
shadowforge extract \
  --stego-file archived-screenshot.png \
  --output recovered-audit.txt
```

**Benefits**:
- Audit trail invisible to system users
- Proves system state at specific time
- Tamper-evident (screenshot modifications affect metadata)
- Single file contains evidence + documentation

---

## Artist Copyright Protection

### Scenario: Protect Digital Art with Hidden Signature

**Challenge**: Prove ownership of digital artwork while allowing distribution.

**Solution**:

```bash
# Artist embeds signature in artwork
SIGNATURE="artist:JohnDoe,created:2025-01-15,\
license:CC-BY-SA,copyright:2025"

shadowforge watermark embed \
  --stego-file digital-artwork.png \
  --metadata "$SIGNATURE"

# Share artwork on social media
# Anyone can enjoy the art
# But artist's signature is embedded

# If stolen or misattributed, verify ownership
shadowforge watermark extract \
  --stego-file suspicious-image.png

# Signature confirms artist and creation date
```

**Benefits**:
- Invisible copyright notice
- Proves ownership if disputes arise
- Doesn't interfere with artwork enjoyment
- Survives format conversions (if quality preserved)

---

## Sensitive Survey Distribution

### Scenario: Distribute Confidential Survey to Participants

**Challenge**: Safely distribute anonymous survey to participants.

**Solution**:

```bash
# Create survey in document
cat > survey.txt << 'EOF'
[CONFIDENTIAL SURVEY - ID: 12345]
1. Question about workplace safety...
2. Question about management practices...
3. Question about compensation...
EOF

# Create multiple versions (different IDs for tracking)
for i in {1..50}; do
  sed "s/ID: 12345/ID: RESP$i/" survey.txt > survey-$i.txt
done

# Distribute hidden in images
shadowforge embed-batch \
  --payloads survey-*.txt \
  --cover generic-landscape.png

# Each survey hidden in same image, different embedding
# Recipient extracts only their version
```

**Benefits**:
- Survey hidden from casual observation
- Responses appear as innocuous images
- Recipient anonymity preserved
- No metadata indicating survey content

---

## Crisis Communication

### Scenario: Secure Communication During Emergencies

**Challenge**: Establish communication when normal channels are compromised.

**Solution**:

```bash
# Create emergency message
echo "Evacuate area immediately. Fallback location: $(date +%s | md5sum | cut -c1-6)" > message.txt

# Embed in multiple innocuous carriers
shadowforge chain create emergency \
  --mode split \
  --payload message.txt \
  --techniques lsb,zero-width \
  --covers photo.jpg document.txt

# Share through multiple public channels:
# - Post as social media background
# - Include as email signature
# - Upload to backup storage
# - Message to broadcast list

# Recipients know to extract message from these sources
```

**Benefits**:
- Message hidden in plain sight
- Multiple distribution methods possible
- No indication of emergency communication
- Deniability if questioned

---

## Best Practices by Use Case

### For Critical Operations

1. **Use distributed embedding** (1:N pattern)
2. **High redundancy** (parity shards ≥ 40% of total)
3. **Multiple transmission methods**
4. **Store manifest separately**
5. **Test extraction before production use**

### For Privacy Protection

1. **Double encryption** (Shadowforge + external)
2. **Zero-width or phase encoding** (maximum stealth)
3. **Large carriers** (preserve statistics)
4. **Capacity under 70%** of available
5. **Archive format** for multiple files

### For Metadata/Watermarking

1. **Use watermark commands**
2. **Zero-width encoding** for text metadata
3. **LSB for image metadata**
4. **Small metadata sizes**
5. **Verify extraction regularly**

---

*Use Cases Documentation for Shadowforge v1.0+*
*Last Updated: December 2025*

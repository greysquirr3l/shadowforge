# Ephemeral Storage Architecture - Shadowforge Security Model

> **"A steganography tool that leaves no trace"**

**Document Version:** 1.0.0
**Last Updated:** December 10, 2025
**Status:** Approved for Implementation
**Security Level:** CRITICAL

---

## Table of Contents

- [Overview](#overview)
- [Security Principles](#security-principles)
- [Architecture by Deployment Mode](#architecture-by-deployment-mode)
- [Encrypted Temp Storage Design](#encrypted-temp-storage-design)
- [Memory Security](#memory-security)
- [Threat Model](#threat-model)
- [Implementation Guidelines](#implementation-guidelines)

---

## Overview

Shadowforge's security model is based on **ephemeral processing** - data exists in memory only during active operations, with cryptographic protection and immediate secure deletion.

### Core Principle

```
┌──────────────────────────────────────────────────────────┐
│ "If it's not in RAM actively being processed,           │
│  it shouldn't exist at all."                             │
└──────────────────────────────────────────────────────────┘
```

### Anti-Pattern

```
❌ FORBIDDEN: Traditional Persistent Storage

User uploads file → Store in database → Process → Return result
                         ↓
                    Forensic evidence
                    Survives seizure
                    Subpoena target
                    Data breach risk
```

### Correct Pattern

```
✅ CORRECT: Ephemeral Processing

User uploads file → Process in RAM → Secure zero → Return result
                         ↓
                    No persistence
                    No forensic trace
                    No seizure risk
                    Self-destructing
```

---

## Security Principles

### 1. **No Persistent Storage**

**Rule:** Shadowforge NEVER stores media data to disk, database, or any persistent medium under its own control.

**Rationale:**
- Steganography tools are targets for surveillance
- Persistent storage creates forensic evidence
- Users manage their own data storage/deletion
- Tool should be stateless and deniable

**Exceptions:**
- User-controlled key storage (encrypted, optional)
- Temporary encrypted cache (API mode only, session-scoped, auto-shredded)
- Audit logs (optional, metadata only, encrypted, user-controlled)

### 2. **Minimal In-Memory Lifetime**

**Rule:** Data exists in RAM only during active processing, measured in seconds/minutes, not hours/days.

**Targets:**
- CLI: Process duration (typically <10 seconds)
- API: Request duration + configurable TTL (max 5 minutes)

### 3. **Cryptographic Protection**

**Rule:** Any temporary storage MUST be encrypted with ephemeral keys.

**Requirements:**
- AES-256-GCM for data encryption
- Keys derived from user session (never stored)
- Keys destroyed immediately after use
- No key reuse across sessions

### 4. **Secure Memory Zeroing**

**Rule:** All sensitive data MUST be overwritten with zeros before deallocation.

**Scope:**
- Cover media buffers
- Secret payload buffers
- Stego output buffers
- Encryption keys
- Temporary working memory

### 5. **Audit Minimalism**

**Rule:** Log operations, not data. Never log file content, paths, or identifying metadata.

**Permitted:**
- Timestamp: `2025-12-10T15:30:00Z`
- Operation: `embed_distributed`
- Status: `success`
- User ID: `session-abc123` (API only, session-scoped)

**Forbidden:**
- File names: ~~`secret_document.pdf`~~
- File sizes: ~~`2.5MB`~~
- File hashes: ~~`sha256:abc123...`~~
- User IP addresses: ~~`192.168.1.1`~~
- File paths: ~~`/home/user/secrets/...`~~

---

## Architecture by Deployment Mode

### CLI Mode (Phase 5) - Zero Persistence

**User Workflow:**
```bash
shadowforge embed --input secret.txt --cover image.png --output stego.png
```

**Internal Flow:**
```
┌─────────────────────────────────────────────────────────┐
│ 1. Read files from disk (user's filesystem)            │
│    - cover.png  → RAM buffer (defer secure zero)       │
│    - secret.txt → RAM buffer (defer secure zero)       │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 2. Process in RAM                                        │
│    - Encrypt with Kyber-1024                            │
│    - Reed-Solomon encode                                │
│    - Embed in cover media                               │
│    - All operations in memory buffers                   │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Write output to disk (user's filesystem)             │
│    - stego.png → user-specified path                    │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 4. Secure cleanup (defer/atexit handlers)               │
│    - Overwrite cover buffer with zeros                  │
│    - Overwrite secret buffer with zeros                 │
│    - Overwrite encryption keys with zeros               │
│    - runtime.KeepAlive() prevents compiler optimization │
└─────────────────────────────────────────────────────────┘
                         ↓
                    EXIT (clean)
```

**No Repository Used:**
```go
// CLI does NOT use Repository pattern
// Direct file I/O only

func (h *EmbedCommandHandler) Handle(ctx context.Context, cmd EmbedCommand) error {
    // Read from user's disk
    coverData, err := os.ReadFile(cmd.CoverPath)
    defer secureZero(coverData)

    secretData, err := os.ReadFile(cmd.SecretPath)
    defer secureZero(secretData)

    // Process in RAM (no repository)
    stego, err := h.stegoService.Embed(ctx, secretData, coverData)
    defer secureZero(stego)

    // Write to user's disk
    err = os.WriteFile(cmd.OutputPath, stego, 0600)

    return err
    // Deferred cleanup runs here
}
```

**Security Properties:**
- ✅ No persistence by Shadowforge
- ✅ User controls all file storage
- ✅ Process memory cleared on exit
- ✅ No traces in system logs (beyond process execution)
- ✅ Survives: Nothing (by design)

---

### API Server Mode (Phase 6) - Encrypted Temp Storage

**User Workflow:**
```bash
curl -X POST https://api.shadowforge.io/v1/embed \
  -F "cover=@image.png" \
  -F "secret=@document.pdf" \
  -H "Authorization: Bearer session-token"
```

**Internal Flow:**
```
┌─────────────────────────────────────────────────────────┐
│ 1. Request arrives (multipart/form-data)                │
│    - Parse upload to RAM                                │
│    - Extract session token                              │
│    - Derive ephemeral encryption key                    │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 2. Optional: Store in EncryptedTempRepository           │
│    - Encrypt with session-derived key                   │
│    - Store in memory/tmpfs                              │
│    - Set TTL: 5 minutes max                             │
│    - Purpose: Handle async processing, long operations  │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Process (same as CLI)                                │
│    - Encrypt, encode, embed                             │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 4. Return result in HTTP response                       │
│    - Stream output to response body                     │
│    - Or return pre-signed S3 URL (user's bucket)        │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 5. Immediate cleanup (CRITICAL)                         │
│    - Delete from EncryptedTempRepository                │
│    - Overwrite encryption key                           │
│    - Secure zero all buffers                            │
│    - Trigger GC                                          │
└─────────────────────────────────────────────────────────┘
```

**Security Properties:**
- ✅ Data encrypted at rest (in temp storage)
- ✅ Encryption key derived from session (not stored)
- ✅ Auto-delete after response sent
- ✅ TTL enforces maximum lifetime (5 min)
- ✅ Survives: Server restart clears everything (tmpfs)

---

## Encrypted Temp Storage Design

### EncryptedTempRepository Interface

```go
// EncryptedTempRepository provides short-lived encrypted storage for API server.
//
// Security guarantees:
// - All data encrypted with AES-256-GCM
// - Encryption keys derived from user session (never stored)
// - Automatic expiration after TTL
// - Secure memory zeroing on deletion
// - Optional tmpfs backing (RAM-only)
//
// Use cases:
// - Async operations (long-running embeds)
// - WebSocket progress updates
// - Batch processing coordination
//
// NOT for:
// - Long-term storage (use user's S3/filesystem)
// - Sensitive key material (use KeyStore with HSM)
// - Audit logs (use encrypted append-only log)
type EncryptedTempRepository interface {
    // Store encrypts and stores data with TTL
    Store(ctx context.Context, key string, data []byte, ttl time.Duration) error

    // Retrieve decrypts and returns data
    Retrieve(ctx context.Context, key string) ([]byte, error)

    // Delete securely removes data (overwrite with zeros)
    Delete(ctx context.Context, key string) error

    // Cleanup removes all expired entries (background goroutine)
    Cleanup(ctx context.Context) error
}
```

### Implementation Design

```go
// EncryptedTempStorage implements EncryptedTempRepository
type EncryptedTempStorage struct {
    mu        sync.RWMutex
    entries   map[string]*encryptedEntry
    keyDeriver KeyDeriver  // Derives AES keys from session tokens
    ttl       time.Duration
    logger    *slog.Logger

    // Optional: Store encrypted blobs on tmpfs (RAM-backed filesystem)
    tmpfsPath string  // e.g., /dev/shm/shadowforge (Linux)
}

type encryptedEntry struct {
    ciphertext []byte      // AES-256-GCM encrypted data
    nonce      []byte      // Unique nonce for GCM
    expiresAt  time.Time   // Auto-delete after this time

    // Optional: Store on tmpfs instead of RAM
    tmpfsPath  string      // Path to encrypted file on tmpfs
}

// Store encrypts data and sets expiration
func (s *EncryptedTempStorage) Store(ctx context.Context, key string, data []byte, ttl time.Duration) error {
    // 1. Derive encryption key from session
    sessionKey := extractSessionKey(ctx)
    aesKey, err := s.keyDeriver.DeriveKey(sessionKey, []byte("temp-storage"))
    if err != nil {
        return fmt.Errorf("key derivation failed: %w", err)
    }
    defer secureZero(aesKey)

    // 2. Encrypt with AES-256-GCM
    nonce := make([]byte, 12)
    if _, err := rand.Read(nonce); err != nil {
        return fmt.Errorf("nonce generation failed: %w", err)
    }

    block, err := aes.NewCipher(aesKey)
    if err != nil {
        return fmt.Errorf("cipher creation failed: %w", err)
    }

    aesgcm, err := cipher.NewGCM(block)
    if err != nil {
        return fmt.Errorf("GCM creation failed: %w", err)
    }

    ciphertext := aesgcm.Seal(nil, nonce, data, nil)

    // 3. Store encrypted entry
    entry := &encryptedEntry{
        ciphertext: ciphertext,
        nonce:      nonce,
        expiresAt:  time.Now().Add(ttl),
    }

    // Optional: Write to tmpfs
    if s.tmpfsPath != "" {
        tmpPath := filepath.Join(s.tmpfsPath, key+".enc")
        if err := os.WriteFile(tmpPath, ciphertext, 0600); err != nil {
            return fmt.Errorf("tmpfs write failed: %w", err)
        }
        entry.tmpfsPath = tmpPath
        entry.ciphertext = nil  // Don't store in RAM too
    }

    s.mu.Lock()
    s.entries[key] = entry
    s.mu.Unlock()

    s.logger.Debug("encrypted data stored",
        slog.String("key", key),
        slog.Duration("ttl", ttl))

    return nil
}

// Retrieve decrypts and returns data
func (s *EncryptedTempStorage) Retrieve(ctx context.Context, key string) ([]byte, error) {
    s.mu.RLock()
    entry, exists := s.entries[key]
    s.mu.RUnlock()

    if !exists {
        return nil, ErrNotFound
    }

    // Check expiration
    if time.Now().After(entry.expiresAt) {
        s.Delete(ctx, key)  // Auto-delete expired
        return nil, ErrExpired
    }

    // Read from tmpfs if needed
    var ciphertext []byte
    if entry.tmpfsPath != "" {
        var err error
        ciphertext, err = os.ReadFile(entry.tmpfsPath)
        if err != nil {
            return nil, fmt.Errorf("tmpfs read failed: %w", err)
        }
    } else {
        ciphertext = entry.ciphertext
    }

    // Derive key and decrypt
    sessionKey := extractSessionKey(ctx)
    aesKey, err := s.keyDeriver.DeriveKey(sessionKey, []byte("temp-storage"))
    if err != nil {
        return nil, fmt.Errorf("key derivation failed: %w", err)
    }
    defer secureZero(aesKey)

    block, err := aes.NewCipher(aesKey)
    if err != nil {
        return nil, fmt.Errorf("cipher creation failed: %w", err)
    }

    aesgcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("GCM creation failed: %w", err)
    }

    plaintext, err := aesgcm.Open(nil, entry.nonce, ciphertext, nil)
    if err != nil {
        return nil, fmt.Errorf("decryption failed: %w", err)
    }

    return plaintext, nil
}

// Delete securely removes entry
func (s *EncryptedTempStorage) Delete(ctx context.Context, key string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    entry, exists := s.entries[key]
    if !exists {
        return nil  // Already deleted
    }

    // Secure zero in-memory data
    if entry.ciphertext != nil {
        secureZero(entry.ciphertext)
    }
    secureZero(entry.nonce)

    // Delete tmpfs file
    if entry.tmpfsPath != "" {
        // Overwrite with random data before deletion
        if err := secureDeleteFile(entry.tmpfsPath); err != nil {
            s.logger.Warn("tmpfs secure delete failed",
                slog.String("error", err.Error()))
        }
    }

    delete(s.entries, key)

    s.logger.Debug("encrypted entry deleted", slog.String("key", key))

    return nil
}

// Cleanup removes expired entries (run as background goroutine)
func (s *EncryptedTempStorage) Cleanup(ctx context.Context) error {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            s.cleanupExpired()
        }
    }
}

func (s *EncryptedTempStorage) cleanupExpired() {
    now := time.Now()

    s.mu.Lock()
    defer s.mu.Unlock()

    var expired []string
    for key, entry := range s.entries {
        if now.After(entry.expiresAt) {
            expired = append(expired, key)
        }
    }

    for _, key := range expired {
        s.Delete(context.Background(), key)
    }

    if len(expired) > 0 {
        s.logger.Info("cleaned up expired entries",
            slog.Int("count", len(expired)))
    }
}
```

### Key Derivation

```go
type KeyDeriver interface {
    // DeriveKey derives an AES-256 key from session token
    DeriveKey(sessionToken, purpose []byte) ([]byte, error)
}

type HKDF256Deriver struct{}

func (d *HKDF256Deriver) DeriveKey(sessionToken, purpose []byte) ([]byte, error) {
    // Use HKDF-SHA256 for key derivation
    // sessionToken = JWT payload or secure random session ID
    // purpose = "temp-storage", "encryption", etc.

    hash := sha256.New
    hkdf := hkdf.New(hash, sessionToken, nil, purpose)

    key := make([]byte, 32)  // 256 bits
    if _, err := io.ReadFull(hkdf, key); err != nil {
        return nil, err
    }

    return key, nil
}
```

### Secure File Deletion

```go
func secureDeleteFile(path string) error {
    // 1. Get file size
    info, err := os.Stat(path)
    if err != nil {
        return err
    }

    // 2. Overwrite with random data (DoD 5220.22-M: 3 passes)
    f, err := os.OpenFile(path, os.O_WRONLY, 0600)
    if err != nil {
        return err
    }
    defer f.Close()

    size := info.Size()
    buf := make([]byte, 4096)

    // Pass 1: Random data
    for i := int64(0); i < size; i += int64(len(buf)) {
        rand.Read(buf)
        f.Write(buf)
    }
    f.Sync()

    // Pass 2: Inverse
    f.Seek(0, 0)
    for i := 0; i < len(buf); i++ {
        buf[i] = 0xFF
    }
    for i := int64(0); i < size; i += int64(len(buf)) {
        f.Write(buf)
    }
    f.Sync()

    // Pass 3: Zeros
    f.Seek(0, 0)
    for i := 0; i < len(buf); i++ {
        buf[i] = 0x00
    }
    for i := int64(0); i < size; i += int64(len(buf)) {
        f.Write(buf)
    }
    f.Sync()

    // 3. Close and delete
    f.Close()
    return os.Remove(path)
}
```

### Configuration

```yaml
# config.yaml - API Server
shadowforge:
  api:
    temp_storage:
      enabled: true
      backend: "tmpfs"  # "memory" or "tmpfs"
      tmpfs_path: "/dev/shm/shadowforge"  # Linux tmpfs (RAM-backed)
      max_ttl: "5m"  # Maximum time-to-live
      cleanup_interval: "1m"
      max_size: "100MB"  # Per-entry limit
      total_limit: "1GB"  # Total storage limit
```

---

## Memory Security

### Secure Zeroing

```go
// secureZero overwrites buffer with zeros to prevent data leakage
func secureZero(b []byte) {
    for i := range b {
        b[i] = 0
    }
    // Prevent compiler optimization
    runtime.KeepAlive(b)
}

// Example usage
func processSecret(secret []byte) error {
    defer secureZero(secret)  // Always defer cleanup

    // Process secret...

    return nil
    // secureZero runs here
}
```

### Memory Allocation Strategy

```go
// Use sync.Pool for frequently allocated buffers
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 64*1024)  // 64KB
    },
}

func processMedia(data []byte) error {
    // Get buffer from pool
    buf := bufferPool.Get().([]byte)
    defer func() {
        secureZero(buf)       // Zero before return
        bufferPool.Put(buf)   // Return to pool
    }()

    // Use buf for processing...

    return nil
}
```

### Preventing Core Dumps

```go
// Disable core dumps in production (Unix systems)
func init() {
    // Set core dump size to 0
    var rlimit syscall.Rlimit
    rlimit.Max = 0
    rlimit.Cur = 0
    syscall.Setrlimit(syscall.RLIMIT_CORE, &rlimit)
}
```

---

## Threat Model

### Threats Mitigated

1. **Forensic Analysis** ✅
   - No persistent storage = no forensic evidence
   - Secure memory zeroing prevents RAM dumps
   - Auto-expiration destroys temporary data

2. **Database Compromise** ✅
   - No database = no database to compromise
   - Encrypted temp storage uses ephemeral keys
   - Key derived from session, never stored

3. **Server Seizure** ✅
   - tmpfs cleared on reboot
   - Encrypted data unrecoverable without session key
   - No long-term data survives

4. **Memory Dump Attacks** ⚠️ (Partially Mitigated)
   - Secure zeroing after use
   - Short in-memory lifetime
   - Encrypted at rest in temp storage
   - **Residual risk:** Live memory while processing

5. **Insider Threats** ✅
   - No persistent storage to exfiltrate
   - Audit logs contain metadata only
   - Encryption keys not accessible to operators

### Threats NOT Mitigated

1. **Live Process Memory Dump** ❌
   - If attacker has root and captures memory during processing
   - **Mitigation:** Minimize processing time, use encrypted temp storage

2. **Compromised Client** ❌
   - If user's machine is compromised, all bets are off
   - **Mitigation:** Document security best practices for users

3. **Network Interception (API mode)** ⚠️
   - HTTPS required, but TLS termination exposes data
   - **Mitigation:** End-to-end encryption, mutual TLS

4. **Swap File Exposure** ⚠️
   - OS may swap RAM to disk
   - **Mitigation:** Disable swap, use mlock(), prefer tmpfs

---

## Implementation Guidelines

### Phase 2.3 (Current) - Testing Infrastructure

✅ **MemoryRepository:**
- Use for unit/integration tests only
- Add security warnings in code comments
- Document "NO PRODUCTION USE"

### Phase 5 - CLI Implementation

✅ **Direct File I/O:**
```go
// NO repository pattern
// Read → Process → Write → Cleanup
```

✅ **Secure Memory Handling:**
```go
defer secureZero(coverData)
defer secureZero(secretData)
defer secureZero(stegoOutput)
```

✅ **Disable Core Dumps:**
```go
syscall.Setrlimit(syscall.RLIMIT_CORE, &rlimit{Max: 0, Cur: 0})
```

### Phase 6 - API Server Implementation

✅ **EncryptedTempRepository:**
```go
repo := NewEncryptedTempStorage(
    ttl: 5 * time.Minute,
    backend: "tmpfs",
    tmpfsPath: "/dev/shm/shadowforge",
)
```

✅ **Background Cleanup:**
```go
go repo.Cleanup(ctx)  // Auto-delete expired entries
```

✅ **Request Handler Pattern:**
```go
func (h *EmbedHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Parse upload
    coverData, _ := parseUpload(r)
    defer secureZero(coverData)

    // 2. Store in encrypted temp (optional, for async)
    key := generateSessionKey(r)
    repo.Store(ctx, key, coverData, 5*time.Minute)

    // 3. Process
    stego, _ := h.service.Embed(ctx, coverData, secretData)
    defer secureZero(stego)

    // 4. Stream response
    w.Write(stego)

    // 5. Delete immediately
    repo.Delete(ctx, key)
}
```

---

## Security Checklist

### For CLI (Phase 5)

- [ ] No repository pattern used
- [ ] All file I/O uses user-controlled paths
- [ ] `defer secureZero()` on all sensitive buffers
- [ ] Core dumps disabled (`RLIMIT_CORE = 0`)
- [ ] No process state persisted between runs
- [ ] Audit logs minimal (timestamps only, no file data)

### For API Server (Phase 6)

- [ ] EncryptedTempRepository implemented
- [ ] AES-256-GCM encryption for temp storage
- [ ] Keys derived from session (HKDF-SHA256)
- [ ] TTL enforced (max 5 minutes)
- [ ] Auto-cleanup background goroutine
- [ ] tmpfs backend configured (`/dev/shm`)
- [ ] Secure file deletion (3-pass overwrite)
- [ ] `defer secureZero()` on all buffers
- [ ] HTTPS/TLS required
- [ ] Rate limiting to prevent storage exhaustion
- [ ] Swap disabled or mlock() used
- [ ] Audit logs metadata only

---

## Future Enhancements

### Phase 7+ - Advanced Security

1. **Hardware Security Module (HSM) Integration**
   - Store user keys in HSM
   - Cryptographic operations in secure enclave

2. **Secure Enclave (Intel SGX / ARM TrustZone)**
   - Process sensitive data in encrypted memory region
   - Protect against root-level memory dumps

3. **Differential Privacy for Audit Logs**
   - Add noise to timing data
   - Prevent statistical analysis of usage patterns

4. **Zero-Knowledge Proofs**
   - Prove operation occurred without revealing data
   - Enable verifiable computation

5. **Self-Destruct Mechanisms**
   - Panic button API endpoint
   - Immediate secure deletion of all data
   - Process termination with memory scrubbing

---

## References

- [NIST SP 800-88 Rev. 1](https://csrc.nist.gov/publications/detail/sp/800-88/rev-1/final) - Media Sanitization
- [DoD 5220.22-M](https://www.epa.gov/sites/default/files/2015-11/documents/dod_5220_22-m_data_sanitization_jan2014.pdf) - Sanitization Standard
- [OWASP Cryptographic Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [Go Secure Coding Practices](https://github.com/OWASP/Go-SCP)

---

**Document Status:** Ready for Implementation
**Next Review:** After Phase 6 API Server Implementation
**Owner:** Security Team

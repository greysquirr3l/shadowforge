# Shadowforge Security Documentation

> **"A steganography tool that leaves no trace"**

## Overview

This directory contains Shadowforge's security architecture documentation, focused on **ephemeral processing** and **zero-persistence** models designed to protect users in adversarial environments.

## Core Security Principle

```
┌──────────────────────────────────────────────────────┐
│ "If it's not in RAM actively being processed,       │
│  it shouldn't exist at all."                         │
└──────────────────────────────────────────────────────┘
```

## Documents

### [ephemeral_storage_architecture.md](./ephemeral_storage_architecture.md)

**Complete specification for Shadowforge's storage security model.**

**Topics covered:**
- ❌ Why persistent databases are FORBIDDEN
- ✅ CLI direct file I/O pattern (Phase 5)
- ✅ API EncryptedTempRepository design (Phase 6)
- ✅ Secure memory handling
- ✅ Threat model and mitigations
- ✅ Implementation checklists

**Status:** Approved for implementation
**Phase:** 2.3+ (Current and future)

## Quick Reference

### Storage Decision Tree

```
┌─────────────────────────────────────────────┐
│ What are you building?                      │
└─────────────────────────────────────────────┘
                  ↓
         ┌────────┴────────┐
         ↓                 ↓
    ┌─────────┐      ┌──────────┐
    │   CLI   │      │   API    │
    │ (Phase  │      │ (Phase   │
    │   5)    │      │   6)     │
    └─────────┘      └──────────┘
         ↓                 ↓
    ┌─────────────────────────────┐
    │ Direct File I/O             │
    │ • No repository             │
    │ • User controls paths       │
    │ • Secure memory zeroing     │
    │ • Process exits cleanly     │
    └─────────────────────────────┘
                             ↓
                   ┌──────────────────────────┐
                   │ EncryptedTempRepository  │
                   │ • AES-256-GCM encryption │
                   │ • Session-scoped keys    │
                   │ • 5-minute max TTL       │
                   │ • Auto-delete on send    │
                   │ • tmpfs backend (RAM)    │
                   └──────────────────────────┘
```

### What About Tests?

```
┌─────────────────────────────────────────────┐
│ MemoryRepository (internal/infrastructure/  │
│                   media/memory_repository.go)│
│                                              │
│ ⚠️  TEST-ONLY IMPLEMENTATION                │
│ ❌ NEVER use in production                  │
│ ✅ Fast unit/integration tests              │
│ ✅ No database dependencies                 │
└─────────────────────────────────────────────┘
```

## Storage Comparison

| Repository Type | Use Case | Encryption | Persistence | Security Level |
|----------------|----------|------------|-------------|----------------|
| **MemoryRepository** | Tests only | ❌ None | ❌ None | ⚠️ Test isolation only |
| **Direct File I/O** | CLI (Phase 5) | ✅ PQC | ✅ User-controlled | ✅ High |
| **EncryptedTempRepository** | API (Phase 6) | ✅ AES-256-GCM | ⏱️ 5 min max | ✅ Very High |
| ~~PostgresRepository~~ | **FORBIDDEN** | ❌ N/A | ❌ N/A | ❌ Security risk |
| ~~S3Repository~~ | **FORBIDDEN** | ❌ N/A | ❌ N/A | ❌ Security risk |

## Security Guarantees

### What Shadowforge Protects Against

- ✅ **Forensic Analysis**: No persistent storage = no forensic evidence
- ✅ **Database Compromise**: No database to compromise
- ✅ **Server Seizure**: tmpfs cleared on reboot, encrypted data unrecoverable
- ✅ **Insider Threats**: No long-term data to exfiltrate
- ✅ **Subpoenas**: No stored data to hand over

### What Shadowforge Does NOT Protect Against

- ❌ **Live Memory Dumps**: Data vulnerable while actively processing
- ❌ **Compromised Client**: User's machine security is their responsibility
- ❌ **Network Interception**: Requires HTTPS/TLS (API mode)
- ⚠️ **Swap File Exposure**: Mitigated by mlock() and tmpfs

## Implementation Status

| Phase | Component | Storage Model | Status |
|-------|-----------|---------------|--------|
| 2.3 | MemoryRepository | In-memory (test-only) | ✅ Complete |
| 5 | CLI | Direct file I/O | ⬜ Planned |
| 6 | API | EncryptedTempRepository | ⬜ Planned |

## Key Files

### Current Implementation
- [internal/infrastructure/media/memory_repository.go](../../internal/infrastructure/media/memory_repository.go)
  - ⚠️ Security warnings in comments
  - ✅ Test-only implementation
  - ❌ NO production use

### Future Implementation
- Phase 5: `internal/interfaces/cli/commands/embed.go` (direct file I/O)
- Phase 6: `internal/infrastructure/storage/encrypted_temp_repository.go`

## Security Checklist

### Before Production Deployment

#### CLI (Phase 5)
- [ ] No repository pattern used
- [ ] All file I/O uses user-controlled paths
- [ ] `defer secureZero()` on all sensitive buffers
- [ ] Core dumps disabled (`RLIMIT_CORE = 0`)
- [ ] No process state persisted between runs
- [ ] Audit logs minimal (timestamps only, no file data)

#### API Server (Phase 6)
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

## Questions?

**Q: Why no PostgreSQL/MongoDB/S3?**
A: Persistent storage creates forensic evidence and makes users vulnerable to server seizure, database compromise, and subpoenas. Shadowforge prioritizes operational security over convenience.

**Q: What about backups?**
A: Users control their own backups by storing stego outputs wherever they choose. Shadowforge never stores data on behalf of users.

**Q: How do I scale the API server?**
A: EncryptedTempRepository uses tmpfs (RAM-backed) and scales horizontally. Session affinity routes users to the same server during processing.

**Q: What if the server crashes during processing?**
A: All data is lost (by design). User must retry the operation. This is a feature, not a bug.

**Q: Can I add persistent storage for `[feature X]`?**
A: Only if:
  1. Data is encrypted with user-controlled keys
  2. User explicitly opts in
  3. Auto-deletion is enforced
  4. Security team reviews the design

## References

- [NIST SP 800-88 Rev. 1](https://csrc.nist.gov/publications/detail/sp/800-88/rev-1/final) - Media Sanitization
- [OWASP Cryptographic Storage](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [Go Secure Coding Practices](https://github.com/OWASP/Go-SCP)

---

**Last Updated:** December 10, 2025
**Owner:** Security Team
**Status:** Active Development

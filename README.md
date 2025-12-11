# Shadowforge

![Shadowforge Logo](docs/assets/shadowforge-logo.png)

> **"Forge secrets in the shadows, shield them from quantum eyes"**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-In%20Development-yellow.svg)](docs/implementation_plan_todo.md)

**Shadowforge** is a production-grade quantum-resistant steganography tool that combines
NIST-approved post-quantum cryptography, Reed-Solomon error correction, and multiple
steganographic techniques. Built with Domain-Driven Design (DDD) and Command Query
Responsibility Segregation (CQRS) patterns.

## 🎯 Project Vision

Shadowforge provides military-grade data hiding with post-quantum security, designed to
withstand both current and future cryptographic attacks. The system supports multiple
distribution patterns, from simple one-to-one embedding to complex distributed networks
with `K-of-N` threshold recovery.

## ⭐ Key Features

### Post-Quantum Cryptography

- **Kyber-1024** (NIST-approved KEM) for encryption
- **Dilithium3** (NIST-approved) for digital signatures
- Constant-time operations to prevent timing attacks
- Secure memory handling with automatic key zeroing

### Reed-Solomon Error Correction

- Configurable data/parity shard ratios
- K-of-N threshold recovery (need only K shards from N total)
- Corruption detection and automatic recovery
- Up to 50% redundancy for maximum fault tolerance

### Steganography Techniques

- **Image**: LSB, DCT (JPEG), Palette-based (GIF/PNG)
- **Audio**: Phase encoding, Echo hiding, LSB audio
- **Text**: Zero-width characters, whitespace manipulation
- Capacity-aware embedding with statistical analysis

### Distribution Patterns

- **One-to-One**: Traditional steganography (1 payload → 1 cover)
- **One-to-Many**: Distributed secret splitting (1 payload → N covers)
- **Many-to-One**: Batch embedding (N payloads → 1 cover)
- **Many-to-Many**: Complex distribution matrix (N payloads → M covers)

### Archive Support

- ZIP, TAR, TAR.GZ format support
- Archive-to-archive workflows
- Nested archive handling
- Automatic format detection

## 🏗️ Architecture

Shadowforge follows **Clean Architecture** principles with clear separation of concerns:

```text
┌─────────────────────────────────────────┐
│ Interface Layer (CLI/API)               │
│ ├─ Cobra CLI commands                   │
│ └─ Echo REST API endpoints              │
├─────────────────────────────────────────┤
│ Application Layer (CQRS)                │
│ ├─ Command Handlers                     │
│ ├─ Query Handlers                       │
│ └─ Application Services                 │
├─────────────────────────────────────────┤
│ Domain Layer (8 Bounded Contexts)       │
│ ├─ Cryptography                         │
│ ├─ Error Correction                     │
│ ├─ Steganography                        │
│ ├─ Media Processing                     │
│ ├─ Security Analysis                    │
│ ├─ Distribution                         │
│ ├─ Reconstruction                       │
│ └─ Archive                              │
├─────────────────────────────────────────┤
│ Infrastructure Layer                    │
│ ├─ CIRCL PQC integration                │
│ ├─ Reed-Solomon implementation          │
│ ├─ File I/O and media processing        │
│ └─ External service adapters            │
└─────────────────────────────────────────┘
```

## 🛠️ Technology Stack

| Category | Technology |
|----------|-----------|
| **Language** | Go 1.21+ |
| **Architecture** | DDD + CQRS + Clean Architecture |
| **PQC** | [cloudflare/circl](https://github.com/cloudflare/circl) |
| **Error Correction** | [klauspost/reedsolomon](https://github.com/klauspost/reedsolomon) |
| **CLI** | [spf13/cobra](https://github.com/spf13/cobra) |
| **API** | [labstack/echo](https://github.com/labstack/echo) |
| **Testing** | [stretchr/testify](https://github.com/stretchr/testify), [golang/mock](https://github.com/golang/mock) |
| **Logging** | log/slog (stdlib) |

## 📦 Installation

**Note**: Shadowforge is currently in active development (Phase 1.3 - Application Layer).

### Prerequisites

- Go 1.21 or higher
- Git

### From Source

```bash
# Clone the repository
git clone https://github.com/greysquirr3l/shadowforge.git
cd shadowforge

# Install dependencies
go mod download

# Build the CLI
make build

# Run tests
make test
```

## 🚀 Quick Start

**Coming Soon**: CLI and API interfaces are currently under development.

### Planned CLI Usage

```bash
# Generate PQC key pair
shadowforge keygen --algorithm kyber1024 --output keys/

# Encrypt and embed (one-to-one)
shadowforge embed \
  --input secret.txt \
  --cover image.png \
  --output stego.png \
  --key keys/public.key

# Distributed embedding (one-to-many with K-of-N recovery)
shadowforge embed-distributed \
  --input document.pdf \
  --covers covers/*.png \
  --data-shards 10 \
  --parity-shards 5 \
  --output-archive stego-bundle.zip

# Extract from stego media
shadowforge extract \
  --input stego.png \
  --key keys/private.key \
  --output recovered.txt

# Analyze capacity
shadowforge analyze capacity --cover image.png --technique lsb
```

### Planned API Usage

```bash
# Start API server
shadowforge-api --port 8080

# Embed via REST API
curl -X POST http://localhost:8080/api/v1/embed \
  -H "Content-Type: application/json" \
  -d '{
    "payload": "base64_encoded_data",
    "cover": "base64_encoded_image",
    "algorithm": "kyber1024",
    "technique": "lsb"
  }'
```

## 📊 Development Status

### Current Phase: **Phase 1.3 - Application Layer (CQRS)**

| Phase | Status | Description |
|-------|--------|-------------|
| **Phase 1.1** | ✅ Complete | Project setup, dependencies, structure |
| **Phase 1.2** | ✅ Complete | Domain layer (8 bounded contexts, 91% coverage) |
| **Phase 1.3** | 🚧 In Progress | Application layer CQRS implementation |
| **Phase 2** | ⏳ Planned | Core domain services |
| **Phase 3** | ⏳ Planned | Steganography techniques |
| **Phase 4** | ⏳ Planned | Distribution patterns |
| **Phase 5** | ⏳ Planned | CLI application |
| **Phase 6** | ⏳ Planned | REST API server |

### Recent Progress

- ✅ CQRS infrastructure (CommandBus, QueryBus) - 94% coverage
- ✅ Crypto domain commands (Encrypt, Decrypt, KeyGen) - 87.8% coverage
- ✅ Repository pattern for key persistence
- ✅ Comprehensive test suite with race detection
- ✅ Zero race conditions, all tests passing

See [docs/implementation_plan_todo.md](docs/implementation_plan_todo.md) for detailed roadmap.

## 🧪 Testing

Shadowforge follows strict testing standards:

- **Domain Logic**: 85%+ coverage required
- **Cryptographic Operations**: 90%+ coverage required
- **Command/Query Handlers**: 80%+ coverage required
- **Race Detection**: All tests run with `-race` flag
- **Table-Driven Tests**: Comprehensive scenario coverage

```bash
# Run all tests
go test -v -race ./...

# Run tests with coverage
go test -v -race -cover ./...

# Run specific domain tests
go test -v -race ./internal/domain/crypto/

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔒 Security

### Security-First Principles

- ✅ Constant-time cryptographic operations
- ✅ Automatic key zeroing after use
- ✅ No panic/recover in production code
- ✅ All errors handled explicitly
- ✅ Input validation on all boundaries
- ✅ Secure random number generation (crypto/rand)
- ✅ Statistical analysis for detectability

### Cryptographic Standards

- **NIST Post-Quantum Standards** (Kyber-1024, Dilithium3)
- **Argon2id** for password-based key derivation
- **AES-256-GCM** for symmetric encryption
- **SHA3/SHAKE** for hashing and extendable output

### Security Audit Status

⚠️ **Pre-Production**: This software is under active development and has NOT been security
audited. Do not use in production environments until Phase 7 (Security Hardening) is
complete.

## 📚 Documentation

- [Complete Specification](INITIAL_PROMPT.md) - Detailed project requirements
- [Architecture Guide](docs/architecture.md) - System design with Mermaid diagrams
- [Implementation Plan](docs/implementation_plan_todo.md) - 9-phase development roadmap
- [Go Instructions](.github/instructions/go.instructions.md) - Development standards
- [Testing Standards](.github/instructions/testing.instructions.md) - Test requirements
- [Problem Resolution](.github/instructions/problem-resolution.instructions.md) - Debugging protocols

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Read the [Go Instructions](.github/instructions/go.instructions.md)
2. Follow Clean Architecture boundaries
3. Maintain 80%+ test coverage
4. Run `go fmt`, `go vet`, and `golangci-lint`
5. All tests must pass with `-race` flag
6. Document security-critical code thoroughly

### Development Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes and test
go test -v -race ./...

# Format and lint
go fmt ./...
go vet ./...
golangci-lint run

# Commit with conventional commits
git commit -m "feat(domain): add new bounded context"

# Push and create PR
git push origin feature/your-feature-name
```

## 🎯 Roadmap

### Phase 5: CLI Application (Q1 2026)

- Complete cross-platform CLI (shadowforge/sforge)
- All distribution patterns implemented
- Comprehensive help and examples

### Phase 6: REST API Server (Q2 2026)

- Full-featured REST API
- Async operation support
- WebSocket progress updates

### Phase 7: Security Hardening (Q2 2026)

- External security audit
- Penetration testing
- Production hardening

### Phase 8: Testing & Documentation (Q3 2026)

- E2E test suite
- Performance benchmarks
- Complete user documentation

### Phase 9: Production Release (Q3 2026)

- CI/CD pipeline
- Multi-platform releases
- Docker images

## 📄 License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## ⚠️ Disclaimer

This software is provided "as is" without warranty of any kind. The developers are not
responsible for any misuse of this software. Shadowforge is designed for legitimate privacy
and security purposes only. Always comply with applicable laws and regulations.

## 🔗 Links

- **GitHub**: <https://github.com/greysquirr3l/shadowforge>
- **Documentation**: [docs/](docs/)
- **Issues**: <https://github.com/greysquirr3l/shadowforge/issues>

---

**Built with 🌮 and quantum-resistant cryptography**

*Last Updated: December 2025*

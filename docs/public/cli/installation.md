# Installation

## System Requirements

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| **OS** | macOS 11, Linux (glibc 2.29+), Windows 10 | Latest stable version |
| **Architecture** | x86-64, ARM64 | x86-64 or ARM64 native |
| **Memory** | 256 MB | 2-4 GB |
| **Disk Space** | 50 MB | 100 MB + working space |

## Installation Methods

### 1. Homebrew (macOS / Linux)

Coming soon! We're preparing Homebrew formula distribution. For now, use the source installation method.

### 2. Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/greysquirr3l/shadowforge/releases):

**macOS (Apple Silicon)**:
```bash
curl -L https://github.com/greysquirr3l/shadowforge/releases/download/v1.0.0/shadowforge-darwin-arm64.tar.gz | tar xz
sudo mv shadowforge /usr/local/bin/
shadowforge version
```

**macOS (Intel)**:
```bash
curl -L https://github.com/greysquirr3l/shadowforge/releases/download/v1.0.0/shadowforge-darwin-amd64.tar.gz | tar xz
sudo mv shadowforge /usr/local/bin/
shadowforge version
```

**Linux (x86-64)**:
```bash
curl -L https://github.com/greysquirr3l/shadowforge/releases/download/v1.0.0/shadowforge-linux-amd64.tar.gz | tar xz
sudo mv shadowforge /usr/local/bin/
shadowforge version
```

**Linux (ARM64)**:
```bash
curl -L https://github.com/greysquirr3l/shadowforge/releases/download/v1.0.0/shadowforge-linux-arm64.tar.gz | tar xz
sudo mv shadowforge /usr/local/bin/
shadowforge version
```

**Windows (PowerShell)**:
```powershell
Invoke-WebRequest -Uri "https://github.com/greysquirr3l/shadowforge/releases/download/v1.0.0/shadowforge-windows-amd64.zip" -OutFile "shadowforge.zip"
Expand-Archive -Path "shadowforge.zip" -DestinationPath "C:\Program Files\shadowforge"
$env:Path += ";C:\Program Files\shadowforge"
shadowforge version
```

### 3. From Source

**Prerequisites**:
- Go 1.21 or later
- Git

**Installation**:
```bash
# Clone the repository
git clone https://github.com/greysquirr3l/shadowforge.git
cd shadowforge

# Build the CLI
go build -o shadowforge ./cmd/cli

# Test it
./shadowforge version

# Optional: Install to system path
sudo mv shadowforge /usr/local/bin/
```

### 4. Docker

**Run in Docker** (no installation needed):
```bash
docker run --rm \
  -v $(pwd):/work \
  -w /work \
  greysquirr3l/shadowforge:latest \
  shadowforge embed --input secret.txt --cover image.png --output stego.png
```

## Verification

After installation, verify the binary works:

```bash
# Check version
shadowforge version

# Run diagnostics
shadowforge formats

# Display help
shadowforge help
```

## Updating

### From Homebrew
```bash
brew upgrade shadowforge
```

### From Source
```bash
cd ~/path/to/shadowforge
git pull origin main
go build -o shadowforge ./cmd/cli
sudo mv shadowforge /usr/local/bin/
```

### From Pre-built Binaries
Download the latest release and follow the installation steps above.

## Shell Completion

### Bash

Add to `~/.bashrc` or `~/.bash_profile`:
```bash
eval "$(shadowforge completion bash)"
```

### Zsh

Add to `~/.zshrc`:
```bash
eval "$(shadowforge completion zsh)"
```

### Fish

```bash
shadowforge completion fish | source
```

### PowerShell

Add to PowerShell profile:
```powershell
Register-ArgumentCompleter -CommandName shadowforge -ScriptBlock {
    # Completion implementation
}
```

## Troubleshooting Installation

### "Command not found: shadowforge"

**Solution**: Add the installation directory to your PATH:
```bash
# If installed to /usr/local/bin
export PATH="/usr/local/bin:$PATH"

# Add to ~/.bashrc or ~/.zshrc permanently
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
```

### "Permission denied" when running binary

**Solution**: Make the binary executable:
```bash
chmod +x /path/to/shadowforge
```

### OpenSSL or library errors on Linux

**Solution**: Install required dependencies:
```bash
# Ubuntu/Debian
sudo apt-get install libssl-dev libc6

# CentOS/RHEL
sudo yum install openssl-devel glibc

# Alpine
apk add openssl-libs libc6
```

### macOS: "shadowforge cannot be opened because the developer cannot be verified"

**Solution**: Allow the app through Gatekeeper:
```bash
xattr -d com.apple.quarantine $(which shadowforge)
```

## Uninstallation

### Remove Binary
```bash
# If installed to /usr/local/bin
rm /usr/local/bin/shadowforge

# If installed elsewhere
which shadowforge  # Find the path
rm <path-to-binary>
```

### Remove Configuration (Optional)
```bash
# Remove cached files
rm -rf ~/.shadowforge/cache

# Remove config files
rm -rf ~/.config/shadowforge  # Linux
rm -rf ~/Library/Application\ Support/shadowforge  # macOS
```

## Next Steps

After installation:
1. Read [Getting Started](./getting-started.md) for basic usage
2. Try the [Quick Start](./README.md#-quick-start) example
3. Review [Best Practices](./best-practices.md) for security guidance

---

**Need help?** See [Troubleshooting](./troubleshooting.md) or check [FAQ](./faq.md)

#!/usr/bin/env bash
# Build script for Shadowforge CLI - Cross-platform builds

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="shadowforge"
CMD_PATH="./cmd/cli"
OUTPUT_DIR="bin"
VERSION=$(cat VERSION 2>/dev/null || echo "dev")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS="-X github.com/greysquirr3l/shadowforge/pkg/version.Version=${VERSION}"
LDFLAGS="${LDFLAGS} -X github.com/greysquirr3l/shadowforge/pkg/version.GitCommit=${GIT_COMMIT}"
LDFLAGS="${LDFLAGS} -X github.com/greysquirr3l/shadowforge/pkg/version.BuildDate=${BUILD_DATE}"

# Platforms to build for
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

print_header() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  🔨 Shadowforge Build Script${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  Version:    ${VERSION}${NC}"
    echo -e "${BLUE}  Commit:     ${GIT_COMMIT}${NC}"
    echo -e "${BLUE}  Build Date: ${BUILD_DATE}${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Clean previous builds
clean() {
    print_info "Cleaning previous builds..."
    rm -rf "${OUTPUT_DIR}"
    mkdir -p "${OUTPUT_DIR}"
    print_success "Clean complete"
}

# Build for current platform only
build_local() {
    print_info "Building for local platform..."

    local output="${OUTPUT_DIR}/${PROJECT_NAME}"
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        output="${output}.exe"
    fi

    go build -ldflags "${LDFLAGS}" -o "${output}" "${CMD_PATH}"

    if [[ -f "${output}" ]]; then
        print_success "Built: ${output}"

        # Make executable
        chmod +x "${output}" 2>/dev/null || true

        # Show binary info
        ls -lh "${output}"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Build for all platforms
build_all() {
    print_info "Building for all platforms..."

    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r GOOS GOARCH <<< "${platform}"

        output="${OUTPUT_DIR}/${PROJECT_NAME}-${GOOS}-${GOARCH}"
        if [[ "$GOOS" == "windows" ]]; then
            output="${output}.exe"
        fi

        echo ""
        print_info "Building ${GOOS}/${GOARCH}..."

        if GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
            -ldflags "${LDFLAGS}" \
            -o "${output}" \
            "${CMD_PATH}"; then

            print_success "Built: ${output} ($(du -h "${output}" | cut -f1))"
        else
            print_error "Failed to build ${GOOS}/${GOARCH}"
        fi
    done
}

# Build optimized release binary
build_release() {
    print_info "Building optimized release binary..."

    local output="${OUTPUT_DIR}/${PROJECT_NAME}"
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        output="${output}.exe"
    fi

    # Additional release flags
    local release_ldflags="${LDFLAGS} -s -w" # Strip debug info

    go build \
        -ldflags "${release_ldflags}" \
        -trimpath \
        -o "${output}" \
        "${CMD_PATH}"

    if [[ -f "${output}" ]]; then
        print_success "Release build complete: ${output}"
        chmod +x "${output}" 2>/dev/null || true
        ls -lh "${output}"

        # Optional: Compress binary
        if command -v upx &> /dev/null; then
            print_info "Compressing binary with UPX..."
            upx --best --lzma "${output}" 2>/dev/null || upx --best "${output}"
            print_success "Compressed: ${output}"
            ls -lh "${output}"
        fi
    else
        print_error "Release build failed"
        exit 1
    fi
}

# Run tests before building
test() {
    print_info "Running tests..."

    if go test -short ./...; then
        print_success "Tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
}

# Show usage
usage() {
    cat << EOF
Usage: $0 [COMMAND]

Commands:
    local       Build for current platform only (default)
    all         Build for all platforms
    release     Build optimized release binary
    clean       Clean build artifacts
    test        Run tests before building
    help        Show this help message

Examples:
    $0               # Build for current platform
    $0 local         # Same as above
    $0 all           # Build for all platforms
    $0 release       # Build optimized release
    $0 test local    # Test then build local

Environment variables:
    VERSION         Override version (default: from VERSION file or "dev")

EOF
}

# Main script
main() {
    print_header

    local command="${1:-local}"

    case "${command}" in
        local)
            clean
            build_local
            ;;
        all)
            clean
            build_all
            ;;
        release)
            clean
            build_release
            ;;
        clean)
            clean
            ;;
        test)
            test
            shift
            main "${@:-local}"
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            print_error "Unknown command: ${command}"
            echo ""
            usage
            exit 1
            ;;
    esac

    echo ""
    print_success "Build complete! 🎉"
}

# Run main
main "$@"

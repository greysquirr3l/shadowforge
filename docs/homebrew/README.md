# Homebrew Distribution (shadowforge)

✅ **Status**: Homebrew tap is live at [greysquirr3l/homebrew-shadowforge](https://github.com/greysquirr3l/homebrew-shadowforge)

This project is a CLI tool, so Homebrew distribution uses a **formula** (not a cask).

## User Installation

Users can install shadowforge via Homebrew:

```bash
brew tap greysquirr3l/shadowforge
brew install shadowforge
```

## Tap Repository

The tap lives at: https://github.com/greysquirr3l/homebrew-shadowforge

Formula location: `Formula/shadowforge.rb`

## Release Assets

The GitHub Actions workflow at `.github/workflows/release.yml` publishes per-platform tarballs on tagged releases:

- `shadowforge_<version>_darwin_amd64.tar.gz`
- `shadowforge_<version>_darwin_arm64.tar.gz`
- `shadowforge_<version>_linux_amd64.tar.gz`
- `shadowforge_<version>_linux_arm64.tar.gz`

A `checksums.txt` file is also uploaded containing SHA-256 sums for all tarballs.

## Recommended Approach

Create a separate **tap** repository (example: `greysquirr3l/homebrew-shadowforge`) and add a binary formula.

- Users install with:
  - `brew tap greysquirr3l/shadowforge`
  - `brew install shadowforge`

## Formula Template

See `docs/homebrew/shadowforge.rb` in this repo as a starting point. In the tap repository it should live at:

- `Formula/shadowforge.rb`

## Release Process

1. Update `VERSION` (if needed)
2. Tag a release: `git tag vX.Y.Z && git push origin vX.Y.Z`
3. GitHub Actions publishes release assets + `checksums.txt`
4. Update the tap formula `version` and the `sha256` values for each platform/arch

### Updating the Tap Formula

In your tap repo (`greysquirr3l/homebrew-shadowforge`), update `Formula/shadowforge.rb`:

- Set `version "X.Y.Z"`
- Update the `sha256` strings by pulling from the release checksums:
  - `curl -L https://github.com/greysquirr3l/shadowforge/releases/download/vX.Y.Z/checksums.txt`
  - Copy the matching hash for each tarball name

Then push the tap update. Users can upgrade with:

- `brew update`
- `brew upgrade shadowforge`

## Notes

- The build is `CGO_ENABLED=0` to keep the binary portable.
- The formula should at minimum run `shadowforge version` in `test do`.

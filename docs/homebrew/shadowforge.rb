# Homebrew formula template for Shadowforge.
#
# Intended usage:
# - Copy to your tap repo as: Formula/shadowforge.rb
# - Update `version`, `url`, and `sha256` blocks for a given release.
#
# NOTE: Homebrew's Ruby DSL selects URL/SHA by platform/arch. Keep the tarball
# layout consistent: tarball contains `shadowforge_<version>_<os>_<arch>/shadowforge`.

class Shadowforge < Formula
  desc "Quantum-resistant steganography tool"
  homepage "https://github.com/greysquirr3l/shadowforge"
  license "Apache-2.0"

  # Example:
  # version "0.6.0"
  version "0.0.0"

  on_macos do
    on_intel do
      url "https://github.com/greysquirr3l/shadowforge/releases/download/v#{version}/shadowforge_#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_ME"
    end

    on_arm do
      url "https://github.com/greysquirr3l/shadowforge/releases/download/v#{version}/shadowforge_#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_ME"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/greysquirr3l/shadowforge/releases/download/v#{version}/shadowforge_#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_ME"
    end

    on_arm do
      url "https://github.com/greysquirr3l/shadowforge/releases/download/v#{version}/shadowforge_#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_ME"
    end
  end

  def install
    candidate = Dir["shadowforge_*/shadowforge"].first
    odie "shadowforge binary not found in archive" if candidate.nil?
    bin.install candidate => "shadowforge"
  end

  test do
    assert_match "shadowforge version", shell_output("#{bin}/shadowforge version")
  end
end

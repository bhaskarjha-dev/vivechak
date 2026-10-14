# typed: false
# frozen_string_literal: true

# Homebrew formula for vck (shorthand CLI alias for Vivechak)
# Install: brew install bhaskarjha-dev/tap/vck
class Vck < Formula
  desc "Shorthand CLI for Vivechak — evidence-grounded research MCP server"
  homepage "https://github.com/bhaskarjha-dev/vivechak"
  version "0.1.0" # Updated by GoReleaser or CI
  license "MIT"

  conflicts_with "vivechak", because: "both install vivechak and vck binaries"

  on_macos do
    url "https://github.com/bhaskarjha-dev/vivechak/releases/download/v#{version}/vivechak_#{version}_darwin_all.tar.gz"
    sha256 "PLACEHOLDER" # Auto-filled by GoReleaser during release — do not edit manually
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/bhaskarjha-dev/vivechak/releases/download/v#{version}/vivechak_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER" # Auto-filled by GoReleaser during release — do not edit manually
    else
      url "https://github.com/bhaskarjha-dev/vivechak/releases/download/v#{version}/vivechak_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER" # Auto-filled by GoReleaser during release — do not edit manually
    end
  end

  def install
    bin.install "vck"
    if File.exist?("vivechak")
      bin.install "vivechak"
    else
      bin.install_symlink "vck" => "vivechak"
    end
  end

  test do
    system "#{bin}/vck", "version"
    system "#{bin}/vivechak", "version"
  end
end

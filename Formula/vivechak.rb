# typed: false
# frozen_string_literal: true

# Homebrew formula for Vivechak
# Install: brew install bhaskarjha-dev/tap/vivechak
class Vivechak < Formula
  desc "Evidence-grounded research for technical decisions — MCP server"
  homepage "https://github.com/bhaskarjha-dev/vivechak"
  version "0.1.0" # Updated by GoReleaser or CI
  license "MIT"

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
    bin.install "vivechak"
    if File.exist?("vck")
      bin.install "vck"
    else
      bin.install_symlink "vivechak" => "vck"
    end
  end

  test do
    system "#{bin}/vivechak", "version"
    system "#{bin}/vck", "version"
  end
end

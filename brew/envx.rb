# Homebrew formula for envx.
#
# Lives in a tap repository (e.g. TJ-programmer/homebrew-envx) as `Formula/envx.rb`.
# Usage:
#   brew tap TJ-programmer/envx
#   brew install envx
#
# On release, update `url` to the new tag and `sha256` (get it with:
#   curl -fsSL <url> | shasum -a 256
# or use `brew bump-formula-pr`).

class Envx < Formula
  desc "A faster, safer drop-in replacement for .env files - project-local env vars, secrets encrypted at rest"
  homepage "https://github.com/TJ-programmer/envx"
  url "https://github.com/TJ-programmer/envx/archive/refs/tags/v0.5.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(
      output: bin/"envx",
      ldflags: "-s -w -X envx/internal/buildinfo.Version=v#{version}"
    )
  end

  test do
    assert_match "envx", shell_output("#{bin}/envx --version")
  end
end

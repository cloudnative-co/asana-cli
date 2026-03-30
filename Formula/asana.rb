class Asana < Formula
  desc "Asana CLI with profile-based auth and automation-friendly output"
  homepage "https://github.com/cloudnative-co/asana-cli"
  license "MIT"

  url "https://github.com/cloudnative-co/asana-cli/archive/refs/tags/v0.1.8.tar.gz"
  sha256 "9ccb7e337603ffe2057d80cadf03ef5729d5505a9e66a0bfae5d3e1a5f39be85"

  # Keep HEAD support for development snapshots.
  head "https://github.com/cloudnative-co/asana-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(output: bin/"asana"), "./cmd/asana"
  end

  test do
    assert_match "Asana CLI", shell_output("#{bin}/asana --help")
  end
end

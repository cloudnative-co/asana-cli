class Asana < Formula
  desc "Asana CLI with profile-based auth and automation-friendly output"
  homepage "https://github.com/cloudnative-co/asana-cli"
  license "MIT"

  url "https://github.com/cloudnative-co/asana-cli/archive/refs/tags/v0.1.7.tar.gz"
  sha256 "cf8e0baa069fc55f0365c4b92a34ad0aeca1671d49383470b912cd40ff7eee0d"

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

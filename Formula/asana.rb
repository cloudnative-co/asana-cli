class Asana < Formula
  desc "Asana CLI with profile-based auth and automation-friendly output"
  homepage "https://github.com/cloudnative-co/asana-cli"
  license "MIT"

  url "https://github.com/cloudnative-co/asana-cli/archive/refs/tags/v0.1.3.tar.gz"
  sha256 "69ea99dfff9317d09f0a2efb836ba38685d89e3e02ffd40eab22d115bd3ea084"




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

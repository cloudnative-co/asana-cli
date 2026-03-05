class Asana < Formula
  desc "Asana CLI with profile-based auth and automation-friendly output"
  homepage "https://github.com/cloudnative-co/asana-cli"
  license "MIT"

  url "https://github.com/cloudnative-co/asana-cli/archive/refs/tags/v0.1.2.tar.gz"
  sha256 "1c5ca27b5f9ff1fe75c1b01d8e956b7b1374f82a71f72a8e8d507e81d56cb902"



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

class Asana < Formula
  desc "Asana CLI with profile-based auth and automation-friendly output"
  homepage "https://github.com/cloudnative-co/asana-cli"
  license "MIT"

  url "https://github.com/cloudnative-co/asana-cli/archive/refs/tags/v0.1.6.tar.gz"
  sha256 "d8ca886c84864876242e2539983865f45bff800d90b094651d7b4e80a087fc3c"

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

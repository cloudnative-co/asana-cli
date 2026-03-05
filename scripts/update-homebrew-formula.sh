#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <tag>"
  echo "Example: $0 v0.1.0"
  exit 1
fi

TAG="$1"
REPO="cloudnative-co/asana-cli"
URL="https://github.com/${REPO}/archive/refs/tags/${TAG}.tar.gz"
FORMULA="Formula/asana.rb"

if [[ ! -f "${FORMULA}" ]]; then
  echo "Formula file not found: ${FORMULA}"
  exit 1
fi

echo "Downloading ${URL} to calculate SHA256..."
TMP_FILE="$(mktemp)"
trap 'rm -f "${TMP_FILE}"' EXIT
curl -fsSL "${URL}" -o "${TMP_FILE}"
SHA256="$(shasum -a 256 "${TMP_FILE}" | awk '{print $1}')"

if ! grep -q '^  head "https://github.com/cloudnative-co/asana-cli.git", branch: "main"$' "${FORMULA}"; then
  echo "Unexpected formula structure; aborting."
  exit 1
fi

awk -v url="${URL}" -v sha="${SHA256}" '
  BEGIN { inserted=0 }
  $0 ~ /^  url "/ { next }
  $0 ~ /^  sha256 "/ { next }
  {
    print $0
    if ($0 ~ /^  license "MIT"$/ && inserted==0) {
      print ""
      print "  url \"" url "\""
      print "  sha256 \"" sha "\""
      inserted=1
    }
  }
' "${FORMULA}" > "${FORMULA}.tmp"

mv "${FORMULA}.tmp" "${FORMULA}"

echo "Updated ${FORMULA}"
echo "  url: ${URL}"
echo "  sha256: ${SHA256}"
echo
echo "Next:"
echo "  1) Review Formula/asana.rb"
echo "  2) brew install cloudnative-co/asana-cli/asana --dry-run"
echo "  3) Commit and push"

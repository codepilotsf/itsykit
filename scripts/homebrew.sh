#!/usr/bin/env bash
# Generate a tap formula from already built release artifacts; never publish it.
set -euo pipefail
cd "$(dirname "$0")/.."
if [[ "${1:-}" == --help ]]; then
  printf 'Usage: RELEASE_BASE_URL=https://host/version bash scripts/homebrew.sh\n\nRequires both Mac release binaries in dist/. Writes dist/itsy.rb.\n\n'
  exit 0
fi
source scripts/version.sh
version=$(flea_version)
base=${RELEASE_BASE_URL:?Set the public HTTPS directory containing release binaries}
[[ "$base" =~ ^https://[A-Za-z0-9./_:-]+$ ]] || { echo 'Use an HTTPS release directory without query parameters.' >&2; exit 1; }
for arch in arm64 amd64; do test -s "dist/flea-darwin-$arch"; done
export FLEA_FORMULA_VERSION="$version" FLEA_FORMULA_BASE="${base%/}"
export FLEA_FORMULA_ARM=$(shasum -a 256 dist/flea-darwin-arm64 | awk '{print $1}')
export FLEA_FORMULA_INTEL=$(shasum -a 256 dist/flea-darwin-amd64 | awk '{print $1}')
awk '{gsub(/@VERSION@/,ENVIRON["FLEA_FORMULA_VERSION"]);gsub(/@BASE@/,ENVIRON["FLEA_FORMULA_BASE"]);gsub(/@ARM_SHA@/,ENVIRON["FLEA_FORMULA_ARM"]);gsub(/@INTEL_SHA@/,ENVIRON["FLEA_FORMULA_INTEL"]);print}' packaging/flea.rb.in > dist/itsy.rb
echo 'Created dist/itsy.rb. Test in a tap before publishing.'

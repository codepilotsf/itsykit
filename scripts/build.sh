#!/usr/bin/env bash
# Contributor build: refresh both servers before compiling the native CLI.
set -euo pipefail
cd "$(dirname "$0")/.."
if [[ -z "${GO_BINARY:-}" ]]; then
  GO_BINARY=$(command -v go || true)
  if [[ -z "$GO_BINARY" ]]; then GO_BINARY=/usr/local/go/bin/go; fi
fi
source scripts/version.sh
version=$(flea_version)
export CGO_ENABLED=0 GOFLAGS=
unset GOOS GOARCH
flags="-s -w -X flea/internal/buildinfo.Version=$version"
stage=$(mktemp -d)
trap 'rm -rf -- "$stage"' EXIT
for arch in amd64 arm64; do
  echo "Building Linux server ($arch)..."
  GOOS=linux GOARCH="$arch" "$GO_BINARY" build -trimpath -ldflags "$flags" -o "$stage/flea-server-linux-$arch" ./cmd/flea-server
done
bash scripts/bundle.sh "$stage"
FLEA_REQUIRE_BUNDLE=1 "$GO_BINARY" test ./internal/serverbundle -count=1
mkdir -p bin
echo 'Building FleaCMS with both servers bundled...'
"$GO_BINARY" build -trimpath -ldflags "$flags" -o "$stage/flea" ./cmd/flea
"$stage/flea" version
# A fresh inode avoids macOS retaining the previous executable's code signature.
# Rename within bin/ also leaves the last working executable intact until ready.
cli_stage=$(mktemp bin/.flea.XXXXXX)
trap 'rm -f -- "$cli_stage"; rm -rf -- "$stage"' EXIT
cp "$stage/flea" "$cli_stage"
chmod 755 "$cli_stage"
mv -f "$cli_stage" bin/flea
echo 'Built bin/flea. Setup is ready without server downloads.'

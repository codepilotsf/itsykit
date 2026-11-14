#!/usr/bin/env bash
# Contributor build: refresh both servers before compiling the native CLI.
set -euo pipefail
cd "$(dirname "$0")/.."
if [[ -z "${GO_BINARY:-}" ]]; then
  GO_BINARY=$(command -v go || true)
  if [[ -z "$GO_BINARY" ]]; then GO_BINARY=/usr/local/go/bin/go; fi
fi
case "$("$GO_BINARY" env GOOS)-$("$GO_BINARY" env GOARCH)" in
  darwin-arm64|linux-amd64) ;;
  *) echo 'CLI builds support Apple Silicon macOS and x86-64 Linux.' >&2; exit 1 ;;
esac
source scripts/version.sh
version=$(itsy_version)
export CGO_ENABLED=0 GOFLAGS=
unset GOOS GOARCH
flags="-s -w -X itsy/internal/buildinfo.Version=$version"
stage=$(mktemp -d)
trap 'rm -rf -- "$stage"' EXIT
for arch in amd64 arm64; do
  GO_BINARY="$GO_BINARY" bash scripts/bundle-app.sh linux "$arch"
  echo "Building Linux server ($arch)..."
  GOOS=linux GOARCH="$arch" "$GO_BINARY" build -trimpath -ldflags "$flags" -o "$stage/itsy-server-linux-$arch" ./cmd/itsy-server
done
bash scripts/bundle.sh "$stage"
FLEA_REQUIRE_BUNDLE=1 "$GO_BINARY" test ./internal/serverbundle -count=1
mkdir -p bin
echo 'Building ItsyWeb with both servers bundled...'
GO_BINARY="$GO_BINARY" bash scripts/bundle-app.sh "$("$GO_BINARY" env GOOS)" "$("$GO_BINARY" env GOARCH)"
FLEA_REQUIRE_APP_BUNDLE=1 "$GO_BINARY" test ./internal/appmanager ./internal/site -run 'TestManaged' -count=1
"$GO_BINARY" build -trimpath -ldflags "$flags" -o "$stage/itsy" ./cmd/itsy
"$stage/itsy" version
# A fresh inode avoids macOS retaining the previous executable's code signature.
# Rename within bin/ also leaves the last working executable intact until ready.
cli_stage=$(mktemp bin/.itsy.XXXXXX)
trap 'rm -f -- "$cli_stage"; rm -rf -- "$stage"' EXIT
cp "$stage/itsy" "$cli_stage"
chmod 755 "$cli_stage"
mv -f "$cli_stage" bin/itsy
ln -sfn itsy bin/flea
echo 'Built bin/itsy (with the retired flea command). Setup is ready without server downloads.'

#!/usr/bin/env bash
# Compile one matching-platform Contact bundle before compiling its parent.
set -euo pipefail
cd "$(dirname "$0")/.."
app_os=${1:?target OS required}
app_arch=${2:?target architecture required}
case "$app_os-$app_arch" in
  linux-amd64|linux-arm64|darwin-amd64|darwin-arm64) ;;
  *) echo 'Unsupported Contact target.' >&2; exit 1 ;;
esac
app_stage=$(mktemp -d)
trap 'rm -rf -- "$app_stage"' EXIT
CGO_ENABLED=0 GOOS="$app_os" GOARCH="$app_arch" "${GO_BINARY:-go}" build -trimpath -ldflags='-s -w' -o "$app_stage/contact" ./cmd/flea-contact
app_destination="internal/appbundle/assets/$app_os-$app_arch"
gzip -n -c "$app_stage/contact" > "$app_destination/contact.gz"
shasum -a 256 "$app_stage/contact" | awk '{print $1}' > "$app_destination/contact.sha256"

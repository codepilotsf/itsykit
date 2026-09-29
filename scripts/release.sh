#!/usr/bin/env bash
set -euo pipefail

kind=${1:?server or client required}
arch=${2:?architecture required}
source "$(dirname "$0")/version.sh"
version=$(flea_version release)
[[ "$arch" == amd64 || "$arch" == arm64 ]] || exit 1
[[ "$(go env GOHOSTARCH)" == "$arch" ]] || { echo 'Use a native runner for this architecture.' >&2; exit 1; }
export GOARCH="$arch"
unset GOOS
go test ./...
go test -race ./...
go vet ./...
mkdir -p dist
flags="-s -w -X flea/internal/buildinfo.Version=$version"
if [[ "$kind" == server ]]; then
  [[ "$(go env GOHOSTOS)" == linux ]] || exit 1
  file="dist/flea-server-linux-$arch"
  CGO_ENABLED=0 go build -trimpath -ldflags "$flags" -o "$file" ./cmd/flea-server
  "$file" -protocol
elif [[ "$kind" == client ]]; then
  system=$(go env GOHOSTOS)
  [[ "$system" == darwin || "$system" == linux ]] || exit 1
  bash scripts/bundle.sh dist
  FLEA_REQUIRE_BUNDLE=1 go test ./internal/serverbundle -count=1
  file="dist/flea-$system-$arch"
  CGO_ENABLED=0 go build -trimpath -ldflags "$flags" -o "$file" ./cmd/flea
  "$file" version
  smoke=$(mktemp -d)
  executable="$PWD/$file"
  (cd "$smoke"; "$executable" new demo --skip-git; cd demo; "$executable" status)
else
  echo 'Expected server or client.' >&2
  exit 1
fi

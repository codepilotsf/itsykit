#!/usr/bin/env bash
set -euo pipefail

kind=${1:?server or client required}
arch=${2:?architecture required}
version=${CI_COMMIT_TAG:?a release tag is required}
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]] || exit 1
[[ "$arch" == amd64 || "$arch" == arm64 ]] || exit 1
[[ "$(go env GOHOSTARCH)" == "$arch" ]] || { echo 'Use a native runner for this architecture.' >&2; exit 1; }
export GOARCH="$arch"
unset GOOS
go test ./...
go test -race ./...
go vet ./...
mkdir -p dist
flags="-s -w -X nanoweb/internal/buildinfo.Version=$version"
if [[ "$kind" == server ]]; then
  [[ "$(go env GOHOSTOS)" == linux ]] || exit 1
  file="dist/nanoweb-server-linux-$arch"
  CGO_ENABLED=0 go build -trimpath -ldflags "$flags" -o "$file" ./cmd/nanoweb-server
  "$file" -protocol
elif [[ "$kind" == client ]]; then
  [[ "$(go env GOHOSTOS)" == darwin ]] || exit 1
  bash scripts/bundle.sh dist
  NANOWEB_REQUIRE_BUNDLE=1 go test ./internal/serverbundle -count=1
  file="dist/nanoweb-darwin-$arch"
  CGO_ENABLED=0 go build -trimpath -ldflags "$flags" -o "$file" ./cmd/nanoweb
  "$file" version
  smoke=$(mktemp -d)
  executable="$PWD/$file"
  (cd "$smoke"; "$executable" new demo --skip-git; cd demo; "$executable" status)
else
  echo 'Expected server or client.' >&2
  exit 1
fi

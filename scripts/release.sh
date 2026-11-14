#!/usr/bin/env bash
set -euo pipefail

kind=${1:?server or client required}
arch=${2:?architecture required}
source "$(dirname "$0")/version.sh"
version=$(itsy_version release)
[[ "$arch" == amd64 || "$arch" == arm64 ]] || exit 1
[[ "$(go env GOHOSTARCH)" == "$arch" ]] || { echo 'Use a native runner for this architecture.' >&2; exit 1; }
if [[ "$kind" == client ]]; then
  case "$(go env GOHOSTOS)-$arch" in
    darwin-arm64|linux-amd64) ;;
    *) echo 'CLI releases support Apple Silicon macOS and x86-64 Linux.' >&2; exit 1 ;;
  esac
fi
export GOARCH="$arch"
unset GOOS
bash scripts/bundle-app.sh "$(go env GOHOSTOS)" "$arch"
export FLEA_REQUIRE_APP_BUNDLE=1
go test ./...
go test -race ./...
go vet ./...
mkdir -p dist
flags="-s -w -X itsy/internal/buildinfo.Version=$version"
if [[ "$kind" == server ]]; then
  [[ "$(go env GOHOSTOS)" == linux ]] || exit 1
  file="dist/flea-server-linux-$arch"
  CGO_ENABLED=0 go build -trimpath -ldflags "$flags" -o "$file" ./cmd/itsy-server
  "$file" -protocol
  if [[ "$arch" == amd64 ]]; then
    bash scripts/bundle-app.sh linux arm64
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$flags" -o dist/flea-server-linux-arm64 ./cmd/itsy-server
  fi
elif [[ "$kind" == client ]]; then
  system=$(go env GOHOSTOS)
  [[ "$system" == darwin || "$system" == linux ]] || exit 1
  bash scripts/bundle.sh dist
  FLEA_REQUIRE_BUNDLE=1 go test ./internal/serverbundle -count=1
  file="dist/flea-$system-$arch"
  CGO_ENABLED=0 go build -trimpath -ldflags "$flags" -o "$file" ./cmd/itsy
  smoke=$(mktemp -d)
  trap 'rm -rf -- "$smoke"' EXIT
  executable="$smoke/itsy"
  ln -s "$PWD/$file" "$executable"
  "$executable" version
  (cd "$smoke"; "$executable" new demo --skip-git; cd demo; "$executable" status)
else
  echo 'Expected server or client.' >&2
  exit 1
fi

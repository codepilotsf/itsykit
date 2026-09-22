#!/usr/bin/env bash
# Shared by local builds and native release jobs. Inputs must be from this build.
set -euo pipefail
source_dir=${1:?directory containing both Linux executables required}
destination="$(dirname "$0")/../internal/serverbundle/assets"
for arch in amd64 arm64; do
  test -s "$source_dir/flea-server-linux-$arch"
done
for arch in amd64 arm64; do
  name="flea-server-linux-$arch"
  gzip -n -c "$source_dir/$name" > "$destination/$name.gz"
  shasum -a 256 "$source_dir/$name" | awk '{print $1}' > "$destination/$name.sha256"
done

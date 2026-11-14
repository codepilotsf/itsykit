#!/usr/bin/env bash
# Prepare public files only. Never uploads or exposes repository source.
set -euo pipefail
source "$(dirname "$0")/version.sh"
version=$(itsy_version release)
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Only stable versions enter the public update channel.' >&2; exit 1; }
output="dist/public/releases/$version"
[[ ! -e "$output" ]] || { echo 'Public staging already exists; use a fresh build directory.' >&2; exit 1; }
files=(itsy-darwin-arm64 itsy-linux-amd64)
for file in "${files[@]}"; do
  [[ -f "dist/$file" && ! -L "dist/$file" ]] || { echo "Missing native-verified artifact: $file" >&2; exit 1; }
done
mkdir -p "$output"
for file in "${files[@]}"; do
  cp "dist/$file" "$output/$file"
  cp "dist/$file" "$output/${file/itsy-/flea-}"
done
(cd "$output"; shasum -a 256 "${files[@]}" flea-darwin-arm64 flea-linux-amd64 > SHA256SUMS)
cp website/site/install.sh dist/public/install.sh
printf '%s\n' "$version" > dist/public/releases/latest.txt
printf 'Prepared %s. Publish versioned files and verify them before latest.txt.\n' "$output"

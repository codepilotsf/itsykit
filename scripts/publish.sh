#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/version.sh"
version="v$(flea_version release)"
base="${CI_API_V4_URL:?}/projects/${CI_PROJECT_ID:?}/packages/generic/flea/$version"
files=(flea-server-linux-amd64 flea-server-linux-arm64 flea-darwin-amd64 flea-darwin-arm64)
for file in "${files[@]}"; do test -s "dist/$file"; done
for file in "${files[@]}"; do
  curl --fail --silent --show-error --header "JOB-TOKEN: ${CI_JOB_TOKEN:?}" --upload-file "dist/$file" "$base/$file"
done
# Users must be able to download without a GitLab account or token.
for file in "${files[@]}"; do
  expected=$(shasum -a 256 "dist/$file" | awk '{print $1}')
  actual=$(curl --fail --silent --show-error "$base/$file" | shasum -a 256 | awk '{print $1}')
  test "$actual" = "$expected" || { echo "Public verification failed for $file" >&2; exit 1; }
done

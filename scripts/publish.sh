#!/usr/bin/env bash
set -euo pipefail
version=${CI_COMMIT_TAG:?}
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]] || exit 1
base="${CI_API_V4_URL:?}/projects/${CI_PROJECT_ID:?}/packages/generic/nanoweb/$version"
files=(nanoweb-server-linux-amd64 nanoweb-server-linux-arm64 nanoweb-darwin-amd64 nanoweb-darwin-arm64)
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

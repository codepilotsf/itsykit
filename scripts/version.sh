#!/usr/bin/env bash
# Shared version validation for local builds, releases, and publishing.
flea_version() {
  local root recorded version numeric prerelease pattern
  root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
  recorded=$(sed -n 's/^var Version = "\([^"]*\)"$/\1/p' "$root/internal/buildinfo/version.go")
  numeric='(0|[1-9][0-9]*)'
  prerelease='(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)'
  pattern="^${numeric}\.${numeric}\.${numeric}(-${prerelease}(\.${prerelease})*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$"
  [[ "$recorded" =~ $pattern ]] || { echo 'The recorded version must be valid SemVer.' >&2; return 1; }
  if [[ "${1:-}" == release ]]; then
    [[ "${CI_COMMIT_TAG:-}" == "v$recorded" ]] || {
      echo "The release tag must match the recorded version: v$recorded" >&2
      return 1
    }
    version=$recorded
  else
    version=${VERSION:-$recorded}
  fi
  [[ "$version" =~ $pattern ]] || { echo 'VERSION must be valid SemVer, such as 0.1.0 or 0.2.0-rc.1.' >&2; return 1; }
  printf '%s\n' "$version"
}

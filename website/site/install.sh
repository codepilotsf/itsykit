#!/usr/bin/env bash
# Install the published CLI without sudo. Keep the whole installer inside a
# function so a truncated download cannot start an incomplete installation.
flea_install() (
  set -euo pipefail
  umask 077
  local system arch file version expected actual destination stage='' locked=false
  local base='https://fleacms.com/releases'
  fail() { printf 'FleaCMS: %s\n' "$1" >&2; exit 1; }
  command -v curl >/dev/null || fail 'Install curl, then run this command again.'
  case "$(uname -s)" in Darwin) system=darwin ;; Linux) system=linux ;; *) fail 'This installer supports macOS and Linux.' ;; esac
  case "$(uname -m)" in arm64|aarch64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) fail 'This processor is not supported.' ;; esac
  file="flea-$system-$arch"
  # The published checksum list decides which platforms have been verified.
  fetch() { curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' --tlsv1.2 --connect-timeout 15 --max-time 180 "$@"; }
  hash() {
    if command -v shasum >/dev/null; then shasum -a 256 "$1" | awk '{print $1}'
    elif command -v sha256sum >/dev/null; then sha256sum "$1" | awk '{print $1}'
    else fail 'A SHA-256 tool (shasum or sha256sum) is required.'; fi
  }
  version=$(fetch --max-filesize 128 "$base/latest.txt") || fail 'No release is available. Your installation was not changed.'
  [[ ${#version} -le 64 && "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || fail 'Invalid release version.'
  destination=${FLEA_INSTALL_DIR:-"$HOME/.local/bin"}
  [[ "$destination" == /* ]] || fail 'FLEA_INSTALL_DIR must be an absolute directory.'
  mkdir -p "$destination" || fail 'Choose a writable FLEA_INSTALL_DIR. No sudo is needed.'
  [[ ! -L "$destination/flea" ]] || fail 'The existing flea is a symlink. Use flea update or choose another directory.'
  [[ ! -e "$destination/flea" || -f "$destination/flea" ]] || fail 'The installation target is not a regular file.'
  mkdir "$destination/.flea-install.lock" 2>/dev/null || fail "Installation is busy or not writable. Check $destination/.flea-install.lock."
  locked=true
  cleanup() {
    if [[ -n "$stage" ]]; then
      rm -f -- "$stage/flea" "$stage/SHA256SUMS" "$stage/version"
      rmdir "$stage" 2>/dev/null || true
    fi
    if [[ "$locked" == true ]]; then rmdir "$destination/.flea-install.lock" 2>/dev/null || true; fi
  }
  trap cleanup EXIT
  trap 'exit 130' INT
  trap 'exit 143' TERM
  stage=$(mktemp -d "$destination/.flea-download.XXXXXXXX") || fail 'Could not stage the download.'
  fetch --max-filesize 8192 --output "$stage/SHA256SUMS" "$base/$version/SHA256SUMS" || fail 'Could not retrieve release checksums.'
  [[ $(wc -c < "$stage/SHA256SUMS") -le 8192 ]] || fail 'Release checksum list is too large.'
  expected=$(awk -v name="$file" '$2 == name && NF == 2 {print $1}' "$stage/SHA256SUMS")
  [[ "$expected" =~ ^[0-9a-f]{64}$ ]] || fail 'This release does not include a verified download for your platform.'
  printf 'Downloading FleaCMS %s…\n' "$version"
  fetch --max-filesize 134217728 --output "$stage/flea" "$base/$version/$file" || fail 'Download failed. Your existing CLI was not changed.'
  [[ $(wc -c < "$stage/flea") -le 134217728 ]] || fail 'Download is too large.'
  actual=$(hash "$stage/flea")
  [[ "$actual" == "$expected" ]] || fail 'Checksum did not match. Your existing CLI was not changed.'
  chmod 755 "$stage/flea"
  "$stage/flea" version > "$stage/version" || fail 'The downloaded CLI cannot run on this machine.'
  [[ $(cat "$stage/version") == "FleaCMS $version" ]] || fail 'The downloaded CLI has an unexpected version.'
  mv -f -- "$stage/flea" "$destination/flea"
  printf '\nInstalled FleaCMS %s in %s/flea\n' "$version" "$destination"
  case ":$PATH:" in
    *":$destination:"*) printf '\nGet started: flea new my-site\n' ;;
    *) printf '\nAdd this directory to PATH in your shell profile, then open a new terminal:\n  export PATH=%q:"$PATH"\n' "$destination" ;;
  esac
)
flea_install

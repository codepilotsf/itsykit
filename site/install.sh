#!/usr/bin/env bash
# Install the published CLI without sudo. Keep the whole installer inside a
# function so a truncated download cannot start an incomplete installation.
itsy_install() (
  set -euo pipefail
  umask 077
  local system arch file version expected actual destination stage='' locked=false
  local base='https://itsyweb.dev/releases'
  fail() { printf 'ItsyWeb: %s\n' "$1" >&2; exit 1; }
  command -v curl >/dev/null || fail 'Install curl, then run this command again.'
  case "$(uname -s)" in Darwin) system=darwin ;; Linux) system=linux ;; *) fail 'This installer supports macOS and Linux.' ;; esac
  case "$(uname -m)" in arm64|aarch64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) fail 'This processor is not supported.' ;; esac
  case "$system-$arch" in
    darwin-arm64|linux-amd64) ;;
    linux-arm64) fail 'Linux ARM64 CLI downloads are no longer supported. Use Apple Silicon macOS or x86-64 Linux.' ;;
    *) fail 'CLI downloads support Apple Silicon macOS and x86-64 Linux.' ;;
  esac
  file="itsy-$system-$arch"
  # The published checksum list decides which platforms have been verified.
  fetch() { curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' --tlsv1.2 --connect-timeout 15 --max-time 180 "$@"; }
  hash() {
    if command -v shasum >/dev/null; then shasum -a 256 "$1" | awk '{print $1}'
    elif command -v sha256sum >/dev/null; then sha256sum "$1" | awk '{print $1}'
    else fail 'A SHA-256 tool (shasum or sha256sum) is required.'; fi
  }
  version=$(fetch --max-filesize 128 "$base/latest.txt") || fail 'No release is available. Your installation was not changed.'
  [[ ${#version} -le 64 && "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || fail 'Invalid release version.'
  destination=${ITSY_INSTALL_DIR:-"$HOME/.local/bin"}
  [[ "$destination" == /* ]] || fail 'ITSY_INSTALL_DIR must be an absolute directory.'
  mkdir -p "$destination" || fail 'Choose a writable ITSY_INSTALL_DIR. No sudo is needed.'
  [[ ! -L "$destination/itsy" ]] || fail 'The existing itsy is a symlink. Use itsy update or choose another directory.'
  [[ ! -e "$destination/itsy" || -f "$destination/itsy" ]] || fail 'The installation target is not a regular file.'
  mkdir "$destination/.itsy-install.lock" 2>/dev/null || fail "Installation is busy or not writable. Check $destination/.itsy-install.lock."
  locked=true
  cleanup() {
    if [[ -n "$stage" ]]; then
      rm -f -- "$stage/itsy" "$stage/SHA256SUMS" "$stage/version"
      rmdir "$stage" 2>/dev/null || true
    fi
    if [[ "$locked" == true ]]; then rmdir "$destination/.itsy-install.lock" 2>/dev/null || true; fi
  }
  trap cleanup EXIT
  trap 'exit 130' INT
  trap 'exit 143' TERM
  stage=$(mktemp -d "$destination/.itsy-download.XXXXXXXX") || fail 'Could not stage the download.'
  fetch --max-filesize 8192 --output "$stage/SHA256SUMS" "$base/$version/SHA256SUMS" || fail 'Could not retrieve release checksums.'
  [[ $(wc -c < "$stage/SHA256SUMS") -le 8192 ]] || fail 'Release checksum list is too large.'
  expected=$(awk -v name="$file" '$2 == name && NF == 2 {print $1}' "$stage/SHA256SUMS")
  [[ "$expected" =~ ^[0-9a-f]{64}$ ]] || fail "ItsyWeb $version has no verified download for $system/$arch. See https://itsyweb.dev/docs for available platforms."
  printf 'Downloading ItsyWeb %s…\n' "$version"
  fetch --max-filesize 134217728 --output "$stage/itsy" "$base/$version/$file" || fail 'Download failed. Your existing CLI was not changed.'
  [[ $(wc -c < "$stage/itsy") -le 134217728 ]] || fail 'Download is too large.'
  actual=$(hash "$stage/itsy")
  [[ "$actual" == "$expected" ]] || fail 'Checksum did not match. Your existing CLI was not changed.'
  chmod 755 "$stage/itsy"
  "$stage/itsy" version > "$stage/version" || fail 'The downloaded CLI cannot run on this machine.'
  [[ $(cat "$stage/version") == "ItsyWeb $version" ]] || fail 'The downloaded CLI has an unexpected version.'
  mv -f -- "$stage/itsy" "$destination/itsy"
  printf '\nInstalled ItsyWeb %s in %s/itsy\n' "$version" "$destination"
  case ":$PATH:" in
    *":$destination:"*) printf '\nGet started: itsy new my-site\n' ;;
    *) printf '\nAdd this directory to PATH in your shell profile, then open a new terminal:\n  export PATH=%q:"$PATH"\n' "$destination" ;;
  esac
)
itsy_install

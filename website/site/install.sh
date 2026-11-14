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
    if command -v shasum >/dev/null; then shasum -a 256 < "$1" | awk '{print $1}'
    elif command -v sha256sum >/dev/null; then sha256sum < "$1" | awk '{print $1}'
    else fail 'A SHA-256 tool (shasum or sha256sum) is required.'; fi
  }
  version=$(fetch --max-filesize 128 "$base/latest.txt") || fail 'No release is available. Your installation was not changed.'
  [[ ${#version} -le 64 && "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || fail 'Invalid release version.'
  destination=${ITSY_INSTALL_DIR:-${FLEA_INSTALL_DIR:-"$HOME/.local/bin"}}
  [[ "$destination" == /* ]] || fail 'ITSY_INSTALL_DIR must be an absolute directory.'
  [[ "$destination" != *:* && "$destination" != *$'\n'* ]] || fail 'ITSY_INSTALL_DIR cannot contain a colon or newline because it must be added to PATH.'
  mkdir -p "$destination" || fail 'Choose a writable ITSY_INSTALL_DIR. No sudo is needed.'
  [[ ! -L "$destination/itsy" ]] || fail 'The existing itsy is a symlink. Use itsy update or choose another directory.'
  [[ ! -e "$destination/itsy" || -f "$destination/itsy" ]] || fail 'The itsy installation target is not a regular file.'
  if [[ -L "$destination/flea" ]]; then
    [[ $(readlink "$destination/flea") == itsy ]] || fail 'The existing flea link belongs to another installation. Choose another directory.'
  fi
  [[ ! -e "$destination/flea" || -f "$destination/flea" ]] || fail 'The installation target is not a regular file.'
  mkdir "$destination/.flea-install.lock" 2>/dev/null || fail "Installation is busy or not writable. Check $destination/.flea-install.lock."
  locked=true
  cleanup() {
    if [[ -n "$stage" ]]; then
      rm -f -- "$stage/itsy" "$stage/SHA256SUMS" "$stage/version" "$stage/flea"
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
  [[ "$expected" =~ ^[0-9a-f]{64}$ ]] || fail "ItsyWeb $version has no verified download for $system/$arch. See https://itsyweb.dev/docs for available platforms."
  printf 'Downloading ItsyWeb %s…\n' "$version"
  fetch --max-filesize 134217728 --output "$stage/itsy" "$base/$version/$file" || fail 'Download failed. Your existing CLI was not changed.'
  [[ $(wc -c < "$stage/itsy") -le 134217728 ]] || fail 'Download is too large.'
  actual=$(hash "$stage/itsy")
  [[ "$actual" == "$expected" ]] || fail 'Checksum did not match. Your existing CLI was not changed.'
  chmod 755 "$stage/itsy"
  "$stage/itsy" version > "$stage/version" || fail 'The downloaded CLI cannot run on this machine.'
  [[ $(cat "$stage/version") == "ItsyWeb $version" ]] || fail 'The downloaded CLI has an unexpected version.'
  ln -s itsy "$stage/flea" || fail 'Could not prepare the retired flea command.'
  mv -f -- "$stage/itsy" "$destination/itsy"
  mv -f -- "$stage/flea" "$destination/flea" || fail 'itsy is installed, but the retired flea command could not be replaced.'
  printf '\nInstalled ItsyWeb %s in %s/itsy\n' "$version" "$destination"

  # This Bash child cannot change the invoking shell's PATH. These guarded
  # entries configure future terminals; print an optional current-shell export.
  local profile login_profile quoted block marker
  local profiles=()
  path_failure() {
    printf '\nItsyWeb is installed, but %s\n' "$1" >&2
    printf 'Fix the startup file and rerun the installer, or add %s to PATH in your shell.\n' "$destination" >&2
    exit 1
  }
  case "${SHELL##*/}" in
    bash)
      # Bash reads only the first existing login profile in this order.
      login_profile="$HOME/.profile"
      for profile in "$HOME/.bash_profile" "$HOME/.bash_login" "$HOME/.profile"; do
        if [[ -e "$profile" || -L "$profile" ]]; then login_profile="$profile"; break; fi
      done
      profiles=("$HOME/.bashrc" "$login_profile")
      ;;
    zsh)
      profiles=("${ZDOTDIR:-$HOME}/.zshrc" "${ZDOTDIR:-$HOME}/.zprofile")
      ;;
    *) path_failure 'automatic PATH setup supports Bash and Zsh. Open one of those shells and rerun with SHELL set to its absolute path, or configure your shell manually.' ;;
  esac
  # POSIX single-quoted literals also work in .profile and safely preserve
  # spaces, quotes, dollar signs and other shell metacharacters in custom paths.
  quoted="'${destination//\'/\'\\\'\'}'"
  marker="# ItsyWeb PATH: $quoted"
  block=$(printf '%s\ncase ":${PATH-}:" in\n  *:%s:*) ;;\n  *) export PATH=%s:"${PATH-}" ;;\nesac' "$marker" "$quoted" "$quoted")
  # Preflight all profiles before appending to any of them. Follow regular
  # symlink targets without replacing links or rewriting existing contents.
  for profile in "${profiles[@]}"; do
    [[ "$profile" == /* ]] || path_failure "the startup file path must be absolute: $profile."
    if [[ -e "$profile" || -L "$profile" ]]; then
      [[ -f "$profile" && -r "$profile" && -w "$profile" ]] || path_failure "the startup file is not a readable, writable regular file: $profile."
    else
      [[ -d "${profile%/*}" && -w "${profile%/*}" ]] || path_failure "the startup directory does not exist or is not writable: ${profile%/*}."
    fi
  done
  for profile in "${profiles[@]}"; do
    if ! grep -Fqx -- "$marker" "$profile" 2>/dev/null; then
      printf '\n%s\n' "$block" >> "$profile" || path_failure "the startup file could not be updated: $profile."
      printf 'Added PATH setup to %s\n' "$profile"
    fi
  done
  printf 'PATH is configured for future %s terminals.\n' "${SHELL##*/}"
  case ":$PATH:" in
    *":$destination:"*) printf '\nGet started: itsy new my-site\n' ;;
    *) printf '\nOpen a new terminal, then run: itsy new my-site\n\nTo keep using this terminal instead, run:\n  export PATH=%s:"$PATH"\n' "$quoted" ;;
  esac
)
itsy_install

#!/usr/bin/env bash
# Package the native CLI. Installation/notarization are separate explicit steps.
set -euo pipefail
export COPYFILE_DISABLE=1
cd "$(dirname "$0")/.."
if [[ "${1:-}" == --help ]]; then
  printf 'Usage: bash scripts/package-macos.sh\n\nPackages bin/flea for /usr/local/bin. Set INSTALLER_IDENTITY to sign.\nAn unsigned development package is not ready for public distribution.\n\n'
  exit 0
fi
[[ $(uname -s) == Darwin ]] || { echo 'Build Mac installers on macOS.' >&2; exit 1; }
source scripts/version.sh
version=$(flea_version)
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Installer packages require a stable numeric version.' >&2; exit 1; }
test -x bin/flea
[[ $(bin/flea version) == "FleaCMS $version" ]] || { echo 'Rebuild bin/flea for this version first.' >&2; exit 1; }
stage=$(mktemp -d /tmp/flea-package.XXXXXX)
trap 'rm -rf -- "$stage"' EXIT
mkdir -p "$stage/root/usr/local/bin" dist
cp -X bin/flea "$stage/root/usr/local/bin/flea"
chmod 755 "$stage/root/usr/local/bin/flea"
args=()
if [[ -n "${APPLICATION_IDENTITY:-}" ]]; then
  codesign --force --options runtime --timestamp --sign "$APPLICATION_IDENTITY" "$stage/root/usr/local/bin/flea"
fi
if [[ -n "${INSTALLER_IDENTITY:-}" ]]; then args+=(--sign "$INSTALLER_IDENTITY"); fi
pkgbuild --root "$stage/root" --filter '\._' --identifier com.fleacms.cli --version "$version" --install-location / ${args[@]+"${args[@]}"} "dist/flea-$version-$(uname -m).pkg"
echo 'Package built, not installed or notarized. Verify signing and notarize before public release.'

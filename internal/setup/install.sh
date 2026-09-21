#!/bin/sh
set -eu

stage=$1
case "$stage" in /tmp/nanoweb-setup.*) ;; *) echo "Invalid staging directory" >&2; exit 1;; esac
test "$(id -u)" = 0 || { echo "Installation requires root or sudo." >&2; exit 1; }
. /etc/os-release
test "$ID" = ubuntu || { echo "Ubuntu is required." >&2; exit 1; }
case "$VERSION_ID" in 22.04|24.04|26.04) ;; *) echo "Unsupported Ubuntu release." >&2; exit 1;; esac
test -d /run/systemd/system
command -v ss >/dev/null || { echo "The ss command (iproute2) is required for the port check." >&2; exit 1; }
if test -n "$(ss -H -ltn 'sport = :80')"; then
  echo "Port 80 is already in use; no changes made." >&2
  exit 1
fi

# Setup is deliberately a first-install operation. Never replace existing data,
# binaries, configuration, accounts, or administrator-managed service units.
for path in /usr/local/bin/nanoweb-server /etc/nanoweb /etc/systemd/system/nanoweb.service /var/lib/nanoweb; do
  if test -e "$path" || test -L "$path"; then
    echo "Refusing to replace existing $path. Setup is for a fresh installation." >&2
    exit 1
  fi
done
if getent passwd nanoweb >/dev/null || getent group nanoweb >/dev/null; then
  echo "The nanoweb account or group already exists; refusing to repurpose it." >&2
  exit 1
fi
if systemctl cat nanoweb.service >/dev/null 2>&1; then
  echo "A nanoweb.service unit already exists." >&2
  exit 1
fi
for file in nanoweb-server server.json nanoweb.service; do
  test -f "$stage/$file" && test ! -L "$stage/$file"
done

installed_binary=0
installed_config=0
installed_unit=0
success=0
cleanup() {
  if test "$success" = 0; then
    if test "$installed_unit" = 1; then
      systemctl disable --now nanoweb.service >/dev/null 2>&1 || true
      rm -f /etc/systemd/system/nanoweb.service
      systemctl daemon-reload || true
    fi
    if test "$installed_binary" = 1; then rm -f /usr/local/bin/nanoweb-server; fi
    if test "$installed_config" = 1; then rm -f /etc/nanoweb/server.json; rmdir /etc/nanoweb || true; fi
    echo "Install failed. Any created nanoweb account and /var/lib/nanoweb were retained for inspection." >&2
  fi
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

useradd --system --user-group --home-dir /var/lib/nanoweb --no-create-home --shell /usr/sbin/nologin nanoweb
install -d -o nanoweb -g nanoweb -m 0750 /var/lib/nanoweb
install -d -m 0755 /etc/nanoweb
installed_config=1
install -m 0644 "$stage/server.json" /etc/nanoweb/server.json
installed_binary=1
install -m 0755 "$stage/nanoweb-server" /usr/local/bin/nanoweb-server
installed_unit=1
install -m 0644 "$stage/nanoweb.service" /etc/systemd/system/nanoweb.service
systemctl daemon-reload
systemctl enable --now nanoweb.service

attempt=0
until /usr/local/bin/nanoweb-server -config /etc/nanoweb/server.json -check >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if test "$attempt" -ge 10; then
    systemctl status nanoweb.service --no-pager >&2 || true
    exit 1
  fi
  sleep 1
done
systemctl is-active --quiet nanoweb.service
success=1
echo "Nanoweb service installed and ready."

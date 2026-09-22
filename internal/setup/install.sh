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
# Check every IPv4/IPv6 TCP and UDP listener, including HTTPS/QUIC. Only the
# systemd main process with the installed NanoWeb executable may be exempted.
check_web_ports() {
  allowed_pid=0
  if test "${1:-}" = --reinstall; then
    allowed_pid=$(systemctl show nanoweb.service --property=MainPID --value) || return 1
    case "$allowed_pid" in ''|*[!0-9]*) allowed_pid=0;; esac
    if test "$allowed_pid" != 0; then
      executable=$(readlink "/proc/$allowed_pid/exe") || return 1
      if test "$executable" != /usr/local/bin/nanoweb-server; then allowed_pid=0; fi
    fi
  fi
  for protocol in t u; do
    listeners=$(ss -H -ln${protocol}p '( sport = :80 or sport = :443 )') || {
      echo "Could not check ports 80 and 443. Setup stopped." >&2
      return 1
    }
    test -n "$listeners" || continue
    printf '%s\n' "$listeners" | while IFS= read -r listener; do
      pids=$(printf '%s\n' "$listener" | grep -o 'pid=[0-9]*' | cut -d= -f2)
      if test -z "$pids"; then
        echo "Ports 80 or 443 are in use, but their owner could not be verified. Setup stopped." >&2
        exit 1
      fi
      for pid in $pids; do
        if test "$allowed_pid" = 0 || test "$pid" != "$allowed_pid"; then
          echo "Another service is using port 80 or 443. Stop it before running NanoWeb setup." >&2
          exit 1
        fi
      done
    done || return 1
  done
}

# This must precede backups, service stops, and any installation changes.
check_web_ports "${2:-}"
# Reinstallation is opt-in and only applies to a recognizable installation.
# Preserve the old files rather than recursively deleting site data.
reuse_account=0
if test "${2:-}" = --reinstall; then
  for path in /usr/local/bin/nanoweb-server /etc/nanoweb /etc/systemd/system/nanoweb.service /var/lib/nanoweb; do
    test ! -L "$path" || { echo "Refusing to replace a symlink: $path" >&2; exit 1; }
  done
  test -f /usr/local/bin/nanoweb-server
  test -f /etc/nanoweb/server.json
  test -f /etc/systemd/system/nanoweb.service
  for file in nanoweb-server server.json nanoweb.service; do
    test -f "$stage/$file" && test ! -L "$stage/$file"
  done
  getent passwd nanoweb >/dev/null
  getent group nanoweb >/dev/null
  backup=$(mktemp -d /var/backups/nanoweb.XXXXXXXXXX)
  chmod 0700 "$backup"
  echo "Previous installation backup: $backup"
  systemctl disable --now nanoweb.service
  mv /usr/local/bin/nanoweb-server "$backup/nanoweb-server"
  mv /etc/nanoweb "$backup/config"
  mv /etc/systemd/system/nanoweb.service "$backup/nanoweb.service"
  if test -e /var/lib/nanoweb; then mv /var/lib/nanoweb "$backup/data"; fi
  systemctl daemon-reload
  reuse_account=1
fi
# Recheck after stopping NanoWeb, before installing replacement files.
check_web_ports

# Setup is deliberately a first-install operation. Never replace existing data,
# binaries, configuration, accounts, or administrator-managed service units.
for path in /usr/local/bin/nanoweb-server /etc/nanoweb /etc/systemd/system/nanoweb.service /var/lib/nanoweb; do
  if test -e "$path" || test -L "$path"; then
    echo "Refusing to replace existing $path. Setup is for a fresh installation." >&2
    exit 1
  fi
done
if test "$reuse_account" = 0 && { getent passwd nanoweb >/dev/null || getent group nanoweb >/dev/null; }; then
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
    if test "$installed_config" = 1; then rm -f /etc/nanoweb/server.json /etc/nanoweb/deploy.lock; rmdir /etc/nanoweb || true; fi
    echo "Install failed. Any created nanoweb account and /var/lib/nanoweb were retained for inspection." >&2
  fi
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

if test "$reuse_account" = 0; then
  useradd --system --user-group --home-dir /var/lib/nanoweb --no-create-home --shell /usr/sbin/nologin nanoweb
fi
install -d -o root -g root -m 0755 /var/lib/nanoweb
install -d -o root -g root -m 0755 /var/lib/nanoweb/sites
install -d -o nanoweb -g nanoweb -m 0700 /var/lib/nanoweb/certs
install -d -m 0755 /etc/nanoweb
installed_config=1
install -o nanoweb -g nanoweb -m 0600 /dev/null /etc/nanoweb/deploy.lock
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

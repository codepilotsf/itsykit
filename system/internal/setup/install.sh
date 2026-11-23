#!/bin/sh
set -eu

stage=$1
case "$stage" in /tmp/flea-setup.*) ;; *) echo "Invalid staging directory" >&2; exit 1;; esac
test "$(id -u)" = 0 || { echo "Installation requires root or sudo." >&2; exit 1; }
. /etc/os-release
test "$ID" = ubuntu || { echo "Ubuntu is required." >&2; exit 1; }
case "$VERSION_ID" in 22.04|24.04|26.04) ;; *) echo "Unsupported Ubuntu release." >&2; exit 1;; esac
test -d /run/systemd/system
command -v ss >/dev/null || { echo "The ss command (iproute2) is required for the port check." >&2; exit 1; }
# Check every IPv4/IPv6 TCP and UDP listener, including HTTPS/QUIC. Only the
# systemd main process with the installed ItsyKit executable may be exempted.
check_web_ports() {
  allowed_pid=0
  if test "${1:-}" = --reinstall; then
    allowed_pid=$(systemctl show flea.service --property=MainPID --value) || return 1
    case "$allowed_pid" in ''|*[!0-9]*) allowed_pid=0;; esac
    if test "$allowed_pid" != 0; then
      executable=$(readlink "/proc/$allowed_pid/exe") || return 1
      if test "$executable" != /usr/local/bin/flea-server; then allowed_pid=0; fi
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
          echo "Another service is using port 80 or 443. Stop it before running ItsyKit setup." >&2
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
  for path in /usr/local/bin/flea-server /etc/flea /etc/systemd/system/flea.service /var/lib/flea; do
    test ! -L "$path" || { echo "Refusing to replace a symlink: $path" >&2; exit 1; }
  done
  test -f /usr/local/bin/flea-server
  test -f /etc/flea/server.json
  test -f /etc/systemd/system/flea.service
  for file in flea-server server.json flea.service; do
    test -f "$stage/$file" && test ! -L "$stage/$file"
  done
  getent passwd flea >/dev/null
  getent group flea >/dev/null
  backup=$(mktemp -d /var/backups/flea.XXXXXXXXXX)
  chmod 0700 "$backup"
  echo "Previous installation backup: $backup"
  systemctl disable --now flea.service
  mv /usr/local/bin/flea-server "$backup/flea-server"
  mv /etc/flea "$backup/config"
  mv /etc/systemd/system/flea.service "$backup/flea.service"
  if test -e /var/lib/flea; then mv /var/lib/flea "$backup/data"; fi
  systemctl daemon-reload
  reuse_account=1
fi
# Recheck after stopping ItsyKit, before installing replacement files.
check_web_ports

# Setup is deliberately a first-install operation. Never replace existing data,
# binaries, configuration, accounts, or administrator-managed service units.
for path in /usr/local/bin/flea-server /etc/flea /etc/systemd/system/flea.service /var/lib/flea; do
  if test -e "$path" || test -L "$path"; then
    echo "Refusing to replace existing $path. Setup is for a fresh installation." >&2
    exit 1
  fi
done
if test "$reuse_account" = 0 && { getent passwd flea >/dev/null || getent group flea >/dev/null; }; then
  echo "The flea account or group already exists; refusing to repurpose it." >&2
  exit 1
fi
if systemctl cat flea.service >/dev/null 2>&1; then
  echo "A flea.service unit already exists." >&2
  exit 1
fi
for file in flea-server server.json flea.service; do
  test -f "$stage/$file" && test ! -L "$stage/$file"
done

installed_binary=0
installed_config=0
installed_unit=0
success=0
cleanup() {
  if test "$success" = 0; then
    if test "$installed_unit" = 1; then
      systemctl disable --now flea.service >/dev/null 2>&1 || true
      rm -f /etc/systemd/system/flea.service
      systemctl daemon-reload || true
    fi
    if test "$installed_binary" = 1; then rm -f /usr/local/bin/flea-server; fi
    if test "$installed_config" = 1; then rm -f /etc/flea/server.json /etc/flea/deploy.lock; rmdir /etc/flea || true; fi
    echo "Install failed. Any created flea account and /var/lib/flea were retained for inspection." >&2
  fi
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

if test "$reuse_account" = 0; then
  useradd --system --user-group --home-dir /var/lib/flea --no-create-home --shell /usr/sbin/nologin flea
fi
install -d -o root -g root -m 0755 /var/lib/flea
install -d -o root -g root -m 0755 /var/lib/flea/sites
install -d -o flea -g flea -m 0700 /var/lib/flea/certs
install -d -m 0755 /etc/flea
installed_config=1
install -o flea -g flea -m 0600 /dev/null /etc/flea/deploy.lock
install -m 0644 "$stage/server.json" /etc/flea/server.json
installed_binary=1
install -m 0755 "$stage/flea-server" /usr/local/bin/flea-server
installed_unit=1
install -m 0644 "$stage/flea.service" /etc/systemd/system/flea.service
systemctl daemon-reload
systemctl enable --now flea.service

attempt=0
until /usr/local/bin/flea-server -config /etc/flea/server.json -check >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if test "$attempt" -ge 10; then
    systemctl status flea.service --no-pager >&2 || true
    exit 1
  fi
  sleep 1
done
systemctl is-active --quiet flea.service
success=1
echo "ItsyKit service installed and ready."

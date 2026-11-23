# Socket activation verification

Private rollout on November 14–15, 2026 UTC, explicitly authorized by the owner.
Runtime source is `9dd6076` on `t3/zero-downtime-itsyweb`. The shared server at
`root@fleacms.com` now runs `0.22.0-sockets.9dd6076`, deployment protocol 30.
This is a private server build. Public releases, tags, downloads, update channels
and customer site source were not published or changed by this rollout.

## Artifacts and local checks

The implementation passed `go test ./...`, `go test -race ./...`, `go vet ./...`
and the bundled build. The rollout reused that source evidence and built the
exact candidate with Go 1.26.9 on native Linux AMD64:

```sh
VERSION=0.22.0-sockets.9dd6076 bash scripts/build.sh
```

The fresh starter `new --starter --skip-git` and `status` smoke passed. Both Linux
servers were refreshed before the CLI; native ARM64 acceptance is not claimed.

- Installed server SHA-256:
  `cc1b153aa71124b485803a8e548324ad0c7614f0b4b1376d113e1089ac7d5194`.
- Previous public 0.22.0 server SHA-256:
  `a77248fca27c14263da31c92263bd6fc55a334830b8d74d4ebb063c4d04103ff`.

## Disposable native systemd checks

An isolated Ubuntu 24.04 AMD64 container ran systemd 255 and the actual production
units and server artifacts, without publishing host ports. Checks passed for:

- Fresh installation and confirmed reinstallation while systemd retained 80/443.
- Automatic standard-unit conversion, injected restart failure rollback, and
  recovery after forcibly interrupting the upgrader with its journal pending.
- Verified HTTPS using a test CA installed only inside the disposable instance.
  HTTP and HTTPS hostname/revision probes passed after conversion and recovery.
- The bundled Contact child under the hardened service; the parent had no
  effective capabilities. No real email was sent.
- Refusal of the actual customized production unit before executable mutation.
- The reviewed one-time custom-unit procedure: successful activation, automatic
  rollback after an injected failure, and explicit recovery after interruption.
  Every recovery restored the exact old executable, unit and memory override.

The generated fixture initially included private local `status` runtime, which
real release archives exclude; that fixture-only runtime was removed. Repeated
fault cases also reached systemd's start-rate limit; failure counters were reset
between independent cases without changing service policy or permission checks.
This verifies native container/systemd behavior, not a full SSH installation on
every supported Ubuntu release. The temporary container and test image were removed.

## Shared-host conversion

Preflight found 22 sites, 34 HTTPS hostnames, active swap, sufficient disk and
memory headroom, and the existing `60-memory.conf` with `GOMEMLIMIT=320MiB`.
The base unit still had the historical FleaCMS description. The automatic
conversion correctly requires manual review for this host.

Concurrent website deployments changed the initial baseline. The first conversion
attempt could not acquire `deploy.lock` and made no changes. After that deployment
finished, a fresh baseline was captured under the same lock used by deployments
and browser saves. The successful conversion held it throughout validation,
backup, the journaled swap, service transition and verification. The pinned
candidate validated all installed sites as the actual `flea` user before downtime.

At 23:56 UTC on November 14, the reviewed procedure installed the canonical
service/socket units and exact candidate executable. The memory drop-in was
preserved byte-for-byte. Both units are active and enabled; only systemd retains
the listening ports while the server restarts. The parent has an empty capability
bounding set and zero effective capabilities, with nine managed child processes.

All 34 hostname/revision probes passed with normal TLS verification. All 34 public
HTTPS homepages retained their baseline status and redirect destinations: 28
returned 200, one returned 301 and five returned 308. Routing, site inventory,
release identities and all 22 active content snapshot hashes matched the locked
cutover baseline. The configuration SHA-256 before and after was:

`07f52c7322df24783b853252327dc8cc2fad9ed2d8837ab3f7be6d100921aa50`.

## Controlled live restart

A further `flea.service` restart ran under `deploy.lock`, with no configuration
mutation. Six new HTTP/HTTPS connections to ItsyWeb, Cubitalk and Atlas used both
IPv4 and IPv6 loopback. TCP connected in approximately 0.1–10 ms. Every request
waited through startup and returned the expected 200 response and hostname/revision
after approximately 11.6 seconds; HTTPS used ordinary certificate verification.

The server PID changed from 774345 to 774817. The kernel listener identities stayed
unchanged: HTTP inode 6942001 and HTTPS inode 6942005. Afterwards all 34 TLS/hostname
probes passed again; configuration and content hashes still matched and neither
deployment nor executable recovery was pending. This prevents connection refusal
during normal restarts, but clients with shorter deadlines can still time out.

The root-only backup is `/var/backups/flea-socket-activation-9dd6076/`. It retains
the previous executable, exact old/new units, memory override, configuration copy,
manifest, reviewed conversion/recovery script and `restart-proof.json`. The normal
`/etc/flea/server.binary.previous` also holds the old executable. Saved configuration
must not be restored over later deployments. Temporary conversion units and remote
staging were removed; no extra persistent application process was introduced.

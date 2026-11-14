# Multi-site VPS capacity check

October 4, 2026, 16:30:45–16:31:41 UTC. Authorized isolated test alongside the
existing Flea service on `fleacms.com`; no production configuration or binaries
were replaced.

## Environment

- Linux x86-64, one vCPU, 956 MiB physical RAM, 1 GiB configured host swap.
- Before testing: 706 MiB available RAM; live service active, approximately
  69 MiB cgroup memory, PID 217528, zero service restarts.
- Test ran as nobody/nogroup in a private network namespace with loopback HTTP,
  70% CPU quota, 384 MiB memory cap, no test swap, 4,096 descriptor limit,
  64-task limit, and 180-second maximum lifetime.
- Disk-backed `/var/tmp` fixtures; no changes to the live site's content,
  configuration, credentials, certificates, process limits, or installed binary.
- Test server built from base commit `27b10b4` plus the uncommitted disk-backed
  asset changes, using Go 1.27.1 on macOS ARM64, cross-compiled for Linux AMD64.
  This was not the installed production executable.

Server SHA-256:
`13733f56afa9f744ac1d4c4b36a73c9e584749560a73a2327e6c7e2f5bf29196`

Probe SHA-256:
`010bbf638637c7aad47eee8ade98df09953649153b4b98d5468acac9a5751631`

## Results

Each site had five pages, JSON-driven homepage content, and 20 MiB of separate
regular asset files. Each phase requested 900 responses at a nominal 60/s:
80% HTML and 20% full 1 MiB asset responses, rotating between sites. A second
read-only validation process loaded the same site set during traffic.

| Sites | Asset files | Startup | Idle server RSS | Peak server RSS | Peak server FDs | p95 response | p99 response |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | 20 MiB | 230 ms | 14.52 MiB | 17.37 MiB | 29 | 2.054 ms | 3.526 ms |
| 10 | 200 MiB | 1,498 ms | 16.24 MiB | 17.73 MiB | 208 | 2.137 ms | 4.365 ms |
| 30 | 600 MiB | 4,259 ms | 19.37 MiB | 20.84 MiB | 609 | 4.130 ms | 25.719 ms |

All three phases completed 900 requests with **zero failures and zero skipped
requests**. Median response times were 0.824, 0.820, and 0.817 ms respectively.

| Sites | Overlapping validation duration | Peak combined server + validator RSS |
| --- | --- | --- |
| 1 | 148 ms | 29.71 MiB |
| 10 | 1,296 ms | 33.63 MiB |
| 30 | 4,615 ms | 38.88 MiB |

The complete unit consumed 12.989 CPU seconds over 55.342 seconds wall time.
Systemd reported 384.4 MiB peak cgroup memory, close to its cap: this includes
the generated files' reclaimable page cache, the probe, and kernel charges, not
just server RSS. Low process RSS does not mean the workload used only 21 MiB
of host memory. The host still had 708 MiB available after completion.

The live service retained PID 217528 and zero restarts, and its local readiness
endpoint returned `{"service":"flea","status":"ready"}` afterward. The
benchmark unit exited successfully. This does not constitute continuous public
uptime monitoring.

## Interpretation and limits

This supports **dozens of modest-traffic, mostly static sites on this 1 GB VPS**
as a plausible target. Disk-backed assets avoided retaining 600 MiB of payload
in the serving process. Descriptor counts grew roughly with the asset count,
so descriptor budgets and old-snapshot cleanup remain important at larger scale.

It does not establish maximum capacity. These are short synthetic runs without
TLS, browser rendering, photo decoding/uploads, logged-in editors, session
activity, real client templates, databases, or mini-app subprocesses. The client
shared the server's resource group and did not test bandwidth over the internet.
Fixtures were generated before measurement; filesystem caches were not cleared.
RSS/FD sampling every 20 ms can miss brief peaks. Combined RSS can double-count
shared pages. The overlap is validation only, not archive transfer/extraction or
a real deployment/restart. No PageSpeed score was measured.

Next: repeat with representative client sites, mixed editor/upload work, a
longer soak, and the first actual mini app before settling production capacity
or always-running app-process policy.

Reproduction details: [capacity probe](../../../scripts/capacity/README.md).

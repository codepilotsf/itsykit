# Isolated Linux capacity probe

Opt-in operator tool, not a normal test or a command to run against live site
configuration. Generates 30 synthetic sites beneath a new `fixtures/` directory
and launches a separately supplied server executable for 1, 10, and 30 sites.
The directory must not already contain `fixtures/`; existing site data is never
an input. It does not install, deploy, restart, or configure the live service.

Build the helper and server locally for the target Linux architecture. Upload
them into a fresh, private, disk-backed temporary directory. Run the helper as
an unprivileged user inside a transient systemd unit with:

- A private network namespace (loopback port 18089), no new privileges, and a
  read-only filesystem except the exact temporary directory.
- `MemoryMax=384M`, `MemorySwapMax=0`, `CPUQuota=70%`, `TasksMax=64`,
  `LimitNOFILE=4096`, and `RuntimeMaxSec=180`.
- Helper environment `GOMEMLIMIT=64MiB`, `GOMAXPROCS=1`. The helper explicitly
  gives server/validator children `GOMEMLIMIT=192MiB`, `GOMAXPROCS=1`, with no
  inherited mail credentials.

Invocation inside that isolation:

```text
flea-capacity-probe -server /absolute/test/flea-server -root /absolute/test/directory
```

Obtain authorization before using a shared VPS. Verify free space, memory, the
test directory's ownership, effective limits, and the live service before and
afterward. Never use `/tmp` on a RAM-backed filesystem for the fixture payload.
Record JSON stdout via the unit journal. After the unit exits and its processes
are gone, remove only the exact generated directory. Do not prune real releases.

## Workload and measurements

Each site has five compiled pages, a 20-item JSON list rendered on its homepage,
and twenty separate 1 MiB regular asset files. The binary payloads are synthetic,
not valid images or a browser workload; inodes are independent across sites.
Fixtures total 600 MiB, created before measurements, so this is not a cold-cache
test. There are no admins, mail, databases, sessions, or mini-app subprocesses.

Each phase issues 900 requests at a nominal 60/s (80% homepage, 20% complete
1 MiB assets), with at most eight outstanding requests and a three-second client
timeout. Sites rotate across requests; not every asset is fetched. A second
`-validate` process loads the same site set halfway through traffic. This probes
overlapping validation, not a full deployment/archive/restart workflow.

Process RSS and descriptor counts are sampled every 20 ms from `/proc`, so
brief peaks may be missed. Latency covers the full response body over loopback,
including load-generator scheduling. It excludes public networking, TLS and
browser work. The client shares the test CPU/memory cap. RSS sums may double
count shared executable pages; cgroup memory also includes fixture file cache
and kernel charges. Do not call RSS the total host RAM requirement.

This is a short capacity sanity check, not a saturation test, long-running leak
test, PageSpeed result, or a guarantee for arbitrary client sites.

See [the recorded VPS run](../../docs/agents/benchmarks/2026-10-04-multisite.md).

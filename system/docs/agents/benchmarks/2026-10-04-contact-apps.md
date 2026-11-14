# Managed Contact on Linux

October 4, 2026, 20:08:38–20:09:21 UTC. Authorized isolated test on
`fleacms.com`; no installed binaries, live sites, credentials or service settings
were changed.

## Environment and method

- Linux x86-64, one vCPU, 956 MiB RAM, 1 GiB configured host swap.
- Before upload/testing: live Flea PID 217528, active, zero restarts.
  Available host RAM was 709 MiB during preflight and 732 MiB after upload.
- A transient systemd unit ran as nobody/nogroup, with a private network and
  private temporary directories, ProtectSystem=strict, ProtectHome=yes and
  NoNewPrivileges=yes. Only the exact test directory was bind-mounted writable.
- MemoryMax=384 MiB, MemorySwapMax=0, CPUQuota=70%, TasksMax=256,
  LimitNOFILE=4096, RuntimeMaxSec=180. Effective settings were checked while running.
- Builds were cross-compiled on macOS ARM64 with Go 1.27.1, CGO disabled, from
  base commit `fdd2e49` plus the uncommitted managed-app work. Executables ran
  natively on Linux; this was not a native Linux build or full Linux test suite.
- Synthetic sites each had three small templates and an enabled Contact app.
  No large assets, admin sessions, databases or mail credentials were present.
  The server used GOMAXPROCS=1 and GOMEMLIMIT=192MiB; app children received their
  ordinary minimal environment, not those parent overrides.

The first attempt failed before any test ran: PrivateTmp hid the uploaded
`/var/tmp` directory. Binding only the exact test directory into the sandbox
resolved this without removing private-network or filesystem isolation.

## Results

Each phase launched one server and one Contact child per site, then made 300
sequential homepage requests, rotating sites and pausing 20 ms plus measurement
time between requests. Memory/FD/thread snapshots followed each request. All
900 requests across the three main phases succeeded.

| Sites/apps | Startup | Idle total PSS | Peak sampled total PSS | Peak summed RSS | Total FDs | Total OS threads | p95 response |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | 153 ms | 25.07 MiB | 26.25 MiB | 26.26 MiB | 22 | 13 | 0.708 ms |
| 10 | 227 ms | 34.77 MiB | 35.60 MiB | 88.67 MiB | 148 | 77 | 0.745 ms |
| 30 | 420 ms | 54.43 MiB | 55.43 MiB | 226.39 MiB | 428 | 221 | 1.557 ms |

PSS is proportional set size: shared executable pages are divided among their
processes rather than counted once per process as in summed RSS. The shared
Contact executable matters here. At 30 sites, the parent had 21.77 MiB PSS and
all 30 children together had 32.66 MiB PSS. A lone child had 6.78 MiB PSS; neither
number should be multiplied blindly to predict arbitrary future workloads.

The entire test unit, including the probe and preceding integration tests,
peaked at **61.6 MiB cgroup memory**, with no test swap. It consumed 10.024 CPU
seconds over 42.699 seconds. Cgroup memory includes more than sampled process
PSS, such as kernel charges and charged file cache; these are different measures.

Startup means that all child PIDs existed and a homepage answered, not that
every app handshake had completed. The separate integration tests below verify
authenticated readiness and actual form behavior.

## Lifecycle and integration checks

These compiled Go tests passed under the same Linux isolation:

- `TestManagedContact`: authenticated launch, validation, successful mocked mail,
  nested redirect, one-time flash, delivery failure and changed-config rejection.
- `TestChildAuthenticationAndParentExit`: missing/wrong/duplicate bearer rejection,
  separate email capability credential and child exit on parent-liveness EOF.
- `TestManagedLifecycle`: multiple managed services, shared executable ownership,
  independent shutdown and executable cleanup.

In each 1/10/30-site phase, SIGKILL of one known test child produced a replacement
after 1,010/1,012/1,023 ms respectively. The homepage remained usable at the
recovery check. Graceful parent shutdown and child exit completed in 3/12/32 ms.

A fourth one-site phase repeated 300 requests with zero failures, then killed
the parent with SIGKILL. Its child stopped within the next 2 ms measurement.
This checks process exit, not durable cleanup after a crash. Exhaustion of all
five restart attempts and repeated long-running crash loops were not tested.

Afterward the live service remained active with the original PID 217528 and
zero restarts; local readiness returned `{"service":"flea","status":"ready"}`.
Host available memory was 698 MiB. No test processes remained. This is a before/
after health check, not continuous public uptime monitoring. The disposable
72 MiB remote test directory was removed after the unit stopped; these generated
fixtures/executables can be rebuilt. No live content or retained releases were removed.

## Interpretation and remaining work

Thirty lightweight Contact-enabled sites fit comfortably in this isolated run
on the 1 GB host. It supports the direction, not a universal site-count guarantee.
Process/thread budgets matter: 30 apps plus their parent used 221 OS threads,
against the live service's currently observed TasksMax of 375. Do not raise that
limit automatically or infer a precise maximum from one short run.

These were tiny synthetic pages with mostly idle app processes. The workload
does not establish sustained submission throughput, TLS/ACME overhead, real
provider delivery, editor/uploads, database apps, full deployment peaks, or
long-running memory stability. No PageSpeed/browser score was measured.
Use representative client sites and a longer mixed workload before promising
production capacity or choosing an always-running policy for heavier apps.

## Provenance

Uploaded hashes matched local SHA-256 values:

```text
b2800b4fedeb782402e4fdc93824d149e858f0dd8bf69822291660f715f61d4c  probe
1101a1b9d21b2e53d483654f1248ae9d2e0ce81d03a71f35c1180da88603cd4e  flea-server
0450b49ec88ce3f43094c8502e9a37d7e8763f989f2281843c318f69610e82c8  site.test
6672c4f3d504eebaacf65f8866b65863779b8495945ce78bdd79bd670f857f8f  runtime.test
baf9a18870d6375e17a08eaf6db14ec519970328554d8aca65edc7a1bccb3cce  manager.test
```

See the [probe instructions](../../../scripts/app-capacity/README.md) and the earlier
[static-site capacity check](2026-10-04-multisite.md). Their workloads differ;
do not treat the tables as a controlled before/after comparison.

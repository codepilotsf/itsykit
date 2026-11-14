# Isolated managed-app probe

Opt-in Linux operator tool. Creates a new `app-fixtures/` directory beneath an
explicit test root, runs a separately supplied server and measures 1/10/30
Contact-enabled sites. It refuses an existing fixture directory. It never reads
live site configuration or installs/replaces a binary.

Build locally for the target architecture after refreshing the matching Contact
bundle with `scripts/build.sh`. Compile `./scripts/app-capacity` and
`./cmd/itsy-server` with CGO_ENABLED=0 and the target GOOS/GOARCH. Obtain approval
before transferring private executables to a remote host. Upload only to a new,
private, disk-backed directory; verify SHA-256 values at both ends.

Run unprivileged in an externally capped transient systemd unit, never directly
against the live service:

- PrivateNetwork=yes, PrivateTmp=yes, ProtectSystem=strict, ProtectHome=yes,
  NoNewPrivileges=yes; permit AF_INET, AF_INET6 and AF_UNIX.
- BindPaths and ReadWritePaths for only the exact test directory. PrivateTmp
  otherwise hides directories uploaded under `/var/tmp`.
- MemoryMax=384M, MemorySwapMax=0, CPUQuota=70%, TasksMax=256,
  LimitNOFILE=4096, RuntimeMaxSec=180. Check host headroom first.
- Probe environment GOMAXPROCS=1, GOMEMLIMIT=64MiB. No email credentials.
  The probe gives its server GOMAXPROCS=1 and GOMEMLIMIT=192MiB; the actual
  supervisor chooses Contact's ordinary minimal environment.

Inside that isolation:

```text
probe -server /absolute/test/flea-server -root /absolute/test/directory
```

It uses loopback port 18089 in the private network namespace. All signals target
PIDs launched by this probe or their directly observed children. It kills one
child per phase to check replacement, gracefully stops the parent, and finally
checks child exit after abrupt parent death in a separate one-site phase.
The external unit must kill remaining processes if the probe fails or times out.
After the unit stops, verify no test processes remain before deleting only the
exact disposable directory. Preserve live releases, data and configuration.

Output is JSON. PSS comes from `/proc/PID/smaps_rollup`; RSS, thread counts and
FDs come from `/proc`. Summed RSS double-counts shared pages. Cgroup memory peak
is a separate systemd measurement. Sequential homepage requests and per-request
sampling are a short idle-app sanity check, not a sustained submission benchmark.
Startup counts child PIDs and a successful homepage, not completed handshakes.

For actual app operations, also run cross-compiled package test executables
under the same isolation with FLEA_REQUIRE_APP_BUNDLE=1:

```text
site.test -test.run '^TestManagedContact$' -test.v -test.timeout 30s
runtime.test -test.run '^TestChildAuthenticationAndParentExit$' -test.v -test.timeout 30s
manager.test -test.run '^TestManagedLifecycle$' -test.v -test.timeout 30s
```

Create them with `go test -c` for `./internal/site`, `./internal/appruntime` and
`./internal/appmanager`. These tests use mock delivery and temporary fixtures;
they do not verify Resend. Native execution of cross-compiled tests is not a
native build or full Linux test suite.

Recorded result: [October 4 Contact check](../../docs/agents/benchmarks/2026-10-04-contact-apps.md).

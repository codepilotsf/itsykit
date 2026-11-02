# Temporary storage for development and deployments

Large builds and deployments can exhaust a RAM-backed `/tmp` or its per-user
quota even when the main disk has space. Check both filesystem capacity and user
quota; `df` alone does not show the user's remaining allowance.

Flea deployment packaging and server setup use Go's standard temporary directory
selection. Set `TMPDIR` for the job when the default is unsuitable. The selected
parent must already exist and be writable. Flea does not change the environment
or retry in another location after an explicit setting fails. Shell packaging
scripts also honor `TMPDIR`. Remote SSH staging is separate and does not inherit
the caller's temporary-directory setting.

On Linux, first check storage placement:

```sh
findmnt -T /tmp
findmnt -T /var/tmp
df -h /tmp /var/tmp
```

When `/var/tmp` is disk-backed, create a private directory for one job:

```sh
(
  scratch=$(mktemp -d /var/tmp/flea-job.XXXXXXXX) || exit 1
  trap 'rm -rf -- "$scratch"' EXIT
  trap 'exit 130' INT
  trap 'exit 143' TERM
  TMPDIR="$scratch" bash scripts/build.sh
)
```

Use `TMPDIR="$scratch" flea deploy` instead for a deployment from a clean,
committed site checkout. Keep the path short because native SSH places its
control socket inside the invocation's directory. Each concurrent job should
create its own directory. `mktemp -d` makes it private; cleanup removes only that
job's tree. Builds leave their final outputs in `bin/` and their normal Go cache
outside scratch. Do not export a machine-wide `TMPDIR` as part of this workflow.

Flea removes its local deployment/setup directory after success, ordinary errors
and handled Ctrl-C/SIGTERM. Remote staged files and SSH control connections have
bounded cleanup contexts independent of the cancelled job. SIGKILL, power loss
and machine crashes cannot run deferred cleanup or shell traps. Inspect leftovers
before removing them; Flea never scans for or deletes other jobs' directories.

Keep worktrees in a durable workspace. Save screenshots, review evidence and
logs that need to survive a job in a project artifact directory or private user
state directory, outside scratch and public site assets. Reusable tool, model and
browser caches should use their producer's documented disk-backed cache location.
Do not move a running tool's environment or cache beneath it.

An age-based backstop needs an explicit, narrowly scoped policy for a known
producer. Protect active jobs with locks honored by that cleanup mechanism and
keep retained evidence outside the cleanup tree. Directory names, age and absence
from one process snapshot do not prove abandonment. Inspect the host's existing
tmpfiles rules and timer before proposing another rule; do not install broad
cleanup rules or change quotas or mounts to fix a producer's lifecycle defect.

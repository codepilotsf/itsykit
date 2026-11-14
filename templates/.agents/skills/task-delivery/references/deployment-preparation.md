# Deployment preparation

Read only when deployment or host resource pressure is in scope. This guide
does not authorize a rollout.

## Resource-aware deployment preparation

When deployment is in scope on a shared or constrained host, inspect available
memory, container/process usage, build location and rollout headroom. Recommend
config changes that reduce unnecessary process duplication and keep compilation
off the app host when an approved local or dedicated builder is available.
Budget limits and concurrency for the actual workload; a low idle measurement
does not establish a safe peak limit. Keep temporary build checkouts and logs on
disk-backed storage when a RAM filesystem or quota is constraining them.

Preserve capacity checks, durable work and sequential rollouts. A configured
staging destination need not become a new running instance merely to validate
config. Use already-authorized recovery actions when capacity is insufficient;
request only an additional scope or ownership decision that remains unresolved.
Do not bypass a failed safety check or treat a successful build as rollout
permission. Verify the deployed process model, resource limits and job health.

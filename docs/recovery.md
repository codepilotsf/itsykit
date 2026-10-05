# Recovery commands

Run these from a site's current Git branch. They use its saved SSH destination,
not a production/staging alias. No HTTP administration endpoint is exposed.

```sh
flea rollback --list
flea rollback
flea rollback <release-id>
flea backup ../my-site.flea-backup
flea restore ../my-site.flea-backup
flea update-server
flea update-server --recover
```

Server-wide site administration needs no local site or saved deployment settings.
It accepts an explicit SSH target (otherwise it asks for the server login, for
example `root@fleacms.com`):

```sh
flea server list [ssh-target]
flea server remove [ssh-target] [site-name]
```

## Rollback

Rollback republishes earlier code as a new release while preserving current admin
snips, multisnips, photos and theme. It moves all of this site's aliases together;
other sites, HTTPS policy and redirects stay unchanged. It asks for confirmation
and refuses if the active release changed after review. Incompatible old templates
fail validation rather than discarding current content.

Only releases reachable through the current activation history are selectable.
Failed deployments are not rollback targets. The latest 1,000 activations are
listed; older files remain on disk. Installations without recorded history must
deploy with the current runtime before a rollback history is available. No
automatic migration or guessing from directory timestamps is performed.

## Backup and restore

Backup downloads a verified, bounded snapshot of the active site's code, public
configuration, live snips, multisnips, photos and theme under the deployment lock.
It refuses existing output files and creates new files with private permissions.
Keep backups outside the website directory and copy them to separate storage.
Backups can contain admin addresses and content; handle them as private data.

Credentials, certificates, deployment settings and sessions are excluded. A backup
is not a whole-server disaster-recovery image. Record server/email setup separately.

Restore requires typing `RESTORE`. It replaces live code, site configuration and
editable content with the backup, so first take a fresh backup if newer changes
might matter. The destination must already exist and match the backup's site and
branch. Host routing, credentials, certificates and session files stay in place.
Restored admin configuration still determines who has access.

Checksums detect corruption, not authorship. Only restore backups you trust.
Extraction rejects links, traversal, malformed archives and oversized payloads.
If live state changes while you review the restore, the operation stops. Failed
activation restores the previous configuration; interrupted activation retains
the existing server-side recovery journal.

## Server upgrades

Upgrade affects every site on the selected server and asks before restarting it.
It validates all installed sites with a pinned copy of the new bundled executable,
keeps the previous executable, journals the swap, restarts, and probes all sites.
Failed verification restores the prior executable. Interrupted upgrades block
further deployments and content saves until `flea update-server --recover` completes.

This updates the executable only. It does not reset sites, migrate data formats,
change SSH permissions, or replace customized systemd units. Incompatible site
formats fail before replacement. Service-policy changes require administrator
review. Keep backups and test upgrades on a disposable server first.

Local automated tests cover failure and recovery paths. Live SSH/systemd upgrade,
restore and power-loss behavior still need disposable-server verification.

## Removing a site

`flea server list [ssh-target]` shows each site's name and branch, then its HTTPS
domains, then its HTTP-only domains, then redirects. It is read-only and does not
take the deployment lock, so it also works while a recovery journal is pending.

`flea server remove [ssh-target] [site-name]` unregisters one site's routing
after the server prints the exact affected domains and you type the exact site
name. Omitting the site name lists the managed sites and asks for one by its
exact name. It is unregister/offline only: it deletes no files. Site files stay
under `/var/lib/flea/sites/<name>` and cached certificates stay under
`/var/lib/flea/certs`, but retained files are not restored automatically:
content lookup follows active routing, so re-deploying the same stable site name
starts from the current project and cannot recover content left behind. Back up
or pull the site before removal if a later deployment must keep its current
content. No restore command is added.

Removal rewrites the shared server config and restarts the service, so other
sites are briefly unavailable. Other sites' routing and ancillary entries are
preserved. When no HTTPS hostnames remain afterwards, automatic HTTPS is disabled
and cached certificates stay on disk.

Removal uses the same deployment/content lock, expected-state digest, recovery
journal, restart/probe and rollback paths as deployment. If the site changed
between review and confirmation, removal stops. If a deployment or executable
upgrade is pending, removal stops. An interrupted removal leaves the existing
`deploy.pending.json`; run `sudo flea-server -recover` to restore the previous
configuration. Removing the last site is supported and still probes server
readiness. Live SSH/systemd removal and recovery still need disposable-server
verification.

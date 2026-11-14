# Blog synchronization and source CSS

### Blog content synchronization

Commit `itsy/snips.json` when seeding Blog posts or cover metadata. Normal
`itsy deploy` preserves live posts and photos by stable ID and adds source-only
IDs. Omitting an ID never deletes a live post or photo. Publish code/CSS normally.

```sh
itsy pull
# Review and commit local Blog content and photos.
itsy deploy --with-content
```

`pull` retrieves Blog posts and cover photos from this branch's saved destination,
keeps local-only IDs and never pulls CSS, templates, configuration or sessions.
Stop dev first; selected local files must be committed. Explicit publication
requires a destination-scoped pull baseline for existing sites. Posts and photos
compare as whole records; conflicting edits require an explicit choice and stop
noninteractively. The expected remote digest is checked under the deployment lock.
`itsy status` reports content differences read-only. `--with-snips` is removed
without an alias. This contract requires deployment protocol 30.

Backups include Blog content and library photos. Rollback preserves current
content while selecting earlier code/CSS. Restore requires an expected snapshot
and explicit confirmation. Retained files are never pruned automatically.


Dev, pull and reset retain their exclusive per-project OS lock. Remote edits and
activation share deploy.lock; never hold it while prompting. Baselines stay in
private Git metadata, scoped to exact destination and project. CSS stays in Git
and is excluded from content sync. Keep expected-state and recovery checks.

# Public 0.15.1 release verification

October 13, 2026. User authorized building, verifying and publishing the public
release. One agent; unrelated main and Mac working changes were preserved.

Source/tag: `d5958ac2ab92971d9f1522ffdddcffdcb78aee97` / `v0.15.1`.
Framework worktree: `/home/sean/Projects/flea-release-0151`, branch
`release/0.15.1`. Mac isolated clone: `/Users/sean/Projects/flea-release-0151`.
Website worktree: `/home/sean/Projects/flea-website-release-0151`, same branch name,
based on `9d7a28d`; deployed source `d6cd9c9` through an isolated clean main clone.
No source branches were merged. The release tag was pushed.

## Artifact verification

Go 1.27.1 clean tagged artifacts (`vcs.modified=false`). Native Apple Silicon
Mac and Linux ARM64 full tests with required real app/server bundles, vet, race
and standard `scripts/build.sh` passed. Preliminary ARM64 Go 1.26.9 checks also
passed. All three final CLIs embed the same Mac-built Linux server bundles;
ARM64 was rebuilt after selecting those bundles, then required bundle/managed-app
checks and final fresh-site smoke tests passed. Linux AMD64 was cross-compiled
on the Mac and smoke-tested natively on the host as nobody, with no host build.
Native AMD64 build/full-suite acceptance is not established.

Fresh clean/starter creation, initial Git commits, status, dev homepage/team/
Contact routes and both new skills passed. Generated Blog templates and all
bundled skills matched canonical starter bytes. Mac ad-hoc signature verification
passed; this does not establish notarization or clean-machine acceptance.

Final anonymous HTTPS downloads and installed binaries match:

- darwin-arm64: `784ca438a304faf344f1a2a6bbb4e2fd5685ced77411b5fc5c313a581b94c9f4`
- linux-amd64: `6c31975387b3712967066cd78533ef0aa6bdd92f27fba53860bb7a6ba36451f3`
- linux-arm64: `ee9d4ede230d14f0d3db74983148c58d526d3a576021a6ac0b37625d9b4e66a4`

## Publication and public checks

Staged beside the disk-backed store at
`/var/lib/flea/downloads/.flea-0151.FVF6MDwf`. Checksums, exact files and permissions
passed before installation under `/etc/flea/deploy.lock`. The immutable version
was moved without overwrite, validated with the established 320 MiB admin Go
memory target, then the shared service was restarted to load its catalogue.
All three public hashes and the retained 0.13.0 manifest passed before advancing
the channel. Host config/content/media/CSS fingerprints were identical across
that restart. No old releases were removed and host memory/swap policy is unchanged.

The website was normally deployed, preserving live content, without explicit sync.
Activation: `20261013-213030-9a92d56ccb61b7b62b20a800`.
Homepage, installation docs, Blog/skills availability, download links and
`latest.txt` now agree on 0.15.1. Current Theme/sync guidance matches the removed
features; historical release notes remain intact. Installer and social image
return 200. All eleven HTTPS domains remain healthy. Content/media/CSS fingerprints
and unrelated site roots/commits match; service is active and no recovery journals
or deployment-window service error entries remain. Final website `flea status`
reports code/content matching the committed deployment.

Public fresh installation, reinstallation, 0.13.0-to-0.15.1 update and already-current
checks passed natively on Mac ARM64, Linux AMD64 (nobody) and Linux ARM64, with
manifest hashes matching. The downloaded binaries also passed fresh-site/dev
checks; task dev servers stopped. Ordinary installed CLIs were left unchanged.
Shared server executable remains the already-installed 0.15.1 / protocol 25.
No real email, new interactive production editor or new Ubuntu/systemd setup was
tested; existing manual QA answers are preserved.

Evidence/website backup: `/tmp/flea-0151-release`, private directory. Local immutable
artifact archive: `/home/sean/fleacms-releases/0.15.1`. Mac public test directory:
`/tmp/flea-0151-public`; native AMD64 test directory:
`/var/tmp/flea-0151-native.T0IR9yVs`. Clean-Mac/notarization, native AMD64 full-suite
and native CI runner acceptance remain open. This is a public test release.

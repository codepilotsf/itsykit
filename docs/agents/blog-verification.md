# Minimal Blog verification

October 13, 2026. Task branch `feat/minimal-blog`, based on merged Theme removal
`fad20d2`, worktree `/home/sean/Projects/flea-minimal-blog`. One writer; the dirty
main checkout and existing checklist answers were preserved.

## Scope and result

Blog includes the private admin post list/editor, explicit draft/publish actions,
rich text, optional library/uploaded cover, published-only route/template views,
strict content records, normal-deploy preservation and explicit whole-post sync.
Posts share the existing atomic content/media snapshot. No extra service,
database, dependency, CLI or API was added. Protocol 25 admits the new contract;
the release candidate is 0.15.0. Commit, push and private server rollout were
authorized after implementation verification; rollout evidence is recorded below.

## Checks

- `go test ./...` and `go vet ./...` passed on native Linux ARM64, Go 1.26.9.
- Full race checks passed. Follow-up changed-package checks also passed, including
  local/managed persistence and concurrent reload tests. `go test -race -p 1 ./...`
  passed on the final candidate with package execution serialized to reduce
  competing work; test goroutines and assertions remained unchanged.
- `bash scripts/build.sh` passed, rebuilding both Linux server bundles and the
  native CLI. No plain parent `go build` was used.
- New regression tests cover draft-route privacy, sanitized body rendering,
  escaped data, forbidden rich-body contexts, mount collisions, validation,
  preservation of removed source IDs, three-way conflicts, local/managed
  publishing, stale versions, authentication, exact Origin, form/header CSRF,
  unexpected uploads, scoped script CSP, deploy preservation, cover backup and
  expected-state restore.
- T3 browser checked a disposable site with the real editor and editstore. A
  synthetic in-memory mail sender completed the normal OTP flow; production
  authentication was not changed. Draft Save returned 404 publicly; Publish
  returned 200 with rich formatting; published Save updated immediately;
  Unpublish returned 404. Cover upload processed and served a synthetic PNG.
  Reload retained the selected library cover. Stale Save kept the draft text;
  Cancel requested confirmation and restored the saved body. Desktop and 375px
  views had no horizontal overflow. Hidden publishing controls stayed hidden.
  Pasting rich HTML retained text/bold while removing images, SVG and scripts.
- Repository Chromium checks passed for ordinary page editing, photos,
  multisnips and reload. Browser and synthetic preview resources were stopped.
- Fresh minimal and starter sites were generated with `--skip-git`. Blog/admin
  templates matched source, enabled Blog accepted a published JSON record, and
  both finished with successful `flea status`. No deployment destinations were
  configured.
- The UI detector found one initially empty preview image; it was replaced with
  a dynamically created image that gets a source only after selecting a photo.
  The final diff has no whitespace errors.

The existing eight-second `TestLocalSavesDuringDevReloads` deadline was exceeded
intermittently before release. Its fixture omitted the toolbar icon cache, which
allowed watcher reloads to fetch missing icons under the content lock. The fixture
now copies canonical bundled icons and stays offline. Three consecutive race
runs passed (6.96, 7.44 and 7.45 seconds), followed by the full tests, vet and full
serial-package race suite with required bundle checks enabled. Assertions, locks
and deadlines were not relaxed. The narrow deadline remains sensitive to machine
load; this change makes the fixture complete, not a runtime deadlock workaround.

This environment used a task-local `shasum -a 256` adapter because `shasum` was
absent. Chromium used the existing cached browser and library-path prerequisites.

## Limits

The implementation verification did not publish artifacts or change live sites.
The subsequently authorized private rollout is recorded separately below. Native Mac/AMD64, real mail, live Ubuntu/systemd and Hub
approval/promotion were not verified. Preview noindex and older site Theme toolbar
adoption remain separate tasks. Existing sites need explicit Blog template/source
adoption while preserving custom layouts. Draft text is private; photo files are
public assets. Comments, taxonomy, scheduling, revisions and inline body photos
remain deferred.

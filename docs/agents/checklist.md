# QA checklist

Use a fresh test site and the rebuilt CLI. Restart any existing dev server after
updating the binary. Mark each result **Pass**, **Fail**, or **Skipped**, and add
notes inline. For failures, include what you did and what happened. Do not paste
API keys, login codes, cookies, or other secrets here.

Start with `flea new breakfast-qa`, enter that folder, and run `flea dev`.
These are manual acceptance checks, not a claim that live delivery, deployment,
or every browser has already been tested.

## Linux ARM64 installer verification, October 12, 2026

Native Linux ARM64 checks passed: full Go tests, vet, bundled build and the full
race suite with real Contact/server bundles required. The real ARM64 CLI passed
installer install/reinstall through local download fixtures, hash verification,
fresh clean-site creation, starter asset comparison and `flea status`. Installer
regressions cover Mac ARM64, Linux AMD64/ARM64, checksum failure and a missing
ARM64 manifest entry. Release-staging regressions check all three hashes and
reject missing ARM64 artifacts and replacement of an existing staged release.

This host lacks shasum; local build/staging checks used a temporary shasum command
adapter to the system sha256sum. Release runners still require real shasum as
documented. No public ARM64 download was published or tested; live 0.10.0 still
omits it. The only attached GitLab runner inspected has the hatchbox tag, not
the native release tags. Native tagged runners must be configured before using
the release pipeline. Existing manual answers below remain unchanged.

## Automated retest, September 28, 2026

Original manual answers below are preserved. Diagnostic regression tests verify
the simplified page, normal dev 404 and opening-line locations for unclosed
conditionals. The executable smoke test verifies recovery after an invalid edit.
Chrome's PublicPreviewBrowser check passed with Theme accordion sections closed
on entry. Full Go tests, vet and race checks passed. The bundled CLI build and
a fresh generated site's `flea status` also passed. This does not replace Safari, Firefox,
real-phone or live Ubuntu acceptance.

Installer/updater tests use synthetic local fixtures: installation/reinstallation,
paths with spaces, metadata and checksums, interrupted downloads, wrong versions,
existing-install preservation and concurrent changes. Public downloads and native
Linux installation have not been verified.

## Local site and editing

### Editor controls and dragging, September 29, 2026

Color/outline follow-up: Edit on uses green #16a34a. Text/media and collection
overlays use matching 2px dashed green #16a34a with 6px corners, including Add tiles.
Outline appearance/disappearance uses a 160ms fade; reduced motion is immediate.
Browser tests verify fade creation, inert exit outlines and rapid-toggle cleanup.
Computed-style browser checks, full Go tests, vet, bundled build and fresh-site
status passed. Updated server assets are served by Bookworm; no content changes.

Toolbar follow-up: removed the wordmark/account divider and idle hints, shortened
the switch label to Edit, used dark-off/green-on colors and compact Theme/Sign out
pills. Chrome checks verify those styles. Live navigation reported opt-in disabled
despite loaded transition CSS; layout heads now link that stylesheet directly.
A native cross-document Chrome test verifies transition ready/finished with the
authenticated editor. Direct linking alone did not clear the live error; an
early native-CSS opt-in in the layout head did. Final test-site commit `7347978`,
release `20260929-162023-cb984c768e27ae0667320f42`. Subsequent automated live
navigation reports viewport-size changes while browser control is attached.
Sean should verify the visible fade without automation attached. Do not mark
live visual transition acceptance complete yet. The seven toolbar refinements
were verified in the live DOM and local screenshot; full tests and vet passed.

Approved direction: default-on Editing switch remembered across pages, second-row
Save/Cancel, Theme suspension, dashed editable boundaries and direct collection
arrangement. Drag references: [Pinterest](https://mobbin.com/screens/1a89b2af-59e6-48ec-8417-4d68ec5eced8)
and [Juicebox](https://mobbin.com/screens/284e4cae-04e3-42dc-a55f-40bc41eb4624).
Implementation includes inset hover/focus controls, persistent touch controls,
floating drag preview, landing placeholder, neighboring movement and reduced motion.
Chrome regressions pass for default-on/off navigation, dirty-mode refusal, Theme
transitions/errors, media dialogs, pointer/keyboard ordering, Escape and animated
drop cleanup. Drag hit testing uses final slots so neighboring animations cannot
trigger oscillating reorder. Desktop/500px toolbar and drag-preview screenshots
inspected. Full Go tests, vet and race checks passed. Bundled build, fresh-site
status and starter-toolbar comparison passed. Updated the disposable server and
deployed release `20260929-155228-73921b0fbc541ebf0b5975e2`; test-site commit
`171db31` contains the toolbar and previously pulled theme. Public editor assets
contain the revised preference and drag logic. The saved heading and uploaded
Sean photo remain intact. Live theme SHA-256 matches the fresh pre-update backup:
`4b6cbfa8aab5306d7164dba106938caf1d96490d046f939dd7f6680cb7270fe2`.
Its current `--black: #112` was already present before the update.
Backup: `/tmp/flea-live-qa.YOhVfCy1/before-editor-controls.flea-backup`.
Safari and real-device acceptance remain pending.

### Photo picker redesign, September 29, 2026

Approved references: [Canva thumbnail library](https://mobbin.com/screens/46ba5a6f-1071-433d-bc97-5daee9db8740)
and [Airbnb compact picker](https://mobbin.com/screens/cbf0b1c0-fa37-4b21-a60d-2d9e1998bdfc).
Implemented thumbnail buttons with selection state, Upload photo, No photo,
larger preview and contextual alt text. Existing descriptions remain read-only.
Browser regression covers selection, upload preview, multipart Save, Cancel,
and rapid reopening without losing the preview. Desktop and 500px screenshots
inspected. Full tests, vet, bundled build and fresh-site status passed.
Safari and real-device visual acceptance remain pending.
Updated the disposable server and verified the new picker markup is served.
Team page returned HTTPS 200; the uploaded Sean photo retained its expected
SHA-256 `1fc05c95c5335d3ff9eeb3810e8a33b864daa25c6b1061bf925d9f75c75a3577`.

### Live reset verification, September 28, 2026

With Sean's explicit approval, `flea setup` replaced the existing Bookworm
installation on root@fleacms.com (Ubuntu 26.04 LTS, x86-64). The previous
installation was moved to `/var/backups/flea.D4DGw28rgk`, mode 0700 root:root.
It was not permanently deleted. The replacement service is active, its readiness
check passes, and it reports protocol 22. Configuration has zero sites. Public
HTTP returns the expected 404 for this empty host. Certificate and lock ownership
checks passed. Both fleacms.com and bookworm.fleacms.com resolve to the host.
These were the immediate post-reset results; deployment results follow.

### Live deployment verification, September 29, 2026

Fresh site `/tmp/flea-live-qa.YOhVfCy1/v1-acceptance` deployed successfully to
https://bookworm.fleacms.com using protocol 22. Sean authorized test-repository
commits, certificate terms/contact and reuse of the backed-up mail configuration.
Framework changes were not committed. First deployment correctly refused the
unextracted starter collection; running dev and committing extraction resolved it.
Activated release: `20260929-144043-dc63617aaf7b0efe5070311f`.

Public HTTPS returned 200 with successful certificate verification. HTTP returned
308 preserving `/contact?qa=1`. Config, snips, private mail templates, AGENTS.md
and the development email-preview route returned 404; /admin returned 200.
Baseline backup downloaded and validated successfully at
`/tmp/flea-live-qa.YOhVfCy1/baseline.flea-backup`.
Sean confirmed receiving the login email and reaching the signed-in toolbar.
Safari did not offer email-code autofill despite Apple Mail on the same device.
At Sean's request, keep the six-box UI unchanged; autofill remains a known
unverified compatibility limitation, not a release-blocking UI change.

### Live recovery checks, September 29, 2026

Final-workflow follow-up: a live contact submission labelled FleaCMS V1 acceptance
test returned HTTP 303 to /thank-you after the sender accepted it. Inbox receipt,
rendering and Reply-To still need Sean's confirmation. Fresh backup saved at
`/tmp/flea-live-qa.YOhVfCy1/before-final-workflow-checks.flea-backup` before further
preservation/recovery testing.

Sean confirmed receipt and said the contact email looks good. Inbox delivery and
visual rendering are accepted; Reply-To header inspection was not separately confirmed.

Starter refresh: fresh generated site validates with 21 snips. Desktop (1360px)
and narrow (500px) Chrome screenshots reviewed for the contact layout. Full Go
tests, vet and bundled build passed; physical phone review is not implied.

Clean/starter follow-up: `flea new` validates with one page; `--starter` and `-s`
create six fictional team members. Reset cancellation, active-dev lock refusal,
non-site/symlink/worktree refusal, Git preservation and private-backup retention
pass automated tests. The rebuilt executable reset a disposable local site and
its recovery copy remains under `/tmp/flea-modes-qa.kZ0pQG/`.
Full tests, vet, race suite and bundled build passed. Chrome at 390px verified
native mobile menu open/Escape close, including signed-in Bookworm with Team,
Contact and Admin links. Bookworm's team replacement was explicitly synchronized;
existing homepage text and retained photo library were preserved. No live reset ran.

Mobile menu refinement: live Chrome at 390px measured menu top and toolbar bottom
at 189px; backdrop also begins at 189px. Hamburger is 30px inside a borderless
48px target. Computed entry/exit transitions are 180ms; Escape closes the menu.
Reduced motion disables these transitions. Full tests, vet and build passed.

Created a fresh pre-test backup at
`/tmp/flea-live-qa.YOhVfCy1/before-recovery-checks.flea-backup`.
Deployed a disposable `/qa-recovery` marker, verified it publicly, then rolled
back. Marker changed from 200 to 404; rollback reported live snips/theme retained.
Restored the pre-test backup successfully (release
`20260929-144722-e37a0196eb438877397c0355`). The server-update command completed
validation/replacement/restart/readiness using the current bundled version.
This is not a version-to-version upgrade or interrupted-update recovery test.
Pull completed successfully afterward. Final public HTTPS returned 200 with
valid TLS, systemd was active and server protocol remained 22. Removed the local
marker in a test-repository commit so later deploys do not reintroduce it.

Owner media/theme survival with deliberate live edits, contact delivery,
interrupted-operation/failure testing, certificate cache inspection and long-term
renewal remain pending. Successful restart alone does not prove these checks.

### Owner text preservation, September 29, 2026

Sean saved the homepage heading as `This is new.` through the live editor while
the local committed snip still contained `Welcome to FleaCMS.`. Normal deployment
activated `20260929-145335-f17626a5a64018277be12471`; a fresh public HTTPS fetch
confirmed the edited heading remained. `flea pull --snips-only` then retrieved
that same edit locally. This verifies live owner text save, deploy preservation
and pull with genuinely differing local/live content. Media, collections,
metadata and theme preservation remain separate acceptance checks.

### Owner theme preservation, September 29, 2026

September 29 follow-up: Sean accepted the current browser/UI pass for now and
asked to move on to live workflow verification. This is user acceptance, not
evidence of additional browser/device combinations being tested.

Sean saved `--black: blue` through Theme while the local stylesheet retained the
original OKLCH value. Normal deployment activated
`20260929-145802-6317b9b2085bd44cc1a57de1`. Public theme SHA-256 was identical
before and after: `63b33afae4ca09dee3b32ec355e3b845792d9578cf800829a98d9cbcbf5bac51`.
`flea pull --theme-only` retrieved the same stylesheet and the local hash matched.
This verifies live Theme save, byte-for-byte preservation on ordinary deploy,
and pull. The six-box code-entry UI remains unchanged.

1. **Topbar modes.** Check Edit and Theme; Photos/Videos buttons should be absent. In each editing mode,
   the other mode buttons should disappear, Cancel and Save changes should take
   their place, and Sign out should remain available. Save should reflect dirty
   state. Check narrow screens and keyboard focus too.
   - Result:
   - Notes:

   Also View Source while signed in and signed out. Signed-in source should have
   one csrf-token meta tag, a declarative toolbar shell and an external script reference,
   not embedded editor libraries/styles or content JSON. Visitors must not receive
   the meta tag or toolbar. Confirm the layout explicitly uses app.current_admin
   and navigation.css is an ordinary site stylesheet.

2. **Developer errors and navigation.** Introduce a template error. Confirm
    the dev error page shows the filename, relevant code and correct line
    numbers, and the terminal logs the error. Fix it and confirm recovery.
    Navigate between pages and check crossfades without a previous-page flash.
    - Result: FAIL
    - Notes: remove the little eyebrow that says 500 development diagnostic at the top, and then also remove everything below the how to fix it hint we don't need try again or go to homepage or the footer. Also, I found that if I put an unclosed expression like an if tag without a closing if tag, the error is misleading. It says unsupported expression, but it should give me more detail to say in this case it was an unclosed conditional. ALso, We don't need a special 404 page for dev mode, just show the regular template.

3. **Theme layout.** Check the plain Theme title, separate brand/neutral/semantic
    color sections, one open accordion at a time, smooth accordion transitions,
    inset select chevrons and unclipped focus outlines. Check label help bubbles,
    Viewport width Min/Max, Text Max Width and Form Max Width.
    - Result: fAIL
    - Notes: brand colors should not be open by default.

4. **Image snips.** Add `<img snip src="/images/test.jpg" alt="Test">` with a
    local JPEG/PNG. Confirm dev imports it into a typed page snip and rewrites
    src/alt to `{snip ID.src}` / `{snip ID.alt}`, preserving classes/dimensions.
    Leave Editing on and click the image. Change alt text and replace its file; Done
    should preview, Save changes should persist, and Cancel should restore it.
    Check unsupported-file errors retain the draft and ordinary images cannot
    be edited. Try a snip inside a text snip: expect a source-located error with
    no partial template/content rewrite. Check shared image edits across pages.
    - Result:
    - Notes:

5. **Collections.** Leave Editing on. Check outlines, inset drag handles,
    remove X icons and an Add tile per collection, without a sidebar/dropdown.
    Test mouse/touch dragging and arrow/Home/End keys; Escape cancels a drag.
    Check the floating preview/shadow, landing placeholder and neighboring motion;
    repeat with reduced motion. Controls appear on hover/focus and remain on touch.
    Edit text and click photos directly. Empty photos show a portrait and Add photo.
    Upload a JPEG/PNG, enter alt text, and check its preview. Save, reload, and
    select that photo for another item from the library. Cancel an upload and
    confirm the original photo returns. Invalid uploads must retain the draft.
    Check Save/Cancel and refresh. Remove all items: Add must remain available,
    and a saved empty collection must stay empty. Test two groups independently;
    controls must never appear in saved content.
    Inspect the page's multisnip array in `flea/snips.json`.
    - Result:
    - Notes:

6. **Contact form and email templates.** Configure a test recipient, submit
    the contact form and check the thank-you page, delivered content and Reply-To.
    Customize the project email templates and check their previews and delivery.
    This check needs real mail credentials; mark Skipped if unavailable.
    - Result:
    - Notes:

    Also check default local preview mode with no Resend credentials: the first
    dev startup asks only for site name and optional owner email. Request a sign-in
    code and submit contact mail; both should open in the browser with sender,
    recipient, subject, date/time and rendered body. Use the code to sign in.
    Check the plain-text disclosure, narrow layout and browser-launch fallback.
    Previews are not sent, expire after 10 minutes and disappear when dev stops.
    Disable dev.email_preview and restart for the separate real-delivery check.

7. **Starter and documentation.** Review the default pages, generated README
    and agent/import guidance. Try the brochure-site import workflow on a small
    disposable example. Review the separate `website/` source for accuracy;
    it has not been published by this work.
    - Result:
    - Notes:

## Optimized local video

8. **Video snips.** Add `<video snip src="/videos/test.mp4" controls></video>`
    using a local fast-start H.264 MP4 with optional AAC-LC audio. Confirm typed
    JSON extraction and generated src/poster bindings. In Edit mode click the
    video, replace its file and choose a JPEG/PNG poster. Done previews privately;
    the same topbar Save changes publishes video and text changes together.
    Test keyboard activation and clear errors. Use your own optimized sample.
    - Result:
    - Notes:

9. **Template and playback.** View the extracted video in ordinary browsing mode.
    Confirm its poster, native controls, play/pause, seeking, sound and download.
    Test Chrome, Firefox, Safari and a phone if available; list those tested.
    Silent MP4s should work too. There is no WebM fallback or transcoding.
    - Result:
    - Notes:

10. **Replace and metadata.** Change the name/poster without selecting a new
    file. Then replace the MP4. The video ID and template references should
    remain stable. Cancel a draft and confirm nothing was changed. Old file
    versions are retained; there is intentionally no delete/prune UI yet.
    - Result:
    - Notes:

11. **Invalid uploads.** Try a non-MP4, an MP4 without fast start, unsupported
    video/audio codecs and a file over 25 MiB. Check clear rejection and retained
    form values. Tiny valid/invalid samples are in `internal/video/testdata/` in
    the framework checkout. Do not expect Flea to optimize or repair files.
    - Result:
    - Notes:

12. **Storage and Git.** Confirm metadata is in `flea/snips.json` and hash-named
    bytes are in `flea/media/`. Git should list metadata/template changes but
    not video bytes. Library limits are 50 entries and 64 MiB of retained files.
    Confirm replacing an image snip still works after a video has been uploaded.
    - Result:
    - Notes:

13. **Concurrent and failed saves.** Open two signed-in tabs, change a video in
    one, then try saving the stale form in the other. It should reject the stale
    save rather than overwrite the newer state. Check an interrupted upload can
    be retried and does not publish a broken catalogue entry.
    - Result:
    - Notes:

## Disposable server only

Use a test server/site with no valuable data. Do not run restore, recovery or
upgrade experiments against a live customer site. Record skipped checks honestly.

14. **First deployment.** Exercise destination setup, domain/HTTPS configuration
    and deployment from a clean commit. Confirm the resulting site, login,
    contact form, photos and videos work. Check playback and seeking remotely.
    - Result:
    - Notes:

15. **Default content preservation.** Make live snip, multisnip, theme, photo and
    video changes. Deploy a local template change normally. Live edits should
    survive while the new template becomes active.
    - Result:
    - Notes:

16. **Pull and fresh clone.** Run `flea pull --snips-only` with dev stopped and
    verify content, photo and video files arrive together. Repeat from a fresh
    clone whose committed video metadata has no local MP4 files. Start dev and
    confirm playback. Check `--theme-only` affects only the theme.
    - Result:
    - Notes:

17. **Explicit publishing and conflicts.** Test `--with-snips` and `--with-theme`
    separately. Make competing local/live changes and confirm conflicts are
    reported, not silently overwritten. A missing local referenced video should
    stop deployment with a useful error.
    - Result:
    - Notes:

18. **Rollback.** Deploy two distinguishable template revisions, edit live
    content and run `flea rollback`. Confirm the code returns to the prior
    revision while current content, theme, photos and videos remain intact.
    - Result:
    - Notes:

19. **Backup and restore.** Create a backup containing snips, multisnips, theme,
    photos and videos. Change the test site, then restore the backup through the
    documented confirmation flow. Verify the restored pages and video bytes.
    - Result:
    - Notes:

20. **Server upgrade and recovery.** Follow `docs/` recovery/upgrade instructions
    on the disposable host. Confirm an interrupted or rejected operation leaves
    a usable prior state. Protocol 22 requires a matching server; never use a
    destructive setup reset as an upgrade.
    - Result:
    - Notes:

21. **Distribution.** Once implemented, test the one-line installer and `flea update`
    on clean Apple Silicon macOS and x86-64 Linux environments. Check installation
    without sudo, PATH, `flea version`, creating a site, reinstall and updating.
    Verify failed downloads/checksums retain the working CLI and that local updates
    do not change the remote server. Test the actual public artifacts and Mac
    platform trust requirements. Homebrew/native packages are deferred for V1.
    - Result:
    - Notes:

22. **Project structure.** Create a fresh site. Confirm pages/layouts/assets are
    under site/, mailers use *.email.html/txt with main.layout.html, config is
    .flea.config.json, and content/skills live under flea/. Start dev, edit a page
    and an admin template, and confirm reloads. Add text/image/video snips,
    and confirm flea/snips.json, unified flea/media/ and ignored flea/runtime/.
    Public routes must still be /, /contact and /css/global.css, without /site/.
    On a disposable deployment, confirm deploy.branches is saved in config and
    not present in a backup. Protocol 22 needs the matching server binary.
    - Result:
    - Notes:

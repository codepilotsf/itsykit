# FleaCMS next steps

This is the working roadmap. Keep priorities and open decisions here, completed
changes in [CHANGELOG.md](CHANGELOG.md), usage in [README.md](README.md), and
architecture in [ARCHITECTURE.md](ARCHITECTURE.md).

## Proposed order

1. Improve admin and owner onboarding.
2. Build the Flea website using Flea.
3. Verify the full workflow on a fresh Ubuntu server.
4. Make Flea available through Homebrew.

Bundled agent skills are also planned and can be developed alongside the website.
The order above is a proposal, not a commitment to start implementation.
Live verification should finish before public distribution.

## Admin and owner onboarding

Current model: each site has approved email addresses in `admin.emails`.
Developers use `flea admin [email]`; owners sign in at `/admin` with an emailed
code. There is no separate account, organization, or role-management system.

- [x] Clarify how developers add owners and apply access changes.
- [x] Make the sign-in URL and first-login instructions easy to find.
- [x] Ship developer-editable admin/ HTML pages using the normal site layout.
- [x] Add scrollable page navigation and cover metadata editing without fleas.
- [x] Add custom layouts and a full-width admin shell with separate page editors.
- [x] Add persistent per-device admin sessions and configurable rolling idle expiry.
- [x] Improve wording for requesting a code, signing in, and editing content.
- [x] Explain missing email configuration and access problems while preserving
  login privacy and security.
- [ ] Verify the whole flow with real email delivery.
- [x] Add a Resend setup wizard with a hidden key, separate sender name/address,
  connection check, private .env save, and automatic local email configuration.

First pass uses a developer-shared `/admin` link; no invitation emails are sent.
Developers view or remove owners through `admin.emails` in the site configuration.
Automated login/editing tests and a browser check of fresh-site setup passed.
Live email delivery and the deployed end-to-end flow remain to be verified.
Additional account structures and roles are not currently in scope.

## Inline owner editing

- [x] Standardize framework, starter, directives and content storage on flea;
  bare `flea` attributes need no label. Old syntax has no compatibility alias.

- [x] Build an isolated Squire prototype with current-page Edit, explicit
  Save/Cancel, and a second formatting bar containing Page details.
- [x] Add current-link editing/removal and individual alignment icons.
- [ ] Revisit undo/redo toolbar buttons if needed; keyboard history remains.
- [ ] Revisit discovery of unlinked pages only when needed; no page menu for now.
- [x] Exercise H1–H4, basic formatting, lists, links, alignment, paste cleanup,
  undo/redo, shared regions, and site typography in Chrome.
- [ ] Review the [prototype](design/inline-editor/README.md) before integration.
- [ ] Verify real clipboard input, keyboard/selection behavior, Safari, Firefox,
  mobile, and accessibility.
- [ ] Extend server sanitization for H1 and restricted alignment with regressions.
- [ ] Integrate authenticated public-page editing with atomic conflict-checked
  saves, preserving CSRF, session, and deployment-lock protections.
- [ ] Keep `/admin` for login; return owners to the homepage in browsing mode.
- [ ] Remove the old dashboard when its replacement is verified. No compatibility
  layer without approval.

Media lives outside rich text. Plan repeatable multi-fleas separately for content
such as team members, with add, edit, reorder, and delete. Small/Large Text and
advanced rich-text features are deferred. The prototype does not change real sites.

## Flea website

Build the public website with Flea to test its own developer and owner workflows.

- [ ] Choose the site location, domain, audience, and initial pages.
- [ ] Explain the plain-HTML workflow and developer versus owner responsibilities.
- [ ] Show features through practical examples and an editing demonstration.
- [ ] Provide getting-started instructions, documentation, and release notes.
- [ ] Add verified installation instructions when distribution is available.

Use the starter conventions and selected Skip logo. Keep public claims consistent
with implemented and verified behavior. Hosting and publishing await a chosen
destination.

## Bundled agent skills

Include useful agent skills by default for developers working on Flea sites.

- [ ] Decide where skills ship, how new sites receive them, and how agents
  discover them. Avoid assuming a particular agent's directory convention.
- [ ] Create an initial skill for importing an existing small brochureware website
  and converting it into a Flea site.
- [ ] Use very basic Prewind styling, readable HTML, shared layouts/partials,
  local assets, and Flea's existing content conventions.
- [ ] Define how the import preserves content, navigation, and useful page
  structure while handling unsupported interactive or third-party features.
- [ ] Verify imports with Flea validation and rendered-page checks, including
  links, assets, and narrow-screen layouts.
- [ ] Test the skill on a representative small site and document how to use it.

The first skill is planned, not implemented. Choose a source site and resolve
packaging before building it. Preserve unrelated customizations in existing sites.

## Live verification before distribution

Local tests cover HTTP/TLS, simulated SSH, recovery, and email logic. They do not
establish successful live installation, email delivery, or public certificate
issuance. These checks need a disposable Ubuntu target, domain, and credentials.

- [ ] Install on a fresh supported Ubuntu server and verify systemd permissions.
- [ ] Deploy and check DNS, public reachability, pages, assets, and private paths.
- [ ] Verify HTTPS issuance, HTTP redirects, and certificate cache reuse.
- [ ] Exercise owner login, editing, and contact-form delivery with real email.
- [ ] Redeploy code and confirm owner edits survive.
- [ ] Exercise failed/interrupted deployment recovery on the disposable target.
- [ ] Record results and remaining limits before publishing.

Monitor certificate renewal after initial verification; issuance alone does not
prove long-term renewal.

## Homebrew and releases

Current source version is 0.2.0. Build and publishing scripts exist; remote
release pipelines and public distribution remain unverified.

- [ ] Choose the release host and Homebrew tap repository.
- [ ] Configure and verify the release pipeline and public downloads.
- [ ] Test Mac artifacts with both Linux server architectures bundled.
- [ ] Create a formula with versioned URLs and verified checksums.
- [ ] Test installation and the first-site workflow on a clean Mac.
- [ ] Publish installation instructions on the Flea website.

Follow [RELEASING.md](RELEASING.md). Creating repositories, pushing tags, and
publishing artifacts require an explicit request; roadmap entries do not
authorize those actions.

## Design discussions to schedule

These topics need a strategy and scope before implementation. They do not change
the proposed order above.

### Deployment access for multiple developers

Plan how multiple developers can deploy the same site without giving each person
root access or unrestricted sudo on the server.

- [ ] Separate server-administrator responsibilities from routine site deployment.
- [ ] Define per-developer identities, access grants, and revocation without shared
  root credentials.
- [ ] Decide how deployment permission is limited to specific sites and branches.
- [ ] Evaluate narrowly authorized server operations, including activation and
  service restart, without granting arbitrary commands or access to other sites
  and server secrets.
- [ ] Plan shared deployment settings, concurrent deploy behavior, and records of
  who deployed which commit.
- [ ] Define tests for unauthorized sites, revoked access, and privilege escalation.

Choose the authorization design before changing SSH or server permissions.
Deployment access remains separate from browser owner access.

### Components alongside partials

Explore whether components would solve concrete reuse problems beyond the
existing partials. No component syntax or implementation has been selected.

- [ ] Collect examples that are awkward with current partials, such as reusable
  cards or buttons with explicit inputs.
- [ ] Compare extending partials with introducing a distinct component concept.
- [ ] Discuss inputs, defaults, child content/slots, and variable scope.
- [ ] Define escaping, validation, and error behavior consistently with Flea's
  existing template safety rules.
- [ ] Assess whether the benefit justifies extra template syntax and complexity;
  preserve the plain-HTML workflow without requiring a frontend build system.

## Deferred work

These are candidates to scope and prioritize, not commitments.

- Visual rich-text editing, images, and revision history.
- Nondestructive server upgrades, rollback, release retention, and remote status/logs.
- Remote content sync and deployment-setting consolidation.
- Review existing legacy flea and deployment-format support for removal under
  the project's no-backward-compatibility policy.
- Parser fuzzing and additional resource-budget tests.
- Independent site runtime versions: keep the shared Go server for now.
  Revisit per-site processes and a routing host only when the need justifies
  the operational complexity.

## Standing decisions

- Prefer simple HTML, JSON, Go, and minimal dependencies.
- No backward compatibility by default during early development. Ask before
  adding any compatibility layer.
- Keep the selected Skip logo and canonical starter files in `templates/`.
- Update this roadmap when work is completed, deferred, or reprioritized.
  Remove stale implementation notes rather than appending a running history.

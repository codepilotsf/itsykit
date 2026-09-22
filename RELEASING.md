# Publishing Flea CMS

End users install only Flea CMS. Setup uses native SSH/SCP and extracts the
matching bundled static Linux server; neither machine needs Go or source code.
Both Linux architectures are gzip-compressed inside each CLI, with SHA-256 hashes.
Setup needs SSH access, not a release host or package registry connection.
Framework contributors and CI still need Go 1.26 or newer. Deploy also uses Git,
as required by the existing committed-source workflow.

## GitLab prerequisites

The repository's .gitlab-ci.yml builds tagged releases on native runners:

- linux-amd64 and linux-arm64 for server builds and tests.
- macos-amd64 and macos-arm64 for CLI builds and tests.

Register trusted shell runners with those exact tags. Each needs Go, Git, Bash,
gzip and shasum; the publishing Linux runner also needs curl. Race tests need a C
compiler on runners only. Go does not run on either end user's machine or VPS.
Jobs verify host architecture and refuse cross-architecture release builds.

Enable the generic package registry and anonymous downloads for this project.
The registry distributes releases but is not used by setup. No user tokens are
embedded. Protect release tags and the runners. Disable duplicate
generic package versions so a published version cannot silently be overwritten.
GitLab package API: https://docs.gitlab.com/user/packages/generic_packages/

## Release procedure

1. Finish and review changes. Update `internal/buildinfo/version.go` and move
   the Unreleased changelog entries into a dated section for that version.
   Commit the changes and run the tests.
2. Create and push a new immutable tag, for example v0.1.0.
3. Both native Linux jobs test and build static server binaries.
4. Both native Mac jobs embed both Linux artifacts and their hashes using
   scripts/bundle.sh, test extraction and build CLIs with the same version.
5. Review the four artifacts, then run the manual publish job. It uploads to
   the generic package registry and verifies unauthenticated downloads against
   local checksums. If a partial publication fails, investigate before retrying;
   do not replace a version already distributed to users.
6. Test the downloaded Mac CLI on a clean Mac and setup on a disposable Ubuntu
   machine for each architecture before announcing the release or updating a tap.

Artifacts are raw executables named flea-darwin-amd64, flea-darwin-arm64,
flea-server-linux-amd64, and flea-server-linux-arm64. A Homebrew formula
can install the appropriate Mac binary directly, with its release SHA-256 and
no Go dependency. Creating/publishing that tap remains a separate step.

## Development builds

Run from anywhere on this Mac:

```sh
GOCACHE=/tmp/flea-go-cache bash /Users/sean/Projects/flea/scripts/build.sh
```

The script locates Go on PATH or /usr/local/go/bin/go (GO_BINARY overrides it).
It cross-compiles both Linux servers with CGO disabled, compresses them with
deterministic gzip headers, stores checksums, tests extraction, and builds the
native CLI at bin/flea. The default version comes from internal/buildinfo/version.go.
Use VERSION=0.2.0-rc.1 for an explicit local build override. Values must be valid
SemVer without a leading v; prerelease and build metadata suffixes are supported.
Release and publish scripts require CI_COMMIT_TAG to match v plus the recorded
version. CLI and server output omit the tag's v prefix.
Contributors need Go 1.26+, Bash, gzip and shasum. Initial builds may download Go
modules. Installed users do not need these build tools.

Generated assets under internal/serverbundle/assets and bin/ are ignored by Git.
Always use the script after changing source: plain go build may embed stale
servers, or none in a fresh checkout. Source-only tests work without bundles;
FLEA_REQUIRE_BUNDLE=1 go test ./internal/serverbundle requires both real
binaries. The server never imports serverbundle, avoiding recursive embedding.
No publishing is necessary for development setup. Extraction is bounded and
checks the hash and static ELF architecture before reset confirmation or upload.

The pipeline is provided but has not been run on the remote GitLab instance.
Runner registration, public registry access, first publication, and real-machine
installation tests must be completed before claiming release readiness.

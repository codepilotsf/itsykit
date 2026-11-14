# ItsyWeb

ItsyWeb keeps its Go framework and public website in one repository:

```text
itsyweb/
  system/   Go CLI, server, starter templates and framework documentation
  website/  Current itsyweb.dev website and public documentation
```

- [Framework development and usage](system/README.md)
- [Website development](website/README.md)
- [Framework agent requirements](system/AGENTS.md)
- [Website agent requirements](website/AGENTS.md)

The canonical remote is `git@git.codepilot.com:itsyweb.git` (CP Git).
Both source histories and existing release tags are preserved. Historical tags
retain their original directory layouts; new work uses the structure above.

Build and check the framework with Go 1.26 or newer:

```sh
cd system
go test ./...
go vet ./...
bash scripts/build.sh
```

Develop the public website from its own directory using the installed CLI:

```sh
cd website
itsy dev
```

`system/templates/` is the canonical generated-site starter. `website/` is
the actual public website, with its own content, design and deployment settings.
Repository consolidation does not deploy either component or republish releases.

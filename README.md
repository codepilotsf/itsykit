# ItsyKit

Write HTML. Skip the build pipeline.

ItsyKit adds layouts, partials, JSON data, and clean URLs to ordinary HTML.
The Go CLI and server handle local development, SSH deployment, automatic HTTPS,
and backups. Add Contact or Blog when you need them.

Created by [CodePilot](https://codepilot.com).

[Get started](https://itsykit.dev/docs) · [Documentation](https://itsykit.dev/docs/templates)

ItsyKit keeps its Go framework and public website in one repository:

```text
itsykit/
  system/   Go CLI, server, starter templates and framework documentation
  website/  Current itsykit.dev website and public documentation
```

- [Framework development and usage](system/README.md)
- [Website development](website/README.md)
- [Framework agent requirements](system/AGENTS.md)
- [Website agent requirements](website/AGENTS.md)

The canonical repository is [codepilotsf/itsykit](https://github.com/codepilotsf/itsykit).
Clone it with:

```sh
git clone https://github.com/codepilotsf/itsykit.git
cd itsykit
```

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

## License

ItsyKit is available under the [MIT License](LICENSE). Bundled third-party
components retain their own license notices.

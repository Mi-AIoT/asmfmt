# Repository Guidelines

## Project Structure & Module Organization

This is a small Go module for formatting Go assembler files.

- `asmfmt.go` contains the formatter library exposed as package `asmfmt`.
- `cmd/asmfmt/` contains the command-line entry point and CLI-specific files.
- `asmfmt_test.go` contains package tests and golden-file comparison logic.
- `cmd/asmfmt/*_test.go` contains CLI/config discovery and integration coverage.
- `testdata/` stores formatter fixtures as paired `*.in` and `*.golden` files.
- `.asmfmt.toml.example` documents supported formatter configuration options.
- `.github/workflows/` defines CI checks for vet, tests, formatting, and releases.
- `refdoc/` is reference material and should not be staged or committed unless explicitly requested.

Keep library behavior in the root package and CLI behavior in `cmd/asmfmt`.

## Build, Test, and Development Commands

- `go test ./...` runs all package and CLI tests.
- `go test ./... -run 'TestCLIConfig|TestConfigResolver|TestFormatWithOptions|TestParseOptionsTOML'` is the focused config-feature regression pass.
- `go test -race -cpu="1,4" -short -v ./...` mirrors the CI race-test pass.
- `go vet ./...` runs static checks used by CI.
- `gofmt -w .` formats Go sources before submitting changes.
- `go test -run TestRewrite -update` refreshes `testdata/*.golden` files when an intentional formatter output change is made.
- `go install ./cmd/asmfmt` builds and installs the local CLI.
- `go run ./cmd/asmfmt [flags] [path ...]` runs the formatter locally, for example `go run ./cmd/asmfmt -d testdata/indent.in`.

## Coding Style & Naming Conventions

Use standard Go style: tabs for indentation in Go code, `gofmt` for formatting, and concise package-level helpers. Prefer clear, lowercase file names and Go identifiers that match existing code style. Keep formatter logic deterministic and avoid unrelated rewrites.

Default formatter output uses tabs for indentation and spaces for alignment. Preserve that default behavior unless the change is behind an explicit formatter option or is fixing a bug in the existing default formatting.

For new formatter options, preserve historical output when no config is present. New behavior should be opt-in unless the change is a bug fix to existing default formatting.

When adding config-related behavior, keep these layers separate:

- parsing and validation in the library,
- config discovery and caching in the CLI,
- formatting behavior in the formatter core.

## Testing Guidelines

Tests use Go's built-in `testing` package. Formatter behavior is validated through golden fixtures in `testdata/`: each `name.in` input should have a corresponding `name.golden` expected output. Add or update focused fixtures for every formatting change, and run `go test ./...` afterward.

On test failure, generated `*.asmfmt` files may appear next to inputs for inspection. Remove those diagnostic files before committing.

When adding a new feature, cover all relevant layers:

- unit tests for parser or formatter behavior,
- CLI tests for config discovery or flag behavior when applicable,
- end-to-end CLI integration tests when behavior depends on reading files, config, or stdin,
- CI workflow updates when the new feature needs explicit regression coverage.

For config-driven formatting changes, verify both:

- the formatter behavior through `FormatWithOptions` or equivalent library-level tests,
- the actual `cmd/asmfmt` executable path with temporary files/configs.

## Development Workflow

Choose the workflow based on the kind of change instead of treating all work the same.

Bug fix:

- reproduce the bug first with a focused test,
- implement the smallest fix that addresses the observed behavior,
- add or update regression coverage close to the layer where the bug was found,
- run the focused test plus `go test ./...` and `go vet ./...`.

New feature:

- align on requirements and a short implementation plan before coding,
- do the same for other tasks when scope, expected behavior, or tradeoffs are unclear,
- define whether the feature belongs in the library, CLI, or both,
- preserve existing default behavior unless the change is explicitly intended to alter defaults,
- add tests for the new behavior before or alongside the implementation,
- update README and related examples when the feature is user-facing,
- update CI when the feature adds a new regression surface that should be exercised explicitly.

Default formatting behavior change:

- treat this as high-risk because it can invalidate many existing fixtures,
- add or update focused fixtures in `testdata/`,
- verify that the change is intentional and deterministic,
- run `go test -run TestRewrite ./...` and inspect fixture diffs carefully before updating goldens.

New formatter config option:

- add parsing and validation in the library,
- keep config discovery and caching logic in the CLI,
- make the new behavior opt-in unless explicitly required otherwise,
- cover parser validation, formatter behavior, CLI discovery, and end-to-end CLI execution,
- document the option in README and `.asmfmt.toml.example`.

Docs or CI only change:

- keep the diff scoped to docs or workflow files,
- avoid mixing unrelated code changes into the same commit,
- run only the checks needed to validate the touched area unless the change affects broader behavior.

## Commit & Pull Request Guidelines

Commit messages should follow Conventional Commits 1.0.0:

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Use a valid type such as `feat`, `fix`, `docs`, `test`, `ci`, `refactor`, `perf`, `build`, `style`, `chore`, or `revert`. Add a scope when useful, write the subject in imperative mood, keep it concise, and avoid trailing punctuation. Add a body when rationale, behavior details, or validation notes matter. Wrap commit body lines at about 80 characters. When the details are easier to scan as separate points, use short flat bullet-style lines in the body instead of one dense paragraph. Add footers such as `BREAKING CHANGE:` or issue references when needed.

Prefer small logical commits instead of one large mixed commit. Split code, tests, CI, and docs when that keeps review easier.

Before committing:

- review `git status --short`,
- avoid staging `refdoc/` material unless explicitly requested,
- remove generated `*.asmfmt` diagnostics,
- run the relevant focused tests plus `go test ./...` and `go vet ./...` when the change is substantial.

Pull requests should include a brief problem statement, the formatter behavior change, updated tests or fixtures, and the commands run locally. Include before/after examples when a formatting rule changes.

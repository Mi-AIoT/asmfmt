# Repository Guidelines

## Project Structure & Module Organization

This is a small Go module for formatting Go assembler files.

- `asmfmt.go` contains the formatter library exposed as package `asmfmt`.
- `cmd/asmfmt/` contains the command-line entry point and CLI-specific files.
- `asmfmt_test.go` contains package tests and golden-file comparison logic.
- `testdata/` stores formatter fixtures as paired `*.in` and `*.golden` files.
- `.github/workflows/` defines CI checks for vet, tests, formatting, and releases.

Keep library behavior in the root package and CLI behavior in `cmd/asmfmt`.

## Build, Test, and Development Commands

- `go test ./...` runs all package and CLI tests.
- `go test -race -cpu="1,4" -short -v ./...` mirrors the CI race-test pass.
- `go vet ./...` runs static checks used by CI.
- `gofmt -w .` formats Go sources before submitting changes.
- `go test -run TestRewrite -update` refreshes `testdata/*.golden` files when an intentional formatter output change is made.
- `go install ./cmd/asmfmt` builds and installs the local CLI.
- `go run ./cmd/asmfmt [flags] [path ...]` runs the formatter locally, for example `go run ./cmd/asmfmt -d testdata/indent.in`.

## Coding Style & Naming Conventions

Use standard Go style: tabs for indentation in Go code, `gofmt` for formatting, and concise package-level helpers. Prefer clear, lowercase file names and Go identifiers that match existing code style. Keep formatter logic deterministic and avoid unrelated rewrites.

Formatter output uses tabs for indentation and spaces for alignment; preserve that behavior when changing formatting rules.

## Testing Guidelines

Tests use Go's built-in `testing` package. Formatter behavior is validated through golden fixtures in `testdata/`: each `name.in` input should have a corresponding `name.golden` expected output. Add or update focused fixtures for every formatting change, and run `go test ./...` afterward.

On test failure, generated `*.asmfmt` files may appear next to inputs for inspection. Remove those diagnostic files before committing.

## Commit & Pull Request Guidelines

Recent history uses short, imperative commit subjects such as `Detect when in a string literal (#47)` and `Support go:build pragma (#41)`. Follow that style: describe the behavior changed, keep the subject concise, and reference issues or PR numbers when applicable.

Pull requests should include a brief problem statement, the formatter behavior change, updated tests or fixtures, and the commands run locally. Include before/after examples when a formatting rule changes.


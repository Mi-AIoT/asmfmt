# asmfmt

[![CI](https://github.com/Mi-AIoT/asmfmt/actions/workflows/go.yml/badge.svg)](https://github.com/Mi-AIoT/asmfmt/actions/workflows/go.yml)
[![Release](https://github.com/Mi-AIoT/asmfmt/actions/workflows/release.yml/badge.svg)](https://github.com/Mi-AIoT/asmfmt/actions/workflows/release.yml)

An assembler formatter for Go / Plan 9 assembly, GAS syntax, and RISC-V code

English | [简体中文](README.zh.md)

`asmfmt` formats assembly source in the same spirit that `gofmt` formats Go
code.

It is primarily aimed at Go / Plan 9 assembly, and also supports a growing
subset of GAS-style syntax used in ELF-oriented codebases, including RISC-V.

Status: `STABLE`. Default output is intended to remain stable unless a bug is
being fixed.

This repository is actively maintained as a fork of
[`klauspost/asmfmt`](https://github.com/klauspost/asmfmt). It keeps the original
formatter behavior where practical, while continuing development for use cases
that are no longer moving upstream.

## Why asmfmt

`asmfmt` is useful when you want assembly code to be easier to review and keep
consistent across a codebase.

It can:

* normalize indentation,
* align operands and end-of-line comments,
* clean up blank lines and semicolon-separated statements,
* preserve historical default behavior when no config is present,
* run as either a CLI tool or a Go library.

## Quick start

Install the CLI from Go toolchain:

```bash
go install github.com/Mi-AIoT/asmfmt/cmd/asmfmt@latest
```

Alternatively, you can download pre-built binaries directly from the [GitHub Releases](https://github.com/Mi-AIoT/asmfmt/releases) page. For beta/nightly features, download the latest master branch build from the [`beta` pre-release](https://github.com/Mi-AIoT/asmfmt/releases/tag/beta).

If you are depending on this fork as a Go module, check the current `go.mod`
module path first. The CLI install path and the library import path may differ
until the module path is moved from the upstream origin.

Format a file in place:

```bash
asmfmt -w path/to/file.s
```

Preview changes as a diff:

```bash
asmfmt -d path/to/file.s
```

Format a directory tree:

```bash
asmfmt ./...
```

`asmfmt` only processes `.s` files. Do not run it on non-assembly inputs.

## Example

Input:

```asm
TEXT foo(SB),$0
MOVQ AX,BX //comment
loop:ADDQ $1,AX;JMP loop
```

Output:

```asm
TEXT foo(SB), $0
	MOVQ AX, BX // comment

loop:
	ADDQ $1, AX
	JMP  loop
```

For larger examples, see
[testdata](https://github.com/Mi-AIoT/asmfmt/tree/master/testdata).

## CLI Usage & Configuration

For the full CLI reference, detailed option explanations, configuration keys, and editor integration setup, please refer to the [User Manual](docs/user_manual.md).

### Basic Command

```bash
# Format a file in place
asmfmt -w path/to/file.s

# Format with a specific configuration
asmfmt -config /path/to/asmfmt.toml ./...
```

### Config Discovery

`asmfmt` automatically searches for a configuration file in the following order:

1. The nearest `.asmfmt.toml` walking upward from the directory of the file being formatted.
2. `~/.asmfmt.toml`
3. `/etc/asmfmt.toml`

For a fully commented configuration template, see [.asmfmt.toml.example](.asmfmt.toml.example).

## Supported syntax

`asmfmt` primarily targets Go / Plan 9 assembly and also supports a conservative
subset of GAS-like syntax.

Covered and tested areas include:

* Go / Plan 9 indentation and alignment rules
* GAS directives commonly used in RISC-V assembly such as `.section`,
  `.attribute`, `.option`, `.insn`, relocation helpers, CFI directives, data
  directives, and symbol visibility directives
* GAS macro blocks such as `.macro`, `.irp`, `.irpc`, `.rept`, `.if`, `.else`,
  and matching end directives
* GAS comments using `#`, `//`, block comments, and ARM/GAS `@` line comments
  when the source is detected as GAS-like
* RISC-V operand forms such as `%hi(...)`, `%lo(...)`, `%pcrel_hi(...)`,
  `%pcrel_lo(...)`, local numeric labels like `1b`, CSR operands, vector
  operands, compressed mnemonics, and `.insn` encodings

Unknown directives and unknown lowercase mnemonics are preserved
conservatively. `asmfmt` will still clean up surrounding whitespace and
alignment, but it does not try to validate or normalize unknown syntax beyond
safe formatting behavior.

## Fork status

This fork is intended to be maintained independently.

Current goals:

* keep the original formatter stable for existing users,
* continue improving GAS and RISC-V coverage,
* accept fixes and features without waiting on upstream activity,
* keep configuration and CLI behavior predictable.

The project still inherits design and historical behavior from the upstream
implementation, but documentation and releases in this repository describe the
forked project rather than the original upstream repository.

## Formatting behavior

Default behavior includes:

* tabs for indentation and spaces for alignment,
* operand alignment,
* end-of-line comment alignment,
* trailing whitespace removal,
* blank-line cleanup,
* label normalization onto their own lines,
* comment spacing cleanup,
* single-line block comment conversion when safe,
* semicolon cleanup and splitting where the detected style permits it.

If you need deterministic behavior across a mixed codebase, set
`source_style` explicitly instead of relying on auto-detection.

## Non-goals

`asmfmt` is a formatter, not an assembler or linter.

It does not aim to:

* validate opcode legality,
* reject unknown directives or vendor mnemonics,
* require binutils or another assembler at runtime,
* guarantee semantic equivalence by default,
* rewrite unrelated Plan 9 formatting behavior to match GAS conventions.

## Library usage

`asmfmt` can also be used as a Go package:

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/klauspost/asmfmt"
)

func main() {
	src := bytes.NewBufferString("TEXT foo(SB),$0\nMOVQ AX,BX\n")
	out, err := asmfmt.Format(src)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(out))
}
```

For configurable formatting, use `asmfmt.FormatWithOptions(...)`.

Note: the current library import path is still the upstream module path shown
above. If this fork adopts its own module path later, the import path in this
example will need to change accordingly.

## Development

Common commands:

```bash
go test ./...
go vet ./...
gofmt -w .
```

Refresh golden files after an intentional formatting change:

```bash
go test -run TestRewrite -update
```

More focused regression commands and contribution workflow details are described
in [AGENTS.md](AGENTS.md).

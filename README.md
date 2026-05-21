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

Install the CLI:

```bash
go install github.com/Mi-AIoT/asmfmt/cmd/asmfmt@latest
```

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

## CLI usage

```text
asmfmt [flags] [path ...]
```

Common flags:

* `-w` overwrite files in place
* `-d` print a diff instead of rewritten source
* `-l` print files whose formatting differs
* `-e` report all errors instead of stopping after the first 10
* `-config` load formatting options from a specific TOML file

If no path is provided, `asmfmt` reads from standard input and writes the
formatted result to standard output. `-w` cannot be used with standard input.

## Configuration

`asmfmt` supports a single TOML configuration file.

Load a config explicitly:

```bash
asmfmt -config /path/to/asmfmt.toml ./...
```

If `-config` is not provided, the first matching config is used in this order:

1. The nearest `.asmfmt.toml` found by walking upward from the formatted file's directory
2. `~/.asmfmt.toml`
3. `/etc/asmfmt.toml`

Rules:

* Only one config file is used.
* Config files are not merged.
* Unknown fields and invalid values are errors.
* Unset fields keep the built-in defaults.
* For standard input, project-directory lookup is skipped unless `-config` is provided.

Minimal example:

```toml
indent_style = "space"
indent_width = 4
source_style = "riscv-gas"
align_comments = true
```

Supported config keys:

* `indent_style`: `tab` or `space`
* `indent_width`: positive integer, used when `indent_style = "space"`
* `align_operands`
* `align_comments`: boolean. In GAS style, comment alignment is only calculated from lines with comments to prevent long comment-less lines from throwing off alignment.
* `align_continuations`
* `max_blank_lines`
* `split_semicolon_statements`
* `newline_before_comments`
* `newline_before_labels`
* `labels_always_on_own_line`
* `line_comment_space`
* `convert_single_line_block_comment`
* `preferred_comment_style`: `preserve` or `slash`
* `source_style`: `auto`, `plan9`, `gas`, or `riscv-gas`
* `indent_gas_directives`: boolean, whether to indent zero-indent GAS directives to the instruction level. Default is `false`.

See [.asmfmt.toml.example](.asmfmt.toml.example) for a fully commented reference.

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

## Editor integration

### GoLand

You can set up a custom File Watcher:

* `Settings -> Tools -> File Watchers`
* create a `<custom>` watcher named `asmfmt`
* `File Type`: `x86 Plan 9 Assembly file`
* `Scope`: `Project Files`
* `Arguments`: `$FilePath$`
* `Output Paths to Refresh`: `$FilePath$`
* `Working Directory`: `$ProjectFileDir$`

Enable:

* `Trigger the watcher regardless of syntax errors`
* `Create output file from stdout`

Disable the remaining advanced options.

![Goland Configuration](https://user-images.githubusercontent.com/5663952/114158973-96eebc80-9925-11eb-9aea-703ce474a7bb.png)

### Emacs

To format assembler files on save:

```elisp
(defun asm-mode-setup ()
  (set (make-local-variable 'gofmt-command) "asmfmt")
  (add-hook 'before-save-hook 'gofmt nil t)
)

(add-hook 'asm-mode-hook 'asm-mode-setup)
```

## Source style differences

`asmfmt` can auto-detect source style, or you can force it with
`source_style` in `.asmfmt.toml`.

The current style modes differ mainly in comment handling, statement splitting,
and syntax detection:

| Style | Typical input | Comment handling | Semicolon splitting | Detection hints | Notes |
| --- | --- | --- | --- | --- | --- |
| `plan9` | Go / Plan 9 assembly | `//` is treated as a line comment. `#` is not treated as a comment marker. `@` line comments are disabled. | Disabled. `;` is not treated as a normal statement separator. | `(SB)`, `(FP)`, `(PC)`, `(SP)`, and uppercase instruction forms tend to identify Plan 9 style. | Best for traditional Go assembler sources. |
| `gas` | Generic GAS-style assembly | `//` and `#` are treated as line comments. `@` line comments are also supported in GAS-like input when used in comment position. | Enabled where the input looks like GAS instruction or directive syntax. | Dot-directives such as `.section` and lowercase instruction forms tend to identify GAS style. | Best for non-Go ELF/GAS assembly that does not need RISC-V-specific detection. |
| `riscv-gas` | RISC-V assembly using GAS syntax | Same comment behavior as `gas`. | Same general semicolon behavior as `gas`. | RISC-V-specific forms such as `%hi(...)`, `%lo(...)`, `%pcrel_hi(...)`, `%pcrel_lo(...)`, `.option`, `.attribute`, `.insn`, `R_RISCV_...`, and common RISC-V register names promote detection to `riscv-gas`. | This is a GAS sub-mode with stronger RISC-V-oriented detection and syntax coverage. |
| `auto` | Mixed or unknown input | Uses the detected style for comment parsing. | Uses the detected style for statement splitting. | Starts conservative, then upgrades based on syntax seen in the file. | Recommended when your codebase is consistent and you want minimal configuration. |

If a mixed codebase needs deterministic results, prefer setting
`source_style` explicitly instead of relying on auto-detection.

# asmfmt
Go Assembler Formatter

English | [简体中文](README.zh.md)

This will format your assembler code in a similar way that `gofmt` formats your Go code.

Read Introduction: [asmfmt: Go Assembler Formatter](https://blog.klauspost.com/asmfmt-assembler-formatter/)

[![Go Reference](https://pkg.go.dev/badge/klauspost/asmfmt.svg)](https://pkg.go.dev/klauspost/asmfmt)
[![Go](https://github.com/klauspost/asmfmt/actions/workflows/go.yml/badge.svg)](https://github.com/klauspost/asmfmt/actions/workflows/go.yml)

See [Example 1](https://files.klauspost.com/diff.html), [Example 2](https://files.klauspost.com/diff2.html), [Example 3](https://files.klauspost.com/diff3.html), or compare files in the [testdata folder](https://github.com/klauspost/asmfmt/tree/master/testdata).

Status: STABLE. The format will only change if bugs are found. Please report any feedback in the issue section.

# install

Binaries can be downloaded from [Releases](https://github.com/klauspost/asmfmt/releases). Unpack the file into your executable path.

To install the standalone formatter from source using Go framework: `go install github.com/klauspost/asmfmt/cmd/asmfmt@latest`.

# updates

* Apr 8, 2021: Add modules info and remove other than main tools. 
* Jan 6, 2021: Fix C comments before line comments like `VPCMPEQB Y8/*(DI)*/, Y0, Y1 // comment...`
* Aug 8, 2016: Don't indent comments before non-indented instruction.
* Jun 10, 2016: Fixed crash with end-of-line comments that contained an end-of-block `/*` part.
* Apr 14, 2016: Fix end of multiline comments in macro definitions.
* Apr 14, 2016: Updated tools to Go 1.5+
* Dec 21, 2015: Space before semi-colons in macro definitions is now trimmed.
* Dec 21, 2015: Fix line comments in macro definitions (only valid with Go 1.5).
* Dec 17, 2015: Comments are better aligned to the following section.
* Dec 17, 2015: Clean semi-colons in multiple instruction per line.

# goland

To set up a custom File Watcher in Goland, 

* Go to Settings -> Tools -> File Watchers
* Press **+** and choose `<custom>` template.
* Name it `asmfmt`
* File Type, Select `x86 Plan 9 Assembly file` (it will apply to all platforms)
* Scope: `Project Files`
* Arguments: `$FilePath$`.
* Output Paths to Refresh: `$FilePath$`
* Working Directory: `$ProjectFileDir$`

Advanced options, Enable:

* [x] Trigger the watcher regardless of syntax errors (IMPORTANT) 
* [x] Create output file from stdout

Disable the rest.

![Goland Configuration](https://user-images.githubusercontent.com/5663952/114158973-96eebc80-9925-11eb-9aea-703ce474a7bb.png)


# emacs

To automatically format assembler, in `.emacs` add:

```
(defun asm-mode-setup ()
  (set (make-local-variable 'gofmt-command) "asmfmt")
  (add-hook 'before-save-hook 'gofmt nil t)
)

(add-hook 'asm-mode-hook 'asm-mode-setup)
```

# usage

`asmfmt [flags] [path ...]`

The flags are similar to `gofmt`, except it will only process `.s` files:
```
	-config string
		Read formatting options from this TOML file.
	-d
		Do not print reformatted sources to standard output.
		If a file's formatting is different than asmfmt's, print diffs
		to standard output.
	-e
		Print all (including spurious) errors.
	-l
		Do not print reformatted sources to standard output.
		If a file's formatting is different from asmfmt's, print its name
		to standard output.
	-w
		Do not print reformatted sources to standard output.
		If a file's formatting is different from asmfmt's, overwrite it
		with asmfmt's version.
```
You should only run `asmfmt` on files that are assembler files. Assembler files cannot be positively identified, so it will mangle non-assembler files.

# configuration

`asmfmt` supports a single TOML configuration file for formatting behavior.

You can load a config explicitly:

```bash
asmfmt -config /path/to/asmfmt.toml ./...
```

If `-config` is not provided, `asmfmt` automatically looks for the first matching file in this order:

1. Project config: the nearest `.asmfmt.toml` found by walking upward from the formatted file's directory.
2. User config: `~/.asmfmt.toml`
3. Global config: `/etc/asmfmt.toml`

Additional rules:

* Only the first matching config is used.
* Config files are not merged.
* Directory formatting reuses the config found from the starting directory for the entire walk.
* Standard input does not do project-directory lookup. It only checks `~/.asmfmt.toml` and `/etc/asmfmt.toml`, unless `-config` is provided.
* Unknown fields and invalid values are treated as errors.
* Unset fields keep the built-in defaults, so no config still matches historical behavior.

A fully commented reference config is available in [.asmfmt.toml.example](.asmfmt.toml.example).

# supported syntax

`asmfmt` primarily formats Go/Plan 9 style assembly, and also supports a growing subset of GAS-style syntax that is common in RISC-V and ELF-oriented assembler sources.

Supported and tested coverage includes:

* Go / Plan 9 assembler indentation and alignment rules.
* GAS directives used in RISC-V assembly, including `.section`, `.attribute`, `.option`, `.insn`, relocation helpers, CFI directives, data directives, and symbol visibility directives.
* GAS macro blocks such as `.macro`, `.irp`, `.irpc`, `.rept`, `.if`, `.else`, and matching end directives.
* GAS comments using `#`, `//`, block comments, and ARM/GAS `@` line comments when the source is detected as GAS-like.
* RISC-V operand forms such as `%hi(...)`, `%lo(...)`, `%pcrel_hi(...)`, `%pcrel_lo(...)`, local numeric labels like `1b`, CSR operands, vector operands, compressed mnemonics, and `.insn` encodings.

Unknown directives and unknown lowercase mnemonics are preserved conservatively. `asmfmt` will still align and indent surrounding code, but it does not try to validate or normalize unknown syntax beyond safe whitespace handling.

# non-goals

`asmfmt` is a formatter, not an assembler or linter.

It intentionally does not:

* validate opcode legality,
* reject unknown directives or vendor mnemonics,
* require binutils or another assembler at runtime,
* guarantee semantic equivalence through default tests,
* rewrite unrelated Plan 9 formatting behavior to follow GAS conventions.

# style detection

Dialect detection is internal and heuristic by default, but configuration can force a syntax mode.

Current style hints are used to distinguish:

* Plan 9 / Go assembler files,
* GAS-like files,
* RISC-V GAS files.

These hints affect whether `#` or `@` starts a comment, whether `;` should split statements, and whether lowercase mnemonics are treated as instruction-stream commands. Detection is intentionally conservative to avoid breaking existing Plan 9 formatting.

If you need deterministic behavior for a mixed codebase, set `source_style` in `.asmfmt.toml` to one of:

* `auto`
* `plan9`
* `gas`
* `riscv-gas`

# formatting

Default formatting behavior:

* Automatic indentation.
* Tabs for indentation and spaces for alignment.
* Remove trailing whitespace.
* Align the first parameter.
* Align end-of-line comments in a block.
* Eliminate repeated blank lines.
* Remove `;` at end of line.
* Insert a blank line before comment blocks, except when preceded by label or another comment block.
* Insert a blank line before labels, except when preceded by comment-only structure that should stay attached.
* Move labels to their own line, except for comment-only handling.
* Retain block breaks between logical sections.
* Convert single-line block comments to line comments.
* Add a space after line comment markers, except in preserved special cases such as `//go:build`.
* Keep a space between parameters.
* Track macros in the same file and exclude them from normal parameter indentation heuristics.
* Keep `TEXT`, `DATA`, `GLOBL`, `FUNCDATA`, `PCDATA`, and labels at level 0 indentation.
* Align `\` in multiline macros.
* Remove whitespace before separating `;`, and insert a space after `;` when followed by another instruction.

Supported configuration keys:

* `indent_style`: `tab` or `space`
* `indent_width`: positive integer, used only when `indent_style = "space"`
* `align_operands`: align the first operand column
* `align_comments`: align end-of-line comments
* `align_continuations`: align trailing `\` continuations
* `max_blank_lines`: maximum consecutive blank lines to preserve
* `split_semicolon_statements`: split `a; b` into separate statements when style permits
* `newline_before_comments`: insert formatter-managed blank lines before comment blocks
* `newline_before_labels`: insert formatter-managed blank lines before labels
* `labels_always_on_own_line`: rewrite `label: instruction` into separate lines
* `line_comment_space`: control whether `// comment` or `//comment` is emitted
* `convert_single_line_block_comment`: control whether `/* comment */` becomes `// comment`
* `preferred_comment_style`: `preserve` or `slash`
* `source_style`: `auto`, `plan9`, `gas`, or `riscv-gas`

# tests

Default verification stays self-contained:

* `go test ./...`
* `go vet ./...`
* `go test -run TestRewrite ./...`

Optional local corpus formatting checks can be enabled with:

* `ASMFMT_CORPUS_DIR=/path/to/corpus go test -run TestOptionalCorpus ./...`

This walks local assembly files and checks that formatting is idempotent. No network access is used.

Optional semantic equivalence checks can be enabled with an assembler and objdump:

* `ASMFMT_AS=riscv64-linux-gnu-as`
* `ASMFMT_OBJDUMP=riscv64-linux-gnu-objdump`
* `ASMFMT_ASFLAGS='-march=rv64gc -mabi=lp64d'`
* `go test -run TestOptionalSemanticEquivalence ./...`

These tests are skipped unless the environment is explicitly configured.

# adding fixtures

When extending syntax coverage:

* add one focused feature or syntax family per fixture when possible,
* always add both `testdata/name.in` and `testdata/name.golden`,
* use `go test -run TestRewrite -update` only for intentional output changes to existing fixtures,
* for brand-new fixtures, generate the initial `.golden` and then rerun `go test -run TestRewrite ./...` to confirm idempotence,
* remove any `*.asmfmt` diagnostics left behind by failing tests before committing.

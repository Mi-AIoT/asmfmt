# GAS/RISC-V Formatting Support Plan

## Goal

Bring `asmfmt` from lightweight RISC-V/GAS compatibility toward comprehensive GAS formatting support while keeping the formatter deterministic, idempotent, and non-validating.

The target is not to become an assembler. The formatter should preserve source semantics, avoid corrupting legal GAS syntax, and produce stable output on repeated formatting.

## Current Baseline

Implemented coverage:

- GAS line comments with `#` and `//`, plus block comments.
- Preprocessor lines such as `#define`, `#include`, `#if`, and `#endif`.
- Quote-aware and block-comment-aware comment detection.
- GAS-style comma splitting with parenthesis/bracket/brace depth tracking.
- Conservative semicolon statement splitting.
- GAS block indentation for `.macro`, `.irp`, `.irpc`, `.rept`, `.if`, `.else`, and matching end directives.
- Common section, symbol, data, object, alignment, `.option`, `.attribute`, `.insn`, and relocation formatting coverage.
- RISC-V base, extension, pseudo instruction, label, and relocation golden fixtures.

Known limits:

- No full GAS grammar or AST.
- No architecture mode selection.
- No assembler-based semantic equivalence testing.
- Macro and expression handling are still conservative text handling, not complete parsing.

## Non-Goals

- Do not validate opcode legality.
- Do not reject unknown directives or unknown instructions.
- Do not require binutils, cgo, or external assembler at runtime.
- Do not change `Format(io.Reader) ([]byte, error)` or CLI behavior.
- Do not rewrite unrelated Plan 9 / Go assembler formatting behavior.

## Phase 1: Parser Infrastructure

Objective: make the tokenizer explicit enough to support additional GAS features without adding more fragile string heuristics.

Tasks:

- Introduce token types for identifiers, directives, labels, numbers, strings, chars, comments, separators, operators, and raw text.
- Track lexical modes:
  - normal text
  - string literal
  - character literal
  - block comment
  - line comment
  - preprocessor line
  - macro body
  - altmacro body
- Preserve original spelling for all tokens so unknown syntax can be emitted unchanged.
- Add tokenizer unit tests for:
  - nested-looking comments inside strings
  - escaped quotes
  - character constants
  - line continuations
  - `#` comment versus immediate text
  - `@` as symbol/type marker versus target comment
  - `;` as separator versus comment-like text
- Keep the current formatter output unchanged for existing non-RISC-V fixtures.

Acceptance:

- `go test -count=1 ./...`
- `go vet ./...`
- Existing golden files remain stable unless a fixture documents a bug fix.

## Phase 2: Macro and Altmacro Support

Objective: handle GAS macro syntax well enough that formatting does not alter macro semantics.

Tasks:

- Parse `.macro` headers conservatively:
  - positional parameters
  - `name=default`
  - `name:req`
  - `name:vararg`
  - empty default values
- Track macro body state until `.endm`.
- Track `.altmacro` and `.noaltmacro` state.
- Support altmacro-sensitive text preservation for:
  - `\name`
  - `\()`
  - `&` concatenation
  - `%expr`
  - `LOCAL name`
- Preserve macro varargs containing commas.
- Preserve macro continuation lines and comments.
- Add fixtures for:
  - default macro args
  - required args
  - varargs with commas
  - nested `.if` inside macros
  - `.rept` inside macros
  - `.altmacro` string and token concatenation

Acceptance:

- Formatting macro-heavy fixtures is idempotent.
- Macro argument text is never reordered or normalized beyond spacing around top-level commas.

## Phase 3: GAS Expression Handling

Objective: avoid corrupting complex GAS expressions while improving spacing decisions around labels, relocations, and directive operands.

Tasks:

- Add a lightweight expression scanner for grouping, not semantic evaluation.
- Track expression nesting for:
  - parentheses
  - brackets
  - braces
  - relocation functions
  - symbol arithmetic
- Preserve operators and spacing in expressions unless a rule is explicitly tested.
- Cover literal forms:
  - decimal, hex, octal, binary where accepted by GAS
  - signed numbers
  - character constants
  - local numeric label references such as `1b` and `1f`
  - current-location symbol `.`
- Add fixtures for:
  - `. - symbol`
  - `symbol1 - symbol2 + (1 << 4)`
  - `%pcrel_lo(1b)(a0)`
  - nested relocation-like function calls
  - expression comments at boundaries

Acceptance:

- Top-level comma splitting remains correct.
- Expression content is preserved when no explicit formatting rule applies.

## Phase 4: Directive Coverage Expansion

Objective: cover the GAS directives that commonly appear in real RISC-V and ELF assembly.

Priority directives:

- Debug and location:
  - `.file`
  - `.loc`
  - `.line`
  - `.app-file`
- CFI:
  - `.cfi_startproc`
  - `.cfi_endproc`
  - `.cfi_def_cfa`
  - `.cfi_def_cfa_register`
  - `.cfi_def_cfa_offset`
  - `.cfi_offset`
  - `.cfi_restore`
  - `.cfi_adjust_cfa_offset`
  - `.cfi_remember_state`
  - `.cfi_restore_state`
  - `.cfi_sections`
  - `.cfi_signal_frame`
- Data and storage:
  - `.ascii`
  - `.space`
  - `.skip`
  - `.fill`
  - `.org`
  - `.incbin`
- Symbol and visibility:
  - `.hidden`
  - `.protected`
  - `.internal`
  - `.symver`
  - `.weakref`
  - `.equiv`
  - `.eqv`
- Struct-like layout:
  - `.struct`
  - `.offset`
  - `.endstruct`
- Section variants:
  - `.pushsection` with flags/type/group
  - `.section` with comdat/group forms
  - `.subsection`
  - `.previous`
- Miscellaneous:
  - `.include`
  - `.err`
  - `.error`
  - `.warning`
  - `.title`
  - `.sbttl`

Tasks:

- Classify directives into:
  - zero-indent control directives
  - indentation-affecting block directives
  - instruction-stream directives
  - data-emitting directives
  - unknown directives
- Add fixture groups by category rather than one huge fixture.
- Keep unknown `.foo` conservative: preserve text and comments, avoid semantic rewriting.

Acceptance:

- Each directive category has golden coverage.
- Unknown directive behavior is documented and tested.

## Phase 5: Architecture-Aware Comment and Separator Policy

Objective: reduce ambiguity for syntax that differs by target architecture while preserving current Plan 9 behavior.

Tasks:

- Add internal source-style detection, not public API changes:
  - Plan 9 / Go assembler style
  - GAS-like style
  - RISC-V GAS hints
- Use style hints to decide:
  - whether `#` can start a comment
  - whether `@` can start a comment
  - whether `;` is a separator or comment-like text
  - whether lowercase instructions are instruction-stream commands
- Add regression fixtures for:
  - ARM immediate `#1`
  - ELF `.type foo, @function`
  - ARM/GAS `@ comment`
  - RISC-V `;` statement separator
  - Plan 9 macro bodies with semicolons

Acceptance:

- Existing ARM and AMD64 Plan 9 fixtures remain stable.
- RISC-V/GAS fixtures retain intended formatting.

## Phase 6: RISC-V Coverage Completion

Objective: expand RISC-V formatting samples beyond common instructions into extension and operand edge cases.

Tasks:

- Add fixtures for ratified and widely used extensions:
  - Zicsr
  - Zifencei
  - Zba/Zbb/Zbc/Zbs
  - Zfh/Zfa where syntax appears in GAS
  - Vector extension load/store/mask combinations
  - compressed instruction variants
  - atomic acquire/release suffixes
- Add CSR operand variants:
  - symbolic CSR names
  - numeric CSR values
  - immediate CSR forms
- Add vector operand edge cases:
  - masks
  - tuples
  - segment loads/stores
  - indexed loads/stores
  - fault-only-first forms
  - policy operands
- Add custom/vendor examples:
  - custom instruction mnemonics
  - `.insn` custom encodings
  - unknown dotted mnemonics

Acceptance:

- Unknown or vendor mnemonics format as ordinary instruction-stream lines without validation errors.
- Operand text remains preserved.

## Phase 7: Real-World Corpus Testing

Objective: catch syntax patterns not covered by hand-written fixtures.

Tasks:

- Add a curated corpus under `testdata` or a separate ignored local corpus runner.
- Candidate sources:
  - binutils RISC-V GAS tests
  - Linux kernel RISC-V assembly snippets
  - compiler runtime assembly
  - embedded startup files
  - vendor SDK assembly files
- For each accepted corpus fixture:
  - keep source license compatible
  - minimize samples to focused fixtures where possible
  - add golden output and idempotence checks
- Add a script or Go test helper to run corpus formatting if files are present.

Acceptance:

- Corpus fixtures are deterministic.
- No network access is required during tests.
- Large external corpora are optional unless checked into the repository.

## Phase 8: Optional Semantic Verification

Objective: provide confidence that formatting preserves assembler output without making external tools a runtime dependency.

Tasks:

- Add an optional test mode gated by environment variable, for example `ASMFMT_AS=...`.
- For selected fixtures:
  - assemble original input
  - assemble formatted output
  - compare object metadata or disassembly
- Keep this outside default `go test ./...` unless the assembler is explicitly configured.
- Document required toolchain versions and target triples.

Acceptance:

- Default test suite remains self-contained.
- Optional semantic tests can be run in CI only when binutils is installed.

## Phase 9: Documentation

Objective: make supported behavior and limits clear to users and contributors.

Tasks:

- Update `README.md` with:
  - supported assembler dialects
  - RISC-V/GAS support summary
  - unknown directive behavior
  - non-validation policy
- Add a contributor note for adding fixtures:
  - one feature per fixture
  - always include `.in` and `.golden`
  - run `go test -run TestRewrite -update` only for intentional output changes
- Document style detection assumptions.

Acceptance:

- Users can tell what “GAS support” means.
- Contributors know how to add new syntax coverage safely.

## Suggested Implementation Order

1. Parser infrastructure cleanup.
2. Macro and altmacro fixtures and state handling.
3. CFI/debug directive coverage.
4. Expression scanner hardening.
5. Architecture-aware comment/separator policy.
6. RISC-V extension fixture expansion.
7. Real-world corpus import.
8. Optional assembler equivalence tests.
9. README and contributor documentation.

## Risk Areas

- Regressing existing Plan 9 formatting because GAS and Plan 9 use overlapping syntax.
- Treating `#`, `@`, or `;` incorrectly without an explicit architecture mode.
- Over-normalizing macro or expression text and changing assembler semantics.
- Adding large fixtures with unclear licensing.
- Making CI depend on external assembler availability.

## Done Criteria

The effort can be considered complete when:

- RISC-V/GAS fixtures cover macro, altmacro, expressions, debug/CFI, section/symbol/data directives, relocations, and common instruction families.
- Existing non-GAS golden fixtures remain stable.
- Formatting is idempotent across all fixtures.
- Unknown directives and unknown instructions are preserved without errors.
- Optional semantic tests, when enabled, show equivalent assembler output for selected real-world samples.

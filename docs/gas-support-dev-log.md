# GAS Support Development Log

This log tracks actual implementation progress against `docs/gas-support-plan.md`.

## 2026-05-21

### Step 0: Planning and Tracking

- Status: completed
- Scope:
  - created an executable checklist in `docs/gas-support-checklist.md`
  - initialized this development log
- Verification:
  - documentation-only change
- Commit:
  - `a5e289e` `docs: break down gas support roadmap`
- Notes:
  - further entries record code changes, tests run, and the resulting commit hash

### Step 1: Lexer Infrastructure

- Status: completed
- Scope:
  - added `lexer.go` with explicit token kinds and lexical modes
  - preserved original token spelling and byte offsets for all produced tokens
  - added `lexer_test.go` coverage for comment markers in strings, escaped quotes, character constants, line continuations, `#` immediates, `@function`, and semicolon splitting contexts
  - kept the formatter hot path unchanged after verifying that directly swapping helper logic introduced output drift
- Verification:
  - `go test ./...`
  - `go vet ./...`
- Commit:
  - `406226d` `refactor: add gas lexer scaffolding`
- Notes:
  - lexer is intentionally introduced as independent infrastructure first; formatter integration will proceed only in behavior-preserving slices

### Step 2: Macro and Altmacro Handling

- Status: completed
- Scope:
  - tracked `.macro`/`.endm` and `.altmacro`/`.noaltmacro` formatter state explicitly
  - preserved ordinary macro body lines as raw text so commas, semicolons, concatenation markers, and inline comments are not reinterpreted
  - merged `.macro` header text from the first `:vararg` parameter onward so comma-bearing vararg tails stay intact
  - added unit coverage for vararg header preservation and macro-body semicolon preservation
  - added `testdata/gas_macros.in` / `.golden` for default args, required args, varargs, nested `.if`, nested `.rept`, and altmacro-sensitive text
- Verification:
  - `go test -run TestRewrite -update`
  - `go test ./...`
  - `go vet ./...`
- Commit:
  - `bdf6146` `fix: preserve gas macro body text`
- Notes:
  - `testdata/riscv_gas.golden` intentionally changed inside a macro body: `lui  \reg, ...` became `lui \reg, ...` because macro body instructions are now preserved instead of alignment-normalized

### Step 3: Expression-Safe Operand Handling

- Status: completed
- Scope:
  - extracted parameter splitting into a dedicated scanner helper instead of embedding the depth-tracking loop directly inside `statement.setParams`
  - kept the scanner grouping-aware for parentheses, brackets, braces, strings, chars, block comments, and semicolon text normalization
  - expanded unit coverage for current-location expressions, relocation calls, signed numbers, local numeric labels, and character constants
  - added `testdata/gas_expressions.in` / `.golden` to lock down complex GAS expression formatting
- Verification:
  - `go test -run TestRewrite -update`
  - `go test ./...`
  - `go vet ./...`
- Commit:
  - `9615085` `refactor: extract gas expression parameter scanner`
- Notes:
  - a stronger case with inline block comments inside operand lists was intentionally deferred; current formatter still treats those comments as structural boundaries, so this step keeps the fixture focused on expression grouping rather than comment fusion

### Step 4: Directive Coverage Expansion

- Status: completed
- Scope:
  - expanded known GAS directive classification across debug/location, CFI, data/storage, symbol/visibility, section/misc, and struct-like directives
  - separated `known` from `zero-indent` so instruction-stream directives like `.insn` remain formatted in-stream while still avoiding the conservative unknown-directive path
  - added conservative raw-text handling for unknown directives based on the first dotted token
  - added grouped fixtures:
    - `testdata/gas_directives_debug.*`
    - `testdata/gas_directives_cfi.*`
    - `testdata/gas_directives_data.*`
    - `testdata/gas_directives_symbols.*`
    - `testdata/gas_directives_sections_misc.*`
    - `testdata/gas_directives_unknown.*`
  - added a focused unit test asserting unknown-directive text preservation
- Verification:
  - `go test -run TestRewrite -update`
  - `go test ./...`
  - `go vet ./...`
- Commit:
  - `8510ddd` `feat: expand gas directive coverage`
- Notes:
  - the first attempt over-classified `.insn` and `.ascii` as zero-indent directives; that was corrected before commit so existing in-stream indentation behavior stayed stable

## Current Progress Summary

- Completed roadmap increments:
  - Step 0 planning and tracking
  - Step 1 lexer infrastructure
  - Step 2 macro and altmacro handling
  - Step 3 expression-safe operand handling
  - Step 4 directive coverage expansion
- Remaining roadmap increments:
  - Step 5 source-style detection
  - Step 6 RISC-V fixture completion

### Step 5: Source-Style Detection

- Status: completed
- Scope:
  - added internal source-style detection for Plan 9, generic GAS, and RISC-V GAS
  - wired style hints into standalone comment-line handling, inline comment detection, and semicolon splitting
  - added explicit `@` line-comment support for GAS-style inputs without breaking `.type foo, @function`
  - kept public API unchanged; style stays an internal formatter state
  - added regression fixtures:
    - `testdata/gas_style_arm.*`
    - `testdata/gas_style_riscv_semicolon.*`
  - added unit coverage for style detection and Plan 9 semicolon policy
- Verification:
  - `go test -run TestRewrite -update`
  - `go test ./...`
  - `go vet ./...`
- Commit:
  - `2e6b81a` `fix: detect gas source style for comments`
- Notes:
  - Plan 9 behavior remains guarded by defaulting to the first strong style hint and only upgrading `gas` to `riscv-gas`

### Step 6: RISC-V Fixture Completion

- Status: completed
- Scope:
  - added focused fixture groups for CSR variants, additional extension mnemonics, vector operand combinations, and custom/vendor syntax
  - covered symbolic CSR names, numeric CSR operands, and immediate CSR forms
  - covered `fence.i`, bit-manip/arithmetic mnemonics, compressed forms, floating-point forms, and atomic acquire/release suffixes
  - covered vector masks, segment loads/stores, indexed loads/stores, fault-only-first forms, and policy operands
  - covered custom lowercase mnemonics and `.insn` encodings without adding validation logic
- Verification:
  - `go test ./...`
  - `go vet ./...`
  - `go test -run TestRewrite ./...`
- Commit:
  - `fd35d25` `test: extend riscv gas fixture coverage`
- Notes:
  - brand-new `.golden` files were generated with `go run ./cmd/asmfmt ...` and then verified for idempotence through `TestRewrite`

## Current Progress Summary

- Completed roadmap increments:
  - Step 0 planning and tracking
  - Step 1 lexer infrastructure
  - Step 2 macro and altmacro handling
  - Step 3 expression-safe operand handling
  - Step 4 directive coverage expansion
  - Step 5 source-style detection
  - Step 6 RISC-V fixture completion
- Remaining roadmap increments:
  - none from the tracked Step 0-6 execution checklist

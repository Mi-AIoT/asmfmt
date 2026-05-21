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
  - pending
- Notes:
  - `testdata/riscv_gas.golden` intentionally changed inside a macro body: `lui  \reg, ...` became `lui \reg, ...` because macro body instructions are now preserved instead of alignment-normalized

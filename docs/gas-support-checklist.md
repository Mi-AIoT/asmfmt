# GAS Support Execution Checklist

This document breaks `docs/gas-support-plan.md` into shippable increments that can be developed, tested, and committed independently.

## Step 0: Planning and Tracking

- [x] Create an execution checklist.
- [x] Create a development log.
- [x] Keep both documents updated after each landed step.

## Step 1: Lexer Infrastructure

Goal: introduce explicit lexical scanning without changing formatter output for existing fixtures.

Deliverables:

- [x] Add internal token kinds for identifiers, directives, labels, numbers, strings, chars, comments, separators, operators, and raw text.
- [x] Add lexical modes for normal text, string, char, block comment, line comment, and preprocessor line.
- [x] Preserve original token spelling for round-tripping.
- [x] Add focused lexer unit tests for:
  - [x] strings containing comment markers
  - [x] escaped quotes
  - [x] character constants
  - [x] line continuations
  - [x] `#` comment versus immediate text
  - [x] `@` operand text versus comment-like usage
  - [x] `;` separator versus literal text
- [x] Keep all current golden files stable.

Suggested commit split:

1. lexer types and scanner scaffolding
2. switch selected helpers to lexer-backed scanning
3. add lexer regression tests

## Step 2: Macro and Altmacro Handling

Goal: preserve GAS macro semantics while continuing to normalize only safe top-level spacing.

Deliverables:

- [x] Track `.macro` body state until `.endm`.
- [x] Parse macro headers conservatively:
  - [x] positional parameters
  - [x] `name=default`
  - [x] `name:req`
  - [x] `name:vararg`
  - [x] empty default values
- [x] Track `.altmacro` / `.noaltmacro`.
- [x] Preserve altmacro-sensitive text:
  - [x] `\name`
  - [x] `\()`
  - [x] `&` concatenation
  - [x] `%expr`
  - [x] `LOCAL name`
- [x] Preserve varargs containing commas.
- [x] Preserve continuation lines and comments inside macro bodies.
- [x] Add fixtures for default args, required args, varargs, nested `.if`, nested `.rept`, and `.altmacro` cases.

Suggested commit split:

1. macro state tracking
2. macro header splitting and vararg preservation
3. altmacro fixtures and regressions

## Step 3: Expression-Safe Operand Handling

Goal: avoid corrupting complex GAS expressions while keeping top-level comma splitting stable.

Deliverables:

- [x] Add a lightweight expression scanner for grouping.
- [x] Track nesting for parentheses, brackets, braces, and relocation-like calls.
- [x] Preserve expression spelling unless a specific formatting rule is covered by tests.
- [x] Add coverage for:
  - [x] decimal / hex / octal / binary
  - [x] signed numbers
  - [x] character constants
  - [x] `1b` / `1f`
  - [x] current-location symbol `.`
- [x] Add fixtures for `. - symbol`, symbol arithmetic, nested relocation calls, and comment boundaries.

Suggested commit split:

1. expression scanner reuse in operand splitting
2. expression fixtures

## Step 4: Directive Coverage Expansion

Goal: expand tested GAS directive support in small, category-based batches.

Deliverables:

- [x] Classify directives into:
  - [x] zero-indent control directives
  - [x] indentation-affecting block directives
  - [x] instruction-stream directives
  - [x] data-emitting directives
  - [x] unknown directives
- [x] Add grouped fixtures for:
  - [x] debug and location directives
  - [x] CFI directives
  - [x] data and storage directives
  - [x] symbol and visibility directives
  - [x] struct-like layout directives
  - [x] section variants
  - [x] miscellaneous directives
- [x] Document and test conservative unknown-directive behavior.

Suggested commit split:

1. directive classification
2. fixture groups for common directives
3. unknown-directive regression coverage

## Step 5: Source-Style Detection

Goal: distinguish Plan 9 / Go assembler inputs from GAS-like inputs without changing the public API.

Deliverables:

- [ ] Add internal style detection for Plan 9, GAS-like, and RISC-V GAS hints.
- [ ] Use style hints to decide `#`, `@`, and `;` treatment.
- [ ] Keep existing Plan 9 fixtures stable.
- [ ] Add regression fixtures for ARM immediates, `.type ... @function`, ARM `@` comments, and RISC-V semicolon separators.

Suggested commit split:

1. style detection
2. comment / separator policy adjustments
3. cross-architecture regression fixtures

## Step 6: RISC-V Fixture Completion

Goal: broaden real-world RISC-V coverage after the safer parsing pieces are in place.

Deliverables:

- [ ] Add extension and operand edge-case samples from real-world syntax.
- [ ] Reuse the earlier parser improvements rather than adding one-off heuristics.
- [ ] Keep formatting idempotent across all new samples.

Suggested commit split:

1. new fixture groups
2. targeted parser fixes only if fixtures expose bugs

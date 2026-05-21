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

- [ ] Track `.macro` body state until `.endm`.
- [ ] Parse macro headers conservatively:
  - [ ] positional parameters
  - [ ] `name=default`
  - [ ] `name:req`
  - [ ] `name:vararg`
  - [ ] empty default values
- [ ] Track `.altmacro` / `.noaltmacro`.
- [ ] Preserve altmacro-sensitive text:
  - [ ] `\name`
  - [ ] `\()`
  - [ ] `&` concatenation
  - [ ] `%expr`
  - [ ] `LOCAL name`
- [ ] Preserve varargs containing commas.
- [ ] Preserve continuation lines and comments inside macro bodies.
- [ ] Add fixtures for default args, required args, varargs, nested `.if`, nested `.rept`, and `.altmacro` cases.

Suggested commit split:

1. macro state tracking
2. macro header splitting and vararg preservation
3. altmacro fixtures and regressions

## Step 3: Expression-Safe Operand Handling

Goal: avoid corrupting complex GAS expressions while keeping top-level comma splitting stable.

Deliverables:

- [ ] Add a lightweight expression scanner for grouping.
- [ ] Track nesting for parentheses, brackets, braces, and relocation-like calls.
- [ ] Preserve expression spelling unless a specific formatting rule is covered by tests.
- [ ] Add coverage for:
  - [ ] decimal / hex / octal / binary
  - [ ] signed numbers
  - [ ] character constants
  - [ ] `1b` / `1f`
  - [ ] current-location symbol `.`
- [ ] Add fixtures for `. - symbol`, symbol arithmetic, nested relocation calls, and comment boundaries.

Suggested commit split:

1. expression scanner reuse in operand splitting
2. expression fixtures

## Step 4: Directive Coverage Expansion

Goal: expand tested GAS directive support in small, category-based batches.

Deliverables:

- [ ] Classify directives into:
  - [ ] zero-indent control directives
  - [ ] indentation-affecting block directives
  - [ ] instruction-stream directives
  - [ ] data-emitting directives
  - [ ] unknown directives
- [ ] Add grouped fixtures for:
  - [ ] debug and location directives
  - [ ] CFI directives
  - [ ] data and storage directives
  - [ ] symbol and visibility directives
  - [ ] struct-like layout directives
  - [ ] section variants
  - [ ] miscellaneous directives
- [ ] Document and test conservative unknown-directive behavior.

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

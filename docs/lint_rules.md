# Style Linter Rules Reference Manual

This document details all style linter rules available in `asmfmt`. These rules can be configured in the `[lint]` section of `.asmfmt.toml`.

## Category 1: Registers and Instructions (`L1xx`)

### L101: `abi_registers`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Enforce ABI register names over hardware register names to improve readability.
* **Non-compliant**:
  ```assembly
  addi x10, x11, 1
  ```
* **Compliant**:
  ```assembly
  addi a0, a1, 1
  ```

### L102: `compressed_instructions`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban compressed instruction mnemonics to keep code clean and uniform.
* **Non-compliant**:
  ```assembly
  c.mv a0, a1
  ```
* **Compliant**:
  ```assembly
  mv a0, a1
  ```

### L103: `operation_immediate`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban aliases for operation-with-immediate instructions.
* **Non-compliant**:
  ```assembly
  add a0, a1, 4
  ```
* **Compliant**:
  ```assembly
  addi a0, a1, 4
  ```

### L104: `relocation_operator_spacing`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban spaces between relocation operators and their opening parenthesis.
* **Non-compliant**:
  ```assembly
  lui a0, %hi (symbol)
  ```
* **Compliant**:
  ```assembly
  lui a0, %hi(symbol)
  ```

### L105: `gp_load_relaxation`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Require loading global pointer register under `.option norelax` to prevent incorrect self-relaxation.
* **Non-compliant**:
  ```assembly
  la gp, _gp
  ```
* **Compliant**:
  ```assembly
  .option push
  .option norelax
  la gp, _gp
  .option pop
  ```

### L106: `csr_names`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: CSR operands must use standard spec names or custom uppercase/`CSR_`-prefixed symbols. Raw numbers and lowercase custom names are banned.
* **Non-compliant**:
  ```assembly
  csrr a0, 0x300
  csrr a0, custom_csr
  ```
* **Compliant**:
  ```assembly
  csrr a0, mstatus
  csrr a0, MY_CSR
  csrr a0, CSR_custom
  ```

### L107: `jump_instruction_selection`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Ban using near jumps (`j`, `jal`, `jr`) to target non-local symbols. Use `call`/`tail` instead.
* **Non-compliant**:
  ```assembly
  j my_global_func
  ```
* **Compliant**:
  ```assembly
  tail my_global_func
  j .L_local_label
  ```

### L108: `pcrel_relocation_label`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Enforce that `%pcrel_lo` refers to a local label pointing to the matching `auipc` instruction.
* **Non-compliant**:
  ```assembly
  addi a0, a0, %pcrel_lo(global_sym)
  ```
* **Compliant**:
  ```assembly
  addi a0, a0, %pcrel_lo(.L_auipc_label)
  ```

---

## Category 2: Directive Usage (`L2xx`)

### L201: `alignment_directives`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban the `.align` directive in favor of `.p2align` or `.balign` for clarity.
* **Non-compliant**:
  ```assembly
  .align 4
  ```
* **Compliant**:
  ```assembly
  .p2align 2
  ```

### L202: `extern_directive`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban the `.extern` directive as it is redundant in GNU Assembler.
* **Non-compliant**:
  ```assembly
  .extern symbol
  ```
* **Compliant**:
  *(Omit the directive, symbols are external by default if undefined)*

### L203: `inline_binary_directives`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Ban `.word`/`.long` for inline binary data (favor `.byte`/`.2byte`/`.4byte`/`.8byte`).
* **Non-compliant**:
  ```assembly
  .word 0x12345678
  ```
* **Compliant**:
  ```assembly
  .4byte 0x12345678
  ```

### L204: `avoid_globl`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban the legacy `.globl` spelling in favor of `.global`.
* **Non-compliant**:
  ```assembly
  .globl main
  ```
* **Compliant**:
  ```assembly
  .global main
  ```

### L205: `leb128_constant_expression`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: `.uleb128` and `.sleb128` must only be used with constant expressions or symbol differences.
* **Non-compliant**:
  ```assembly
  .uleb128 my_symbol
  ```
* **Compliant**:
  ```assembly
  .uleb128 B - A
  .uleb128 42
  ```

### L206: `avoid_space_skip_directives`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban `.space` and `.skip` directives in favor of `.zero`.
* **Non-compliant**:
  ```assembly
  .space 16
  ```
* **Compliant**:
  ```assembly
  .zero 16
  ```

### L207: `operator_precedence_parentheses`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Require explicit parentheses to set operator precedence when mixing operators of different precedence levels in expressions.
* **Non-compliant**:
  ```assembly
  .long A * B >> 2
  ```
* **Compliant**:
  ```assembly
  .long A * (B >> 2)
  ```

### L208: `end_directive_last`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Enforce that the `.end` directive is the last statement in the file.
* **Non-compliant**:
  ```assembly
  .end
  addi a0, a0, 1
  ```
* **Compliant**:
  ```assembly
  addi a0, a0, 1
  .end
  ```

---

## Category 3: Structure and Formatting (`L3xx`)

### L301: `local_labels`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce that local labels start with `.L_` prefix.
* **Non-compliant**:
  ```assembly
  .Lloop:
  ```
* **Compliant**:
  ```assembly
  .L_loop:
  ```

### L302: `current_point_label`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban the current point `.` label usage in instruction operands.
* **Non-compliant**:
  ```assembly
  j .
  ```
* **Compliant**:
  ```assembly
  .L_loop:
      j .L_loop
  ```

### L303: `pointer_offset_shorthand`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce memory pointer offset shorthand syntax.
* **Non-compliant**:
  ```assembly
  lw a0, (a1)
  ```
* **Compliant**:
  ```assembly
  lw a0, 0(a1)
  ```

### L304: `option_push_pop`
* **Target Scope**: RISC-V Specific (`riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Enforce balanced `.option push` and `.option pop` directives.
* **Non-compliant**:
  ```assembly
  .option push
  .option norelax
  ```
* **Compliant**:
  ```assembly
  .option push
  .option norelax
  .option pop
  ```

### L305: `symbol_preamble_footer`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce correct symbol declarations (alignment directive preceding `.type`, and matching `.size` footer).
* **Non-compliant**:
  ```assembly
  .type my_func, @function
  my_func:
      ret
  ```
* **Compliant**:
  ```assembly
  .p2align 2
  .type my_func, @function
  my_func:
      ret
  .size my_func, .-my_func
  ```

### L306: `cfi_start_end_balance`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce balanced `.cfi_startproc` and `.cfi_endproc` directives.
* **Non-compliant**:
  ```assembly
  .cfi_startproc
  ret
  ```
* **Compliant**:
  ```assembly
  .cfi_startproc
  ret
  .cfi_endproc
  ```

### L307: `macro_balance`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce balanced `.macro` and `.endm` directives.
* **Non-compliant**:
  ```assembly
  .macro my_macro
  addi a0, a0, 1
  ```
* **Compliant**:
  ```assembly
  .macro my_macro
  addi a0, a0, 1
  .endm
  ```

### L308: `instruction_sequence_termination`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce that every function sequence or code block ends with a terminator instruction to prevent PC wandering.
* **Non-compliant**:
  ```assembly
  my_func:
      addi a0, a0, 1
  ```
* **Compliant**:
  ```assembly
  my_func:
      addi a0, a0, 1
      ret
  ```

### L309: `function_doxygen_comment`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce Doxygen-style comment blocks before global function declarations.
* **Non-compliant**:
  ```assembly
  .global my_func
  .type my_func, @function
  my_func:
  ```
* **Compliant**:
  ```assembly
  /**
   * @brief Increments a0
   */
  .global my_func
  .type my_func, @function
  my_func:
  ```

### L310: `avoid_hash_and_at_comments`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce comment syntaxes using `//` or `/* */` and ban `#` and `@` line comments.
* **Non-compliant**:
  ```assembly
  # this is a comment
  addi a0, a0, 1 @ increment
  ```
* **Compliant**:
  ```assembly
  // this is a comment
  addi a0, a0, 1 /* increment */
  ```

### L311: `double_label_declaration`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban consecutive labels on the same address without instructions or directives in between.
* **Non-compliant**:
  ```assembly
  label_a:
  label_b:
      addi a0, a0, 1
  ```
* **Compliant**:
  ```assembly
  label_a:
      addi a0, a0, 1
  label_b:
  ```

### L312: `reserved_label_names`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban developer-defined labels/macros that conflict with reserved keywords, registers, or instructions.
* **Non-compliant**:
  ```assembly
  text:
  addi:
  a0:
  ```
* **Compliant**:
  ```assembly
  my_label:
  ```

### L313: `unreachable_code`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Flag instructions located directly after a terminator instruction without a preceding label.
* **Non-compliant**:
  ```assembly
  ret
  addi a0, a0, 1
  ```
* **Compliant**:
  ```assembly
  ret
  .L_other_block:
      addi a0, a0, 1
  ```

### L314: `single_return_statement`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce that a function has at most one return instruction.
* **Non-compliant**:
  ```assembly
  my_func:
      beqz a0, .L_zero
      addi a0, a0, 1
      ret
  .L_zero:
      ret
  ```
* **Compliant**:
  ```assembly
  my_func:
      beqz a0, .L_zero
      addi a0, a0, 1
      j .L_out
  .L_zero:
      li a0, 0
  .L_out:
      ret
  ```

### L315: `no_fallthrough_to_function`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Ban code falling through into a function from a prior instruction sequence.
* **Non-compliant**:
  ```assembly
  func1:
      addi a0, a0, 1
  func2:
      ret
  ```
* **Compliant**:
  ```assembly
  func1:
      addi a0, a0, 1
      ret
  func2:
      ret
  ```

### L316: `no_jump_out_of_function`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban using a direct jump instruction (`j`, `jmp`) to jump out of a function's boundary.
* **Non-compliant**:
  ```assembly
  func1:
      j func2
  ```
* **Compliant**:
  ```assembly
  func1:
      tail func2
  ```

### L317: `no_recursive_calls`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Ban a function from calling itself recursively.
* **Non-compliant**:
  ```assembly
  my_func:
      call my_func
  ```
* **Compliant**:
  *(Implement recursion via iteration or explicit stack frames if needed)*

### L318: `copyright_and_license`
* **Target Scope**: All (`plan9` / `gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Every file must start with a copyright notice and an SPDX license identifier.
* **Non-compliant**:
  *(Empty file or file without license header)*
* **Compliant**:
  ```assembly
  // Copyright 2026 The Authors.
  // SPDX-License-Identifier: Apache-2.0
  ```

### L319: `developer_name_length`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `error`
* **Rationale**: Enforce that developer-defined names (labels, macros) are fewer than 31 characters.
* **Non-compliant**:
  ```assembly
  this_is_an_extremely_long_label_name:
  ```
* **Compliant**:
  ```assembly
  short_label:
  ```

### L320: `line_length_limit`
* **Target Scope**: All (`plan9` / `gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Enforce that every line contains at most 120 characters.
* **Non-compliant**:
  ```assembly
  // This line is extremely long and has more than one hundred and twenty characters which exceeds the style guide limits...
  ```
* **Compliant**:
  ```assembly
  // This line is shorter.
  ```

### L321: `label_naming_style`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Validate casing conventions of defined labels according to target style.
* **Options**: `"snake_case"`, `"camelCase"`, `"PascalCase"`, or `"any"`.
* **Non-compliant** (when style is `snake_case`):
  ```assembly
  myLabel:
  ```
* **Compliant** (when style is `snake_case`):
  ```assembly
  my_label:
  ```

### L322: `macro_naming_style`
* **Target Scope**: Generic GAS (`gas` / `riscv-gas`)
* **Default Severity**: `warning`
* **Rationale**: Validate casing conventions of preprocessor macro names.
* **Options**: `"UPPER_SNAKE_CASE"`, `"snake_case"`, or `"any"`.
* **Non-compliant** (when style is `UPPER_SNAKE_CASE`):
  ```assembly
  .macro my_macro
  ```
* **Compliant** (when style is `UPPER_SNAKE_CASE`):
  ```assembly
  .macro MY_MACRO
  ```

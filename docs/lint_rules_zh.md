# 代码风格检查规则参考手册

本文档详细介绍了 `asmfmt` 中所有可用的代码风格检查（Linter）规则。这些规则可以在 `.asmfmt.toml` 的 `[lint]` 部分中进行配置。

## 第一类：寄存器与指令 (`L1xx`)

### L101: `abi_registers`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 强制使用 ABI 寄存器名称，而不是硬件寄存器名称，以提高代码的可读性。
* **不合规示例**:
  ```assembly
  addi x10, x11, 1
  ```
* **合规示例**:
  ```assembly
  addi a0, a1, 1
  ```

### L102: `compressed_instructions`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止使用压缩指令助记符，以保持代码清洁和统一。
* **不合规示例**:
  ```assembly
  c.mv a0, a1
  ```
* **合规示例**:
  ```assembly
  mv a0, a1
  ```

### L103: `operation_immediate`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止使用带有立即数的通用操作指令别名。
* **不合规示例**:
  ```assembly
  add a0, a1, 4
  ```
* **合规示例**:
  ```assembly
  addi a0, a1, 4
  ```

### L104: `relocation_operator_spacing`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止在重定位操作符与其开括号之间留有空格。
* **不合规示例**:
  ```assembly
  lui a0, %hi (symbol)
  ```
* **合规示例**:
  ```assembly
  lui a0, %hi(symbol)
  ```

### L105: `gp_load_relaxation`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 要求在禁用链接器松弛（即在 `.option norelax` 下）时加载全局指针寄存器，以防止链接时的错误松弛。
* **不合规示例**:
  ```assembly
  la gp, _gp
  ```
* **合规示例**:
  ```assembly
  .option push
  .option norelax
  la gp, _gp
  .option pop
  ```

### L106: `csr_names`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `error`
* **原理**: CSR 操作数必须使用标准的规范名称或自定义的大写/以 `CSR_` 为前缀的符号。禁止使用原始数字常量或小写自定义名称。
* **不合规示例**:
  ```assembly
  csrr a0, 0x300
  csrr a0, custom_csr
  ```
* **合规示例**:
  ```assembly
  csrr a0, mstatus
  csrr a0, MY_CSR
  csrr a0, CSR_custom
  ```

### L107: `jump_instruction_selection`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 禁止使用近跳转（`j`、`jal`、`jr`）跳转到全局/非本地符号。应使用 `call`/`tail` 代替。
* **不合规示例**:
  ```assembly
  j my_global_func
  ```
* **合规示例**:
  ```assembly
  tail my_global_func
  j .L_local_label
  ```

### L108: `pcrel_relocation_label`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 强制要求 `%pcrel_lo` 重定位操作符必须引用一个本地标签（指向匹配的 `auipc` 指令），而不是全局符号名称。
* **不合规示例**:
  ```assembly
  addi a0, a0, %pcrel_lo(global_sym)
  ```
* **合规示例**:
  ```assembly
  addi a0, a0, %pcrel_lo(.L_auipc_label)
  ```

---

## 第二类：指令/伪指令使用 (`L2xx`)

### L201: `alignment_directives`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止使用 `.align` 指令，推荐使用 `.p2align` 或 `.balign` 以明确对齐字节。
* **不合规示例**:
  ```assembly
  .align 4
  ```
* **合规示例**:
  ```assembly
  .p2align 2
  ```

### L202: `extern_directive`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止使用 `.extern` 指令，因为它在 GNU 汇编器中是冗余的。
* **不合规示例**:
  ```assembly
  .extern symbol
  ```
* **合规示例**:
  *(直接省去该伪指令，未定义的符号默认即为外部符号)*

### L203: `inline_binary_directives`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 禁止使用 `.word`/`.long` 来定义内联二进制数据，推荐使用 `.byte`/`.2byte`/`.4byte`/`.8byte`。
* **不合规示例**:
  ```assembly
  .word 0x12345678
  ```
* **合规示例**:
  ```assembly
  .4byte 0x12345678
  ```

### L204: `avoid_globl`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止使用过时的 `.globl` 拼写，推荐使用 `.global`。
* **不合规示例**:
  ```assembly
  .globl main
  ```
* **合规示例**:
  ```assembly
  .global main
  ```

### L205: `leb128_constant_expression`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 强制要求 `.uleb128` 和 `.sleb128` 只能与常量表达式或符号差（如 `B - A`）一起使用，不能使用原始符号地址。
* **不合规示例**:
  ```assembly
  .uleb128 my_symbol
  ```
* **合规示例**:
  ```assembly
  .uleb128 B - A
  .uleb128 42
  ```

### L206: `avoid_space_skip_directives`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止使用 `.space` 和 `.skip` 指令，改用 `.zero` 进行零初始化。
* **不合规示例**:
  ```assembly
  .space 16
  ```
* **合规示例**:
  ```assembly
  .zero 16
  ```

### L207: `operator_precedence_parentheses`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 在表达式中混合使用不同优先级的操作符时，必须使用显式括号来确定优先级。
* **不合规示例**:
  ```assembly
  .long A * B >> 2
  ```
* **合规示例**:
  ```assembly
  .long A * (B >> 2)
  ```

### L208: `end_directive_last`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 强制要求 `.end` 指令是汇编文件中的最后一条语句。
* **不合规示例**:
  ```assembly
  .end
  addi a0, a0, 1
  ```
* **合规示例**:
  ```assembly
  addi a0, a0, 1
  .end
  ```

---

## 第三类：结构与格式化 (`L3xx`)

### L301: `local_labels`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制要求本地标签必须以 `.L_` 前缀开头。
* **不合规示例**:
  ```assembly
  .Lloop:
  ```
* **合规示例**:
  ```assembly
  .L_loop:
  ```

### L302: `current_point_label`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止在指令操作数中使用当前点 `.` 标签。
* **不合规示例**:
  ```assembly
  j .
  ```
* **合规示例**:
  ```assembly
  .L_loop:
      j .L_loop
  ```

### L303: `pointer_offset_shorthand`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制要求内存指针偏移量的简写语法（例如 `0(a0)`，而不是 `(a0)`）。
* **不合规示例**:
  ```assembly
  lw a0, (a1)
  ```
* **合规示例**:
  ```assembly
  lw a0, 0(a1)
  ```

### L304: `option_push_pop`
* **目标范围**: RISC-V 专用 (`riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 强制要求平衡 `.option push` 和 `.option pop` 指令。
* **不合规示例**:
  ```assembly
  .option push
  .option norelax
  ```
* **合规示例**:
  ```assembly
  .option push
  .option norelax
  .option pop
  ```

### L305: `symbol_preamble_footer`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制执行正确的符号声明序列（`.type` 之前有对齐指令，且有匹配的 `.size` 尾部）。
* **不合规示例**:
  ```assembly
  .type my_func, @function
  my_func:
      ret
  ```
* **合规示例**:
  ```assembly
  .p2align 2
  .type my_func, @function
  my_func:
      ret
  .size my_func, .-my_func
  ```

### L306: `cfi_start_end_balance`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制要求平衡的 `.cfi_startproc` 和 `.cfi_endproc` 指令。
* **不合规示例**:
  ```assembly
  .cfi_startproc
  ret
  ```
* **合规示例**:
  ```assembly
  .cfi_startproc
  ret
  .cfi_endproc
  ```

### L307: `macro_balance`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制要求平衡的 `.macro` 和 `.endm` 指令。
* **不合规示例**:
  ```assembly
  .macro my_macro
  addi a0, a0, 1
  ```
* **合规示例**:
  ```assembly
  .macro my_macro
  addi a0, a0, 1
  .endm
  ```

### L308: `instruction_sequence_termination`
* **Target Scope**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制要求每个函数序列或代码块以终止指令结束，以防止 PC 跑飞。
* **不合规示例**:
  ```assembly
  my_func:
      addi a0, a0, 1
  ```
* **合规示例**:
  ```assembly
  my_func:
      addi a0, a0, 1
      ret
  ```

### L309: `function_doxygen_comment`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制要求全局函数声明前有 Doxygen 风格的注释块。
* **不合规示例**:
  ```assembly
  .global my_func
  .type my_func, @function
  my_func:
  ```
* **合规示例**:
  ```assembly
  /**
   * @brief 递增 a0 寄存器
   */
  .global my_func
  .type my_func, @function
  my_func:
  ```

### L310: `avoid_hash_and_at_comments`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制使用 `//` 或 `/* */` 的注释语法，禁止使用 `#` 和 `@` 行注释。
* **不合规示例**:
  ```assembly
  # 这是一个注释
  addi a0, a0, 1 @ 递增
  ```
* **合规示例**:
  ```assembly
  // 这是一个注释
  addi a0, a0, 1 /* 递增 */
  ```

### L311: `double_label_declaration`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止在同一地址上声明连续的标签而中间没有指令或指令。
* **不合规示例**:
  ```assembly
  label_a:
  label_b:
      addi a0, a0, 1
  ```
* **合规示例**:
  ```assembly
  label_a:
      addi a0, a0, 1
  label_b:
  ```

### L312: `reserved_label_names`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止使用与保留关键字、指令或寄存器名称冲突的自定义标签/宏名称。
* **不合规示例**:
  ```assembly
  text:
  addi:
  a0:
  ```
* **合规示例**:
  ```assembly
  my_label:
  ```

### L313: `unreachable_code`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 标记直接位于终止指令之后且没有前导标签的指令。
* **不合规示例**:
  ```assembly
  ret
  addi a0, a0, 1
  ```
* **合规示例**:
  ```assembly
  ret
  .L_other_block:
      addi a0, a0, 1
  ```

### L314: `single_return_statement`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 强制要求一个函数最多只有一个返回指令。
* **不合规示例**:
  ```assembly
  my_func:
      beqz a0, .L_zero
      addi a0, a0, 1
      ret
  .L_zero:
      ret
  ```
* **合规示例**:
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
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 禁止代码从前一个指令序列直接落入（fall through）到一个新的函数中。
* **不合规示例**:
  ```assembly
  func1:
      addi a0, a0, 1
  func2:
      ret
  ```
* **合规示例**:
  ```assembly
  func1:
      addi a0, a0, 1
      ret
  func2:
      ret
  ```

### L316: `no_jump_out_of_function`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止使用直接跳转指令（`j`、`jmp`）跳转到当前函数边界之外的目标。
* **不合规示例**:
  ```assembly
  func1:
      j func2
  ```
* **合规示例**:
  ```assembly
  func1:
      tail func2
  ```

### L317: `no_recursive_calls`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 禁止函数直接递归调用自身。
* **不合规示例**:
  ```assembly
  my_func:
      call my_func
  ```
* **合规示例**:
  *(如需要，通过迭代或显式栈帧实现递归)*

### L318: `copyright_and_license`
* **目标范围**: 所有类型 (`plan9` / `gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 每个文件必须以版权声明和 SPDX 许可证标识符开头。
* **不合规示例**:
  *(空文件或无版权头部的汇编文件)*
* **合规示例**:
  ```assembly
  // Copyright 2026 The Authors.
  // SPDX-License-Identifier: Apache-2.0
  ```

### L319: `developer_name_length`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `error`
* **原理**: 强制要求开发者自定义的名称（标签、宏）少于 31 个字符。
* **不合规示例**:
  ```assembly
  this_is_an_extremely_long_label_name:
  ```
* **合规示例**:
  ```assembly
  short_label:
  ```

### L320: `line_length_limit`
* **目标范围**: 所有类型 (`plan9` / `gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 限制每行最多包含 120 个字符。
* **不合规示例**:
  ```assembly
  // 这是一行非常长的行，其包含的字符数量超过了一百二十个，超出了规范限制...
  ```
* **合规示例**:
  ```assembly
  // 这是一行较短的行。
  ```

### L321: `label_naming_style`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 验证自定义标签的命名大小写规范。
* **可配置项**: `"snake_case"`, `"camelCase"`, `"PascalCase"`, 或 `"any"`。
* **不合规示例** (配置为 `snake_case` 时):
  ```assembly
  myLabel:
  ```
* **合规示例** (配置为 `snake_case` 时):
  ```assembly
  my_label:
  ```

### L322: `macro_naming_style`
* **目标范围**: 通用 GAS (`gas` / `riscv-gas`)
* **默认严重级别**: `warning`
* **原理**: 验证预处理宏名称的大小写规范。
* **可配置项**: `"UPPER_SNAKE_CASE"`, `"snake_case"`, 或 `"any"`。
* **不合规示例** (配置为 `UPPER_SNAKE_CASE` 时):
  ```assembly
  .macro my_macro
  ```
* **合规示例** (配置为 `UPPER_SNAKE_CASE` 时):
  ```assembly
  .macro MY_MACRO
  ```

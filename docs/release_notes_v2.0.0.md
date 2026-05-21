# Release Notes: v2.0.0

This release marks a major milestone for `asmfmt`. Moving beyond the upstream `v1.3.2` codebase, `v2.0.0` introduces comprehensive GNU Assembler (GAS) and RISC-V syntax support, a highly customizable TOML configuration engine, semantic equivalence verification, and pre-built release automation.

---

## English Version

### 1. GNU Assembler (GAS) & RISC-V Support
* **RISC-V GAS Formatter**: Complete formatting coverage for RISC-V oriented assembly written in GAS syntax. (commit c78a973)
* **Expanded GAS Directives**: Full support for `.section`, `.attribute`, `.option`, `.insn`, CFI directives (e.g. `.cfi_startproc`, `.cfi_endproc`), data directives, and relocation helpers. (commit 8510ddd)
* **Advanced Comment Parsing**: Added support for GAS-specific comment markers: `#`, `//`, block comments (`/* ... */`), and ARM/GAS `@` line comments. (commit 2e6b81a)
* **Macro Preservation**: Robust support for GAS macro blocks (`.macro`, `.irp`, `.irpc`, `.rept`, `.if`, `.else`), preventing formatting from corrupting or flattening lines inside macro bodies. (commit bdf6146)

### 2. Extensible TOML Configuration Support
* **Project-Level `.asmfmt.toml` & Config Discovery**: Read formatting options from a TOML configuration file. Supports automatic configuration resolution walking upward from the file's directory, falling back to `~/.asmfmt.toml` and `/etc/asmfmt.toml`. (commit 169de5e)
* **New Layout Options**: Added new formatting configuration parameters including:
  * `indent_style` / `indent_width`: Customize tabs or spaces for indentation.
  * `align_operands` / `align_comments`: Toggle alignment for operands and comments.
  * `align_continuations`: Toggle alignment for trailing `\` continuation lines in macros.
  * `max_blank_lines`: Control consecutive blank lines to preserve.
  * `source_style`: Force formatting parser style (`auto`, `plan9`, `gas`, `riscv-gas`).
  * `indent_gas_directives`: Indent zero-indent GAS directives to code level. (commit ce13a2f)

### 3. Layout & Indentation Fixes
* **Refactored Comment Alignment**: Comment alignment in GAS mode is now calculated only against lines containing comments. This prevents long comment-less lines (like macro definitions) from pushing comments on other lines far to the right. (commit 3f0b272)
* **Preprocessor Block Alignment**: Restored correct indentation levels inside C-preprocess `#if` / `#else` / `#endif` blocks. (commits ca543bf, 2df112b)
* **Preserve Local Directives**: Fixed issues where zero-indent GAS directives inside macros or blocks were incorrectly de-indented. (commit 3614359)

### 4. CLI & Developer Tooling
* **`-init` Flag**: Automatically generate a default `.asmfmt.toml` configuration template in the current directory. (commit eabca18)
* **`-version` Flag & Usage Header**: Display version number, git hash, and build time (with automatic fallback to Go VCS build info). (commits f054a9e, ac34d8b)
* **`-cpuprofile` Clarification**: Updated flag description to specify debugging/profiling intent. (commit 8d65678)
* **Makefile Integration**: Added a root `Makefile` for streamlined building, testing, linting, and formatting. (commit ffa452d)

### 5. Documentation & Automated CI
* **Structured Manuals**: Added comprehensive English and Chinese user manuals (`docs/user_manual.md`, `docs/user_manual_zh.md`) covering all configurations and editor setups. (commits cb8cbf0, 8609047)
* **Semantic Verification**: Integrated RISC-V assembler binutils (`as` & `objdump`) semantic equivalence testing in CI, guaranteeing that formatting changes do not alter compile-time binary output. (commit cbba8d0)
* **Pre-built Beta Releases & CI Checkout Fix**: Automated GitHub Actions to compile and release pre-built snapshots of the `master` branch to the `beta` tag, fixing deep tag checkouts. (commits 3419687, 2ec9140)

---

## 中文版说明 (Chinese Version)

本版本是 `asmfmt` 维护分支的一个里程碑版本。相比于上游官方 `v1.3.2` 版本，`v2.0.0` 引入了对 GNU 汇编（GAS）和 RISC-V 语法的全面支持、高可定制性的 TOML 配置引擎、语义等价性自动化校验以及预构建发布流程。

### 1. GNU 汇编 (GAS) 与 RISC-V 语法支持
* **RISC-V GAS 格式化**: 针对以 GAS 语法编写的 RISC-V 汇编代码提供了完整的格式化覆盖。(commit c78a973)
* **扩展 GAS Directive**: 完整支持 `.section`、`.attribute`、`.option`、`.insn`、CFI 指令（如 `.cfi_startproc`、`.cfi_endproc`）、数据定义指令和重定位辅助指令。(commit 8510ddd)
* **高级注释解析**: 增加了对 GAS 专属注释标记 `#`、`//`、块注释（`/* ... */`）以及在检测为 GAS 格式时 ARM/GAS 风格的 `@` 行注释支持。(commit 2e6b81a)
* **宏代码块保护**: 对 GAS 宏指令块（`.macro`、`.irp`、`.irpc`、`.rept`、`.if`、`.else` 等）提供了稳健的解析保护，防止格式化破坏或折叠宏体内的代码结构。(commit bdf6146)

### 2. 灵活的 TOML 配置支持
* **项目级 `.asmfmt.toml` 与配置自动查找**: 支持从 TOML 配置文件中读取特定格式化规则。运行时自动从当前格式化文件目录向上逐级查找 `.asmfmt.toml`，并支持 fallback 到用户主目录 `~/.asmfmt.toml` 和系统目录 `/etc/asmfmt.toml`。(commit 169de5e)
* **丰富的新增排版选项**: 新增的一系列格式化控制选项包含：
  * `indent_style` / `indent_width`：自定义使用 tab 还是空格缩进及空格数量。
  * `align_operands` / `align_comments`：开关操作数和行尾注释对齐。
  * `align_continuations`：控制宏多行拼接符 `\` 的右对齐。
  * `max_blank_lines`：最大连续保留空行数。
  * `source_style`：强制指定解析器语法风格（`auto`、`plan9`、`gas`、`riscv-gas`）。
  * `indent_gas_directives`：控制是否将原零缩进的 GAS directive 缩进对齐到代码层级。(commit ce13a2f)

### 3. 排版与缩进修复
* **优化注释对齐**: 在 GAS 模式下，行尾注释的对齐位置只根据含有注释的行计算，防止未包含注释的长代码行（如宏定义）将其他行的行尾注释强行推至极右侧。(commit 3f0b272)
* **预处理块对齐**: 修复了 C 语言预处理 `#if` / `#else` / `#endif` 内部代码行对齐失效的问题，能够正确维持并恢复缩进层级。(commits ca543bf, 2df112b)
* **保留局部指令缩进**: 修复了宏或局部代码块内部零缩进 GAS 指令被错误取消缩进的排版缺陷。(commit 3614359)

### 4. 命令行与开发工具链
* **`-init` 选项**: 支持在当前目录下自动生成带完整选项注释的默认 `.asmfmt.toml` 模板配置文件。(commit eabca18)
* **`-version` 选项与使用说明标头**: 命令行输出版本号、Git Hash 和编译时间（支持 Go VCS 信息自动回退解析）。(commits f054a9e, ac34d8b)
* **`-cpuprofile` 描述优化**: 优化了帮助说明，明确说明该选项仅供开发调试和大规模格式化性能调优使用。(commit 8d65678)
* **Makefile 集成**: 新增根目录 `Makefile`，方便开发者进行本地构建、测试、格式化与静态检查。(commit ffa452d)

### 5. 完善的文档与 CI/CD 构建
* **中英文用户手册**: 编写了详尽的中英文用户手册（`docs/user_manual.md`、`docs/user_manual_zh.md`），解耦并简化了 README，提供了配置详解及主流编辑器集成指南。(commits cb8cbf0, 8609047)
* **语义等价性校验**: CI 流程中引入了基于 RISC-V 交叉编译工具链（`as` 和 `objdump`）的语义等价性自动化校验，保障格式化操作绝对不改变编译出的机器码。(commit cbba8d0)
* **Beta 版本自动发布及构建源历史修复**: 每次向 `master` 分支推送代码且 CI 跑通后，会自动将多平台预构建二进制产物发布并更新到 `beta` 预发布标签下，修复了浅拉取带来的版本号错误问题。(commits 3419687, 2ec9140)

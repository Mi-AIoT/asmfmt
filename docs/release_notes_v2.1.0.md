# Release Notes: v2.1.0

> **NOTE:** This is a version in development and has not been formally released yet.  
> **注意：** 本版本处于开发阶段，尚未正式发布。


This release introduces a comprehensive **assembler style linter** for RISC-V assembly code validation, a new self-upgrade feature (`-update` flag) with custom update server configurations, and improved CI/CD linting capability by returning a non-zero exit code when formatting differences are detected.

---

## English Version

### 1. Assembler Style Linter (NEW)
* **Comprehensive Linter Subsystem**: Added a complete code style linter that validates assembly code against best practices and style guidelines. Includes 30+ configurable rules organized into three categories: RISC-V specific (L1xx), Generic GAS (L2xx), and Structure/Formatting (L3xx). (commit a8c7289)
* **RISC-V Specific Rules**: Enforces ABI register naming, bans compressed instructions, validates relocation operator spacing, ensures GP load relaxation, and checks CSR naming conventions. (commits a8c7289)
* **Generic GAS Rules**: Enforces alignment directive usage, bans redundant directives, validates operator precedence, and checks symbol declaration patterns. (commits a8c7289)
* **Inline Control Comments**: Added support for per-line and per-file inline comment directives to selectively disable lint rules using `// asmfmt-disable-next-line L101` or `/* asmfmt-disable-line L101 */` syntax. Supports both C++ and assembly comment styles. (commits c632eb2, 91def2b)
* **Copyright & SPDX Validation**: Added `copyright_require_spdx` and `copyright_format` options to validate copyright headers and SPDX license identifiers. (commit 8b6ff02)
* **Declarative Macros Support**: Added `declarative_macros` config option to handle custom macros as directives for rule validation. (commit 229e2b1)

### 2. Bug Fixes
* **Unreachable Code Detection**: Fixed false positives for empty lines in rule L313, preventing blank lines from triggering unreachable code warnings. (commit feda76a)
* **Instruction Termination Reporting**: Fixed reporting of un-terminated sequences to point to the exact instruction line rather than the end of the function. (commit ff17d26)
* **GAS Directive Indentation**: Fixed `.extern` directive indentation when GAS directive config is enabled. (commit 0de823a)

### 3. In-Place Self-Upgrade Feature
* **Self-Upgrade CLI Flag**: Added the `-update <target>` CLI option to upgrade the `asmfmt` binary in-place to the `latest` version, a `beta` build, or a specific Release Tag (e.g. `v2.1.0`). (commit 60f8fb6)
* **Custom Environment Overrides**: Added support for overriding the upgrade repository via `ASMFMT_UPGRADE_REPO` (defaults to `Mi-AIoT/asmfmt`) and the GitHub API base URL via `ASMFMT_UPDATE_URL` (defaults to `https://api.github.com`), facilitating local tests and custom mirror usages. (commits 60f8fb6, b05cd38)

### 4. CI Verification Enhancements
* **Non-Zero Exit Code on Diff**: Updated the `-d` (diff) and `-l` (list files) flags to return a non-zero exit code (`1`) when any formatting differences are detected. This simplifies integration check steps in CI environments without requiring log parsing wrappers. (commit 15ec8fd)
* **Exit Code Test Coverage**: Added integration tests to verify successful/failed exit code behavior for `-d`, `-l`, `-w` and standard output modes under formatting differences. (commit 15ec8fd)

### 5. CI/CD & Build Improvements
* **Exclude Non-v Tags from Snapshot Versioning**: Deletes non-v-prefixed local tags (such as `beta`) before running GoReleaser in snapshot builds, ensuring that automated master branch beta releases retrieve the latest `v`-prefixed semantic version (e.g., `v2.0.0-next` instead of `beta-next`). (commit bc0c12d)
* **Automated Release Notes for Beta Builds**: Configured the beta release pipeline to automatically generate release notes using GitHub Releases API to show changes since the last tag. (commit 06f528e)
* **Dynamic Custom Release Notes**: Added capability to dynamically fetch release notes by checking for custom files at build time. (commit 98fbf48)

### 6. Bilingual Documentation
* **Lint Rules Manual**: Created comprehensive documentation for all 30+ style linter rules in both English (`docs/lint_rules.md`) and Chinese (`docs/lint_rules_zh.md`). (commits e4135d5, bcf6ab3, cb8d1e6, 042b1f6, 2cb48ff)
* **Manual & README Updates**: Documented the `-update` flag, exit code behavior changes, and environment variables bilingually in English and Chinese manuals (`docs/user_manual.md`, `docs/user_manual_zh.md`) and project READMEs. (commits 29f83c5, b05cd38)
* **Development Guidelines**: Updated `AGENTS.md` guidelines to specify that both English and Chinese versions of documentation must be updated concurrently whenever changes are made. (commits 29f83c5, 29f83c5)

---

## 中文版说明 (Chinese Version)

本版本为 `asmfmt` 引入了全新的**汇编代码风格检查器**、自更新功能（`-update` 选项），并支持自定义更新服务器配置。同时，本版本通过在检测到文件格式不一致时返回非零退出码，极大地增强了工具在 CI/CD 静态校验环境中的集成易用性。

### 1. 汇编代码风格检查器（新增）
* **完整 Linter 子系统**: 新增完整的代码风格检查器，用于验证汇编代码是否符合最佳实践和风格指南。包含 30+ 条可配置规则，分为三大类：RISC-V 专用规则 (L1xx)、通用 GAS 规则 (L2xx) 和结构/格式规则 (L3xx)。(commit a8c7289)
* **RISC-V 专用规则**: 检查 ABI 寄存器命名、禁止压缩指令、验证重定位操作符间距、确保 GP 寄存器加载时的松弛选项、检查 CSR 命名约定等。(commits a8c7289)
* **通用 GAS 规则**: 检查对齐指令用法、禁止冗余指令、验证运算符优先级、检查符号声明模式等。(commits a8c7289)
* **内联控制注释**: 支持使用内联注释指令逐行或逐文件选择性禁用规则，如 `// asmfmt-disable-next-line L101` 或 `/* asmfmt-disable-line L101 */` 语法。支持 C++ 和汇编注释风格。(commits c632eb2, 91def2b)
* **版权与 SPDX 验证**: 新增 `copyright_require_spdx` 和 `copyright_format` 选项，用于验证版权头和 SPDX 许可证标识符。(commit 8b6ff02)
* **声明式宏支持**: 新增 `declarative_macros` 配置选项，将自定义宏作为指令处理以进行规则验证。(commit 229e2b1)

### 2. Bug 修复
* **不可达代码检测**: 修复了 L313 规则中空行导致的误报问题，防止空行触发不可达代码警告。(commit feda76a)
* **指令终止报告**: 修复了未终止指令序列的报告位置，现在会指向准确的指令行而非函数末尾。(commit ff17d26)
* **GAS 指令缩进**: 修复了启用 GAS 指令配置时 `.extern` 指令的缩进问题。(commit 0de823a)

### 3. 就地自更新功能
* **命令行自更新标志**: 新增 `-update <target>` 命令行选项，支持就地在线将 `asmfmt` 二进制文件升级到 `latest`（最新发布版）、`beta`（测试版）或指定的 Release Tag（如 `v2.1.0`）。(commit 60f8fb6)
* **环境变量自定义配置**: 升级功能支持使用 `ASMFMT_UPGRADE_REPO`（默认为 `Mi-AIoT/asmfmt`）覆盖更新仓库源，以及使用 `ASMFMT_UPDATE_URL`（默认为 `https://api.github.com`）覆盖 GitHub API 地址，从而允许指向本地测试服务器或企业镜像源。(commits 60f8fb6, b05cd38)

### 4. CI 格式校验集成增强
* **格式差异返回非零退出码**: 更新了 `-d`（输出 diff 补丁）和 `-l`（仅列出差异文件）参数的行为。当检测到任何文件的排版格式与 `asmfmt` 标准不一致时，工具将以状态码 `1` 退出。这允许 CI 静态检查流水线直接阻断不合规的提交，而无需再依赖额外的日志文本过滤脚本。(commit 15ec8fd)
* **退出状态码测试覆盖**: 编写了完整的集成测试，全面验证了在存在格式差异时，`-d`、`-l`、`-w` 和标准输出模式下的正确/失败退出状态码表现。(commit 15ec8fd)

### 5. CI/CD 构建优化
* **排除非 v 开头标签干扰快照版本号**: 在构建测试版快照时，CI 流水线会在本地自动清除如 `beta` 这样非 `v` 开头的标签，以确保 GoReleaser 能够基于最新版本的正式 `v` 标签（如 `v2.0.0`）生成快照版本号（例如 `v2.0.0-next`，而非 `beta-next`）。(commit bc0c12d)
* **Beta 构建自动生成 Release Note**: 优化了 Beta 发布流程，启用 GitHub 自动 Release Note 接口，自动检索并呈现自上一个版本以来的 Commits 和 PRs。(commit 06f528e)
* **动态自定义 Release Notes**: 增加了在 CI 阶段动态检测并载入自定义发布日志的脚本逻辑。(commit 98fbf48)

### 6. 双语文档与规范
* **Linter 规则手册**: 为所有 30+ 条风格检查规则创建了完整的中英文文档（`docs/lint_rules.md` 和 `docs/lint_rules_zh.md`）。(commits e4135d5, bcf6ab3, cb8d1e6, 042b1f6, 2cb48ff)
* **README 及用户手册更新**: 同步更新了中英文的 README 及用户手册（`docs/user_manual.md`、`docs/user_manual_zh.md`），详尽阐述了自更新标志、退出码变更和相关环境变量的使用方法。(commits 29f83c5, b05cd38)
* **仓库开发指南完善**: 完善了 `AGENTS.md` 中的开发流程规范，规定每次更新文档时，**必须同时更新**中文和英文版本。(commits 29f83c5, 29f83c5)

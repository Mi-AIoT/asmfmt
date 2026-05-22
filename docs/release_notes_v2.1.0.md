# Release Notes: v2.1.0

This release introduces a new self-upgrade feature (`-update` flag) with custom update server configurations, and improves CI/CD linting capability by returning a non-zero exit code when formatting differences are detected.

---

## English Version

### 1. In-Place Self-Upgrade Feature
* **Self-Upgrade CLI Flag**: Added the `-update <target>` CLI option to upgrade the `asmfmt` binary in-place to the `latest` version, a `beta` build, or a specific Release Tag (e.g. `v2.1.0`). (commit 60f8fb6)
* **Custom Environment Overrides**: Added support for overriding the upgrade repository via `ASMFMT_UPGRADE_REPO` (defaults to `Mi-AIoT/asmfmt`) and the GitHub API base URL via `ASMFMT_UPDATE_URL` (defaults to `https://api.github.com`), facilitating local tests and custom mirror usages. (commits 60f8fb6, b05cd38)

### 2. CI Verification Enhancements
* **Non-Zero Exit Code on Diff**: Updated the `-d` (diff) and `-l` (list files) flags to return a non-zero exit code (`1`) when any formatting differences are detected. This simplifies integration check steps in CI environments without requiring log parsing wrappers. (commit 15ec8fd)
* **Exit Code Test Coverage**: Added integration tests to verify successful/failed exit code behavior for `-d`, `-l`, `-w` and standard output modes under formatting differences. (commit 15ec8fd)

### 3. CI/CD & Build Improvements
* **Exclude Non-v Tags from Snapshot Versioning**: Deletes non-v-prefixed local tags (such as `beta`) before running GoReleaser in snapshot builds, ensuring that automated master branch beta releases retrieve the latest `v`-prefixed semantic version (e.g., `v2.0.0-next` instead of `beta-next`). (commit bc0c12d)
* **Automated Release Notes for Beta Builds**: Configured the beta release pipeline to automatically generate release notes using GitHub Releases API to show changes since the last tag. (commit 06f528e)
* **Dynamic Custom Release Notes**: Added capability to dynamically fetch release notes by checking for custom files at build time. (commit 98fbf48)

### 4. Bilingual Documentation
* **Manual & README Updates**: Documented the `-update` flag, exit code behavior changes, and environment variables bilingually in English and Chinese manuals (`docs/user_manual.md`, `docs/user_manual_zh.md`) and project READMEs. (commits 29f83c5, b05cd38)
* **Development Guidelines**: Updated `AGENTS.md` guidelines to specify that both English and Chinese versions of documentation must be updated concurrently whenever changes are made. (commits 29f83c5, 29f83c5)

---

## 中文版说明 (Chinese Version)

本版本为 `asmfmt` 引入了全新的自更新功能（`-update` 选项），并支持自定义更新服务器配置。同时，本版本通过在检测到文件格式不一致时返回非零退出码，极大地增强了工具在 CI/CD 静态校验环境中的集成易用性。

### 1. 就地自更新功能
* **命令行自更新标志**: 新增 `-update <target>` 命令行选项，支持就地在线将 `asmfmt` 二进制文件升级到 `latest`（最新发布版）、`beta`（测试版）或指定的 Release Tag（如 `v2.1.0`）。(commit 60f8fb6)
* **环境变量自定义配置**: 升级功能支持使用 `ASMFMT_UPGRADE_REPO`（默认为 `Mi-AIoT/asmfmt`）覆盖更新仓库源，以及使用 `ASMFMT_UPDATE_URL`（默认为 `https://api.github.com`）覆盖 GitHub API 地址，从而允许指向本地测试服务器或企业镜像源。(commits 60f8fb6, b05cd38)

### 2. CI 格式校验集成增强
* **格式差异返回非零退出码**: 更新了 `-d`（输出 diff 补丁）和 `-l`（仅列出差异文件）参数的行为。当检测到任何文件的排版格式与 `asmfmt` 标准不一致时，工具将以状态码 `1` 退出。这允许 CI 静态检查流水线直接阻断不合规的提交，而无需再依赖额外的日志文本过滤脚本。(commit 15ec8fd)
* **退出状态码测试覆盖**: 编写了完整的集成测试，全面验证了在存在格式差异时，`-d`、`-l`、`-w` 和标准输出模式下的正确/失败退出状态码表现。(commit 15ec8fd)

### 3. CI/CD 构建优化
* **排除非 v 开头标签干扰快照版本号**: 在构建测试版快照时，CI 流水线会在本地自动清除如 `beta` 这样非 `v` 开头的标签，以确保 GoReleaser 能够基于最新版本的正式 `v` 标签（如 `v2.0.0`）生成快照版本号（例如 `v2.0.0-next`，而非 `beta-next`）。(commit bc0c12d)
* **Beta 构建自动生成 Release Note**: 优化了 Beta 发布流程，启用 GitHub 自动 Release Note 接口，自动检索并呈现自上一个版本以来的 Commits 和 PRs。(commit 06f528e)
* **动态自定义 Release Notes**: 增加了在 CI 阶段动态检测并载入自定义发布日志的脚本逻辑。(commit 98fbf48)

### 4. 双语文档与规范
* **README 及用户手册更新**: 同步更新了中英文的 README 及用户手册（`docs/user_manual.md`、`docs/user_manual_zh.md`），详尽阐述了自更新标志、退出码变更和相关环境变量的使用方法。(commits 29f83c5, b05cd38)
* **仓库开发指南完善**: 完善了 `AGENTS.md` 中的开发流程规范，规定每次更新文档时，**必须同时更新**中文和英文版本。(commits 29f83c5, 29f83c5)

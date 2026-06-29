# Release Notes: v2.2.0

> **NOTE:** This is a version in development and has not been formally released yet.  
> **注意：** 本版本处于开发阶段，尚未正式发布。

---

## English Version

### CI/CD Workflow Improvements

* **Release Tag Validation**: Added validation to ensure release tags follow semantic version format (vX.Y.Z) and rejects non-version tags to prevent accidental releases. (commit 12e9d02)
* **Beta Release Concurrency Control**: Added concurrency control to beta deployment workflow to prevent race conditions when multiple master pushes occur simultaneously. (commit 12e9d02)
* **Enhanced Beta Release Notes**: Beta releases now automatically include:
  - Commit history since latest release tag in `hash: message` format
  - Latest development release notes from docs/ directory
  Provides full visibility into all changes in development. (commit 97c4531)
* **Beta Release Timestamp Fix**: Fixed beta release timestamp not updating by deleting existing release before creating a new one. This ensures GitHub displays the correct updated timestamp instead of the original creation time. (commit 5225c28)
* **Release Process Documentation**: Added comprehensive release process guidelines to AGENTS.md covering:
  - Release note location, naming conventions, and bilingual format
  - Development vs released status tracking
  - Bilingual documentation requirements
  - Post-release workflow procedures. (commit f5af748)

### Documentation

* **Release Process Guidelines**: Documented standard release workflow, tag validation requirements, and documentation synchronization rules. (commit f5af748)

---

## 中文版说明 (Chinese Version)

### CI/CD 工作流改进

* **Release Tag 验证**: 新增 tag 格式验证，确保 release tag 遵循语义化版本格式 (vX.Y.Z)，拒绝非版本标签以防止意外发布。(commit 12e9d02)
* **Beta Release 并发控制**: 新增 beta 部署工作流的并发控制，防止多个 master 推送同时发生时的竞态条件。(commit 12e9d02)
* **增强的 Beta Release Notes**: Beta releases 现在自动包含：
  - 自最新 release tag 以来的 commit 历史，格式为 `hash: message`
  - docs/ 目录中的最新开发中 release notes
  提供开发中所有更改的完整可见性。(commit 97c4531)
* **Beta Release 发布时间修复**: 修复 beta release 发布时间不更新的问题。通过在创建新 release 前先删除现有的 beta release，确保 GitHub 显示正确的更新时间而非原始创建时间。(commit 5225c28)
* **Release 流程文档**: 在 AGENTS.md 中添加了全面的 release 流程指南，涵盖：
  - Release note 位置、命名约定和双语格式
  - 开发中 vs 已发布状态跟踪
  - 双语文档要求
  - 发布后工作流程序。(commit f5af748)

### 文档

* **Release 流程指南**: 记录了标准 release 工作流、tag 验证要求和文档同步规则。(commit f5af748)

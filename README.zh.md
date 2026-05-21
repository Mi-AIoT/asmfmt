# asmfmt
Go 汇编格式化工具

[English](README.md) | 简体中文

`asmfmt` 会像 `gofmt` 处理 Go 代码一样，对 Go 汇编源码进行格式化。

介绍文章：[asmfmt: Go Assembler Formatter](https://blog.klauspost.com/asmfmt-assembler-formatter/)

[![Go Reference](https://pkg.go.dev/badge/klauspost/asmfmt.svg)](https://pkg.go.dev/klauspost/asmfmt)
[![Go](https://github.com/klauspost/asmfmt/actions/workflows/go.yml/badge.svg)](https://github.com/klauspost/asmfmt/actions/workflows/go.yml)

可以查看 [Example 1](https://files.klauspost.com/diff.html)、[Example 2](https://files.klauspost.com/diff2.html)、[Example 3](https://files.klauspost.com/diff3.html)，或者直接对比 [testdata](https://github.com/klauspost/asmfmt/tree/master/testdata) 目录中的示例。

状态：`STABLE`。除非修复 bug，否则格式不会随意变化。欢迎反馈问题。

# 安装

可执行文件可以从 [Releases](https://github.com/klauspost/asmfmt/releases) 下载，解压后放到可执行路径中即可。

如果通过 Go 安装命令行版本：

```bash
go install github.com/klauspost/asmfmt/cmd/asmfmt@latest
```

# 更新记录

* 2021-04-08：补充 module 信息，移除非主工具内容。
* 2021-01-06：修复行注释前出现 C 风格注释时的问题，例如 `VPCMPEQB Y8/*(DI)*/, Y0, Y1 // comment...`
* 2016-08-08：修复非缩进行前注释被错误缩进的问题。
* 2016-06-10：修复行尾注释中包含 `/*` 时的崩溃问题。
* 2016-04-14：修复宏定义中多行块注释结尾处理问题。
* 2016-04-14：工具升级到 Go 1.5+。
* 2015-12-21：修复宏定义中分号前的空格清理。
* 2015-12-21：修复宏定义中的行注释处理（仅 Go 1.5 生效）。
* 2015-12-17：改进行注释对齐效果。
* 2015-12-17：清理一行多条指令中的分号格式。

# Goland

如果要在 Goland 中配置自定义 File Watcher：

* 打开 `Settings -> Tools -> File Watchers`
* 点击 `+`，选择 `<custom>` 模板
* 名称填写 `asmfmt`
* `File Type` 选择 `x86 Plan 9 Assembly file`（会应用到所有平台）
* `Scope` 选择 `Project Files`
* `Arguments` 填写 `$FilePath$`
* `Output Paths to Refresh` 填写 `$FilePath$`
* `Working Directory` 填写 `$ProjectFileDir$`

高级选项中启用：

* [x] Trigger the watcher regardless of syntax errors
* [x] Create output file from stdout

其余选项关闭即可。

![Goland Configuration](https://user-images.githubusercontent.com/5663952/114158973-96eebc80-9925-11eb-9aea-703ce474a7bb.png)

# Emacs

如果要在 `.emacs` 中保存前自动格式化汇编文件，可以加入：

```elisp
(defun asm-mode-setup ()
  (set (make-local-variable 'gofmt-command) "asmfmt")
  (add-hook 'before-save-hook 'gofmt nil t)
)

(add-hook 'asm-mode-hook 'asm-mode-setup)
```

# 用法

`asmfmt [flags] [path ...]`

命令行参数整体风格与 `gofmt` 类似，但只会处理 `.s` 文件：

```text
	-config string
		从这个 TOML 文件读取格式化配置。
	-d
		不直接输出格式化结果。
		如果文件格式与 asmfmt 输出不同，则输出 diff。
	-e
		打印所有错误（而不是只打印前 10 个不同行的错误）。
	-l
		不直接输出格式化结果。
		如果文件格式与 asmfmt 输出不同，则输出文件名。
	-w
		不直接输出格式化结果。
		如果文件格式与 asmfmt 输出不同，则覆盖原文件。
```

只应对真正的汇编文件运行 `asmfmt`。汇编文件无法被绝对准确地识别，所以它会破坏非汇编文件的内容。

# 配置文件

`asmfmt` 支持通过单个 TOML 配置文件控制格式行为。

显式指定配置文件：

```bash
asmfmt -config /path/to/asmfmt.toml ./...
```

如果没有提供 `-config`，`asmfmt` 会按如下优先级查找第一份匹配的配置：

1. 项目级配置：从被格式化文件所在目录开始向上查找最近的 `.asmfmt.toml`
2. 用户级配置：`~/.asmfmt.toml`
3. 全局配置：`/etc/asmfmt.toml`

补充规则：

* 只使用第一份命中的配置文件。
* 不会合并多份配置。
* 对目录执行格式化时，会先基于起始目录确定配置，然后整次遍历复用同一份配置。
* 对标准输入，不做项目目录查找；只会检查 `~/.asmfmt.toml` 和 `/etc/asmfmt.toml`，除非显式传入 `-config`。
* 未知字段、非法枚举值、非法数字都会直接报错。
* 未设置的字段会继续使用内置默认值，因此没有配置文件时行为与历史版本一致。

仓库中提供了带详细注释的参考配置：[.asmfmt.toml.example](.asmfmt.toml.example)。

# 支持的语法

`asmfmt` 主要面向 Go / Plan 9 风格汇编，同时也支持一部分常见的 GAS 风格语法，尤其是 RISC-V 和 ELF 场景中常用的那部分。

已支持并覆盖测试的内容包括：

* Go / Plan 9 汇编的缩进和对齐规则
* RISC-V 汇编常见 GAS 指令伪操作，例如 `.section`、`.attribute`、`.option`、`.insn`、重定位辅助、CFI 指令、数据伪操作、符号可见性相关伪操作
* GAS 宏块语法，例如 `.macro`、`.irp`、`.irpc`、`.rept`、`.if`、`.else` 及其对应结束指令
* GAS 注释形式，包括 `#`、`//`、块注释，以及在检测为 GAS 风格时支持 ARM/GAS 的 `@` 行注释
* RISC-V 操作数形式，例如 `%hi(...)`、`%lo(...)`、`%pcrel_hi(...)`、`%pcrel_lo(...)`、`1b` 这类局部数字标签、CSR 操作数、向量寄存器操作数、压缩指令助记符和 `.insn` 编码格式

未知的 directive 和未知的小写助记符会尽量保守保留。`asmfmt` 会继续整理周围的缩进和空白，但不会尝试做超出安全范围的语法归一化或合法性判断。

# 非目标

`asmfmt` 是 formatter，不是 assembler，也不是 linter。

明确不做：

* opcode 合法性校验
* 拒绝未知 directive 或厂商自定义助记符
* 运行时依赖 binutils 或其他 assembler
* 默认测试中保证语义完全等价
* 为了贴合 GAS 约定而重写无关的 Plan 9 既有格式行为

# 风格检测

默认情况下，方言检测是内部启发式逻辑，但现在也可以通过配置强制指定语法模式。

当前会区分：

* Plan 9 / Go 汇编文件
* GAS 风格文件
* RISC-V GAS 文件

这些判断会影响：

* `#` 或 `@` 是否被当作注释起始符
* `;` 是否应该拆分为多条语句
* 小写助记符是否按指令流命令来处理

检测逻辑依然保持保守，以避免破坏已有的 Plan 9 格式行为。

如果需要稳定、可预测的行为，可以在 `.asmfmt.toml` 中设置 `source_style` 为以下之一：

* `auto`
* `plan9`
* `gas`
* `riscv-gas`

# 格式化行为

默认格式化行为包括：

* 自动缩进
* 使用 tab 做缩进，使用空格做对齐
* 移除行尾空白
* 对齐第一列操作数
* 对齐同一块中的行尾注释
* 消除多余空行
* 移除行尾多余的 `;`
* 在注释块前插入空行，但不会在标签后或注释块后重复插入无意义空行
* 在标签前插入空行，以增强基本块的可读性
* 将 `label: instruction` 形式的标签拆到独立一行
* 保留逻辑块之间的分隔
* 将单行块注释转换为行注释
* 在行注释标记后补一个空格，`//go:build` 这类特殊情形会保留原始形式
* 参数之间保持统一空格
* 跟踪同一文件中的宏定义，避免它们影响普通参数对齐
* `TEXT`、`DATA`、`GLOBL`、`FUNCDATA`、`PCDATA` 和 label 保持 0 级缩进
* 对齐多行宏中的 `\`
* 去掉 `;` 前多余空白；如果后面还有下一条语句，则补一个空格

支持的配置项：

* `indent_style`：`tab` 或 `space`
* `indent_width`：正整数，仅在 `indent_style = "space"` 时生效
* `align_operands`：是否对齐第一列操作数
* `align_comments`：是否对齐行尾注释
* `align_continuations`：是否对齐多行续行反斜杠 `\`
* `max_blank_lines`：允许保留的最大连续空行数
* `split_semicolon_statements`：是否将 `a; b` 拆成多条语句
* `newline_before_comments`：是否在注释块前插入 formatter 管理的空行
* `newline_before_labels`：是否在标签前插入 formatter 管理的空行
* `labels_always_on_own_line`：是否把 `label: instruction` 改写为多行
* `line_comment_space`：控制输出 `// comment` 还是 `//comment`
* `convert_single_line_block_comment`：控制是否把 `/* comment */` 转成 `// comment`
* `preferred_comment_style`：`preserve` 或 `slash`
* `source_style`：`auto`、`plan9`、`gas`、`riscv-gas`

# 测试

默认验证命令：

* `go test ./...`
* `go vet ./...`
* `go test -run TestRewrite ./...`

可选的本地语料格式化测试：

* `ASMFMT_CORPUS_DIR=/path/to/corpus go test -run TestOptionalCorpus ./...`

这个测试会遍历本地汇编文件，并验证格式化结果是幂等的，不会使用网络。

可选的语义等价性测试：

* `ASMFMT_AS=riscv64-linux-gnu-as`
* `ASMFMT_OBJDUMP=riscv64-linux-gnu-objdump`
* `ASMFMT_ASFLAGS='-march=rv64gc -mabi=lp64d'`
* `go test -run TestOptionalSemanticEquivalence ./...`

只有在显式设置这些环境变量时，这些测试才会运行，否则会自动跳过。

# 新增 fixture

扩展语法覆盖时，建议遵循以下规则：

* 尽量每个 fixture 只聚焦一个特性或一类语法
* 始终同时添加 `testdata/name.in` 和 `testdata/name.golden`
* 只有在明确要修改既有输出时，才使用 `go test -run TestRewrite -update`
* 新增 fixture 时，先生成初始 `.golden`，再重新运行 `go test -run TestRewrite ./...` 验证幂等性
* 测试失败后遗留的 `*.asmfmt` 诊断文件，在提交前要删除

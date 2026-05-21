# asmfmt

[![CI](https://github.com/Mi-AIoT/asmfmt/actions/workflows/go.yml/badge.svg)](https://github.com/Mi-AIoT/asmfmt/actions/workflows/go.yml)
[![Release](https://github.com/Mi-AIoT/asmfmt/actions/workflows/release.yml/badge.svg)](https://github.com/Mi-AIoT/asmfmt/actions/workflows/release.yml)

适用于 Go / Plan 9 汇编、GAS 语法和 RISC-V 代码的汇编格式化工具

[English](README.md) | 简体中文

`asmfmt` 会像 `gofmt` 处理 Go 代码一样，对汇编源码进行格式化。

它主要面向 Go / Plan 9 风格汇编，同时也支持一部分 ELF 场景中常见的
GAS 风格语法，包括 RISC-V。

状态：`STABLE`。默认输出应保持稳定，除非是在修复 bug。

这个仓库是 [`klauspost/asmfmt`](https://github.com/klauspost/asmfmt) 的活跃
维护 fork。在尽量保持原始格式化行为的前提下，继续为上游多年未推进的使用
场景演进能力。

## 为什么用 asmfmt

如果你希望汇编代码更易读、更容易 review，并在代码库内保持统一风格，
`asmfmt` 会比较合适。

它可以：

* 统一缩进，
* 对齐操作数和行尾注释，
* 清理空行和分号分隔语句，
* 在没有配置文件时保持历史默认行为，
* 既能作为 CLI 使用，也能作为 Go 库使用。

## 快速开始

安装 CLI：

```bash
go install github.com/Mi-AIoT/asmfmt/cmd/asmfmt@latest
```

如果你要把这个 fork 当作 Go module 依赖，请先确认当前 `go.mod` 里的
module path。CLI 的安装路径和库的 import path 目前可能还不一致，直到
module path 从上游路径迁移出来为止。

就地格式化文件：

```bash
asmfmt -w path/to/file.s
```

以 diff 形式预览修改：

```bash
asmfmt -d path/to/file.s
```

格式化整个目录树：

```bash
asmfmt ./...
```

`asmfmt` 只处理 `.s` 文件。不要把它用于非汇编输入。

## 示例

输入：

```asm
TEXT foo(SB),$0
MOVQ AX,BX //comment
loop:ADDQ $1,AX;JMP loop
```

输出：

```asm
TEXT foo(SB), $0
	MOVQ AX, BX // comment

loop:
	ADDQ $1, AX
	JMP  loop
```

更完整的示例可以看
[testdata](https://github.com/Mi-AIoT/asmfmt/tree/master/testdata)。

## CLI 用法

```text
asmfmt [flags] [path ...]
```

常用参数：

* `-w` 直接覆盖原文件
* `-d` 输出 diff，而不是直接输出格式化后的源码
* `-l` 输出格式与 `asmfmt` 不一致的文件名
* `-e` 报告全部错误，而不是在前 10 个不同位置错误后停止
* `-config` 从指定 TOML 文件加载格式化选项

如果不提供路径，`asmfmt` 会从标准输入读取，并把格式化结果写到标准输出。
`-w` 不能和标准输入一起使用。

## 配置

`asmfmt` 支持单个 TOML 配置文件。

显式指定配置文件：

```bash
asmfmt -config /path/to/asmfmt.toml ./...
```

如果没有提供 `-config`，会按以下顺序使用第一份命中的配置：

1. 从被格式化文件所在目录向上查找最近的 `.asmfmt.toml`
2. `~/.asmfmt.toml`
3. `/etc/asmfmt.toml`

规则：

* 一次只使用一个配置文件。
* 多份配置不会合并。
* 未知字段和非法值都会报错。
* 未设置的字段会保留内置默认值。
* 对标准输入，除非显式传入 `-config`，否则不会做项目目录查找。

最小示例：

```toml
indent_style = "space"
indent_width = 4
source_style = "riscv-gas"
align_comments = true
```

支持的配置项：

* `indent_style`：`tab` 或 `space`
* `indent_width`：正整数，仅在 `indent_style = "space"` 时生效
* `align_operands`
* `align_comments`
* `align_continuations`
* `max_blank_lines`
* `split_semicolon_statements`
* `newline_before_comments`
* `newline_before_labels`
* `labels_always_on_own_line`
* `line_comment_space`
* `convert_single_line_block_comment`
* `preferred_comment_style`：`preserve` 或 `slash`
* `source_style`：`auto`、`plan9`、`gas` 或 `riscv-gas`

完整注释版配置参考见 [.asmfmt.toml.example](.asmfmt.toml.example)。

## 支持的语法

`asmfmt` 主要面向 Go / Plan 9 汇编，同时也支持一部分保守处理的 GAS 风格语法。

已覆盖并测试的范围包括：

* Go / Plan 9 汇编的缩进和对齐规则
* RISC-V 汇编中常见的 GAS directive，例如 `.section`、`.attribute`、
  `.option`、`.insn`、重定位辅助、CFI directive、数据 directive 和
  符号可见性 directive
* GAS 宏块，例如 `.macro`、`.irp`、`.irpc`、`.rept`、`.if`、`.else`
  以及对应的结束 directive
* GAS 注释形式，包括 `#`、`//`、块注释，以及在检测为 GAS 风格时支持
  ARM/GAS 的 `@` 行注释
* RISC-V 操作数形式，例如 `%hi(...)`、`%lo(...)`、`%pcrel_hi(...)`、
  `%pcrel_lo(...)`、`1b` 这种局部数字标签、CSR 操作数、向量寄存器操作数、
  压缩指令助记符和 `.insn` 编码

未知的 directive 和未知的小写助记符会被保守保留。`asmfmt` 仍会整理周围
的空白和对齐，但不会尝试在安全格式化之外验证或规范化未知语法。

## Fork 状态

这个 fork 计划独立维护。

当前目标：

* 为现有用户保持原有格式化行为稳定，
* 持续增强 GAS 和 RISC-V 支持，
* 不依赖上游活跃度继续接受修复和新特性，
* 保持配置和 CLI 行为可预测。

这个项目仍继承自上游实现的设计和历史行为，但当前仓库中的文档和发布内容
描述的是 fork 后的项目，而不是原始上游仓库。

## 格式化行为

默认行为包括：

* 使用 tab 缩进，使用空格对齐，
* 对齐操作数，
* 对齐行尾注释，
* 删除行尾空白，
* 清理多余空行，
* 将 label 规范到独立行，
* 清理注释空格，
* 在安全情况下把单行块注释转换为行注释，
* 在检测到的风格允许时清理并拆分分号语句。

如果你的代码库风格混杂，但又需要稳定一致的结果，建议显式设置
`source_style`，而不是依赖自动检测。

## 非目标

`asmfmt` 是 formatter，不是 assembler，也不是 linter。

它不打算：

* 校验 opcode 合法性，
* 拒绝未知 directive 或厂商自定义助记符，
* 在运行时依赖 binutils 或其他 assembler，
* 默认保证语义等价，
* 为了贴合 GAS 约定而重写无关的 Plan 9 既有格式行为。

## 作为库使用

`asmfmt` 也可以作为 Go package 使用：

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/klauspost/asmfmt"
)

func main() {
	src := bytes.NewBufferString("TEXT foo(SB),$0\nMOVQ AX,BX\n")
	out, err := asmfmt.Format(src)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(out))
}
```

如果需要可配置格式化，使用 `asmfmt.FormatWithOptions(...)`。

注意：当前库的 import path 仍然是上面示例中的上游 module path。以后如果
这个 fork 切换到自己的 module path，这里的 import path 也需要一起调整。

## 开发

常用命令：

```bash
go test ./...
go vet ./...
gofmt -w .
```

如果是有意的格式化输出变更，可以这样刷新 golden 文件：

```bash
go test -run TestRewrite -update
```

更聚焦的回归命令和贡献流程细节见 [AGENTS.md](AGENTS.md)。

## 编辑器集成

### GoLand

可以配置一个自定义 File Watcher：

* `Settings -> Tools -> File Watchers`
* 创建一个名为 `asmfmt` 的 `<custom>` watcher
* `File Type`：`x86 Plan 9 Assembly file`
* `Scope`：`Project Files`
* `Arguments`：`$FilePath$`
* `Output Paths to Refresh`：`$FilePath$`
* `Working Directory`：`$ProjectFileDir$`

启用：

* `Trigger the watcher regardless of syntax errors`
* `Create output file from stdout`

其余高级选项关闭即可。

![Goland Configuration](https://user-images.githubusercontent.com/5663952/114158973-96eebc80-9925-11eb-9aea-703ce474a7bb.png)

### Emacs

如果希望在保存时自动格式化汇编文件：

```elisp
(defun asm-mode-setup ()
  (set (make-local-variable 'gofmt-command) "asmfmt")
  (add-hook 'before-save-hook 'gofmt nil t)
)

(add-hook 'asm-mode-hook 'asm-mode-setup)
```

## Source style 差异

`asmfmt` 可以自动检测源码风格，也可以通过 `.asmfmt.toml` 里的
`source_style` 强制指定。

当前几种风格模式的主要差异在于注释处理、分号拆分和语法检测：

| 风格 | 典型输入 | 注释处理 | 分号拆分 | 检测线索 | 说明 |
| --- | --- | --- | --- | --- | --- |
| `plan9` | Go / Plan 9 汇编 | `//` 会被当作行注释。`#` 不会被当作注释起始符。`@` 行注释关闭。 | 关闭。`；` 不会被当作普通语句分隔符处理。 | `(SB)`、`(FP)`、`(PC)`、`(SP)` 以及大写指令形式通常会识别为 Plan 9 风格。 | 适合传统 Go 汇编源码。 |
| `gas` | 通用 GAS 风格汇编 | `//` 和 `#` 会被当作行注释。在 GAS 风格输入中，`@` 处于注释位置时也会被识别为行注释。 | 当输入看起来像 GAS 指令或 directive 语法时启用。 | `.section` 这类点号 directive 和小写指令形式通常会识别为 GAS 风格。 | 适合不依赖 RISC-V 特定检测的 ELF/GAS 汇编。 |
| `riscv-gas` | 使用 GAS 语法的 RISC-V 汇编 | 注释行为与 `gas` 相同。 | 分号处理与 `gas` 大体相同。 | `%hi(...)`、`%lo(...)`、`%pcrel_hi(...)`、`%pcrel_lo(...)`、`.option`、`.attribute`、`.insn`、`R_RISCV_...` 以及常见 RISC-V 寄存器名会把检测提升为 `riscv-gas`。 | 这是一个偏向 RISC-V 的 GAS 子模式，检测和语法覆盖更强。 |
| `auto` | 混合或未知输入 | 使用检测出来的风格处理注释。 | 使用检测出来的风格决定是否拆分语句。 | 初始比较保守，随着文件中出现的语法特征逐步升级判断。 | 如果代码库风格一致、又想少配配置项，推荐使用。 |

如果代码库里混用了多种风格，但你又需要可预测的稳定结果，建议显式设置
`source_style`，而不是依赖自动检测。

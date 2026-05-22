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

使用 Go 工具链安装 CLI：

```bash
go install github.com/Mi-AIoT/asmfmt/cmd/asmfmt@latest
```

或者，你也可以直接从 [GitHub Releases](https://github.com/Mi-AIoT/asmfmt/releases) 页面下载针对不同系统和架构预编译好的二进制文件。如果要体验最新 master 分支的功能，可从 [`beta` 预发布标签](https://github.com/Mi-AIoT/asmfmt/releases/tag/beta) 下载获取。

你也可以使用自更新参数将工具自动升级到最新版本、beta 测试版本或特定的发布版本标签：

```bash
asmfmt -update latest
```

如果需要覆盖默认的仓库源或更新服务器（例如用于本地测试或镜像源），可以使用 `ASMFMT_UPGRADE_REPO` 和 `ASMFMT_UPDATE_URL` 环境变量。

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

## 命令行用法与配置

完整的命令行参考、各选项的详细说明、支持的配置项以及编辑器集成配置，请参阅 [用户手册](docs/user_manual_zh.md)。

### 基本命令

```bash
# 就地格式化文件
asmfmt -w path/to/file.s

# 使用指定配置文件进行格式化
asmfmt -config /path/to/asmfmt.toml ./...
```

### 配置文件查找顺序

如果不提供 `-config` 选项，`asmfmt` 会自动按以下顺序查找配置文件：

1. 从被格式化文件所在目录自底向上查找最近的 `.asmfmt.toml` 文件
2. `~/.asmfmt.toml`
3. `/etc/asmfmt.toml`

完整注释版配置参考见 [.asmfmt.toml.example](.asmfmt.toml.example)。

## 代码风格检查 (Linter)

`asmfmt` 内置了代码风格检查器，可对汇编文件进行 RISC-V 和通用 GAS 代码规范检查。

使用 `-lint` 选项运行风格检查：
```bash
asmfmt -lint path/to/file.s
```

您可以在 `.asmfmt.toml` 的 `[lint]` 部分中配置规则及其严重级别。有关规则列表，请参见 [代码风格检查规则参考手册](docs/lint_rules_zh.md)；有关配置详细说明，请参阅 [用户手册](docs/user_manual_zh.md)。

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

# asmfmt 用户手册

`asmfmt` 是一个用于格式化 Go / Plan 9 汇编、GNU 汇编 (GAS) 语法以及 RISC-V 汇编代码的工具。它可以规范缩进、对齐操作数及注释、清理多余空行，并支持通过 TOML 配置文件定义特定项目的格式化风格。

---

## 1. 安装方式

### 通过 Go 工具链编译安装
使用 Go 工具链在本地构建并安装 CLI 工具：

```bash
go install github.com/Mi-AIoT/asmfmt/cmd/asmfmt@latest
```

安装后，`asmfmt` 二进制文件会被放置于你的 `$GOPATH/bin`（通常是 `~/go/bin`）目录下。请确保该目录已加入到系统的 `PATH` 环境路径中。

### 预构建二进制文件下载
此外，您也可以直接在 GitHub Releases 页面下载针对不同系统和架构编译好的二进制文件：
* **稳定版本 (Stable Releases)**: 您可以在 [Releases](https://github.com/Mi-AIoT/asmfmt/releases) 页面下载正式发布的稳定版本（如 `v2.0.0`）。
* **测试版本 (Beta / Nightly Builds)**: 每次 `master` 分支的提交在 CI 中构建成功后，都会自动生成最新的测试版本，并强制更新发布到 [`beta` 预发布标签](https://github.com/Mi-AIoT/asmfmt/releases/tag/beta) 下。您可以通过该标签获取最新的 master 分支功能。

---

## 2. 命令行选项说明

`asmfmt` 的基本命令格式为：
```bash
asmfmt [flags] [path ...]
```

默认情况下，运行 `asmfmt -h` 或 `asmfmt --help` 会在帮助信息的顶部打印版本信息（包含版本号、Git Commit Hash 及构建时间）。

### 选项清单与详细作用：

#### `-config <file>`
* **作用**：指定用于格式化校验的 TOML 配置文件路径。

#### `-init`
* **作用**：在当前目录下创建一份包含默认设置及详细选项注释的 `.asmfmt.toml` 默认配置文件。为了防止覆盖已有配置，如果当前目录下已存在该文件，将报错且不进行覆盖。
* **默认行为**：若不加该参数，`asmfmt` 会在被格式化文件所在目录自底向上查找最近的 `.asmfmt.toml` 文件；若没有找到，则查找 `~/.asmfmt.toml` 或 `/etc/asmfmt.toml`。

#### `-w`
* **作用**：将格式化后的结果直接写入覆盖源文件（就地修改）。
* **限制**：当输入为标准输入流（stdin）时，无法使用此选项。

#### `-d`
* **作用**：在终端中仅输出格式化前后的 diff 补丁，不修改源文件本身。
* **退出状态码**：如果检测到任何格式差异，将以状态码 `1` 退出。

#### `-l`
* **作用**：仅列出格式与 `asmfmt` 规定不一致的文件名，而不输出任何格式化内容。
* **退出状态码**：如果检测到任何格式差异，将以状态码 `1` 退出。
* **使用场景**：在 CI 静态校验检查中非常有用，可用于拦截不符合格式规范的提交。
* **示例**：
  ```bash
  # 格式化检查命令
  asmfmt -l file1.s file2.s
  
  # 若 file2.s 的格式不合规，命令将仅在标准输出中输出：
  file2.s
  ```

#### `-e`
* **作用**：解除报错条数限制，打印源文件中遇到的所有解析错误。
* **默认行为**：在不加 `-e` 时，如果一个文件出现很多处语法错误，工具默认只打印前 10 个在不同行上的报错，然后输出 `(too many errors)` 强行终止程序。使用 `-e` 则可以打印所有错误。
* **特别注意**：`-e` 不是运行模式开关。当源文件无错误并成功完成格式化时，格式化后的内容仍会被输出到标准输出（终端）；仅当有语法解析错误导致格式化失败时，该选项才会列出所有的解析报错。
* **示例**：
  ```bash
  # 假设 bad.s 有 15 处错误，默认只会输出 10 个：
  asmfmt bad.s
  
  # 使用 -e 输出所有 15 处错误：
  asmfmt -e bad.s
  ```

#### `-cpuprofile <file>`
* **作用**：将 CPU 性能剖析数据输出到指定文件。这是一个开发/调试级别的高级选项，主要用于工具本身的性能调优和在大规模文件格式化场景下分析性能瓶颈。

#### `-version`
* **作用**：打印版本信息（包括版本号、Git Commit Hash 及构建时间）并退出。

#### `-update <target>`
* **作用**：就地在线更新/升级 `asmfmt` 自身二进制文件。可选的值为：`latest`（最新发布版）、`beta`（测试版）或指定的 Release Tag（如 `v2.0.0`）。
* **支持的环境变量**：
  * `ASMFMT_UPGRADE_REPO`：覆盖默认的仓库源（默认为 `Mi-AIoT/asmfmt`）。
  * `ASMFMT_UPDATE_URL`：覆盖默认的 GitHub API 基础 URL（默认为 `https://api.github.com`），常用于集成测试或指向自定义镜像源。

---

## 3. 配置文件参考 (`.asmfmt.toml`)

你可以使用 TOML 格式自定义配置项目的排版规则。支持的配置键如下：

* **`indent_style`**：`"tab"` 或 `"space"`。设置使用制表符缩进还是空格缩进。默认值为 `"tab"`。
* **`indent_width`**：正整数。当 `indent_style = "space"` 时，指定每级缩进的空格数。默认值为 `8`。
* **`align_operands`**：布尔值。在一组连续指令行之间对齐第一个操作数。默认值为 `true`。
* **`align_comments`**：布尔值。对齐行尾注释。默认值为 `true`。
  * *注*：在 `gas` 或 `riscv-gas` 风格下，行尾注释对齐将仅根据带有注释的行计算宽度。这能防止因为某一行是不带注释的超长指令（如宏定义），把其他较短行的注释无意义地推到极右侧。
* **`align_continuations`**：布尔值。对齐多行宏体中末尾的 `\` 换行符。默认值为 `true`。
* **`max_blank_lines`**：非负整数。保留连续空行的最大数量（`0` 表示清除所有空行）。默认值为 `1`。
* **`split_semicolon_statements`**：布尔值。在允许的汇编风格下，将分号分隔的多条语句拆分到单独的行中。默认值为 `true`。
* **`newline_before_comments`**：布尔值。在新注释块开头的独立注释行之前插入一个空行。默认值为 `true`。
* **`newline_before_labels`**：布尔值。在标签或其他 0 级指令段之前插入一个空行。默认值为 `true`。
* **`labels_always_on_own_line`**：布尔值。强制将带有后续指令的内联标签（如 `loop: addi a0, a0, 1`）拆分到独立行。默认值为 `true`。
* **`line_comment_space`**：布尔值。在格式化引入的注释符号后强制插入一个空格（例如将 `//comment` 规范为 `// comment`）。默认值为 `true`。
* **`convert_single_line_block_comment`**：布尔值。在安全的情况下将单行块注释转换为普通行注释（如 `/* comment */` 转换为 `// comment`）。默认值为 `true`。
* **`preferred_comment_style`**：`"preserve"` 或 `"slash"`。`preserve` 保持文件原有的注释符风格（如 `#`、`@`、`//`）；`slash` 会把格式化插入的注释统一规范为 `//`。默认值为 `"preserve"`。
* **`source_style`**：指定汇编代码源风格。可选值有：
  * `"auto"`：根据文件内容关键字自动识别。
  * `"plan9"`：强制指定为 Go / Plan 9 汇编格式（使用大写助记符，`;` 不换行）。
  * `"gas"`：强制指定为常规 GNU 汇编格式（允许小写，支持 `#` 和 `//` 注释，`;` 可换行）。
  * `"riscv-gas"`：强制指定为偏向 RISC-V 约定的 GAS 格式（具有更强的寄存器及指令助记符侦测能力）。
  * 默认值为 `"auto"`。
* **`indent_gas_directives`**：布尔值。是否将零缩进的 GAS 伪指令（如 `.global`、`.type`、`.word`）缩进到当前的指令/宏级别。默认值为 `false`。

---

### 代码风格检查配置 (`[lint]`)

代码风格检查选项可以在 `[lint]` 部分指定。每个规则名称都可以配置为 `"error"`、`"warning"` 或 `"ignore"`。

* **`label_naming_style`**：`"snake_case"`, `"camelCase"`, `"PascalCase"`, 或 `"any"`（跳过命名检查）。默认值为 `"snake_case"`。
* **`macro_naming_style`**：`"UPPER_SNAKE_CASE"`, `"snake_case"`, 或 `"any"`（跳过命名检查）。默认值为 `"UPPER_SNAKE_CASE"`。
* **`copyright_require_spdx`**：`true` 或 `false`。规则 `L318` (`copyright_and_license`) 是否需要 SPDX 许可证标识符。默认值为 `true`。
* **`copyright_format`**：用于强制执行特定版权格式的正则表达式字符串。若为空，则默认匹配 "copyright" 或 "©"。默认值为 `""`。

#### 行内注释控制 (Inline Linter Control)

您可以通过在汇编文件中添加行内注释，在特定区域临时关闭或重新开启特定规则或所有规则的检查：
* **持久关闭特定检查**：使用 `// asmfmt:disable <规则ID或名称>...` 或 `/* asmfmt:disable <规则ID或名称>... */`。支持通过空格分隔同时指定多个规则。
* **持久关闭所有检查**：使用 `// asmfmt:disable` 或 `/* asmfmt:disable */`。
* **持久重新开启检查**：使用 `// asmfmt:enable <规则ID或名称>...` 或 `// asmfmt:enable`（重新开启所有）。
* **仅关闭当前行检查**：在当前行末尾或行内使用 `// asmfmt:disable-line [规则ID或名称]...`。若不指定规则，则默认关闭当前行的所有检查。
* **仅关闭下一行检查**：在目标行的上一行使用 `// asmfmt:disable-next-line [规则ID或名称]...`。若不指定规则，则默认关闭下一行的所有检查。

*注意：控制注释不需要顶头放置，可以包含前置空格或缩进，也允许直接写在指令行的末尾作为后缀注释。*

示例：
```assembly
	// 带有缩进的注释，持久关闭多个规则
	// asmfmt:disable L101 L303
	addi x10, x11, 1   # 跳过 L101 检查
	lw a0, (a1)        # 跳过 L303 检查
	// asmfmt:enable L101 L303

	// 仅通过行尾注释关闭当前行检查
	addi x10, x11, 1   # asmfmt:disable-line L101

	// 仅关闭下一行检查
	// asmfmt:disable-next-line L101
	addi x10, x11, 1   # 此处跳过 L101 检查
	addi x10, x11, 1   # 此处正常报告 L101 警告
```

完整的规则列表和示例，请参阅 [代码风格检查规则参考手册](lint_rules_zh.md)。

## 4. 开发与调试

若需要贡献代码或更新 golden 文件，可以使用以下命令：

* **运行单元与回归测试**：
  ```bash
  go test ./...
  ```
* **运行静态代码检查**：
  ```bash
  go vet ./...
  ```
* **一键刷新测试 golden 预期文件**（格式化修改经过验证且符合预期时使用）：
  ```bash
  go test -run TestRewrite -update
  ```

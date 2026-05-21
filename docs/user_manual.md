# asmfmt User Manual

`asmfmt` is a tool for formatting Go / Plan 9 assembly, GNU Assembler (GAS) syntax, and RISC-V assembly code. It normalizes indentation, aligns operands and end-of-line comments, cleans up blank lines, and supports TOML configuration files to define project-specific formatting styles.

---

## 1. Installation

Build and install the CLI tool locally using the Go toolchain:

```bash
go install github.com/Mi-AIoT/asmfmt/cmd/asmfmt@latest
```

This installs the `asmfmt` binary into your `$GOPATH/bin` (typically `~/go/bin`) directory. Ensure this directory is added to your system `PATH` environment variable.

---

## 2. CLI Options

The basic command-line syntax is:
```bash
asmfmt [flags] [path ...]
```

When running `asmfmt -h` or `asmfmt --help`, the tool prints version information (version number, git commit hash, and build time) at the very top of the help message by default.

### Flags & Descriptions:

#### `-config <file>`
* **Description**: Specifies the path to a TOML configuration file containing formatting rules.

#### `-init`
* **Description**: Creates a default `.asmfmt.toml` configuration file in the current directory with default settings and comments explaining all options. It reports an error if `.asmfmt.toml` already exists in the current directory to avoid overwriting existing configurations.
* **Default**: If not specified, `asmfmt` walks upward from the directory of the file being formatted to find the nearest `.asmfmt.toml`. If none is found, it falls back to `~/.asmfmt.toml` or `/etc/asmfmt.toml`.

#### `-w`
* **Description**: Writes the formatting results directly back to the source files (in-place modification).
* **Restriction**: This option cannot be used when the input is standard input (stdin).

#### `-d`
* **Description**: Prints formatting changes as a unified diff to standard output without modifying the source files.

#### `-l`
* **Description**: Lists only the filenames of files whose formatting differs from `asmfmt` standards, without outputting any formatted content.
* **Use Case**: Commonly used in CI static check pipelines to block non-compliant code submissions.
* **Example**:
  ```bash
  asmfmt -l file1.s file2.s
  
  # If file2.s is not compliant, the output will be:
  file2.s
  ```

#### `-e`
* **Description**: Disables error reporting limits and prints all parser/syntax errors encountered.
* **Default**: By default, `asmfmt` stops after reporting the first 10 errors on different lines and prints `(too many errors)`. With `-e`, all errors are printed.
* **Note**: `-e` does not alter output behavior for successful runs. If the source file has no syntax errors, the formatted code will still be printed to standard output. It only takes effect on parsing failures by listing all errors.
* **Example**:
  ```bash
  # If bad.s has 15 errors, default limits to 10:
  asmfmt bad.s
  
  # Use -e to print all 15 errors:
  asmfmt -e bad.s
  ```

#### `-cpuprofile <file>`
* **Description**: Outputs CPU profiling data to the specified file. Mainly used for performance tuning when formatting large sets of files.

#### `-version`
* **Description**: Prints version information (version number, git commit hash, and build time) and exits.

---

## 3. Configuration Options (`.asmfmt.toml`)

You can customize code styling rules via a TOML file. Supported keys:

* **`indent_style`**: `"tab"` or `"space"`. Indentation character to use. Default is `"tab"`.
* **`indent_width`**: Positive integer. Number of spaces per indentation level when `indent_style = "space"`. Default is `8`.
* **`align_operands`**: Boolean. Align the first operand across consecutive instruction lines. Default is `true`.
* **`align_comments`**: Boolean. Align end-of-line comments across a block. Default is `true`.
  * *Note*: In `gas` and `riscv-gas` styles, comment alignment is only calculated from lines that actually contain comments. This prevents long comment-less lines (like macro definitions) from pushing comments on other lines far to the right.
* **`align_continuations`**: Boolean. Align trailing `\` continuation markers in multiline macro bodies. Default is `true`.
* **`max_blank_lines`**: Non-negative integer. Maximum number of consecutive blank lines to preserve. `0` removes all blank lines. Default is `1`.
* **`split_semicolon_statements`**: Boolean. Split semicolon-separated statements onto separate lines where the style permits. Default is `true`.
* **`newline_before_comments`**: Boolean. Insert a blank line before standalone comment lines starting a new comment block. Default is `true`.
* **`newline_before_labels`**: Boolean. Insert a blank line before labels or other level-0 entries. Default is `true`.
* **`labels_always_on_own_line`**: Boolean. Force labels onto their own line when they have trailing instructions (e.g. `loop: addi a0, a0, 1` becomes a multi-line format). Default is `true`.
* **`line_comment_space`**: Boolean. Ensure a space is inserted after comment markers (e.g. normalizing `//comment` to `// comment`). Default is `true`.
* **`convert_single_line_block_comment`**: Boolean. Safely convert single-line block comments to line comments (e.g., `/* comment */` to `// comment`). Default is `true`.
* **`preferred_comment_style`**: `"preserve"` or `"slash"`. `"preserve"` keeps original comment markers (`#`, `@`, `//`); `"slash"` normalizes formatted line comments to `//`. Default is `"preserve"`.
* **`source_style`**: Force a specific source format or detect automatically. Supported values:
  * `"auto"`: Auto-detect formatting style based on keywords and syntax heuristics.
  * `"plan9"`: Plan 9 / Go assembly (uppercase instructions, semicolon splitting disabled).
  * `"gas"`: GNU Assembler syntax (lowercase instructions, `#` and `//` comments, semicolon splitting enabled).
  * `"riscv-gas"`: RISC-V oriented GAS syntax (enhanced detection for RISC-V registers and instructions).
  * Default is `"auto"`.
* **`indent_gas_directives`**: Boolean. Indent zero-indent GAS directives (like `.global`, `.type`, `.word`, `.byte`, `.section`) to the instruction/macro level. Default is `false`.

---

## 4. Source Style Differences

`asmfmt` behaves differently depending on the source style (either forced by `source_style` or auto-detected):

| Style | Typical Input | Comment Handling | Semicolon Splitting | Detection Hints | Notes |
| --- | --- | --- | --- | --- | --- |
| `plan9` | Go / Plan 9 assembly | `//` is treated as a line comment. `#` is not treated as a comment marker. `@` line comments are disabled. | Disabled. `;` is not treated as a normal statement separator. | `(SB)`, `(FP)`, `(PC)`, `(SP)`, and uppercase instruction forms tend to identify Plan 9 style. | Best for traditional Go assembler sources. |
| `gas` | Generic GAS-style assembly | `//` and `#` are treated as line comments. `@` line comments are also supported in GAS-like input when used in comment position. | Enabled where the input looks like GAS instruction or directive syntax. | Dot-directives such as `.section` and lowercase instruction forms tend to identify GAS style. | Best for non-Go ELF/GAS assembly that does not need RISC-V-specific detection. |
| `riscv-gas` | RISC-V assembly using GAS syntax | Same comment behavior as `gas`. | Same general semicolon behavior as `gas`. | RISC-V-specific forms such as `%hi(...)`, `%lo(...)`, `%pcrel_hi(...)`, `%pcrel_lo(...)`, `.option`, `.attribute`, `.insn`, `R_RISCV_...`, and common RISC-V register names promote detection to `riscv-gas`. | This is a GAS sub-mode with stronger RISC-V-oriented detection and syntax coverage. |
| `auto` | Mixed or unknown input | Uses the detected style for comment parsing. | Uses the detected style for statement splitting. | Starts conservative, then upgrades based on syntax seen in the file. | Recommended when your codebase is consistent and you want minimal configuration. |

---

## 5. Development and Testing

If you are contributing code or verifying style changes:

* **Run all tests**:
  ```bash
  go test ./...
  ```
* **Run static checks**:
  ```bash
  go vet ./...
  ```
* **Update golden expectation fixtures** (after verifying changes are correct):
  ```bash
  go test -run TestRewrite -update
  ```

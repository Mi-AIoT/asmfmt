package asmfmt

import (
	"strings"
	"testing"
)

type lintTestCase struct {
	name      string
	style     string
	config    string
	code      string
	expectIDs []string // Rule IDs we expect to find violations for
}

func TestLinterRules(t *testing.T) {
	cases := []lintTestCase{
		{
			name:  "Inline disable specific rule ID",
			style: "riscv-gas",
			code: `// Copyright
// SPDX-License-Identifier: Apache
// asmfmt:disable L101
	addi x10, x11, 1
// asmfmt:enable L101
	addi x10, x11, 1
`,
			expectIDs: []string{"L101"},
		},
		{
			name:  "Inline disable specific rule name",
			style: "riscv-gas",
			code: `// Copyright
// SPDX-License-Identifier: Apache
// asmfmt:disable abi_registers
	addi x10, x11, 1
// asmfmt:enable abi_registers
	addi x10, x11, 1
`,
			expectIDs: []string{"L101"},
		},
		{
			name:  "Inline disable all rules",
			style: "riscv-gas",
			code: `// Copyright
// SPDX-License-Identifier: Apache
// asmfmt:disable
	addi x10, x11, 1
	lw a0, (a1)
// asmfmt:enable
	addi x10, x11, 1
`,
			expectIDs: []string{"L101"},
		},
		{
			name:  "Inline disable multiple rules at once",
			style: "riscv-gas",
			code: `// Copyright
// SPDX-License-Identifier: Apache
// asmfmt:disable L101 L303
	addi x10, x11, 1
	lw a0, (a1)
// asmfmt:enable L101 L303
	addi x10, x11, 1
`,
			expectIDs: []string{"L101"},
		},
		{
			name:  "Inline disable with leading spaces and trailing comments",
			style: "riscv-gas",
			code: `// Copyright
// SPDX-License-Identifier: Apache
	// asmfmt:disable L101
	addi x10, x11, 1
	// asmfmt:enable L101
	addi x10, x11, 1
`,
			expectIDs: []string{"L101"},
		},
		{
			name:  "Inline disable-line specific rule",
			style: "riscv-gas",
			code: `// Copyright
// SPDX-License-Identifier: Apache
	addi x10, x11, 1   // asmfmt:disable-line L101
	addi x10, x11, 1
`,
			expectIDs: []string{"L101"},
		},
		{
			name:  "Inline disable-next-line specific rule",
			style: "riscv-gas",
			code: `// Copyright
// SPDX-License-Identifier: Apache
	// asmfmt:disable-next-line L101
	addi x10, x11, 1
	addi x10, x11, 1
`,
			expectIDs: []string{"L101"},
		},
		{
			name:  "Inline disable-line all rules",
			style: "riscv-gas",
			code: `// Copyright
// SPDX-License-Identifier: Apache
	addi x10, x11, 1   // asmfmt:disable-line
	addi x10, x11, 1
`,
			expectIDs: []string{"L101"},
		},
		{
			name:  "Declarative macros do not start instruction sequences",
			style: "gas",
			config: `[lint]
declarative_macros = ["MY_SECTION_BEGIN", "MY_GLOBAL_FUNC", "MY_SECTION_END"]`,
			code: `// Copyright
// SPDX-License-Identifier: Apache
	MY_SECTION_BEGIN
	MY_GLOBAL_FUNC my_func
my_func:
	ret
	MY_SECTION_END
`,
			expectIDs: []string{},
		},
		{
			name:      "L101 ABI registers invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi x10, x11, 1\n",
			expectIDs: []string{"L101"},
		},
		{
			name:      "L101 ABI registers valid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi a0, a1, 1\n",
			expectIDs: []string{},
		},
		{
			name:      "L102 compressed instructions invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tc.mv a0, a1\n",
			expectIDs: []string{"L102"},
		},
		{
			name:      "L103 operation immediate invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tadd a0, a1, 4\n",
			expectIDs: []string{"L103"},
		},
		{
			name:      "L103 operation immediate valid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi a0, a1, 4\n",
			expectIDs: []string{},
		},
		{
			name:      "L104 relocation operator spacing invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tlui a0, %hi (sym)\n",
			expectIDs: []string{"L104"},
		},
		{
			name:      "L104 relocation operator spacing valid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tlui a0, %hi(sym)\n",
			expectIDs: []string{},
		},
		{
			name:      "L105 gp load relaxation invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tla gp, _gp\n",
			expectIDs: []string{"L105"},
		},
		{
			name:      "L105 gp load relaxation valid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.option push\n\t.option norelax\n\tla gp, _gp\n\t.option pop\n",
			expectIDs: []string{},
		},
		{
			name:      "L106 csr names invalid raw constant",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tcsrr a0, 0x300\n",
			expectIDs: []string{"L106"},
		},
		{
			name:      "L106 csr names invalid lowercase custom",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tcsrr a0, my_custom_csr\n",
			expectIDs: []string{"L106"},
		},
		{
			name:      "L106 csr names valid standard",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tcsrr a0, mstatus\n",
			expectIDs: []string{},
		},
		{
			name:      "L106 csr names valid custom uppercase",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tcsrr a0, MY_CUSTOM_CSR\n",
			expectIDs: []string{},
		},
		{
			name:      "L106 csr names valid custom prefix",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tcsrr a0, CSR_my_csr\n",
			expectIDs: []string{},
		},
		{
			name:      "L107 jump instruction selection invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tj global_func\n",
			expectIDs: []string{"L107"},
		},
		{
			name:      "L107 jump instruction selection valid local",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tj .L_local_lbl\n",
			expectIDs: []string{},
		},
		{
			name:      "L108 pcrel relocation label invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi a0, a0, %pcrel_lo(global_sym)\n",
			expectIDs: []string{"L108"},
		},
		{
			name:      "L108 pcrel relocation label valid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi a0, a0, %pcrel_lo(.L_local_lbl)\n",
			expectIDs: []string{},
		},
		{
			name:      "L201 alignment directives invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.align 4\n",
			expectIDs: []string{"L201"},
		},
		{
			name:      "L202 extern directive invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.extern sym\n",
			expectIDs: []string{"L202"},
		},
		{
			name:      "L203 inline binary directives invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.word 0x12345678\n",
			expectIDs: []string{"L203"},
		},
		{
			name:      "L204 avoid globl invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.globl sym\n",
			expectIDs: []string{"L204"},
		},
		{
			name:      "L205 leb128 constant expression invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.uleb128 my_symbol\n",
			expectIDs: []string{"L205"},
		},
		{
			name:      "L205 leb128 constant expression valid diff",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.uleb128 B - A\n",
			expectIDs: []string{},
		},
		{
			name:      "L206 avoid space skip directives invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.space 16\n",
			expectIDs: []string{"L206"},
		},
		{
			name:      "L207 operator precedence parentheses invalid",
			style:     "gas",
			config:    "[lint]\ninline_binary_directives = \"ignore\"",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.long A * B >> 2\n",
			expectIDs: []string{"L207"},
		},
		{
			name:      "L207 operator precedence parentheses valid",
			style:     "gas",
			config:    "[lint]\ninline_binary_directives = \"ignore\"",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.long A * (B >> 2)\n",
			expectIDs: []string{},
		},
		{
			name:      "L208 end directive last invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.end\n\taddi a0, a0, 1\n",
			expectIDs: []string{"L208"},
		},
		{
			name:      "L301 local labels invalid prefix",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n.Lloop:\n\tret\n",
			expectIDs: []string{"L301"},
		},
		{
			name:      "L301 local labels valid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n.L_loop:\n\tret\n",
			expectIDs: []string{},
		},
		{
			name:      "L302 current point label invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tj .\n",
			expectIDs: []string{"L302"},
		},
		{
			name:      "L303 pointer offset shorthand invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tlw a0, (a1)\n",
			expectIDs: []string{"L303"},
		},
		{
			name:      "L304 option push pop invalid",
			style:     "riscv-gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.option push\n",
			expectIDs: []string{"L304"},
		},
		{
			name:      "L305 symbol preamble footer invalid type align",
			style:     "gas",
			config:    "[lint]\nfunction_doxygen_comment = \"ignore\"",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.type my_func, @function\nmy_func:\n\tret\n\t.size my_func, .-my_func\n",
			expectIDs: []string{"L305"},
		},
		{
			name:      "L305 symbol preamble footer invalid size",
			style:     "gas",
			config:    "[lint]\nfunction_doxygen_comment = \"ignore\"",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.p2align 2\n\t.type my_func, @function\nmy_func:\n\tret\n",
			expectIDs: []string{"L305"},
		},
		{
			name:      "L306 cfi start end balance invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.cfi_startproc\n\tret\n",
			expectIDs: []string{"L306"},
		},
		{
			name:      "L307 macro balance invalid",
			style:     "gas",
			config:    "[lint]\nmacro_naming_style = \"any\"",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.macro my_macro\n\tret\n",
			expectIDs: []string{"L307"},
		},
		{
			name:      "L308 instruction sequence termination invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nmy_func:\n\taddi a0, a0, 1\n",
			expectIDs: []string{"L308"},
		},
		{
			name:      "L308 instruction sequence termination valid with fallthrough comment",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nmy_func:\n\taddi a0, a0, 1\n\t// fall through\n",
			expectIDs: []string{},
		},
		{
			name:      "L309 function doxygen comment invalid",
			style:     "gas",
			config:    "[lint]\nsymbol_preamble_footer = \"ignore\"",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.global my_func\n\t.type my_func, @function\nmy_func:\n\tret\n\t.size my_func, .-my_func\n",
			expectIDs: []string{"L309"},
		},
		{
			name:      "L309 function doxygen comment valid",
			style:     "gas",
			config:    "[lint]\nsymbol_preamble_footer = \"ignore\"",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n/**\n * @brief my function\n */\n\t.global my_func\n\t.type my_func, @function\nmy_func:\n\tret\n\t.size my_func, .-my_func\n",
			expectIDs: []string{},
		},
		{
			name:      "L310 avoid hash and at comments invalid hash",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi a0, a0, 1 # increment\n",
			expectIDs: []string{"L310"},
		},
		{
			name:      "L310 avoid hash and at comments invalid at",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi a0, a0, 1 @ increment\n",
			expectIDs: []string{"L310"},
		},
		{
			name:      "L311 double label declaration invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nlabel_a:\nlabel_b:\n\tret\n",
			expectIDs: []string{"L311"},
		},
		{
			name:      "L312 reserved label names invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\neax:\n\tret\n",
			expectIDs: []string{"L312"},
		},
		{
			name:      "L313 unreachable code invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tret\n\taddi a0, a0, 1\n",
			expectIDs: []string{"L313"},
		},
		{
			name:      "L313 unreachable code valid with blank lines and comments",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\tret\n\n\t// comment line\nmy_label:\n\taddi a0, a0, 1\n",
			expectIDs: []string{},
		},
		{
			name:      "L314 single return statement invalid",
			style:     "gas",
			config:    "[lint]\nunreachable_code = \"ignore\"",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nmy_func:\n\tret\n\tret\n",
			expectIDs: []string{"L314"},
		},
		{
			name:      "L315 no fallthrough to function invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nfunc1:\n\taddi a0, a0, 1\nfunc2:\n\tret\n",
			expectIDs: []string{"L315"},
		},
		{
			name:      "L316 no jump out of function invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nfunc1:\n\tj func2\nfunc2:\n\tret\n",
			expectIDs: []string{"L316"},
		},
		{
			name:      "L317 no recursive calls invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nmy_func:\n\tcall my_func\n",
			expectIDs: []string{"L317"},
		},
		{
			name:      "L318 copyright and license invalid",
			style:     "gas",
			code:      "\taddi a0, a0, 1\n",
			expectIDs: []string{"L318"},
		},
		{
			name:      "L318 copyright only with require_spdx = false",
			style:     "gas",
			config:    "[lint]\ncopyright_require_spdx = false",
			code:      "// Copyright 2026 Nuclei\n\taddi a0, a0, 1\n",
			expectIDs: []string{},
		},
		{
			name:      "L318 copyright only with require_spdx = true",
			style:     "gas",
			config:    "[lint]\ncopyright_require_spdx = true",
			code:      "// Copyright 2026 Nuclei\n\taddi a0, a0, 1\n",
			expectIDs: []string{"L318"},
		},
		{
			name:      "L318 copyright pattern invalid",
			style:     "gas",
			config:    "[lint]\ncopyright_format = \"^// Copyright \\\\(c\\\\) [0-9]{4} Nuclei\"\ncopyright_require_spdx = false",
			code:      "// Copyright 2026 Nuclei\n\taddi a0, a0, 1\n",
			expectIDs: []string{"L318"},
		},
		{
			name:      "L318 copyright pattern valid",
			style:     "gas",
			config:    "[lint]\ncopyright_format = \"^// Copyright \\\\(c\\\\) [0-9]{4} Nuclei\"\ncopyright_require_spdx = false",
			code:      "// Copyright (c) 2026 Nuclei\n\taddi a0, a0, 1\n",
			expectIDs: []string{},
		},
		{
			name:      "L319 developer name length invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nthis_is_an_extremely_long_label_name_that_exceeds_thirty_one_characters:\n\tret\n",
			expectIDs: []string{"L319"},
		},
		{
			name:      "L320 line length limit invalid",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n// This line is extremely long and will definitely exceed the standard one hundred and twenty character limit specified in the rules...\n",
			expectIDs: []string{"L320"},
		},
		{
			name:      "L321 label naming style invalid camelCase",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\nmyLabel:\n\tret\n",
			expectIDs: []string{"L321"},
		},
		{
			name:      "L322 macro naming style invalid lowercase",
			style:     "gas",
			code:      "// Copyright\n// SPDX-License-Identifier: Apache\n\t.macro my_macro\n\t.endm\n",
			expectIDs: []string{"L322"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := DefaultOptions()
			opts.SourceStyle = tc.style

			// Load config if provided
			if tc.config != "" {
				var err error
				opts, err = ParseOptionsTOML([]byte(tc.config))
				if err != nil {
					t.Fatalf("failed to parse test config: %v", err)
				}
			}

			problems, err := Lint("test.s", strings.NewReader(tc.code), opts)
			if err != nil {
				t.Fatalf("unexpected lint error: %v", err)
			}

			// Print problems for debugging if mismatch
			t.Logf("Case %s problems: %v", tc.name, problems)

			// Verify we find all expected rule IDs
			foundMap := make(map[string]bool)
			for _, p := range problems {
				foundMap[p.RuleID] = true
			}

			for _, expID := range tc.expectIDs {
				if !foundMap[expID] {
					t.Errorf("expected violation of rule %s, but got none. Detected problems: %v", expID, problems)
				}
			}

			// Verify we do not find unexpected rule IDs (excluding L318, L320, etc. if not expected)
			for id := range foundMap {
				expected := false
				for _, expID := range tc.expectIDs {
					if id == expID {
						expected = true
						break
					}
				}
				// Allow copyright and line limit failures in tests where we intentionally omitted headers/lengths,
				// unless they were expected
				if !expected {
					if id == "L318" || id == "L320" || id == "L308" {
						continue
					}
					t.Errorf("unexpected violation of rule %s: %v", id, problems)
				}
			}
		})
	}
}

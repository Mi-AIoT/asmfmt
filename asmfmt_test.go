package asmfmt

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update .golden files")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func runTest(t *testing.T, in, out string) {
	f, err := os.Open(in)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	got, err := Format(f)
	if err != nil {
		t.Error(in, "-", err)
		return
	}

	expected, err := ioutil.ReadFile(out)
	if err != nil && !*update {
		t.Error(out, "-", err)
		return
	}

	// Convert expected file to LF in case someone did it for us.
	expected = []byte(strings.Replace(string(expected), "\r\n", "\n", -1))

	if !bytes.Equal(got, expected) {
		if *update {
			if in != out {
				if err := ioutil.WriteFile(out, got, 0666); err != nil {
					t.Error(err)
				}
				return
			}
			// in == out: don't accidentally destroy input
			t.Errorf("WARNING: -update did not rewrite input file %s", in)
		}

		t.Errorf("(gofmt %s) != %s (see %s.asmfmt)", in, out, in)
		d, err := diff(expected, got)
		if err == nil {
			t.Errorf("%s", d)
		}
		if err := ioutil.WriteFile(in+".asmfmt", got, 0666); err != nil {
			t.Error(err)
		}
	}
}

// TestRewrite processes testdata/*.input files and compares them to the
// corresponding testdata/*.golden files. The gofmt flags used to process
// a file must be provided via a comment of the form
//
//	//gofmt flags
//
// in the processed file within the first 20 lines, if any.
func TestRewrite(t *testing.T) {
	// determine input files
	match, err := filepath.Glob("testdata/*.in")
	if err != nil {
		t.Fatal(err)
	}

	for _, in := range match {
		out := in // for files where input and output are identical
		if strings.HasSuffix(in, ".in") {
			out = in[:len(in)-len(".in")] + ".golden"
		}
		runTest(t, in, out)
		if in != out {
			// Check idempotence.
			runTest(t, out, out)
		}
	}
}

func diff(b1, b2 []byte) (data []byte, err error) {
	f1, err := ioutil.TempFile("", "asmfmt")
	if err != nil {
		return
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "asmfmt")
	if err != nil {
		return
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.Write(b1)
	f2.Write(b2)

	data, err = exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return

}

// Go files must fail.
func TestGoFile(t *testing.T) {
	input := `package main

	func main() {
	}
	`
	_, err := Format(bytes.NewBuffer([]byte(input)))
	if err == nil {
		t.Error("go file not detected")
		return
	}
}

// Files containg zero byte values must fail.
func TestZeroByteFile(t *testing.T) {
	var input = []byte{13, 13, 10, 0, 0, 0, 13}
	_, err := Format(bytes.NewBuffer(input))
	if err == nil {
		t.Fatal("file containing zero (0) byte values not rejected")
		return
	}
}

func TestFindLineComment(t *testing.T) {
	tests := []struct {
		line string
		pos  int
		mark string
	}{
		{`addi a0, a0, 1 # increment`, 15, "#"},
		{`addi a0, a0, 1#not-comment`, -1, ""},
		{`add r0, r1, #1`, -1, ""},
		{`.string "not # a comment" # comment`, 26, "#"},
		{`.byte '#' // slash comment`, 10, "//"},
		{`addi a0, a0, 1 /* # not comment */ # comment`, 35, "#"},
	}
	for _, tt := range tests {
		pos, mark := findLineComment(tt.line, styleRiscvGas)
		if pos != tt.pos || mark != tt.mark {
			t.Fatalf("findLineComment(%q) = %d, %q; want %d, %q", tt.line, pos, mark, tt.pos, tt.mark)
		}
	}
}

func TestGasParamSplitting(t *testing.T) {
	tests := []struct {
		line   string
		params []string
	}{
		{`.section .foo,"ax",@progbits`, []string{".foo", `"ax"`, "@progbits"}},
		{`ld a0, %pcrel_lo(1b)(a0)`, []string{"a0", "%pcrel_lo(1b)(a0)"}},
		{`.insn r 0x33, 0, 0, a0, a1, a2`, []string{"r 0x33", "0", "0", "a0", "a1", "a2"}},
		{`.macro load reg:req, values:vararg`, []string{"load reg:req", "values:vararg"}},
		{`.set delta, . - symbol`, []string{"delta", ". - symbol"}},
		{`.word symbol1 - symbol2 + (1 << 4)`, []string{"symbol1 - symbol2 + (1 << 4)"}},
		{`.reloc 0, R_RISCV_PCREL_LO12_I, nested(%pcrel_hi(sym + 4))`, []string{"0", "R_RISCV_PCREL_LO12_I", "nested(%pcrel_hi(sym + 4))"}},
		{`.byte '\n', '.', -1, 0x10`, []string{`'\n'`, `'.'`, `-1`, `0x10`}},
	}
	for _, tt := range tests {
		st := newStatement(tt.line, nil)
		if st == nil {
			t.Fatalf("newStatement(%q) = nil", tt.line)
		}
		if strings.Join(st.params, "|") != strings.Join(tt.params, "|") {
			t.Fatalf("newStatement(%q).params = %#v; want %#v", tt.line, st.params, tt.params)
		}
	}
}

func TestSplitStatements(t *testing.T) {
	got := splitStatements(`addi a0, a0, 1; ret # done`, styleRiscvGas)
	want := []string{`addi a0, a0, 1`, `ret # done`}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("splitStatements = %#v; want %#v", got, want)
	}
	if shouldSplitSemicolonStatements(`#define X a; b \`) {
		t.Fatal("continued macro should not use semicolon splitting")
	}
}

func TestGasBlockIndentation(t *testing.T) {
	input := `.macro wrap reg
.if \reg
addi a0, a0, 1
.else
ret
.endif
.endm
`
	want := `.macro wrap reg
	.if \reg
		addi a0, a0, 1
	.else
		ret
	.endif
.endm
`
	got, err := Format(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("Format block indentation:\n%s", got)
	}
}

func TestMacroVarargMerging(t *testing.T) {
	st := newStatement(`.macro forward head, tail:vararg, 1, 2`, nil)
	if st == nil {
		t.Fatal("newStatement returned nil")
	}
	want := []string{"forward head", "tail:vararg, 1, 2"}
	if strings.Join(st.params, "|") != strings.Join(want, "|") {
		t.Fatalf("macro params = %#v; want %#v", st.params, want)
	}
}

func TestMacroBodyPreservesSemicolons(t *testing.T) {
	input := `.macro wrap reg
addi \reg, \reg, 1; addi \reg, \reg, 2
.endm
`
	want := `.macro wrap reg
	addi \reg, \reg, 1; addi \reg, \reg, 2
.endm
`
	got, err := Format(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("Format macro body:\n%s", got)
	}
}

func TestRiscvConditionalBranchKeepsIndentedBlock(t *testing.T) {
	input := `func_label:
    li x5, 1
    beq x5, x0, 1f
    csrr x18, mattri0_base
    csrr x19, mattri0_mask
    fence
1:
    ret
`
	opts := DefaultOptions()
	opts.IndentStyle = "space"
	opts.IndentWidth = 4
	opts.SourceStyle = "riscv-gas"
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := `func_label:
    li   x5, 1
    beq  x5, x0, 1f
    csrr x18, mattri0_base
    csrr x19, mattri0_mask
    fence
1:
    ret
`
	if string(got) != want {
		t.Fatalf("Format branch indentation:\n%s", got)
	}
}

func TestRiscvStandaloneCommentsStayInsideIndentedBlock(t *testing.T) {
	input := `func_label:
    li x5, 0x100
    // save control state
    // VERSION
    LOAD_UW x20, (x5)
    bne x20, x15, error
    // INFO
    LOAD_UW x24, 4(x5)
`
	opts := DefaultOptions()
	opts.IndentStyle = "space"
	opts.IndentWidth = 4
	opts.SourceStyle = "riscv-gas"
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := `func_label:
    li x5, 0x100

    // save control state
    // VERSION
    LOAD_UW x20, (x5)
    bne x20, x15, error

    // INFO
    LOAD_UW x24, 4(x5)
`
	if string(got) != want {
		t.Fatalf("Format standalone comments:\n%s", got)
	}
}

func TestBannerBlockCommentIsPreserved(t *testing.T) {
	input := `/***************************************************************************************************
 * banner line
 **************************************************************************************************/
`
	got, err := Format(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != input {
		t.Fatalf("Format banner block comment:\n%s", got)
	}
}

func TestStandaloneBlockCommentIsCanonicalized(t *testing.T) {
	input := `    /*
    parameters:
      a0: first value

    returns:
      a0: result
    */
func_label:
    ret
`
	opts := DefaultOptions()
	opts.IndentStyle = "space"
	opts.IndentWidth = 4
	opts.SourceStyle = "riscv-gas"
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := `/*
 * parameters:
 * a0: first value
 *
 * returns:
 * a0: result
 */
func_label:
    ret
`
	if string(got) != want {
		t.Fatalf("Format standalone block comment:\n%s", got)
	}
}

func TestUnknownDirectivePreservesText(t *testing.T) {
	input := `.foo a,b,@c # keep spacing
`
	got, err := Format(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != input {
		t.Fatalf("Format unknown directive:\n%s", got)
	}
}

func TestIndentedStandaloneBlockCommentIsZeroIndentedInOnePass(t *testing.T) {
	input := `    /*
        * parameters:
     * a0: first value
     */
func_label:
    ret
`
	opts := DefaultOptions()
	opts.IndentStyle = "space"
	opts.IndentWidth = 4
	opts.SourceStyle = "riscv-gas"
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := `/*
 * parameters:
 * a0: first value
 */
func_label:
    ret
`
	if string(got) != want {
		t.Fatalf("first format = %q, want %q", got, want)
	}
	again, err := FormatWithOptions(strings.NewReader(string(got)), opts)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(got) {
		t.Fatalf("second format changed output:\n%s", again)
	}
}

func TestSourceStyleDetection(t *testing.T) {
	tests := []struct {
		line string
		want sourceStyle
	}{
		{`TEXT runtime·memmove(SB), NOSPLIT, $4-12`, stylePlan9},
		{`.type foo, @function`, styleGas},
		{`addi a0, a0, 1 # inc`, styleRiscvGas},
		{`mov r0, r1 @ comment`, styleGas},
	}
	for _, tt := range tests {
		if got := detectSourceStyle(tt.line); got != tt.want {
			t.Fatalf("detectSourceStyle(%q) = %v; want %v", tt.line, got, tt.want)
		}
	}
}

func TestAtLineCommentStyle(t *testing.T) {
	pos, mark := findLineComment(`mov r0, r1 @ comment`, styleGas)
	if pos < 0 || mark != "@" {
		t.Fatalf("gas @ comment not detected: %d %q", pos, mark)
	}
	pos, mark = findLineComment(`.type foo, @function`, styleGas)
	if pos != -1 || mark != "" {
		t.Fatalf(".type @function misdetected: %d %q", pos, mark)
	}
	pos, mark = findLineComment(`add r0, r1, #1`, styleGas)
	if pos != -1 || mark != "" {
		t.Fatalf("ARM immediate misdetected: %d %q", pos, mark)
	}
}

func TestPlan9SemicolonsStayInline(t *testing.T) {
	if shouldSplitSemicolonStatementsForStyle(`REP; MOVSL`, stylePlan9, true) {
		t.Fatal("plan9 semicolon unexpectedly split")
	}
}

func TestParseOptionsTOMLRejectsUnknownField(t *testing.T) {
	_, err := ParseOptionsTOML([]byte("unknown_field = true\n"))
	if err == nil || !strings.Contains(err.Error(), "unknown config field") {
		t.Fatalf("ParseOptionsTOML unknown field error = %v", err)
	}
}

func TestParseOptionsTOMLRejectsInvalidEnum(t *testing.T) {
	_, err := ParseOptionsTOML([]byte("source_style = \"weird\"\n"))
	if err == nil || !strings.Contains(err.Error(), "invalid source_style") {
		t.Fatalf("ParseOptionsTOML invalid enum error = %v", err)
	}
}

func TestFormatWithOptionsSpaceIndent(t *testing.T) {
	input := "TEXT foo(SB),$0\nMOVQ AX,BX\n"
	opts := DefaultOptions()
	opts.IndentStyle = "space"
	opts.IndentWidth = 2
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "TEXT foo(SB), $0\n  MOVQ AX, BX\n"
	if string(got) != want {
		t.Fatalf("FormatWithOptions space indent:\n%s", got)
	}
}

func TestFormatWithOptionsDisableAlignment(t *testing.T) {
	input := "addi a0, a1, 1 # one\nlonginstruction a0, a1, 2 # two\n"
	opts := DefaultOptions()
	opts.AlignOperands = false
	opts.AlignComments = false
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "addi a0, a1, 1 # one\nlonginstruction a0, a1, 2 # two\n"
	if string(got) != want {
		t.Fatalf("FormatWithOptions alignment:\n%s", got)
	}
}

func TestFormatWithOptionsDisableSemicolonSplit(t *testing.T) {
	input := "addi a0, a0, 1; ret # done\n"
	opts := DefaultOptions()
	opts.SplitSemicolonStatements = false
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "addi a0, a0, 1; ret # done\n"
	if string(got) != want {
		t.Fatalf("FormatWithOptions semicolon split:\n%s", got)
	}
}

func TestFormatWithOptionsDisableLineCommentSpace(t *testing.T) {
	input := "// comment\naddi a0, a0, 1 // note\n"
	opts := DefaultOptions()
	opts.LineCommentSpace = false
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "//comment\naddi a0, a0, 1 //note\n"
	if string(got) != want {
		t.Fatalf("FormatWithOptions line comment spacing:\n%s", got)
	}
}

func TestFormatWithOptionsPreserveSingleLineBlockComment(t *testing.T) {
	input := "/* comment */\n"
	opts := DefaultOptions()
	opts.ConvertSingleLineBlockComment = false
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "/* comment */\n"
	if string(got) != want {
		t.Fatalf("FormatWithOptions block comment:\n%s", got)
	}
}

func TestFormatWithOptionsKeepLabelInline(t *testing.T) {
	input := "loop: addi a0, a0, 1\n"
	opts := DefaultOptions()
	opts.LabelsAlwaysOnOwnLine = false
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "loop: addi a0, a0, 1\n"
	if string(got) != want {
		t.Fatalf("FormatWithOptions label handling:\n%s", got)
	}
}

func TestFormatWithOptionsForcePlan9Style(t *testing.T) {
	input := "rep; movsl\n"
	opts := DefaultOptions()
	opts.SourceStyle = "plan9"
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "rep; movsl\n"
	if string(got) != want {
		t.Fatalf("FormatWithOptions forced plan9:\n%s", got)
	}
}

func TestFormatWithOptionsForceGasStyle(t *testing.T) {
	input := "MOV R0 @ one\nLONGNAME R0, R1 @ two\n"
	opts := DefaultOptions()
	opts.SourceStyle = "gas"
	opts.PreferredCommentStyle = "slash"
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "// one") || !strings.Contains(string(got), "// two") {
		t.Fatalf("FormatWithOptions forced gas:\n%s", got)
	}
}

func TestMacroAndLabelIndentation(t *testing.T) {
	input := `.macro STL_GLOBAL_FUNC x
	.global \x
	.type \x, %function
.endm
`
	got, err := Format(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	want := `.macro STL_GLOBAL_FUNC x
	.global \x
	.type \x, %function
.endm
`
	if string(got) != want {
		t.Fatalf("Format macro directive indentation got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatWithOptionsIndentGASDirectives(t *testing.T) {
	input := "foo:\n.global bar\n.type bar, @function\n\taddi a0, a0, 1\n.macro mymacro\n.word 42\n.endm\n"
	opts := DefaultOptions()
	opts.IndentGASDirectives = true
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "foo:\n\t.global bar\n\t.type bar, @function\n\taddi a0, a0, 1\n.macro mymacro\n\t.word 42\n.endm\n"
	if string(got) != want {
		t.Fatalf("FormatWithOptionsIndentGASDirectives got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatWithOptionsIndentGASDirectivesExtern(t *testing.T) {
	input := ".section .text\n.global _start\n.type _start, %function\n.extern utils_uart_init\n.extern utils_print_str\n_start:\n\tcall utils_uart_init\n"
	opts := DefaultOptions()
	opts.IndentGASDirectives = true
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "\t.section .text\n\t.global _start\n\t.type _start, %function\n\t.extern utils_uart_init\n\t.extern utils_print_str\n_start:\n\tcall utils_uart_init\n"
	if string(got) != want {
		t.Fatalf("TestFormatWithOptionsIndentGASDirectivesExtern got:\n%s\nwant:\n%s", got, want)
	}
}

func TestPreProcessorIndentation(t *testing.T) {
	input := "foo:\n#if __riscv_xlen == 32\n\tsw x14, (a3)\n#else\n\tfmv.d.x f0, x14\n#endif\n\tret\n"
	opts := DefaultOptions()
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "foo:\n#if __riscv_xlen == 32\n\tsw x14, (a3)\n\n#else\n\tfmv.d.x f0, x14\n\n#endif\n\tret\n"
	if string(got) != want {
		t.Fatalf("TestPreProcessorIndentation got:\n%s\nwant:\n%s", got, want)
	}
}

func TestGasTerminatorIndentationBypass(t *testing.T) {
	input := "stl_timer_handler:\n\taddi a0, a0, 1\n\tret\n\n\t// ===== software Interrupt Handler =====\n\tSTL_GLOBAL_FUNC stl_sft_handler\n\nstl_sft_handler:\n\tret\n"
	opts := DefaultOptions()
	opts.SourceStyle = "riscv-gas"
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "stl_timer_handler:\n\taddi a0, a0, 1\n\tret\n\n\t// ===== software Interrupt Handler =====\n\tSTL_GLOBAL_FUNC stl_sft_handler\n\nstl_sft_handler:\n\tret\n"
	if string(got) != want {
		t.Fatalf("TestGasTerminatorIndentationBypass got:\n%s\nwant:\n%s", got, want)
	}
}

func TestCommentedPreprocessorIndentation(t *testing.T) {
	input := "#define A 1\n// #define B 2\n#define C 3\n"
	opts := DefaultOptions()
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "#define A 1\n// #define B 2\n#define C 3\n"
	if string(got) != want {
		t.Fatalf("TestCommentedPreprocessorIndentation got:\n%s\nwant:\n%s", got, want)
	}
}

func TestCommentAlignmentWithLongNoCommentLine(t *testing.T) {
	input := "foo:\n\tSTL_CHECK_CRC_XPR STL_CORE_EXU_ISA_P_SIG0, STL_CORE_EXU_ISA_P_SIG1\n\tSTL_ASSEMBLE_RESTORE_CONTEXT // restore integer register and set stl status\n"
	opts := DefaultOptions()
	opts.SourceStyle = "riscv-gas"
	got, err := FormatWithOptions(strings.NewReader(input), opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "foo:\n\tSTL_CHECK_CRC_XPR STL_CORE_EXU_ISA_P_SIG0, STL_CORE_EXU_ISA_P_SIG1\n\tSTL_ASSEMBLE_RESTORE_CONTEXT // restore integer register and set stl status\n"
	if string(got) != want {
		t.Fatalf("TestCommentAlignmentWithLongNoCommentLine got:\n%s\nwant:\n%s", got, want)
	}
}

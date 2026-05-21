package asmfmt

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type sourceStyle int

const (
	styleUnknown sourceStyle = iota
	stylePlan9
	styleGas
	styleRiscvGas
)

func (f *fstate) observeStyle(line string) {
	if !f.detectStyle {
		return
	}
	hint := detectSourceStyle(line)
	switch {
	case hint == styleUnknown:
		return
	case f.style == styleUnknown:
		f.style = hint
	case f.style == styleGas && hint == styleRiscvGas:
		f.style = hint
	}
}

func detectSourceStyle(line string) sourceStyle {
	s := strings.TrimSpace(line)
	if s == "" {
		return styleUnknown
	}
	if isPreProcessorLine(s) {
		return styleUnknown
	}
	if strings.Contains(s, "(SB)") || strings.Contains(s, "(FP)") || strings.Contains(s, "(PC)") || strings.Contains(s, "(SP)") {
		return stylePlan9
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return styleUnknown
	}
	head := fields[0]
	if strings.HasPrefix(head, ".") {
		if looksLikeRiscvGas(line) {
			return styleRiscvGas
		}
		return styleGas
	}
	r, _ := utf8.DecodeRuneInString(head)
	if unicode.IsUpper(r) {
		return stylePlan9
	}
	if unicode.IsLower(r) {
		if looksLikeRiscvGas(line) {
			return styleRiscvGas
		}
		return styleGas
	}
	return styleUnknown
}

func looksLikeRiscvGas(line string) bool {
	hints := []string{
		"%hi(", "%lo(", "%pcrel_hi(", "%pcrel_lo(",
		".option", ".attribute", ".insn", "R_RISCV_",
	}
	for _, hint := range hints {
		if strings.Contains(line, hint) {
			return true
		}
	}
	registerHints := []string{
		" a0", " a1", " a2", " a3", " a4", " a5", " a6", " a7",
		" s0", " s1", " s2", " s3", " s4", " s5", " s6", " s7", " s8", " s9", " s10", " s11",
		" t0", " t1", " t2", " t3", " t4", " t5", " t6",
		" ra", " sp", " gp", " tp", " zero",
	}
	padded := " " + line + " "
	for _, hint := range registerHints {
		if strings.Contains(padded, hint+" ") || strings.Contains(padded, hint+",") || strings.Contains(padded, hint+")") {
			return true
		}
	}
	return false
}

func hashStartsComment(s string, i int, style sourceStyle) bool {
	if style == stylePlan9 {
		return false
	}
	return isHashLineComment(s, i)
}

func atStartsComment(s string, i int, style sourceStyle) bool {
	if style != styleGas && style != styleRiscvGas {
		return false
	}
	if i == 0 || !unicode.IsSpace(rune(s[i-1])) {
		return false
	}
	return i+1 == len(s) || unicode.IsSpace(rune(s[i+1]))
}

func isStandaloneCommentLine(s string, style sourceStyle) (string, bool) {
	if strings.HasPrefix(s, "//") {
		return "//", true
	}
	if strings.HasPrefix(s, "#") && !isPreProcessorInstruction(strings.Fields(s)[0]) {
		if style != stylePlan9 {
			return "#", true
		}
	}
	return "", false
}

func shouldSplitSemicolonStatementsForStyle(s string, style sourceStyle, enabled bool) bool {
	if !enabled {
		return false
	}
	if style == stylePlan9 {
		return false
	}
	if strings.HasPrefix(s, "#") || !strings.Contains(s, ";") || strings.HasSuffix(strings.TrimSpace(s), `\`) {
		return false
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return false
	}
	inst := fields[0]
	if strings.HasPrefix(inst, ".") {
		return style == styleGas || style == styleRiscvGas || style == styleUnknown
	}
	r, _ := utf8.DecodeRuneInString(inst)
	if style == styleGas || style == styleRiscvGas {
		return unicode.IsLower(r)
	}
	return unicode.IsLower(r)
}

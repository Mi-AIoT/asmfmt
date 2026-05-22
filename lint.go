package asmfmt

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// RuleScope defines the target scope of a rule.
type RuleScope int

const (
	ScopeAll RuleScope = iota
	ScopeGas
	ScopeRiscv
)

// Problem represents a style rule violation.
type Problem struct {
	Filename string
	Line     int
	RuleID   string
	RuleName string
	Message  string
	Severity string
}

func (p Problem) String() string {
	return fmt.Sprintf("%s:%d: [%s][%s] %s (%s)", p.Filename, p.Line, p.RuleID, p.RuleName, p.Message, p.Severity)
}

// Rule defines the interface that all linter rules must satisfy.
type Rule interface {
	ID() string
	Name() string
	Scope() RuleScope
	Check(st statement, lineNum int, state *lintState) *Problem
}

type optionState struct {
	relaxationEnabled bool
}

type lintLine struct {
	LineNum int
	Raw     string
	Sts     []statement
}

type lintState struct {
	filename            string
	rawLines            []string
	currentRawLine      string
	optionStack         []optionState
	relaxationEnabled   bool
	pushLines           []int
	cfiStartLines       []int
	macroStartLines     []int
	lastNonCommentSt    *statement
	lastNonCommentLine  int
	lastInstructionLine int
	activeFunctions     map[string]int
	globalSymbols       map[string]bool

	inInstructionSequence              bool
	lastInstructionWasTerminator       bool
	lastLineCommentContainsFallthrough bool

	lastWasLabel          bool
	lastWasTerminatorInst bool

	currentFunc     string
	funcReturnCount int

	hasCopyright bool
	hasSPDX      bool

	custom           map[string]interface{}
	disabledRules    map[string]bool
	allRulesDisabled bool
}

func isCommentOrBlank(s string, style sourceStyle) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	if _, ok := isStandaloneCommentLine(s, style); ok {
		return true
	}
	if strings.HasPrefix(s, "/*") || strings.HasPrefix(s, "*") || strings.HasSuffix(s, "*/") {
		return true
	}
	return false
}

func newLintState(filename string, rawLines []string) *lintState {
	return &lintState{
		filename:                     filename,
		rawLines:                     rawLines,
		relaxationEnabled:            true,
		activeFunctions:              make(map[string]int),
		globalSymbols:                make(map[string]bool),
		custom:                       make(map[string]interface{}),
		lastInstructionWasTerminator: true,
		disabledRules:                make(map[string]bool),
	}
}

func (state *lintState) isRuleDisabled(ruleID, ruleName string) bool {
	if state.allRulesDisabled {
		return true
	}
	return state.disabledRules[strings.ToUpper(ruleID)] || state.disabledRules[strings.ToLower(ruleName)]
}

// Helper to determine if a string is a valid register name
func isRegister(s string) bool {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ",")
	s = strings.TrimSpace(s)
	switch s {
	case "zero", "ra", "sp", "gp", "tp", "fp":
		return true
	}
	if len(s) < 2 {
		return false
	}
	switch s[0] {
	case 'x', 'f':
		val := s[1:]
		if val == "0" {
			return true
		}
		if len(val) > 0 && val[0] >= '1' && val[0] <= '9' {
			var n int
			_, err := fmt.Sscanf(val, "%d", &n)
			return err == nil && n >= 0 && n <= 31
		}
	case 'a', 't', 's':
		if s[0] == 'a' && len(s) == 2 && s[1] >= '0' && s[1] <= '7' {
			return true
		}
		if s[0] == 't' && len(s) == 2 && s[1] >= '0' && s[1] <= '6' {
			return true
		}
		if s[0] == 's' && len(s) >= 2 {
			val := s[1:]
			if val == "0" || val == "1" || val == "10" || val == "11" {
				return true
			}
			if len(val) == 1 && val[0] >= '2' && val[0] <= '9' {
				return true
			}
		}
	}
	if strings.HasPrefix(s, "ft") || strings.HasPrefix(s, "fs") || strings.HasPrefix(s, "fa") {
		val := s[2:]
		if len(val) > 0 {
			var n int
			_, err := fmt.Sscanf(val, "%d", &n)
			if err == nil {
				if strings.HasPrefix(s, "fa") && n >= 0 && n <= 7 {
					return true
				}
				if (strings.HasPrefix(s, "ft") || strings.HasPrefix(s, "fs")) && n >= 0 && n <= 11 {
					return true
				}
			}
		}
	}
	return false
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return true
	}
	if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		return true
	}
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isAllUppercase(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func matchesNamingStyle(name string, style string) bool {
	if style == "any" || style == "" {
		return true
	}
	if strings.HasPrefix(name, ".") {
		name = name[1:]
		if strings.HasPrefix(name, "L_") {
			name = name[2:]
		} else if strings.HasPrefix(name, "L") {
			name = name[1:]
		}
	}
	if name == "" {
		return true
	}
	switch style {
	case "snake_case":
		for i, r := range name {
			if i == 0 {
				if !unicode.IsLower(r) && r != '_' {
					return false
				}
			} else {
				if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
					return false
				}
			}
		}
		return true
	case "camelCase":
		for i, r := range name {
			if i == 0 {
				if !unicode.IsLower(r) {
					return false
				}
			} else {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					return false
				}
			}
		}
		return true
	case "PascalCase":
		for i, r := range name {
			if i == 0 {
				if !unicode.IsUpper(r) {
					return false
				}
			} else {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					return false
				}
			}
		}
		return true
	}
	return true
}

func matchesMacroNamingStyle(name string, style string) bool {
	if style == "any" || style == "" {
		return true
	}
	switch style {
	case "UPPER_SNAKE_CASE":
		for i, r := range name {
			if i == 0 {
				if !unicode.IsUpper(r) && r != '_' {
					return false
				}
			} else {
				if !unicode.IsUpper(r) && !unicode.IsDigit(r) && r != '_' {
					return false
				}
			}
		}
		return true
	case "snake_case":
		for i, r := range name {
			if i == 0 {
				if !unicode.IsLower(r) && r != '_' {
					return false
				}
			} else {
				if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
					return false
				}
			}
		}
		return true
	}
	return true
}

func isReservedName(name string) bool {
	name = strings.ToLower(name)
	if isRegister(name) {
		return true
	}
	switch name {
	case "eax", "ebx", "ecx", "edx", "esi", "edi", "ebp", "esp",
		"rax", "rbx", "rcx", "rdx", "rsi", "rdi", "rbp", "rsp",
		"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15", "lr", "pc":
		return true
	}
	if isKnownGasDirectiveName("." + name) {
		return true
	}
	commonInsts := map[string]bool{
		"add": true, "addi": true, "sub": true, "mul": true, "div": true, "and": true, "andi": true, "or": true, "ori": true, "xor": true, "xori": true,
		"sll": true, "slli": true, "srl": true, "srli": true, "sra": true, "srai": true, "lui": true, "auipc": true, "jal": true, "jalr": true,
		"beq": true, "bne": true, "blt": true, "bge": true, "bltu": true, "bgeu": true, "lb": true, "lh": true, "lw": true, "lbu": true, "lhu": true,
		"sb": true, "sh": true, "sw": true, "fence": true, "ecall": true, "ebreak": true, "csrrw": true, "csrrs": true, "csrrc": true,
		"csrrwi": true, "csrrsi": true, "csrrci": true, "ret": true, "call": true, "tail": true, "mv": true, "li": true, "neg": true, "not": true,
		"seqz": true, "snez": true, "sltz": true, "sgtz": true, "beqz": true, "bnez": true, "j": true, "jr": true, "la": true, "lla": true,
		"mov": true, "jmp": true, "push": true, "pop": true, "cmp": true, "nop": true,
	}
	return commonInsts[name]
}

func isTerminatorInstruction(inst string) bool {
	inst = strings.ToLower(inst)
	switch inst {
	case "ret", "j", "jmp", "tail", "mret", "sret", "uret", "unimp", "jr":
		return true
	}
	return false
}

// Registry and Rules definition
var (
	hwRegRegex             = regexp.MustCompile(`\b(x0|x[1-9]|x[1-2][0-9]|x3[0-1]|f0|f[1-9]|f[1-2][0-9]|f3[0-1])\b`)
	relocationSpacingRegex = regexp.MustCompile(`(?i)%(hi|lo|pcrel_hi|pcrel_lo|got_pcrel_hi|tprel_hi|tprel_lo|tprel_add|tls_ie_pcrel_hi|tls_gd_pcrel_hi)\s+\(`)
	pcrelLoRegex           = regexp.MustCompile(`%pcrel_lo\(([^)]+)\)`)
	identifierRegex        = regexp.MustCompile(`^[a-zA-Z_.][a-zA-Z0-9_.]*$`)
	offsetShorthandRegex   = regexp.MustCompile(`^\([a-zA-Z0-9]+\)$`)
)

func hasPrecedenceViolation(expr string) bool {
	for {
		start := strings.LastIndex(expr, "(")
		if start == -1 {
			break
		}
		end := strings.Index(expr[start:], ")")
		if end == -1 {
			break
		}
		end = start + end
		expr = expr[:start] + "x" + expr[end+1:]
	}
	hasMult := false
	hasAdd := false
	hasShift := false
	hasBit := false
	temp := expr
	if strings.Contains(temp, "<<") || strings.Contains(temp, ">>") {
		hasShift = true
		temp = strings.ReplaceAll(temp, "<<", " ")
		temp = strings.ReplaceAll(temp, ">>", " ")
	}
	if strings.Contains(temp, "*") || strings.Contains(temp, "/") {
		hasMult = true
	}
	for i, r := range temp {
		if r == '+' || r == '-' {
			isBinary := false
			for j := i - 1; j >= 0; j-- {
				if temp[j] != ' ' && temp[j] != '\t' {
					prevChar := temp[j]
					if prevChar != '+' && prevChar != '-' && prevChar != '*' && prevChar != '/' && prevChar != '&' && prevChar != '|' && prevChar != '^' {
						isBinary = true
					}
					break
				}
			}
			if isBinary {
				hasAdd = true
			}
		}
	}
	if strings.Contains(temp, "&") || strings.Contains(temp, "|") || strings.Contains(temp, "^") {
		hasBit = true
	}
	count := 0
	if hasMult {
		count++
	}
	if hasAdd {
		count++
	}
	if hasShift {
		count++
	}
	if hasBit {
		count++
	}
	return count > 1
}

func hasPrecedingDoxygen(lines []lintLine, labelIdx int) bool {
	for i := labelIdx - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i].Raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ".") {
			continue
		}
		if strings.HasPrefix(line, "///") || strings.HasPrefix(line, "//!") {
			return true
		}
		if strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") || strings.HasSuffix(line, "*/") {
			for j := i; j >= 0; j-- {
				l2 := strings.TrimSpace(lines[j].Raw)
				if strings.HasPrefix(l2, "/*") {
					if strings.HasPrefix(l2, "/**") || strings.HasPrefix(l2, "/*!") {
						return true
					}
					return false
				}
			}
			return false
		}
		return false
	}
	return false
}

func hasCurrentPoint(s string) bool {
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '.' {
			if i > 0 && (unicode.IsLetter(runes[i-1]) || unicode.IsDigit(runes[i-1]) || runes[i-1] == '_' || runes[i-1] == '.') {
				continue
			}
			if i+1 < len(runes) && (unicode.IsLetter(runes[i+1]) || runes[i+1] == '_' || runes[i+1] == '.') {
				continue
			}
			return true
		}
	}
	return false
}

func getRules() []Rule {
	return []Rule{
		&ruleL101{}, &ruleL102{}, &ruleL103{}, &ruleL104{}, &ruleL105{}, &ruleL106{}, &ruleL107{}, &ruleL108{},
		&ruleL201{}, &ruleL202{}, &ruleL203{}, &ruleL204{}, &ruleL205{}, &ruleL206{}, &ruleL207{}, &ruleL208{},
		&ruleL301{}, &ruleL302{}, &ruleL303{}, &ruleL304{}, &ruleL305{}, &ruleL306{}, &ruleL307{}, &ruleL308{},
		&ruleL309{}, &ruleL310{}, &ruleL311{}, &ruleL312{}, &ruleL313{}, &ruleL314{}, &ruleL315{}, &ruleL316{},
		&ruleL317{}, &ruleL318{}, &ruleL319{}, &ruleL320{}, &ruleL321{}, &ruleL322{},
	}
}

// Lint runs target-scoped rules line-by-line using a lintState and returns all violations.
func Lint(filename string, in io.Reader, opts Options) ([]Problem, error) {
	nopts, err := opts.normalize()
	if err != nil {
		return nil, err
	}
	var rawLines []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		rawLines = append(rawLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Detect style
	fileStyle := nopts.sourceStyle
	if fileStyle == styleUnknown {
		detected := styleUnknown
		for _, line := range rawLines {
			hint := detectSourceStyle(line)
			if hint != styleUnknown {
				if detected == styleUnknown {
					detected = hint
				} else if detected == styleGas && hint == styleRiscvGas {
					detected = hint
				}
			}
		}
		if detected == styleUnknown {
			detected = styleGas // fallback for linter
		}
		fileStyle = detected
	}

	state := newLintState(filename, rawLines)
	state.relaxationEnabled = true
	state.custom["label_naming_style"] = nopts.lint.labelNamingStyle
	state.custom["macro_naming_style"] = nopts.lint.macroNamingStyle
	state.custom["copyright_require_spdx"] = nopts.lint.copyrightRequireSpdx
	state.custom["copyright_format"] = nopts.lint.copyrightFormat
	if nopts.lint.copyrightFormat != "" {
		if re, err := regexp.Compile(nopts.lint.copyrightFormat); err == nil {
			state.custom["copyright_regex"] = re
		}
	}

	// Build global symbol and local function map
	// First pass: locate all labels and symbols
	var lines []lintLine
	defines := make(map[string]struct{})
	for idx, line := range rawLines {
		lineNum := idx + 1
		s := strings.TrimSpace(line)

		if isCommentOrBlank(s, fileStyle) {
			lines = append(lines, lintLine{
				LineNum: lineNum,
				Raw:     line,
				Sts:     nil,
			})
			continue
		}

		// Semicolon splitting
		var parts []string
		if shouldSplitSemicolonStatementsForStyle(s, fileStyle, nopts.splitSemicolonStatements) {
			parts = splitStatements(s, fileStyle)
		} else {
			parts = []string{s}
		}

		var sts []statement
		for _, part := range parts {
			if part == "" {
				continue
			}
			var st *statement
			if isConservativeUnknownGasDirective(part) {
				fields := strings.Fields(part)
				st = &statement{
					instruction: fields[0],
					function:    true,
				}
				if len(fields) > 1 {
					rest := strings.Join(fields[1:], " ")
					st.setParams(rest)
				}
			} else {
				st = newStatementWithStyle(part, defines, fileStyle)
			}
			if st != nil {
				if def := st.define(); def != "" {
					defines[def] = struct{}{}
				}
				sts = append(sts, *st)
			}
		}
		lines = append(lines, lintLine{
			LineNum: lineNum,
			Raw:     line,
			Sts:     sts,
		})
	}

	// Gather all global and function symbol names first
	for _, l := range lines {
		for _, st := range l.Sts {
			if st.instruction == ".global" || st.instruction == ".globl" {
				for _, p := range st.params {
					state.globalSymbols[p] = true
				}
			}
			if st.instruction == ".type" && len(st.params) >= 2 {
				if strings.Contains(st.params[1], "function") {
					state.globalSymbols[st.params[0]] = true
				}
			}
		}
	}

	var problems []Problem
	rules := getRules()

	// Second pass: lint each line and statement
	for _, l := range lines {
		state.currentRawLine = l.Raw
		lineNum := l.LineNum

		// Parse inline linter control comments
		if idx := strings.Index(l.Raw, "asmfmt:disable"); idx != -1 {
			rest := l.Raw[idx+len("asmfmt:disable"):]
			if endIdx := strings.Index(rest, "*/"); endIdx != -1 {
				rest = rest[:endIdx]
			}
			fields := strings.Fields(rest)
			if len(fields) == 0 {
				state.allRulesDisabled = true
			} else {
				for _, f := range fields {
					val := strings.TrimSpace(f)
					if strings.HasPrefix(strings.ToUpper(val), "L") {
						state.disabledRules[strings.ToUpper(val)] = true
					} else {
						state.disabledRules[strings.ToLower(val)] = true
					}
				}
			}
		}
		if idx := strings.Index(l.Raw, "asmfmt:enable"); idx != -1 {
			rest := l.Raw[idx+len("asmfmt:enable"):]
			if endIdx := strings.Index(rest, "*/"); endIdx != -1 {
				rest = rest[:endIdx]
			}
			fields := strings.Fields(rest)
			if len(fields) == 0 {
				state.allRulesDisabled = false
				state.disabledRules = make(map[string]bool)
			} else {
				for _, f := range fields {
					val := strings.TrimSpace(f)
					if strings.HasPrefix(strings.ToUpper(val), "L") {
						delete(state.disabledRules, strings.ToUpper(val))
					} else {
						delete(state.disabledRules, strings.ToLower(val))
					}
				}
			}
		}

		// Pre-scan raw line for copyright/SPDX
		lower := strings.ToLower(l.Raw)
		if re, ok := state.custom["copyright_regex"].(*regexp.Regexp); ok {
			if re.MatchString(l.Raw) {
				state.hasCopyright = true
			}
		} else {
			if strings.Contains(lower, "copyright") || strings.Contains(l.Raw, "©") {
				state.hasCopyright = true
			}
		}
		if strings.Contains(l.Raw, "SPDX-License-Identifier") {
			state.hasSPDX = true
		}

		// Comment updates
		if strings.Contains(lower, "fall through") || strings.Contains(lower, "fallthrough") {
			state.lastLineCommentContainsFallthrough = true
		}

		// Handle empty lines or pure comment lines
		if len(l.Sts) == 0 {
			st := statement{}
			for _, r := range rules {
				if r.Scope() == ScopeRiscv && fileStyle != styleRiscvGas {
					continue
				}
				if r.Scope() == ScopeGas && fileStyle != styleGas && fileStyle != styleRiscvGas {
					continue
				}
				severity := nopts.lint.severities[r.Name()]
				if severity == "ignore" || severity == "" {
					continue
				}
				if state.isRuleDisabled(r.ID(), r.Name()) {
					continue
				}
				prob := r.Check(st, lineNum, state)
				if prob != nil {
					prob.Severity = severity
					problems = append(problems, *prob)
				}
			}
			continue
		}

		for i := 0; i < len(l.Sts); i++ {
			st := &l.Sts[i]
			if st.isLabel() {
				st.function = false
			}
			if !strings.Contains(st.instruction, "(") && !st.isPreProcessor() && !st.isGasDirective() && !st.isLabel() {
				st.function = false
			}

			// Run all rule checks for this statement BEFORE updating state
			for _, r := range rules {
				if r.Scope() == ScopeRiscv && fileStyle != styleRiscvGas {
					continue
				}
				if r.Scope() == ScopeGas && fileStyle != styleGas && fileStyle != styleRiscvGas {
					continue
				}
				severity := nopts.lint.severities[r.Name()]
				if severity == "ignore" || severity == "" {
					continue
				}
				if state.isRuleDisabled(r.ID(), r.Name()) {
					continue
				}
				prob := r.Check(*st, lineNum, state)
				if prob != nil {
					prob.Severity = severity
					problems = append(problems, *prob)
				}
			}

			// Update state AFTER checks
			if st.instruction == ".option" && len(st.params) > 0 {
				switch st.params[0] {
				case "norelax":
					state.relaxationEnabled = false
				case "relax":
					state.relaxationEnabled = true
				case "push":
					state.optionStack = append(state.optionStack, optionState{relaxationEnabled: state.relaxationEnabled})
					state.pushLines = append(state.pushLines, lineNum)
				case "pop":
					if len(state.optionStack) > 0 {
						state.relaxationEnabled = state.optionStack[len(state.optionStack)-1].relaxationEnabled
						state.optionStack = state.optionStack[:len(state.optionStack)-1]
						state.pushLines = state.pushLines[:len(state.pushLines)-1]
					}
				}
			}

			if st.instruction == ".cfi_startproc" {
				state.cfiStartLines = append(state.cfiStartLines, lineNum)
			} else if st.instruction == ".cfi_endproc" {
				if len(state.cfiStartLines) > 0 {
					state.cfiStartLines = state.cfiStartLines[:len(state.cfiStartLines)-1]
				}
			}

			if st.instruction == ".macro" {
				state.macroStartLines = append(state.macroStartLines, lineNum)
			} else if st.instruction == ".endm" {
				if len(state.macroStartLines) > 0 {
					state.macroStartLines = state.macroStartLines[:len(state.macroStartLines)-1]
				}
			}

			if st.instruction == ".type" && len(st.params) >= 2 {
				if strings.Contains(st.params[1], "function") {
					state.activeFunctions[st.params[0]] = lineNum
				}
			}
			if st.instruction == ".size" && len(st.params) >= 1 {
				delete(state.activeFunctions, st.params[0])
			}

			if st.isLabel() {
				labelName := strings.TrimSuffix(st.instruction, ":")
				if !isNumeric(labelName) && !strings.HasPrefix(labelName, ".L") {
					state.currentFunc = labelName
					state.funcReturnCount = 0
					state.inInstructionSequence = false
					state.lastInstructionWasTerminator = true
				}
			}

			if st.instruction != "" && !st.function && !st.continued {
				state.lastNonCommentSt = st
				state.lastNonCommentLine = lineNum

				if !st.isGasDirective() && !st.isLabel() && !st.isPreProcessor() {
					state.inInstructionSequence = true
					state.lastInstructionWasTerminator = isTerminatorInstruction(st.instruction)
					state.lastLineCommentContainsFallthrough = false
					state.lastWasTerminatorInst = isTerminatorInstruction(st.instruction)
					state.lastInstructionLine = lineNum
				}
				if st.isLabel() {
					state.lastWasLabel = true
				} else {
					state.lastWasLabel = false
				}
			}
		}
	}

	// EOF checks - Call Check with a dummy eof statement
	eofSt := statement{instruction: ".end_of_file"}
	eofLine := len(rawLines)
	if eofLine == 0 {
		eofLine = 1
	}
	for _, r := range rules {
		if r.Scope() == ScopeRiscv && fileStyle != styleRiscvGas {
			continue
		}
		if r.Scope() == ScopeGas && fileStyle != styleGas && fileStyle != styleRiscvGas {
			continue
		}
		severity := nopts.lint.severities[r.Name()]
		if severity == "ignore" || severity == "" {
			continue
		}
		if state.isRuleDisabled(r.ID(), r.Name()) {
			continue
		}
		prob := r.Check(eofSt, eofLine, state)
		if prob != nil {
			prob.Severity = severity
			problems = append(problems, *prob)
		}
	}

	return problems, nil
}

// Rule L101: abi_registers
type ruleL101 struct{}

func (r *ruleL101) ID() string       { return "L101" }
func (r *ruleL101) Name() string     { return "abi_registers" }
func (r *ruleL101) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL101) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	for _, p := range st.params {
		if hwRegRegex.MatchString(p) {
			match := hwRegRegex.FindString(p)
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("register %q should be replaced with its ABI name", match),
			}
		}
	}
	return nil
}

// Rule L102: compressed_instructions
type ruleL102 struct{}

func (r *ruleL102) ID() string       { return "L102" }
func (r *ruleL102) Name() string     { return "compressed_instructions" }
func (r *ruleL102) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL102) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	if strings.HasPrefix(strings.ToLower(st.instruction), "c.") {
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  fmt.Sprintf("compressed instruction mnemonic %q is banned", st.instruction),
		}
	}
	return nil
}

// Rule L103: operation_immediate
type ruleL103 struct{}

func (r *ruleL103) ID() string       { return "L103" }
func (r *ruleL103) Name() string     { return "operation_immediate" }
func (r *ruleL103) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL103) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	switch strings.ToLower(st.instruction) {
	case "add", "and", "or", "xor", "sll", "srl", "sra", "addw", "sllw", "srlw", "sraw":
		if len(st.params) == 3 && !isRegister(st.params[2]) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("operation-with-immediate should use %si format instead of %s", st.instruction, st.instruction),
			}
		}
	}
	return nil
}

// Rule L104: relocation_operator_spacing
type ruleL104 struct{}

func (r *ruleL104) ID() string       { return "L104" }
func (r *ruleL104) Name() string     { return "relocation_operator_spacing" }
func (r *ruleL104) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL104) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if relocationSpacingRegex.MatchString(state.currentRawLine) {
		match := relocationSpacingRegex.FindString(state.currentRawLine)
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  fmt.Sprintf("spaces between relocation operator and its parenthesis in %q are banned", match),
		}
	}
	return nil
}

// Rule L105: gp_load_relaxation
type ruleL105 struct{}

func (r *ruleL105) ID() string       { return "L105" }
func (r *ruleL105) Name() string     { return "gp_load_relaxation" }
func (r *ruleL105) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL105) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	inst := strings.ToLower(st.instruction)
	if (inst == "la" || inst == "lla") && len(st.params) > 0 {
		reg := strings.TrimSpace(st.params[0])
		if reg == "gp" && state.relaxationEnabled {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  "loading global pointer (gp) must be performed with linker relaxation disabled (.option norelax)",
			}
		}
	}
	return nil
}

// Rule L106: csr_names
type ruleL106 struct{}

func (r *ruleL106) ID() string       { return "L106" }
func (r *ruleL106) Name() string     { return "csr_names" }
func (r *ruleL106) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL106) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	inst := strings.ToLower(st.instruction)
	var csr string
	switch inst {
	case "csrrw", "csrrs", "csrrc", "csrrwi", "csrrsi", "csrrci", "csrr", "csrs", "csrc", "csrsi", "csrci":
		if len(st.params) >= 2 {
			csr = strings.TrimSpace(st.params[1])
		}
	case "csrw", "csrwi":
		if len(st.params) >= 1 {
			csr = strings.TrimSpace(st.params[0])
		}
	}
	if csr != "" {
		if isNumeric(csr) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("raw numeric constant CSR %q is banned; use standard name or custom uppercase symbol", csr),
			}
		}
		// Custom symbols must be uppercase or start with CSR_
		if !isStandardCSR(csr) && !isAllUppercase(csr) && !strings.HasPrefix(csr, "CSR_") {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("custom CSR %q must be all-uppercase or start with 'CSR_'", csr),
			}
		}
	}
	return nil
}

func isStandardCSR(name string) bool {
	if strings.HasPrefix(name, "hpmcounter") || strings.HasPrefix(name, "mhpmcounter") ||
		strings.HasPrefix(name, "hpmcounterh") || strings.HasPrefix(name, "mhpmcounterh") ||
		strings.HasPrefix(name, "mhpmevent") ||
		strings.HasPrefix(name, "pmpcfg") || strings.HasPrefix(name, "pmpaddr") {
		pfx := ""
		if strings.HasPrefix(name, "mhpmcounterh") {
			pfx = "mhpmcounterh"
		} else if strings.HasPrefix(name, "mhpmcounter") {
			pfx = "mhpmcounter"
		} else if strings.HasPrefix(name, "hpmcounterh") {
			pfx = "hpmcounterh"
		} else if strings.HasPrefix(name, "hpmcounter") {
			pfx = "hpmcounter"
		} else if strings.HasPrefix(name, "mhpmevent") {
			pfx = "mhpmevent"
		} else if strings.HasPrefix(name, "pmpcfg") {
			pfx = "pmpcfg"
		} else if strings.HasPrefix(name, "pmpaddr") {
			pfx = "pmpaddr"
		}

		val := name[len(pfx):]
		var num int
		_, err := fmt.Sscanf(val, "%d", &num)
		if err == nil {
			if pfx == "pmpcfg" {
				return num >= 0 && num <= 15
			}
			if pfx == "pmpaddr" {
				return num >= 0 && num <= 63
			}
			return num >= 3 && num <= 31
		}
	}
	switch name {
	case "ustatus", "uie", "utvec", "uscratch", "uepc", "ucause", "utval", "uip",
		"fflags", "frm", "fcsr", "cycle", "time", "instret", "cycleh", "timeh", "instreth",
		"sstatus", "sedeleg", "sideleg", "sie", "stvec", "scounteren", "sscratch", "sepc",
		"scause", "stval", "sip", "satp",
		"hstatus", "hedeleg", "hideleg", "hie", "htvec", "hcounteren", "hgeie",
		"hscratch", "hepc", "hcause", "hbadaddr", "hip", "hgeip", "htinst", "hgatp",
		"mstatus", "misa", "medeleg", "mideleg", "mie", "mtvec", "mcounteren", "mstatush",
		"mscratch", "mepc", "mcause", "mtval", "mip", "mtinst", "mtval2", "mconfigptr",
		"mcycle", "mtime", "minstret", "mcycleh", "mtimeh", "minstreth",
		"tselect", "tdata1", "tdata2", "tdata3", "dcsr", "dpc", "dscratch0", "dscratch1":
		return true
	}
	return false
}

// Rule L107: jump_instruction_selection
type ruleL107 struct{}

func (r *ruleL107) ID() string       { return "L107" }
func (r *ruleL107) Name() string     { return "jump_instruction_selection" }
func (r *ruleL107) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL107) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	inst := strings.ToLower(st.instruction)
	if inst == "j" || inst == "jal" || inst == "jr" {
		target := ""
		if inst == "jal" && len(st.params) == 2 {
			target = strings.TrimSpace(st.params[1])
		} else if len(st.params) > 0 {
			target = strings.TrimSpace(st.params[0])
		}
		if target != "" {
			if isRegister(target) {
				return nil
			}
			isLocal := strings.HasPrefix(target, ".L_") || regexp.MustCompile(`^[0-9]+[fb]$`).MatchString(target)
			if !isLocal {
				return &Problem{
					Filename: state.filename,
					Line:     lineNum,
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Message:  fmt.Sprintf("near jump %s to non-local symbol %q is banned; use call/tail instead", inst, target),
				}
			}
		}
	}
	return nil
}

// Rule L108: pcrel_relocation_label
type ruleL108 struct{}

func (r *ruleL108) ID() string       { return "L108" }
func (r *ruleL108) Name() string     { return "pcrel_relocation_label" }
func (r *ruleL108) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL108) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	for _, p := range st.params {
		matches := pcrelLoRegex.FindAllStringSubmatch(p, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				lbl := strings.TrimSpace(m[1])
				if !strings.HasPrefix(lbl, ".L_") {
					return &Problem{
						Filename: state.filename,
						Line:     lineNum,
						RuleID:   r.ID(),
						RuleName: r.Name(),
						Message:  fmt.Sprintf("%%pcrel_lo operand %q must refer to a local label (starting with '.L_')", lbl),
					}
				}
			}
		}
	}
	return nil
}

// Rule L201: alignment_directives
type ruleL201 struct{}

func (r *ruleL201) ID() string       { return "L201" }
func (r *ruleL201) Name() string     { return "alignment_directives" }
func (r *ruleL201) Scope() RuleScope { return ScopeGas }
func (r *ruleL201) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if strings.ToLower(st.instruction) == ".align" {
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  "directive '.align' is banned; use '.p2align' or '.balign' instead",
		}
	}
	return nil
}

// Rule L202: extern_directive
type ruleL202 struct{}

func (r *ruleL202) ID() string       { return "L202" }
func (r *ruleL202) Name() string     { return "extern_directive" }
func (r *ruleL202) Scope() RuleScope { return ScopeGas }
func (r *ruleL202) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if strings.ToLower(st.instruction) == ".extern" {
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  "directive '.extern' is banned",
		}
	}
	return nil
}

// Rule L203: inline_binary_directives
type ruleL203 struct{}

func (r *ruleL203) ID() string       { return "L203" }
func (r *ruleL203) Name() string     { return "inline_binary_directives" }
func (r *ruleL203) Scope() RuleScope { return ScopeGas }
func (r *ruleL203) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	inst := strings.ToLower(st.instruction)
	if inst == ".word" || inst == ".long" {
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  fmt.Sprintf("directive %q is banned for inline binary data; favor .byte/.2byte/.4byte/.8byte", inst),
		}
	}
	return nil
}

// Rule L204: avoid_globl
type ruleL204 struct{}

func (r *ruleL204) ID() string       { return "L204" }
func (r *ruleL204) Name() string     { return "avoid_globl" }
func (r *ruleL204) Scope() RuleScope { return ScopeGas }
func (r *ruleL204) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if strings.ToLower(st.instruction) == ".globl" {
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  "legacy directive '.globl' is banned; use '.global' instead",
		}
	}
	return nil
}

// Rule L205: leb128_constant_expression
type ruleL205 struct{}

func (r *ruleL205) ID() string       { return "L205" }
func (r *ruleL205) Name() string     { return "leb128_constant_expression" }
func (r *ruleL205) Scope() RuleScope { return ScopeGas }
func (r *ruleL205) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	inst := strings.ToLower(st.instruction)
	if (inst == ".uleb128" || inst == ".sleb128") && len(st.params) > 0 {
		expr := strings.TrimSpace(st.params[0])
		if identifierRegex.MatchString(expr) && !isNumeric(expr) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("%s must be used with a constant or symbol difference, not a raw symbol %q", inst, expr),
			}
		}
	}
	return nil
}

// Rule L206: avoid_space_skip_directives
type ruleL206 struct{}

func (r *ruleL206) ID() string       { return "L206" }
func (r *ruleL206) Name() string     { return "avoid_space_skip_directives" }
func (r *ruleL206) Scope() RuleScope { return ScopeGas }
func (r *ruleL206) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	inst := strings.ToLower(st.instruction)
	if inst == ".space" || inst == ".skip" {
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  fmt.Sprintf("directive %q is banned; use '.zero' instead", inst),
		}
	}
	return nil
}

// Rule L207: operator_precedence_parentheses
type ruleL207 struct{}

func (r *ruleL207) ID() string       { return "L207" }
func (r *ruleL207) Name() string     { return "operator_precedence_parentheses" }
func (r *ruleL207) Scope() RuleScope { return ScopeGas }
func (r *ruleL207) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	for _, p := range st.params {
		if hasPrecedenceViolation(p) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("explicit parentheses required for operator precedence in %q", p),
			}
		}
	}
	return nil
}

// Rule L208: end_directive_last
type ruleL208 struct{}

func (r *ruleL208) ID() string       { return "L208" }
func (r *ruleL208) Name() string     { return "end_directive_last" }
func (r *ruleL208) Scope() RuleScope { return ScopeGas }
func (r *ruleL208) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if strings.ToLower(st.instruction) == ".end" {
		// Verify if this is not the last non-comment statement in file
		// Done in Lint() by checking subsequent lines/statements
		// But here, we can set custom state to check later or perform the forward-scan in Check if needed.
		// Since we have state.rawLines, we can scan from lineNum to end:
		for i := lineNum; i < len(state.rawLines); i++ {
			trimmed := strings.TrimSpace(state.rawLines[i])
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "/*") {
				return &Problem{
					Filename: state.filename,
					Line:     lineNum,
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Message:  "directive '.end' must be the last statement in the file",
				}
			}
		}
	}
	return nil
}

// Rule L301: local_labels
type ruleL301 struct{}

func (r *ruleL301) ID() string       { return "L301" }
func (r *ruleL301) Name() string     { return "local_labels" }
func (r *ruleL301) Scope() RuleScope { return ScopeGas }
func (r *ruleL301) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.isLabel() {
		name := strings.TrimSuffix(st.instruction, ":")
		if strings.HasPrefix(name, ".L") && !strings.HasPrefix(name, ".L_") {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("local label %q must start with '.L_' prefix", name),
			}
		}
	}
	return nil
}

// Rule L302: current_point_label
type ruleL302 struct{}

func (r *ruleL302) ID() string       { return "L302" }
func (r *ruleL302) Name() string     { return "current_point_label" }
func (r *ruleL302) Scope() RuleScope { return ScopeGas }
func (r *ruleL302) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	for _, p := range st.params {
		if hasCurrentPoint(p) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("use of current point '.' label in operand %q is banned", p),
			}
		}
	}
	return nil
}

// Rule L303: pointer_offset_shorthand
type ruleL303 struct{}

func (r *ruleL303) ID() string       { return "L303" }
func (r *ruleL303) Name() string     { return "pointer_offset_shorthand" }
func (r *ruleL303) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL303) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	for _, p := range st.params {
		if offsetShorthandRegex.MatchString(p) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("memory pointer shorthand %q must include explicit offset (e.g. 0%s)", p, p),
			}
		}
	}
	return nil
}

// Rule L304: option_push_pop
type ruleL304 struct{}

func (r *ruleL304) ID() string       { return "L304" }
func (r *ruleL304) Name() string     { return "option_push_pop" }
func (r *ruleL304) Scope() RuleScope { return ScopeRiscv }
func (r *ruleL304) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		if len(state.pushLines) > 0 {
			return &Problem{
				Filename: state.filename,
				Line:     state.pushLines[len(state.pushLines)-1],
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  "unbalanced '.option push' without matching '.option pop'",
			}
		}
		return nil
	}
	if st.instruction == ".option" && len(st.params) > 0 && st.params[0] == "pop" {
		if len(state.pushLines) == 0 {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  "unbalanced '.option pop' without matching '.option push'",
			}
		}
	}
	return nil
}

// Rule L305: symbol_preamble_footer
type ruleL305 struct{}

func (r *ruleL305) ID() string       { return "L305" }
func (r *ruleL305) Name() string     { return "symbol_preamble_footer" }
func (r *ruleL305) Scope() RuleScope { return ScopeGas }
func (r *ruleL305) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		for name, line := range state.activeFunctions {
			return &Problem{
				Filename: state.filename,
				Line:     line,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("missing matching '.size' footer for function %q", name),
			}
		}
		return nil
	}
	if st.instruction == ".type" && len(st.params) >= 2 {
		if strings.Contains(st.params[1], "function") {
			// Check if alignment directive preceded it
			precededByAlign := false
			if state.lastNonCommentSt != nil {
				prevInst := strings.ToLower(state.lastNonCommentSt.instruction)
				if prevInst == ".p2align" || prevInst == ".balign" || prevInst == ".align" {
					precededByAlign = true
				}
			}
			if !precededByAlign {
				return &Problem{
					Filename: state.filename,
					Line:     lineNum,
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Message:  fmt.Sprintf("function %q type declaration must be preceded by alignment directive (.p2align/.balign)", st.params[0]),
				}
			}
		}
	}
	return nil
}

// Rule L306: cfi_start_end_balance
type ruleL306 struct{}

func (r *ruleL306) ID() string       { return "L306" }
func (r *ruleL306) Name() string     { return "cfi_start_end_balance" }
func (r *ruleL306) Scope() RuleScope { return ScopeGas }
func (r *ruleL306) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		if len(state.cfiStartLines) > 0 {
			return &Problem{
				Filename: state.filename,
				Line:     state.cfiStartLines[len(state.cfiStartLines)-1],
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  "unbalanced '.cfi_startproc' without matching '.cfi_endproc'",
			}
		}
		return nil
	}
	if st.instruction == ".cfi_endproc" && len(state.cfiStartLines) == 0 {
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  "unbalanced '.cfi_endproc' without matching '.cfi_startproc'",
		}
	}
	return nil
}

// Rule L307: macro_balance
type ruleL307 struct{}

func (r *ruleL307) ID() string       { return "L307" }
func (r *ruleL307) Name() string     { return "macro_balance" }
func (r *ruleL307) Scope() RuleScope { return ScopeGas }
func (r *ruleL307) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		if len(state.macroStartLines) > 0 {
			return &Problem{
				Filename: state.filename,
				Line:     state.macroStartLines[len(state.macroStartLines)-1],
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  "unbalanced '.macro' without matching '.endm'",
			}
		}
		return nil
	}
	if st.instruction == ".endm" && len(state.macroStartLines) == 0 {
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  "unbalanced '.endm' without matching '.macro'",
		}
	}
	return nil
}

// Rule L308: instruction_sequence_termination
type ruleL308 struct{}

func (r *ruleL308) ID() string       { return "L308" }
func (r *ruleL308) Name() string     { return "instruction_sequence_termination" }
func (r *ruleL308) Scope() RuleScope { return ScopeGas }
func (r *ruleL308) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		if state.inInstructionSequence && !state.lastInstructionWasTerminator && !state.lastLineCommentContainsFallthrough {
			line := state.lastInstructionLine
			if line == 0 {
				line = state.lastNonCommentLine
			}
			return &Problem{
				Filename: state.filename,
				Line:     line,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  "instruction sequence or code block must end with a terminator (like ret, j, tail, mret, unimp) or explicit 'fallthrough' comment",
			}
		}
		return nil
	}
	if st.isLabel() {
		labelName := strings.TrimSuffix(st.instruction, ":")
		if !isNumeric(labelName) && !strings.HasPrefix(labelName, ".L") {
			// Transitioning to new function/global symbol block
			if state.inInstructionSequence && !state.lastInstructionWasTerminator && !state.lastLineCommentContainsFallthrough {
				line := state.lastInstructionLine
				if line == 0 {
					line = state.lastNonCommentLine
				}
				return &Problem{
					Filename: state.filename,
					Line:     line,
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Message:  "instruction sequence or code block must end with a terminator (like ret, j, tail, mret, unimp) or explicit 'fallthrough' comment",
				}
			}
		}
	}
	return nil
}

// Rule L309: function_doxygen_comment
type ruleL309 struct{}

func (r *ruleL309) ID() string       { return "L309" }
func (r *ruleL309) Name() string     { return "function_doxygen_comment" }
func (r *ruleL309) Scope() RuleScope { return ScopeGas }
func (r *ruleL309) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.isLabel() {
		labelName := strings.TrimSuffix(st.instruction, ":")
		if !isNumeric(labelName) && !strings.HasPrefix(labelName, ".L") {
			// If it's global or declared as function, it needs Doxygen
			isFunc := state.globalSymbols[labelName]
			if isFunc {
				// Find line index in rawLines
				labelIdx := lineNum - 1
				if !hasPrecedingDoxygen(state.rawLinesToLintLines(lineNum), labelIdx) {
					return &Problem{
						Filename: state.filename,
						Line:     lineNum,
						RuleID:   r.ID(),
						RuleName: r.Name(),
						Message:  fmt.Sprintf("global function %q must have preceding Doxygen comment block (starts with '/**' or '///')", labelName),
					}
				}
			}
		}
	}
	return nil
}

func (state *lintState) rawLinesToLintLines(lineNum int) []lintLine {
	var l []lintLine
	for idx, rl := range state.rawLines {
		l = append(l, lintLine{
			LineNum: idx + 1,
			Raw:     rl,
		})
	}
	return l
}

// Rule L310: avoid_hash_and_at_comments
type ruleL310 struct{}

func (r *ruleL310) ID() string       { return "L310" }
func (r *ruleL310) Name() string     { return "avoid_hash_and_at_comments" }
func (r *ruleL310) Scope() RuleScope { return ScopeGas }
func (r *ruleL310) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.commentMark == "#" || st.commentMark == "@" {
		// Verify not a preprocessor line
		if st.commentMark == "#" && isPreProcessorInstruction(st.instruction) {
			return nil
		}
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  fmt.Sprintf("comments using %q marker are banned; use '//' or '/* */'", st.commentMark),
		}
	}
	return nil
}

// Rule L311: double_label_declaration
type ruleL311 struct{}

func (r *ruleL311) ID() string       { return "L311" }
func (r *ruleL311) Name() string     { return "double_label_declaration" }
func (r *ruleL311) Scope() RuleScope { return ScopeGas }
func (r *ruleL311) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.isLabel() {
		if state.lastWasLabel {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  "consecutive labels declared on the same address without instructions/directives are banned",
			}
		}
	}
	return nil
}

// Rule L312: reserved_label_names
type ruleL312 struct{}

func (r *ruleL312) ID() string       { return "L312" }
func (r *ruleL312) Name() string     { return "reserved_label_names" }
func (r *ruleL312) Scope() RuleScope { return ScopeGas }
func (r *ruleL312) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.isLabel() {
		name := strings.TrimSuffix(st.instruction, ":")
		if isReservedName(name) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("label name %q conflicts with reserved keyword, instruction, or register", name),
			}
		}
	}
	if st.instruction == ".macro" && len(st.params) > 0 {
		name := strings.TrimSpace(st.params[0])
		if isReservedName(name) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("macro name %q conflicts with reserved keyword, instruction, or register", name),
			}
		}
	}
	return nil
}

// Rule L313: unreachable_code
type ruleL313 struct{}

func (r *ruleL313) ID() string       { return "L313" }
func (r *ruleL313) Name() string     { return "unreachable_code" }
func (r *ruleL313) Scope() RuleScope { return ScopeGas }
func (r *ruleL313) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.isLabel() {
		state.lastWasTerminatorInst = false
		return nil
	}
	if !st.isGasDirective() && !st.isLabel() && !st.isPreProcessor() && !st.function {
		if state.lastWasTerminatorInst {
			state.lastWasTerminatorInst = false // flag once
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  "unreachable code: instruction follows a terminator without a preceding label",
			}
		}
	}
	return nil
}

// Rule L314: single_return_statement
type ruleL314 struct{}

func (r *ruleL314) ID() string       { return "L314" }
func (r *ruleL314) Name() string     { return "single_return_statement" }
func (r *ruleL314) Scope() RuleScope { return ScopeGas }
func (r *ruleL314) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if strings.ToLower(st.instruction) == "ret" {
		state.funcReturnCount++
		if state.funcReturnCount > 1 {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("multiple return statements in function %q are banned", state.currentFunc),
			}
		}
	}
	return nil
}

// Rule L315: no_fallthrough_to_function
type ruleL315 struct{}

func (r *ruleL315) ID() string       { return "L315" }
func (r *ruleL315) Name() string     { return "no_fallthrough_to_function" }
func (r *ruleL315) Scope() RuleScope { return ScopeGas }
func (r *ruleL315) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.isLabel() {
		labelName := strings.TrimSuffix(st.instruction, ":")
		if !isNumeric(labelName) && !strings.HasPrefix(labelName, ".L") {
			if state.inInstructionSequence && !state.lastInstructionWasTerminator && !state.lastLineCommentContainsFallthrough {
				line := state.lastInstructionLine
				if line == 0 {
					line = state.lastNonCommentLine
				}
				return &Problem{
					Filename: state.filename,
					Line:     line,
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Message:  fmt.Sprintf("code falls through into new function label %q from previous block", labelName),
				}
			}
		}
	}
	return nil
}

// Rule L316: no_jump_out_of_function
type ruleL316 struct{}

func (r *ruleL316) ID() string       { return "L316" }
func (r *ruleL316) Name() string     { return "no_jump_out_of_function" }
func (r *ruleL316) Scope() RuleScope { return ScopeGas }
func (r *ruleL316) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	inst := strings.ToLower(st.instruction)
	if inst == "j" || inst == "jmp" {
		if len(st.params) > 0 {
			target := strings.TrimSpace(st.params[0])
			if !isRegister(target) && !strings.HasPrefix(target, ".L") && !regexp.MustCompile(`^[0-9]+[fb]$`).MatchString(target) {
				if state.currentFunc != "" && target != state.currentFunc {
					return &Problem{
						Filename: state.filename,
						Line:     lineNum,
						RuleID:   r.ID(),
						RuleName: r.Name(),
						Message:  fmt.Sprintf("direct jump %s to target %q is jumping out of function %q", inst, target, state.currentFunc),
					}
				}
			}
		}
	}
	return nil
}

// Rule L317: no_recursive_calls
type ruleL317 struct{}

func (r *ruleL317) ID() string       { return "L317" }
func (r *ruleL317) Name() string     { return "no_recursive_calls" }
func (r *ruleL317) Scope() RuleScope { return ScopeGas }
func (r *ruleL317) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" || st.isGasDirective() || st.isLabel() || st.isPreProcessor() || st.function {
		return nil
	}
	inst := strings.ToLower(st.instruction)
	if inst == "call" || inst == "jal" || inst == "tail" || inst == "j" || inst == "jmp" {
		target := ""
		if inst == "jal" && len(st.params) == 2 {
			target = strings.TrimSpace(st.params[1])
		} else if len(st.params) > 0 {
			target = strings.TrimSpace(st.params[0])
		}
		if target != "" && state.currentFunc != "" && target == state.currentFunc {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("recursive call %s to own function %q is banned", inst, target),
			}
		}
	}
	return nil
}

// Rule L318: copyright_and_license
type ruleL318 struct{}

func (r *ruleL318) ID() string       { return "L318" }
func (r *ruleL318) Name() string     { return "copyright_and_license" }
func (r *ruleL318) Scope() RuleScope { return ScopeAll }
func (r *ruleL318) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		requireSpdx := true
		if val, ok := state.custom["copyright_require_spdx"].(bool); ok {
			requireSpdx = val
		}
		if !state.hasCopyright || (requireSpdx && !state.hasSPDX) {
			msg := "file must start with a copyright notice"
			if requireSpdx {
				msg = "file must start with a copyright notice and an SPDX license identifier"
			}
			return &Problem{
				Filename: state.filename,
				Line:     1,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  msg,
			}
		}
	}
	return nil
}

// Rule L319: developer_name_length
type ruleL319 struct{}

func (r *ruleL319) ID() string       { return "L319" }
func (r *ruleL319) Name() string     { return "developer_name_length" }
func (r *ruleL319) Scope() RuleScope { return ScopeGas }
func (r *ruleL319) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.isLabel() {
		name := strings.TrimSuffix(st.instruction, ":")
		if !isNumeric(name) && len(name) >= 31 {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("label name %q has %d characters (must be fewer than 31)", name, len(name)),
			}
		}
	}
	if st.instruction == ".macro" && len(st.params) > 0 {
		name := strings.TrimSpace(st.params[0])
		if len(name) >= 31 {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("macro name %q has %d characters (must be fewer than 31)", name, len(name)),
			}
		}
	}
	return nil
}

// Rule L320: line_length_limit
type ruleL320 struct{}

func (r *ruleL320) ID() string       { return "L320" }
func (r *ruleL320) Name() string     { return "line_length_limit" }
func (r *ruleL320) Scope() RuleScope { return ScopeAll }
func (r *ruleL320) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	// Avoid reporting multiple times on a line split by semicolon
	if state.custom["l320_reported"] == lineNum {
		return nil
	}
	if utf8.RuneCountInString(state.currentRawLine) > 120 {
		state.custom["l320_reported"] = lineNum
		return &Problem{
			Filename: state.filename,
			Line:     lineNum,
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Message:  fmt.Sprintf("line exceeds 120 characters limit (%d characters)", utf8.RuneCountInString(state.currentRawLine)),
		}
	}
	return nil
}

// Rule L321: label_naming_style
type ruleL321 struct{}

func (r *ruleL321) ID() string       { return "L321" }
func (r *ruleL321) Name() string     { return "label_naming_style" }
func (r *ruleL321) Scope() RuleScope { return ScopeGas }
func (r *ruleL321) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	// Get normal options or retrieve from state to check target casing style
	if st.isLabel() {
		name := strings.TrimSuffix(st.instruction, ":")
		if !isNumeric(name) && !strings.HasPrefix(name, ".L") {
			// Retrieve active casing style from state or default:
			// Since we can't access opts directly, we should read it from state.custom which holds the active style if we pass it,
			// or we can pass the config properties.
			// Let's make sure the naming styles are passed from Lint to state!
			// We can add them to state.custom during Lint initialization.
			style, _ := state.custom["label_naming_style"].(string)
			if style == "" {
				style = "snake_case" // fallback default
			}
			if !matchesNamingStyle(name, style) {
				return &Problem{
					Filename: state.filename,
					Line:     lineNum,
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Message:  fmt.Sprintf("label name %q does not match configured case style %q", name, style),
				}
			}
		}
	}
	return nil
}

// Rule L322: macro_naming_style
type ruleL322 struct{}

func (r *ruleL322) ID() string       { return "L322" }
func (r *ruleL322) Name() string     { return "macro_naming_style" }
func (r *ruleL322) Scope() RuleScope { return ScopeGas }
func (r *ruleL322) Check(st statement, lineNum int, state *lintState) *Problem {
	if st.instruction == ".end_of_file" {
		return nil
	}
	if st.instruction == ".macro" && len(st.params) > 0 {
		name := strings.TrimSpace(st.params[0])
		style, _ := state.custom["macro_naming_style"].(string)
		if style == "" {
			style = "UPPER_SNAKE_CASE" // fallback default
		}
		if !matchesMacroNamingStyle(name, style) {
			return &Problem{
				Filename: state.filename,
				Line:     lineNum,
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Message:  fmt.Sprintf("macro name %q does not match configured case style %q", name, style),
			}
		}
	}
	return nil
}

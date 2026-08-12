package asmfmt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Format the input and return the formatted data.
// If any error is encountered, no data will be returned.
func Format(in io.Reader) ([]byte, error) {
	return FormatWithOptions(in, DefaultOptions())
}

// FormatWithOptions formats the input and return the formatted data.
// If any error is encountered, no data will be returned.
func FormatWithOptions(in io.Reader, opts Options) ([]byte, error) {
	nopts, err := opts.normalize()
	if err != nil {
		return nil, err
	}
	src := bufio.NewReaderSize(in, 512<<10)
	dst := &bytes.Buffer{}
	state := fstate{
		out:         dst,
		defines:     make(map[string]struct{}),
		opts:        nopts,
		style:       nopts.sourceStyle,
		detectStyle: nopts.sourceStyle == styleUnknown,
	}
	for {
		data, _, err := src.ReadLine()
		if err == io.EOF {
			state.flush()
			break
		}
		if err != nil {
			return nil, err
		}
		err = state.addLine(data)
		if err != nil {
			return nil, err
		}
	}
	return dst.Bytes(), nil
}

type fstate struct {
	out               *bytes.Buffer
	insideBlock       bool // Block comment
	blockStandalone   bool
	blockIndentation  int
	blockCanonical    bool
	blockBuffered     bool
	blockBuffer       []string
	indentation       int // Indentation level
	lastEmpty         bool
	lastComment       bool
	lastBlockComment  bool
	lastStar          bool // Block comment, last line started with a star.
	lastLabel         bool
	anyContents       bool
	lastContinued     bool // Last line continued
	gasBlock          int
	inMacro           bool
	altMacro          bool
	style             sourceStyle
	detectStyle       bool
	queued            []statement
	comments          []string
	defines           map[string]struct{}
	opts              normalizedOptions
	blankLines        int
	emptyAfterComment bool
}

type statement struct {
	instruction string
	params      []string // Parameters
	comment     string   // Without slashes
	commentMark string   // Line comment marker.
	function    bool     // Probably define call
	continued   bool     // Multiline statement, continues on next line
	contComment bool     // Multiline statement, comment only
}

// Add a new input line.
// Since you are looking at ths code:
// This code has grown over a considerable amount of time,
// and deserves a rewrite with proper parsing instead of this hodgepodge.
// Its output is stable, and could be used as reference for a rewrite.
func (f *fstate) addLine(b []byte) error {
	if bytes.Contains(b, []byte{0}) {
		return fmt.Errorf("zero (0) byte in input. file is unlikely an assembler file")
	}
	raw := string(b)
	s := raw
	// Inside block comment
	if f.insideBlock {
		trimmed := strings.TrimSpace(s)
		if f.blockBuffered {
			f.blockBuffer = append(f.blockBuffer, s)
			if strings.Contains(s, "*/") {
				f.emitBufferedBlockComment()
			}
			f.blankLines = 0
			f.lastEmpty = false
			f.lastComment = true
			f.lastBlockComment = true
			return nil
		}
		if f.blockStandalone && f.blockCanonical {
			defer func() {
				f.blankLines = 0
				f.lastEmpty = false
				f.lastComment = true
				f.lastBlockComment = true
			}()
			if strings.Contains(trimmed, "*/") {
				f.writeIndentLevel(f.blockIndentation)
				f.out.WriteString(" */\n")
				f.insideBlock = false
				f.blockStandalone = false
				f.blockCanonical = false
				return nil
			}
			body := strings.TrimSpace(trimBlockCommentLeader(trimmed))
			f.writeIndentLevel(f.blockIndentation)
			if body == "" {
				f.out.WriteString(" *\n")
			} else {
				f.out.WriteString(" * " + body + "\n")
			}
			return nil
		}
		if isDecorativeBlockCommentLine(strings.TrimSpace(s)) {
			defer func() {
				f.blankLines = 0
				f.lastEmpty = false
				f.lastComment = true
				f.lastBlockComment = true
			}()
			if f.blockStandalone {
				f.writeIndentLevel(f.blockIndentation)
				fmt.Fprintln(f.out, trimmed)
			} else {
				fmt.Fprintln(f.out, s)
			}
			if strings.Contains(s, "*/") {
				f.insideBlock = false
				f.blockStandalone = false
				f.blockCanonical = false
			}
			return nil
		}
		defer func() {
			f.blankLines = 0
			f.lastEmpty = false
			f.lastComment = true
			f.lastBlockComment = true
		}()
		if strings.Contains(s, "*/") {
			if f.blockStandalone && trimmed == "*/" {
				f.writeIndentLevel(f.blockIndentation)
				f.out.WriteString("*/\n")
				f.insideBlock = false
				f.blockStandalone = false
				f.blockCanonical = false
				return nil
			}
			ends := strings.Index(s, "*/")
			end := s[:ends]
			if strings.HasPrefix(strings.TrimSpace(s), "*") && f.lastStar {
				end = strings.TrimSpace(end) + " "
			}
			end = end + "*/"
			f.insideBlock = false
			f.blockStandalone = false
			f.blockCanonical = false
			s = strings.TrimSpace(s[ends+2:])
			if strings.HasSuffix(s, "\\") {
				end = end + " \\"
				if len(s) == 1 {
					s = ""
				}
			}
			f.out.WriteString(end + "\n")
			if len(s) == 0 {
				return nil
			}
		} else {
			if f.blockStandalone {
				if strings.HasPrefix(trimmed, "*") {
					f.writeIndentLevel(f.blockIndentation)
					f.out.WriteByte(' ')
					f.out.WriteString(trimmed)
					f.out.WriteByte('\n')
					f.lastStar = true
					return nil
				}
				f.writeIndentLevel(f.blockIndentation)
				fmt.Fprintln(f.out, trimmed)
				f.lastStar = false
				return nil
			}
			// Insert a space on lines that begin with '*'
			if strings.HasPrefix(strings.TrimSpace(s), "*") {
				s = strings.TrimSpace(s)
				f.out.WriteByte(' ')
				f.lastStar = true
			} else {
				f.lastStar = false
			}
			fmt.Fprintln(f.out, s)
			return nil
		}
	}
	s = strings.TrimSpace(s)
	f.observeStyle(s)

	// Comment is the only line content.
	if mark, ok := isStandaloneCommentLine(s, f.style); ok {
		if f.lastEmpty && f.emptyAfterComment && f.out.Len() > 0 {
			f.out.Truncate(f.out.Len() - 1)
			f.lastEmpty = false
			f.blankLines = 0
		}
		// Non-comment content is now added.
		defer func() {
			f.anyContents = true
			f.blankLines = 0
			f.emptyAfterComment = false
			f.lastEmpty = false
			f.lastStar = false
			f.lastBlockComment = false
		}()

		s = strings.TrimPrefix(s, mark)
		if len(f.queued) > 0 {
			f.flush()
		}
		// Newline before comments
		if len(f.comments) == 0 {
			f.newLine(f.opts.newlineBeforeComments)
		}

		// Preserve whitespace if the first character after the comment
		// is a whitespace
		ts := strings.TrimSpace(s)
		var q string
		preserveRaw := (len(s) > 0 && strings.ContainsAny(string(s[0]), `+/`)) || (mark == "//" && len(s) >= 8 && s[:8] == "go:build")
		if preserveRaw || (f.opts.lineCommentSpace && ts != s && len(ts) > 0) {
			q = fmt.Sprint(f.preferredCommentMark(mark) + s)
		} else if len(ts) > 0 {
			q = fmt.Sprint(f.preferredCommentMark(mark) + f.commentSeparator() + ts)
		} else {
			q = fmt.Sprint(f.preferredCommentMark(mark))
		}
		f.comments = append(f.comments, q)
		f.lastComment = true
		f.lastBlockComment = false
		return nil
	}

	rawIndented := len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t')

	// Handle end-of blockcomments.
	if starts := findBlockCommentStart(s); starts >= 0 && !strings.HasSuffix(s, `\`) {
		ends := strings.Index(s, "*/")
		lineComment := strings.Index(s, "//")
		if lineComment >= 0 {
			if lineComment < starts {
				goto exitcomm
			}
			if lineComment < ends && !f.insideBlock {
				goto exitcomm
			}
			if ends > starts && ends < lineComment {
				// If there is something left between the end and the line comment, keep it.
				if len(strings.TrimSpace(s[ends:lineComment])) > 0 {
					goto exitcomm
				}
			}
		}
		pre := s[:starts]
		pre = strings.TrimSpace(pre)

		if len(pre) > 0 {
			if strings.HasSuffix(s, `\`) {
				goto exitcomm
			}
			// Add items before the comment section as a line.
			if ends > starts && ends >= len(s)-2 && f.opts.convertSingleLineBlockComment {
				comm := strings.TrimSpace(s[starts+2 : ends])
				return f.addLine([]byte(pre + " " + f.preferredCommentMark("//") + f.commentSeparator() + comm))
			}
			err := f.addLine([]byte(pre))
			if err != nil {
				return err
			}
		}

		f.flush()

		if starts == 0 && ends < 0 && isDecorativeBlockCommentLine(strings.TrimSpace(s)) {
			fmt.Fprintln(f.out, s)
			f.insideBlock = true
			f.blockStandalone = rawIndented
			f.blockIndentation = f.indentation
			f.blockCanonical = false
			f.lastComment = true
			f.lastBlockComment = true
			f.lastStar = false
			return nil
		}

		// Convert single line /* comment */ to // Comment
		if ends > starts && ends >= len(s)-2 && f.opts.convertSingleLineBlockComment {
			return f.addLine([]byte(f.preferredCommentMark("//") + f.commentSeparator() + strings.TrimSpace(s[starts+2:ends])))
		}

		// Comments inside multiline defines.
		if strings.HasSuffix(s, `\`) {
			f.indent()
			s = strings.TrimSpace(strings.TrimSuffix(s, `\`)) + ` \`
		}

		// Otherwise output
		if starts == 0 && len(pre) == 0 && rawIndented && ends < 0 {
			f.insideBlock = true
			f.blockBuffered = true
			f.blockIndentation = f.indentation
			f.blockBuffer = []string{raw}
			f.lastComment = true
			f.lastBlockComment = true
			f.lastStar = false
			return nil
		}
		if starts == 0 && len(pre) == 0 {
			f.blockStandalone = false
			f.blockCanonical = false
			fmt.Fprint(f.out, "/*")
			s = strings.TrimSpace(s[starts+2:])
			f.insideBlock = ends < 0
			f.lastComment = true
			f.lastBlockComment = true
			f.lastStar = true
			if len(s) == 0 {
				f.out.WriteByte('\n')
				return nil
			}
			f.out.WriteByte(' ')
			f.out.WriteString(s + "\n")
			return nil
		} else {
			f.blockStandalone = false
			f.blockCanonical = false
		}
		fmt.Fprint(f.out, "/*")
		s = strings.TrimSpace(s[starts+2:])
		f.insideBlock = ends < 0
		f.lastComment = true
		f.lastBlockComment = true
		f.lastStar = true
		if len(s) == 0 {
			f.out.WriteByte('\n')
			return nil
		}
		f.out.WriteByte(' ')
		f.out.WriteString(s + "\n")
		return nil
	}
exitcomm:

	if len(s) == 0 {
		f.flush()

		// No more than two empty lines in a row
		// cannot start with NL
		if f.blankLines >= f.opts.maxBlankLines || !f.anyContents {
			return nil
		}
		if f.lastContinued {
			f.indentation = 0
			f.lastContinued = false
		}
		f.emptyAfterComment = f.lastComment && f.lastBlockComment
		f.blankLines++
		f.lastEmpty = true
		return f.out.WriteByte('\n')
	}
	f.blankLines = 0
	f.emptyAfterComment = false

	if !f.inMacro && shouldSplitSemicolonStatementsForStyle(s, f.style, f.opts.splitSemicolonStatements) {
		parts := splitStatements(s, f.style)
		if len(parts) > 1 {
			for _, part := range parts {
				if err := f.addLine([]byte(part)); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Non-comment content is now added.
	defer func() {
		f.anyContents = true
		f.blankLines = 0
		f.lastEmpty = false
		f.lastStar = false
		f.lastComment = false
		f.lastBlockComment = false
	}()

	var st *statement
	if isConservativeUnknownGasDirective(s) {
		st = &statement{instruction: s, function: true}
	} else if f.inMacro && isMacroBodyText(s) {
		st = &statement{instruction: s, function: true}
	} else {
		st = newStatementWithStyle(s, f.defines, f.style)
	}
	if st == nil {
		return nil
	}
	defer f.updateMacroState(*st)
	if def := st.define(); def != "" {
		f.defines[def] = struct{}{}
	}
	if st.instruction == "package" {
		if _, ok := f.defines["package"]; !ok {
			return fmt.Errorf("package instruction found. Go files are not supported")
		}
	}

	// Move anything that isn't a comment to the next line
	if f.opts.labelsAlwaysOnOwnLine && st.isLabel() && len(st.params) > 0 && !st.continued {
		idx := strings.Index(s, ":")
		st = newStatement(s[:idx+1], f.defines)
		defer f.addLine([]byte(s[idx+1:]))
	}

	if st.isGasBlockMiddle() || st.isGasBlockEnd() {
		f.flush()
		if f.gasBlock > 0 {
			f.gasBlock--
		}
		if f.indentation > f.gasBlock {
			f.indentation = f.gasBlock
		}
	}

	// Should this line be at level 0?
	if st.level0() && !(st.continued && f.lastContinued) {
		if st.isTEXT() && len(f.queued) == 0 && len(f.comments) > 0 {
			f.indentation = 0
		}
		if st.isGasBlockStart() || st.isGasBlockMiddle() {
			f.flush()
			f.indentation = f.gasBlock
		}
		prevIndentation := f.indentation
		f.flush()
		if st.isGasZeroDirective() {
			if f.opts.indentGASDirectives {
				f.indentation = f.instructionIndentation()
			} else {
				f.indentation = f.gasBlock
			}
		}

		// Add newline before jump targets, but not before GAS directives
		// used inside an instruction stream.
		if !st.isGasDirective() {
			allowNewline := f.opts.newlineBeforeLabels || !st.isLabel()
			if st.isNumericLocalLabel() {
				allowNewline = false
			}
			f.newLine(allowNewline)
		}

		if !st.isGasDirective() {
			f.indentation = 0
		} else if st.isGasZeroDirective() {
			if f.opts.indentGASDirectives {
				f.indentation = f.instructionIndentation()
			} else {
				f.indentation = f.gasBlock
			}
		}
		f.queued = append(f.queued, *st)
		f.flush()

		if st.isGasBlockStart() || st.isGasBlockMiddle() {
			f.gasBlock++
			f.indentation = f.gasBlock
		} else if st.isGasDirective() {
			if f.opts.indentGASDirectives && st.isGasZeroDirective() {
				f.indentation = f.instructionIndentation()
			} else if st.isGasZeroDirective() {
				f.indentation = prevIndentation
			} else {
				f.indentation = f.gasBlock
			}
		} else if st.isPreProcessor() {
			f.indentation = prevIndentation
		} else if !st.isGlobal() {
			f.indentation = 1
		}
		f.lastLabel = true
		return nil
	}

	defer func() {
		f.lastLabel = false
	}()
	f.queued = append(f.queued, *st)
	if st.isTerminator() || (f.lastContinued && !st.continued) {
		// Terminators should always be at level 1
		f.indentation = 1
		if f.gasBlock > f.indentation {
			f.indentation = f.gasBlock
		}
		f.flush()
		if f.style != styleGas && f.style != styleRiscvGas {
			f.indentation = f.gasBlock
		}
	} else if st.isCommand() {
		// handles cases where a JMP/RET isn't a terminator
		f.indentation = 1
	}
	f.lastContinued = st.continued
	return nil
}

func (f *fstate) instructionIndentation() int {
	if f.gasBlock > 0 {
		return f.gasBlock
	}
	return 1
}

// indent the current line with current indentation.
func (f *fstate) indent() {
	for i := 0; i < f.indentation; i++ {
		f.writeIndentLevel(1)
	}
}

func (f *fstate) writeIndentLevel(level int) {
	for i := 0; i < level; i++ {
		if f.opts.indentStyle == "space" {
			for j := 0; j < f.opts.indentWidth; j++ {
				f.out.WriteByte(' ')
			}
			continue
		}
		f.out.WriteByte('\t')
	}
}

func (f *fstate) indentString(level int) string {
	if level <= 0 {
		return ""
	}
	if f.opts.indentStyle == "space" {
		return strings.Repeat(" ", level*f.opts.indentWidth)
	}
	return strings.Repeat("\t", level)
}

// flush any queued comments and commands
func (f *fstate) flush() {
	for _, line := range f.comments {
		body := line
		if strings.HasPrefix(body, "//") {
			body = body[2:]
		} else if strings.HasPrefix(body, "#") {
			body = body[1:]
		}
		if isCommentedPreProcessor(body) {
			fmt.Fprintln(f.out, line)
		} else {
			f.indent()
			fmt.Fprintln(f.out, line)
		}
	}
	f.comments = nil
	f.opts.sourceStyle = f.style
	s := formatStatements(f.queued, f.opts)
	for _, line := range s {
		f.indent()
		fmt.Fprintln(f.out, line)
	}
	f.queued = nil
}

func (f *fstate) updateMacroState(st statement) {
	switch st.instruction {
	case ".macro":
		f.inMacro = true
	case ".endm":
		f.inMacro = false
	case ".altmacro":
		f.altMacro = true
	case ".noaltmacro":
		f.altMacro = false
	}
}

func isMacroBodyText(s string) bool {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return false
	}
	head := fields[0]
	return !strings.HasPrefix(head, ".") && !strings.HasSuffix(head, ":") && !isPreProcessorInstruction(head)
}

func isDecorativeBlockCommentLine(s string) bool {
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "/*") {
		s = s[2:]
	}
	if strings.HasSuffix(s, "*/") {
		s = s[:len(s)-2]
	}
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return false
	}
	for _, r := range s {
		if r != '*' {
			return false
		}
	}
	return true
}

func trimBlockCommentLeader(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "*") {
		s = strings.TrimSpace(s[1:])
	}
	return s
}

func (f *fstate) emitBufferedBlockComment() {
	lines := append([]string(nil), f.blockBuffer...)
	f.blockBuffer = nil
	f.blockBuffered = false
	f.insideBlock = false

	canonical := len(lines) >= 2 && strings.TrimSpace(lines[0]) == "/*"
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, `\`) {
			canonical = false
			break
		}
	}

	if !canonical {
		prefix := f.indentString(f.blockIndentation)
		for _, line := range lines {
			fmt.Fprintln(f.out, strings.TrimPrefix(line, prefix))
		}
		f.blockStandalone = false
		f.blockCanonical = false
		return
	}

	f.out.WriteString("/*\n")
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "*/") {
			f.out.WriteString(" */\n")
			break
		}
		body := strings.TrimSpace(trimBlockCommentLeader(trimmed))
		if body == "" {
			f.out.WriteString(" *\n")
		} else {
			f.out.WriteString(" * " + body + "\n")
		}
	}
	f.blockStandalone = false
	f.blockCanonical = false
}

func isConservativeUnknownGasDirective(s string) bool {
	fields := strings.Fields(s)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], ".") {
		return false
	}
	return !isKnownGasDirectiveName(fields[0])
}

// Add a newline, unless last line was empty or a comment
func (f *fstate) newLine(enabled bool) {
	if !enabled || f.opts.maxBlankLines == 0 {
		return
	}
	// Always newline before comment-only line.
	if !f.lastEmpty && !f.lastComment && !f.lastLabel && f.anyContents {
		f.out.WriteByte('\n')
		f.blankLines = 1
		f.lastEmpty = true
	}
}

func (f *fstate) preferredCommentMark(mark string) string {
	if f.opts.preferredCommentStyle == "slash" {
		return "//"
	}
	return mark
}

func (f *fstate) commentSeparator() string {
	if f.opts.lineCommentSpace {
		return " "
	}
	return ""
}

// newStatement will parse a line and return it as a statement.
// Will return nil if the line is empty after whitespace removal.
func newStatement(s string, defs map[string]struct{}) *statement {
	return newStatementWithStyle(s, defs, styleUnknown)
}

func newStatementWithStyle(s string, defs map[string]struct{}, style sourceStyle) *statement {
	s = strings.TrimSpace(s)
	st := statement{}

	// Split into fields
	fields := strings.Fields(s)
	if len(fields) < 1 {
		return nil
	}

	// Fix where a comment start if any.
	// We need to make sure that the comment isn't embedded in a string literal.
	if !isPreProcessorInstruction(fields[0]) {
		if startcom, mark := findLineComment(s, style); startcom > 0 {
			st.comment = strings.TrimSpace(s[startcom+len(mark):])
			st.commentMark = mark
			s = strings.TrimSpace(s[:startcom])
			fields = strings.Fields(s)
			if len(fields) < 1 {
				return nil
			}
		}
	}
	st.instruction = fields[0]

	// Handle defined macro calls
	if len(defs) > 0 {
		inst := strings.Split(st.instruction, "(")[0]
		if _, ok := defs[inst]; ok {
			st.function = true
		}
	}
	if strings.HasPrefix(s, "/*") {
		st.function = true
	}
	// We may not have it defined as a macro, if defined in an external
	// .h file, so we try to detect the remaining ones.
	if strings.ContainsAny(st.instruction, "(_") {
		st.function = true
	}
	if len(st.params) > 0 && strings.HasPrefix(st.params[0], "(") {
		st.function = true
	}
	if st.function {
		st.instruction = s
	}

	if st.instruction == "\\" && len(st.comment) > 0 {
		st.instruction = fmt.Sprintf("\\ // %s", st.comment)
		st.comment = ""
		st.function = true
		st.continued = true
		st.contComment = true
	}

	s = strings.TrimPrefix(s, st.instruction)
	st.instruction = strings.Replace(st.instruction, "\t", " ", -1)
	s = strings.TrimSpace(s)

	st.setParams(s)

	// Remove trailing ;
	if len(st.params) > 0 {
		st.params[len(st.params)-1] = strings.TrimSuffix(st.params[len(st.params)-1], ";")
	} else {
		st.instruction = strings.TrimSuffix(st.instruction, ";")
	}

	// Register line continuations.
	if len(st.params) > 0 {
		p := st.params[len(st.params)-1]
		if st.willContinue() {
			p = strings.TrimSuffix(st.params[len(st.params)-1], `\`)
			p = strings.TrimSpace(p)
			if len(p) > 0 {
				st.params[len(st.params)-1] = p
			} else {
				st.params = st.params[:len(st.params)-1]
			}
			st.continued = true
		}
	}
	if strings.HasSuffix(st.instruction, `\`) && !st.contComment {
		i := strings.TrimSuffix(st.instruction, `\`)
		st.instruction = strings.TrimSpace(i)
		st.continued = true
	}

	if len(st.params) == 0 && !st.isLabel() {
		st.function = true
	}

	return &st
}

// setParams will add the string given as parameters.
// Inline comments are retained.
// There will be a space after ",", unless inside a comment.
// A tab is replaced by a space for consistent indentation.
func (st *statement) setParams(s string) {
	st.params = splitParams(s, st.usesGasParams(), st.isPreProcessor())
	if st.instruction == ".macro" {
		st.params = mergeMacroVarargParams(st.params)
	}
}

func splitParams(s string, trackDepth, preserveTabs bool) []string {
	params := make([]string, 0)
	runes := []rune(s)
	last := rune(0)
	inComment := false
	inStringLiteral := false
	inCharLiteral := false
	depth := 0
	out := make([]rune, 0, len(runes))
	for _, r := range runes {
		switch r {
		case '"':
			if last != '\\' && inStringLiteral {
				inStringLiteral = false
			} else if last != '\\' && !inStringLiteral {
				inStringLiteral = true
			}
		case '\'':
			if last != '\\' && inCharLiteral {
				inCharLiteral = false
			} else if last != '\\' && !inCharLiteral {
				inCharLiteral = true
			}
		case ',':
			if inComment || inStringLiteral || inCharLiteral || (trackDepth && depth > 0) {
				break
			}
			c := strings.TrimSpace(string(out))
			if len(c) > 0 {
				params = append(params, c)
			}
			out = out[0:0]
			continue
		case '/':
			if last == '*' && inComment {
				inComment = false
			}
		case '*':
			if last == '/' {
				inComment = true
			}
		case '(', '[', '{':
			if trackDepth && !inComment && !inStringLiteral && !inCharLiteral {
				depth++
			}
		case ')', ']', '}':
			if trackDepth && !inComment && !inStringLiteral && !inCharLiteral && depth > 0 {
				depth--
			}
		case '\t':
			if !preserveTabs {
				r = ' '
			}
		case ';':
			if inComment || inStringLiteral || inCharLiteral {
				break
			}
			out = []rune(strings.TrimSpace(string(out)) + "; ")
			last = r
			continue
		}
		if last == ';' && unicode.IsSpace(r) {
			continue
		}
		last = r
		out = append(out, r)
	}
	c := strings.TrimSpace(string(out))
	if len(c) > 0 {
		params = append(params, c)
	}
	return params
}

func mergeMacroVarargParams(params []string) []string {
	for i, p := range params {
		if strings.Contains(p, ":vararg") {
			if i == len(params)-1 {
				return params
			}
			merged := append([]string{}, params[:i]...)
			merged = append(merged, strings.Join(params[i:], ", "))
			return merged
		}
	}
	return params
}

// Return true if this line should be at indentation level 0.
func (st statement) level0() bool {
	return st.isLabel() || st.isTEXT() || st.isPreProcessor() || st.isGasDirective()
}

// Will return true if the statement is a label.
func (st statement) isLabel() bool {
	return strings.HasSuffix(st.instruction, ":")
}

func (st statement) isNumericLocalLabel() bool {
	if !st.isLabel() {
		return false
	}
	name := strings.TrimSuffix(st.instruction, ":")
	if name == "" {
		return false
	}
	for _, r := range name {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isPreProcessor will return if the statement is a preprocessor statement.
func (st statement) isPreProcessor() bool {
	return isPreProcessorInstruction(st.instruction)
}

// isGlobal returns true if the current instruction is
// a global. Currently that is DATA, GLOBL, FUNCDATA and PCDATA
func (st statement) isGlobal() bool {
	up := strings.ToUpper(st.instruction)
	switch up {
	case "DATA", "GLOBL", "FUNCDATA", "PCDATA":
		return true
	default:
		return false
	}
}

// isTEXT returns true if the instruction is "TEXT"
// or one of the "isGlobal" types
func (st statement) isTEXT() bool {
	up := strings.ToUpper(st.instruction)
	return up == "TEXT" || st.isGlobal()
}

// We attempt to identify "terminators", after which
// indentation is likely to be level 0.
func (st statement) isTerminator() bool {
	up := strings.ToUpper(st.instruction)
	switch up {
	case "RET", "JMP", "J", "JR":
		return true
	}
	low := strings.ToLower(st.instruction)
	if low != st.instruction {
		return false
	}
	switch low {
	case "j", "jr", "ret", "tail", "ebreak", "ecall", "mret", "sret", "uret":
		return true
	default:
		return false
	}
}

// Detects commands based on case.
func (st statement) isCommand() bool {
	if st.isLabel() || st.isPreProcessor() || st.isGasDirective() {
		return false
	}
	up := strings.ToUpper(st.instruction)
	return up == st.instruction
}

func (st statement) isGasDirective() bool {
	return strings.HasPrefix(st.instruction, ".") && !st.isLabel()
}

func (st statement) usesGasParams() bool {
	if st.isGasDirective() {
		return true
	}
	r, _ := utf8.DecodeRuneInString(st.instruction)
	return unicode.IsLower(r)
}

func (st statement) isGasBlockStart() bool {
	switch st.instruction {
	case ".macro", ".irp", ".irpc", ".rept", ".if", ".ifdef", ".ifndef", ".ifnotdef",
		".ifb", ".ifnb", ".ifc", ".ifnc", ".ifeq", ".ifne", ".ifge", ".ifgt", ".ifle", ".iflt",
		".struct":
		return true
	default:
		return false
	}
}

func (st statement) isGasBlockMiddle() bool {
	switch st.instruction {
	case ".elseif", ".else":
		return true
	default:
		return false
	}
}

func (st statement) isGasBlockEnd() bool {
	switch st.instruction {
	case ".endm", ".endr", ".endif", ".endstruct":
		return true
	default:
		return false
	}
}

func (st statement) isGasZeroDirective() bool {
	if !st.isGasDirective() {
		return false
	}
	switch st.instruction {
	case ".text", ".data", ".rodata", ".bss", ".section", ".pushsection", ".popsection",
		".previous", ".subsection", ".globl", ".global", ".local", ".weak", ".comm",
		".extern",
		".common", ".file", ".ident", ".size", ".type", ".attribute", ".option",
		".altmacro", ".noaltmacro",
		".align", ".p2align", ".balign", ".equ", ".set", ".byte", ".2byte", ".half",
		".short", ".4byte", ".word", ".long", ".8byte", ".dword", ".quad", ".float",
		".double", ".string", ".asciz", ".zero",
		".org", ".incbin", ".sleb128", ".uleb128", ".variant_cc", ".reloc", ".loc",
		".line", ".app-file", ".hidden", ".protected", ".internal", ".symver", ".weakref",
		".equiv", ".eqv", ".offset", ".include", ".err", ".error", ".warning", ".title",
		".sbttl", ".cfi_startproc", ".cfi_endproc", ".cfi_def_cfa", ".cfi_def_cfa_register",
		".cfi_def_cfa_offset", ".cfi_offset", ".cfi_restore", ".cfi_adjust_cfa_offset",
		".cfi_remember_state", ".cfi_restore_state", ".cfi_sections", ".cfi_signal_frame":
		return true
	default:
		return false
	}
}

func (st statement) isKnownGasDirective() bool {
	return isKnownGasDirectiveName(st.instruction)
}

func isKnownGasDirectiveName(name string) bool {
	st := statement{instruction: name}
	return st.isGasZeroDirective() || st.isGasBlockStart() || st.isGasBlockMiddle() || st.isGasBlockEnd() ||
		st.isGasInstructionStreamDirective() || st.isGasDataDirective()
}

func (st statement) isGasInstructionStreamDirective() bool {
	switch st.instruction {
	case ".insn":
		return true
	default:
		return false
	}
}

func (st statement) isGasDataDirective() bool {
	switch st.instruction {
	case ".ascii", ".space", ".skip", ".fill":
		return true
	default:
		return false
	}
}

// Detect if last character is '\', indicating a multiline statement.
func (st statement) willContinue() bool {
	if st.continued {
		return true
	}
	if len(st.params) == 0 {
		return false
	}
	return strings.HasSuffix(st.params[len(st.params)-1], `\`)
}

// define returns the macro defined in this line.
// if none is defined "" is returned.
func (st statement) define() string {
	if st.instruction == "#define" && len(st.params) > 0 {
		r := strings.TrimSpace(strings.Split(st.params[0], "(")[0])
		r = strings.Trim(r, `\`)
		return r
	}
	return ""
}

func (st *statement) cleanParams() {
	// Remove whitespace before semicolons
	if strings.HasSuffix(st.instruction, ";") {
		s := strings.TrimSuffix(st.instruction, ";")
		st.instruction = strings.TrimSpace(s) + ";"
	}
}

// formatStatements will format a slice of statements and return each line
// as a separate string.
// Comments and line-continuation (\) are aligned with spaces.
func formatStatements(s []statement, opts normalizedOptions) []string {
	if opts.sourceStyle == styleGas || opts.sourceStyle == styleRiscvGas {
		return formatStatementsCustom(s, opts)
	}
	if opts.alignOperands && opts.alignComments && opts.alignContinuations && opts.lineCommentSpace && opts.preferredCommentStyle == "preserve" {
		return formatStatementsDefault(s)
	}
	return formatStatementsCustom(s, opts)
}

func formatStatementsDefault(s []statement) []string {
	res := make([]string, len(s))
	maxParam := 0 // Length of longest parameter
	maxInstr := 0 // Length of longest instruction WITH parameters.
	maxAlone := 0 // Length of longest instruction without parameters.
	for i, x := range s {
		// Clean up and store
		x.cleanParams()
		s[i] = x

		il := len([]rune(x.instruction)) + 1 // Instruction length
		l := il
		// Ignore length if we are a define "function"
		// or we are a parameterless instruction.
		if l > maxInstr && !x.function && !(x.isCommand() && len(x.params) == 0) {
			maxInstr = l
		}
		if x.function && il > maxAlone {
			maxAlone = il
		}
		if len(x.params) > 1 {
			l = 2 * (len(x.params) - 1) // Spaces between parameters
		} else {
			l = 0
		}
		// Add parameters
		for _, y := range x.params {
			l += len([]rune(y))
		}
		l++
		if l > maxParam {
			maxParam = l
		}
	}

	maxParam += maxInstr
	if maxInstr == 0 {
		maxInstr = maxAlone
	}

	for i, x := range s {
		if x.contComment {
			res[i] = x.instruction
			continue
		}
		r := x.instruction
		p := strings.Join(x.params, ", ")
		if len(x.params) > 0 || len(x.comment) > 0 {
			for len(r) < maxInstr {
				r += " "
			}
		}
		r = r + p
		if len(x.comment) > 0 && !x.continued {
			it := maxParam - len([]rune(r))
			for i := 0; i < it; i++ {
				r += " "
			}
			r += fmt.Sprintf("%s %s", x.defaultLineCommentMark(), x.comment)
		}

		if x.continued {
			// Find continuation placement.
			it := maxParam - len([]rune(r))
			if maxAlone > maxParam {
				it = maxAlone - len([]rune(r))
			}
			for i := 0; i < it; i++ {
				r += " "
			}
			r += `\`
			// Add comment, if any.
			if len(x.comment) > 0 {
				r += " " + x.defaultLineCommentMark() + " " + x.comment
			}
		}
		res[i] = r
	}
	return res
}

func formatStatementsCustom(s []statement, opts normalizedOptions) []string {
	res := make([]string, len(s))
	maxInstr := 0
	for i, x := range s {
		x.cleanParams()
		s[i] = x

		il := len([]rune(x.instruction)) + 1
		if il > maxInstr && !x.function && !(x.isCommand() && len(x.params) == 0) {
			maxInstr = il
		}
	}

	maxCodeLen := 0
	for _, x := range s {
		if x.contComment {
			continue
		}
		if len(x.comment) > 0 || x.continued {
			l := len([]rune(renderStatementCode(x, maxInstr, opts)))
			if l > maxCodeLen {
				maxCodeLen = l
			}
		}
	}

	for i, x := range s {
		if x.contComment {
			res[i] = x.instruction
			continue
		}
		r := renderStatementCode(x, maxInstr, opts)
		if len(x.comment) > 0 && !x.continued {
			if opts.alignComments {
				it := maxCodeLen - len([]rune(r)) + 1
				for i := 0; i < it; i++ {
					r += " "
				}
			} else {
				r += " "
			}
			r += x.lineCommentMark(opts) + commentTextSeparator(opts, x.comment) + x.comment
		}

		if x.continued {
			if opts.alignContinuations {
				it := maxCodeLen - len([]rune(r)) + 1
				for i := 0; i < it; i++ {
					r += " "
				}
			} else {
				r += " "
			}
			r += `\`
			if len(x.comment) > 0 {
				r += " " + x.lineCommentMark(opts) + commentTextSeparator(opts, x.comment) + x.comment
			}
		}
		res[i] = r
	}
	return res
}

func renderStatementCode(st statement, maxInstr int, opts normalizedOptions) string {
	r := st.instruction
	if opts.alignOperands {
		if len(st.params) > 0 || len(st.comment) > 0 {
			for len([]rune(r)) < maxInstr {
				r += " "
			}
		}
	} else if len(st.params) > 0 {
		r += " "
	}
	return r + strings.Join(st.params, ", ")
}

func commentTextSeparator(opts normalizedOptions, text string) string {
	if !opts.lineCommentSpace || text == "" {
		return ""
	}
	return " "
}

func (st statement) defaultLineCommentMark() string {
	if st.commentMark != "" {
		return st.commentMark
	}
	return "//"
}

func (st statement) lineCommentMark(opts normalizedOptions) string {
	if opts.preferredCommentStyle == "slash" {
		return "//"
	}
	return st.defaultLineCommentMark()
}

func findLineComment(s string, style sourceStyle) (int, string) {
	inStringLiteral := false
	inCharLiteral := false
	inBlockComment := false
	last := rune(0)
	for i, r := range s {
		switch r {
		case '"':
			if !inCharLiteral && last != '\\' {
				inStringLiteral = !inStringLiteral
			}
		case '\'':
			if !inStringLiteral && last != '\\' {
				inCharLiteral = !inCharLiteral
			}
		case '/':
			if !inStringLiteral && !inCharLiteral && i+1 < len(s) && s[i+1] == '/' {
				return i, "//"
			}
			if !inStringLiteral && !inCharLiteral && inBlockComment && last == '*' {
				inBlockComment = false
			}
		case '*':
			if !inStringLiteral && !inCharLiteral && last == '/' {
				inBlockComment = true
			}
		case '#':
			if !inStringLiteral && !inCharLiteral && !inBlockComment && hashStartsComment(s, i, style) {
				return i, "#"
			}
		case '@':
			if !inStringLiteral && !inCharLiteral && !inBlockComment && atStartsComment(s, i, style) {
				return i, "@"
			}
		}
		last = r
	}
	return -1, ""
}

func isHashLineComment(s string, i int) bool {
	if i == 0 {
		return true
	}
	if !unicode.IsSpace(rune(s[i-1])) {
		return false
	}
	return i+1 == len(s) || unicode.IsSpace(rune(s[i+1]))
}

func findBlockCommentStart(s string) int {
	inStringLiteral := false
	inCharLiteral := false
	last := rune(0)
	for i, r := range s {
		switch r {
		case '"':
			if !inCharLiteral && last != '\\' {
				inStringLiteral = !inStringLiteral
			}
		case '\'':
			if !inStringLiteral && last != '\\' {
				inCharLiteral = !inCharLiteral
			}
		case '*':
			if !inStringLiteral && !inCharLiteral && last == '/' {
				return i - 1
			}
		}
		last = r
	}
	return -1
}

func splitStatements(s string, style sourceStyle) []string {
	var parts []string
	inStringLiteral := false
	inCharLiteral := false
	inBlockComment := false
	last := rune(0)
	start := 0
	for i, r := range s {
		switch r {
		case '"':
			if !inCharLiteral && last != '\\' {
				inStringLiteral = !inStringLiteral
			}
		case '\'':
			if !inStringLiteral && last != '\\' {
				inCharLiteral = !inCharLiteral
			}
		case '/':
			if !inStringLiteral && !inCharLiteral && i+1 < len(s) && s[i+1] == '/' {
				return appendSemicolonPart(parts, s[start:])
			}
			if !inStringLiteral && !inCharLiteral && inBlockComment && last == '*' {
				inBlockComment = false
			}
		case '*':
			if !inStringLiteral && !inCharLiteral && last == '/' {
				inBlockComment = true
			}
		case '#':
			if !inStringLiteral && !inCharLiteral && !inBlockComment && hashStartsComment(s, i, style) {
				return appendSemicolonPart(parts, s[start:])
			}
		case '@':
			if !inStringLiteral && !inCharLiteral && !inBlockComment && atStartsComment(s, i, style) {
				return appendSemicolonPart(parts, s[start:])
			}
		case ';':
			if !inStringLiteral && !inCharLiteral && !inBlockComment {
				parts = appendSemicolonPart(parts, s[start:i])
				start = i + 1
			}
		}
		last = r
	}
	parts = appendSemicolonPart(parts, s[start:])
	if len(parts) <= 1 {
		return nil
	}
	return parts
}

func shouldSplitSemicolonStatements(s string) bool {
	return shouldSplitSemicolonStatementsForStyle(s, styleUnknown, true)
}

func appendSemicolonPart(parts []string, s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return parts
	}
	return append(parts, s)
}

func isCommentedPreProcessor(s string) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "#") {
		return false
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return false
	}
	return isPreProcessorInstruction(fields[0])
}

func isPreProcessorInstruction(s string) bool {
	switch s {
	case "#define", "#include", "#if", "#ifdef", "#ifndef", "#else", "#elif", "#endif", "#undef", "#error", "#warning", "#pragma":
		return true
	default:
		return false
	}
}

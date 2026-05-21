package asmfmt

import (
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokenIdentifier tokenKind = iota
	tokenDirective
	tokenLabel
	tokenNumber
	tokenString
	tokenChar
	tokenComment
	tokenSeparator
	tokenOperator
	tokenRawText
)

type lexicalMode int

const (
	modeNormal lexicalMode = iota
	modeStringLiteral
	modeCharLiteral
	modeBlockComment
	modeLineComment
	modePreprocessorLine
)

type token struct {
	kind tokenKind
	mode lexicalMode
	text string
	pos  int
}

type lexer struct {
	src   string
	pos   int
	mode  lexicalMode
	atBOL bool
}

func lexLine(s string) []token {
	l := lexer{src: s, atBOL: true}
	var tokens []token
	for {
		tok, ok := l.next()
		if !ok {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func (l *lexer) next() (token, bool) {
	for l.pos < len(l.src) {
		start := l.pos
		switch l.mode {
		case modeStringLiteral:
			return l.scanQuoted(start, '"', tokenString, modeStringLiteral), true
		case modeCharLiteral:
			return l.scanQuoted(start, '\'', tokenChar, modeCharLiteral), true
		case modeBlockComment:
			return l.scanBlockComment(start), true
		case modeLineComment:
			tok := token{kind: tokenComment, mode: modeLineComment, text: l.src[start:], pos: start}
			l.pos = len(l.src)
			return tok, true
		}

		r := l.src[l.pos]
		if unicode.IsSpace(rune(r)) {
			l.pos++
			if l.atBOL {
				continue
			}
			continue
		}
		if l.atBOL && r == '#' && isPreProcessorLine(l.src[start:]) {
			l.mode = modePreprocessorLine
		}
		l.atBOL = false

		if r == '"' {
			l.mode = modeStringLiteral
			return l.scanQuoted(start, '"', tokenString, modeStringLiteral), true
		}
		if r == '\'' {
			l.mode = modeCharLiteral
			return l.scanQuoted(start, '\'', tokenChar, modeCharLiteral), true
		}
		if r == '/' && l.peek('/') {
			l.mode = modeLineComment
			return l.next()
		}
		if r == '/' && l.peek('*') {
			l.mode = modeBlockComment
			return l.scanBlockComment(start), true
		}
		if r == '#' && isHashLineComment(l.src, l.pos) && l.mode != modePreprocessorLine {
			l.mode = modeLineComment
			return l.next()
		}
		if isSeparator(r) {
			l.pos++
			return token{kind: tokenSeparator, mode: l.mode, text: l.src[start:l.pos], pos: start}, true
		}
		if isOperator(r) {
			l.pos++
			return token{kind: tokenOperator, mode: l.mode, text: l.src[start:l.pos], pos: start}, true
		}
		if isNumberStart(l.src, l.pos) {
			l.pos++
			for l.pos < len(l.src) && isNumberPart(l.src[l.pos]) {
				l.pos++
			}
			return token{kind: tokenNumber, mode: l.mode, text: l.src[start:l.pos], pos: start}, true
		}
		if isIdentifierStart(r) || r == '.' || r == '\\' || r == '%' {
			l.pos++
			for l.pos < len(l.src) && isIdentifierPart(l.src[l.pos]) {
				l.pos++
			}
			kind := tokenIdentifier
			if strings.HasPrefix(l.src[start:l.pos], ".") {
				kind = tokenDirective
			}
			if l.pos < len(l.src) && l.src[l.pos] == ':' {
				l.pos++
				kind = tokenLabel
			}
			return token{kind: kind, mode: l.mode, text: l.src[start:l.pos], pos: start}, true
		}
		l.pos++
		return token{kind: tokenRawText, mode: l.mode, text: l.src[start:l.pos], pos: start}, true
	}
	return token{}, false
}

func (l *lexer) scanQuoted(start int, quote byte, kind tokenKind, mode lexicalMode) token {
	l.pos++
	for l.pos < len(l.src) {
		if l.src[l.pos] == quote && !isEscaped(l.src, l.pos) {
			l.pos++
			break
		}
		l.pos++
	}
	l.mode = modeNormal
	return token{kind: kind, mode: mode, text: l.src[start:l.pos], pos: start}
}

func (l *lexer) scanBlockComment(start int) token {
	l.pos += 2
	for l.pos < len(l.src) {
		if l.src[l.pos-1] == '*' && l.src[l.pos] == '/' {
			l.pos++
			l.mode = modeNormal
			return token{kind: tokenComment, mode: modeBlockComment, text: l.src[start:l.pos], pos: start}
		}
		l.pos++
	}
	return token{kind: tokenComment, mode: modeBlockComment, text: l.src[start:l.pos], pos: start}
}

func (l *lexer) peek(next byte) bool {
	return l.pos+1 < len(l.src) && l.src[l.pos+1] == next
}

func isEscaped(s string, pos int) bool {
	escapes := 0
	for i := pos - 1; i >= 0 && s[i] == '\\'; i-- {
		escapes++
	}
	return escapes%2 == 1
}

func isPreProcessorLine(s string) bool {
	fields := strings.Fields(s)
	return len(fields) > 0 && isPreProcessorInstruction(fields[0])
}

func isSeparator(b byte) bool {
	switch b {
	case ',', ';', '(', ')', '[', ']', '{', '}':
		return true
	default:
		return false
	}
}

func isOperator(b byte) bool {
	switch b {
	case '#', '@', '&', '+', '-', '*', '/', '=', ':':
		return true
	default:
		return false
	}
}

func isIdentifierStart(b byte) bool {
	return b == '_' || unicode.IsLetter(rune(b))
}

func isIdentifierPart(b byte) bool {
	return isIdentifierStart(b) || unicode.IsDigit(rune(b)) || strings.ContainsRune(".$", rune(b))
}

func isNumberStart(s string, pos int) bool {
	if pos >= len(s) {
		return false
	}
	if unicode.IsDigit(rune(s[pos])) {
		return true
	}
	if (s[pos] == '+' || s[pos] == '-') && pos+1 < len(s) && unicode.IsDigit(rune(s[pos+1])) {
		return true
	}
	return false
}

func isNumberPart(b byte) bool {
	return unicode.IsDigit(rune(b)) || unicode.IsLetter(rune(b)) || b == 'x' || b == 'X'
}

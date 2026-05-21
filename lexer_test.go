package asmfmt

import "testing"

func TestLexLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []token
	}{
		{
			name: "string hides comment markers",
			line: `.ascii "# not comment" # comment`,
			want: []token{
				{kind: tokenDirective, mode: modeNormal, text: `.ascii`, pos: 0},
				{kind: tokenString, mode: modeStringLiteral, text: `"# not comment"`, pos: 7},
				{kind: tokenComment, mode: modeLineComment, text: `# comment`, pos: 23},
			},
		},
		{
			name: "escaped quote in string",
			line: `.ascii "say \"hi\""`,
			want: []token{
				{kind: tokenDirective, mode: modeNormal, text: `.ascii`, pos: 0},
				{kind: tokenString, mode: modeStringLiteral, text: `"say \"hi\""`, pos: 7},
			},
		},
		{
			name: "character constant",
			line: `.byte '\n' # newline`,
			want: []token{
				{kind: tokenDirective, mode: modeNormal, text: `.byte`, pos: 0},
				{kind: tokenChar, mode: modeCharLiteral, text: `'\n'`, pos: 6},
				{kind: tokenComment, mode: modeLineComment, text: `# newline`, pos: 11},
			},
		},
		{
			name: "line continuation",
			line: `#define X a \\`,
			want: []token{
				{kind: tokenOperator, mode: modePreprocessorLine, text: `#`, pos: 0},
				{kind: tokenIdentifier, mode: modePreprocessorLine, text: `define`, pos: 1},
				{kind: tokenIdentifier, mode: modePreprocessorLine, text: `X`, pos: 8},
				{kind: tokenIdentifier, mode: modePreprocessorLine, text: `a`, pos: 10},
				{kind: tokenIdentifier, mode: modePreprocessorLine, text: `\`, pos: 12},
				{kind: tokenIdentifier, mode: modePreprocessorLine, text: `\`, pos: 13},
			},
		},
		{
			name: "# immediate is not comment",
			line: `add r0, r1, #1`,
			want: []token{
				{kind: tokenIdentifier, mode: modeNormal, text: `add`, pos: 0},
				{kind: tokenIdentifier, mode: modeNormal, text: `r0`, pos: 4},
				{kind: tokenSeparator, mode: modeNormal, text: `,`, pos: 6},
				{kind: tokenIdentifier, mode: modeNormal, text: `r1`, pos: 8},
				{kind: tokenSeparator, mode: modeNormal, text: `,`, pos: 10},
				{kind: tokenOperator, mode: modeNormal, text: `#`, pos: 12},
				{kind: tokenNumber, mode: modeNormal, text: `1`, pos: 13},
			},
		},
		{
			name: "@ marker stays operand text",
			line: `.type foo, @function`,
			want: []token{
				{kind: tokenDirective, mode: modeNormal, text: `.type`, pos: 0},
				{kind: tokenIdentifier, mode: modeNormal, text: `foo`, pos: 6},
				{kind: tokenSeparator, mode: modeNormal, text: `,`, pos: 9},
				{kind: tokenOperator, mode: modeNormal, text: `@`, pos: 11},
				{kind: tokenIdentifier, mode: modeNormal, text: `function`, pos: 12},
			},
		},
		{
			name: "semicolon splits but string survives",
			line: `addi a0, a0, 1; .ascii ";"`,
			want: []token{
				{kind: tokenIdentifier, mode: modeNormal, text: `addi`, pos: 0},
				{kind: tokenIdentifier, mode: modeNormal, text: `a0`, pos: 5},
				{kind: tokenSeparator, mode: modeNormal, text: `,`, pos: 7},
				{kind: tokenIdentifier, mode: modeNormal, text: `a0`, pos: 9},
				{kind: tokenSeparator, mode: modeNormal, text: `,`, pos: 11},
				{kind: tokenNumber, mode: modeNormal, text: `1`, pos: 13},
				{kind: tokenSeparator, mode: modeNormal, text: `;`, pos: 14},
				{kind: tokenDirective, mode: modeNormal, text: `.ascii`, pos: 16},
				{kind: tokenString, mode: modeStringLiteral, text: `";"`, pos: 23},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lexLine(tt.line)
			if len(got) != len(tt.want) {
				t.Fatalf("len(tokens) = %d; want %d (%#v)", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("token[%d] = %#v; want %#v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

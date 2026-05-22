package asmfmt

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Options configures asmfmt formatting behavior.
type Options struct {
	IndentStyle                   string      `toml:"indent_style"`
	IndentWidth                   int         `toml:"indent_width"`
	AlignOperands                 bool        `toml:"align_operands"`
	AlignComments                 bool        `toml:"align_comments"`
	AlignContinuations            bool        `toml:"align_continuations"`
	MaxBlankLines                 int         `toml:"max_blank_lines"`
	SplitSemicolonStatements      bool        `toml:"split_semicolon_statements"`
	NewlineBeforeComments         bool        `toml:"newline_before_comments"`
	NewlineBeforeLabels           bool        `toml:"newline_before_labels"`
	LabelsAlwaysOnOwnLine         bool        `toml:"labels_always_on_own_line"`
	LineCommentSpace              bool        `toml:"line_comment_space"`
	ConvertSingleLineBlockComment bool        `toml:"convert_single_line_block_comment"`
	PreferredCommentStyle         string      `toml:"preferred_comment_style"`
	SourceStyle                   string      `toml:"source_style"`
	IndentGASDirectives           bool        `toml:"indent_gas_directives"`
	Lint                          LintOptions `toml:"lint"`
}

// LintOptions defines configuration for the assembler style linter.
type LintOptions struct {
	AbiRegisters                   string `toml:"abi_registers"`
	CompressedInstructions         string `toml:"compressed_instructions"`
	OperationImmediate             string `toml:"operation_immediate"`
	RelocationOperatorSpacing      string `toml:"relocation_operator_spacing"`
	GpLoadRelaxation               string `toml:"gp_load_relaxation"`
	CsrNames                       string `toml:"csr_names"`
	JumpInstructionSelection       string `toml:"jump_instruction_selection"`
	PcrelRelocationLabel           string `toml:"pcrel_relocation_label"`
	AlignmentDirectives            string `toml:"alignment_directives"`
	ExternDirective                string `toml:"extern_directive"`
	InlineBinaryDirectives         string `toml:"inline_binary_directives"`
	AvoidGlobl                     string `toml:"avoid_globl"`
	Leb128ConstantExpression       string `toml:"leb128_constant_expression"`
	AvoidSpaceSkipDirectives       string `toml:"avoid_space_skip_directives"`
	OperatorPrecedenceParentheses  string `toml:"operator_precedence_parentheses"`
	EndDirectiveLast               string `toml:"end_directive_last"`
	LocalLabels                    string `toml:"local_labels"`
	CurrentPointLabel              string `toml:"current_point_label"`
	PointerOffsetShorthand         string `toml:"pointer_offset_shorthand"`
	OptionPushPop                  string `toml:"option_push_pop"`
	SymbolPreambleFooter           string `toml:"symbol_preamble_footer"`
	CfiStartEndBalance             string `toml:"cfi_start_end_balance"`
	MacroBalance                   string `toml:"macro_balance"`
	InstructionSequenceTermination string `toml:"instruction_sequence_termination"`
	FunctionDoxygenComment         string `toml:"function_doxygen_comment"`
	AvoidHashAndAtComments         string `toml:"avoid_hash_and_at_comments"`
	DoubleLabelDeclaration         string `toml:"double_label_declaration"`
	ReservedLabelNames             string `toml:"reserved_label_names"`
	UnreachableCode                string `toml:"unreachable_code"`
	SingleReturnStatement          string `toml:"single_return_statement"`
	NoFallthroughToFunction        string `toml:"no_fallthrough_to_function"`
	NoJumpOutOfFunction            string `toml:"no_jump_out_of_function"`
	NoRecursiveCalls               string `toml:"no_recursive_calls"`
	CopyrightAndLicense            string `toml:"copyright_and_license"`
	DeveloperNameLength            string `toml:"developer_name_length"`
	LineLengthLimit                string `toml:"line_length_limit"`

	// Casing style parameters
	LabelNamingStyle string `toml:"label_naming_style"`
	MacroNamingStyle string `toml:"macro_naming_style"`
}

type normalizedOptions struct {
	indentStyle                   string
	indentWidth                   int
	alignOperands                 bool
	alignComments                 bool
	alignContinuations            bool
	maxBlankLines                 int
	splitSemicolonStatements      bool
	newlineBeforeComments         bool
	newlineBeforeLabels           bool
	labelsAlwaysOnOwnLine         bool
	lineCommentSpace              bool
	convertSingleLineBlockComment bool
	preferredCommentStyle         string
	sourceStyle                   sourceStyle
	indentGASDirectives           bool
	lint                          normalizedLintOptions
}

type normalizedLintOptions struct {
	severities       map[string]string
	labelNamingStyle string
	macroNamingStyle string
}

// DefaultLintOptions returns the default configurations for the linter.
func DefaultLintOptions() LintOptions {
	return LintOptions{
		AbiRegisters:                   "error",
		CompressedInstructions:         "error",
		OperationImmediate:             "error",
		RelocationOperatorSpacing:      "error",
		GpLoadRelaxation:               "error",
		CsrNames:                       "error",
		JumpInstructionSelection:       "warning",
		PcrelRelocationLabel:           "error",
		AlignmentDirectives:            "error",
		ExternDirective:                "error",
		InlineBinaryDirectives:         "warning",
		AvoidGlobl:                     "error",
		Leb128ConstantExpression:       "error",
		AvoidSpaceSkipDirectives:       "error",
		OperatorPrecedenceParentheses:  "warning",
		EndDirectiveLast:               "error",
		LocalLabels:                    "warning",
		CurrentPointLabel:              "error",
		PointerOffsetShorthand:         "warning",
		OptionPushPop:                  "error",
		SymbolPreambleFooter:           "warning",
		CfiStartEndBalance:             "warning",
		MacroBalance:                   "warning",
		InstructionSequenceTermination: "warning",
		FunctionDoxygenComment:         "warning",
		AvoidHashAndAtComments:         "warning",
		DoubleLabelDeclaration:         "error",
		ReservedLabelNames:             "error",
		UnreachableCode:                "error",
		SingleReturnStatement:          "warning",
		NoFallthroughToFunction:        "warning",
		NoJumpOutOfFunction:            "error",
		NoRecursiveCalls:               "error",
		CopyrightAndLicense:            "warning",
		DeveloperNameLength:            "error",
		LineLengthLimit:                "warning",
		LabelNamingStyle:               "snake_case",
		MacroNamingStyle:               "UPPER_SNAKE_CASE",
	}
}

// DefaultOptions returns the formatter defaults used when no config is loaded.
func DefaultOptions() Options {
	return Options{
		IndentStyle:                   "tab",
		IndentWidth:                   8,
		AlignOperands:                 true,
		AlignComments:                 true,
		AlignContinuations:            true,
		MaxBlankLines:                 1,
		SplitSemicolonStatements:      true,
		NewlineBeforeComments:         true,
		NewlineBeforeLabels:           true,
		LabelsAlwaysOnOwnLine:         true,
		LineCommentSpace:              true,
		ConvertSingleLineBlockComment: true,
		PreferredCommentStyle:         "preserve",
		SourceStyle:                   "auto",
		IndentGASDirectives:           false,
		Lint:                          DefaultLintOptions(),
	}
}

// ParseOptionsTOML decodes formatter options from TOML input.
func ParseOptionsTOML(data []byte) (Options, error) {
	opts := DefaultOptions()
	md, err := toml.Decode(string(data), &opts)
	if err != nil {
		return Options{}, err
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		names := make([]string, 0, len(undecoded))
		for _, key := range undecoded {
			names = append(names, key.String())
		}
		return Options{}, fmt.Errorf("unknown config field(s): %s", strings.Join(names, ", "))
	}
	if _, err := opts.normalize(); err != nil {
		return Options{}, err
	}
	return opts, nil
}

// LoadOptionsFile reads formatter options from a TOML file.
func LoadOptionsFile(path string) (Options, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Options{}, err
	}
	opts, err := ParseOptionsTOML(data)
	if err != nil {
		return Options{}, fmt.Errorf("%s: %w", path, err)
	}
	return opts, nil
}

func validateSeverity(val string, name string) error {
	if val == "" {
		return nil
	}
	switch val {
	case "error", "warning", "ignore":
		return nil
	default:
		return fmt.Errorf("invalid severity %q for rule %q (must be error, warning, or ignore)", val, name)
	}
}

func (o Options) normalize() (normalizedOptions, error) {
	n := normalizedOptions{
		indentStyle:                   o.IndentStyle,
		indentWidth:                   o.IndentWidth,
		alignOperands:                 o.AlignOperands,
		alignComments:                 o.AlignComments,
		alignContinuations:            o.AlignContinuations,
		maxBlankLines:                 o.MaxBlankLines,
		splitSemicolonStatements:      o.SplitSemicolonStatements,
		newlineBeforeComments:         o.NewlineBeforeComments,
		newlineBeforeLabels:           o.NewlineBeforeLabels,
		labelsAlwaysOnOwnLine:         o.LabelsAlwaysOnOwnLine,
		lineCommentSpace:              o.LineCommentSpace,
		convertSingleLineBlockComment: o.ConvertSingleLineBlockComment,
		preferredCommentStyle:         o.PreferredCommentStyle,
		indentGASDirectives:           o.IndentGASDirectives,
	}
	switch o.IndentStyle {
	case "tab", "space":
	default:
		return normalizedOptions{}, fmt.Errorf("invalid indent_style %q", o.IndentStyle)
	}
	if o.IndentWidth <= 0 {
		return normalizedOptions{}, fmt.Errorf("indent_width must be greater than 0")
	}
	if o.MaxBlankLines < 0 {
		return normalizedOptions{}, fmt.Errorf("max_blank_lines must be greater than or equal to 0")
	}
	switch o.PreferredCommentStyle {
	case "preserve", "slash":
	default:
		return normalizedOptions{}, fmt.Errorf("invalid preferred_comment_style %q", o.PreferredCommentStyle)
	}
	switch o.SourceStyle {
	case "auto":
		n.sourceStyle = styleUnknown
	case "plan9":
		n.sourceStyle = stylePlan9
	case "gas":
		n.sourceStyle = styleGas
	case "riscv-gas":
		n.sourceStyle = styleRiscvGas
	default:
		return normalizedOptions{}, fmt.Errorf("invalid source_style %q", o.SourceStyle)
	}

	// Validate and normalize lint options
	n.lint.severities = make(map[string]string)

	rules := map[string]string{
		"abi_registers":                    o.Lint.AbiRegisters,
		"compressed_instructions":          o.Lint.CompressedInstructions,
		"operation_immediate":              o.Lint.OperationImmediate,
		"relocation_operator_spacing":      o.Lint.RelocationOperatorSpacing,
		"gp_load_relaxation":               o.Lint.GpLoadRelaxation,
		"csr_names":                        o.Lint.CsrNames,
		"jump_instruction_selection":       o.Lint.JumpInstructionSelection,
		"pcrel_relocation_label":           o.Lint.PcrelRelocationLabel,
		"alignment_directives":             o.Lint.AlignmentDirectives,
		"extern_directive":                 o.Lint.ExternDirective,
		"inline_binary_directives":         o.Lint.InlineBinaryDirectives,
		"avoid_globl":                      o.Lint.AvoidGlobl,
		"leb128_constant_expression":       o.Lint.Leb128ConstantExpression,
		"avoid_space_skip_directives":      o.Lint.AvoidSpaceSkipDirectives,
		"operator_precedence_parentheses":  o.Lint.OperatorPrecedenceParentheses,
		"end_directive_last":               o.Lint.EndDirectiveLast,
		"local_labels":                     o.Lint.LocalLabels,
		"current_point_label":              o.Lint.CurrentPointLabel,
		"pointer_offset_shorthand":         o.Lint.PointerOffsetShorthand,
		"option_push_pop":                  o.Lint.OptionPushPop,
		"symbol_preamble_footer":           o.Lint.SymbolPreambleFooter,
		"cfi_start_end_balance":            o.Lint.CfiStartEndBalance,
		"macro_balance":                    o.Lint.MacroBalance,
		"instruction_sequence_termination": o.Lint.InstructionSequenceTermination,
		"function_doxygen_comment":         o.Lint.FunctionDoxygenComment,
		"avoid_hash_and_at_comments":       o.Lint.AvoidHashAndAtComments,
		"double_label_declaration":         o.Lint.DoubleLabelDeclaration,
		"reserved_label_names":             o.Lint.ReservedLabelNames,
		"unreachable_code":                 o.Lint.UnreachableCode,
		"single_return_statement":          o.Lint.SingleReturnStatement,
		"no_fallthrough_to_function":       o.Lint.NoFallthroughToFunction,
		"no_jump_out_of_function":          o.Lint.NoJumpOutOfFunction,
		"no_recursive_calls":               o.Lint.NoRecursiveCalls,
		"copyright_and_license":            o.Lint.CopyrightAndLicense,
		"developer_name_length":            o.Lint.DeveloperNameLength,
		"line_length_limit":                o.Lint.LineLengthLimit,
	}

	for name, val := range rules {
		if err := validateSeverity(val, name); err != nil {
			return normalizedOptions{}, err
		}
		if val != "" {
			n.lint.severities[name] = val
		} else {
			n.lint.severities[name] = "ignore"
		}
	}

	// Validate casing parameters
	switch o.Lint.LabelNamingStyle {
	case "snake_case", "camelCase", "PascalCase", "any":
		n.lint.labelNamingStyle = o.Lint.LabelNamingStyle
	default:
		return normalizedOptions{}, fmt.Errorf("invalid label_naming_style %q", o.Lint.LabelNamingStyle)
	}

	switch o.Lint.MacroNamingStyle {
	case "UPPER_SNAKE_CASE", "snake_case", "any":
		n.lint.macroNamingStyle = o.Lint.MacroNamingStyle
	default:
		return normalizedOptions{}, fmt.Errorf("invalid macro_naming_style %q", o.Lint.MacroNamingStyle)
	}

	if o.Lint.LabelNamingStyle != "any" {
		n.lint.severities["label_naming_style"] = "warning"
	}
	if o.Lint.MacroNamingStyle != "any" {
		n.lint.severities["macro_naming_style"] = "warning"
	}

	return n, nil
}

package asmfmt

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Options configures asmfmt formatting behavior.
type Options struct {
	IndentStyle                   string `toml:"indent_style"`
	IndentWidth                   int    `toml:"indent_width"`
	AlignOperands                 bool   `toml:"align_operands"`
	AlignComments                 bool   `toml:"align_comments"`
	AlignContinuations            bool   `toml:"align_continuations"`
	MaxBlankLines                 int    `toml:"max_blank_lines"`
	SplitSemicolonStatements      bool   `toml:"split_semicolon_statements"`
	NewlineBeforeComments         bool   `toml:"newline_before_comments"`
	NewlineBeforeLabels           bool   `toml:"newline_before_labels"`
	LabelsAlwaysOnOwnLine         bool   `toml:"labels_always_on_own_line"`
	LineCommentSpace              bool   `toml:"line_comment_space"`
	ConvertSingleLineBlockComment bool   `toml:"convert_single_line_block_comment"`
	PreferredCommentStyle         string `toml:"preferred_comment_style"`
	SourceStyle                   string `toml:"source_style"`
	IndentGASDirectives           bool   `toml:"indent_gas_directives"`
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
	return n, nil
}

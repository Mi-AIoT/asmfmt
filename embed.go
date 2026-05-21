package asmfmt

import _ "embed"

// DefaultConfigTemplate contains the raw contents of the example configuration.
//
//go:embed .asmfmt.toml.example
var DefaultConfigTemplate []byte

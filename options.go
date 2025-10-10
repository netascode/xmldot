// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

// Options configures xmldot behavior for advanced use cases.
// Zero value (Options{}) uses default safe behavior.
//
// Example:
//
//	opts := &Options{
//	    CaseSensitive: false,
//	    Indent:        "  ",
//	}
//	result := GetWithOptions(xml, path, opts)
type Options struct {
	// CaseSensitive controls path matching case sensitivity.
	// Default: true (case-sensitive matching)
	// When false, path matching ignores case differences.
	CaseSensitive bool

	// Indent specifies indentation for formatted output (Set operations).
	// Empty string (default) preserves original formatting.
	// Use "  " or "\t" for pretty printing.
	Indent string

	// PreserveWhitespace controls whitespace handling in text content.
	// Default: false (trim whitespace from text values)
	// Phase 6: Reserved for future implementation.
	PreserveWhitespace bool

	// Namespaces maps namespace prefixes to URIs (future use).
	// Phase 6: Reserved for future implementation.
	Namespaces map[string]string
}

// DefaultOptions returns a pointer to Options with recommended defaults.
// This function is provided for convenience and documentation purposes.
//
// Default values:
//   - CaseSensitive: true (case-sensitive matching)
//   - Indent: "" (preserve original formatting)
//   - PreserveWhitespace: false (trim whitespace)
//   - Namespaces: nil (no namespace mapping)
//
// Example:
//
//	opts := DefaultOptions()
//	opts.CaseSensitive = false
//	result := GetWithOptions(xml, path, opts)
func DefaultOptions() *Options {
	return &Options{
		CaseSensitive:      true,
		Indent:             "",
		PreserveWhitespace: false,
		Namespaces:         nil,
	}
}

// isDefaultOptions checks if the given Options pointer uses all default values.
// This allows optimization to use the fast path for default options.
// nil pointer is treated as default options.
func isDefaultOptions(opts *Options) bool {
	if opts == nil {
		return true
	}
	return opts.CaseSensitive &&
		opts.Indent == "" &&
		!opts.PreserveWhitespace &&
		opts.Namespaces == nil
}

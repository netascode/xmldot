// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import "errors"

// Error types for Set and Delete operations

var (
	// ErrInvalidPath is returned when the path syntax is invalid or unsupported.
	ErrInvalidPath = errors.New("invalid path syntax")

	// ErrMalformedXML is returned when the input XML document is malformed
	// or cannot be parsed.
	ErrMalformedXML = errors.New("malformed XML document")

	// ErrInvalidValue is returned when the value cannot be converted to XML
	// or is inappropriate for the operation.
	ErrInvalidValue = errors.New("invalid value for XML")
)

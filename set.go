// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
)

// Set modifies the xml at the specified path with the given value and returns
// the modified XML. If the path does not exist, intermediate elements are created.
//
// The value can be:
//   - string, int, float, bool - converted to text content
//   - []byte - inserted as raw XML
//   - nil - removes the element (same as Delete)
//
// Security Considerations:
//
// This function implements security protections:
//   - XML validation: Input XML is validated for well-formedness before processing
//   - Document size limit: Documents larger than MaxDocumentSize (10MB) are rejected
//   - Value validation: Values are properly escaped to prevent XML injection
//   - Path validation: Paths are validated for correct syntax
//
// Error Handling:
//
// Returns ErrMalformedXML if:
//   - XML has unclosed tags
//   - XML has mismatched opening/closing tags
//   - XML exceeds size limits
//   - XML fails well-formedness checks
//
// Example:
//
//	xml := `<root><user><name>John</name></user></root>`
//	modified, _ := Set(xml, "root.user.age", 30)
//	// modified: <root><user><name>John</name><age>30</age></user></root>
func Set(xml, path string, value interface{}) (string, error) {
	result, err := SetBytes([]byte(xml), path, value)
	if err != nil {
		return xml, err
	}
	return string(result), nil
}

// SetBytes is like Set but accepts and returns xml as byte slices for efficiency.
func SetBytes(xml []byte, path string, value interface{}) ([]byte, error) {
	return SetBytesWithOptions(xml, path, value, DefaultOptions())
}

// SetRaw embeds pre-formatted XML at the specified path without parsing or escaping.
// The raw XML must be well-formed. This function performs basic validation to ensure
// the raw XML doesn't contain unmatched tags.
//
// Example:
//
//	xml := `<root></root>`
//	modified, _ := SetRaw(xml, "root.data", "<item><name>Test</name></item>")
//	// modified: <root><data><item><name>Test</name></item></data></root>
func SetRaw(xml, path, rawxml string) (string, error) {
	// Basic validation: check for balanced tags
	if err := validateRawXML(rawxml); err != nil {
		return xml, err
	}

	// Use Set with []byte to insert raw XML
	return Set(xml, path, []byte(rawxml))
}

// validateRawXML performs basic validation on raw XML to prevent injection
func validateRawXML(rawxml string) error {
	// Track opening tags on a stack to verify they match closing tags
	var tagStack []string
	i := 0

	for i < len(rawxml) {
		if rawxml[i] == '<' {
			if i+1 >= len(rawxml) {
				return ErrInvalidValue
			}

			next := rawxml[i+1]
			if next == '/' {
				// Closing tag - must match top of stack
				i += 2
				tagNameStart := i

				// Extract tag name
				for i < len(rawxml) && rawxml[i] != '>' && rawxml[i] != ' ' {
					i++
				}
				tagName := rawxml[tagNameStart:i]

				// Verify stack not empty
				if len(tagStack) == 0 {
					return ErrInvalidValue
				}

				// Verify tag matches
				if tagStack[len(tagStack)-1] != tagName {
					return ErrInvalidValue
				}

				// Pop from stack
				tagStack = tagStack[:len(tagStack)-1]

				// Skip to end of closing tag
				for i < len(rawxml) && rawxml[i] != '>' {
					i++
				}
				if i < len(rawxml) {
					i++ // Skip '>'
				}
				continue
			} else if next == '!' || next == '?' {
				// Comment, CDATA, or PI - skip
				if i+2 < len(rawxml) && rawxml[i+2] == '-' {
					// Comment: <!--
					endPos := strings.Index(rawxml[i:], "-->")
					if endPos == -1 {
						return ErrInvalidValue
					}
					i += endPos + 3
				} else if strings.HasPrefix(rawxml[i:], "<![CDATA[") {
					// CDATA - validate no nested CDATA or suspicious patterns
					cdataStart := i + 9
					cdataEnd := strings.Index(rawxml[cdataStart:], "]]>")
					if cdataEnd == -1 {
						return fmt.Errorf("%w: unclosed CDATA section", ErrInvalidValue)
					}

					// Check CDATA content for nested CDATA (invalid XML)
					cdataContent := rawxml[cdataStart : cdataStart+cdataEnd]
					if strings.Contains(cdataContent, "<![CDATA[") {
						return fmt.Errorf("%w: nested CDATA sections not allowed", ErrInvalidValue)
					}

					// NOTE: Tag-like structures in CDATA are allowed per XML spec
					// but we can optionally warn about suspicious patterns

					i = cdataStart + cdataEnd + 3
				} else if strings.HasPrefix(rawxml[i:], "<!DOCTYPE") {
					// DOCTYPE - REJECT for security
					return fmt.Errorf("%w: DOCTYPE declarations not allowed in SetRaw", ErrInvalidValue)
				} else if strings.HasPrefix(rawxml[i:], "<!ENTITY") {
					// ENTITY declarations - REJECT for security
					return fmt.Errorf("%w: ENTITY declarations not allowed in SetRaw", ErrInvalidValue)
				} else {
					// PI or other - skip to end
					endPos := strings.Index(rawxml[i:], ">")
					if endPos == -1 {
						return ErrInvalidValue
					}
					i += endPos + 1
				}
				continue
			}

			// Opening tag - extract tag name
			i++
			tagNameStart := i
			for i < len(rawxml) && rawxml[i] != '>' && rawxml[i] != ' ' && rawxml[i] != '/' {
				i++
			}
			tagName := rawxml[tagNameStart:i]

			if tagName == "" {
				return ErrInvalidValue
			}

			// Check for self-closing
			isSelfClosing := false
			for i < len(rawxml) {
				if rawxml[i] == '/' && i+1 < len(rawxml) && rawxml[i+1] == '>' {
					isSelfClosing = true
					i += 2
					break
				} else if rawxml[i] == '>' {
					i++
					break
				}
				i++
			}

			// Push to stack if not self-closing
			if !isSelfClosing {
				tagStack = append(tagStack, tagName)
			}
			continue
		}
		i++
	}

	// Verify all tags were closed
	if len(tagStack) != 0 {
		return ErrInvalidValue
	}

	return nil
}

// SetMany performs multiple Set operations, applying each modification
// sequentially. This is more convenient than calling Set multiple times manually.
// If multiple paths overlap, later operations take precedence.
//
// Performance Characteristics:
//   - Time complexity: O(n²) where n is the number of operations, as each operation
//     reparses and rebuilds the entire document
//   - Memory: Allocates a new document copy for each operation
//   - Best for: Small to medium batch sizes (< 100 operations)
//   - For very large batches (10K+ operations), performance degrades significantly
//     due to sequential document processing
//
// The sequential processing approach provides predictable semantics but means
// performance scales quadratically with operation count. Future optimizations
// may enable true single-pass processing without API changes.
//
// Benchmark data (on Apple M4 Pro):
//   - 10 ops:    ~20,000 ns/op   (30KB allocated)
//   - 100 ops:   ~1,750,000 ns/op (2.3MB allocated)
//   - 1000 ops:  ~183,000,000 ns/op (268MB allocated)
//   - 10K ops:   ~21,600,000,000 ns/op (29.8GB allocated)
//
// Example:
//
//	xml := `<root><user><name>John</name></user></root>`
//	paths := []string{"root.user.age", "root.user.email"}
//	values := []interface{}{30, "john@example.com"}
//	modified, _ := SetMany(xml, paths, values)
//	// modified: <root><user><name>John</name><age>30</age><email>john@example.com</email></user></root>
func SetMany(xml string, paths []string, values []interface{}) (string, error) {
	result, err := SetManyBytes([]byte(xml), paths, values)
	if err != nil {
		return xml, err
	}
	return string(result), nil
}

// SetManyBytes is like SetMany but accepts and returns xml as byte slices for efficiency.
func SetManyBytes(xml []byte, paths []string, values []interface{}) ([]byte, error) {
	// Security check: reject documents that are too large
	if len(xml) > MaxDocumentSize {
		return xml, ErrMalformedXML
	}

	// Validate inputs
	if len(paths) != len(values) {
		return xml, fmt.Errorf("%w: paths and values length mismatch", ErrInvalidPath)
	}

	// Empty inputs - return original XML
	if len(paths) == 0 {
		return xml, nil
	}

	// Apply operations sequentially
	// For true single-pass optimization, we'd need to batch parse/build
	// But for Phase 4, sequential application is acceptable
	result := xml
	var err error

	for i := 0; i < len(paths); i++ {
		result, err = SetBytes(result, paths[i], values[i])
		if err != nil {
			return xml, fmt.Errorf("error setting path %q: %w", paths[i], err)
		}
	}

	return result, nil
}

// Delete removes the element or attribute at the specified path and returns
// the modified XML. If the path does not exist, the original XML is returned
// unchanged (no error is returned).
//
// This function can delete:
//   - Element nodes and all their children
//   - Attributes from elements
//   - Specific array elements by index
//
// Security Considerations:
//
// This function validates input XML for well-formedness before processing.
// Returns ErrMalformedXML if the XML has unclosed tags, mismatched tags,
// or fails other well-formedness checks.
//
// Example:
//
//	xml := `<root><user><name>John</name><email>john@example.com</email></user></root>`
//	modified, _ := Delete(xml, "root.user.email")
//	// modified: <root><user><name>John</name></user></root>
func Delete(xml, path string) (string, error) {
	result, err := DeleteBytes([]byte(xml), path)
	if err != nil {
		return xml, err
	}
	return string(result), nil
}

// DeleteBytes is like Delete but accepts and returns xml as byte slices.
func DeleteBytes(xml []byte, path string) ([]byte, error) {
	return DeleteBytesWithOptions(xml, path, DefaultOptions())
}

// DeleteMany removes multiple paths sequentially. If multiple paths overlap
// (e.g., parent and child), the parent deletion takes precedence. Paths are
// processed in the order provided, and non-existent paths are silently skipped.
// Duplicate paths are automatically deduplicated.
//
// Performance Characteristics:
//   - Time complexity: O(n²) where n is the number of operations, as each operation
//     reparses and rebuilds the entire document
//   - Memory: Allocates a new document copy for each operation
//   - Best for: Small to medium batch sizes (< 100 operations)
//   - For very large batches (10K+ operations), performance degrades significantly
//     due to sequential document processing
//
// The sequential processing approach ensures earlier deletions affect later paths
// (e.g., deleting a parent removes its children). Future optimizations may enable
// true single-pass processing without API changes.
//
// Benchmark data (on Apple M4 Pro):
//   - 10 ops:    ~13,000 ns/op   (21KB allocated)
//   - 100 ops:   ~840,000 ns/op (1.3MB allocated)
//   - 1000 ops:  ~84,000,000 ns/op (144MB allocated)
//   - 10K ops:   ~9,200,000,000 ns/op (15.7GB allocated)
//
// Example:
//
//	xml := `<root><user><name>John</name><age>30</age><email>john@example.com</email></user></root>`
//	modified, _ := DeleteMany(xml, "root.user.age", "root.user.email")
//	// modified: <root><user><name>John</name></user></root>
func DeleteMany(xml string, paths ...string) (string, error) {
	result, err := DeleteManyBytes([]byte(xml), paths...)
	if err != nil {
		return xml, err
	}
	return string(result), nil
}

// DeleteManyBytes is like DeleteMany but accepts and returns xml as byte slices for efficiency.
func DeleteManyBytes(xml []byte, paths ...string) ([]byte, error) {
	// Security check: reject documents that are too large
	if len(xml) > MaxDocumentSize {
		return xml, ErrMalformedXML
	}

	// Empty inputs - return original XML
	if len(paths) == 0 {
		return xml, nil
	}

	// Deduplicate paths while preserving order
	seen := make(map[string]bool, len(paths))
	uniquePaths := make([]string, 0, len(paths))
	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}

	// Apply deletions sequentially
	// For true single-pass optimization, we'd need to batch parse/build
	// But for Phase 4, sequential application is acceptable and handles
	// parent-child relationships correctly (earlier deletions affect later ones)
	result := xml
	var err error

	for _, path := range uniquePaths {
		result, err = DeleteBytes(result, path)
		if err != nil {
			return xml, fmt.Errorf("error deleting path %q: %w", path, err)
		}
	}

	return result, nil
}

// SetWithOptions is like Set but accepts Options for behavioral control.
// Most users should use Set(); this function is for advanced use cases.
//
// Options allows customizing behavior such as:
//   - Case-insensitive path matching (CaseSensitive: false)
//   - Output indentation (Indent: "  " or "\t")
//
// Performance: If opts is nil or uses all default values, this function uses
// a fast path that calls the standard Set() directly.
//
// Example (case-insensitive set with indentation):
//
//	xml := `<ROOT><USER><NAME>John</NAME></USER></ROOT>`
//	opts := &Options{CaseSensitive: false, Indent: "  "}
//	modified, _ := SetWithOptions(xml, "root.user.age", 30, opts)
//	// modified: pretty-printed XML with proper indentation
//
// Concurrency: SetWithOptions is safe for concurrent use from multiple goroutines.
func SetWithOptions(xml, path string, value interface{}, opts *Options) (string, error) {
	result, err := SetBytesWithOptions([]byte(xml), path, value, opts)
	if err != nil {
		return xml, err
	}
	return string(result), nil
}

// SetBytesWithOptions is like SetWithOptions but accepts and returns xml as byte slices for efficiency.
// Security: Documents larger than MaxDocumentSize (10MB) are rejected to prevent
// memory exhaustion attacks.
func SetBytesWithOptions(xml []byte, path string, value interface{}, opts *Options) ([]byte, error) {
	// Security check: reject documents that are too large
	if len(xml) > MaxDocumentSize {
		return xml, ErrMalformedXML
	}

	// Validate XML well-formedness unless in optimistic mode (future feature)
	// This prevents crashes from malformed XML discovered by fuzz testing
	if !ValidBytes(xml) {
		return xml, ErrMalformedXML
	}

	// Handle nil value as deletion
	if value == nil {
		return DeleteBytesWithOptions(xml, path, opts)
	}

	// Parse the path with options-aware parsing
	segments := parsePathWithOptions(path, opts)
	if len(segments) == 0 {
		return xml, ErrInvalidPath
	}

	// Create builder with options
	builder := newXMLBuilderWithOptions(xml, opts)

	// Execute the set operation
	if err := builder.setElement(segments, value); err != nil {
		return xml, err
	}

	return []byte(builder.getResult()), nil
}

// DeleteBytesWithOptions is like DeleteBytes but accepts Options for behavioral control.
// This is used internally by SetBytesWithOptions when value is nil.
func DeleteBytesWithOptions(xml []byte, path string, opts *Options) ([]byte, error) {
	// Security check: reject documents that are too large
	if len(xml) > MaxDocumentSize {
		return xml, ErrMalformedXML
	}

	// Validate XML well-formedness unless in optimistic mode (future feature)
	// This prevents crashes from malformed XML discovered by fuzz testing
	if !ValidBytes(xml) {
		return xml, ErrMalformedXML
	}

	// Parse the path with options-aware parsing
	segments := parsePathWithOptions(path, opts)
	if len(segments) == 0 {
		return xml, ErrInvalidPath
	}

	// Create builder with options
	builder := newXMLBuilderWithOptions(xml, opts)

	// Execute the delete operation
	if err := builder.deleteElement(segments); err != nil {
		return xml, err
	}

	return []byte(builder.getResult()), nil
}

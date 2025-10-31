// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// MaxPathSegments is the maximum number of path segments allowed in a query path.
// This prevents DoS attacks with extremely complex paths.
const MaxPathSegments = 100

// MaxFieldNameLength is the maximum length of a field name in #.field syntax.
// This prevents DoS attacks with extremely long field names.
const MaxFieldNameLength = 256

// Path cache for performance optimization
// Thread-safe LRU-style cache for parsed paths to avoid repeated parsing
var (
	pathCache      = make(map[string][]PathSegment)
	pathCacheMu    sync.RWMutex
	pathCacheLimit = 256 // Keep cache size reasonable
)

// SegmentType represents the type of a path segment.
type SegmentType int

const (
	// SegmentElement represents an element name in the path.
	SegmentElement SegmentType = iota
	// SegmentAttribute represents an attribute access (@attribute).
	SegmentAttribute
	// SegmentIndex represents an array index access ([n] or .n).
	SegmentIndex
	// SegmentWildcard represents a wildcard match (* or **).
	SegmentWildcard
	// SegmentFilter represents a query filter ([condition]).
	SegmentFilter
	// SegmentText represents text content access (%).
	SegmentText
	// SegmentCount represents array length/count (#).
	SegmentCount
	// SegmentFieldExtraction represents field extraction from all array elements (#.field).
	SegmentFieldExtraction
)

// IndexIntent represents the semantic intent of an index operation.
// This allows clean separation between syntax (what user wrote) and
// semantics (what should happen).
type IndexIntent int

const (
	// IntentReplace replaces an existing element at a specific index.
	// Example: items.item.0 replaces the first item
	IntentReplace IndexIntent = iota

	// IntentAppend creates a NEW element at the end of the array.
	// Only triggered by -1 index in Set/SetRaw operations.
	// Example: items.item.-1 appends a new item
	IntentAppend

	// IntentAccess reads an element (used in Get, not Set/SetRaw).
	// Example: items.item.-1 accesses last element (Get only)
	IntentAccess

	// Future extensibility:
	// IntentInsertBefore (for -2, -3, etc.)
	// IntentConditionalAppend (append only if not exists)
)

// PathSegment represents a single segment in a parsed path.
type PathSegment struct {
	// Type is the type of this path segment.
	Type SegmentType
	// Value is the element or attribute name.
	Value string
	// Index is the array index if Type is SegmentIndex.
	Index int
	// Intent represents the semantic intent for index operations (only set for SegmentIndex).
	// Enables clean separation of syntax (Index=-1) from semantics (Intent=IntentAppend).
	Intent IndexIntent
	// Wildcard indicates if this is a recursive wildcard (**).
	Wildcard bool
	// Filter contains the filter expression if Type is SegmentFilter.
	Filter *Filter
	// FilterAll indicates if #()# syntax is used (returns ALL matches instead of first).
	// Only applies when Type is SegmentFilter.
	FilterAll bool
	// Field is the field name for FieldExtraction type (#.field syntax).
	// The field can be an element name, attribute (@attr), or text (%).
	Field string
	// Modifiers contains modifiers to apply after this segment matches (Phase 6).
	// Modifiers execute in order after the path segment resolves.
	// Example: "items.item|@sort|@first" applies @sort then @first to "item" results.
	Modifiers []string
}

// Filter represents a query filter condition.
type Filter struct {
	// Path is the path to evaluate in the filter (e.g., "age" or "@id").
	Path string
	// Op is the comparison operator.
	Op FilterOp
	// Value is the value to compare against.
	Value string
}

// parsePath parses a path string into a slice of PathSegments.
// Supported syntax:
//   - "root.child.element" - element path
//   - "element.@attribute" - attribute access
//   - "elements.element.0" - array index
//   - "elements.element.#" - array count
//   - "element.%" - text content only
//   - "root.*.name" - single-level wildcard
//   - "root.**.price" - recursive wildcard
//
// Security: Paths with more than MaxPathSegments segments are rejected.
// Performance: Uses a thread-safe LRU cache to avoid re-parsing common paths.
func parsePath(path string) []PathSegment {
	if path == "" {
		return nil
	}

	// Check cache first (read lock)
	pathCacheMu.RLock()
	if cached, ok := pathCache[path]; ok {
		pathCacheMu.RUnlock()
		// Return a copy to prevent modification of cached data
		result := make([]PathSegment, len(cached))
		copy(result, cached)
		return result
	}
	pathCacheMu.RUnlock()

	// Parse the path
	segments := parsePathInternal(path)

	// Cache the result (write lock)
	if segments != nil {
		pathCacheMu.Lock()
		// Simple cache eviction: clear cache when it exceeds limit
		if len(pathCache) >= pathCacheLimit {
			pathCache = make(map[string][]PathSegment)
		}
		// Store a copy to prevent external modification
		cached := make([]PathSegment, len(segments))
		copy(cached, segments)
		pathCache[path] = cached
		pathCacheMu.Unlock()
	}

	return segments
}

// parsePathInternal performs the actual path parsing logic.
// This is separated from parsePath to enable caching.
func parsePathInternal(path string) []PathSegment {
	parts := splitPath(path)

	// Security check: enforce maximum path segment count
	if len(parts) > MaxPathSegments {
		return nil
	}

	segments := make([]PathSegment, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Parse modifiers from this path component (Phase 6)
		pathPart, modifiers := parseModifiers(part)

		seg := PathSegment{
			Modifiers: modifiers, // Store modifiers for this segment
		}

		// Check for GJSON filter syntax #(...) or #(...)#
		if strings.HasPrefix(pathPart, "#(") {
			// Validate proper closing
			var validSyntax bool
			var filterAll bool
			var endIdx int

			if strings.HasSuffix(pathPart, ")#") {
				// Check it's exactly )# at the end, not multiple #'s
				if len(pathPart) >= 3 && pathPart[len(pathPart)-2] == ')' && pathPart[len(pathPart)-1] == '#' {
					validSyntax = true
					filterAll = true
					endIdx = len(pathPart) - 2
				}
			} else if strings.HasSuffix(pathPart, ")") {
				validSyntax = true
				filterAll = false
				endIdx = len(pathPart) - 1
			}

			// Skip malformed filters
			if !validSyntax {
				continue
			}

			seg.Type = SegmentFilter
			seg.FilterAll = filterAll

			// Extract condition between markers
			startIdx := len("#(")

			// Security check: validate closing marker exists and condition is not empty
			if endIdx <= startIdx {
				// Empty condition - skip this segment
				continue
			}

			condition := pathPart[startIdx:endIdx]
			filter, err := parseFilterCondition(condition)
			if err != nil {
				// Invalid filter condition (e.g., control characters) - reject entire path
				// Returning nil causes Get() to return Null result
				return nil
			}
			seg.Filter = filter
			segments = append(segments, seg)
			continue
		}

		// Note: Old bracket filter syntax [condition] is not supported
		// Only GJSON-style #(condition) syntax is supported
		// Note: #.field syntax is handled in post-processing (see end of function)

		if strings.HasPrefix(pathPart, "@") {
			// Attribute access
			seg.Type = SegmentAttribute
			seg.Value = pathPart[1:]
		} else if pathPart == "%" {
			// Text content
			seg.Type = SegmentText
		} else if pathPart == "#" {
			// Array count
			seg.Type = SegmentCount
		} else if pathPart == "*" {
			// Single-level wildcard
			seg.Type = SegmentWildcard
			seg.Wildcard = false
		} else if pathPart == "**" {
			// Recursive wildcard
			seg.Type = SegmentWildcard
			seg.Wildcard = true
		} else if isNumeric(pathPart) {
			// Array index (numeric)
			seg.Type = SegmentIndex
			seg.Index, _ = strconv.Atoi(pathPart)
		} else {
			// Element name
			seg.Type = SegmentElement
			seg.Value = pathPart
		}

		segments = append(segments, seg)
	}

	// Post-processing: Convert # followed by element/attribute/text into field extraction
	// This handles the GJSON #.field pattern which gets split by splitPath
	processedSegments := make([]PathSegment, 0, len(segments))
	for i := 0; i < len(segments); i++ {
		seg := segments[i]

		// Check if this is # followed by another segment
		if seg.Type == SegmentCount && i+1 < len(segments) {
			nextSeg := segments[i+1]

			// Convert # followed by element/attribute/text into field extraction
			if nextSeg.Type == SegmentElement || nextSeg.Type == SegmentAttribute || nextSeg.Type == SegmentText {
				// Create field extraction segment
				fieldSeg := PathSegment{
					Type:      SegmentFieldExtraction,
					Modifiers: nextSeg.Modifiers, // Preserve modifiers from the field segment
				}

				// Determine field name based on next segment type
				switch nextSeg.Type {
				case SegmentAttribute:
					fieldSeg.Field = "@" + nextSeg.Value
				case SegmentText:
					fieldSeg.Field = "%"
				default:
					fieldSeg.Field = nextSeg.Value
				}

				// Validate field name
				if isValidFieldName(fieldSeg.Field) && len(fieldSeg.Field) <= MaxFieldNameLength {
					processedSegments = append(processedSegments, fieldSeg)
					i++ // Skip the next segment since we consumed it
					continue
				}
			}
		}

		// Not a field extraction pattern, keep segment as-is
		processedSegments = append(processedSegments, seg)
	}

	return processedSegments
}

// splitPath splits a path on dots, handling escapes
func splitPath(path string) []string {
	if path == "" {
		return nil
	}

	var parts []string
	var current strings.Builder
	escaped := false

	for i := 0; i < len(path); i++ {
		c := path[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == '.' {
			// Split point
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	// Add the last part (even if empty, for cases like "root.child.")
	parts = append(parts, current.String())

	return parts
}

// isNumeric checks if a string is a valid integer
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			if c == '-' && len(s) > 1 {
				// Allow negative indices
				continue
			}
			return false
		}
	}
	return true
}

// isValidFieldName validates a field name for #.field extraction
// Allows element names, @attribute, and % (text content)
func isValidFieldName(fieldName string) bool {
	if fieldName == "" {
		return false
	}

	// Allow special cases
	if fieldName == "%" {
		return true // Text content extraction
	}

	// Handle attribute extraction @attr
	if strings.HasPrefix(fieldName, "@") {
		if len(fieldName) == 1 {
			return false // Just @ is invalid
		}
		// Validate attribute name (after @)
		attrName := fieldName[1:]
		return isValidIdentifier(attrName)
	}

	// Regular element name
	return isValidIdentifier(fieldName)
}

// isValidIdentifier checks if a string is a valid XML identifier (element/attribute name)
// Allows: letters, digits, hyphens, underscores, colons (for namespaces)
// Must start with letter or underscore
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// First character must be letter or underscore
	first := s[0]
	if (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') && first != '_' {
		return false
	}

	// Remaining characters can be letters, digits, hyphens, underscores, or colons
	for i := 1; i < len(s); i++ {
		c := s[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
			(c < '0' || c > '9') && c != '-' && c != '_' && c != ':' {
			return false
		}
	}

	return true
}

// matches checks if a path segment matches an element name
// Phase 6: Implements namespace-aware matching for prefixed elements
func (seg PathSegment) matches(elementName string) bool {
	switch seg.Type {
	case SegmentElement:
		// Split namespace from path segment value
		pathPrefix, pathLocal := splitNamespace(seg.Value)

		// Split namespace from element name
		elemPrefix, elemLocal := splitNamespace(elementName)

		// If path has prefix, match exactly (prefix:localname == prefix:localname)
		if pathPrefix != "" {
			return pathPrefix == elemPrefix && pathLocal == elemLocal
		}

		// If path has no prefix, match by local name only (backward compatible)
		return pathLocal == elemLocal

	case SegmentWildcard:
		return true // Wildcards match any element
	default:
		return false
	}
}

// matchesWithOptions checks if a path segment matches an element name with Options support.
// Phase 6: Implements case-insensitive matching when opts.CaseSensitive is false.
// Phase 6: Implements namespace-aware matching for prefixed elements.
// The segment value is expected to be pre-lowercased by parsePathWithOptions for case-insensitive matching.
func (seg PathSegment) matchesWithOptions(elementName string, opts *Options) bool {
	// Treat nil options as default options to prevent nil pointer dereferences
	if opts == nil {
		opts = DefaultOptions()
	}

	switch seg.Type {
	case SegmentElement:
		// Split namespace from path segment value
		pathPrefix, pathLocal := splitNamespace(seg.Value)

		// Split namespace from element name
		elemPrefix, elemLocal := splitNamespace(elementName)

		// Normalize case if needed
		if !opts.CaseSensitive {
			pathPrefix = toLowerASCII(pathPrefix)
			pathLocal = toLowerASCII(pathLocal)
			elemPrefix = toLowerASCII(elemPrefix)
			elemLocal = toLowerASCII(elemLocal)
		}

		// If path has prefix, match exactly (prefix:localname == prefix:localname)
		if pathPrefix != "" {
			return pathPrefix == elemPrefix && pathLocal == elemLocal
		}

		// If path has no prefix, match by local name only (backward compatible)
		return pathLocal == elemLocal

	case SegmentWildcard:
		return true // Wildcards match any element
	default:
		return false
	}
}

// toLowerASCII converts ASCII letters to lowercase (fast path, no Unicode support needed).
// This is used for case-insensitive matching of element and attribute names.
func toLowerASCII(s string) string {
	// Fast path: check if string needs lowercasing
	hasUpper := false
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return s
	}

	// Lowercase the string
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		b[i] = c
	}
	return string(b)
}

// splitNamespace splits a name into namespace prefix and local name.
// Returns ("", localName) if no colon is present (no namespace).
// Returns ("prefix", "localName") if colon is present.
//
// Security: If the prefix length exceeds MaxNamespacePrefixLength,
// treats the entire name as unprefixed to prevent memory exhaustion attacks.
//
// ⚠️ NAMESPACE SUPPORT - IMPORTANT LIMITATIONS ⚠️
//
// This library provides BASIC namespace prefix matching only.
// It does NOT implement the full XML Namespaces specification (https://www.w3.org/TR/xml-names/).
//
// What Works:
//   - Prefix-aware matching: "soap:Envelope" matches <soap:Envelope>
//   - Backward compatible: "Envelope" matches both <Envelope> and <soap:Envelope>
//   - Attribute prefixes: "@xmlns:ns" matches xmlns:ns attributes
//
// What DOES NOT Work (Critical Limitations):
//
//	❌ NO namespace URI resolution (xmlns attributes are completely ignored)
//	❌ NO validation that prefixes are declared
//	❌ NO default namespace support (xmlns="..." is ignored)
//	❌ Elements with same local name but different namespace URIs are NOT distinguished
//	❌ No namespace inheritance or scoping
//
// Example Demonstrating Limitations:
//
//	XML:
//	  <root xmlns:a="http://example.com/a" xmlns:b="http://example.com/b">
//	    <a:item>Value A</a:item>
//	    <b:item>Value B</b:item>
//	  </root>
//
//	Path: "root.a:item"
//	  ✓ Matches <a:item> based on prefix string only
//	  ✗ Does NOT verify that 'a' maps to "http://example.com/a"
//	  ✗ Library has NO concept that a:item and b:item are semantically different
//
// When to Use This Library:
//
//	✓ Simple XML with predictable namespace prefixes
//	✓ Internal APIs where namespace URIs don't change
//	✓ Performance-critical applications where full namespace processing is too slow
//	✓ Documents where you control the namespace prefix conventions
//
// When NOT to Use (Use encoding/xml instead):
//
//	✗ Processing arbitrary XML from external sources
//	✗ Need semantic namespace URI resolution
//	✗ Standards-compliant XML processing (SOAP, XHTML, etc.)
//	✗ Documents with dynamic or varying namespace prefixes
//	✗ Applications requiring namespace validation
//
// Security Note:
//   - Prefix length limited to MaxNamespacePrefixLength (256) to prevent DoS attacks
//   - Control characters in prefixes are validated and rejected
//
// For production systems requiring full XML Namespaces support, use encoding/xml.
func splitNamespace(name string) (prefix, localName string) {
	idx := strings.IndexByte(name, ':')
	if idx == -1 {
		return "", name
	}

	// Security: Check prefix length
	if idx > MaxNamespacePrefixLength {
		return "", name // Treat as unprefixed if too long
	}

	prefix = name[:idx]
	localName = name[idx+1:]

	// Security: Validate prefix doesn't contain control characters or null bytes
	// This prevents injection attacks and ensures well-formed namespace prefixes
	for i := 0; i < len(prefix); i++ {
		c := prefix[i]
		if c < 0x20 || c == 0x7F {
			// Control character detected (including null byte, tab, newline, etc.)
			// Treat entire name as unprefixed to reject malformed input
			return "", name
		}
	}

	return prefix, localName
}

// resolveIndexIntent determines the semantic intent of an index segment
// based on context (Set vs Get) and validation rules.
//
// Context:
//   - opContext: "set" for Set/SetRaw, "get" for Get/Delete
//
// Rules:
//   - Index >= 0: always IntentReplace (or IntentAccess in Get)
//   - Index == -1 in Set: IntentAppend if allowed, error if nested path
//   - Index == -1 in Get: IntentAccess (existing behavior)
//   - Index < -1: Reserved for future use, currently error
//
// Returns:
//   - IndexIntent and nil if valid
//   - IntentReplace and error if invalid (e.g., nested append)
func resolveIndexIntent(seg PathSegment, segIndex int, segments []PathSegment, opContext string) (IndexIntent, error) {
	if seg.Type != SegmentIndex {
		return IntentReplace, fmt.Errorf("not an index segment")
	}

	// Non-negative indices: always replace/access
	if seg.Index >= 0 {
		if opContext == "set" {
			return IntentReplace, nil
		}
		return IntentAccess, nil
	}

	// Negative indices
	switch seg.Index {
	case -1:
		if opContext == "get" {
			return IntentAccess, nil // Existing Get behavior (not yet implemented)
		}

		// Set operation: check if this is last segment
		if segIndex != len(segments)-1 {
			// Nested path like item.-1.child is not supported
			return IntentReplace, fmt.Errorf("%w: append (-1 index) cannot be used with nested paths (e.g., item.-1.child)", ErrInvalidPath)
		}

		return IntentAppend, nil

	default:
		// -2, -3, etc: reserved for future use
		return IntentReplace, fmt.Errorf("%w: index %d reserved for future use; only -1 (append) is currently supported", ErrInvalidPath, seg.Index)
	}
}

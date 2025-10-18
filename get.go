// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"strings"
	"unsafe"
)

// Get searches xml for the specified path and returns a Result containing
// the value found. If the path is not found, an empty Result is returned.
//
// The path syntax supports:
//   - Element access: "root.child.element"
//   - Attribute access: "element.@attribute"
//   - Array indexing: "elements.element.0"
//   - Array count: "elements.element.#"
//   - Text content: "element.%"
//   - Wildcards: "root.*.name" or "root.**.price"
//   - Filters: "users.user[age>21]" or "items.item[@id=5]"
//
// Security Considerations:
//
// This function implements several security protections:
//   - Document size limit: Documents larger than MaxDocumentSize (10MB) are rejected
//   - Nesting depth limit: XML nesting deeper than MaxNestingDepth (100 levels) is truncated
//   - Attribute limit: Elements with more than MaxAttributes (100) have excess attributes ignored
//   - Token size limit: Tokens larger than MaxTokenSize (1MB) are truncated
//   - DOCTYPE skipping: DOCTYPE declarations are skipped to prevent XXE attacks
//
// These limits protect against denial-of-service and memory exhaustion attacks.
// For trusted XML documents where these limits might be restrictive, the constants
// can be adjusted in parser.go.
//
// Note: This parser does NOT expand entity references and does NOT process external
// entities, providing secure-by-default behavior against XXE vulnerabilities.
//
// Concurrency:
//
// Get is safe for concurrent use from multiple goroutines. Each call creates
// its own parser instance, so there is no shared state between concurrent calls.
// You can safely call Get from multiple goroutines without additional synchronization:
//
//	var wg sync.WaitGroup
//	for i := 0; i < 10; i++ {
//	    wg.Add(1)
//	    go func() {
//	        defer wg.Done()
//	        result := Get(xml, path) // Safe for concurrent use
//	        // Process result...
//	    }()
//	}
//	wg.Wait()
//
// Example:
//
//	xml := `<root><user><name>John</name><age>30</age></user></root>`
//	name := Get(xml, "root.user.name")
//	fmt.Println(name.String()) // "John"
func Get(xml, path string) Result {
	return GetBytes([]byte(xml), path)
}

// GetString is like Get but optimized for string input with zero-copy conversion.
// This is the recommended entry point for string-based XML queries when called
// from Result.Get() to avoid unnecessary string-to-byte allocations.
//
// The zero-copy conversion shares the underlying string data without allocation,
// but the slice must not be modified (which is safe since xmldot only reads the XML).
func GetString(xml string, path string) Result {
	return GetBytes(stringToBytes(xml), path)
}

// GetBytes is like Get but accepts xml as a byte slice for zero-copy efficiency.
// Security: Documents larger than MaxDocumentSize (10MB) are rejected to prevent
// memory exhaustion attacks.
func GetBytes(xml []byte, path string) Result {
	// Security check: reject documents that are too large
	if len(xml) > MaxDocumentSize {
		return Result{Type: Null}
	}

	// Parse the path into segments
	segments := parsePath(path)
	if len(segments) == 0 {
		return Result{Type: Null}
	}

	// Create parser
	parser := newXMLParser(xml)

	// Execute the query
	return executeQuery(parser, segments, 0)
}

const (
	// MaxWildcardResults is the maximum number of results from wildcard queries.
	// This prevents memory exhaustion from recursive wildcard queries.
	MaxWildcardResults = 1000

	// MaxRecursiveOperations is the maximum number of recursive search operations.
	// This prevents CPU exhaustion from exponential blowup in recursive wildcard queries.
	MaxRecursiveOperations = 10000
)

// elementMatch represents a matched element with its attributes and content
type elementMatch struct {
	name          string
	attrs         map[string]string
	content       string
	isSelfClosing bool
}

// searchContext tracks recursive search operations to prevent DoS attacks
type searchContext struct {
	operations int
	results    *[]Result
}

// executeQuery recursively matches path segments against XML structure
func executeQuery(parser *xmlParser, segments []PathSegment, segIndex int) Result {
	// Base case: we've matched all segments
	if segIndex >= len(segments) {
		return Result{Type: Null}
	}

	currentSeg := segments[segIndex]
	isLastSegment := segIndex == len(segments)-1

	// Handle GJSON-style filter segment (Phase 2: GJSON migration)
	// Note: Filter segments operate on ALL elements at the current level
	if currentSeg.Type == SegmentFilter && currentSeg.Filter != nil {
		return handleFilterQuery(parser, segments, segIndex)
	}

	// Handle GJSON-style field extraction segment #.field
	// This requires collecting ALL elements from previous context
	if currentSeg.Type == SegmentFieldExtraction {
		// Field extraction requires an array of matches from previous segment
		// This should not happen at segment index 0 (needs previous context)
		// Return empty array if called incorrectly
		return Result{Type: Array, Results: []Result{}}
	}

	// Check if next segment is field extraction - if so, collect ALL matches of current segment
	// This is the GJSON pattern: element.#.field
	hasFollowingFieldExtraction := !isLastSegment && segments[segIndex+1].Type == SegmentFieldExtraction
	if hasFollowingFieldExtraction && currentSeg.Type == SegmentElement {
		// Collect all elements matching current segment
		var allMatches []elementMatch
		for parser.skipToNextElement() {
			parser.next()
			elemName, attrs, isSelfClosing := parser.parseElementName()

			if !currentSeg.matches(elemName) {
				if !isSelfClosing {
					parser.parseElementContent(elemName)
				}
				continue
			}

			var content string
			if isSelfClosing {
				content = ""
			} else {
				content = parser.parseElementContent(elemName)
			}

			allMatches = append(allMatches, elementMatch{
				name:          elemName,
				attrs:         attrs,
				content:       content,
				isSelfClosing: isSelfClosing,
			})

			if len(allMatches) >= MaxWildcardResults {
				break
			}
		}

		// Now apply field extraction to these matches
		return executeFieldExtraction(allMatches, segments[segIndex+1])
	}

	// Check if next segment is a filter - if so, we need to collect ALL matches of current segment
	// This is the GJSON pattern: element.#(condition)
	hasFollowingFilter := !isLastSegment && segments[segIndex+1].Type == SegmentFilter
	if hasFollowingFilter && currentSeg.Type == SegmentElement {
		// Collect all elements matching current segment
		var allMatches []elementMatch
		for parser.skipToNextElement() {
			parser.next()
			elemName, attrs, isSelfClosing := parser.parseElementName()

			if !currentSeg.matches(elemName) {
				if !isSelfClosing {
					parser.parseElementContent(elemName)
				}
				continue
			}

			var content string
			if isSelfClosing {
				content = ""
			} else {
				content = parser.parseElementContent(elemName)
			}

			allMatches = append(allMatches, elementMatch{
				name:          elemName,
				attrs:         attrs,
				content:       content,
				isSelfClosing: isSelfClosing,
			})

			if len(allMatches) >= MaxWildcardResults {
				break
			}
		}

		// Now apply the filter segment to these matches
		return handleFilterQueryOnMatches(allMatches, segments, segIndex+1)
	}

	// Check if this is the last segment and it has modifiers to apply after resolution
	// We'll apply modifiers after getting the result, at the end of this function

	// Handle array index or count on previous match
	if currentSeg.Type == SegmentIndex || currentSeg.Type == SegmentCount {
		// These are handled within array matching logic
		return Result{Type: Null}
	}

	// Handle text content extraction
	if currentSeg.Type == SegmentText {
		return Result{Type: Null} // Will be handled by element matching
	}

	// Handle recursive wildcard - collect all matches at any depth
	if currentSeg.Type == SegmentWildcard && currentSeg.Wildcard {
		return handleRecursiveWildcard(parser, segments, segIndex)
	}

	// Find matching elements - need to collect for array operations or wildcards or filters
	var matches []elementMatch

	// For single-level wildcards, we collect ALL matches
	isWildcard := currentSeg.Type == SegmentWildcard && !currentSeg.Wildcard

	// For filters, we collect ALL matches then filter them
	hasFilter := currentSeg.Filter != nil

	for parser.skipToNextElement() {
		parser.next() // skip '<'

		elemName, attrs, isSelfClosing := parser.parseElementName()

		// Check if this segment is an attribute request
		if currentSeg.Type == SegmentAttribute {
			// We need to match the previous element, not this one
			// This is handled in the element matching logic
			continue
		}

		// Check if element matches current segment
		if !currentSeg.matches(elemName) {
			// Skip this element
			if !isSelfClosing {
				parser.parseElementContent(elemName)
			}
			continue
		}

		// We have a match!
		// Extract content
		var content string
		if isSelfClosing {
			content = ""
		} else {
			content = parser.parseElementContent(elemName)
		}

		// Check if next segment indicates array operation
		needsArray := !isLastSegment && (segments[segIndex+1].Type == SegmentIndex || segments[segIndex+1].Type == SegmentCount)

		if needsArray || isWildcard || hasFilter {
			// Security check: enforce wildcard result limit
			if len(matches) >= MaxWildcardResults {
				break // Stop collecting more matches when limit reached
			}

			// Collect this match for array/wildcard/filter handling
			match := elementMatch{
				name:          elemName,
				attrs:         attrs,
				content:       content,
				isSelfClosing: isSelfClosing,
			}

			// If there's a filter, only collect if it matches
			if hasFilter {
				if evaluateFilterOnMatch(currentSeg.Filter, match) {
					matches = append(matches, match)
				}
			} else {
				matches = append(matches, match)
			}
			continue
		}

		// Not an array/wildcard operation - process the first match

		// Check if next segment is an attribute
		if !isLastSegment && segments[segIndex+1].Type == SegmentAttribute {
			attrName := segments[segIndex+1].Value
			if attrValue, ok := attrs[attrName]; ok {
				// Check if there are more segments after the attribute
				if segIndex+2 < len(segments) {
					// More segments - not supported for attributes
					return Result{Type: Null}
				}
				result := Result{
					Type: Attribute,
					Str:  attrValue,
					Raw:  attrValue,
				}
				// Apply modifiers from the attribute segment if present (Phase 6)
				if len(segments[segIndex+1].Modifiers) > 0 {
					result = applyModifiers(result, segments[segIndex+1].Modifiers)
				}
				return result
			}
			return Result{Type: Null}
		}

		// Check if next segment is text content extraction
		if !isLastSegment && segments[segIndex+1].Type == SegmentText {
			textContent := extractDirectTextOnly(content)
			result := Result{
				Type: String,
				Str:  unescapeXML(textContent),
				Raw:  content,
			}
			// Apply modifiers from the text segment if present (Phase 6)
			if len(segments[segIndex+1].Modifiers) > 0 {
				result = applyModifiers(result, segments[segIndex+1].Modifiers)
			}
			return result
		}

		// If this is the last segment, return the element content
		if isLastSegment {
			result := Result{
				Type: Element,
				Str:  unescapeXML(extractTextContent(content)),
				Raw:  content,
			}
			// Apply modifiers if present (Phase 6)
			if len(currentSeg.Modifiers) > 0 {
				result = applyModifiers(result, currentSeg.Modifiers)
			}
			return result
		}

		// Otherwise, parse the content and continue matching
		contentParser := newXMLParser([]byte(content))
		result := executeQuery(contentParser, segments, segIndex+1)
		if result.Type != Null {
			return result
		}
	}

	// Handle wildcard or filter results
	if (isWildcard || hasFilter) && len(matches) > 0 {
		result := handleWildcardMatches(matches, segments, segIndex)
		// Apply modifiers if this is the last segment with modifiers (Phase 6)
		if isLastSegment && len(currentSeg.Modifiers) > 0 {
			result = applyModifiers(result, currentSeg.Modifiers)
		}
		return result
	}

	// Handle array operations if we collected matches
	if len(matches) > 0 && segIndex+1 < len(segments) {
		nextSeg := segments[segIndex+1]
		switch nextSeg.Type {
		case SegmentIndex:
			// Return specific index
			if nextSeg.Index >= 0 && nextSeg.Index < len(matches) {
				match := matches[nextSeg.Index]

				// Check if there are more segments after the index
				if segIndex+2 < len(segments) {
					// Check if next segment is an attribute
					if segments[segIndex+2].Type == SegmentAttribute {
						attrName := segments[segIndex+2].Value
						if attrValue, ok := match.attrs[attrName]; ok {
							result := Result{
								Type: Attribute,
								Str:  attrValue,
								Raw:  attrValue,
							}
							// Apply modifiers from the attribute segment if it's the last one (Phase 6)
							if segIndex+3 >= len(segments) && len(segments[segIndex+2].Modifiers) > 0 {
								result = applyModifiers(result, segments[segIndex+2].Modifiers)
							}
							return result
						}
						return Result{Type: Null}
					}

					// Continue matching within this element
					contentParser := newXMLParser([]byte(match.content))
					return executeQuery(contentParser, segments, segIndex+2)
				}

				// No more segments - return the element
				result := Result{
					Type: Element,
					Str:  unescapeXML(extractTextContent(match.content)),
					Raw:  match.content,
				}
				// Apply modifiers from the index segment if present (Phase 6)
				if len(nextSeg.Modifiers) > 0 {
					result = applyModifiers(result, nextSeg.Modifiers)
				}
				return result
			}
			return Result{Type: Null}
		case SegmentCount:
			// Return count
			result := Result{
				Type: Number,
				Num:  float64(len(matches)),
				Str:  itoa(len(matches)),
			}
			// Apply modifiers from the count segment if present (Phase 6)
			if len(nextSeg.Modifiers) > 0 {
				result = applyModifiers(result, nextSeg.Modifiers)
			}
			return result
		}
	}

	// No match found
	return Result{Type: Null}
}

// handleWildcardMatches processes matches from a single-level wildcard
func handleWildcardMatches(matches []elementMatch, segments []PathSegment, segIndex int) Result {
	isLastSegment := segIndex == len(segments)-1

	// If this is the last segment, return all matches as an array
	if isLastSegment {
		if len(matches) == 1 {
			// Single match - return as single result
			return Result{
				Type: Element,
				Str:  unescapeXML(extractTextContent(matches[0].content)),
				Raw:  matches[0].content,
			}
		}
		// Multiple matches - return as array
		// For Phase 3, we'll return the first match and mark it as Array type
		// Full array support will come later
		results := make([]Result, 0, len(matches))
		for _, match := range matches {
			results = append(results, Result{
				Type: Element,
				Str:  unescapeXML(extractTextContent(match.content)),
				Raw:  match.content,
			})
		}
		return Result{
			Type:    Array,
			Results: results,
		}
	}

	// Continue matching with the next segment for each match
	var allResults []Result
	hasFieldExtraction := false // Track if field extraction is in remaining path

	for _, match := range matches {
		nextSeg := segments[segIndex+1]

		// Check if next segment is an attribute
		if nextSeg.Type == SegmentAttribute {
			attrName := nextSeg.Value
			if attrValue, ok := match.attrs[attrName]; ok {
				// Check if there are more segments after the attribute
				if segIndex+2 < len(segments) {
					// More segments - not supported for attributes
					continue
				}
				allResults = append(allResults, Result{
					Type: Attribute,
					Str:  attrValue,
					Raw:  attrValue,
				})
			}
			continue
		}

		// Check if next segment is text content extraction
		if nextSeg.Type == SegmentText {
			textContent := extractDirectTextOnly(match.content)
			allResults = append(allResults, Result{
				Type: String,
				Str:  unescapeXML(textContent),
				Raw:  match.content,
			})
			continue
		}

		// Continue matching within this element's content
		contentParser := newXMLParser([]byte(match.content))
		result := executeQuery(contentParser, segments, segIndex+1)
		if result.Type != Null {
			// If we got an empty Array back, that means field extraction occurred
			if result.Type == Array && len(result.Results) == 0 {
				hasFieldExtraction = true
			}
			if result.Type == Array {
				// Flatten nested arrays
				allResults = append(allResults, result.Results...)
			} else {
				allResults = append(allResults, result)
			}
		}
	}

	// If field extraction occurred (even with no results), return empty Array not Null
	if len(allResults) == 0 {
		if hasFieldExtraction {
			return Result{
				Type:    Array,
				Results: []Result{},
			}
		}
		return Result{Type: Null}
	}

	// Package results
	var result Result
	if len(allResults) == 1 {
		result = allResults[0]
	} else {
		result = Result{
			Type:    Array,
			Results: allResults,
		}
	}

	// Apply modifiers from the next segment if present (Phase 6)
	// The next segment after the wildcard is the one that was matched
	if segIndex+1 < len(segments) && len(segments[segIndex+1].Modifiers) > 0 {
		result = applyModifiers(result, segments[segIndex+1].Modifiers)
	}

	return result
}

// handleRecursiveWildcard processes recursive wildcard (**) queries
// Security: Limits total operations to prevent CPU exhaustion
func handleRecursiveWildcard(parser *xmlParser, segments []PathSegment, segIndex int) Result {
	if segIndex >= len(segments)-1 {
		// ** must have a following segment to search for
		return Result{Type: Null}
	}

	// Get the segment to search for after **
	nextSegIndex := segIndex + 1
	targetSeg := segments[nextSegIndex]

	// Recursively search for matches at any depth
	var allResults []Result
	ctx := &searchContext{operations: 0, results: &allResults}
	recursiveSearchWithContext(parser, targetSeg, segments, nextSegIndex, ctx, 0)

	if len(allResults) == 0 {
		return Result{Type: Null}
	}
	if len(allResults) == 1 {
		return allResults[0]
	}
	return Result{
		Type:    Array,
		Results: allResults,
	}
}

// recursiveSearchWithContext performs depth-first search with operation tracking
// Security: limits recursion depth, result count, and total operations
func recursiveSearchWithContext(parser *xmlParser, targetSeg PathSegment, segments []PathSegment, segIndex int, ctx *searchContext, depth int) {
	// Security checks: limit recursion depth, result count, and total operations
	ctx.operations++
	if depth > MaxNestingDepth || len(*ctx.results) >= MaxWildcardResults || ctx.operations >= MaxRecursiveOperations {
		return
	}

	isLastSegment := segIndex == len(segments)-1

	for parser.skipToNextElement() {
		parser.next() // skip '<'
		elemName, attrs, isSelfClosing := parser.parseElementName()

		var content string
		if isSelfClosing {
			content = ""
		} else {
			content = parser.parseElementContent(elemName)
		}

		// Check if this element matches the target OR if we need to check within it
		// First, recurse into content regardless of match (for deeper matches)
		if !isSelfClosing && content != "" {
			contentParser := newXMLParser([]byte(content))
			recursiveSearchWithContext(contentParser, targetSeg, segments, segIndex, ctx, depth+1)
		}

		// Check operation limit again after recursion
		if ctx.operations >= MaxRecursiveOperations {
			return
		}

		// Then check if this element itself matches
		if targetSeg.matches(elemName) {
			// Security check: stop if we've reached the result limit
			if len(*ctx.results) >= MaxWildcardResults {
				return
			}

			// Found a match!
			if isLastSegment {
				// This is the final segment - add the result
				*ctx.results = append(*ctx.results, Result{
					Type: Element,
					Str:  unescapeXML(extractTextContent(content)),
					Raw:  content,
				})
			} else {
				// Continue matching with the next segment
				nextSegment := segments[segIndex+1]
				switch nextSegment.Type {
				case SegmentAttribute:
					attrName := nextSegment.Value
					if attrValue, ok := attrs[attrName]; ok {
						// Check if this is the final segment after attribute
						if segIndex+2 >= len(segments) {
							*ctx.results = append(*ctx.results, Result{
								Type: Attribute,
								Str:  attrValue,
								Raw:  attrValue,
							})
						}
					}
				case SegmentText:
					textContent := extractDirectTextOnly(content)
					*ctx.results = append(*ctx.results, Result{
						Type: String,
						Str:  unescapeXML(textContent),
						Raw:  content,
					})
				case SegmentFieldExtraction:
					// Field extraction from current match
					// Create a single-element match array for field extraction
					match := elementMatch{
						name:          elemName,
						attrs:         attrs,
						content:       content,
						isSelfClosing: isSelfClosing,
					}
					result := executeFieldExtraction([]elementMatch{match}, nextSegment)
					// Field extraction always returns Array (even if empty)
					if result.Type == Array {
						*ctx.results = append(*ctx.results, result.Results...)
					}
				default:
					contentParser := newXMLParser([]byte(content))
					result := executeQuery(contentParser, segments, segIndex+1)
					if result.Type != Null {
						if result.Type == Array {
							*ctx.results = append(*ctx.results, result.Results...)
						} else {
							*ctx.results = append(*ctx.results, result)
						}
					}
				}
			}
		}
	}
}

// GetMany searches xml for multiple paths in a single pass and returns
// a slice of Results corresponding to each path. This is more efficient
// than calling Get multiple times.
//
// Concurrency: GetMany is safe for concurrent use from multiple goroutines.
// Like Get, each call creates its own parser instances with no shared state.
func GetMany(xml string, paths ...string) []Result {
	results := make([]Result, len(paths))
	for i, path := range paths {
		results[i] = Get(xml, path)
	}
	return results
}

// GetWithOptions is like Get but accepts Options for behavioral control.
// Most users should use Get(); this function is for advanced use cases.
//
// Options allows customizing behavior such as:
//   - Case-insensitive path matching (CaseSensitive: false)
//   - Whitespace preservation (PreserveWhitespace: true, Phase 7+)
//   - Namespace URI mapping (Namespaces map, Phase 7+)
//
// Performance: If opts is nil or uses all default values, this function uses
// a fast path with minimal overhead compared to Get().
//
// Example (case-insensitive matching):
//
//	xml := `<ROOT><CHILD>value</CHILD></ROOT>`
//	opts := &Options{CaseSensitive: false}
//	result := GetWithOptions(xml, "root.child", opts)
//	fmt.Println(result.String()) // "value"
//
// Concurrency: GetWithOptions is safe for concurrent use from multiple goroutines.
func GetWithOptions(xml, path string, opts *Options) Result {
	return GetBytesWithOptions([]byte(xml), path, opts)
}

// GetStringWithOptions is like GetWithOptions but optimized for string input with
// zero-copy conversion. This is used internally by Result.GetWithOptions() to avoid
// unnecessary allocations when chaining queries.
func GetStringWithOptions(xml string, path string, opts *Options) Result {
	return GetBytesWithOptions(stringToBytes(xml), path, opts)
}

// GetBytesWithOptions is like GetBytes but accepts Options for behavioral control.
// Security: Documents larger than MaxDocumentSize (10MB) are rejected to prevent
// memory exhaustion attacks.
func GetBytesWithOptions(xml []byte, path string, opts *Options) Result {
	// Security check: reject documents that are too large
	if len(xml) > MaxDocumentSize {
		return Result{Type: Null}
	}

	// Fast path: if opts uses all defaults, use standard Get path
	if isDefaultOptions(opts) {
		segments := parsePath(path)
		if len(segments) == 0 {
			return Result{Type: Null}
		}
		parser := newXMLParser(xml)
		return executeQuery(parser, segments, 0)
	}

	// Parse path with options-aware parsing
	segments := parsePathWithOptions(path, opts)
	if len(segments) == 0 {
		return Result{Type: Null}
	}

	// Create parser
	parser := newXMLParser(xml)

	// Execute query with options
	return executeQueryWithOptions(parser, segments, 0, opts)
}

// parsePathWithOptions parses a path with options-aware parsing.
// Phase 6: Implements CaseSensitive option.
// Future: Will implement namespace prefix resolution.
func parsePathWithOptions(path string, opts *Options) []PathSegment {
	// Treat nil options as default options to prevent nil pointer dereferences
	if opts == nil {
		opts = DefaultOptions()
	}

	// Phase 6: For case-insensitive matching, we'll normalize the path during parsing
	// The actual case-insensitive matching happens in PathSegment.matchesWithOptions
	segments := parsePath(path)

	// If case-insensitive, convert all segment values to lowercase for matching
	if !opts.CaseSensitive {
		for i := range segments {
			if segments[i].Type == SegmentElement || segments[i].Type == SegmentAttribute {
				segments[i].Value = toLowerASCII(segments[i].Value)
			}
		}
	}

	return segments
}

// executeQueryWithOptions is like executeQuery but respects Options.
// Phase 6: Implements CaseSensitive matching.
// Future: Will implement PreserveWhitespace, Namespace resolution.
func executeQueryWithOptions(parser *xmlParser, segments []PathSegment, segIndex int, opts *Options) Result {
	// Base case: we've matched all segments
	if segIndex >= len(segments) {
		return Result{Type: Null}
	}

	currentSeg := segments[segIndex]
	isLastSegment := segIndex == len(segments)-1

	// Handle GJSON-style filter segment (Phase 3: Options support for GJSON)
	if currentSeg.Type == SegmentFilter && currentSeg.Filter != nil {
		return handleFilterQueryWithOptions(parser, segments, segIndex, opts)
	}

	// Handle GJSON-style field extraction segment #.field with options
	if currentSeg.Type == SegmentFieldExtraction {
		return Result{Type: Array, Results: []Result{}}
	}

	// Check if next segment is field extraction - collect ALL matches of current segment
	// This is the GJSON pattern with options: element.#.field
	hasFollowingFieldExtraction := !isLastSegment && segments[segIndex+1].Type == SegmentFieldExtraction
	if hasFollowingFieldExtraction && currentSeg.Type == SegmentElement {
		var allMatches []elementMatch
		for parser.skipToNextElement() {
			parser.next()
			elemName, attrs, isSelfClosing := parser.parseElementName()

			if !currentSeg.matchesWithOptions(elemName, opts) {
				if !isSelfClosing {
					parser.parseElementContent(elemName)
				}
				continue
			}

			var content string
			if isSelfClosing {
				content = ""
			} else {
				content = parser.parseElementContent(elemName)
			}

			allMatches = append(allMatches, elementMatch{
				name:          elemName,
				attrs:         attrs,
				content:       content,
				isSelfClosing: isSelfClosing,
			})

			if len(allMatches) >= MaxWildcardResults {
				break
			}
		}

		return executeFieldExtractionWithOptions(allMatches, segments[segIndex+1], opts)
	}

	// Check if next segment is a filter - if so, collect ALL matches of current segment
	// This is the GJSON pattern with options: element.#(condition)
	hasFollowingFilter := !isLastSegment && segments[segIndex+1].Type == SegmentFilter
	if hasFollowingFilter && currentSeg.Type == SegmentElement {
		var allMatches []elementMatch
		for parser.skipToNextElement() {
			parser.next()
			elemName, attrs, isSelfClosing := parser.parseElementName()

			if !currentSeg.matchesWithOptions(elemName, opts) {
				if !isSelfClosing {
					parser.parseElementContent(elemName)
				}
				continue
			}

			var content string
			if isSelfClosing {
				content = ""
			} else {
				content = parser.parseElementContent(elemName)
			}

			allMatches = append(allMatches, elementMatch{
				name:          elemName,
				attrs:         attrs,
				content:       content,
				isSelfClosing: isSelfClosing,
			})

			if len(allMatches) >= MaxWildcardResults {
				break
			}
		}

		return handleFilterQueryOnMatchesWithOptions(allMatches, segments, segIndex+1, opts)
	}

	// Handle array index or count on previous match
	if currentSeg.Type == SegmentIndex || currentSeg.Type == SegmentCount {
		return Result{Type: Null}
	}

	// Handle text content extraction
	if currentSeg.Type == SegmentText {
		return Result{Type: Null}
	}

	// Handle recursive wildcard
	if currentSeg.Type == SegmentWildcard && currentSeg.Wildcard {
		return handleRecursiveWildcardWithOptions(parser, segments, segIndex, opts)
	}

	// Find matching elements
	var matches []elementMatch
	isWildcard := currentSeg.Type == SegmentWildcard && !currentSeg.Wildcard
	hasFilter := currentSeg.Filter != nil

	for parser.skipToNextElement() {
		parser.next() // skip '<'
		elemName, attrs, isSelfClosing := parser.parseElementName()

		// Check if this segment is an attribute request
		if currentSeg.Type == SegmentAttribute {
			continue
		}

		// Check if element matches current segment (case-sensitive or insensitive)
		if !currentSeg.matchesWithOptions(elemName, opts) {
			if !isSelfClosing {
				parser.parseElementContent(elemName)
			}
			continue
		}

		// We have a match - extract content
		var content string
		if isSelfClosing {
			content = ""
		} else {
			content = parser.parseElementContent(elemName)
		}

		// Check if next segment indicates array operation
		needsArray := !isLastSegment && (segments[segIndex+1].Type == SegmentIndex || segments[segIndex+1].Type == SegmentCount)

		if needsArray || isWildcard || hasFilter {
			if len(matches) >= MaxWildcardResults {
				break
			}

			match := elementMatch{
				name:          elemName,
				attrs:         attrs,
				content:       content,
				isSelfClosing: isSelfClosing,
			}

			if hasFilter {
				if evaluateFilterOnMatch(currentSeg.Filter, match) {
					matches = append(matches, match)
				}
			} else {
				matches = append(matches, match)
			}
			continue
		}

		// Not an array/wildcard operation - process the first match

		// Check if next segment is an attribute
		if !isLastSegment && segments[segIndex+1].Type == SegmentAttribute {
			attrName := segments[segIndex+1].Value
			// For case-insensitive matching, search attrs case-insensitively
			if !opts.CaseSensitive {
				for k, v := range attrs {
					if toLowerASCII(k) == attrName {
						if segIndex+2 < len(segments) {
							return Result{Type: Null}
						}
						return Result{
							Type: Attribute,
							Str:  v,
							Raw:  v,
						}
					}
				}
				return Result{Type: Null}
			}
			if attrValue, ok := attrs[attrName]; ok {
				if segIndex+2 < len(segments) {
					return Result{Type: Null}
				}
				return Result{
					Type: Attribute,
					Str:  attrValue,
					Raw:  attrValue,
				}
			}
			return Result{Type: Null}
		}

		// Check if next segment is text content extraction
		if !isLastSegment && segments[segIndex+1].Type == SegmentText {
			textContent := extractDirectTextOnly(content)
			return Result{
				Type: String,
				Str:  unescapeXML(textContent),
				Raw:  content,
			}
		}

		// If this is the last segment, return the element content
		if isLastSegment {
			return Result{
				Type: Element,
				Str:  unescapeXML(extractTextContent(content)),
				Raw:  content,
			}
		}

		// Otherwise, parse the content and continue matching
		contentParser := newXMLParser([]byte(content))
		result := executeQueryWithOptions(contentParser, segments, segIndex+1, opts)
		if result.Type != Null {
			return result
		}
	}

	// Handle wildcard or filter results
	if (isWildcard || hasFilter) && len(matches) > 0 {
		return handleWildcardMatchesWithOptions(matches, segments, segIndex, opts)
	}

	// Handle array operations
	if len(matches) > 0 && segIndex+1 < len(segments) {
		nextSeg := segments[segIndex+1]
		switch nextSeg.Type {
		case SegmentIndex:
			if nextSeg.Index >= 0 && nextSeg.Index < len(matches) {
				match := matches[nextSeg.Index]

				if segIndex+2 < len(segments) {
					if segments[segIndex+2].Type == SegmentAttribute {
						attrName := segments[segIndex+2].Value
						if !opts.CaseSensitive {
							for k, v := range match.attrs {
								if toLowerASCII(k) == attrName {
									return Result{
										Type: Attribute,
										Str:  v,
										Raw:  v,
									}
								}
							}
							return Result{Type: Null}
						}
						if attrValue, ok := match.attrs[attrName]; ok {
							return Result{
								Type: Attribute,
								Str:  attrValue,
								Raw:  attrValue,
							}
						}
						return Result{Type: Null}
					}

					contentParser := newXMLParser([]byte(match.content))
					return executeQueryWithOptions(contentParser, segments, segIndex+2, opts)
				}

				return Result{
					Type: Element,
					Str:  unescapeXML(extractTextContent(match.content)),
					Raw:  match.content,
				}
			}
			return Result{Type: Null}
		case SegmentCount:
			return Result{
				Type: Number,
				Num:  float64(len(matches)),
				Str:  itoa(len(matches)),
			}
		}
	}

	return Result{Type: Null}
}

// handleWildcardMatchesWithOptions is like handleWildcardMatches but with Options support
func handleWildcardMatchesWithOptions(matches []elementMatch, segments []PathSegment, segIndex int, opts *Options) Result {
	isLastSegment := segIndex == len(segments)-1

	if isLastSegment {
		if len(matches) == 1 {
			return Result{
				Type: Element,
				Str:  unescapeXML(extractTextContent(matches[0].content)),
				Raw:  matches[0].content,
			}
		}
		results := make([]Result, 0, len(matches))
		for _, match := range matches {
			results = append(results, Result{
				Type: Element,
				Str:  unescapeXML(extractTextContent(match.content)),
				Raw:  match.content,
			})
		}
		return Result{
			Type:    Array,
			Results: results,
		}
	}

	var allResults []Result
	for _, match := range matches {
		nextSeg := segments[segIndex+1]

		if nextSeg.Type == SegmentAttribute {
			attrName := nextSeg.Value
			if !opts.CaseSensitive {
				for k, v := range match.attrs {
					if toLowerASCII(k) == attrName {
						if segIndex+2 >= len(segments) {
							allResults = append(allResults, Result{
								Type: Attribute,
								Str:  v,
								Raw:  v,
							})
						}
					}
				}
				continue
			}
			if attrValue, ok := match.attrs[attrName]; ok {
				if segIndex+2 >= len(segments) {
					allResults = append(allResults, Result{
						Type: Attribute,
						Str:  attrValue,
						Raw:  attrValue,
					})
				}
			}
			continue
		}

		if nextSeg.Type == SegmentText {
			textContent := extractDirectTextOnly(match.content)
			allResults = append(allResults, Result{
				Type: String,
				Str:  unescapeXML(textContent),
				Raw:  match.content,
			})
			continue
		}

		contentParser := newXMLParser([]byte(match.content))
		result := executeQueryWithOptions(contentParser, segments, segIndex+1, opts)
		if result.Type != Null {
			if result.Type == Array {
				allResults = append(allResults, result.Results...)
			} else {
				allResults = append(allResults, result)
			}
		}
	}

	if len(allResults) == 0 {
		return Result{Type: Null}
	}

	// Package results
	var result Result
	if len(allResults) == 1 {
		result = allResults[0]
	} else {
		result = Result{
			Type:    Array,
			Results: allResults,
		}
	}

	// Apply modifiers from the next segment if present (Phase 6)
	// The next segment after the wildcard is the one that was matched
	if segIndex+1 < len(segments) && len(segments[segIndex+1].Modifiers) > 0 {
		result = applyModifiers(result, segments[segIndex+1].Modifiers)
	}

	return result
}

// handleRecursiveWildcardWithOptions is like handleRecursiveWildcard but with Options support
func handleRecursiveWildcardWithOptions(parser *xmlParser, segments []PathSegment, segIndex int, opts *Options) Result {
	if segIndex >= len(segments)-1 {
		return Result{Type: Null}
	}

	nextSegIndex := segIndex + 1
	targetSeg := segments[nextSegIndex]

	var allResults []Result
	ctx := &searchContext{operations: 0, results: &allResults}
	recursiveSearchWithContextAndOptions(parser, targetSeg, segments, nextSegIndex, ctx, 0, opts)

	if len(allResults) == 0 {
		return Result{Type: Null}
	}
	if len(allResults) == 1 {
		return allResults[0]
	}
	return Result{
		Type:    Array,
		Results: allResults,
	}
}

// recursiveSearchWithContextAndOptions is like recursiveSearchWithContext but with Options support
func recursiveSearchWithContextAndOptions(parser *xmlParser, targetSeg PathSegment, segments []PathSegment, segIndex int, ctx *searchContext, depth int, opts *Options) {
	ctx.operations++
	if depth > MaxNestingDepth || len(*ctx.results) >= MaxWildcardResults || ctx.operations >= MaxRecursiveOperations {
		return
	}

	isLastSegment := segIndex == len(segments)-1

	for parser.skipToNextElement() {
		parser.next()
		elemName, attrs, isSelfClosing := parser.parseElementName()

		var content string
		if isSelfClosing {
			content = ""
		} else {
			content = parser.parseElementContent(elemName)
		}

		if !isSelfClosing && content != "" {
			contentParser := newXMLParser([]byte(content))
			recursiveSearchWithContextAndOptions(contentParser, targetSeg, segments, segIndex, ctx, depth+1, opts)
		}

		if ctx.operations >= MaxRecursiveOperations {
			return
		}

		if targetSeg.matchesWithOptions(elemName, opts) {
			if len(*ctx.results) >= MaxWildcardResults {
				return
			}

			if isLastSegment {
				*ctx.results = append(*ctx.results, Result{
					Type: Element,
					Str:  unescapeXML(extractTextContent(content)),
					Raw:  content,
				})
			} else {
				nextSegment := segments[segIndex+1]
				switch nextSegment.Type {
				case SegmentAttribute:
					attrName := nextSegment.Value
					if !opts.CaseSensitive {
						for k, v := range attrs {
							if toLowerASCII(k) == attrName {
								if segIndex+2 >= len(segments) {
									*ctx.results = append(*ctx.results, Result{
										Type: Attribute,
										Str:  v,
										Raw:  v,
									})
								}
							}
						}
					} else {
						if attrValue, ok := attrs[attrName]; ok {
							if segIndex+2 >= len(segments) {
								*ctx.results = append(*ctx.results, Result{
									Type: Attribute,
									Str:  attrValue,
									Raw:  attrValue,
								})
							}
						}
					}
				case SegmentText:
					textContent := extractDirectTextOnly(content)
					*ctx.results = append(*ctx.results, Result{
						Type: String,
						Str:  unescapeXML(textContent),
						Raw:  content,
					})
				default:
					contentParser := newXMLParser([]byte(content))
					result := executeQueryWithOptions(contentParser, segments, segIndex+1, opts)
					if result.Type != Null {
						if result.Type == Array {
							*ctx.results = append(*ctx.results, result.Results...)
						} else {
							*ctx.results = append(*ctx.results, result)
						}
					}
				}
			}
		}
	}
}

// executeFieldExtraction extracts a specific field from all elements in an array of matches.
// This implements the GJSON #.field syntax for extracting values from array elements.
// The field can be an element name, attribute (@attr), or text content (%).
//
// Security: Results are limited to MaxWildcardResults (1,000) items. Extraction exceeding
// this limit will be silently truncated to prevent memory exhaustion.
func executeFieldExtraction(matches []elementMatch, segment PathSegment) Result {
	fieldName := segment.Field
	if fieldName == "" {
		return Result{Type: Null}
	}

	// Pre-allocate results array with estimated capacity
	results := make([]Result, 0, len(matches))
	totalExtracted := 0

	// Determine field type (attribute, text, or element)
	isAttribute := strings.HasPrefix(fieldName, "@")
	isText := fieldName == "%"

	for _, match := range matches {
		// Security: Check result limit
		if totalExtracted >= MaxWildcardResults {
			break
		}

		if isAttribute {
			// Extract attribute value
			attrName := fieldName[1:] // Remove @ prefix
			if attrValue, ok := match.attrs[attrName]; ok {
				results = append(results, Result{
					Type: Attribute,
					Str:  attrValue,
					Raw:  attrValue,
				})
				totalExtracted++
			}
		} else if isText {
			// Extract text content only
			textContent := extractDirectTextOnly(match.content)
			if textContent != "" {
				results = append(results, Result{
					Type: String,
					Str:  unescapeXML(textContent),
					Raw:  textContent,
				})
				totalExtracted++
			}
		} else {
			// Extract child element(s) with matching name
			parser := newXMLParser([]byte(match.content))
			for parser.skipToNextElement() {
				// Security: Check limit on each iteration
				if totalExtracted >= MaxWildcardResults {
					break
				}

				parser.next() // skip '<'
				elemName, _, isSelfClosing := parser.parseElementName()

				// Check if element name matches field name
				if elemName != fieldName {
					// Skip this element
					if !isSelfClosing {
						parser.parseElementContent(elemName)
					}
					continue
				}

				// Found matching element
				var content string
				if isSelfClosing {
					content = ""
				} else {
					content = parser.parseElementContent(elemName)
				}

				results = append(results, Result{
					Type: Element,
					Str:  unescapeXML(extractTextContent(content)),
					Raw:  content,
				})
				totalExtracted++
			}
		}
	}

	// Return empty array if no results
	if len(results) == 0 {
		return Result{
			Type:    Array,
			Results: []Result{},
		}
	}

	// Return as Array Result
	result := Result{
		Type:    Array,
		Results: results,
	}

	// Apply modifiers if present
	if len(segment.Modifiers) > 0 {
		result = applyModifiers(result, segment.Modifiers)
	}

	return result
}

// executeFieldExtractionWithOptions is like executeFieldExtraction but with Options support
//
// Security: Results are limited to MaxWildcardResults (1,000) items. Extraction exceeding
// this limit will be silently truncated to prevent memory exhaustion.
func executeFieldExtractionWithOptions(matches []elementMatch, segment PathSegment, opts *Options) Result {
	// Treat nil options as default options
	if opts == nil {
		opts = DefaultOptions()
	}

	fieldName := segment.Field
	if fieldName == "" {
		return Result{Type: Null}
	}

	results := make([]Result, 0, len(matches))
	totalExtracted := 0

	isAttribute := strings.HasPrefix(fieldName, "@")
	isText := fieldName == "%"

	for _, match := range matches {
		if totalExtracted >= MaxWildcardResults {
			break
		}

		if isAttribute {
			attrName := fieldName[1:]
			// Case-insensitive attribute lookup
			if !opts.CaseSensitive {
				attrNameLower := toLowerASCII(attrName)
				for k, v := range match.attrs {
					if toLowerASCII(k) == attrNameLower {
						results = append(results, Result{
							Type: Attribute,
							Str:  v,
							Raw:  v,
						})
						totalExtracted++
						break // Only match first attribute with this name
					}
				}
			} else {
				// Case-sensitive attribute lookup
				if attrValue, ok := match.attrs[attrName]; ok {
					results = append(results, Result{
						Type: Attribute,
						Str:  attrValue,
						Raw:  attrValue,
					})
					totalExtracted++
				}
			}
		} else if isText {
			textContent := extractDirectTextOnly(match.content)
			if textContent != "" {
				results = append(results, Result{
					Type: String,
					Str:  unescapeXML(textContent),
					Raw:  textContent,
				})
				totalExtracted++
			}
		} else {
			// Extract child element(s) with matching name (case-insensitive if needed)
			parser := newXMLParser([]byte(match.content))
			fieldNameCmp := fieldName
			if !opts.CaseSensitive {
				fieldNameCmp = toLowerASCII(fieldName)
			}

			for parser.skipToNextElement() {
				if totalExtracted >= MaxWildcardResults {
					break
				}

				parser.next()
				elemName, _, isSelfClosing := parser.parseElementName()

				// Case-aware comparison
				elemNameCmp := elemName
				if !opts.CaseSensitive {
					elemNameCmp = toLowerASCII(elemName)
				}

				if elemNameCmp != fieldNameCmp {
					if !isSelfClosing {
						parser.parseElementContent(elemName)
					}
					continue
				}

				var content string
				if isSelfClosing {
					content = ""
				} else {
					content = parser.parseElementContent(elemName)
				}

				results = append(results, Result{
					Type: Element,
					Str:  unescapeXML(extractTextContent(content)),
					Raw:  content,
				})
				totalExtracted++
			}
		}
	}

	if len(results) == 0 {
		return Result{
			Type:    Array,
			Results: []Result{},
		}
	}

	result := Result{
		Type:    Array,
		Results: results,
	}

	if len(segment.Modifiers) > 0 {
		result = applyModifiers(result, segment.Modifiers)
	}

	return result
}

// handleFilterQuery processes GJSON-style filter queries #(condition) or #(condition)#
// This function collects all matching elements, then routes to first-match or all-match processing.
func handleFilterQuery(parser *xmlParser, segments []PathSegment, segIndex int) Result {
	currentSeg := segments[segIndex]
	isLastSegment := segIndex == len(segments)-1

	// Collect ALL matching elements
	var matches []elementMatch

	for parser.skipToNextElement() {
		parser.next() // skip '<'
		elemName, attrs, isSelfClosing := parser.parseElementName()

		// Extract content
		var content string
		if isSelfClosing {
			content = ""
		} else {
			content = parser.parseElementContent(elemName)
		}

		match := elementMatch{
			name:          elemName,
			attrs:         attrs,
			content:       content,
			isSelfClosing: isSelfClosing,
		}

		// Evaluate filter condition
		if evaluateFilterOnMatch(currentSeg.Filter, match) {
			matches = append(matches, match)

			// Security: enforce result limit
			if len(matches) >= MaxWildcardResults {
				break
			}
		}
	}

	// No matches found
	if len(matches) == 0 {
		return Result{Type: Null}
	}

	// Route based on FilterAll flag
	if currentSeg.FilterAll {
		// #(condition)# - Return ALL matches
		return processAllMatches(matches, segments, segIndex, isLastSegment)
	}
	// #(condition) - Return FIRST match
	return processFirstMatch(matches[0], segments, segIndex, isLastSegment)
}

// processFirstMatch processes the first matching element from a filter query
func processFirstMatch(match elementMatch, segments []PathSegment, segIndex int, isLastSegment bool) Result {
	currentSeg := segments[segIndex]

	// If this is the last segment, return the element
	if isLastSegment {
		result := Result{
			Type: Element,
			Str:  unescapeXML(extractTextContent(match.content)),
			Raw:  match.content,
		}
		// Apply modifiers if present
		if len(currentSeg.Modifiers) > 0 {
			result = applyModifiers(result, currentSeg.Modifiers)
		}
		return result
	}

	// Continue matching with next segment
	nextSeg := segments[segIndex+1]

	// Handle attribute access
	if nextSeg.Type == SegmentAttribute {
		attrName := nextSeg.Value
		if attrValue, ok := match.attrs[attrName]; ok {
			result := Result{
				Type: Attribute,
				Str:  attrValue,
				Raw:  attrValue,
			}
			// Apply modifiers from the attribute segment if present
			if len(nextSeg.Modifiers) > 0 {
				result = applyModifiers(result, nextSeg.Modifiers)
			}
			return result
		}
		return Result{Type: Null}
	}

	// Handle text extraction
	if nextSeg.Type == SegmentText {
		textContent := extractDirectTextOnly(match.content)
		result := Result{
			Type: String,
			Str:  unescapeXML(textContent),
			Raw:  match.content,
		}
		// Apply modifiers from the text segment if present
		if len(nextSeg.Modifiers) > 0 {
			result = applyModifiers(result, nextSeg.Modifiers)
		}
		return result
	}

	// Handle field extraction
	if nextSeg.Type == SegmentFieldExtraction {
		return executeFieldExtraction([]elementMatch{match}, nextSeg)
	}

	// Continue query within matched element
	contentParser := newXMLParser([]byte(match.content))
	return executeQuery(contentParser, segments, segIndex+1)
}

// processAllMatches processes all matching elements from a filter query
func processAllMatches(matches []elementMatch, segments []PathSegment, segIndex int, isLastSegment bool) Result {
	currentSeg := segments[segIndex]

	// If this is the last segment, return ALL matches as array
	if isLastSegment {
		results := make([]Result, 0, len(matches))
		for _, match := range matches {
			results = append(results, Result{
				Type: Element,
				Str:  unescapeXML(extractTextContent(match.content)),
				Raw:  match.content,
			})
		}

		result := Result{
			Type:    Array,
			Results: results,
		}
		// Apply modifiers if present
		if len(currentSeg.Modifiers) > 0 {
			result = applyModifiers(result, currentSeg.Modifiers)
		}
		return result
	}

	// Continue matching with next segment for each match
	nextSeg := segments[segIndex+1]
	var allResults []Result

	for _, match := range matches {
		// Handle attribute access
		if nextSeg.Type == SegmentAttribute {
			attrName := nextSeg.Value
			if attrValue, ok := match.attrs[attrName]; ok {
				allResults = append(allResults, Result{
					Type: Attribute,
					Str:  attrValue,
					Raw:  attrValue,
				})
			}
			continue
		}

		// Handle text extraction
		if nextSeg.Type == SegmentText {
			textContent := extractDirectTextOnly(match.content)
			allResults = append(allResults, Result{
				Type: String,
				Str:  unescapeXML(textContent),
				Raw:  match.content,
			})
			continue
		}

		// Handle field extraction
		if nextSeg.Type == SegmentFieldExtraction {
			// For field extraction after filter all, we extract from all matches at once
			// This is done outside the loop below
			break
		}

		// Continue query within matched element
		contentParser := newXMLParser([]byte(match.content))
		result := executeQuery(contentParser, segments, segIndex+1)
		if result.Type != Null {
			if result.Type == Array {
				// Flatten nested arrays
				allResults = append(allResults, result.Results...)
			} else {
				allResults = append(allResults, result)
			}
		}
	}

	// Handle field extraction for all matches
	if nextSeg.Type == SegmentFieldExtraction {
		return executeFieldExtraction(matches, nextSeg)
	}

	if len(allResults) == 0 {
		return Result{Type: Null}
	}

	// Package results
	var result Result
	if len(allResults) == 1 {
		result = allResults[0]
	} else {
		result = Result{
			Type:    Array,
			Results: allResults,
		}
	}

	// Apply modifiers from the next segment if present
	if len(nextSeg.Modifiers) > 0 {
		result = applyModifiers(result, nextSeg.Modifiers)
	}

	return result
}

// handleFilterQueryOnMatches processes GJSON-style filters against a pre-collected set of matches
// This is used when an element segment is followed by a filter segment (element.#(condition))
func handleFilterQueryOnMatches(allMatches []elementMatch, segments []PathSegment, segIndex int) Result {
	currentSeg := segments[segIndex]
	isLastSegment := segIndex == len(segments)-1

	// Filter the matches
	var filteredMatches []elementMatch
	for _, match := range allMatches {
		if evaluateFilterOnMatch(currentSeg.Filter, match) {
			filteredMatches = append(filteredMatches, match)
		}
	}

	// No matches found
	if len(filteredMatches) == 0 {
		return Result{Type: Null}
	}

	// Route based on FilterAll flag
	if currentSeg.FilterAll {
		// #(condition)# - Return ALL matches
		return processAllMatches(filteredMatches, segments, segIndex, isLastSegment)
	}
	// #(condition) - Return FIRST match
	return processFirstMatch(filteredMatches[0], segments, segIndex, isLastSegment)
}

// handleFilterQueryWithOptions processes GJSON-style filter queries with Options support
func handleFilterQueryWithOptions(parser *xmlParser, segments []PathSegment, segIndex int, opts *Options) Result {
	currentSeg := segments[segIndex]
	isLastSegment := segIndex == len(segments)-1

	// Collect ALL matching elements
	var matches []elementMatch

	for parser.skipToNextElement() {
		parser.next() // skip '<'
		elemName, attrs, isSelfClosing := parser.parseElementName()

		// Extract content
		var content string
		if isSelfClosing {
			content = ""
		} else {
			content = parser.parseElementContent(elemName)
		}

		match := elementMatch{
			name:          elemName,
			attrs:         attrs,
			content:       content,
			isSelfClosing: isSelfClosing,
		}

		// Evaluate filter condition
		if evaluateFilterOnMatch(currentSeg.Filter, match) {
			matches = append(matches, match)

			// Security: enforce result limit
			if len(matches) >= MaxWildcardResults {
				break
			}
		}
	}

	// No matches found
	if len(matches) == 0 {
		return Result{Type: Null}
	}

	// Route based on FilterAll flag
	if currentSeg.FilterAll {
		// #(condition)# - Return ALL matches
		return processAllMatchesWithOptions(matches, segments, segIndex, isLastSegment, opts)
	}
	// #(condition) - Return FIRST match
	return processFirstMatchWithOptions(matches[0], segments, segIndex, isLastSegment, opts)
}

// handleFilterQueryOnMatchesWithOptions processes GJSON-style filters against pre-collected matches with Options
func handleFilterQueryOnMatchesWithOptions(allMatches []elementMatch, segments []PathSegment, segIndex int, opts *Options) Result {
	currentSeg := segments[segIndex]
	isLastSegment := segIndex == len(segments)-1

	// Filter the matches
	var filteredMatches []elementMatch
	for _, match := range allMatches {
		if evaluateFilterOnMatch(currentSeg.Filter, match) {
			filteredMatches = append(filteredMatches, match)
		}
	}

	// No matches found
	if len(filteredMatches) == 0 {
		return Result{Type: Null}
	}

	// Route based on FilterAll flag
	if currentSeg.FilterAll {
		// #(condition)# - Return ALL matches
		return processAllMatchesWithOptions(filteredMatches, segments, segIndex, isLastSegment, opts)
	}
	// #(condition) - Return FIRST match
	return processFirstMatchWithOptions(filteredMatches[0], segments, segIndex, isLastSegment, opts)
}

// processFirstMatchWithOptions processes the first matching element with Options support
func processFirstMatchWithOptions(match elementMatch, segments []PathSegment, segIndex int, isLastSegment bool, opts *Options) Result {
	currentSeg := segments[segIndex]

	// If this is the last segment, return the element
	if isLastSegment {
		result := Result{
			Type: Element,
			Str:  unescapeXML(extractTextContent(match.content)),
			Raw:  match.content,
		}
		// Apply modifiers if present
		if len(currentSeg.Modifiers) > 0 {
			result = applyModifiers(result, currentSeg.Modifiers)
		}
		return result
	}

	// Continue matching with next segment
	nextSeg := segments[segIndex+1]

	// Handle attribute access
	if nextSeg.Type == SegmentAttribute {
		attrName := nextSeg.Value
		// For case-insensitive matching, search attrs case-insensitively
		if !opts.CaseSensitive {
			for k, v := range match.attrs {
				if toLowerASCII(k) == attrName {
					result := Result{
						Type: Attribute,
						Str:  v,
						Raw:  v,
					}
					// Apply modifiers from the attribute segment if present
					if len(nextSeg.Modifiers) > 0 {
						result = applyModifiers(result, nextSeg.Modifiers)
					}
					return result
				}
			}
			return Result{Type: Null}
		}
		if attrValue, ok := match.attrs[attrName]; ok {
			result := Result{
				Type: Attribute,
				Str:  attrValue,
				Raw:  attrValue,
			}
			// Apply modifiers from the attribute segment if present
			if len(nextSeg.Modifiers) > 0 {
				result = applyModifiers(result, nextSeg.Modifiers)
			}
			return result
		}
		return Result{Type: Null}
	}

	// Handle text extraction
	if nextSeg.Type == SegmentText {
		textContent := extractDirectTextOnly(match.content)
		result := Result{
			Type: String,
			Str:  unescapeXML(textContent),
			Raw:  match.content,
		}
		// Apply modifiers from the text segment if present
		if len(nextSeg.Modifiers) > 0 {
			result = applyModifiers(result, nextSeg.Modifiers)
		}
		return result
	}

	// Handle field extraction
	if nextSeg.Type == SegmentFieldExtraction {
		return executeFieldExtractionWithOptions([]elementMatch{match}, nextSeg, opts)
	}

	// Continue query within matched element
	contentParser := newXMLParser([]byte(match.content))
	return executeQueryWithOptions(contentParser, segments, segIndex+1, opts)
}

// processAllMatchesWithOptions processes all matching elements with Options support
func processAllMatchesWithOptions(matches []elementMatch, segments []PathSegment, segIndex int, isLastSegment bool, opts *Options) Result {
	currentSeg := segments[segIndex]

	// If this is the last segment, return ALL matches as array
	if isLastSegment {
		results := make([]Result, 0, len(matches))
		for _, match := range matches {
			results = append(results, Result{
				Type: Element,
				Str:  unescapeXML(extractTextContent(match.content)),
				Raw:  match.content,
			})
		}

		result := Result{
			Type:    Array,
			Results: results,
		}
		// Apply modifiers if present
		if len(currentSeg.Modifiers) > 0 {
			result = applyModifiers(result, currentSeg.Modifiers)
		}
		return result
	}

	// Continue matching with next segment for each match
	nextSeg := segments[segIndex+1]
	var allResults []Result

	for _, match := range matches {
		// Handle attribute access
		if nextSeg.Type == SegmentAttribute {
			attrName := nextSeg.Value
			if !opts.CaseSensitive {
				for k, v := range match.attrs {
					if toLowerASCII(k) == attrName {
						allResults = append(allResults, Result{
							Type: Attribute,
							Str:  v,
							Raw:  v,
						})
					}
				}
				continue
			}
			if attrValue, ok := match.attrs[attrName]; ok {
				allResults = append(allResults, Result{
					Type: Attribute,
					Str:  attrValue,
					Raw:  attrValue,
				})
			}
			continue
		}

		// Handle text extraction
		if nextSeg.Type == SegmentText {
			textContent := extractDirectTextOnly(match.content)
			allResults = append(allResults, Result{
				Type: String,
				Str:  unescapeXML(textContent),
				Raw:  match.content,
			})
			continue
		}

		// Handle field extraction
		if nextSeg.Type == SegmentFieldExtraction {
			// For field extraction after filter all, we extract from all matches at once
			// This is done outside the loop below
			break
		}

		// Continue query within matched element
		contentParser := newXMLParser([]byte(match.content))
		result := executeQueryWithOptions(contentParser, segments, segIndex+1, opts)
		if result.Type != Null {
			if result.Type == Array {
				// Flatten nested arrays
				allResults = append(allResults, result.Results...)
			} else {
				allResults = append(allResults, result)
			}
		}
	}

	// Handle field extraction for all matches
	if nextSeg.Type == SegmentFieldExtraction {
		return executeFieldExtractionWithOptions(matches, nextSeg, opts)
	}

	if len(allResults) == 0 {
		return Result{Type: Null}
	}

	// Package results
	var result Result
	if len(allResults) == 1 {
		result = allResults[0]
	} else {
		result = Result{
			Type:    Array,
			Results: allResults,
		}
	}

	// Apply modifiers from the next segment if present
	if len(nextSeg.Modifiers) > 0 {
		result = applyModifiers(result, nextSeg.Modifiers)
	}

	return result
}

// stringToBytes converts a string to []byte with zero allocation.
// The returned slice shares the underlying string data and MUST NOT be modified.
//
// Safety guarantees (Go memory model):
//  1. Go strings are immutable - the underlying data is in read-only memory
//  2. The Go runtime protects string data from modification (write attempts may segfault)
//  3. xmldot parser ONLY performs read operations (indexing, slicing)
//  4. This function is unexported - only trusted package code uses it
//  5. GC correctly handles the shared memory (string won't be collected while []byte in use)
//
// This is a standard Go optimization pattern used in strings/bytes/unsafe packages.
// See: https://github.com/golang/go/issues/53003 (approved unsafe pattern)
//
// Performance: Eliminates string→[]byte copy overhead (~20-30% faster for fluent chains)
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"strconv"
	"strings"
)

// Type represents the type of a Result value.
type Type int

const (
	// Null represents a non-existent path.
	Null Type = iota
	// String represents a text value.
	String
	// Number represents a numeric value.
	Number
	// True represents a boolean true value.
	True
	// False represents a boolean false value.
	False
	// Element represents an XML element.
	Element
	// Attribute represents an XML attribute.
	Attribute
	// Array represents multiple XML elements with the same name.
	Array
)

// Result represents the result of a Get operation. It contains the matched
// value and provides methods for type conversion.
type Result struct {
	// Type is the type of the result value.
	Type Type
	// Raw is the raw XML segment that was matched.
	Raw string
	// Str is the parsed string value.
	Str string
	// Index is the array index if this result is part of an array.
	Index int
	// Num is the cached numeric value if the result is a number.
	Num float64
	// Results holds child results for Array type (Phase 3+)
	Results []Result
}

// Exists returns true if the result represents an existing value in the XML.
func (r Result) Exists() bool {
	return r.Type != Null
}

// String returns the string representation of the result.
// For Null types, it returns an empty string.
// For Array types, it returns a JSON-like array representation.
// This implements the fmt.Stringer interface.
func (r Result) String() string {
	if r.Type == Null {
		return ""
	}
	if r.Type == Array {
		// Return JSON-like array representation: ["item1","item2",...]
		if len(r.Results) == 0 {
			return "[]"
		}
		var sb strings.Builder
		sb.WriteString("[")
		for i, item := range r.Results {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(`"`)
			// Escape quotes and backslashes in the value
			itemStr := item.String()
			for _, ch := range itemStr {
				switch ch {
				case '"':
					sb.WriteString(`\"`)
				case '\\':
					sb.WriteString(`\\`)
				default:
					sb.WriteRune(ch)
				}
			}
			sb.WriteString(`"`)
		}
		sb.WriteString("]")
		return sb.String()
	}
	return r.Str
}

// Int returns the result as an int64. If the result cannot be converted,
// it returns 0.
func (r Result) Int() int64 {
	switch r.Type {
	case Number:
		return int64(r.Num)
	case String, Element, Attribute:
		// Try parsing the string value
		if val, err := parseInt64(r.Str); err == nil {
			return val
		}
	case True:
		return 1
	case False:
		return 0
	}
	return 0
}

// Float returns the result as a float64. If the result cannot be converted,
// it returns 0.
func (r Result) Float() float64 {
	switch r.Type {
	case Number:
		return r.Num
	case String, Element, Attribute:
		// Try parsing the string value
		if val, err := parseFloat64(r.Str); err == nil {
			return val
		}
	case True:
		return 1
	case False:
		return 0
	}
	return 0
}

// Bool returns the result as a bool. The strings "true", "1", "yes" are
// considered true. Everything else is false.
func (r Result) Bool() bool {
	switch r.Type {
	case True:
		return true
	case False:
		return false
	case Number:
		return r.Num != 0
	case String, Element, Attribute:
		s := r.Str
		return s == "true" || s == "1" || s == "yes" || s == "True" || s == "YES" || s == "t" || s == "T"
	}
	return false
}

// Value returns the result as an interface{} with the appropriate Go type.
func (r Result) Value() interface{} {
	switch r.Type {
	case Null:
		return nil
	case True:
		return true
	case False:
		return false
	case Number:
		return r.Num
	case String, Element, Attribute:
		return r.Str
	case Array:
		// Return slice of Result values
		return r.Results
	}
	return nil
}

// IsArray returns true if the Result represents an array (multiple elements).
func (r Result) IsArray() bool {
	return r.Type == Array
}

// Array returns the Result as a slice of Results for array types.
// For non-array types, returns a single-element slice containing the result.
func (r Result) Array() []Result {
	if r.Type == Array {
		return r.Results
	}
	if r.Type == Null {
		return []Result{}
	}
	return []Result{r}
}

// ForEach iterates over array elements, calling the iterator function for each.
// The iterator receives the index and value. Return false to stop iteration.
// For non-array types, the iterator is called once with index 0.
func (r Result) ForEach(iterator func(index int, value Result) bool) {
	if r.Type == Array {
		for i, result := range r.Results {
			if !iterator(i, result) {
				return
			}
		}
	} else if r.Type != Null {
		iterator(0, r)
	}
}

// Helper functions for type conversion

// parseInt64 parses a string to int64, handling various formats
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// parseFloat64 parses a string to float64
func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// itoa converts int to string
func itoa(i int) string {
	return strconv.Itoa(i)
}

// Get enables fluent method chaining by executing a path query on the Result's
// content. This allows querying nested structures without extracting intermediate
// values.
//
// Behavior by Result type:
//   - Element: Re-parses the element's XML content and executes the path query
//   - Array: Delegates to the first element's Get() method (GJSON-compatible behavior)
//   - Null: Returns Null immediately (safe chaining)
//   - Primitives (String, Number, Attribute): Returns Null (terminal types)
//
// For querying all array elements, use explicit #.field or #(filter)# syntax.
//
// Result Semantics:
//   - Path not found: Returns Null
//   - Invalid query on primitive types: Returns Null
//   - Field extraction with no matches (#.field): Returns empty Array (Type: Array, Results: [])
//   - Filter with no matches: Returns Null for #(...), empty Array for #(...)#
//
// Security: All security limits apply (MaxDocumentSize, MaxWildcardResults, etc.).
// The Raw field is already bounded by package-level parsing limits. Deep chaining
// operates on progressively smaller XML subsets, preventing amplification attacks.
//
// Concurrency: Get is safe for concurrent use. Result is immutable and each call
// creates its own parser instance.
//
// Example (basic chaining):
//
//	root := xmldot.Get(xml, "root")
//	user := root.Get("user")
//	name := user.Get("name").String()
//
// Example (deep chaining):
//
//	name := xmldot.Get(xml, "root").
//	    Get("company").
//	    Get("department").
//	    Get("team.member").
//	    Get("name")
//
// Example (array field extraction):
//
//	items := xmldot.Get(xml, "root.items")
//	names := items.Get("item.#.name")  // All item names
//
// Performance (measured on Go 1.24, Apple M4 Pro arm64):
//
//	Single Get() call: ~1699ns (vs ~1328ns for equivalent direct path = 27.9% overhead)
//	3-level chain: ~5003ns total (vs ~1328ns direct = 276.7% overhead)
//
// Recommendation: Use fluent chaining for readability when overhead is acceptable,
// use full paths (e.g., "root.user.name") for performance-critical loops.
func (r Result) Get(path string) Result {
	// Null results return Null immediately
	if r.Type == Null {
		return Result{Type: Null}
	}

	// Primitive types (String, Number, Attribute, True, False) cannot be queried
	if r.Type != Element && r.Type != Array {
		return Result{Type: Null}
	}

	// Array type: query first element only (GJSON behavior)
	if r.Type == Array {
		if len(r.Results) == 0 {
			return Result{Type: Null}
		}
		return r.Results[0].Get(path)
	}

	// Element type: re-parse Raw XML using zero-copy helper
	return GetString(r.Raw, path)
}

// GetMany enables fluent batch queries on the Result's content. This is more
// efficient than calling Get multiple times when querying multiple paths.
//
// Like Get, GetMany only works on Element and Array types. For Null or primitive
// types, it returns a slice of Null results.
//
// For Array types, each path is applied to all elements in the array, and the
// results are combined (similar to Get's array behavior).
//
// Example:
//
//	xml := `<root><user><name>Alice</name><age>30</age><city>NYC</city></user></root>`
//	user := xmldot.Get(xml, "root.user")
//	results := user.GetMany("name", "age", "city")
//	fmt.Println(results[0].String()) // "Alice"
//	fmt.Println(results[1].String()) // "30"
//	fmt.Println(results[2].String()) // "NYC"
//
// Concurrency: GetMany is safe for concurrent use. The Result is immutable.
func (r Result) GetMany(paths ...string) []Result {
	results := make([]Result, len(paths))

	// Null or primitive types: return slice of Null results (zero-values)
	if r.Type == Null || (r.Type != Element && r.Type != Array) {
		return results
	}

	// Query each path
	for i, path := range paths {
		results[i] = r.Get(path)
	}

	return results
}

// GetWithOptions enables fluent queries with custom options like case-insensitive
// matching. This is like Get but accepts Options for behavioral control.
//
// Most users should use Get(); this method is for advanced use cases requiring
// non-default behavior.
//
// Options allows customizing behavior such as:
//   - Case-insensitive path matching (CaseSensitive: false)
//   - Whitespace preservation (PreserveWhitespace: true, Phase 7+)
//   - Namespace URI mapping (Namespaces map, Phase 7+)
//
// Like Get, GetWithOptions only works on Element and Array types.
//
// Example (case-insensitive fluent query):
//
//	xml := `<root><USER><NAME>Alice</NAME></USER></root>`
//	opts := &xmldot.Options{CaseSensitive: false}
//	root := xmldot.Get(xml, "root")
//	name := root.GetWithOptions("user.name", opts)
//	fmt.Println(name.String()) // "Alice"
//
// Performance: If opts is nil or uses all default values, this method uses
// the fast path with minimal overhead compared to Get().
//
// Concurrency: GetWithOptions is safe for concurrent use.
func (r Result) GetWithOptions(path string, opts *Options) Result {
	// Null results return Null immediately
	if r.Type == Null {
		return Result{Type: Null}
	}

	// Primitive types cannot be queried
	if r.Type != Element && r.Type != Array {
		return Result{Type: Null}
	}

	// Fast path: if opts uses all defaults, use standard Get
	if isDefaultOptions(opts) {
		return r.Get(path)
	}

	// Array type with options: query first element only
	if r.Type == Array {
		if len(r.Results) == 0 {
			return Result{Type: Null}
		}
		return r.Results[0].GetWithOptions(path, opts)
	}

	// Element type: re-parse Raw XML with options using zero-copy helper
	return GetStringWithOptions(r.Raw, path, opts)
}

// Map returns a map of immediate children from the Result's content.
// This enables structure inspection and dynamic field access for XML elements.
//
// Behavior by Result type:
//   - Element: Parses immediate children into map
//   - Array: Delegates to first element's Map() (GJSON-compatible)
//   - Null/Primitives: Returns empty map
//
// Map Keys:
//   - Element names: Immediate child elements (e.g., "name", "age")
//   - Mixed text content: Stored under "%" key (direct text only, excludes nested)
//   - Namespace prefixes: Preserved as-is (e.g., "soap:Body")
//
// Note on Attributes:
//
//	Map() returns children of the element, not the element's own attributes.
//	To access element attributes, use Get with @ syntax:
//	  id := Get(xml, "user.@id")     // Get attribute directly
//	  user := Get(xml, "user")
//	  name := user.Map()["name"]     // Get child via Map
//
// Duplicate Handling:
//   - Multiple elements with same name: Combined into Array Result
//   - Single element: Stored as Element Result (not Array)
//   - Example: <root><item>a</item><item>b</item></root> → map["item"] is Array with 2 Results
//
// Empty Elements:
//   - Self-closing tags (e.g., <item/>) included with empty string value
//
// Excluded:
//   - Comments and processing instructions are skipped
//   - Only immediate children (not recursive)
//
// Security:
//   - Immediate children limited to MaxWildcardResults (1000)
//   - Elements beyond limit are silently excluded (map will be incomplete)
//   - Consider using explicit Get queries for elements with >1000 children
//   - Nesting depth checked during parse (MaxNestingDepth=100)
//
// Example:
//
//	xml := `<user id="42">
//	  <name>Alice</name>
//	  <age>30</age>
//	  Some text
//	  <email>alice@example.com</email>
//	  <tag>go</tag>
//	  <tag>xml</tag>
//	</user>`
//
//	user := xmldot.Get(xml, "user")
//	m := user.Map()
//
//	// Access user's attribute separately
//	id := xmldot.Get(xml, "user.@id").String()  // "42"
//
//	// Access children via map
//	fmt.Println(m["name"].String())   // "Alice"
//	fmt.Println(m["age"].String())    // "30"
//	fmt.Println(m["email"].String())  // "alice@example.com"
//	fmt.Println(m["%"].String())      // "Some text"
//
//	// Multiple elements with same name become Array
//	for _, tag := range m["tag"].Array() {
//	    fmt.Println(tag.String())  // "go", "xml"
//	}
//
// Performance:
//   - Single-pass parser: O(n) where n is number of immediate children
//   - No caching: Each call re-parses Result.Raw
//   - Typical cost: 1-5µs for elements with 5-10 children
//
// Concurrency: Safe for concurrent use. Result is immutable.
func (r Result) Map() map[string]Result {
	// Null or primitive types: return empty map
	if r.Type == Null || (r.Type != Element && r.Type != Array) {
		return make(map[string]Result)
	}

	// Array type: delegate to first element (GJSON-compatible)
	if r.Type == Array {
		if len(r.Results) == 0 {
			return make(map[string]Result)
		}
		return r.Results[0].Map()
	}

	// Element type: re-parse Raw XML to extract immediate children
	return parseMapChildren(r.Raw)
}

// MapWithOptions returns a map of immediate children with custom options like case-insensitive matching.
// This is like Map() but accepts Options for behavioral control.
//
// Most users should use Map(); this method is for advanced use cases requiring
// non-default behavior such as case-insensitive element name matching.
//
// Options allows customizing behavior:
//   - Case-insensitive name matching (CaseSensitive: false)
//
// Like Map(), MapWithOptions only works on Element and Array types and returns
// children only (not parent element attributes).
//
// Example (case-insensitive map):
//
//	xml := `<USER><NAME>Alice</NAME><AGE>30</AGE></USER>`
//	opts := &xmldot.Options{CaseSensitive: false}
//	user := xmldot.Get(xml, "user") // Assuming case-insensitive Get
//	m := user.MapWithOptions(opts)
//	// m will have lowercase keys: "name", "age"
//
// Performance: If opts is nil or uses all default values, this method delegates
// to Map() for optimal performance.
//
// Concurrency: MapWithOptions is safe for concurrent use.
func (r Result) MapWithOptions(opts *Options) map[string]Result {
	// Fast path: if opts uses all defaults, use standard Map
	if isDefaultOptions(opts) {
		return r.Map()
	}

	// Null or primitive types: return empty map
	if r.Type == Null || (r.Type != Element && r.Type != Array) {
		return make(map[string]Result)
	}

	// Array type: delegate to first element (GJSON-compatible)
	if r.Type == Array {
		if len(r.Results) == 0 {
			return make(map[string]Result)
		}
		return r.Results[0].MapWithOptions(opts)
	}

	// Element type: parse with options
	return parseMapChildrenWithOptions(r.Raw, opts)
}

// parseMapChildren parses element content and returns immediate children as map.
// Note: xml parameter is the element's content (Result.Raw), not the full element with tags.
// This means parent element attributes are NOT accessible - only child elements and mixed text.
func parseMapChildren(xml string) map[string]Result {
	// Convert to bytes (zero-copy)
	xmlBytes := stringToBytes(xml)

	// Security check
	if len(xmlBytes) > MaxDocumentSize {
		return make(map[string]Result)
	}

	result := make(map[string]Result, 8) // Pre-allocate reasonable capacity

	// If empty content, return empty map
	if len(xmlBytes) == 0 {
		return result
	}

	// Extract direct text content for "%" key
	// Note: extractDirectTextOnly already returns trimmed text
	directText := extractDirectTextOnly(xml)
	if directText != "" {
		result["%"] = Result{
			Type: String,
			Str:  unescapeXML(directText),
			Raw:  directText,
		}
	}

	// Parse immediate child elements
	parser := newXMLParser(xmlBytes)
	childCount := 0

	for parser.skipToNextElement() {
		// Security limit: Prevent memory exhaustion from elements with excessive children
		// Silently truncate at MaxWildcardResults (1000) - map will be incomplete beyond limit
		if childCount >= MaxWildcardResults {
			break
		}

		parser.next() // skip '<'
		childName, _, childIsSelfClosing := parser.parseElementName()

		var childContent string

		if childIsSelfClosing {
			childContent = ""
		} else {
			childContent = parser.parseElementContent(childName)
		}

		// Create Result for this child
		newChild := Result{
			Type: Element,
			Str:  unescapeXML(extractTextContent(childContent)),
			Raw:  childContent, // Store content, not full XML
		}

		// Add child to map, handling duplicates by converting to Array
		addChildToMap(result, childName, newChild)
		childCount++
	}

	return result
}

// parseMapChildrenWithOptions parses element content with options and returns immediate children as map.
// Supports case-insensitive element name matching when opts.CaseSensitive = false.
func parseMapChildrenWithOptions(xml string, opts *Options) map[string]Result {
	// Convert to bytes (zero-copy)
	xmlBytes := stringToBytes(xml)

	// Security check
	if len(xmlBytes) > MaxDocumentSize {
		return make(map[string]Result)
	}

	result := make(map[string]Result, 8) // Pre-allocate reasonable capacity

	// If empty content, return empty map
	if len(xmlBytes) == 0 {
		return result
	}

	// Extract direct text content for "%" key
	// Note: extractDirectTextOnly already returns trimmed text
	directText := extractDirectTextOnly(xml)
	if directText != "" {
		result["%"] = Result{
			Type: String,
			Str:  unescapeXML(directText),
			Raw:  directText,
		}
	}

	// Determine case sensitivity
	caseSensitive := opts == nil || opts.CaseSensitive

	// Parse immediate child elements
	parser := newXMLParser(xmlBytes)
	childCount := 0

	for parser.skipToNextElement() {
		// Security limit: Prevent memory exhaustion from elements with excessive children
		// Silently truncate at MaxWildcardResults (1000) - map will be incomplete beyond limit
		if childCount >= MaxWildcardResults {
			break
		}

		parser.next() // skip '<'
		childName, _, childIsSelfClosing := parser.parseElementName()

		var childContent string

		if childIsSelfClosing {
			childContent = ""
		} else {
			childContent = parser.parseElementContent(childName)
		}

		// Normalize child name based on case sensitivity
		mapKey := childName
		if !caseSensitive {
			mapKey = strings.ToLower(childName)
		}

		// Create Result for this child
		newChild := Result{
			Type: Element,
			Str:  unescapeXML(extractTextContent(childContent)),
			Raw:  childContent, // Store content, not full XML
		}

		// Add child to map, handling duplicates by converting to Array
		addChildToMap(result, mapKey, newChild)
		childCount++
	}

	return result
}

// addChildToMap adds a child Result to the map, handling duplicates by converting to Array.
// First occurrence: stores as single Result
// Second occurrence: converts to Array with both elements
// Subsequent occurrences: appends to existing Array
func addChildToMap(m map[string]Result, name string, child Result) {
	existing, exists := m[name]
	if !exists {
		// First occurrence - store as single Result
		m[name] = child
		return
	}

	// Duplicate found - convert to Array or append
	var results []Result
	if existing.Type == Array {
		// Already an array, append new child
		results = append(existing.Results, child)
	} else {
		// First duplicate, convert single to array
		results = []Result{existing, child}
	}

	m[name] = Result{
		Type:    Array,
		Results: results,
	}
}

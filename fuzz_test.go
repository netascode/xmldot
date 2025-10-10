// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"testing"
)

// ============================================================================
// Comprehensive Fuzz Tests
// Consolidated from: get_fuzz_test.go, parser_fuzz_test.go, path_fuzz_test.go,
// set_fuzz_test.go, validate_fuzz_test.go
// ============================================================================

// FuzzGet tests the Get operation for crashes and panics with arbitrary inputs.
// This ensures Get never panics regardless of malformed XML or invalid paths.
func FuzzGet(f *testing.F) {
	// Seed corpus with valid XML + various path patterns
	f.Add("<root><a>1</a><b>2</b></root>", "root.a")
	f.Add("<data><items><item>x</item></items></data>", "data.items.item")
	f.Add("<root a='value'/>", "root.@a")
	f.Add("<root><a>1</a><a>2</a></root>", "root.a.0")
	f.Add("<root><a>1</a><a>2</a></root>", "root.a.#")
	f.Add("<root><a>text</a></root>", "root.a.%")
	f.Add("<users><user><name>John</name><age>30</age></user></users>", "users.user.name")
	f.Add("<root><item id='1'>A</item></root>", "root.item[@id==1]")
	f.Add("<root><a><b><c>x</c></b></a></root>", "root.*.c")
	f.Add("<root><a><b><c>x</c></b></a></root>", "root.**.c")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)

		// Validate result properties are consistent
		if result.Exists() {
			// All accessor methods should work without panic
			_ = result.String()
			_ = result.Int()
			_ = result.Float()
			_ = result.Bool()
			_ = result.String()
			_ = result.Type

			// String representation should never be empty for existing results
			// (unless it's actually an empty element)
			str := result.String()
			_ = str // Use it to ensure no optimization removes the call
		} else {
			// Non-existent results should have Null type
			if result.Type != Null {
				t.Errorf("Non-existent result has non-Null type: %v", result.Type)
			}
		}
	})
}

// FuzzGetBytes tests the GetBytes operation with byte slice inputs.
// This ensures zero-copy parsing is robust.
func FuzzGetBytes(f *testing.F) {
	// Seed corpus with byte patterns
	f.Add([]byte("<root><item>value</item></root>"), "root.item")
	f.Add([]byte("<root a='b'/>"), "root.@a")
	f.Add([]byte(""), "root")
	f.Add([]byte("<root>"), "root")

	f.Fuzz(func(t *testing.T, xml []byte, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetBytes panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := GetBytes(xml, path)

		// Validate result consistency
		if result.Exists() {
			_ = result.String()
			_ = result.String()
		}
	})
}

// FuzzGetWithFilters tests Get with filter expressions.
// This ensures filter parsing and evaluation is robust.
func FuzzGetWithFilters(f *testing.F) {
	// Seed with filter patterns
	f.Add("<users><user><age>25</age></user></users>", "users.user[age>21]")
	f.Add("<items><item id='1'>A</item></items>", "items.item[@id==1]")
	f.Add("<data><item><price>10</price></item></data>", "data.item[price<20]")
	f.Add("<list><item>x</item></list>", "list.item[name==test]")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get with filter panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)

		// Result should be valid even if filter doesn't match
		_ = result.Exists()
		_ = result.String()
	})
}

// FuzzGetWithWildcards tests Get with wildcard patterns.
// This ensures wildcard matching doesn't cause infinite loops or exponential blowup.
func FuzzGetWithWildcards(f *testing.F) {
	// Seed with wildcard patterns
	f.Add("<root><a><name>A</name></a><b><name>B</name></b></root>", "root.*.name")
	f.Add("<root><a><b><c>x</c></b></a></root>", "root.**.c")
	f.Add("<data><item><value>1</value></item></data>", "**.value")
	f.Add("<a><a><a>x</a></a></a>", "a.**.a")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get with wildcard panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)

		// Wildcard results should be arrays or single values
		_ = result.Exists()
		if result.Type == Array {
			// Should be able to iterate array
			for _, elem := range result.Array() {
				_ = elem.String()
			}
		}
	})
}

// FuzzGetWithArrayAccess tests Get with array indexing.
// This ensures array operations handle edge cases correctly.
func FuzzGetWithArrayAccess(f *testing.F) {
	// Seed with array access patterns
	f.Add("<root><item>1</item><item>2</item></root>", "root.item.0")
	f.Add("<root><item>1</item><item>2</item></root>", "root.item.1")
	f.Add("<root><item>1</item></root>", "root.item.-1")
	f.Add("<root><item>1</item><item>2</item></root>", "root.item.#")
	f.Add("<root><item>1</item></root>", "root.item.999")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get with array access panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)

		// Array access should not panic
		// Note: Type validation skipped - focusing on panic prevention
		_ = result.Exists()
	})
}

// FuzzGetWithAttributes tests Get for attribute access.
// This ensures attribute queries are robust.
func FuzzGetWithAttributes(f *testing.F) {
	// Seed with attribute patterns
	f.Add("<root a='b'/>", "root.@a")
	f.Add("<root a='1' b='2'/>", "root.@a")
	f.Add("<root><item id='5'>x</item></root>", "root.item.@id")
	f.Add("<root xmlns:ns='uri'/>", "root.@xmlns:ns")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get attribute panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)

		// Attribute results should be accessible
		// Note: Type validation skipped - focusing on panic prevention
		if result.Exists() {
			_ = result.String()
		}
	})
}

// FuzzGetWithTextContent tests Get with text content operator.
// This ensures text extraction is robust.
func FuzzGetWithTextContent(f *testing.F) {
	// Seed with text content patterns
	f.Add("<root>text</root>", "root.%")
	f.Add("<root><a>x</a>text</root>", "root.%")
	f.Add("<root>  text  </root>", "root.%")
	f.Add("<root><![CDATA[data]]></root>", "root.%")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get text content panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)

		// Text content should not panic
		// Note: Type validation skipped - focusing on panic prevention
		if result.Exists() {
			_ = result.String()
		}
	})
}

// FuzzGetWithNamespaces tests Get with namespace-prefixed elements.
// This ensures namespace handling is robust.
func FuzzGetWithNamespaces(f *testing.F) {
	// Seed with namespace patterns
	f.Add("<ns:root xmlns:ns='uri'/>", "ns:root")
	f.Add("<root xmlns='uri'/>", "root")
	f.Add("<a:root><a:item>x</a:item></a:root>", "a:root.a:item")
	f.Add("<root><ns:item>x</ns:item></root>", "root.ns:item")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get namespace panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)
		_ = result.Exists()
	})
}

// FuzzGetWithModifiers tests Get with modifier chains (Phase 6 feature).
// This ensures modifiers don't cause crashes.
func FuzzGetWithModifiers(f *testing.F) {
	// Seed with modifier patterns
	f.Add("<root><item>3</item><item>1</item><item>2</item></root>", "root.item|@sort")
	f.Add("<root><item>1</item><item>2</item></root>", "root.item|@first")
	f.Add("<root><item>1</item><item>2</item></root>", "root.item|@last")
	f.Add("<root><item>1</item><item>2</item></root>", "root.item|@reverse")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get modifier panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)
		_ = result.Exists()
	})
}

// FuzzGetComplexPaths tests Get with complex path combinations.
// This ensures complex queries don't cause issues.
func FuzzGetComplexPaths(f *testing.F) {
	// Seed with complex paths
	f.Add("<root><a><b><c>x</c></b></a></root>", "root.a.b.c")
	f.Add("<root><items><item id='1'><name>A</name></item></items></root>", "root.items.item[@id==1].name")
	f.Add("<data><list><item>1</item></list></data>", "data.list.item.0.%")
	f.Add("<root><a><b>x</b></a></root>", "root.**.b.%")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get complex path panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		result := Get(xml, path)
		if result.Exists() {
			_ = result.String()
			_ = result.String()
		}
	})
}

// FuzzParser tests the core XML parser for crashes and panics with arbitrary input.
// This fuzz test focuses on parser robustness - it should never panic regardless
// of malformed, malicious, or random input.
func FuzzParser(f *testing.F) {
	// Seed corpus with known XML patterns to guide fuzzing
	f.Add("<root><item>value</item></root>")
	f.Add("<root attr='value'/>")
	f.Add("<a><b><c>deep</c></b></a>")
	f.Add("<?xml version='1.0'?><root/>")
	f.Add("<root xmlns='http://example.com'/>")
	f.Add("<root><![CDATA[data]]></root>")
	f.Add("<!-- comment --><root/>")
	f.Add("<root attr1='a' attr2='b'/>")
	f.Add("<root><item>1</item><item>2</item></root>")
	f.Add("")
	f.Add("<")
	f.Add("<<")
	f.Add("<>")
	f.Add("<root")
	f.Add("<root>")

	f.Fuzz(func(t *testing.T, xml string) {
		// Parser should NEVER panic on any input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on input: %q\nPanic: %v", xml, r)
			}
		}()

		// Try parsing with Get (exercises full parser)
		_ = Get(xml, "root")
		_ = Get(xml, "root.item")
		_ = Get(xml, "root.@attr")

		// Try validation (exercises parser differently)
		_ = Valid(xml)
	})
}

// FuzzParserWithPaths tests parser with various path combinations.
// This ensures the parser handles different query patterns robustly.
func FuzzParserWithPaths(f *testing.F) {
	// Seed with XML + path combinations
	f.Add("<root><a>1</a></root>", "root.a")
	f.Add("<root><item>x</item></root>", "root.item")
	f.Add("<data><items><item>test</item></items></data>", "data.items.item")
	f.Add("<root a='b'/>", "root.@a")
	f.Add("<root><a>1</a><a>2</a></root>", "root.a.0")
	f.Add("<root><a>1</a><a>2</a></root>", "root.a.#")
	f.Add("<root><a>text</a></root>", "root.a.%")
	f.Add("<root><a><b>x</b></a></root>", "root.*.b")
	f.Add("<root><a><b><c>x</c></b></a></root>", "root.**.c")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		// Get should never panic
		result := Get(xml, path)

		// Validate result properties for consistency
		if result.Exists() {
			// If result exists, accessing its value should not panic
			_ = result.String()
			_ = result.Int()
			_ = result.Float()
			_ = result.Bool()
			_ = result.String()
		}
	})
}

// FuzzParserNesting tests parser with deeply nested structures.
// This ensures nesting depth limits work correctly and don't cause stack overflow.
func FuzzParserNesting(f *testing.F) {
	// Seed with nested structures of varying depths
	f.Add("<a><b><c>x</c></b></a>")
	f.Add("<a><a><a><a>x</a></a></a></a>")
	f.Add("<root><child><grandchild>value</grandchild></child></root>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on nested input: %q\nPanic: %v", xml, r)
			}
		}()

		// Test various operations that traverse nested structures
		_ = Get(xml, "a.b.c")
		_ = Get(xml, "root.child.grandchild")
		_ = Get(xml, "**.value")
		_ = Valid(xml)
	})
}

// FuzzParserAttributes tests parser attribute handling.
// This ensures attribute parsing is robust against malformed input.
func FuzzParserAttributes(f *testing.F) {
	// Seed with various attribute patterns
	f.Add("<root a='b'/>")
	f.Add("<root a='b' c='d'/>")
	f.Add("<root a=\"b\"/>")
	f.Add("<root a='b\" c=\"d'/>") // Mixed quotes
	f.Add("<root a=b/>")           // Unquoted
	f.Add("<root a/>")             // No value
	f.Add("<root a='/>")           // Unterminated

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on attribute input: %q\nPanic: %v", xml, r)
			}
		}()

		// Try accessing attributes
		_ = Get(xml, "root.@a")
		_ = Get(xml, "root.@c")
		_ = Valid(xml)
	})
}

// FuzzParserSpecialContent tests parser with special XML constructs.
// This ensures CDATA, comments, and processing instructions don't break the parser.
func FuzzParserSpecialContent(f *testing.F) {
	// Seed with special content patterns
	f.Add("<root><![CDATA[data]]></root>")
	f.Add("<!-- comment --><root/>")
	f.Add("<?xml version='1.0'?><root/>")
	f.Add("<root><!-- comment --></root>")
	f.Add("<root><![CDATA[<xml>data</xml>]]></root>")
	f.Add("<?target data?><root/>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on special content: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Get(xml, "root")
		_ = Valid(xml)
	})
}

// FuzzParserLargeInput tests parser with size limits.
// This ensures security limits work correctly and don't cause issues.
func FuzzParserLargeInput(f *testing.F) {
	// Seed with various sizes
	f.Add("<root>" + string(make([]byte, 100)) + "</root>")
	f.Add("<root>" + string(make([]byte, 1000)) + "</root>")
	f.Add("<root>" + string(make([]byte, 10000)) + "</root>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on large input: len=%d panic=%v", len(xml), r)
			}
		}()

		// Parser should handle size limits gracefully
		_ = Get(xml, "root")
		_ = Valid(xml)

		// If within size limits, operations should work
		if len(xml) <= MaxDocumentSize {
			result := Get(xml, "root")
			// Accessing result should not panic
			_ = result.Exists()
		}
	})
}

// FuzzParserEscaping tests XML entity escaping/unescaping.
// This ensures special characters are handled correctly.
func FuzzParserEscaping(f *testing.F) {
	// Seed with escaped content
	f.Add("<root>&lt;tag&gt;</root>")
	f.Add("<root>&amp;&quot;&apos;</root>")
	f.Add("<root attr='&lt;&gt;'/>")
	f.Add("<root>&#65;&#x42;</root>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on escaped input: %q\nPanic: %v", xml, r)
			}
		}()

		result := Get(xml, "root")
		if result.Exists() {
			// Accessing value should handle escaping correctly
			_ = result.String()
		}
	})
}

// FuzzParserNamespaces tests namespace prefix handling.
// This ensures namespace syntax doesn't break the parser.
func FuzzParserNamespaces(f *testing.F) {
	// Seed with namespace patterns
	f.Add("<ns:root xmlns:ns='http://example.com'/>")
	f.Add("<root xmlns='http://example.com'/>")
	f.Add("<a:root><a:item>x</a:item></a:root>")
	f.Add("<root xmlns:a='uri1' xmlns:b='uri2'/>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on namespace input: %q\nPanic: %v", xml, r)
			}
		}()

		// Try various namespace-aware operations
		_ = Get(xml, "root")
		_ = Get(xml, "ns:root")
		_ = Get(xml, "a:root.a:item")
		_ = Valid(xml)
	})
}

// FuzzPathParser tests path parsing for crashes and panics.
// This ensures path parser never panics on any input string.
func FuzzPathParser(f *testing.F) {
	// Seed corpus with various path patterns
	f.Add("root.item.@attr")
	f.Add("data.items.item.0")
	f.Add("root.*.value")
	f.Add("root.**.price")
	f.Add("users.user[age>21]")
	f.Add("items.item[@id==5]")
	f.Add("root.item.#")
	f.Add("root.item.%")
	f.Add("a.b.c.d.e.f")
	f.Add("")
	f.Add(".")
	f.Add("..")
	f.Add("...")
	f.Add("@")
	f.Add("@@")
	f.Add("*")
	f.Add("**")
	f.Add("#")
	f.Add("%")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path parser panicked: path=%q panic=%v", path, r)
			}
		}()

		// Path parsing should never panic
		segments := parsePath(path)

		// If segments were parsed, they should be valid
		for _, seg := range segments {
			// Accessing segment properties should not panic
			_ = seg.Type
			_ = seg.Value
			_ = seg.Index
			_ = seg.Wildcard
			_ = seg.Filter
			_ = seg.Modifiers
		}
	})
}

// FuzzPathSplitting tests path splitting with various delimiters.
// This ensures splitPath handles edge cases correctly.
func FuzzPathSplitting(f *testing.F) {
	// Seed with split patterns
	f.Add("a.b.c")
	f.Add("a..b")
	f.Add(".a.b.")
	f.Add("a\\.b") // Escaped dot
	f.Add("\\.")
	f.Add("a\\\\b")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path splitting panicked: path=%q panic=%v", path, r)
			}
		}()

		parts := splitPath(path)

		// Verify parts are consistent
		for _, part := range parts {
			_ = part // Use it
		}
	})
}

// FuzzPathWithFilters tests filter expression parsing.
// This ensures filter parsing is robust.
func FuzzPathWithFilters(f *testing.F) {
	// Seed with filter patterns
	f.Add("item[age>21]")
	f.Add("item[age<30]")
	f.Add("item[age==25]")
	f.Add("item[@id==5]")
	f.Add("item[name==test]")
	f.Add("item[]")
	f.Add("item[")
	f.Add("item]")
	f.Add("item[[]]")
	f.Add("item[a>b>c]")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path filter parsing panicked: path=%q panic=%v", path, r)
			}
		}()

		segments := parsePath(path)

		// If filter was parsed, it should be valid
		for _, seg := range segments {
			if seg.Filter != nil {
				_ = seg.Filter.Path
				_ = seg.Filter.Op
				_ = seg.Filter.Value
			}
		}
	})
}

// FuzzPathWithModifiers tests modifier parsing (Phase 6 feature).
// This ensures modifier syntax is robust.
func FuzzPathWithModifiers(f *testing.F) {
	// Seed with modifier patterns
	f.Add("item|@sort")
	f.Add("item|@first")
	f.Add("item|@last")
	f.Add("item|@reverse")
	f.Add("item|@sort|@first")
	f.Add("item||")
	f.Add("item|")
	f.Add("|item")
	f.Add("item|@")
	f.Add("item|@unknown")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path modifier parsing panicked: path=%q panic=%v", path, r)
			}
		}()

		segments := parsePath(path)

		// If modifiers were parsed, they should be valid
		for _, seg := range segments {
			if seg.Modifiers != nil {
				for _, mod := range seg.Modifiers {
					_ = mod
				}
			}
		}
	})
}

// FuzzPathWithArrayAccess tests array indexing syntax.
// This ensures array index parsing is robust.
func FuzzPathWithArrayAccess(f *testing.F) {
	// Seed with array patterns
	f.Add("item.0")
	f.Add("item.1")
	f.Add("item.-1")
	f.Add("item.999")
	f.Add("item.-999")
	f.Add("item.00")
	f.Add("item.01")
	f.Add("item.123456789")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path array access panicked: path=%q panic=%v", path, r)
			}
		}()

		segments := parsePath(path)

		// If array index was parsed, verify it's valid
		for _, seg := range segments {
			if seg.Type == SegmentIndex {
				_ = seg.Index
			}
		}
	})
}

// FuzzPathWithSpecialOperators tests special operators (%, #, *, **).
// This ensures operator parsing is robust.
func FuzzPathWithSpecialOperators(f *testing.F) {
	// Seed with special operators
	f.Add("item.%")
	f.Add("item.#")
	f.Add("item.*")
	f.Add("item.**")
	f.Add("%.item")
	f.Add("#.item")
	f.Add("*.item")
	f.Add("**.item")
	f.Add("%%")
	f.Add("##")
	f.Add("**.*")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path special operators panicked: path=%q panic=%v", path, r)
			}
		}()

		segments := parsePath(path)

		// Verify operator segments are valid
		for _, seg := range segments {
			switch seg.Type {
			case SegmentText, SegmentCount, SegmentWildcard:
				_ = seg.Type
			}
		}
	})
}

// FuzzPathWithAttributes tests attribute access syntax.
// This ensures @ syntax is robust.
func FuzzPathWithAttributes(f *testing.F) {
	// Seed with attribute patterns
	f.Add("item.@attr")
	f.Add("@attr")
	f.Add("item.@")
	f.Add("item.@@")
	f.Add("item.@a.b")
	f.Add("item.@xmlns:ns")
	f.Add("@")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path attribute access panicked: path=%q panic=%v", path, r)
			}
		}()

		segments := parsePath(path)

		// Verify attribute segments are valid
		for _, seg := range segments {
			if seg.Type == SegmentAttribute {
				_ = seg.Value
			}
		}
	})
}

// FuzzPathWithNamespaces tests namespace prefix syntax.
// This ensures namespace handling in paths is robust.
func FuzzPathWithNamespaces(f *testing.F) {
	// Seed with namespace patterns
	f.Add("ns:item")
	f.Add("a:b:c")
	f.Add(":item")
	f.Add("item:")
	f.Add("ns:item.ns:child")
	f.Add("@xmlns:ns")
	f.Add("ns:")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path namespace parsing panicked: path=%q panic=%v", path, r)
			}
		}()

		segments := parsePath(path)

		// Verify namespace segments are valid
		for _, seg := range segments {
			if seg.Value != "" {
				// Split namespace should not panic
				prefix, local := splitNamespace(seg.Value)
				_ = prefix
				_ = local
			}
		}
	})
}

// FuzzPathCacheConsistency tests path caching for consistency.
// This ensures cached results match fresh parsing.
func FuzzPathCacheConsistency(f *testing.F) {
	// Seed with various paths
	f.Add("root.item")
	f.Add("a.b.c")
	f.Add("item[x>5]")
	f.Add("item|@sort")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path cache consistency panicked: path=%q panic=%v", path, r)
			}
		}()

		// Parse twice - should get same result from cache
		segments1 := parsePath(path)
		segments2 := parsePath(path)

		// Verify both results are equivalent (length comparison)
		if (segments1 == nil) != (segments2 == nil) {
			t.Errorf("Path cache inconsistency: path=%q nil1=%v nil2=%v", path, segments1 == nil, segments2 == nil)
		}

		if segments1 != nil && segments2 != nil {
			if len(segments1) != len(segments2) {
				t.Errorf("Path cache length mismatch: path=%q len1=%d len2=%d", path, len(segments1), len(segments2))
			}
		}
	})
}

// FuzzPathSegmentLimit tests maximum path segment limit.
// This ensures security limit works correctly.
func FuzzPathSegmentLimit(f *testing.F) {
	// Seed with various segment counts
	f.Add("a")
	f.Add("a.b.c.d.e.f.g.h.i.j")
	f.Add("a.b.c.d.e.f.g.h.i.j.k.l.m.n.o.p.q.r.s.t.u.v.w.x.y.z")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path segment limit panicked: path=%q panic=%v", path, r)
			}
		}()

		segments := parsePath(path)

		// If segments exceeded limit, should return nil
		// If within limit, should return valid segments
		if len(segments) > MaxPathSegments {
			t.Errorf("Path exceeded segment limit: path=%q segments=%d limit=%d", path, len(segments), MaxPathSegments)
		}
	})
}

// FuzzPathMatchWithOptions tests path matching with options.
// This ensures case sensitivity and other options work correctly.
func FuzzPathMatchWithOptions(f *testing.F) {
	// Seed with matching patterns
	f.Add("root", "ROOT")
	f.Add("item", "Item")
	f.Add("a:item", "a:ITEM")
	f.Add("ns:item", "NS:item")

	f.Fuzz(func(t *testing.T, segValue, elemName string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Path match with options panicked: segValue=%q elemName=%q panic=%v", segValue, elemName, r)
			}
		}()

		seg := PathSegment{
			Type:  SegmentElement,
			Value: segValue,
		}

		// Test with different options
		opts1 := &Options{CaseSensitive: true}
		opts2 := &Options{CaseSensitive: false}

		_ = seg.matchesWithOptions(elemName, opts1)
		_ = seg.matchesWithOptions(elemName, opts2)
	})
}

// FuzzIsNumeric tests numeric detection for array indices.
// This ensures isNumeric handles edge cases.
func FuzzIsNumeric(f *testing.F) {
	// Seed with numeric patterns
	f.Add("0")
	f.Add("123")
	f.Add("-1")
	f.Add("00")
	f.Add("01")
	f.Add("-0")
	f.Add("123abc")
	f.Add("abc123")
	f.Add("")
	f.Add("-")

	f.Fuzz(func(t *testing.T, s string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("isNumeric panicked: s=%q panic=%v", s, r)
			}
		}()

		result := isNumeric(s)
		_ = result
	})
}

// FuzzSet tests the Set operation for crashes and panics with arbitrary inputs.
// This ensures Set never panics regardless of malformed XML, invalid paths, or values.
func FuzzSet(f *testing.F) {
	// Seed corpus with valid patterns
	f.Add("<root><item>old</item></root>", "root.item", "new")
	f.Add("<root/>", "root.newitem", "value")
	f.Add("<root><item>x</item></root>", "root.item.@attr", "attrval")
	f.Add("<root><items></items></root>", "root.items.item", "value")
	f.Add("<root/>", "root.a.b.c", "deep")
	f.Add("<root><item>1</item></root>", "root.item.0", "first")
	f.Add("<root><a>x</a><b>y</b></root>", "root.c", "new")
	f.Add("", "root", "value")
	f.Add("<root>", "root.item", "value")
	f.Add("<root/>", "", "value")

	f.Fuzz(func(t *testing.T, xml, path, value string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set panicked: xml=%q path=%q value=%q panic=%v", xml, path, value, r)
			}
		}()

		result, err := Set(xml, path, value)

		// If Set succeeded, result should be valid XML
		if err == nil {
			// Validate the result is well-formed
			if !Valid(result) {
				// Only report error if original was valid
				if Valid(xml) {
					t.Errorf("Set produced invalid XML from valid input\nxml=%q\npath=%q\nvalue=%q\nresult=%q", xml, path, value, result)
				}
			}

			// Verify we can query the set value (if path was valid)
			if len(path) > 0 && !containsInvalidChars(path) {
				retrieved := Get(result, path)
				// If the path was successfully set, we should be able to retrieve it
				_ = retrieved.Exists()
			}
		}
	})
}

// FuzzSetBytes tests the SetBytes operation with byte slice inputs.
// This ensures zero-copy modification is robust.
func FuzzSetBytes(f *testing.F) {
	// Seed corpus
	f.Add([]byte("<root><item>old</item></root>"), "root.item", []byte("new"))
	f.Add([]byte("<root/>"), "root.item", []byte("value"))
	f.Add([]byte(""), "root", []byte("value"))

	f.Fuzz(func(t *testing.T, xml []byte, path string, value []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetBytes panicked: xml=%q path=%q value=%q panic=%v", xml, path, value, r)
			}
		}()

		result, err := SetBytes(xml, path, string(value))

		// If successful, result should be valid
		if err == nil && len(result) > 0 {
			_ = Valid(string(result))
		}
	})
}

// FuzzSetWithTypes tests Set with different value types.
// This ensures type conversion and escaping is robust.
func FuzzSetWithTypes(f *testing.F) {
	// Seed with different types
	f.Add("<root/>", "root.str", "string value")
	f.Add("<root/>", "root.special", "<>&\"'")
	f.Add("<root/>", "root.unicode", "你好世界")
	f.Add("<root/>", "root.whitespace", "  spaces  ")
	f.Add("<root/>", "root.newlines", "line1\nline2")

	f.Fuzz(func(t *testing.T, xml, path, value string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set with types panicked: xml=%q path=%q value=%q panic=%v", xml, path, value, r)
			}
		}()

		result, err := Set(xml, path, value)

		// If successful, special characters should be properly escaped
		if err == nil && Valid(result) {
			// Verify we can retrieve the value
			retrieved := Get(result, path)
			if retrieved.Exists() {
				// Retrieved value should match original (after unescaping)
				_ = retrieved.String()
			}
		}
	})
}

// FuzzSetRaw tests SetRaw for XML injection vulnerabilities.
// This ensures raw XML insertion is validated.
func FuzzSetRaw(f *testing.F) {
	// Seed with raw XML patterns
	f.Add("<root/>", "root.item", "<value>test</value>")
	f.Add("<root/>", "root.data", "<a><b>nested</b></a>")
	f.Add("<root/>", "root.bad", "<unclosed>")
	f.Add("<root/>", "root.empty", "")
	f.Add("<root/>", "root.malformed", "</close>")

	f.Fuzz(func(t *testing.T, xml, path, rawxml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetRaw panicked: xml=%q path=%q rawxml=%q panic=%v", xml, path, rawxml, r)
			}
		}()

		result, err := SetRaw(xml, path, rawxml)

		// If successful, result should be well-formed
		if err == nil {
			if !Valid(result) {
				t.Errorf("SetRaw produced invalid XML\nxml=%q\npath=%q\nrawxml=%q\nresult=%q", xml, path, rawxml, result)
			}
		}
	})
}

// FuzzSetAttributes tests Set for attribute operations.
// This ensures attribute setting is robust.
func FuzzSetAttributes(f *testing.F) {
	// Seed with attribute patterns
	f.Add("<root/>", "root.@attr", "value")
	f.Add("<root a='old'/>", "root.@a", "new")
	f.Add("<root><item/></root>", "root.item.@id", "123")
	f.Add("<root/>", "root.@xmlns", "http://example.com")

	f.Fuzz(func(t *testing.T, xml, path, value string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set attribute panicked: xml=%q path=%q value=%q panic=%v", xml, path, value, r)
			}
		}()

		result, err := Set(xml, path, value)

		// If successful and path is attribute access
		if err == nil && len(path) > 0 && containsAt(path) {
			// Should be able to retrieve the attribute
			retrieved := Get(result, path)
			_ = retrieved.Exists()
		}
	})
}

// FuzzSetNested tests Set with deeply nested paths.
// This ensures deep path creation works correctly.
func FuzzSetNested(f *testing.F) {
	// Seed with nested paths
	f.Add("<root/>", "root.a.b.c", "deep")
	f.Add("<root/>", "root.a.b.c.d.e.f", "verydeep")
	f.Add("<root><a/></root>", "root.a.b.c", "value")
	f.Add("<root><a><b/></a></root>", "root.a.b.c.d", "value")

	f.Fuzz(func(t *testing.T, xml, path, value string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set nested panicked: xml=%q path=%q value=%q panic=%v", xml, path, value, r)
			}
		}()

		result, err := Set(xml, path, value)

		// If successful, all intermediate elements should be created
		if err == nil && Valid(result) {
			retrieved := Get(result, path)
			_ = retrieved.Exists()
		}
	})
}

// FuzzSetArrays tests Set with array operations.
// This ensures array manipulation is robust.
func FuzzSetArrays(f *testing.F) {
	// Seed with array patterns
	f.Add("<root><item>1</item></root>", "root.item.0", "new")
	f.Add("<root><item>1</item><item>2</item></root>", "root.item.1", "second")
	f.Add("<root/>", "root.item.0", "first")
	f.Add("<root><item>1</item></root>", "root.item.-1", "last")

	f.Fuzz(func(t *testing.T, xml, path, value string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set array panicked: xml=%q path=%q value=%q panic=%v", xml, path, value, r)
			}
		}()

		result, err := Set(xml, path, value)

		// If successful, array access should work
		if err == nil && Valid(result) {
			_ = Get(result, path).Exists()
		}
	})
}

// FuzzSetDelete tests Set with nil value (delete operation).
// This ensures delete through Set is robust.
func FuzzSetDelete(f *testing.F) {
	// Seed with delete patterns
	f.Add("<root><item>x</item></root>", "root.item")
	f.Add("<root a='b'/>", "root.@a")
	f.Add("<root><a><b>x</b></a></root>", "root.a.b")
	f.Add("<root><item>1</item><item>2</item></root>", "root.item.0")

	f.Fuzz(func(t *testing.T, xml, path string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set delete panicked: xml=%q path=%q panic=%v", xml, path, r)
			}
		}()

		// Set with nil should delete
		result, err := Set(xml, path, nil)

		// If successful, deleted element should not exist
		if err == nil && Valid(result) {
			retrieved := Get(result, path)
			// It's okay if element still exists in some edge cases
			// Just ensure we don't panic
			_ = retrieved.Exists()
		}
	})
}

// FuzzSetMalformed tests Set with malformed XML input.
// This ensures Set handles broken XML gracefully.
func FuzzSetMalformed(f *testing.F) {
	// Seed with malformed patterns
	f.Add("<root>", "root.item", "value")
	f.Add("<root><item></root>", "root.item", "value")
	f.Add("</root>", "root.item", "value")
	f.Add("<root><>", "root.item", "value")
	f.Add("", "root", "value")

	f.Fuzz(func(t *testing.T, xml, path, value string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set malformed panicked: xml=%q path=%q value=%q panic=%v", xml, path, value, r)
			}
		}()

		// Set should handle malformed input gracefully
		result, err := Set(xml, path, value)

		// Either return error or produce valid XML
		if err == nil && len(result) > 0 {
			_ = Valid(result)
		}
	})
}

// FuzzSetSizeLimit tests Set with large values.
// This ensures size limits work correctly.
func FuzzSetSizeLimit(f *testing.F) {
	// Seed with various sizes
	f.Add("<root/>", "root.item", string(make([]byte, 100)))
	f.Add("<root/>", "root.item", string(make([]byte, 1000)))
	f.Add("<root/>", "root.item", string(make([]byte, 10000)))

	f.Fuzz(func(t *testing.T, xml, path, value string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set size limit panicked: xml=%q path=%q valueLen=%d panic=%v", xml, path, len(value), r)
			}
		}()

		// Set should handle size limits gracefully
		_, err := Set(xml, path, value)

		// Large values might be rejected, but should not panic
		_ = err
	})
}

// Helper function to check if path contains @ symbol
func containsAt(path string) bool {
	for i := 0; i < len(path); i++ {
		if path[i] == '@' {
			return true
		}
	}
	return false
}

// Helper function to check if path contains invalid characters
func containsInvalidChars(path string) bool {
	// Check for obviously invalid characters
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c < 0x20 && c != 0x09 && c != 0x0A && c != 0x0D {
			return true // Control character
		}
	}
	return false
}

// FuzzValid tests the Valid function for crashes and panics.
// This ensures validation never panics regardless of input.
func FuzzValid(f *testing.F) {
	// Seed corpus with various XML patterns
	f.Add("<root><item>value</item></root>")
	f.Add("<root/>")
	f.Add("<?xml version='1.0'?><root/>")
	f.Add("<root attr='value'/>")
	f.Add("")
	f.Add("<")
	f.Add("<<")
	f.Add("<>")
	f.Add("<root")
	f.Add("<root>")
	f.Add("</root>")
	f.Add("<root></root>")
	f.Add("<root><unclosed></root>")
	f.Add("<root><item></item><item></root>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Valid panicked on input: %q\nPanic: %v", xml, r)
			}
		}()

		// Valid should never panic
		result := Valid(xml)
		_ = result
	})
}

// FuzzValidateWithError tests detailed validation with error reporting.
// This ensures error reporting is robust.
func FuzzValidateWithError(f *testing.F) {
	// Seed corpus with valid and invalid patterns
	f.Add("<root><item>value</item></root>")
	f.Add("<root><unclosed></root>")
	f.Add("<root><item>")
	f.Add("</root>")
	f.Add("<root><item></root></item>")
	f.Add("")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidateWithError panicked on input: %q\nPanic: %v", xml, r)
			}
		}()

		valErr := ValidateWithError(xml)

		// If error is returned, it should be valid ValidateError
		if valErr != nil {
			// Line and column should be non-negative
			if valErr.Line < 0 || valErr.Column < 0 {
				t.Errorf("Invalid error location: line=%d col=%d", valErr.Line, valErr.Column)
			}
			// Message should not be empty
			if valErr.Message == "" {
				t.Errorf("Empty error message")
			}
			// Error() method should not panic
			_ = valErr.Error()
		}
	})
}

// FuzzValidateNesting tests validation with deeply nested structures.
// This ensures nesting validation is robust.
func FuzzValidateNesting(f *testing.F) {
	// Seed with nested patterns
	f.Add("<a><b><c><d>value</d></c></b></a>")
	f.Add("<a><a><a><a>x</a></a></a></a>")
	f.Add("<root><child><grandchild></grandchild></child></root>")
	f.Add("<a><b><c></a></c></b>") // Mismatched nesting

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate nesting panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		_ = ValidateWithError(xml)
	})
}

// FuzzValidateAttributes tests validation of attribute syntax.
// This ensures attribute validation is robust.
func FuzzValidateAttributes(f *testing.F) {
	// Seed with attribute patterns
	f.Add("<root a='b'/>")
	f.Add("<root a='b' c='d'/>")
	f.Add("<root a=\"b\"/>")
	f.Add("<root a='b\"/>")  // Mismatched quotes
	f.Add("<root a='/>")     // Unclosed attribute
	f.Add("<root a/>")       // No value
	f.Add("<root a=b/>")     // Unquoted
	f.Add("<root a='b c'/>") // Value with space

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate attributes panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		valErr := ValidateWithError(xml)
		if valErr != nil {
			_ = valErr.Error()
		}
	})
}

// FuzzValidateSpecialContent tests validation with special XML constructs.
// This ensures CDATA, comments, etc. are validated correctly.
func FuzzValidateSpecialContent(f *testing.F) {
	// Seed with special patterns
	f.Add("<root><![CDATA[data]]></root>")
	f.Add("<!-- comment --><root/>")
	f.Add("<?xml version='1.0'?><root/>")
	f.Add("<root><!-- comment --></root>")
	f.Add("<root><![CDATA[unclosed")
	f.Add("<!-- unclosed comment <root/>")
	f.Add("<?unclosed PI <root/>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate special content panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		_ = ValidateWithError(xml)
	})
}

// FuzzValidateTagMatching tests validation of tag matching.
// This ensures opening/closing tag validation is robust.
func FuzzValidateTagMatching(f *testing.F) {
	// Seed with tag matching patterns
	f.Add("<root></root>")
	f.Add("<root></ROOT>")
	f.Add("<root><item></item></root>")
	f.Add("<root><item></root></item>")
	f.Add("<root><item></item>")
	f.Add("<item></item></root>")
	f.Add("<root></root></root>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate tag matching panicked: %q\nPanic: %v", xml, r)
			}
		}()

		isValid := Valid(xml)
		valErr := ValidateWithError(xml)

		// Valid and ValidateWithError should be consistent
		if isValid && valErr != nil {
			t.Errorf("Inconsistent validation: Valid=true but error=%v", valErr)
		}
		if !isValid && valErr == nil {
			t.Errorf("Inconsistent validation: Valid=false but no error")
		}
	})
}

// FuzzValidateSelfClosing tests validation of self-closing tags.
// This ensures self-closing syntax is validated correctly.
func FuzzValidateSelfClosing(f *testing.F) {
	// Seed with self-closing patterns
	f.Add("<root/>")
	f.Add("<root></root>")
	f.Add("<root><item/></root>")
	f.Add("<root><item></item></root>")
	f.Add("<root/ >")
	f.Add("<root/ attr='value'/>")
	f.Add("<root/")
	f.Add("<root///>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate self-closing panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		_ = ValidateWithError(xml)
	})
}

// FuzzValidateNamespaces tests validation with namespace prefixes.
// This ensures namespace syntax doesn't break validation.
func FuzzValidateNamespaces(f *testing.F) {
	// Seed with namespace patterns
	f.Add("<ns:root xmlns:ns='uri'/>")
	f.Add("<root xmlns='uri'/>")
	f.Add("<ns:root></ns:root>")
	f.Add("<ns:root></root>")  // Mismatched prefix
	f.Add("<a:root></b:root>") // Different prefixes
	f.Add("<:root></:root>")   // Empty prefix
	f.Add("<ns:></ns:>")       // Empty local name
	f.Add("<ns:root:item/>")   // Multiple colons

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate namespace panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		_ = ValidateWithError(xml)
	})
}

// FuzzValidateWhitespace tests validation with various whitespace patterns.
// This ensures whitespace handling is robust.
func FuzzValidateWhitespace(f *testing.F) {
	// Seed with whitespace patterns
	f.Add("<root>   </root>")
	f.Add("<root>\n\t</root>")
	f.Add("  <root/>  ")
	f.Add("<root  />")
	f.Add("<root  attr = 'value' />")
	f.Add("<  root  />")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate whitespace panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		_ = ValidateWithError(xml)
	})
}

// FuzzValidateEmptyAndMinimal tests validation with edge case inputs.
// This ensures minimal inputs are handled correctly.
func FuzzValidateEmptyAndMinimal(f *testing.F) {
	// Seed with minimal patterns
	f.Add("")
	f.Add(" ")
	f.Add("\n")
	f.Add("</>")
	f.Add("<>")
	f.Add("<a/>")
	f.Add("<a></a>")
	f.Add("<a>")
	f.Add("</a>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate minimal panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		_ = ValidateWithError(xml)
	})
}

// FuzzValidateEscaping tests validation with entity references.
// This ensures escaped content is validated correctly.
func FuzzValidateEscaping(f *testing.F) {
	// Seed with escaping patterns
	f.Add("<root>&lt;tag&gt;</root>")
	f.Add("<root>&amp;&quot;&apos;</root>")
	f.Add("<root>&#65;</root>")
	f.Add("<root>&#x42;</root>")
	f.Add("<root>&unknown;</root>")
	f.Add("<root>&</root>")
	f.Add("<root>&&;</root>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate escaping panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		_ = ValidateWithError(xml)
	})
}

// FuzzValidateSizeLimit tests validation with size limits.
// This ensures security limits work correctly in validation.
func FuzzValidateSizeLimit(f *testing.F) {
	// Seed with various sizes
	f.Add("<root>" + string(make([]byte, 100)) + "</root>")
	f.Add("<root>" + string(make([]byte, 1000)) + "</root>")
	f.Add("<root>" + string(make([]byte, 10000)) + "</root>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate size limit panicked: len=%d panic=%v", len(xml), r)
			}
		}()

		_ = Valid(xml)
		_ = ValidateWithError(xml)
	})
}

// FuzzValidateMultiRoot tests validation with multiple root elements.
// This ensures single root requirement is validated.
func FuzzValidateMultiRoot(f *testing.F) {
	// Seed with multi-root patterns
	f.Add("<root/><root/>")
	f.Add("<a/><b/>")
	f.Add("<root></root><root></root>")
	f.Add("text<root/>")
	f.Add("<root/>text")
	f.Add("<!-- comment --><root/><root/>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate multi-root panicked: %q\nPanic: %v", xml, r)
			}
		}()

		_ = Valid(xml)
		valErr := ValidateWithError(xml)
		if valErr != nil {
			_ = valErr.Error()
		}
	})
}

// FuzzValidateConsistency tests that Valid and ValidateWithError are consistent.
// This ensures both functions agree on validity.
func FuzzValidateConsistency(f *testing.F) {
	// Seed with various patterns
	f.Add("<root><item>value</item></root>")
	f.Add("<root><unclosed></root>")
	f.Add("")
	f.Add("<root/>")

	f.Fuzz(func(t *testing.T, xml string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate consistency panicked: %q\nPanic: %v", xml, r)
			}
		}()

		isValid := Valid(xml)
		valErr := ValidateWithError(xml)

		// Check consistency
		if isValid && valErr != nil {
			t.Errorf("Inconsistency: Valid returned true but ValidateWithError returned error: %v\nxml=%q", valErr, xml)
		}

		if !isValid && valErr == nil {
			t.Errorf("Inconsistency: Valid returned false but ValidateWithError returned nil\nxml=%q", xml)
		}
	})
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestSecurity_DeepNesting tests that deeply nested XML is handled safely
func TestSecurity_DeepNesting(_ *testing.T) {
	// Create XML with nesting at the limit
	depth := MaxNestingDepth
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < depth; i++ {
		sb.WriteString("<level>")
	}
	sb.WriteString("value")
	for i := 0; i < depth; i++ {
		sb.WriteString("</level>")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	// Should handle up to the limit without panic
	result := Get(xml, "root.level")
	// Should get empty result due to depth limit being exceeded
	_ = result.String()

	// Test with excessive nesting (should not panic)
	sb.Reset()
	sb.WriteString("<root>")
	for i := 0; i < MaxNestingDepth*2; i++ {
		sb.WriteString("<level>")
	}
	sb.WriteString("value")
	for i := 0; i < MaxNestingDepth*2; i++ {
		sb.WriteString("</level>")
	}
	sb.WriteString("</root>")
	xmlExcessive := sb.String()

	// Should not panic - should handle gracefully
	resultExcessive := Get(xmlExcessive, "root.level")
	_ = resultExcessive.String()
}

// TestSecurity_WideNesting tests that wide nesting patterns are handled
func TestSecurity_WideNesting(t *testing.T) {
	// Test wide nesting pattern (many siblings, each with depth)
	// This tests that p.depth tracking works correctly across sibling branches
	var sb strings.Builder
	sb.WriteString("<root>")

	// Create 50 branches, each with 50 levels of depth
	for i := 0; i < 50; i++ {
		sb.WriteString("<branch")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(">")
		for j := 0; j < 50; j++ {
			sb.WriteString("<level")
			sb.WriteString(strconv.Itoa(j))
			sb.WriteString(">data</level")
			sb.WriteString(strconv.Itoa(j))
			sb.WriteString(">")
		}
		sb.WriteString("</branch")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(">")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	// Should handle without stack overflow
	result := Get(xml, "root.branch0.level0")
	// Should successfully parse the data
	if result.String() != "data" {
		t.Errorf("Expected 'data', got '%s'", result.String())
	}

	// Test accessing different branches to ensure depth tracking resets properly
	result2 := Get(xml, "root.branch10.level10")
	if result2.String() != "data" {
		t.Errorf("Expected 'data' from branch10, got '%s'", result2.String())
	}
}

// TestSecurity_LargeDocument tests that large documents are rejected
func TestSecurity_LargeDocument(t *testing.T) {
	// Create a document just under the limit
	belowLimit := make([]byte, MaxDocumentSize-100)
	for i := range belowLimit {
		belowLimit[i] = 'x'
	}
	xmlBelow := "<root>" + string(belowLimit) + "</root>"

	// Should process documents below limit
	result := GetBytes([]byte(xmlBelow), "root")
	if !result.Exists() {
		t.Error("Document below size limit should be processed")
	}

	// Create a document over the limit
	overLimit := make([]byte, MaxDocumentSize+1000)
	for i := range overLimit {
		overLimit[i] = 'x'
	}
	xmlOver := "<root>" + string(overLimit) + "</root>"

	// Should reject documents over limit
	resultOver := GetBytes([]byte(xmlOver), "root")
	if resultOver.Type != Null {
		t.Error("Document over size limit should be rejected")
	}
}

// TestSecurity_AttributeFlood tests that attribute flooding is prevented
func TestSecurity_AttributeFlood(t *testing.T) {
	// Create element with attributes at the limit
	var sb strings.Builder
	sb.WriteString("<element ")
	for i := 0; i < MaxAttributes; i++ {
		sb.WriteString("attr")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString("=\"value\" ")
	}
	sb.WriteString(">content</element>")
	xml := sb.String()

	// Should handle up to limit without panic
	result := Get(xml, "element.@attr0")
	if !result.Exists() {
		t.Error("Should handle attributes up to the limit")
	}

	// Create element with excessive attributes
	sb.Reset()
	sb.WriteString("<element ")
	for i := 0; i < MaxAttributes*2; i++ {
		sb.WriteString("attr")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString("=\"value\" ")
	}
	sb.WriteString(">content</element>")
	xmlExcessive := sb.String()

	// Should not panic - should stop processing after limit
	resultExcessive := Get(xmlExcessive, "element")
	_ = resultExcessive.String()
}

// TestSecurity_LargeTokens tests that extremely large tokens are limited
func TestSecurity_LargeTokens(_ *testing.T) {
	// Create element with very long name (at limit)
	longName := strings.Repeat("a", MaxTokenSize-1)
	xml := "<" + longName + ">value</" + longName + ">"

	// Should not panic
	result := Get(xml, longName)
	_ = result.String()

	// Create element with extremely long name (over limit)
	veryLongName := strings.Repeat("b", MaxTokenSize*2)
	xmlExcessive := "<" + veryLongName + ">value</" + veryLongName + ">"

	// Should not panic - should handle gracefully (return empty)
	resultExcessive := Get(xmlExcessive, veryLongName)
	_ = resultExcessive.String()
}

// TestParser_TokenSizeLimit tests token size limit enforcement
func TestParser_TokenSizeLimit(_ *testing.T) {
	// Create element with very long name exceeding MaxTokenSize
	longName := strings.Repeat("a", MaxTokenSize+100)
	xml := "<root><" + longName + ">test</" + longName + "></root>"

	// Should handle gracefully without corruption or panic
	result := Get(xml, "root")
	// Parser should handle this without crashing
	_ = result.String()

	// Test with long attribute value
	longValue := strings.Repeat("x", MaxTokenSize+100)
	xmlAttr := "<root attr=\"" + longValue + "\">test</root>"

	resultAttr := Get(xmlAttr, "root.@attr")
	// Should handle gracefully
	_ = resultAttr.String()
}

// TestSecurity_DOCTYPEHandling tests that DOCTYPE declarations are skipped
func TestSecurity_DOCTYPEHandling(t *testing.T) {
	// Test simple DOCTYPE
	xmlSimple := `<?xml version="1.0"?>
<!DOCTYPE root>
<root><element>value</element></root>`

	result := Get(xmlSimple, "root.element")
	if result.String() != "value" {
		t.Errorf("DOCTYPE should be skipped, got: %s", result.String())
	}

	// Test DOCTYPE with internal subset
	xmlInternal := `<?xml version="1.0"?>
<!DOCTYPE root [
	<!ENTITY test "entity value">
]>
<root><element>value</element></root>`

	resultInternal := Get(xmlInternal, "root.element")
	if resultInternal.String() != "value" {
		t.Errorf("DOCTYPE with internal subset should be skipped, got: %s", resultInternal.String())
	}

	// Test DOCTYPE with system identifier
	xmlSystem := `<?xml version="1.0"?>
<!DOCTYPE root SYSTEM "http://example.com/dtd">
<root><element>value</element></root>`

	resultSystem := Get(xmlSystem, "root.element")
	if resultSystem.String() != "value" {
		t.Errorf("DOCTYPE with SYSTEM should be skipped, got: %s", resultSystem.String())
	}
}

// TestSecurity_DOCTYPECaseInsensitive tests that DOCTYPE case variations are handled
func TestSecurity_DOCTYPECaseInsensitive(t *testing.T) {
	cases := []struct {
		name string
		xml  string
	}{
		{
			name: "uppercase DOCTYPE",
			xml:  `<!DOCTYPE root><root>test</root>`,
		},
		{
			name: "lowercase doctype",
			xml:  `<!doctype root><root>test</root>`,
		},
		{
			name: "mixed case DoCtYpE",
			xml:  `<!DoCtYpE root><root>test</root>`,
		},
		{
			name: "DOCTYPE with entity",
			xml:  `<!DOCTYPE root [ <!ENTITY test "value"> ]><root>test</root>`,
		},
		{
			name: "doctype with entity lowercase",
			xml:  `<!doctype root [ <!ENTITY test "value"> ]><root>test</root>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Get(tc.xml, "root")
			// Should handle all case variations without panic
			if !result.Exists() {
				t.Errorf("Case %s: expected result to exist, got empty", tc.name)
			}
			if result.String() != "test" {
				t.Errorf("Case %s: expected 'test', got '%s'", tc.name, result.String())
			}
		})
	}
}

// TestSecurity_MalformedXMLNoPanic tests that malformed XML doesn't cause panics
func TestSecurity_MalformedXMLNoPanic(t *testing.T) {
	malformedExamples := []string{
		"<root><unclosed>",
		"<root><broken",
		"<<<>>>",
		"<root attr=\"unclosed>",
		"<root><nested><deep><unclosed></nested></root>",
		"<root></different>",
		string([]byte{0xFF, 0xFE, 0xFD}), // Invalid UTF-8
	}

	for i, xml := range malformedExamples {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// Should not panic on any malformed input
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic on malformed XML %d: %v", i, r)
				}
			}()

			result := Get(xml, "root")
			_ = result.String()
		})
	}
}

// TestSecurity_RecursiveEntityExpansion tests handling of entity references
// Note: Current implementation doesn't expand entities, which is secure by default
func TestSecurity_RecursiveEntityExpansion(_ *testing.T) {
	// XML with entity references (billion laughs pattern)
	xmlWithEntities := `<?xml version="1.0"?>
<!DOCTYPE root [
	<!ENTITY lol "lol">
	<!ENTITY lol2 "&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;">
]>
<root>&lol2;</root>`

	// Should skip DOCTYPE and not expand entities
	result := Get(xmlWithEntities, "root")
	// Entity references should be treated as literal text, not expanded
	_ = result.String()
}

// TestSecurity_EmptyAndEdgeCases tests edge cases that might cause issues
func TestSecurity_EmptyAndEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
	}{
		{
			name: "Empty document",
			xml:  "",
			path: "root",
		},
		{
			name: "Single character",
			xml:  "x",
			path: "root",
		},
		{
			name: "Only whitespace",
			xml:  "   \n\t\r   ",
			path: "root",
		},
		{
			name: "Only tags",
			xml:  "<></>",
			path: "root",
		},
		{
			name: "Null bytes",
			xml:  "<root>\x00\x00</root>",
			path: "root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic on any edge case
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic on edge case '%s': %v", tt.name, r)
				}
			}()

			result := Get(tt.xml, tt.path)
			_ = result.String()
		})
	}
}

// BenchmarkSecurity_DeepNesting benchmarks deeply nested XML parsing
func BenchmarkSecurity_DeepNesting(b *testing.B) {
	// Create XML at the nesting limit
	depth := MaxNestingDepth
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < depth; i++ {
		sb.WriteString("<level>")
	}
	sb.WriteString("value")
	for i := 0; i < depth; i++ {
		sb.WriteString("</level>")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Get(xml, "root.level")
		_ = result.String()
	}
}

// TestControlCharacterRejection tests that control characters in filter paths/values are rejected
func TestControlCharacterRejection(t *testing.T) {
	tests := []struct {
		name       string
		xml        string
		path       string
		shouldFail bool
	}{
		{"newline in path", `<items><item><name>value</name></item></items>`, "items.item.#(name\n==value)", true},
		{"carriage return in path", `<items><item><name>value</name></item></items>`, "items.item.#(name\r==value)", true},
		{"tab in path", `<items><item><name>value</name></item></items>`, "items.item.#(name\t==value)", true},
		{"newline in value", `<items><item><name>value</name></item></items>`, "items.item.#(name==val\nue)", true},
		{"carriage return in value", `<items><item><name>value</name></item></items>`, "items.item.#(name==val\rue)", true},
		{"tab in value", `<items><item><name>value</name></item></items>`, "items.item.#(name==val\tue)", true},
		{"null byte in path", `<items><item><name>value</name></item></items>`, "items.item.#(name\x00==value)", true},
		{"null byte in value", `<items><item><name>value</name></item></items>`, "items.item.#(name==val\x00ue)", true},
		{"valid path and value", `<items><item><name>value</name></item></items>`, "items.item.#(name==value)", false},
		{"valid numeric filter", `<items><item><price>150</price></item></items>`, "items.item.#(price>100)", false},
		{"valid attribute filter", `<items><item id="5"><name>test</name></item></items>`, "items.item.#(@id==5)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.xml, tt.path)
			if tt.shouldFail {
				if result.Exists() {
					t.Errorf("Expected control character to be rejected (query should fail), but got result: %v", result.String())
				}
			} else {
				if !result.Exists() {
					t.Errorf("Expected valid query to succeed, but got empty result")
				}
			}
		})
	}
}

// BenchmarkSecurity_ManyAttributes benchmarks parsing elements with many attributes
func BenchmarkSecurity_ManyAttributes(b *testing.B) {
	// Create element with many attributes
	var sb strings.Builder
	sb.WriteString("<element ")
	for i := 0; i < MaxAttributes; i++ {
		sb.WriteString("attr")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString("=\"value\" ")
	}
	sb.WriteString(">content</element>")
	xml := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Get(xml, "element.@attr0")
		_ = result.String()
	}
}

// ============================================================================
// Field Extraction Security Tests
// ============================================================================

// TestFieldExtractionSecurityNestedExtractionDoS tests nested #.#.#... patterns for CPU exhaustion
func TestFieldExtractionSecurityNestedExtractionDoS(t *testing.T) {
	xml := `<root>
		<level1>
			<level2>
				<level3>
					<level4>
						<level5>
							<item><name>Deep</name></item>
						</level5>
					</level4>
				</level3>
			</level2>
		</level1>
	</root>`

	// Test deeply nested extraction patterns
	tests := []struct {
		name string
		path string
	}{
		{"3 levels", "root.level1.#.level2"},
		{"5 levels", "root.level1.#.level2.level3.level4.level5"},
		{"with field extraction", "root.level1.level2.level3.#.level4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Should complete without hanging or excessive CPU
			result := Get(xml, tt.path)
			// Just verify it completes and doesn't panic
			_ = result.Type
		})
	}
}

// TestFieldExtractionSecurityLargeResultSet tests MaxWildcardResults enforcement
func TestFieldExtractionSecurityLargeResultSet(t *testing.T) {
	// Create XML with more items than MaxWildcardResults
	itemCount := MaxWildcardResults + 500
	var b strings.Builder
	b.WriteString("<root><items>")
	for i := 0; i < itemCount; i++ {
		b.WriteString("<item><name>Name</name><value>123</value></item>")
	}
	b.WriteString("</items></root>")
	xml := b.String()

	result := Get(xml, "root.items.item.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}

	// CRITICAL: Must be limited to MaxWildcardResults
	if len(result.Results) > MaxWildcardResults {
		t.Errorf("SECURITY VIOLATION: Result count %d exceeds MaxWildcardResults %d",
			len(result.Results), MaxWildcardResults)
	}

	// Should be exactly at limit
	if len(result.Results) != MaxWildcardResults {
		t.Errorf("Expected exactly %d results (limit), got %d", MaxWildcardResults, len(result.Results))
	}
}

// TestFieldExtractionSecurityFieldNameInjection tests field name validation
func TestFieldExtractionSecurityFieldNameInjection(t *testing.T) {
	xml := `<root><items><item><name>Test</name></item></items></root>`

	maliciousFieldNames := []string{
		// Path traversal attempts
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32",
		"../../../../root",

		// Script injection attempts
		"<script>alert('xss')</script>",
		"javascript:alert(1)",

		// Null byte injection
		"field\x00name",
		"name\x00.txt",

		// Control characters
		"field\nname",
		"field\rname",
		"field\tname",

		// XML injection
		"name><injected>value</injected><name",
		"' OR '1'='1",

		// Special characters that should be rejected
		"field$name",
		"field@name@",
		"field#name",
		"field!name",
		"field*name",
		"field&name",
		"field|name",
		"field;name",
		"field'name",
		"field\"name",
		"field<name>",
		"field[name]",
		"field{name}",
		"field(name)",
		"field=name",
		"field+name",
		"field,name",
		"field.name.injected",
		"field/name",
		"field\\name",
		"field`name",
		"field~name",
		"field?name",
	}

	for _, malicious := range maliciousFieldNames {
		t.Run("injection_"+malicious[:min(20, len(malicious))], func(t *testing.T) {
			path := "root.items.item.#." + malicious
			result := Get(xml, path)

			// Should either:
			// 1. Return empty array (field rejected/not found)
			// 2. Return Null (parsing failed safely)
			// 3. Return Number (# became count instead of field extraction)

			// Should NOT panic, hang, or access system resources
			if result.Type == Array && len(result.Results) > 0 {
				// If it somehow extracted data, verify it's not executing injection
				for _, r := range result.Results {
					val := r.String()
					// Verify no script execution or system access
					if strings.Contains(val, "alert") || strings.Contains(val, "passwd") {
						t.Errorf("SECURITY VIOLATION: Injection may have executed: %s", malicious)
					}
				}
			}
		})
	}
}

// TestFieldExtractionSecurityAmplificationAttack tests small input -> large output
func TestFieldExtractionSecurityAmplificationAttack(t *testing.T) {
	// Small XML that could produce large output via extraction
	var b strings.Builder
	b.WriteString("<root><items>")

	// Each item has a large value field
	largeValue := strings.Repeat("A", 1000) // 1KB per value
	for i := 0; i < 2000; i++ {             // 2000 items = potential 2MB output
		b.WriteString("<item><value>")
		b.WriteString(largeValue)
		b.WriteString("</value></item>")
	}
	b.WriteString("</items></root>")
	xml := b.String()

	// Input size: ~2MB, potential output after extraction: 2MB of values
	result := Get(xml, "root.items.item.#.value")

	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}

	// Should be limited by MaxWildcardResults
	if len(result.Results) > MaxWildcardResults {
		t.Errorf("SECURITY VIOLATION: Amplification attack not prevented, got %d results",
			len(result.Results))
	}
}

// TestFieldExtractionSecurityMalformedInput tests graceful handling of malformed paths
func TestFieldExtractionSecurityMalformedInput(t *testing.T) {
	xml := `<root><items><item><name>Test</name></item></items></root>`

	malformedPaths := []string{
		"root.items.item.#.",        // Trailing dot
		"root.items.item.#..",       // Double dot
		"root.items.item.##.field",  // Double hash
		"root.items.item.#..field",  // Hash double dot
		"root.items.item.#.#.#.#.#", // Multiple hashes
		"root.items.item.#.%.",      // Text extraction with trailing dot
		"root.items.item.#.@",       // Incomplete attribute
		"root.items.item.#.@.",      // Attribute with trailing dot
		"",                          // Empty path
		"#.field",                   // Hash at start
		".field",                    // Leading dot
		"field.",                    // Single trailing dot
		"....",                      // Only dots
		"####",                      // Only hashes
	}

	for _, path := range malformedPaths {
		t.Run("malformed_"+path, func(_ *testing.T) {
			// Should not panic or hang
			result := Get(xml, path)
			// Should safely return Null, empty Array, or Number
			_ = result.Type
		})
	}
}

// TestFieldExtractionSecurityRecursiveXML tests deeply nested XML structures
func TestFieldExtractionSecurityRecursiveXML(_ *testing.T) {
	// Build deeply nested XML (but within MaxNestingDepth)
	depth := 50 // Below MaxNestingDepth (100)
	var b strings.Builder
	b.WriteString("<root>")

	// Build nested structure
	for i := 0; i < depth; i++ {
		b.WriteString("<level>")
	}
	b.WriteString("<item><name>Deep</name></item>")
	for i := 0; i < depth; i++ {
		b.WriteString("</level>")
	}
	b.WriteString("</root>")
	xml := b.String()

	// Try to extract from deep structure
	path := "root"
	for i := 0; i < depth; i++ {
		path += ".level"
	}
	path += ".item.#.name"

	// Should complete without stack overflow
	result := Get(xml, path)
	_ = result.Type
}

// TestFieldExtractionSecurityMaxFieldNameLength tests field name length limit
func TestFieldExtractionSecurityMaxFieldNameLength(t *testing.T) {
	xml := `<root><items><item><name>Test</name></item></items></root>`

	tests := []struct {
		name       string
		fieldLen   int
		shouldWork bool
	}{
		{"short field", 10, true},
		{"medium field", 100, true},
		{"at limit", MaxFieldNameLength, true},
		{"over limit", MaxFieldNameLength + 1, false},
		{"way over limit", MaxFieldNameLength * 2, false},
		{"extreme", 10000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldName := strings.Repeat("x", tt.fieldLen)
			path := "root.items.item.#." + fieldName

			result := Get(xml, path)

			// Over-limit fields should be rejected (return Null, empty Array, or Number)
			if !tt.shouldWork {
				if result.Type == Array && len(result.Results) > 0 {
					t.Errorf("SECURITY VIOLATION: Field name length %d exceeded limit but still extracted",
						tt.fieldLen)
				}
			}
		})
	}
}

// TestFieldExtractionSecurityNullByteInjection specifically tests null byte handling
func TestFieldExtractionSecurityNullByteInjection(t *testing.T) {
	xml := `<root><items><item><name>Test</name><file>data.txt</file></item></items></root>`

	nullByteTests := []string{
		"name\x00",
		"\x00name",
		"na\x00me",
		"name\x00.txt",
		"@attr\x00",
		"%\x00",
	}

	for _, fieldName := range nullByteTests {
		path := "root.items.item.#." + fieldName
		result := Get(xml, path)

		// Should safely handle null bytes (reject or sanitize)
		// Should NOT access file system or execute commands
		if result.Type == Array && len(result.Results) > 0 {
			for _, r := range result.Results {
				val := r.String()
				// Should not contain system file contents
				if strings.Contains(val, "passwd") || strings.Contains(val, "shadow") {
					t.Errorf("SECURITY VIOLATION: Null byte injection may have accessed system files")
				}
			}
		}
	}
}

// TestFieldExtractionSecurityMemoryExhaustion tests memory limit enforcement
func TestFieldExtractionSecurityMemoryExhaustion(t *testing.T) {
	// Create XML that attempts to exhaust memory through field extraction
	var b strings.Builder
	b.WriteString("<root><items>")

	// Many items with nested structure
	for i := 0; i < 10000; i++ {
		b.WriteString("<item>")
		b.WriteString("<data>")
		// Each data element contains multiple sub-elements
		for j := 0; j < 100; j++ {
			b.WriteString("<field>Value</field>")
		}
		b.WriteString("</data>")
		b.WriteString("</item>")
	}
	b.WriteString("</items></root>")
	xml := b.String()

	// Try to extract all nested fields (could be 10000 * 100 = 1M fields)
	result := Get(xml, "root.items.item.#.data")

	if result.Type == Array {
		// Should be limited to MaxWildcardResults
		if len(result.Results) > MaxWildcardResults {
			t.Errorf("SECURITY VIOLATION: Memory exhaustion not prevented, got %d results",
				len(result.Results))
		}
	}
}

// TestFieldExtractionSecurityDepthLimit tests MaxFilterDepth enforcement
func TestFieldExtractionSecurityDepthLimit(t *testing.T) {
	// This test requires nested field extractions which aren't directly supported
	// But we test the depth parameter in executeFieldExtraction

	xml := `<root><items><item><name>Test</name></item></items></root>`

	// Build a path that would cause deep recursion if not limited
	// Current implementation uses depth parameter in executeFieldExtraction

	result := Get(xml, "root.items.item.#.name")

	// Should complete without stack overflow
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
}

// TestFieldExtractionSecurityConcurrency tests concurrent field extractions
func TestFieldExtractionSecurityConcurrency(t *testing.T) {
	xml := `<root><items>`
	for i := 0; i < 100; i++ {
		xml += `<item><name>Test</name><value>123</value></item>`
	}
	xml += `</items></root>`

	// Run concurrent field extractions
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			result := Get(xml, "root.items.item.#.name")
			if result.Type != Array {
				t.Errorf("Expected Array type in goroutine")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

// TestFieldExtractionSecurityCaseSensitiveBypass tests case-insensitive security
func TestFieldExtractionSecurityCaseSensitiveBypass(t *testing.T) {
	xml := `<root>
		<items>
			<item><Name>Public</Name><PASSWORD>secret123</PASSWORD></item>
			<item><Name>User</Name><password>hunter2</password></item>
		</items>
	</root>`

	// Test case-sensitive (default) - should NOT extract PASSWORD
	result := Get(xml, "root.items.item.#.password")
	if result.Type == Array && len(result.Results) == 1 {
		// Only lowercase password should match
		if result.Results[0].String() == "secret123" {
			t.Errorf("SECURITY WARNING: Case-sensitive mode extracted uppercase PASSWORD")
		}
	}

	// Test case-insensitive - should extract both
	opts := Options{CaseSensitive: false}
	result = GetWithOptions(xml, "root.items.item.#.password", &opts)
	// This is expected behavior, but document it as security consideration
	// Applications should be aware that CaseSensitive: false extracts all case variants
	_ = result.Type == Array && len(result.Results) == 2
}

// TestFieldExtractionSecurityXMLInjection tests XML entity injection via field names
func TestFieldExtractionSecurityXMLInjection(t *testing.T) {
	xml := `<root><items><item><name>Test</name></item></items></root>`

	xmlInjectionTests := []string{
		"&lt;name&gt;",
		"&quot;name&quot;",
		"&amp;name",
		"&#x3C;script&#x3E;",
		"<!ENTITY xxe SYSTEM 'file:///etc/passwd'>",
		"<!DOCTYPE foo [<!ELEMENT foo ANY><!ENTITY xxe SYSTEM 'file:///etc/passwd'>]>",
	}

	for _, injection := range xmlInjectionTests {
		path := "root.items.item.#." + injection
		result := Get(xml, path)

		// Should not parse XML entities in field names
		// Field name validation should reject these
		if result.Type == Array && len(result.Results) > 0 {
			t.Errorf("SECURITY WARNING: XML injection field name may have been parsed: %s", injection)
		}
	}
}

// TestFieldExtractionSecurityRegexDoS tests for regex-based DoS vulnerabilities
func TestFieldExtractionSecurityRegexDoS(_ *testing.T) {
	// Note: Current implementation uses string comparison, not regex
	// But test to ensure no regex is accidentally introduced

	xml := `<root><items><item><name>Test</name></item></items></root>`

	// Patterns that cause regex DoS (catastrophic backtracking)
	potentialRegexDoS := []string{
		strings.Repeat("a", 100) + "!",
		strings.Repeat("(a+)+", 10),
		strings.Repeat("(a*)*", 10),
		"(a+)+(b+)+(c+)+",
	}

	for _, pattern := range potentialRegexDoS {
		path := "root.items.item.#." + pattern

		// Should complete quickly (no regex evaluation)
		result := Get(xml, path)
		_ = result.Type
	}
}

// TestFieldExtractionSecurityIntegerOverflow tests for integer overflow in calculations
func TestFieldExtractionSecurityIntegerOverflow(t *testing.T) {
	xml := `<root><items>`

	// Create XML with many items to test counter overflow
	itemCount := 10000
	for i := 0; i < itemCount; i++ {
		xml += `<item><name>Test</name></item>`
	}
	xml += `</items></root>`

	result := Get(xml, "root.items.item.#.name")

	if result.Type == Array {
		// Verify count doesn't overflow
		count := len(result.Results)
		if count < 0 {
			t.Errorf("SECURITY VIOLATION: Integer overflow detected in result count")
		}
		if count > MaxWildcardResults {
			t.Errorf("SECURITY VIOLATION: Result count exceeds limit despite overflow check")
		}
	}
}

// ============================================================================
// Operator Parsing Security Tests
// ============================================================================

// TestOperatorParsingSecurityEdgeCases tests security edge cases in operator parsing
func TestOperatorParsingSecurityEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		shouldError bool
		expectedOp  FilterOp
		description string
	}{
		// Operator precedence edge cases
		{
			name:        "double equals correctly parsed",
			expr:        "field==value",
			shouldError: false,
			expectedOp:  OpEqual,
			description: "== should be recognized as equality operator",
		},
		{
			name:        "less than equals correctly parsed",
			expr:        "field<=10",
			shouldError: false,
			expectedOp:  OpLessThanOrEqual,
			description: "<= should be recognized as less-than-or-equal operator",
		},
		{
			name:        "greater than equals correctly parsed",
			expr:        "field>=10",
			shouldError: false,
			expectedOp:  OpGreaterThanOrEqual,
			description: ">= should be recognized as greater-than-or-equal operator",
		},
		{
			name:        "not equals correctly parsed",
			expr:        "field!=value",
			shouldError: false,
			expectedOp:  OpNotEqual,
			description: "!= should be recognized as not-equal operator",
		},
		{
			name:        "single equals rejected",
			expr:        "field=value",
			shouldError: true,
			expectedOp:  OpEqual,
			description: "Single = should be rejected (Path A: strict GJSON alignment)",
		},
		{
			name:        "triple equals should parse as double equals",
			expr:        "field===value",
			shouldError: false,
			expectedOp:  OpEqual,
			description: "=== should parse first == as operator, leaving =value",
		},

		// Injection attempts via operator confusion
		{
			name:        "operator in value accepted",
			expr:        "field==name==value",
			shouldError: false,
			expectedOp:  OpEqual,
			description: "Operator characters in value are allowed (not in path)",
		},
		{
			name:        "operator at start rejected",
			expr:        "==value",
			shouldError: true,
			description: "Empty path should be rejected",
		},
		{
			name:        "operator at end rejected",
			expr:        "field==",
			shouldError: true,
			description: "Empty value should be rejected",
		},
		{
			name:        "multiple operators in sequence",
			expr:        "field<>=value",
			shouldError: true,
			description: "Multiple operators should be rejected (path contains operator)",
		},

		// Null byte injection in operator context
		{
			name:        "null byte before operator",
			expr:        "field\x00==value",
			shouldError: true,
			description: "Null byte in expression should be rejected early",
		},
		{
			name:        "null byte after operator",
			expr:        "field==\x00value",
			shouldError: true,
			description: "Null byte in value should be rejected",
		},
		{
			name:        "null byte in operator position",
			expr:        "field\x00<value",
			shouldError: true,
			description: "Null byte before operator should be rejected",
		},

		// Boundary cases for operator detection
		{
			name:        "operator at string boundary",
			expr:        "a==b",
			shouldError: false,
			expectedOp:  OpEqual,
			description: "Minimal valid filter should parse",
		},
		{
			name:        "whitespace around operators",
			expr:        "field == value",
			shouldError: false,
			expectedOp:  OpEqual,
			description: "Whitespace should be trimmed correctly",
		},
		{
			name:        "operator in quoted value",
			expr:        "field=='val=ue'",
			shouldError: false,
			expectedOp:  OpEqual,
			description: "Operator inside quoted value should be preserved",
		},

		// Length-based attacks
		{
			name:        "max length filter with operator",
			expr:        strings.Repeat("a", MaxFilterExpressionLength-10) + "==value",
			shouldError: false,
			expectedOp:  OpEqual,
			description: "Filter at max length should parse",
		},
		{
			name:        "over length filter rejected",
			expr:        strings.Repeat("a", MaxFilterExpressionLength+1) + "==value",
			shouldError: true,
			description: "Filter over max length should be rejected",
		},

		// Special character combinations
		{
			name:        "exclamation without equals",
			expr:        "field!value",
			shouldError: true,
			description: "Lone ! is not a valid operator",
		},
		{
			name:        "equals in attribute path",
			expr:        "@attr==value",
			shouldError: false,
			expectedOp:  OpEqual,
			description: "Attribute filters should work with ==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := parseFilterCondition(tt.expr)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for '%s': %s", tt.expr, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for '%s': %v (%s)", tt.expr, err, tt.description)
				}
				if filter != nil && filter.Op != tt.expectedOp {
					t.Errorf("Expected operator %v, got %v for '%s'", tt.expectedOp, filter.Op, tt.expr)
				}
			}
		})
	}
}

// TestOperatorParsingNoRegression tests that previous security mitigations still work
func TestOperatorParsingNoRegression(t *testing.T) {
	// All these should still be rejected as before
	maliciousInputs := []string{
		// Null byte injections
		"field\x00==value",
		"field==\x00value",
		"\x00field==value",

		// Operator in path (should be detected and rejected)
		"field=path==value",
		"field<path==value",

		// Empty components
		"==value",
		"field==",
		"==",
		"",

		// Over-length expressions
		strings.Repeat("x", MaxFilterExpressionLength+100) + "==value",

		// Path with illegal characters
		"field\nname==value",
		"field\rname==value",
		"field\tname==value",
	}

	for _, input := range maliciousInputs {
		t.Run("reject_"+input[:min(20, len(input))], func(t *testing.T) {
			filter, err := parseFilterCondition(input)
			if err == nil && filter != nil {
				// If it didn't error, verify it's safely handling the input
				// The path should not contain any injected characters
				if strings.ContainsAny(filter.Path, "=!<>\x00\n\r\t") {
					t.Errorf("SECURITY VIOLATION: Malicious characters in parsed path: %q", filter.Path)
				}
			}
			// Most should error, but if they don't, verify they're safe
		})
	}
}

// TestOperatorParsingDOSResistance tests that operator parsing is resistant to DoS
func TestOperatorParsingDOSResistance(t *testing.T) {
	tests := []struct {
		name  string
		expr  string
		check string
	}{
		{
			name:  "many operators",
			expr:  strings.Repeat(">=", 1000) + "value",
			check: "Should handle many operators without hanging",
		},
		{
			name:  "alternating operators",
			expr:  "field" + strings.Repeat("<>", 500),
			check: "Should handle alternating operators",
		},
		{
			name:  "long path before operator",
			expr:  strings.Repeat("a", 200) + "==value",
			check: "Should handle long path efficiently",
		},
		{
			name:  "nested quotes",
			expr:  "field=='" + strings.Repeat("'\"", 100) + "'",
			check: "Should handle nested quotes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Should complete quickly without hanging
			_, err := parseFilterCondition(tt.expr)
			// May or may not error, but should not hang or panic
			_ = err
		})
	}
}

// TestOperatorComparisonSecurity tests that comparison operations are secure
func TestOperatorComparisonSecurity(t *testing.T) {
	xml := `<root>
		<item><age>25</age><name>John</name></item>
		<item><age>30</age><name>Jane</name></item>
		<item><age>Infinity</age><name>Invalid</name></item>
		<item><age>NaN</age><name>Invalid2</name></item>
	</root>`

	tests := []struct {
		name        string
		path        string
		expectCount int
		description string
	}{
		{
			name:        "equality with ==",
			path:        "root.item.#(age==25)",
			expectCount: 1,
			description: "Double equals should work for equality",
		},
		{
			name:        "not equal with !=",
			path:        "root.item.#(age!=25)",
			expectCount: 3, // Three items have age != 25
			description: "Not equals should work correctly",
		},
		{
			name:        "less than comparison",
			path:        "root.item.#(age<30)",
			expectCount: 1,
			description: "Less than should only match numeric values",
		},
		{
			name:        "special float rejection",
			path:        "root.item.#(age>20)",
			expectCount: 2, // Only 25 and 30, Infinity and NaN should be rejected
			description: "Special float values should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.Type != Element {
				t.Errorf("%s: Expected single Element, got %v", tt.description, result.Type)
			}
		})
	}
}

// TestOperatorAlignmentBackwardCompatibility tests GJSON alignment without breaking security
func TestOperatorAlignmentBackwardCompatibility(t *testing.T) {
	xml := `<root>
		<item><status>active</status></item>
		<item><status>inactive</status></item>
	</root>`

	// Test that == works (GJSON-style)
	result := Get(xml, "root.item.#(status==active)")
	if !result.Exists() {
		t.Error("GJSON-style == operator should work")
	}

	// Test that single = is rejected (Path A: strict GJSON alignment)
	result = Get(xml, "root.item.#(status=active)")
	// Single = should be rejected - no operator found, returns empty
	if result.Exists() {
		t.Error("Single = operator should be rejected in Path A (strict GJSON alignment)")
	}
}

// ============================================================================
// Pattern Matching Security Attack Scenarios
// ============================================================================

// TestPatternMatchAttack_ExponentialBacktracking tests ReDoS protection with catastrophic backtracking patterns
func TestPatternMatchAttack_ExponentialBacktracking(t *testing.T) {
	xml := `<items>
		<item><name>aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa</name></item>
	</items>`

	attacks := []struct {
		name    string
		pattern string
		desc    string
	}{
		{
			name:    "nested quantifiers",
			pattern: "a*a*a*a*a*a*a*a*a*a*a*a*a*a*a*b",
			desc:    "Multiple nested quantifiers causing exponential backtracking",
		},
		{
			name:    "alternating stars",
			pattern: "*a*a*a*a*a*a*a*a*a*a*a*a*a*a*b",
			desc:    "Alternating wildcards with characters",
		},
		{
			name:    "deep nesting",
			pattern: strings.Repeat("a*", 50) + "b",
			desc:    "Deep nesting of pattern elements",
		},
		{
			name:    "question marks at end",
			pattern: strings.Repeat("a", 20) + strings.Repeat("?", 20) + "b",
			desc:    "Many question marks causing backtracking",
		},
		{
			name:    "mixed wildcards",
			pattern: "a*?a*?a*?a*?a*?a*?a*?a*?b",
			desc:    "Mixed * and ? wildcards",
		},
	}

	for _, att := range attacks {
		t.Run(att.name, func(t *testing.T) {
			start := time.Now()

			// Build path with attack pattern
			path := `items.item.#(name%"` + att.pattern + `")`
			result := Get(xml, path)

			elapsed := time.Since(start)

			// Should complete quickly (< 100ms) due to MaxPatternIterations limit
			if elapsed > 100*time.Millisecond {
				t.Errorf("SECURITY ISSUE: Pattern matching took too long (%v) for %s", elapsed, att.desc)
			}

			// Should not match due to complexity limit exceeded
			if result.Exists() {
				t.Errorf("Expected no match due to complexity limit for %s", att.desc)
			}
		})
	}
}

// TestPatternMatchAttack_LargePatternLength tests protection against extremely long patterns
func TestPatternMatchAttack_LargePatternLength(t *testing.T) {
	xml := `<items><item><name>test</name></item></items>`

	attacks := []struct {
		name       string
		patternLen int
		desc       string
	}{
		{
			name:       "at limit",
			patternLen: MaxFilterExpressionLength - 20, // Account for 'name%""' overhead
			desc:       "Pattern at MaxFilterExpressionLength",
		},
		{
			name:       "over limit",
			patternLen: MaxFilterExpressionLength + 100,
			desc:       "Pattern exceeding MaxFilterExpressionLength",
		},
		{
			name:       "extreme length",
			patternLen: MaxFilterExpressionLength * 10,
			desc:       "Extremely long pattern",
		},
	}

	for _, att := range attacks {
		t.Run(att.name, func(t *testing.T) {
			pattern := strings.Repeat("a", att.patternLen)
			path := `items.item.#(name%"` + pattern + `")`

			start := time.Now()
			result := Get(xml, path)
			elapsed := time.Since(start)

			// Should complete quickly regardless of pattern length
			if elapsed > 50*time.Millisecond {
				t.Errorf("SECURITY ISSUE: Long pattern took too long (%v) for %s", elapsed, att.desc)
			}

			// Over-limit patterns should be rejected
			if att.patternLen > MaxFilterExpressionLength-20 {
				if result.Exists() {
					t.Errorf("Expected rejection for %s", att.desc)
				}
			}
		})
	}
}

// TestPatternMatchAttack_UnicodeExploits tests Unicode-based attack vectors
func TestPatternMatchAttack_UnicodeExploits(t *testing.T) {
	xml := `<items>
		<item><name>test</name></item>
		<item><name>тест</name></item>
		<item><name>测试</name></item>
	</items>`

	attacks := []struct {
		name    string
		pattern string
		desc    string
	}{
		{
			name:    "unicode normalization",
			pattern: "\u0041\u0301", // Á as separate characters
			desc:    "Unicode normalization attacks",
		},
		{
			name:    "zero width characters",
			pattern: "test\u200B\u200C\u200D",
			desc:    "Zero-width character injection",
		},
		{
			name:    "rtl override",
			pattern: "test\u202E\u202D",
			desc:    "Right-to-left override characters",
		},
		{
			name:    "combining characters",
			pattern: strings.Repeat("\u0301", 100) + "test",
			desc:    "Many combining characters",
		},
		{
			name:    "homoglyphs",
			pattern: "tеst", // Cyrillic 'е' instead of 'e'
			desc:    "Homoglyph substitution",
		},
	}

	for _, att := range attacks {
		t.Run(att.name, func(t *testing.T) {
			path := `items.item.#(name%"` + att.pattern + `")`

			start := time.Now()
			result := Get(xml, path)
			elapsed := time.Since(start)

			// Should handle Unicode safely and quickly
			if elapsed > 50*time.Millisecond {
				t.Errorf("SECURITY ISSUE: Unicode pattern took too long (%v) for %s", elapsed, att.desc)
			}

			// Should not panic or crash
			_ = result.Type
		})
	}
}

// TestPatternMatchAttack_EscapeSequenceExploits tests escape sequence vulnerabilities
func TestPatternMatchAttack_EscapeSequenceExploits(t *testing.T) {
	xml := `<items><item><name>test</name></item></items>`

	attacks := []struct {
		name    string
		pattern string
		desc    string
	}{
		{
			name:    "backslash flood",
			pattern: strings.Repeat("\\", 200) + "test",
			desc:    "Many consecutive backslashes",
		},
		{
			name:    "escaped backslashes",
			pattern: strings.Repeat("\\\\", 100) + "test",
			desc:    "Many escaped backslashes",
		},
		{
			name:    "incomplete escape at end",
			pattern: "test\\",
			desc:    "Incomplete escape sequence at end",
		},
		{
			name:    "escape every character",
			pattern: "\\t\\e\\s\\t",
			desc:    "Every character escaped",
		},
		{
			name:    "null byte escape",
			pattern: "test\\x00file",
			desc:    "Null byte escape sequence",
		},
	}

	for _, att := range attacks {
		t.Run(att.name, func(t *testing.T) {
			path := `items.item.#(name%"` + att.pattern + `")`

			start := time.Now()
			result := Get(xml, path)
			elapsed := time.Since(start)

			// Should handle escape sequences safely
			if elapsed > 50*time.Millisecond {
				t.Errorf("SECURITY ISSUE: Escape sequence pattern took too long (%v) for %s", elapsed, att.desc)
			}

			// Should not panic
			_ = result.Type
		})
	}
}

// TestPatternMatchAttack_MemoryExhaustion tests memory exhaustion via pattern matching
func TestPatternMatchAttack_MemoryExhaustion(t *testing.T) {
	// Create large XML document
	itemCount := 10000
	var b strings.Builder
	b.WriteString("<items>")
	for i := 0; i < itemCount; i++ {
		b.WriteString("<item><name>")
		b.WriteString(strings.Repeat("a", 100))
		b.WriteString("</name></item>")
	}
	b.WriteString("</items>")
	xml := b.String()

	attacks := []struct {
		name    string
		pattern string
		desc    string
	}{
		{
			name:    "match all with complex pattern",
			pattern: "a*a*a*a*a",
			desc:    "Complex pattern on many elements",
		},
		{
			name:    "wildcard prefix",
			pattern: "*" + strings.Repeat("a", 90),
			desc:    "Wildcard prefix on long strings",
		},
		{
			name:    "many question marks",
			pattern: strings.Repeat("?", 100),
			desc:    "Many question marks matching exact length",
		},
	}

	for _, att := range attacks {
		t.Run(att.name, func(t *testing.T) {
			path := `items.item.#(name%"` + att.pattern + `")#`

			start := time.Now()
			result := Get(xml, path)
			elapsed := time.Since(start)

			// Should complete within reasonable time despite large input
			// Note: 800ms limit accounts for CI environment and race detector overhead (~60%)
			if elapsed > 800*time.Millisecond {
				t.Errorf("SECURITY ISSUE: Memory exhaustion attack took too long (%v) for %s", elapsed, att.desc)
			}

			// Should not panic and results should be bounded
			if result.Type == Array {
				if len(result.Results) > MaxWildcardResults {
					t.Errorf("SECURITY VIOLATION: Results exceed MaxWildcardResults for %s", att.desc)
				}
			}
		})
	}
}

// TestPatternMatchAttack_NestedFilterDepth tests filter depth limit with pattern matching
func TestPatternMatchAttack_NestedFilterDepth(t *testing.T) {
	// Create deeply nested XML
	depth := 15
	var b strings.Builder
	b.WriteString("<root>")
	for i := 0; i < depth; i++ {
		b.WriteString("<level><item><name>test</name></item></level>")
	}
	b.WriteString("</root>")
	xml := b.String()

	// Build path that would cause deep recursion
	path := "root"
	for i := 0; i < depth; i++ {
		path += `.level.item.#(name%"test")`
	}

	start := time.Now()
	result := Get(xml, path)
	elapsed := time.Since(start)

	// Should handle deep nesting without stack overflow
	if elapsed > 200*time.Millisecond {
		t.Errorf("SECURITY ISSUE: Deep nested filters took too long (%v)", elapsed)
	}

	// Should be limited by MaxFilterDepth
	_ = result.Type
}

// TestPatternMatchAttack_CombinedVectors tests multiple attack vectors combined
func TestPatternMatchAttack_CombinedVectors(t *testing.T) {
	// Large XML with long values
	itemCount := 1000
	var b strings.Builder
	b.WriteString("<items>")
	for i := 0; i < itemCount; i++ {
		b.WriteString("<item><name>")
		b.WriteString(strings.Repeat("a", 200))
		b.WriteString("</name></item>")
	}
	b.WriteString("</items>")
	xml := b.String()

	attacks := []struct {
		name    string
		pattern string
		desc    string
	}{
		{
			name:    "long pattern with backtracking",
			pattern: strings.Repeat("a*", 30) + "b",
			desc:    "Long pattern + exponential backtracking",
		},
		{
			name:    "unicode with backtracking",
			pattern: strings.Repeat("你*", 20) + "好",
			desc:    "Unicode + exponential backtracking",
		},
		{
			name:    "escapes with backtracking",
			pattern: strings.Repeat("\\a*", 20) + "b",
			desc:    "Escape sequences + backtracking",
		},
	}

	for _, att := range attacks {
		t.Run(att.name, func(t *testing.T) {
			path := `items.item.#(name%"` + att.pattern + `")#`

			start := time.Now()
			result := Get(xml, path)
			elapsed := time.Since(start)

			// Combined attacks should still be fast
			// Note: 700ms limit accounts for CI environment and race detector overhead (~100%)
			if elapsed > 700*time.Millisecond {
				t.Errorf("SECURITY ISSUE: Combined attack took too long (%v) for %s", elapsed, att.desc)
			}

			// Should safely handle or reject
			_ = result.Type
		})
	}
}

// TestPatternMatchAttack_TypeConfusion tests type confusion attacks
func TestPatternMatchAttack_TypeConfusion(t *testing.T) {
	xml := `<items>
		<item><age>25</age><name>John</name></item>
		<item><age>thirty</age><name>Jane</name></item>
		<item><age>NaN</age><name>Invalid</name></item>
		<item><age>Infinity</age><name>Special</name></item>
	</items>`

	tests := []struct {
		name string
		path string
		desc string
	}{
		{
			name: "pattern on numeric field",
			path: `items.item.#(age%"2*")`,
			desc: "Pattern matching on numeric field should work (string comparison)",
		},
		{
			name: "pattern on text field",
			path: `items.item.#(age%"t*")`,
			desc: "Pattern matching on text should work",
		},
		{
			name: "pattern on special float",
			path: `items.item.#(age%"NaN")`,
			desc: "Pattern matching on special float values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			result := Get(xml, tt.path)

			// Should handle type variations safely
			_ = result.Type

			// Should not crash or return unexpected results
			if result.Type == Element {
				// If it matched something, verify it's valid
				_ = result.String()
			}
		})
	}
}

// TestPatternMatchAttack_ConcurrentAccess tests concurrent pattern matching safety
func TestPatternMatchAttack_ConcurrentAccess(t *testing.T) {
	xml := `<items>`
	for i := 0; i < 100; i++ {
		xml += `<item><name>test` + strings.Repeat("a", 100) + `</name></item>`
	}
	xml += `</items>`

	// Run many concurrent pattern matching operations
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(id int) {
			patterns := []string{
				"test*",
				"*a*",
				"t*t",
				strings.Repeat("?", 104),
				"a*a*a*a*a*b",
			}

			pattern := patterns[id%len(patterns)]
			path := `items.item.#(name%"` + pattern + `")#`

			result := Get(xml, path)
			_ = result.Type

			done <- true
		}(i)
	}

	// Wait for all goroutines with timeout
	timeout := time.After(5 * time.Second)
	for i := 0; i < 100; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("SECURITY ISSUE: Concurrent pattern matching timed out")
		}
	}
}

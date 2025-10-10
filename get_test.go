// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"testing"
)

// Test basic element access
func TestGet_BasicElement(t *testing.T) {
	xml := `<root><user><name>John</name><age>30</age></user></root>`

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Single nested element",
			path: "root.user.name",
			want: "John",
		},
		{
			name: "Another nested element",
			path: "root.user.age",
			want: "30",
		},
		{
			name: "Missing path",
			path: "root.user.email",
			want: "",
		},
		{
			name: "Partial path",
			path: "root.user",
			want: "John30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			got := result.String()
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// Test deeply nested elements
func TestGet_DeeplyNested(t *testing.T) {
	xml := `<root><level1><level2><level3><level4>value</level4></level3></level2></level1></root>`

	result := Get(xml, "root.level1.level2.level3.level4")
	if got := result.String(); got != "value" {
		t.Errorf("Get() = %q, want %q", got, "value")
	}
}

// Test attribute access
func TestGet_Attributes(t *testing.T) {
	xml := `<root><user id="123" active="true"><name>John</name></user></root>`

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Attribute access",
			path: "root.user.@id",
			want: "123",
		},
		{
			name: "Another attribute",
			path: "root.user.@active",
			want: "true",
		},
		{
			name: "Missing attribute",
			path: "root.user.@missing",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			got := result.String()
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.path, got, tt.want)
			}
			if tt.want != "" && result.Type != Attribute {
				t.Errorf("Result type = %v, want Attribute", result.Type)
			}
		})
	}
}

// Test text content extraction
func TestGet_TextContent(t *testing.T) {
	xml := `<item>
		This is text
		<note>A note</note>
		More text
	</item>`

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Full content with tags",
			path: "item",
			want: "This is text A note More text",
		},
		{
			name: "Text only",
			path: "item.%",
			want: "This is text More text",
		},
		{
			name: "Nested element",
			path: "item.note",
			want: "A note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			got := result.String()
			// Normalize whitespace for comparison
			got = normalizeWhitespace(got)
			want := normalizeWhitespace(tt.want)
			if got != want {
				t.Errorf("Get(%q) = %q, want %q", tt.path, got, want)
			}
		})
	}
}

// Test array handling
func TestGet_Arrays(t *testing.T) {
	xml := `<root>
		<users>
			<user><name>Alice</name></user>
			<user><name>Bob</name></user>
			<user><name>Carol</name></user>
		</users>
	</root>`

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "First element",
			path: "root.users.user.0.name",
			want: "Alice",
		},
		{
			name: "Second element",
			path: "root.users.user.1.name",
			want: "Bob",
		},
		{
			name: "Third element",
			path: "root.users.user.2.name",
			want: "Carol",
		},
		{
			name: "Out of bounds",
			path: "root.users.user.10.name",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			got := result.String()
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// Test array count
func TestGet_ArrayCount(t *testing.T) {
	xml := `<root>
		<users>
			<user><name>Alice</name></user>
			<user><name>Bob</name></user>
			<user><name>Carol</name></user>
		</users>
	</root>`

	result := Get(xml, "root.users.user.#")
	if result.Type != Number {
		t.Errorf("Result type = %v, want Number", result.Type)
	}
	if got := result.Int(); got != 3 {
		t.Errorf("Array count = %d, want 3", got)
	}
}

// Test empty array
func TestGet_EmptyArray(t *testing.T) {
	xml := `<root><users></users></root>`

	result := Get(xml, "root.users.user.#")
	if got := result.Int(); got != 0 {
		t.Errorf("Empty array count = %d, want 0", got)
	}
}

// Test self-closing tags
func TestGet_SelfClosing(t *testing.T) {
	xml := `<root><empty/><item value="test"/></root>`

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Self-closing element",
			path: "root.empty",
			want: "",
		},
		{
			name: "Self-closing with attribute",
			path: "root.item.@value",
			want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			got := result.String()
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// Test XML escaping
func TestGet_XMLEscaping(t *testing.T) {
	xml := `<root><text>&lt;tag&gt; &amp; &quot;quotes&quot;</text></root>`

	result := Get(xml, "root.text")
	want := "<tag> & \"quotes\""
	if got := result.String(); got != want {
		t.Errorf("Get() = %q, want %q", got, want)
	}
}

// Test malformed XML (graceful handling)
func TestGet_MalformedXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
	}{
		{
			name: "Unclosed tag",
			xml:  `<root><user>`,
			path: "root.user",
		},
		{
			name: "Missing closing bracket",
			xml:  `<root><user</root>`,
			path: "root.user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Should not panic
			result := Get(tt.xml, tt.path)
			_ = result.String()
		})
	}
}

// Test GetBytes
func TestGetBytes(t *testing.T) {
	xml := []byte(`<root><user><name>John</name></user></root>`)

	result := GetBytes(xml, "root.user.name")
	if got := result.String(); got != "John" {
		t.Errorf("GetBytes() = %q, want %q", got, "John")
	}
}

// Test GetMany
func TestGetMany(t *testing.T) {
	xml := `<root>
		<user id="123">
			<name>John</name>
			<age>30</age>
		</user>
	</root>`

	results := GetMany(xml,
		"root.user.name",
		"root.user.age",
		"root.user.@id",
	)

	if len(results) != 3 {
		t.Fatalf("GetMany() returned %d results, want 3", len(results))
	}

	wants := []string{"John", "30", "123"}
	for i, want := range wants {
		if got := results[i].String(); got != want {
			t.Errorf("GetMany() result[%d] = %q, want %q", i, got, want)
		}
	}
}

// Test complex real-world-like XML
func TestGet_ComplexXML(t *testing.T) {
	xml := `<?xml version="1.0"?>
	<catalog>
		<book id="bk101" category="programming">
			<author>Gambardella, Matthew</author>
			<title>XML Developer's Guide</title>
			<genre>Computer</genre>
			<price>44.95</price>
			<publish_date>2000-10-01</publish_date>
		</book>
		<book id="bk102" category="fiction">
			<author>Ralls, Kim</author>
			<title>Midnight Rain</title>
			<genre>Fantasy</genre>
			<price>5.95</price>
			<publish_date>2000-12-16</publish_date>
		</book>
	</catalog>`

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "First book title",
			path: "catalog.book.0.title",
			want: "XML Developer's Guide",
		},
		{
			name: "Second book author",
			path: "catalog.book.1.author",
			want: "Ralls, Kim",
		},
		{
			name: "Book count",
			path: "catalog.book.#",
			want: "2",
		},
		{
			name: "First book category attribute",
			path: "catalog.book.0.@category",
			want: "programming",
		},
		{
			name: "First book price",
			path: "catalog.book.0.price",
			want: "44.95",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			got := result.String()
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// Test Exists method on results
func TestGet_Exists(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	tests := []struct {
		name   string
		path   string
		exists bool
	}{
		{
			name:   "Existing path",
			path:   "root.user.name",
			exists: true,
		},
		{
			name:   "Missing path",
			path:   "root.user.email",
			exists: false,
		},
		{
			name:   "Partial path",
			path:   "root.user",
			exists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if got := result.Exists(); got != tt.exists {
				t.Errorf("Get(%q).Exists() = %v, want %v", tt.path, got, tt.exists)
			}
		})
	}
}

// Helper function to normalize whitespace for comparison
func normalizeWhitespace(s string) string {
	// Simple normalization: collapse multiple spaces/tabs/newlines to single space
	var result []byte
	lastWasSpace := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if !lastWasSpace {
				result = append(result, ' ')
				lastWasSpace = true
			}
		} else {
			result = append(result, c)
			lastWasSpace = false
		}
	}
	// Trim leading/trailing spaces
	str := string(result)
	if len(str) > 0 && str[0] == ' ' {
		str = str[1:]
	}
	if len(str) > 0 && str[len(str)-1] == ' ' {
		str = str[:len(str)-1]
	}
	return str
}

// Example functions for godoc

// ExampleGet demonstrates basic element access using path syntax
func ExampleGet() {
	xml := `<catalog>
		<book>
			<title>The Go Programming Language</title>
			<author>Alan Donovan</author>
			<price>44.99</price>
		</book>
	</catalog>`

	// Access nested element
	title := Get(xml, "catalog.book.title")
	fmt.Println(title.String())
	// Output: The Go Programming Language
}

// ExampleGet_attribute demonstrates attribute access using @ syntax
func ExampleGet_attribute() {
	xml := `<book id="123" category="programming">
		<title>Learning Go</title>
	</book>`

	// Access attribute using @
	id := Get(xml, "book.@id")
	fmt.Println(id.String())

	category := Get(xml, "book.@category")
	fmt.Println(category.String())
	// Output:
	// 123
	// programming
}

// ExampleGet_textContent demonstrates text content extraction using % operator
func ExampleGet_textContent() {
	xml := `<item>This is direct text<note>A nested note</note>More direct text</item>`

	// Get all text content (including nested elements)
	allText := Get(xml, "item")
	fmt.Println(allText.String())

	// Get only direct text content (excluding nested elements) using %
	directText := Get(xml, "item.%")
	fmt.Println(directText.String())
	// Output:
	// This is direct textA nested noteMore direct text
	// This is direct textMore direct text
}

// ExampleGet_arrayIndex demonstrates array indexing
func ExampleGet_arrayIndex() {
	xml := `<catalog>
		<book><title>First Book</title></book>
		<book><title>Second Book</title></book>
		<book><title>Third Book</title></book>
	</catalog>`

	// Access first book (index 0)
	first := Get(xml, "catalog.book.0.title")
	fmt.Println(first.String())

	// Access second book (index 1)
	second := Get(xml, "catalog.book.1.title")
	fmt.Println(second.String())

	// Access third book (index 2)
	third := Get(xml, "catalog.book.2.title")
	fmt.Println(third.String())
	// Output:
	// First Book
	// Second Book
	// Third Book
}

// ExampleGet_arrayCount demonstrates array counting using # operator
func ExampleGet_arrayCount() {
	xml := `<users>
		<user><name>Alice</name></user>
		<user><name>Bob</name></user>
		<user><name>Carol</name></user>
	</users>`

	// Count number of user elements
	count := Get(xml, "users.user.#")
	fmt.Printf("Total users: %d\n", count.Int())
	// Output: Total users: 3
}

// ExampleGetMany demonstrates efficient multi-path querying
func ExampleGetMany() {
	xml := `<book id="123">
		<title>Go Programming</title>
		<author>John Doe</author>
		<price>39.99</price>
	</book>`

	// Get multiple paths in one call
	results := GetMany(xml,
		"book.title",
		"book.author",
		"book.price",
		"book.@id",
	)

	fmt.Println("Title:", results[0].String())
	fmt.Println("Author:", results[1].String())
	fmt.Printf("Price: $%.2f\n", results[2].Float())
	fmt.Println("ID:", results[3].String())
	// Output:
	// Title: Go Programming
	// Author: John Doe
	// Price: $39.99
	// ID: 123
}

// ExampleResult_Exists demonstrates checking if a path exists
func ExampleResult_Exists() {
	xml := `<user>
		<name>John</name>
		<age>30</age>
	</user>`

	// Check if elements exist
	name := Get(xml, "user.name")
	fmt.Println("Name exists:", name.Exists())

	email := Get(xml, "user.email")
	fmt.Println("Email exists:", email.Exists())
	// Output:
	// Name exists: true
	// Email exists: false
}

// ExampleGet_wildcard demonstrates single-level wildcard queries
func ExampleGet_wildcard() {
	xml := `<store>
		<electronics>
			<name>Electronics Department</name>
		</electronics>
		<clothing>
			<name>Clothing Department</name>
		</clothing>
		<food>
			<name>Food Department</name>
		</food>
	</store>`

	// Get all department names using wildcard
	result := Get(xml, "store.*.name")
	for _, dept := range result.Array() {
		fmt.Println(dept.String())
	}
	// Output:
	// Electronics Department
	// Clothing Department
	// Food Department
}

// ExampleGet_recursiveWildcard demonstrates recursive wildcard queries
func ExampleGet_recursiveWildcard() {
	xml := `<catalog>
		<category>
			<item><price>10.99</price></item>
			<subcategory>
				<item><price>25.50</price></item>
			</subcategory>
		</category>
	</catalog>`

	// Find all prices at any depth using recursive wildcard
	result := Get(xml, "catalog.**.price")
	for _, price := range result.Array() {
		fmt.Printf("$%.2f\n", price.Float())
	}
	// Output:
	// $10.99
	// $25.50
}

// ExampleGet_filter demonstrates filtering elements by conditions
func ExampleGet_filter() {
	xml := `<products>
		<product>
			<name>Widget A</name>
			<price>15.99</price>
		</product>
		<product>
			<name>Widget B</name>
			<price>25.99</price>
		</product>
		<product>
			<name>Widget C</name>
			<price>35.99</price>
		</product>
	</products>`

	// Filter products with price greater than 20 - use #()# for all matches
	result := Get(xml, "products.product.#(price>20)#.name")
	for _, name := range result.Array() {
		fmt.Println(name.String())
	}
	// Output:
	// Widget B
	// Widget C
}

// ExampleGet_modifier demonstrates using result modifiers
func ExampleGet_modifier() {
	xml := `<item>apple</item>`

	// Get item and apply uppercase modifier (after registering it)
	upperMod := NewModifierFunc("testupper", func(r Result) Result {
		return Result{Type: r.Type, Str: strings.ToUpper(r.Str), Raw: r.Raw}
	})
	_ = RegisterModifier("testupper", upperMod)
	defer func() { _ = UnregisterModifier("testupper") }()

	result := Get(xml, "item|@testupper")
	fmt.Println(result.String())
	// Output: APPLE
}

// ExampleGetWithOptions demonstrates using options for case-insensitive queries
func ExampleGetWithOptions() {
	xml := `<Document>
		<Person>
			<Name>John Doe</Name>
		</Person>
	</Document>`

	// Case-insensitive query (matches despite different casing)
	opts := Options{CaseSensitive: false}
	result := GetWithOptions(xml, "document.person.name", &opts)
	fmt.Println(result.String())
	// Output: John Doe
}

// ============================================================================
// Field Extraction Tests (#.field syntax)
// ============================================================================

// TestFieldExtractionBasic tests basic #.field extraction
func TestFieldExtractionBasic(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Alice</name><age>30</age></item>
			<item><name>Bob</name><age>25</age></item>
			<item><name>Carol</name><age>35</age></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}
	if result.Results[0].String() != "Alice" {
		t.Errorf("Expected Alice, got %s", result.Results[0].String())
	}
	if result.Results[1].String() != "Bob" {
		t.Errorf("Expected Bob, got %s", result.Results[1].String())
	}
	if result.Results[2].String() != "Carol" {
		t.Errorf("Expected Carol, got %s", result.Results[2].String())
	}
}

// TestFieldExtractionAttribute tests #.@attr extraction
func TestFieldExtractionAttribute(t *testing.T) {
	xml := `<root>
		<items>
			<item id="1" name="First"/>
			<item id="2" name="Second"/>
			<item id="3" name="Third"/>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.@id")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}
	if result.Results[0].String() != "1" {
		t.Errorf("Expected 1, got %s", result.Results[0].String())
	}
	if result.Results[1].String() != "2" {
		t.Errorf("Expected 2, got %s", result.Results[1].String())
	}
}

// TestFieldExtractionText tests #.% (text content) extraction
func TestFieldExtractionText(t *testing.T) {
	xml := `<root>
		<items>
			<item>Text1<sub>Nested</sub></item>
			<item>Text2</item>
			<item>Text3<sub>More</sub></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.%")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// Text-only extraction should get direct text, not nested elements
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionEmpty tests #.field with empty array
func TestFieldExtractionEmpty(t *testing.T) {
	xml := `<root><items></items></root>`

	result := Get(xml, "root.items.item.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionSingleElement tests #.field with single element
func TestFieldExtractionSingleElement(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>OnlyOne</name></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].String() != "OnlyOne" {
		t.Errorf("Expected OnlyOne, got %s", result.Results[0].String())
	}
}

// TestFieldExtractionMissingField tests #.field when field doesn't exist
func TestFieldExtractionMissingField(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Alice</name></item>
			<item><name>Bob</name></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.age")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// No matching fields = empty array
	if len(result.Results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionNonExistentPath tests #.field with non-existent path
func TestFieldExtractionNonExistentPath(t *testing.T) {
	xml := `<root><items></items></root>`

	result := Get(xml, "root.nonexistent.item.#.name")
	// When the path doesn't exist before field extraction, we get Null, not Array
	if result.Type != Null && result.Type != Array {
		t.Errorf("Expected Null or Array type, got %v", result.Type)
	}
	if result.Type == Array && len(result.Results) != 0 {
		t.Errorf("Expected 0 results if Array, got %d", len(result.Results))
	}
}

// TestFieldExtractionPartialMatch tests #.field when some elements have field, others don't
func TestFieldExtractionPartialMatch(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Alice</name><age>30</age></item>
			<item><name>Bob</name></item>
			<item><name>Carol</name><age>35</age></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.age")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// Only 2 items have age
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0].String() != "30" {
		t.Errorf("Expected 30, got %s", result.Results[0].String())
	}
	if result.Results[1].String() != "35" {
		t.Errorf("Expected 35, got %s", result.Results[1].String())
	}
}

// TestFieldExtractionNested tests #.#.field (nested extraction)
func TestFieldExtractionNested(t *testing.T) {
	xml := `<root>
		<categories>
			<category>
				<products>
					<product><price>10</price></product>
					<product><price>20</price></product>
				</products>
			</category>
			<category>
				<products>
					<product><price>30</price></product>
				</products>
			</category>
		</categories>
	</root>`

	// Extract all prices from all products across all categories
	result := Get(xml, "root.categories.category.#.products")
	if result.Type != Array {
		t.Errorf("Expected Array type for categories, got %v", result.Type)
	}

	// Now extract prices from each products element
	// This requires iterating and doing nested extraction
	// For now, just test that we can get products
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 category results, got %d", len(result.Results))
	}
}

// TestFieldExtractionWithFilter tests combining #(condition) with #.field
func TestFieldExtractionWithFilter(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Alice</name><age>30</age></item>
			<item><name>Bob</name><age>25</age></item>
			<item><name>Carol</name><age>35</age></item>
		</items>
	</root>`

	// Test: Filter first match + field extraction
	// Pattern: items.item.#(age>28).#.name
	// Should return: First name where age>28 (Alice)
	result := Get(xml, "root.items.item.#(age>28).#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type for first match field extraction, got %v", result.Type)
	}
	// First match filter returns single element, field extraction should extract all matching fields from that element
	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result (first match), got %d", len(result.Results))
	}
	if len(result.Results) > 0 && result.Results[0].String() != "Alice" {
		t.Errorf("Expected Alice (first match with age>28), got %s", result.Results[0].String())
	}
}

// TestFieldExtractionFilterAll tests combining #(condition)# (all matches) with #.field
func TestFieldExtractionFilterAll(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Alice</name><age>30</age></item>
			<item><name>Bob</name><age>25</age></item>
			<item><name>Carol</name><age>35</age></item>
		</items>
	</root>`

	// Test: Filter all matches + field extraction
	// Pattern: items.item.#(age>28)#.#.name
	// Should return: All names where age>28 (Alice and Carol)
	result := Get(xml, "root.items.item.#(age>28)#.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type for all matches field extraction, got %v", result.Type)
	}
	// Should get names from both items where age>28 (Alice and Carol)
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results (all matches with age>28), got %d", len(result.Results))
	}
	if len(result.Results) >= 2 {
		if result.Results[0].String() != "Alice" {
			t.Errorf("Expected Alice as first result, got %s", result.Results[0].String())
		}
		if result.Results[1].String() != "Carol" {
			t.Errorf("Expected Carol as second result, got %s", result.Results[1].String())
		}
	}
}

// TestFieldExtractionFilterAllAttributes tests #(condition)#.#.@attr pattern
func TestFieldExtractionFilterAllAttributes(t *testing.T) {
	xml := `<root>
		<items>
			<item id="1"><name>Alice</name><score>95</score></item>
			<item id="2"><name>Bob</name><score>75</score></item>
			<item id="3"><name>Carol</name><score>88</score></item>
		</items>
	</root>`

	// Test: Filter all matches + element extraction
	// Pattern: items.item.#(score>80)#.#.name
	// Should return: Names of items with score>80 (Alice and Carol)
	result := Get(xml, "root.items.item.#(score>80)#.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
	if len(result.Results) >= 2 {
		if result.Results[0].String() != "Alice" {
			t.Errorf("Expected Alice, got %s", result.Results[0].String())
		}
		if result.Results[1].String() != "Carol" {
			t.Errorf("Expected Carol, got %s", result.Results[1].String())
		}
	}
}

// TestFieldExtractionFilterAllAttributeExtraction tests #(condition)#.@attr pattern (attribute on matched element)
func TestFieldExtractionFilterAllAttributeExtraction(t *testing.T) {
	xml := `<root>
		<items>
			<item id="1" name="Alice"><score>95</score></item>
			<item id="2" name="Bob"><score>75</score></item>
			<item id="3" name="Carol"><score>88</score></item>
		</items>
	</root>`

	// Test: Filter all matches + attribute extraction from matched elements
	// Pattern: items.item.#(score>80)#.@name
	// Should return: @name attributes of items with score>80 (Alice and Carol)
	result := Get(xml, "root.items.item.#(score>80)#.@name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
	if len(result.Results) >= 2 {
		if result.Results[0].String() != "Alice" {
			t.Errorf("Expected Alice, got %s", result.Results[0].String())
		}
		if result.Results[1].String() != "Carol" {
			t.Errorf("Expected Carol, got %s", result.Results[1].String())
		}
	}
}

// TestFieldExtractionWithModifiers tests #.field with modifiers
func TestFieldExtractionWithModifiers(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Charlie</name></item>
			<item><name>Alice</name></item>
			<item><name>Bob</name></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.name|@sort")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}
	// Should be sorted alphabetically
	if result.Results[0].String() != "Alice" {
		t.Errorf("Expected Alice first after sort, got %s", result.Results[0].String())
	}
	if result.Results[1].String() != "Bob" {
		t.Errorf("Expected Bob second after sort, got %s", result.Results[1].String())
	}
	if result.Results[2].String() != "Charlie" {
		t.Errorf("Expected Charlie third after sort, got %s", result.Results[2].String())
	}
}

// TestFieldExtractionMultipleValues tests when field appears multiple times in one element
func TestFieldExtractionMultipleValues(t *testing.T) {
	xml := `<root>
		<items>
			<item><tag>tag1</tag><tag>tag2</tag></item>
			<item><tag>tag3</tag></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.tag")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// Should extract ALL tag elements (3 total)
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionCDATA tests #.field with CDATA content
func TestFieldExtractionCDATA(t *testing.T) {
	xml := `<root>
		<items>
			<item><data><![CDATA[Value1]]></data></item>
			<item><data><![CDATA[Value2]]></data></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.data")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionNamespace tests #.field with namespaced elements
func TestFieldExtractionNamespace(t *testing.T) {
	xml := `<root xmlns:ns="http://example.com">
		<items>
			<item><ns:name>Alice</ns:name></item>
			<item><ns:name>Bob</ns:name></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.ns:name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionSecurity tests MaxWildcardResults limit
func TestFieldExtractionSecurity(t *testing.T) {
	// Create XML with many items
	xml := `<root><items>`
	for i := 0; i < 1500; i++ {
		xml += `<item><name>Item</name></item>`
	}
	xml += `</items></root>`

	result := Get(xml, "root.items.item.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// Should be limited to MaxWildcardResults (1000)
	if len(result.Results) != MaxWildcardResults {
		t.Errorf("Expected %d results (limit), got %d", MaxWildcardResults, len(result.Results))
	}
}

// TestFieldExtractionFieldNameValidation tests field name validation
func TestFieldExtractionFieldNameValidation(t *testing.T) {
	xml := `<root><items><item><name>Test</name></item></items></root>`

	// Test various valid field names
	validNames := []string{
		"root.items.item.#.name",     // Simple name
		"root.items.item.#.@attr",    // Attribute
		"root.items.item.#.%",        // Text content
		"root.items.item.#.ns:name",  // Namespace prefix
		"root.items.item.#.my-field", // Hyphen
		"root.items.item.#.my_field", // Underscore
		"root.items.item.#._field",   // Leading underscore
	}

	for _, path := range validNames {
		result := Get(xml, path)
		// All should parse without error (may return empty array if field doesn't exist)
		if result.Type != Array && result.Type != Null {
			t.Errorf("Path %s should parse correctly", path)
		}
	}
}

// TestFieldExtractionLongFieldName tests MaxFieldNameLength security limit
func TestFieldExtractionLongFieldName(t *testing.T) {
	xml := `<root><items><item><name>Test</name></item></items></root>`

	// Create a field name longer than MaxFieldNameLength
	longName := "field"
	for len(longName) < MaxFieldNameLength+10 {
		longName += "x"
	}

	path := "root.items.item.#." + longName
	result := Get(xml, path)

	// Long field names are rejected during parsing, resulting in # (count) segment
	// which returns a number (count of items), not field extraction
	// This is the expected secure behavior - silently reject invalid segments
	if result.Type == Array {
		// If somehow it got through as Array, it should be empty
		if len(result.Results) != 0 {
			t.Errorf("Long field name should not extract values")
		}
	}
	// Otherwise the # becomes a count operation, which is fine (returns Number type)
}

// TestFieldExtractionWithOptions tests field extraction with case-insensitive option
func TestFieldExtractionWithOptions(t *testing.T) {
	xml := `<root>
		<items>
			<item><Name>Alice</Name></item>
			<item><NAME>Bob</NAME></item>
			<item><name>Carol</name></item>
		</items>
	</root>`

	opts := Options{CaseSensitive: false}
	result := GetWithOptions(xml, "root.items.item.#.name", &opts)

	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// Should match all three (Name, NAME, name) with case-insensitive search
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results with case-insensitive, got %d", len(result.Results))
	}
}

// TestFieldExtractionAttributeWithOptions tests attribute extraction with case-insensitive option
func TestFieldExtractionAttributeWithOptions(t *testing.T) {
	xml := `<root>
		<items>
			<item Id="1"/>
			<item ID="2"/>
			<item id="3"/>
		</items>
	</root>`

	opts := Options{CaseSensitive: false}
	result := GetWithOptions(xml, "root.items.item.#.@id", &opts)

	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// Should match all three with case-insensitive attribute search
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionChainedModifiers tests multiple modifiers on field extraction
func TestFieldExtractionChainedModifiers(t *testing.T) {
	xml := `<root>
		<items>
			<item><value>10</value></item>
			<item><value>5</value></item>
			<item><value>20</value></item>
			<item><value>15</value></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.value|@sort|@reverse|@first")
	// Should sort, reverse (descending), then take first (highest value)
	if result.String() != "20" {
		t.Errorf("Expected 20 (highest value), got %s", result.String())
	}
}

// TestFieldExtractionMixedContent tests field extraction with mixed text and element content
func TestFieldExtractionMixedContent(t *testing.T) {
	xml := `<root>
		<items>
			<item>
				<data>Text <b>bold</b> more</data>
			</item>
			<item>
				<data>Plain text</data>
			</item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.data")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
	// First result should include all text content
	if result.Results[0].String() == "" {
		t.Errorf("Expected non-empty text content")
	}
}

// TestFieldExtractionSelfClosingElements tests field extraction from self-closing elements
func TestFieldExtractionSelfClosingElements(t *testing.T) {
	xml := `<root>
		<items>
			<item id="1"><name/></item>
			<item id="2"><name/></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// Should extract both self-closing name elements
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionCount tests combining #.field with array count
func TestFieldExtractionCount(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Alice</name></item>
			<item><name>Bob</name></item>
			<item><name>Carol</name></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}

	// Use Array() to access results and check count
	arr := result.Array()
	if len(arr) != 3 {
		t.Errorf("Expected 3 items in array, got %d", len(arr))
	}
}

// TestFieldExtractionForEach tests iterating over extracted fields
func TestFieldExtractionForEach(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Alice</name></item>
			<item><name>Bob</name></item>
			<item><name>Carol</name></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.name")

	count := 0
	names := []string{}
	result.ForEach(func(_ int, r Result) bool {
		count++
		names = append(names, r.String())
		return true
	})

	if count != 3 {
		t.Errorf("Expected 3 iterations, got %d", count)
	}
	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}
	if names[0] != "Alice" || names[1] != "Bob" || names[2] != "Carol" {
		t.Errorf("Names don't match expected: %v", names)
	}
}

// TestFieldExtractionArrayIndexing tests accessing specific index from extracted fields
func TestFieldExtractionArrayIndexing(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>Alice</name></item>
			<item><name>Bob</name></item>
			<item><name>Carol</name></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.name")

	// Access specific indexes
	if result.Results[0].String() != "Alice" {
		t.Errorf("Expected Alice at index 0, got %s", result.Results[0].String())
	}
	if result.Results[2].String() != "Carol" {
		t.Errorf("Expected Carol at index 2, got %s", result.Results[2].String())
	}
}

// TestFieldExtractionEmptyFieldContent tests extraction when field exists but is empty
func TestFieldExtractionEmptyFieldContent(t *testing.T) {
	xml := `<root>
		<items>
			<item><name></name></item>
			<item><name>Bob</name></item>
		</items>
	</root>`

	result := Get(xml, "root.items.item.#.name")
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	// Both elements should be extracted (even if first is empty)
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
}

// Benchmarks

// BenchmarkFieldExtractionSimple tests performance of simple field extraction
func BenchmarkFieldExtractionSimple(b *testing.B) {
	xml := `<root>
		<items>
			<item><name>Alice</name><age>30</age></item>
			<item><name>Bob</name><age>25</age></item>
			<item><name>Carol</name><age>35</age></item>
			<item><name>Dave</name><age>28</age></item>
			<item><name>Eve</name><age>32</age></item>
		</items>
	</root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Get(xml, "root.items.item.#.name")
		if result.Type != Array {
			b.Fatal("Expected Array type")
		}
	}
}

// BenchmarkFieldExtractionNested tests performance of nested field extraction
func BenchmarkFieldExtractionNested(b *testing.B) {
	xml := `<root>
		<categories>
			<category>
				<products>
					<product><price>10</price><name>P1</name></product>
					<product><price>20</price><name>P2</name></product>
				</products>
			</category>
			<category>
				<products>
					<product><price>30</price><name>P3</name></product>
					<product><price>40</price><name>P4</name></product>
				</products>
			</category>
		</categories>
	</root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Get(xml, "root.categories.category.#.products")
		if result.Type != Array {
			b.Fatal("Expected Array type")
		}
	}
}

// BenchmarkFieldExtractionLarge tests performance with large arrays
func BenchmarkFieldExtractionLarge(b *testing.B) {
	// Create XML with 100 items
	xml := `<root><items>`
	for i := 0; i < 100; i++ {
		xml += `<item><name>Name</name><value>123</value></item>`
	}
	xml += `</items></root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Get(xml, "root.items.item.#.name")
		if result.Type != Array {
			b.Fatal("Expected Array type")
		}
		if len(result.Results) != 100 {
			b.Fatalf("Expected 100 results, got %d", len(result.Results))
		}
	}
}

// TestFieldExtractionAfterWildcard tests field extraction after single-level wildcard
func TestFieldExtractionAfterWildcard(t *testing.T) {
	xml := `<root>
		<category>
			<item><name>A</name><value>1</value></item>
			<item><name>B</name><value>2</value></item>
		</category>
		<category>
			<item><name>C</name><value>3</value></item>
		</category>
	</root>`

	// Test: wildcard then field extraction
	// This should get all <item> elements under any <category>, then extract <name> from each
	result := Get(xml, "root.*.item.#.name")

	// The wildcard matches both category elements
	// Each category has items, and we extract names from all items
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}

	// Should get names: A, B, C
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}

	expectedNames := []string{"A", "B", "C"}
	for i, expected := range expectedNames {
		if i >= len(result.Results) {
			break
		}
		if result.Results[i].String() != expected {
			t.Errorf("Expected result[%d] = %s, got %s", i, expected, result.Results[i].String())
		}
	}
}

// TestFieldExtractionAfterRecursiveWildcard tests field extraction after recursive wildcard
func TestFieldExtractionAfterRecursiveWildcard(t *testing.T) {
	xml := `<root>
		<a>
			<item><name>A</name></item>
		</a>
		<b>
			<c>
				<item><name>B</name></item>
			</c>
		</b>
		<d>
			<e>
				<f>
					<item><name>C</name></item>
				</f>
			</e>
		</d>
	</root>`

	// Test: recursive wildcard then field extraction
	// This should find all <item> elements at any depth, then extract <name> from each
	result := Get(xml, "root.**.item.#.name")

	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}

	// Should find all three items and extract their names
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}

	expectedNames := []string{"A", "B", "C"}
	for i, expected := range expectedNames {
		if i >= len(result.Results) {
			break
		}
		if result.Results[i].String() != expected {
			t.Errorf("Expected result[%d] = %s, got %s", i, expected, result.Results[i].String())
		}
	}
}

// TestFieldExtractionWildcardMultipleFields tests wildcard with multiple matching elements
func TestFieldExtractionWildcardMultipleFields(t *testing.T) {
	xml := `<root>
		<section id="1">
			<product><price>10</price></product>
			<product><price>20</price></product>
		</section>
		<section id="2">
			<product><price>30</price></product>
		</section>
	</root>`

	result := Get(xml, "root.*.product.#.price")

	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}

	// Should extract all prices: 10, 20, 30
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}

	expectedPrices := []string{"10", "20", "30"}
	for i, expected := range expectedPrices {
		if i >= len(result.Results) {
			break
		}
		if result.Results[i].String() != expected {
			t.Errorf("Expected result[%d] = %s, got %s", i, expected, result.Results[i].String())
		}
	}
}

// TestFieldExtractionWildcardNoMatches tests wildcard + field extraction with no results
func TestFieldExtractionWildcardNoMatches(t *testing.T) {
	xml := `<root>
		<category>
			<item><value>1</value></item>
		</category>
		<category>
			<item><value>2</value></item>
		</category>
	</root>`

	// Try to extract non-existent field
	result := Get(xml, "root.*.item.#.name")

	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}

	// Should return empty array since no <name> elements exist
	if len(result.Results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(result.Results))
	}
}

// TestFieldExtractionWildcardAttribute tests wildcard + attribute field extraction
func TestFieldExtractionWildcardAttribute(t *testing.T) {
	xml := `<root>
		<section>
			<item id="1" name="First"/>
			<item id="2" name="Second"/>
		</section>
		<section>
			<item id="3" name="Third"/>
		</section>
	</root>`

	result := Get(xml, "root.*.item.#.@id")

	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}

	// Should extract all id attributes: 1, 2, 3
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}

	expectedIDs := []string{"1", "2", "3"}
	for i, expected := range expectedIDs {
		if i >= len(result.Results) {
			break
		}
		if result.Results[i].String() != expected {
			t.Errorf("Expected result[%d] = %s, got %s", i, expected, result.Results[i].String())
		}
	}
}

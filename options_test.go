// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"testing"
)

// ============================================================================
// Basic Options Tests
// ============================================================================

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if !opts.CaseSensitive {
		t.Error("Expected CaseSensitive to be true")
	}
	if opts.Indent != "" {
		t.Error("Expected Indent to be empty string")
	}
	if opts.PreserveWhitespace {
		t.Error("Expected PreserveWhitespace to be false")
	}
	if opts.Namespaces != nil {
		t.Error("Expected Namespaces to be nil")
	}
}

func TestOptionsStructInitialization(t *testing.T) {
	opts := &Options{
		CaseSensitive:      false,
		Indent:             "  ",
		PreserveWhitespace: true,
		Namespaces:         map[string]string{"ns": "http://example.com"},
	}
	if opts.CaseSensitive {
		t.Error("Expected CaseSensitive to be false")
	}
	if opts.Indent != "  " {
		t.Errorf("Expected Indent to be '  ', got %q", opts.Indent)
	}
	if !opts.PreserveWhitespace {
		t.Error("Expected PreserveWhitespace to be true")
	}
	if opts.Namespaces == nil || opts.Namespaces["ns"] != "http://example.com" {
		t.Error("Expected Namespaces to contain mapping")
	}
}

func TestOptionsZeroValue(t *testing.T) {
	var opts Options
	// Zero value for bool fields is false
	if opts.CaseSensitive {
		t.Error("Expected zero-value CaseSensitive to be false")
	}
	if opts.Indent != "" {
		t.Error("Expected zero-value Indent to be empty")
	}
	if opts.PreserveWhitespace {
		t.Error("Expected zero-value PreserveWhitespace to be false")
	}
	if opts.Namespaces != nil {
		t.Error("Expected zero-value Namespaces to be nil")
	}
}

func TestOptionsAllFieldsSet(t *testing.T) {
	opts := &Options{
		CaseSensitive:      false,
		Indent:             "\t",
		PreserveWhitespace: true,
		Namespaces:         map[string]string{"x": "http://x.com", "y": "http://y.com"},
	}
	if opts.CaseSensitive {
		t.Error("Expected CaseSensitive to be false")
	}
	if opts.Indent != "\t" {
		t.Errorf("Expected Indent to be tab, got %q", opts.Indent)
	}
	if !opts.PreserveWhitespace {
		t.Error("Expected PreserveWhitespace to be true")
	}
	if len(opts.Namespaces) != 2 {
		t.Errorf("Expected Namespaces to have 2 entries, got %d", len(opts.Namespaces))
	}
}

func TestIsDefaultOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     *Options
		expected bool
	}{
		{
			name:     "default options",
			opts:     DefaultOptions(),
			expected: true,
		},
		{
			name:     "nil options",
			opts:     nil,
			expected: true,
		},
		{
			name:     "zero value (not default because CaseSensitive defaults to false)",
			opts:     &Options{},
			expected: false,
		},
		{
			name:     "case insensitive",
			opts:     &Options{CaseSensitive: false},
			expected: false,
		},
		{
			name:     "with indent",
			opts:     &Options{CaseSensitive: true, Indent: "  "},
			expected: false,
		},
		{
			name:     "with whitespace",
			opts:     &Options{CaseSensitive: true, PreserveWhitespace: true},
			expected: false,
		},
		{
			name:     "with namespaces",
			opts:     &Options{CaseSensitive: true, Namespaces: map[string]string{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDefaultOptions(tt.opts)
			if result != tt.expected {
				t.Errorf("isDefaultOptions() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// GetWithOptions Tests
// ============================================================================

func TestGetWithOptionsCaseSensitive(t *testing.T) {
	xml := `<ROOT><CHILD>value</CHILD></ROOT>`

	// Case-sensitive (default) - should not match
	opts := &Options{CaseSensitive: true}
	result := GetWithOptions(xml, "root.child", opts)
	if result.Exists() {
		t.Error("Expected no match with case-sensitive search for lowercase path")
	}

	// Case-insensitive - should match
	opts.CaseSensitive = false
	result = GetWithOptions(xml, "root.child", opts)
	if !result.Exists() {
		t.Error("Expected match with case-insensitive search")
	}
	if result.String() != "value" {
		t.Errorf("Expected 'value', got %q", result.String())
	}
}

func TestGetWithOptionsDefaultOptionsFastPath(t *testing.T) {
	xml := `<root><child>value</child></root>`

	// Using default options should use fast path
	opts := DefaultOptions()
	result := GetWithOptions(xml, "root.child", opts)
	if !result.Exists() {
		t.Error("Expected match with default options")
	}
	if result.String() != "value" {
		t.Errorf("Expected 'value', got %q", result.String())
	}
}

func TestGetWithOptionsCaseInsensitiveNested(t *testing.T) {
	xml := `<ROOT><User><Name>John</Name><Age>30</Age></User></ROOT>`

	opts := &Options{CaseSensitive: false}
	result := GetWithOptions(xml, "root.user.name", opts)
	if !result.Exists() {
		t.Error("Expected match with case-insensitive nested path")
	}
	if result.String() != "John" {
		t.Errorf("Expected 'John', got %q", result.String())
	}

	result = GetWithOptions(xml, "ROOT.USER.AGE", opts)
	if !result.Exists() {
		t.Error("Expected match with uppercase path")
	}
	if result.String() != "30" {
		t.Errorf("Expected '30', got %q", result.String())
	}
}

func TestGetWithOptionsCaseInsensitiveAttribute(t *testing.T) {
	xml := `<root><user ID="123" Name="John"/></root>`

	opts := &Options{CaseSensitive: false}
	result := GetWithOptions(xml, "root.user.@id", opts)
	if !result.Exists() {
		t.Error("Expected match with case-insensitive attribute")
	}
	if result.String() != "123" {
		t.Errorf("Expected '123', got %q", result.String())
	}

	result = GetWithOptions(xml, "root.user.@NAME", opts)
	if !result.Exists() {
		t.Error("Expected match with uppercase attribute name")
	}
	if result.String() != "John" {
		t.Errorf("Expected 'John', got %q", result.String())
	}
}

func TestGetWithOptionsCaseInsensitiveArrays(t *testing.T) {
	xml := `<ROOT><USERS><User>Alice</User><User>Bob</User><User>Charlie</User></USERS></ROOT>`

	opts := &Options{CaseSensitive: false}
	result := GetWithOptions(xml, "root.users.user.1", opts)
	if !result.Exists() {
		t.Error("Expected match with case-insensitive array index")
	}
	if result.String() != "Bob" {
		t.Errorf("Expected 'Bob', got %q", result.String())
	}

	result = GetWithOptions(xml, "ROOT.USERS.USER.#", opts)
	if !result.Exists() {
		t.Error("Expected match with case-insensitive array count")
	}
	if result.Int() != 3 {
		t.Errorf("Expected 3, got %d", result.Int())
	}
}

func TestGetWithOptionsCaseInsensitiveWildcards(t *testing.T) {
	xml := `<ROOT><User><Name>Alice</Name></User><User><Name>Bob</Name></User></ROOT>`

	opts := &Options{CaseSensitive: false}
	result := GetWithOptions(xml, "root.*.name", opts)
	if !result.Exists() {
		t.Error("Expected match with case-insensitive wildcard")
	}
	if result.Type != Array {
		t.Errorf("Expected Array type, got %v", result.Type)
	}
	if len(result.Array()) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Array()))
	}
}

func TestGetWithOptionsCaseInsensitiveRecursiveWildcard(t *testing.T) {
	xml := `<ROOT><Level1><Level2><NAME>Alice</NAME></Level2></Level1></ROOT>`

	opts := &Options{CaseSensitive: false}
	result := GetWithOptions(xml, "root.**.name", opts)
	if !result.Exists() {
		t.Error("Expected match with case-insensitive recursive wildcard")
	}
	if result.String() != "Alice" {
		t.Errorf("Expected 'Alice', got %q", result.String())
	}
}

func TestGetBytesWithOptions(t *testing.T) {
	xml := []byte(`<ROOT><CHILD>value</CHILD></ROOT>`)

	opts := &Options{CaseSensitive: false}
	result := GetBytesWithOptions(xml, "root.child", opts)
	if !result.Exists() {
		t.Error("Expected match with GetBytesWithOptions")
	}
	if result.String() != "value" {
		t.Errorf("Expected 'value', got %q", result.String())
	}
}

// ============================================================================
// SetWithOptions Tests
// ============================================================================

func TestSetWithOptionsDefaultOptionsFastPath(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	opts := DefaultOptions()
	result, err := SetWithOptions(xml, "root.user.age", 30, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(result, "<age>30</age>") {
		t.Errorf("Expected result to contain '<age>30</age>', got %q", result)
	}
}

func TestSetWithOptionsIndent(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	opts := &Options{CaseSensitive: true, Indent: "  "}
	result, err := SetWithOptions(xml, "root.user.age", 30, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that the result contains indentation (newlines)
	if !strings.Contains(result, "\n") {
		t.Errorf("Expected indented output to contain newlines, got %q", result)
	}
	if !strings.Contains(result, "<age>30</age>") {
		t.Errorf("Expected result to contain '<age>30</age>', got %q", result)
	}
}

func TestSetWithOptionsCaseInsensitive(t *testing.T) {
	xml := `<ROOT><USER><NAME>John</NAME></USER></ROOT>`

	opts := &Options{CaseSensitive: false}
	result, err := SetWithOptions(xml, "root.user.age", 30, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the new element was added
	if !strings.Contains(result, "<age>30</age>") {
		t.Errorf("Expected result to contain '<age>30</age>', got %q", result)
	}

	// Verify we can read it back case-insensitively
	getName := GetWithOptions(result, "root.user.name", opts)
	if getName.String() != "John" {
		t.Errorf("Expected 'John', got %q", getName.String())
	}
}

func TestSetWithOptionsCreateNew(t *testing.T) {
	xml := `<root></root>`

	opts := &Options{CaseSensitive: false, Indent: "  "}
	result, err := SetWithOptions(xml, "root.user.name", "Alice", opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that elements were created
	if !strings.Contains(result, "<user>") {
		t.Errorf("Expected result to contain '<user>', got %q", result)
	}
	if !strings.Contains(result, "<name>Alice</name>") {
		t.Errorf("Expected result to contain '<name>Alice</name>', got %q", result)
	}
}

func TestSetWithOptionsAttributes(t *testing.T) {
	xml := `<root><user id="123"/></root>`

	opts := &Options{CaseSensitive: false}
	result, err := SetWithOptions(xml, "root.user.@name", "Alice", opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that attribute was added
	if !strings.Contains(result, `name="Alice"`) {
		t.Errorf("Expected result to contain 'name=\"Alice\"', got %q", result)
	}

	// Verify we can read it back
	getName := GetWithOptions(result, "root.user.@NAME", opts)
	if getName.String() != "Alice" {
		t.Errorf("Expected 'Alice', got %q", getName.String())
	}
}

func TestSetWithOptionsPreserveStructure(t *testing.T) {
	xml := `<root><user><name>John</name><age>30</age></user></root>`

	opts := &Options{CaseSensitive: true, Indent: ""}
	result, err := SetWithOptions(xml, "root.user.age", 31, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify name is preserved
	getName := Get(result, "root.user.name")
	if getName.String() != "John" {
		t.Errorf("Expected 'John', got %q", getName.String())
	}

	// Verify age was updated
	getAge := Get(result, "root.user.age")
	if getAge.Int() != 31 {
		t.Errorf("Expected 31, got %d", getAge.Int())
	}
}

func TestSetBytesWithOptions(t *testing.T) {
	xml := []byte(`<ROOT><USER><NAME>John</NAME></USER></ROOT>`)

	opts := &Options{CaseSensitive: false}
	result, err := SetBytesWithOptions(xml, "root.user.age", 30, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(string(result), "<age>30</age>") {
		t.Errorf("Expected result to contain '<age>30</age>', got %q", string(result))
	}
}

func TestSetWithOptionsIndentDeepNesting(t *testing.T) {
	xml := `<root></root>`

	opts := &Options{CaseSensitive: true, Indent: "  "}
	result, err := SetWithOptions(xml, "root.level1.level2.level3.level4.value", "deep", opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Count indentation levels
	lines := strings.Split(result, "\n")
	deepestIndent := 0
	for _, line := range lines {
		if strings.Contains(line, "<value>") {
			// Count leading spaces
			indent := 0
			for _, ch := range line {
				if ch == ' ' {
					indent++
				} else {
					break
				}
			}
			deepestIndent = indent / 2 // 2 spaces per indent level
			break
		}
	}

	// We expect at least 4 levels of indentation (level1, level2, level3, level4)
	if deepestIndent < 4 {
		t.Errorf("Expected at least 4 levels of indentation, got %d", deepestIndent)
	}
}

// ============================================================================
// Performance & Edge Cases
// ============================================================================

func BenchmarkGetVsGetWithOptionsDefault(b *testing.B) {
	xml := `<root><user><name>John</name><age>30</age></user></root>`
	opts := DefaultOptions()

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Get(xml, "root.user.name")
		}
	})

	b.Run("GetWithOptions-Default", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetWithOptions(xml, "root.user.name", opts)
		}
	})
}

func BenchmarkSetVsSetWithOptionsDefault(b *testing.B) {
	xml := `<root><user><name>John</name></user></root>`
	opts := DefaultOptions()

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Set(xml, "root.user.age", 30)
		}
	})

	b.Run("SetWithOptions-Default", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = SetWithOptions(xml, "root.user.age", 30, opts)
		}
	})
}

func TestOptionsEmptyIndent(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	opts := &Options{CaseSensitive: true, Indent: ""}
	result, err := SetWithOptions(xml, "root.user.age", 30, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Empty indent should not add newlines
	if strings.Count(result, "\n") > strings.Count(xml, "\n") {
		t.Errorf("Expected no additional newlines with empty indent")
	}
}

func TestOptionsCustomIndentString(t *testing.T) {
	xml := `<root></root>`

	// Test with 4 spaces
	opts := &Options{CaseSensitive: true, Indent: "    "}
	result, err := SetWithOptions(xml, "root.child.value", "test", opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(result, "    ") {
		t.Errorf("Expected result to contain 4-space indentation")
	}

	// Test with tabs
	opts.Indent = "\t"
	result, err = SetWithOptions(xml, "root.child.value", "test", opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(result, "\t") {
		t.Errorf("Expected result to contain tab indentation")
	}
}

func TestOptionsCombinedCaseInsensitiveAndIndent(t *testing.T) {
	xml := `<ROOT><USER><NAME>John</NAME></USER></ROOT>`

	opts := &Options{CaseSensitive: false, Indent: "  "}
	result, err := SetWithOptions(xml, "root.user.age", 30, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify both case-insensitive matching worked and indentation was applied
	if !strings.Contains(result, "<age>30</age>") {
		t.Errorf("Expected result to contain '<age>30</age>', got %q", result)
	}

	// Verify we can read back case-insensitively
	getAge := GetWithOptions(result, "ROOT.USER.AGE", opts)
	if getAge.Int() != 30 {
		t.Errorf("Expected 30, got %d", getAge.Int())
	}
}

// Benchmarks for Options

func BenchmarkOptionsPointerVsValue(b *testing.B) {
	xml := `<root><items><item id="1">First</item><item id="2">Second</item></items></root>`
	path := "root.items.item"
	opts := &Options{CaseSensitive: false}

	b.Run("WithPointer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GetWithOptions(xml, path, opts)
		}
	})
}

func BenchmarkGetWithOptionsDefaultVsGet(b *testing.B) {
	xml := `<root><items><item id="1">First</item><item id="2">Second</item></items></root>`
	path := "root.items.item"

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Get(xml, path)
		}
	})

	b.Run("GetWithOptionsNil", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GetWithOptions(xml, path, nil)
		}
	})

	b.Run("GetWithOptionsDefault", func(b *testing.B) {
		opts := DefaultOptions()
		for i := 0; i < b.N; i++ {
			GetWithOptions(xml, path, opts)
		}
	})

	b.Run("GetWithOptionsCaseInsensitive", func(b *testing.B) {
		opts := &Options{CaseSensitive: false}
		for i := 0; i < b.N; i++ {
			GetWithOptions(xml, path, opts)
		}
	})
}

// Example functions for godoc

// ExampleGetWithOptions_caseInsensitive demonstrates case-insensitive queries
func ExampleGetWithOptions_caseInsensitive() {
	xml := `<Document>
		<Person>
			<FirstName>John</FirstName>
			<LastName>Doe</LastName>
		</Person>
	</Document>`

	// Case-insensitive query (matches despite different casing)
	opts := Options{CaseSensitive: false}
	firstName := GetWithOptions(xml, "document.person.firstname", &opts)
	lastName := GetWithOptions(xml, "document.person.lastname", &opts)

	fmt.Printf("%s %s\n", firstName.String(), lastName.String())
	// Output: John Doe
}

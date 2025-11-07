// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"strings"
	"testing"
)

// TestSet_AppendOperation tests the -1 index append functionality
func TestSet_AppendOperation(t *testing.T) {
	tests := []struct {
		name   string
		xml    string
		path   string
		value  interface{}
		verify func(t *testing.T, result string)
	}{
		{
			name:  "append to empty parent (creates first element)",
			xml:   `<root></root>`,
			path:  "root.item.-1",
			value: "first",
			verify: func(t *testing.T, result string) {
				first := Get(result, "root.item.0")
				if first.String() != "first" {
					t.Errorf("item.0 = %v, want first", first.String())
				}
				count := Get(result, "root.item.#")
				if count.Int() != 1 {
					t.Errorf("item.# = %v, want 1", count.Int())
				}
			},
		},
		{
			name:  "append to single element (treat as 1-element array)",
			xml:   `<root><item>A</item></root>`,
			path:  "root.item.-1",
			value: "B",
			verify: func(t *testing.T, result string) {
				first := Get(result, "root.item.0")
				if first.String() != "A" {
					t.Errorf("item.0 = %v, want A", first.String())
				}
				second := Get(result, "root.item.1")
				if second.String() != "B" {
					t.Errorf("item.1 = %v, want B", second.String())
				}
				count := Get(result, "root.item.#")
				if count.Int() != 2 {
					t.Errorf("item.# = %v, want 2", count.Int())
				}
			},
		},
		{
			name:  "append to multi-element array",
			xml:   `<root><item>A</item><item>B</item><item>C</item></root>`,
			path:  "root.item.-1",
			value: "D",
			verify: func(t *testing.T, result string) {
				fourth := Get(result, "root.item.3")
				if fourth.String() != "D" {
					t.Errorf("item.3 = %v, want D", fourth.String())
				}
				count := Get(result, "root.item.#")
				if count.Int() != 4 {
					t.Errorf("item.# = %v, want 4", count.Int())
				}
			},
		},
		{
			name:  "append with nested parent path",
			xml:   `<root><items></items></root>`,
			path:  "root.items.item.-1",
			value: "first",
			verify: func(t *testing.T, result string) {
				first := Get(result, "root.items.item.0")
				if first.String() != "first" {
					t.Errorf("item.0 = %v, want first", first.String())
				}
			},
		},
		{
			name:  "append with missing parent (auto-create)",
			xml:   `<root></root>`,
			path:  "root.items.item.-1",
			value: "first",
			verify: func(t *testing.T, result string) {
				first := Get(result, "root.items.item.0")
				if first.String() != "first" {
					t.Errorf("item.0 = %v, want first", first.String())
				}
				// Verify parent was created
				items := Get(result, "root.items")
				if !items.Exists() {
					t.Error("root.items should exist")
				}
			},
		},
		{
			name:  "append with int value",
			xml:   `<root><count>10</count></root>`,
			path:  "root.count.-1",
			value: 20,
			verify: func(t *testing.T, result string) {
				first := Get(result, "root.count.0")
				if first.String() != "10" {
					t.Errorf("count.0 = %v, want 10", first.String())
				}
				second := Get(result, "root.count.1")
				if second.String() != "20" {
					t.Errorf("count.1 = %v, want 20", second.String())
				}
			},
		},
		{
			name:  "append with float value",
			xml:   `<root><price>9.99</price></root>`,
			path:  "root.price.-1",
			value: 19.99,
			verify: func(t *testing.T, result string) {
				second := Get(result, "root.price.1")
				if second.String() != "19.99" {
					t.Errorf("price.1 = %v, want 19.99", second.String())
				}
			},
		},
		{
			name:  "append with bool value",
			xml:   `<root><flag>true</flag></root>`,
			path:  "root.flag.-1",
			value: false,
			verify: func(t *testing.T, result string) {
				second := Get(result, "root.flag.1")
				if second.String() != "false" {
					t.Errorf("flag.1 = %v, want false", second.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			tt.verify(t, result)
		})
	}
}

// TestSetRaw_AppendOperation tests SetRaw with -1 index
func TestSetRaw_AppendOperation(t *testing.T) {
	tests := []struct {
		name   string
		xml    string
		path   string
		rawXML string
		verify func(t *testing.T, result string)
	}{
		{
			name:   "append raw XML element",
			xml:    `<root><item>A</item></root>`,
			path:   "root.item.-1",
			rawXML: "<nested>B</nested>",
			verify: func(t *testing.T, result string) {
				second := Get(result, "root.item.1.nested")
				if second.String() != "B" {
					t.Errorf("item.1.nested = %v, want B", second.String())
				}
			},
		},
		{
			name:   "append complex raw XML",
			xml:    `<root></root>`,
			path:   "root.product.-1",
			rawXML: "<id>123</id><name>Widget</name><price>9.99</price>",
			verify: func(t *testing.T, result string) {
				id := Get(result, "root.product.0.id")
				if id.String() != "123" {
					t.Errorf("product.0.id = %v, want 123", id.String())
				}
				name := Get(result, "root.product.0.name")
				if name.String() != "Widget" {
					t.Errorf("product.0.name = %v, want Widget", name.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetRaw(tt.xml, tt.path, tt.rawXML)
			if err != nil {
				t.Fatalf("SetRaw() error = %v", err)
			}
			tt.verify(t, result)
		})
	}
}

// TestSet_AppendOperationErrors tests error cases for append
func TestSet_AppendOperationErrors(t *testing.T) {
	tests := []struct {
		name      string
		xml       string
		path      string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:      "reject -2 index",
			xml:       `<root><item>A</item></root>`,
			path:      "root.item.-2",
			value:     "B",
			wantErr:   true,
			errString: "reserved for future use",
		},
		{
			name:      "reject -3 index",
			xml:       `<root><item>A</item></root>`,
			path:      "root.item.-3",
			value:     "B",
			wantErr:   true,
			errString: "reserved for future use",
		},
		{
			name:      "reject nested path after -1",
			xml:       `<root><item>A</item></root>`,
			path:      "root.item.-1.child",
			value:     "B",
			wantErr:   true,
			errString: "nested paths",
		},
		{
			name:      "reject nested attribute after -1",
			xml:       `<root><item>A</item></root>`,
			path:      "root.item.-1.@attr",
			value:     "B",
			wantErr:   true,
			errString: "nested paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Set(tt.xml, tt.path, tt.value)
			if tt.wantErr && err == nil {
				t.Errorf("Set() should return error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Set() unexpected error = %v", err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("Set() error = %v, want error containing %q", err, tt.errString)
			}
		})
	}
}

// TestSet_AppendMultiple tests multiple sequential appends
func TestSet_AppendMultiple(t *testing.T) {
	xml := `<root></root>`

	// Append first element
	var err error
	xml, err = Set(xml, "root.item.-1", "A")
	if err != nil {
		t.Fatalf("First append error: %v", err)
	}

	// Append second element
	xml, err = Set(xml, "root.item.-1", "B")
	if err != nil {
		t.Fatalf("Second append error: %v", err)
	}

	// Append third element
	xml, err = Set(xml, "root.item.-1", "C")
	if err != nil {
		t.Fatalf("Third append error: %v", err)
	}

	// Verify all elements
	count := Get(xml, "root.item.#")
	if count.Int() != 3 {
		t.Errorf("item.# = %d, want 3", count.Int())
	}

	first := Get(xml, "root.item.0")
	if first.String() != "A" {
		t.Errorf("item.0 = %v, want A", first.String())
	}

	second := Get(xml, "root.item.1")
	if second.String() != "B" {
		t.Errorf("item.1 = %v, want B", second.String())
	}

	third := Get(xml, "root.item.2")
	if third.String() != "C" {
		t.Errorf("item.2 = %v, want C", third.String())
	}
}

// TestSet_AppendSelfClosingParent tests append to self-closing parent elements
func TestSet_AppendSelfClosingParent(t *testing.T) {
	tests := []struct {
		name   string
		xml    string
		path   string
		value  string
		verify func(t *testing.T, result string)
	}{
		{
			name:  "append to self-closing parent (no attributes)",
			xml:   `<root><items/></root>`,
			path:  "root.items.item.-1",
			value: "first",
			verify: func(t *testing.T, result string) {
				first := Get(result, "root.items.item.0")
				if first.String() != "first" {
					t.Errorf("item.0 = %v, want first", first.String())
				}
				// Verify self-closing tag was converted
				if strings.Contains(result, "<items/>") {
					t.Error("Self-closing tag should be converted to full element")
				}
			},
		},
		{
			name:  "append to self-closing parent (with attributes)",
			xml:   `<root><items id="list1" class="active"/></root>`,
			path:  "root.items.item.-1",
			value: "first",
			verify: func(t *testing.T, result string) {
				// Verify element created
				first := Get(result, "root.items.item.0")
				if first.String() != "first" {
					t.Errorf("item.0 = %v, want first", first.String())
				}
				// Verify attributes preserved
				id := Get(result, "root.items.@id")
				if id.String() != "list1" {
					t.Errorf("items.@id = %v, want list1 (attribute should be preserved)", id.String())
				}
				class := Get(result, "root.items.@class")
				if class.String() != "active" {
					t.Errorf("items.@class = %v, want active", class.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			tt.verify(t, result)
		})
	}
}

// TestSet_AppendWithOptions tests append with different options
func TestSet_AppendWithOptions(t *testing.T) {
	t.Run("append with indentation", func(t *testing.T) {
		xml := `<root>
  <item>first</item>
</root>`
		opts := &Options{Indent: "  "}
		result, err := SetWithOptions(xml, "root.item.-1", "second", opts)
		if err != nil {
			t.Fatalf("SetWithOptions() error = %v", err)
		}

		// Verify second element was appended
		count := Get(result, "root.item.#")
		if count.Int() != 2 {
			t.Errorf("item.# = %d, want 2", count.Int())
		}

		// Result should contain indented XML
		if !strings.Contains(result, "\n") {
			t.Error("Expected indented XML but got inline")
		}
	})
}

// TestSet_AppendMixedSiblings tests append with mixed element types in parent
func TestSet_AppendMixedSiblings(t *testing.T) {
	xml := `<root><item>A</item><other>X</other><item>B</item><foo>Y</foo></root>`
	result, err := Set(xml, "root.item.-1", "C")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify count of item elements
	count := Get(result, "root.item.#")
	if count.Int() != 3 {
		t.Errorf("item.# = %d, want 3", count.Int())
	}

	// Verify new item is appended after last item (after B)
	third := Get(result, "root.item.2")
	if third.String() != "C" {
		t.Errorf("item.2 = %v, want C", third.String())
	}

	// Verify other elements unchanged
	other := Get(result, "root.other")
	if other.String() != "X" {
		t.Errorf("other = %v, want X (should be unchanged)", other.String())
	}
}

// TestSet_AppendPreservesAttributes tests that parent attributes are preserved
func TestSet_AppendPreservesAttributes(t *testing.T) {
	xml := `<root><items id="list1" class="active"><item>A</item></items></root>`
	result, err := Set(xml, "root.items.item.-1", "B")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify element appended
	second := Get(result, "root.items.item.1")
	if second.String() != "B" {
		t.Errorf("item.1 = %v, want B", second.String())
	}

	// Verify parent attributes preserved
	id := Get(result, "root.items.@id")
	if id.String() != "list1" {
		t.Errorf("items.@id = %v, want list1", id.String())
	}

	class := Get(result, "root.items.@class")
	if class.String() != "active" {
		t.Errorf("items.@class = %v, want active", class.String())
	}
}

// TestSet_AppendDocumentSizeLimit tests that append respects MaxDocumentSize
func TestSet_AppendDocumentSizeLimit(t *testing.T) {
	// Create a document near the size limit
	largeValue := strings.Repeat("x", MaxDocumentSize-100)
	xml := `<root></root>`

	// First append should succeed (document still under limit)
	result, err := Set(xml, "root.item.-1", "small")
	if err != nil {
		t.Fatalf("First append should succeed: %v", err)
	}

	// Try to append a value that would exceed MaxDocumentSize
	_, err = Set(result, "root.item.-1", largeValue)
	if err == nil {
		t.Error("Append should fail when resulting document exceeds MaxDocumentSize")
	}
	if err != nil && !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Errorf("Expected size limit error, got: %v", err)
	}
}

// TestSet_AppendRootLevel tests the fix for root-level append creating siblings instead of nesting
func TestSet_AppendRootLevel(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		value    string
		expected string
	}{
		{
			name:     "append different root name creates sibling",
			xml:      `<user>Alice</user>`,
			path:     "item.-1",
			value:    "first",
			expected: `<user>Alice</user><item>first</item>`,
		},
		{
			name:     "append matching root name creates sibling array",
			xml:      `<item>A</item>`,
			path:     "item.-1",
			value:    "B",
			expected: `<item>A</item><item>B</item>`,
		},
		{
			name:     "append to empty XML creates first root",
			xml:      ``,
			path:     "item.-1",
			value:    "first",
			expected: `<item>first</item>`,
		},
		{
			name:     "append to multi-root different name",
			xml:      `<user>Alice</user><name>test</name>`,
			path:     "item.-1",
			value:    "first",
			expected: `<user>Alice</user><name>test</name><item>first</item>`,
		},
		{
			name:     "append to multi-root matching name",
			xml:      `<item>A</item><item>B</item>`,
			path:     "item.-1",
			value:    "C",
			expected: `<item>A</item><item>B</item><item>C</item>`,
		},
		{
			name:     "multiple sequential appends to root level",
			xml:      `<user>Alice</user>`,
			path:     "item.-1",
			value:    "first",
			expected: `<user>Alice</user><item>first</item>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Set() = %q, want %q", result, tt.expected)
			}

			// Verify we can query the new element
			elemName := tt.path[:len(tt.path)-3] // Remove ".-1"
			query := Get(result, elemName)
			if !query.Exists() {
				t.Errorf("Element %q should exist after append", elemName)
			}
		})
	}

	// Test multiple sequential appends
	t.Run("multiple sequential root appends", func(t *testing.T) {
		xml := `<user>Alice</user>`
		xml, _ = Set(xml, "item.-1", "first")
		xml, _ = Set(xml, "item.-1", "second")

		expected := `<user>Alice</user><item>first</item><item>second</item>`
		if xml != expected {
			t.Errorf("Multiple appends = %q, want %q", xml, expected)
		}

		// Verify count
		count := Get(xml, "item.#")
		if count.Int() != 2 {
			t.Errorf("item.# = %d, want 2", count.Int())
		}
	})
}

// TestSet_AppendToEmptyArrayEdgeCases tests the fix for appending to empty arrays.
// Prior to v0.4.3, appending to an empty array inserted at the beginning of the parent element.
// After v0.4.3, it correctly inserts at the end of the parent's content.
func TestSet_AppendToEmptyArrayEdgeCases(t *testing.T) {
	t.Run("append to empty parent with other children", func(t *testing.T) {
		xml := `<root><other>existing</other></root>`
		result, err := Set(xml, "root.item.-1", "new")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// Verify new item exists
		item := Get(result, "root.item.0")
		if item.String() != "new" {
			t.Errorf("item.0 = %v, want 'new'", item.String())
		}

		// Verify existing content preserved
		other := Get(result, "root.other")
		if other.String() != "existing" {
			t.Errorf("other = %v, want 'existing'", other.String())
		}

		// Verify order: new item should come AFTER existing content
		if !strings.Contains(result, "<other>existing</other><item>new</item>") {
			t.Errorf("Expected item after other, got: %s", result)
		}
	})

	t.Run("append to completely empty parent", func(t *testing.T) {
		xml := `<root></root>`
		result, err := Set(xml, "root.item.-1", "first")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		expected := `<root><item>first</item></root>`
		if result != expected {
			t.Errorf("got %s, want %s", result, expected)
		}
	})

	t.Run("append to empty parent with whitespace", func(t *testing.T) {
		xml := `<root>
  </root>`
		result, err := Set(xml, "root.item.-1", "value")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// Verify item was created
		item := Get(result, "root.item.0")
		if item.String() != "value" {
			t.Errorf("item.0 = %v, want 'value'", item.String())
		}

		// Should append at end (before closing tag)
		if !strings.HasSuffix(strings.TrimSpace(result), "<item>value</item></root>") {
			t.Errorf("Expected item at end of parent, got: %s", result)
		}
	})

	t.Run("multiple appends to initially empty parent", func(t *testing.T) {
		xml := `<root></root>`

		result, err := Set(xml, "root.item.-1", "first")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		result, err = Set(result, "root.item.-1", "second")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		result, err = Set(result, "root.item.-1", "third")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// Verify count
		count := Get(result, "root.item.#")
		if count.Int() != 3 {
			t.Errorf("item.# = %d, want 3", count.Int())
		}

		// Verify order
		first := Get(result, "root.item.0")
		second := Get(result, "root.item.1")
		third := Get(result, "root.item.2")

		if first.String() != "first" {
			t.Errorf("item.0 = %v, want 'first'", first.String())
		}
		if second.String() != "second" {
			t.Errorf("item.1 = %v, want 'second'", second.String())
		}
		if third.String() != "third" {
			t.Errorf("item.2 = %v, want 'third'", third.String())
		}
	})

	t.Run("append to empty parent with mixed siblings", func(t *testing.T) {
		xml := `<root><a>1</a><b>2</b></root>`
		result, err := Set(xml, "root.item.-1", "new")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// Verify all existing elements preserved
		a := Get(result, "root.a")
		if a.String() != "1" {
			t.Errorf("a = %v, want '1'", a.String())
		}

		b := Get(result, "root.b")
		if b.String() != "2" {
			t.Errorf("b = %v, want '2'", b.String())
		}

		// Verify new item exists
		item := Get(result, "root.item.0")
		if item.String() != "new" {
			t.Errorf("item.0 = %v, want 'new'", item.String())
		}

		// Verify order: item should come AFTER existing siblings
		if !strings.Contains(result, "<a>1</a><b>2</b><item>new</item>") {
			t.Errorf("Expected item after siblings, got: %s", result)
		}
	})

	t.Run("regression - append to non-empty array still works", func(t *testing.T) {
		xml := `<root><item>A</item><item>B</item></root>`
		result, err := Set(xml, "root.item.-1", "C")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// Should append after B (existing behavior preserved)
		third := Get(result, "root.item.2")
		if third.String() != "C" {
			t.Errorf("item.2 = %v, want 'C'", third.String())
		}

		// Verify A and B unchanged
		first := Get(result, "root.item.0")
		second := Get(result, "root.item.1")
		if first.String() != "A" || second.String() != "B" {
			t.Error("Existing items corrupted")
		}
	})

	t.Run("SetRaw append to empty parent", func(t *testing.T) {
		xml := `<root></root>`
		result, err := SetRaw(xml, "root.item.-1", "<name>Test</name>")
		if err != nil {
			t.Fatalf("SetRaw() error = %v", err)
		}

		name := Get(result, "root.item.0.name")
		if name.String() != "Test" {
			t.Errorf("item.0.name = %v, want 'Test'", name.String())
		}
	})

	t.Run("deeply nested empty parent", func(t *testing.T) {
		xml := `<root><level1><level2></level2></level1></root>`
		result, err := Set(xml, "root.level1.level2.item.-1", "deep")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		item := Get(result, "root.level1.level2.item.0")
		if item.String() != "deep" {
			t.Errorf("deep item = %v, want 'deep'", item.String())
		}
	})
}

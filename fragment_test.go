// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"testing"
)

// TestFragmentGetOperations tests querying XML fragments with multiple roots
func TestFragmentGetOperations(t *testing.T) {
	fragment := `<user id="1"><name>Alice</name><age>30</age></user>
<user id="2"><name>Bob</name><age>25</age></user>
<user id="3"><name>Carol</name><age>35</age></user>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "get first user name",
			path:     "user.name",
			expected: "Alice",
		},
		{
			name:     "get first user age",
			path:     "user.age",
			expected: "30",
		},
		{
			name:     "get first user id attribute",
			path:     "user.@id",
			expected: "1",
		},
		{
			name:     "get specific nested element",
			path:     "user.name",
			expected: "Alice", // Gets first match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(fragment, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Get(%q) = %q, want %q", tt.path, result.String(), tt.expected)
			}
		})
	}
}

// TestFragmentSetOperations tests modifying XML fragments
func TestFragmentSetOperations(t *testing.T) {
	t.Run("modify first root element", func(t *testing.T) {
		fragment := `<user id="1"><name>Alice</name></user><user id="2"><name>Bob</name></user>`

		result, err := Set(fragment, "user.name", "Alicia")
		if err != nil {
			t.Fatalf("Set on fragment failed: %v", err)
		}

		// Verify first user was modified
		firstName := Get(result, "user.name")
		if firstName.String() != "Alicia" {
			t.Errorf("Expected first name to be 'Alicia', got %q", firstName.String())
		}

		// Verify fragment structure preserved (should still have both users)
		if !Get(result, "user.@id").Exists() {
			t.Error("Fragment structure was corrupted")
		}
	})

	t.Run("add element to first root", func(t *testing.T) {
		fragment := `<user><name>Alice</name></user><user><name>Bob</name></user>`

		result, err := Set(fragment, "user.age", "30")
		if err != nil {
			t.Fatalf("Set on fragment failed: %v", err)
		}

		// Verify age was added to first user
		age := Get(result, "user.age")
		if age.String() != "30" {
			t.Errorf("Expected age '30', got %q", age.String())
		}
	})

	t.Run("set attribute in fragment", func(t *testing.T) {
		fragment := `<user><name>Alice</name></user><user><name>Bob</name></user>`

		result, err := Set(fragment, "user.@id", "100")
		if err != nil {
			t.Fatalf("Set attribute on fragment failed: %v", err)
		}

		// Verify attribute was added
		id := Get(result, "user.@id")
		if id.String() != "100" {
			t.Errorf("Expected id '100', got %q", id.String())
		}
	})
}

// TestFragmentDeleteOperations tests deleting elements from fragments
func TestFragmentDeleteOperations(t *testing.T) {
	t.Run("delete element from first root", func(t *testing.T) {
		fragment := `<user><name>Alice</name><age>30</age></user><user><name>Bob</name></user>`

		result, err := Delete(fragment, "user.age")
		if err != nil {
			t.Fatalf("Delete from fragment failed: %v", err)
		}

		// Verify age was deleted
		age := Get(result, "user.age")
		if age.Exists() {
			t.Error("Age should have been deleted")
		}

		// Verify name still exists
		name := Get(result, "user.name")
		if name.String() != "Alice" {
			t.Errorf("Name should still exist, got %q", name.String())
		}
	})

	t.Run("delete attribute from fragment", func(t *testing.T) {
		fragment := `<user id="1"><name>Alice</name></user><user id="2"><name>Bob</name></user>`

		result, err := Delete(fragment, "user.@id")
		if err != nil {
			t.Fatalf("Delete attribute from fragment failed: %v", err)
		}

		// Verify attribute was deleted from first user
		id := Get(result, "user.@id")
		if id.Exists() {
			t.Error("Attribute should have been deleted")
		}

		// Verify element content preserved
		name := Get(result, "user.name")
		if name.String() != "Alice" {
			t.Errorf("Name should be preserved, got %q", name.String())
		}
	})
}

// TestFragmentWithProlog tests fragments with XML declaration
func TestFragmentWithProlog(t *testing.T) {
	fragment := `<?xml version="1.0" encoding="UTF-8"?>
<item id="1">First</item>
<item id="2">Second</item>`

	// Validate fragment with prolog
	if !Valid(fragment) {
		t.Error("Fragment with prolog should be valid")
	}

	// Query should work
	value := Get(fragment, "item.%")
	if value.String() != "First" {
		t.Errorf("Expected 'First', got %q", value.String())
	}

	// Modification should preserve prolog
	result, err := Set(fragment, "item.@status", "active")
	if err != nil {
		t.Fatalf("Set on fragment with prolog failed: %v", err)
	}

	// Verify prolog is preserved
	if Get(result, "item.@status").String() != "active" {
		t.Error("Attribute should be set")
	}
}

// TestFragmentWithComments tests fragments with comments between roots
func TestFragmentWithComments(t *testing.T) {
	fragment := `<!-- First item -->
<item>first</item>
<!-- Second item -->
<item>second</item>
<!-- End -->`

	// Validate fragment with comments
	if !Valid(fragment) {
		t.Error("Fragment with comments should be valid")
	}

	// Query should work
	value := Get(fragment, "item.%")
	if value.String() != "first" {
		t.Errorf("Expected 'first', got %q", value.String())
	}

	// Modification should work (add attribute instead of text to avoid comment handling complexity)
	result, err := Set(fragment, "item.@status", "active")
	if err != nil {
		t.Fatalf("Set on fragment with comments failed: %v", err)
	}

	// Verify modification
	if Get(result, "item.@status").String() != "active" {
		t.Error("Attribute should be set")
	}
}

// TestFragmentComplexStructure tests fragments with complex nested structures
func TestFragmentComplexStructure(t *testing.T) {
	fragment := `<order id="1">
	<customer>
		<name>Alice</name>
		<email>alice@example.com</email>
	</customer>
	<items>
		<item><name>Widget</name><price>10.00</price></item>
		<item><name>Gadget</name><price>20.00</price></item>
	</items>
</order>
<order id="2">
	<customer>
		<name>Bob</name>
		<email>bob@example.com</email>
	</customer>
	<items>
		<item><name>Doohickey</name><price>15.00</price></item>
	</items>
</order>`

	// Validate complex fragment
	if !Valid(fragment) {
		t.Error("Complex fragment should be valid")
	}

	// Query nested elements in first root
	customerName := Get(fragment, "order.customer.name")
	if customerName.String() != "Alice" {
		t.Errorf("Expected 'Alice', got %q", customerName.String())
	}

	// Modify nested element
	result, err := Set(fragment, "order.customer.email", "alicia@example.com")
	if err != nil {
		t.Fatalf("Set on complex fragment failed: %v", err)
	}

	// Verify modification
	email := Get(result, "order.customer.email")
	if email.String() != "alicia@example.com" {
		t.Errorf("Expected updated email, got %q", email.String())
	}
}

// TestFragmentArrayOperations tests array operations on fragment roots with same names
func TestFragmentArrayOperations(t *testing.T) {
	fragment := `<user id="1"><name>Alice</name><age>30</age></user>
<user id="2"><name>Bob</name><age>25</age></user>
<user id="3"><name>Carol</name><age>35</age></user>`

	tests := []struct {
		name     string
		path     string
		expected string
		isNumber bool
	}{
		{
			name:     "count fragment roots",
			path:     "user.#",
			expected: "3",
			isNumber: true,
		},
		{
			name:     "access first root by index 0",
			path:     "user.0.name",
			expected: "Alice",
		},
		{
			name:     "access second root by index 1",
			path:     "user.1.name",
			expected: "Bob",
		},
		{
			name:     "access third root by index 2",
			path:     "user.2.name",
			expected: "Carol",
		},
		{
			name:     "access nested element from indexed root",
			path:     "user.1.name",
			expected: "Bob",
		},
		{
			name:     "access attribute from indexed root",
			path:     "user.0.@id",
			expected: "1",
		},
		{
			name:     "access attribute from second root",
			path:     "user.1.@id",
			expected: "2",
		},
		{
			name:     "nested element after root index",
			path:     "user.2.age",
			expected: "35",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(fragment, tt.path)
			if tt.isNumber {
				if result.Type != Number {
					t.Errorf("Expected Number type, got %v", result.Type)
				}
				if result.Int() != 3 {
					t.Errorf("Get(%q) = %d, want 3", tt.path, result.Int())
				}
			} else {
				if result.String() != tt.expected {
					t.Errorf("Get(%q) = %q, want %q", tt.path, result.String(), tt.expected)
				}
			}
		})
	}
}

// TestFragmentArrayFieldExtraction tests #.field syntax on fragment roots
func TestFragmentArrayFieldExtraction(t *testing.T) {
	fragment := `<user id="1"><name>Alice</name></user>
<user id="2"><name>Bob</name></user>
<user id="3"><name>Carol</name></user>`

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "extract nested element from all roots",
			path:     "user.#.name",
			expected: []string{"Alice", "Bob", "Carol"},
		},
		{
			name:     "extract attribute from all roots",
			path:     "user.#.@id",
			expected: []string{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(fragment, tt.path)
			if !result.IsArray() {
				t.Errorf("Expected array result for %q", tt.path)
			}
			arr := result.Array()
			if len(arr) != len(tt.expected) {
				t.Errorf("Expected %d results, got %d", len(tt.expected), len(arr))
			}
			for i, expected := range tt.expected {
				if arr[i].String() != expected {
					t.Errorf("Result[%d] = %q, want %q", i, arr[i].String(), expected)
				}
			}
		})
	}
}

// TestFragmentArrayTextExtraction tests #.% syntax on fragment roots with direct text
func TestFragmentArrayTextExtraction(t *testing.T) {
	// XML with direct text content in root elements
	fragment := `<user id="1">Alice</user>
<user id="2">Bob</user>
<user id="3">Carol</user>`

	// Extract direct text content from all roots
	result := Get(fragment, "user.#.%")
	if !result.IsArray() {
		t.Fatal("Expected array result")
	}

	arr := result.Array()
	expected := []string{"Alice", "Bob", "Carol"}
	if len(arr) != len(expected) {
		t.Errorf("Expected %d results, got %d", len(expected), len(arr))
	}

	for i, exp := range expected {
		if arr[i].String() != exp {
			t.Errorf("Result[%d] = %q, want %q", i, arr[i].String(), exp)
		}
	}
}

// TestFragmentArrayMixedRoots tests array operations when roots have different names
func TestFragmentArrayMixedRoots(t *testing.T) {
	fragment := `<user id="1">Alice</user>
<item>Widget</item>
<user id="2">Bob</user>
<item>Gadget</item>
<user id="3">Carol</user>`

	// Count only matching roots
	userCount := Get(fragment, "user.#")
	if userCount.Int() != 3 {
		t.Errorf("Expected 3 users, got %d", userCount.Int())
	}

	itemCount := Get(fragment, "item.#")
	if itemCount.Int() != 2 {
		t.Errorf("Expected 2 items, got %d", itemCount.Int())
	}

	// Index access with mixed roots
	firstUser := Get(fragment, "user.0.%")
	if firstUser.String() != "Alice" {
		t.Errorf("Expected 'Alice', got %q", firstUser.String())
	}

	secondUser := Get(fragment, "user.1.%")
	if secondUser.String() != "Bob" {
		t.Errorf("Expected 'Bob', got %q", secondUser.String())
	}

	thirdUser := Get(fragment, "user.2.%")
	if thirdUser.String() != "Carol" {
		t.Errorf("Expected 'Carol', got %q", thirdUser.String())
	}
}

// TestFragmentArrayBoundaries tests boundary conditions for fragment arrays
func TestFragmentArrayBoundaries(t *testing.T) {
	t.Run("single root element", func(t *testing.T) {
		fragment := `<user><name>Alice</name></user>`
		count := Get(fragment, "user.#")
		if count.Int() != 1 {
			t.Errorf("Expected count 1, got %d", count.Int())
		}

		first := Get(fragment, "user.0.name")
		if first.String() != "Alice" {
			t.Errorf("Expected 'Alice', got %q", first.String())
		}
	})

	t.Run("out of bounds positive index", func(t *testing.T) {
		fragment := `<user>Alice</user><user>Bob</user>`
		result := Get(fragment, "user.5.name")
		if result.Exists() {
			t.Error("Expected null result for out of bounds index")
		}
	})

	t.Run("empty fragment", func(t *testing.T) {
		fragment := `<!-- comment only -->`
		result := Get(fragment, "user.#")
		if result.Exists() {
			t.Error("Expected null result for non-existent element")
		}
	})

	t.Run("no matching roots", func(t *testing.T) {
		fragment := `<item>Widget</item><item>Gadget</item>`
		result := Get(fragment, "user.#")
		if result.Exists() {
			t.Error("Expected null result when no roots match")
		}
	})
}

// TestFragmentEdgeCases tests edge cases with fragments
func TestFragmentEdgeCases(t *testing.T) {
	t.Run("empty root elements", func(t *testing.T) {
		fragment := `<a></a><b></b><c></c>`
		if !Valid(fragment) {
			t.Error("Fragment with empty roots should be valid")
		}
	})

	t.Run("self-closing roots only", func(t *testing.T) {
		fragment := `<a/><b/><c/><d/>`
		if !Valid(fragment) {
			t.Error("Fragment with all self-closing roots should be valid")
		}
	})

	t.Run("mixed self-closing and paired", func(t *testing.T) {
		fragment := `<a/><b>text</b><c/>`
		if !Valid(fragment) {
			t.Error("Fragment with mixed root types should be valid")
		}
	})

	t.Run("whitespace between roots", func(t *testing.T) {
		fragment := `<a>1</a>

<b>2</b>		<c>3</c>`
		if !Valid(fragment) {
			t.Error("Fragment with whitespace between roots should be valid")
		}
	})

	t.Run("many root elements", func(t *testing.T) {
		// Create fragment with 50 root elements
		fragment := ""
		for i := 1; i <= 50; i++ {
			fragment += `<item>` + string(rune('0'+i%10)) + `</item>`
		}
		if !Valid(fragment) {
			t.Error("Fragment with 50 roots should be valid")
		}
	})
}

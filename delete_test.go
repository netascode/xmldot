// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"testing"
)

// Test element deletion (P2.7)
func TestDelete_Element(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected string
	}{
		{
			name:     "delete single element",
			xml:      `<root><user><name>John</name><age>30</age></user></root>`,
			path:     "root.user.age",
			expected: `<root><user><name>John</name></user></root>`,
		},
		{
			name:     "delete nested element",
			xml:      `<root><level1><level2><value>test</value></level2></level1></root>`,
			path:     "root.level1.level2",
			expected: `<root><level1></level1></root>`,
		},
		{
			name:     "delete element with children",
			xml:      `<root><user><name>John</name><age>30</age></user><other>data</other></root>`,
			path:     "root.user",
			expected: `<root><other>data</other></root>`,
		},
		{
			name:     "delete from simple structure",
			xml:      `<root><a>1</a><b>2</b><c>3</c></root>`,
			path:     "root.b",
			expected: `<root><a>1</a><c>3</c></root>`,
		},
		{
			name:     "delete self-closing element",
			xml:      `<root><item/><other>data</other></root>`,
			path:     "root.item",
			expected: `<root><other>data</other></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Delete(tt.xml, tt.path)
			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("Delete() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test attribute deletion
func TestDelete_Attribute(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected string
	}{
		{
			name:     "delete single attribute",
			xml:      `<user id="123" active="true"><name>John</name></user>`,
			path:     "user.@active",
			expected: `<user id="123"><name>John</name></user>`,
		},
		{
			name:     "delete only attribute",
			xml:      `<user id="123"><name>John</name></user>`,
			path:     "user.@id",
			expected: `<user><name>John</name></user>`,
		},
		{
			name:     "delete attribute from nested element",
			xml:      `<root><user id="123" active="true"><name>John</name></user></root>`,
			path:     "root.user.@id",
			expected: `<root><user active="true"><name>John</name></user></root>`,
		},
		{
			name:     "delete attribute from self-closing tag",
			xml:      `<item id="1" value="test"/>`,
			path:     "item.@value",
			expected: `<item id="1"/>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Delete(tt.xml, tt.path)
			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
			// Normalize attribute order for comparison
			if !attributesEqual(result, tt.expected) {
				t.Errorf("Delete() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test deletion of non-existent paths (should be no-op)
func TestDelete_NonExistent(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected string
	}{
		{
			name:     "delete non-existent element",
			xml:      `<root><user><name>John</name></user></root>`,
			path:     "root.user.age",
			expected: `<root><user><name>John</name></user></root>`,
		},
		{
			name:     "delete non-existent attribute",
			xml:      `<user id="123"><name>John</name></user>`,
			path:     "user.@active",
			expected: `<user id="123"><name>John</name></user>`,
		},
		{
			name:     "delete with invalid path",
			xml:      `<root><user><name>John</name></user></root>`,
			path:     "root.nonexistent.element",
			expected: `<root><user><name>John</name></user></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Delete(tt.xml, tt.path)
			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("Delete() = %v, want %v (should be unchanged)", result, tt.expected)
			}
		})
	}
}

// Test array element deletion
func TestDelete_ArrayElement(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected string
	}{
		{
			name: "delete first array element",
			xml: `<root>
				<item>first</item>
				<item>second</item>
				<item>third</item>
			</root>`,
			path:     "root.item.0",
			expected: "second", // Should verify second element becomes first
		},
		{
			name: "delete middle array element",
			xml: `<root>
				<item>first</item>
				<item>second</item>
				<item>third</item>
			</root>`,
			path:     "root.item.1",
			expected: "2", // Should verify count reduces to 2
		},
		{
			name: "delete last array element",
			xml: `<root>
				<item>first</item>
				<item>second</item>
				<item>third</item>
			</root>`,
			path:     "root.item.2",
			expected: "2", // Should verify count reduces to 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Delete(tt.xml, tt.path)
			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}

			// Verify by checking either count or first element
			if tt.path == "root.item.0" {
				// Check first element is now what was second
				firstItem := Get(result, "root.item.0")
				if firstItem.String() != tt.expected {
					t.Errorf("After Delete(), first item = %v, want %v", firstItem.String(), tt.expected)
				}
			} else {
				// Check count
				count := Get(result, "root.item.#")
				if count.String() != tt.expected {
					t.Errorf("After Delete(), count = %v, want %v", count.String(), tt.expected)
				}
			}
		})
	}
}

// Test DeleteBytes
func TestDeleteBytes(t *testing.T) {
	xml := []byte(`<root><user><name>John</name><age>30</age></user></root>`)
	expected := []byte(`<root><user><name>John</name></user></root>`)

	result, err := DeleteBytes(xml, "root.user.age")
	if err != nil {
		t.Fatalf("DeleteBytes() error = %v", err)
	}

	if string(result) != string(expected) {
		t.Errorf("DeleteBytes() = %s, want %s", result, expected)
	}
}

// Test error cases
func TestDelete_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		path    string
		wantErr error
	}{
		{
			name:    "empty path",
			xml:     `<root></root>`,
			path:    "",
			wantErr: ErrInvalidPath,
		},
		{
			name:    "document too large",
			xml:     string(make([]byte, MaxDocumentSize+1)),
			path:    "root.test",
			wantErr: ErrMalformedXML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Delete(tt.xml, tt.path)
			if err == nil {
				t.Fatalf("Delete() expected error, got nil")
			}
			if err != tt.wantErr {
				t.Errorf("Delete() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// Integration test: Delete followed by Get
func TestDelete_Integration(t *testing.T) {
	xml := `<root><user><name>John</name><age>30</age><email>john@example.com</email></user></root>`

	// Delete the age element
	modified, err := Delete(xml, "root.user.age")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify age is gone
	age := Get(modified, "root.user.age")
	if age.Exists() {
		t.Errorf("Get() after Delete() should not exist, but got %v", age.String())
	}

	// Verify other elements still exist
	name := Get(modified, "root.user.name")
	if name.String() != "John" {
		t.Errorf("Get() name after Delete() = %v, want John", name.String())
	}

	email := Get(modified, "root.user.email")
	if email.String() != "john@example.com" {
		t.Errorf("Get() email after Delete() = %v, want john@example.com", email.String())
	}
}

// Integration test: Set and Delete combination
func TestSetDelete_Integration(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	// Set a new value
	modified, err := Set(xml, "root.user.age", 30)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify it was set
	age := Get(modified, "root.user.age")
	if age.Int() != 30 {
		t.Errorf("Get() after Set() = %v, want 30", age.Int())
	}

	// Now delete it
	modified2, err := Delete(modified, "root.user.age")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	age2 := Get(modified2, "root.user.age")
	if age2.Exists() {
		t.Errorf("Get() after Delete() should not exist")
	}

	// Verify original data is still there
	name := Get(modified2, "root.user.name")
	if name.String() != "John" {
		t.Errorf("Get() name = %v, want John", name.String())
	}
}

// Test delete last element leaves parent empty
func TestDelete_LastElement(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	result, err := Delete(xml, "root.user.name")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	expected := `<root><user></user></root>`
	if result != expected {
		t.Errorf("Delete() = %v, want %v", result, expected)
	}

	// Verify parent still exists but is empty
	user := Get(result, "root.user")
	if !user.Exists() {
		t.Errorf("Parent element should still exist after deleting last child")
	}
}

// Benchmark Delete operations
func BenchmarkDelete_Element(b *testing.B) {
	xml := `<root><user><name>John</name><age>30</age><email>john@example.com</email></user></root>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Delete(xml, "root.user.age")
	}
}

func BenchmarkDelete_Attribute(b *testing.B) {
	xml := `<root><user id="123" active="true"><name>John</name></user></root>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Delete(xml, "root.user.@active")
	}
}

func BenchmarkDelete_NonExistent(b *testing.B) {
	xml := `<root><user><name>John</name></user></root>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Delete(xml, "root.user.nonexistent")
	}
}

// ============================================================================
// Coverage Tests for DeleteBytesWithOptions
// ============================================================================

// TestDeleteBytesWithOptions tests delete operations with custom options
func TestDeleteBytesWithOptions(t *testing.T) {
	tests := []struct {
		name      string
		xml       string
		path      string
		opts      *Options
		wantMatch string
		wantErr   bool
	}{
		{
			name:      "delete with case-insensitive matching",
			xml:       `<ROOT><USER><NAME>John</NAME><AGE>30</AGE></USER></ROOT>`,
			path:      "root.user.age",
			opts:      &Options{CaseSensitive: false},
			wantMatch: "<NAME>John</NAME>",
			wantErr:   false,
		},
		{
			name:      "delete with default options (fast path)",
			xml:       `<root><user><name>John</name><age>30</age></user></root>`,
			path:      "root.user.age",
			opts:      DefaultOptions(),
			wantMatch: "<name>John</name>",
			wantErr:   false,
		},
		{
			name:      "delete non-existent with case-insensitive",
			xml:       `<ROOT><USER><NAME>John</NAME></USER></ROOT>`,
			path:      "root.user.nonexistent",
			opts:      &Options{CaseSensitive: false},
			wantMatch: "<NAME>John</NAME>",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DeleteBytesWithOptions([]byte(tt.xml), tt.path, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteBytesWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				resultStr := string(result)
				if !stringContains(resultStr, tt.wantMatch) {
					t.Errorf("DeleteBytesWithOptions() result = %q, want match %q", resultStr, tt.wantMatch)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

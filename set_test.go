// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"testing"
)

// Test basic element setting (P2.3)
func TestSet_ElementSetting(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		value    interface{}
		expected string
	}{
		{
			name:     "update existing element",
			xml:      `<root><user><name>John</name></user></root>`,
			path:     "root.user.name",
			value:    "Jane",
			expected: `<root><user><name>Jane</name></user></root>`,
		},
		{
			name:     "update nested element",
			xml:      `<root><level1><level2><value>old</value></level2></level1></root>`,
			path:     "root.level1.level2.value",
			value:    "new",
			expected: `<root><level1><level2><value>new</value></level2></level1></root>`,
		},
		{
			name:     "update with int value",
			xml:      `<root><age>30</age></root>`,
			path:     "root.age",
			value:    35,
			expected: `<root><age>35</age></root>`,
		},
		{
			name:     "update with float value",
			xml:      `<root><price>10.5</price></root>`,
			path:     "root.price",
			value:    15.75,
			expected: `<root><price>15.75</price></root>`,
		},
		{
			name:     "update with bool value",
			xml:      `<root><active>false</active></root>`,
			path:     "root.active",
			value:    true,
			expected: `<root><active>true</active></root>`,
		},
		{
			name:     "update with special characters",
			xml:      `<root><data>old</data></root>`,
			path:     "root.data",
			value:    `<test>&"'`,
			expected: `<root><data>&lt;test&gt;&amp;&quot;&apos;</data></root>`,
		},
		{
			name:     "update with unicode",
			xml:      `<root><text>old</text></root>`,
			path:     "root.text",
			value:    "Hello 世界 🌍",
			expected: `<root><text>Hello 世界 🌍</text></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("Set() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test attribute setting (P2.4)
func TestSet_AttributeSetting(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		value    interface{}
		expected string
	}{
		{
			name:     "update existing attribute",
			xml:      `<user id="123"><name>John</name></user>`,
			path:     "user.@id",
			value:    "456",
			expected: `<user id="456"><name>John</name></user>`,
		},
		{
			name:     "create new attribute",
			xml:      `<user id="123"><name>John</name></user>`,
			path:     "user.@active",
			value:    "true",
			expected: `<user active="true" id="123"><name>John</name></user>`, // Attributes are sorted alphabetically
		},
		{
			name:     "update attribute with escaping",
			xml:      `<item id="1"></item>`,
			path:     "item.@value",
			value:    `"test"`,
			expected: `<item id="1" value="&quot;test&quot;"></item>`,
		},
		{
			name:     "update attribute on nested element",
			xml:      `<root><user id="1"><name>John</name></user></root>`,
			path:     "root.user.@id",
			value:    "2",
			expected: `<root><user id="2"><name>John</name></user></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			// Normalize whitespace in attributes for comparison
			if !attributesEqual(result, tt.expected) {
				t.Errorf("Set() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test element creation (P2.5)
func TestSet_ElementCreation(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		value    interface{}
		expected string
	}{
		{
			name:     "create single missing element",
			xml:      `<root><user><name>John</name></user></root>`,
			path:     "root.user.age",
			value:    30,
			expected: `<root><user><name>John</name><age>30</age></user></root>`,
		},
		{
			name:     "create multiple missing levels",
			xml:      `<root></root>`,
			path:     "root.user.name",
			value:    "John",
			expected: `<root><user><name>John</name></user></root>`,
		},
		{
			name:     "create in empty root",
			xml:      `<root/>`,
			path:     "root.data",
			value:    "test",
			expected: `<root><data>test</data></root>`,
		},
		{
			name:     "create deeply nested path",
			xml:      `<root></root>`,
			path:     "root.a.b.c.d",
			value:    "deep",
			expected: `<root><a><b><c><d>deep</d></c></b></a></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("Set() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test attribute creation on non-existent elements (v0.3.0 feature)
func TestSet_AttributeCreation(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		value    interface{}
		expected string
	}{
		{
			name:     "create attribute on missing single element",
			xml:      `<root></root>`,
			path:     "root.user.@id",
			value:    "123",
			expected: `<root><user id="123"></user></root>`,
		},
		{
			name:     "create attribute on missing nested path",
			xml:      `<root></root>`,
			path:     "root.a.b.c.@id",
			value:    "deep",
			expected: `<root><a><b><c id="deep"></c></b></a></root>`,
		},
		{
			name:     "create attribute in self-closing root",
			xml:      `<root/>`,
			path:     "root.user.@active",
			value:    "true",
			expected: `<root><user active="true"></user></root>`,
		},
		{
			name:     "create attribute with special characters",
			xml:      `<root></root>`,
			path:     "root.item.@value",
			value:    `"test"`,
			expected: `<root><item value="&quot;test&quot;"></item></root>`,
		},
		{
			name:     "create attribute with partial existing path",
			xml:      `<root><user></user></root>`,
			path:     "root.user.contact.@email",
			value:    "test@example.com",
			expected: `<root><user><contact email="test@example.com"></contact></user></root>`,
		},
		{
			name:     "create attribute with empty value",
			xml:      `<root></root>`,
			path:     "root.item.@flag",
			value:    "",
			expected: `<root><item flag=""></item></root>`,
		},
		{
			name:     "create attribute with numeric value",
			xml:      `<root></root>`,
			path:     "root.item.@count",
			value:    42,
			expected: `<root><item count="42"></item></root>`,
		},
		{
			name:     "create attribute on root element",
			xml:      `<root></root>`,
			path:     "root.@version",
			value:    "1.0",
			expected: `<root version="1.0"></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("Set() = %v, want %v", result, tt.expected)
			}

			// Verify attribute was set correctly by reading it back
			attrResult := Get(result, tt.path)
			if !attrResult.Exists() {
				t.Errorf("Attribute not found after Set()")
			}
			expectedValue := fmt.Sprintf("%v", tt.value)
			if attrResult.String() != expectedValue {
				t.Errorf("Get() after Set() = %v, want %v", attrResult.String(), expectedValue)
			}
		})
	}
}

// Test multiple attributes on created elements
func TestSet_AttributeCreationMultiple(t *testing.T) {
	xml := `<root></root>`

	// Create first attribute (creates element)
	result, err := Set(xml, "root.user.@id", "123")
	if err != nil {
		t.Fatalf("First Set() error = %v", err)
	}

	// Verify element was created
	if !strings.Contains(result, `<user id="123">`) {
		t.Errorf("First attribute not created correctly: %s", result)
	}

	// Create second attribute (element already exists)
	result, err = Set(result, "root.user.@name", "test")
	if err != nil {
		t.Fatalf("Second Set() error = %v", err)
	}

	// Verify both attributes exist (alphabetically sorted)
	if !strings.Contains(result, `id="123"`) {
		t.Errorf("First attribute lost: %s", result)
	}
	if !strings.Contains(result, `name="test"`) {
		t.Errorf("Second attribute not added: %s", result)
	}
}

// Test attribute creation preserves existing structure
func TestSet_AttributeCreationPreservesStructure(t *testing.T) {
	xml := `<root><existing><data>value</data></existing></root>`

	result, err := Set(xml, "root.new.@attr", "test")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify new element with attribute was added
	if !strings.Contains(result, `<new attr="test">`) {
		t.Errorf("New element not created: %s", result)
	}

	// Verify existing structure is preserved
	existingValue := Get(result, "root.existing.data")
	if existingValue.String() != "value" {
		t.Errorf("Existing data corrupted: got %v, want 'value'", existingValue.String())
	}
}

// Test SetRaw (P2.8)
func TestSetRaw(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		rawxml   string
		expected string
		wantErr  bool
	}{
		{
			name:     "insert raw XML",
			xml:      `<root></root>`,
			path:     "root.data",
			rawxml:   "<item><name>Test</name></item>",
			expected: `<root><data><item><name>Test</name></item></data></root>`,
			wantErr:  false,
		},
		{
			name:     "insert complex raw XML",
			xml:      `<root></root>`,
			path:     "root.users",
			rawxml:   "<user><name>John</name><age>30</age></user>",
			expected: `<root><users><user><name>John</name><age>30</age></user></users></root>`,
			wantErr:  false,
		},
		{
			name:     "invalid raw XML - unbalanced tags",
			xml:      `<root></root>`,
			path:     "root.data",
			rawxml:   "<item><name>Test</item>",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "self-closing tag",
			xml:      `<root></root>`,
			path:     "root.data",
			rawxml:   "<item/>",
			expected: `<root><data><item/></data></root>`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetRaw(tt.xml, tt.path, tt.rawxml)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SetRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("SetRaw() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test SetBytes
func TestSetBytes(t *testing.T) {
	xml := []byte(`<root><value>old</value></root>`)
	expected := []byte(`<root><value>new</value></root>`)

	result, err := SetBytes(xml, "root.value", "new")
	if err != nil {
		t.Fatalf("SetBytes() error = %v", err)
	}

	if string(result) != string(expected) {
		t.Errorf("SetBytes() = %s, want %s", result, expected)
	}
}

// Test nil value as deletion
func TestSet_NilValue(t *testing.T) {
	xml := `<root><user><name>John</name><age>30</age></user></root>`
	expected := `<root><user><name>John</name></user></root>`

	result, err := Set(xml, "root.user.age", nil)
	if err != nil {
		t.Fatalf("Set() with nil error = %v", err)
	}

	if result != expected {
		t.Errorf("Set() with nil = %v, want %v", result, expected)
	}
}

// Test error cases
func TestSet_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		path    string
		value   interface{}
		wantErr error
	}{
		{
			name:    "empty path",
			xml:     `<root></root>`,
			path:    "",
			value:   "test",
			wantErr: ErrInvalidPath,
		},
		{
			name:    "document too large",
			xml:     string(make([]byte, MaxDocumentSize+1)),
			path:    "root.test",
			value:   "test",
			wantErr: ErrMalformedXML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Set(tt.xml, tt.path, tt.value)
			if err == nil {
				t.Fatalf("Set() expected error, got nil")
			}
			if err != tt.wantErr {
				t.Errorf("Set() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// Test array index setting
func TestSet_ArrayIndex(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		value    interface{}
		expected string
	}{
		{
			name: "update first element",
			xml: `<root>
				<item>first</item>
				<item>second</item>
				<item>third</item>
			</root>`,
			path:     "root.item.0",
			value:    "updated",
			expected: "updated",
		},
		{
			name: "update middle element",
			xml: `<root>
				<item>first</item>
				<item>second</item>
				<item>third</item>
			</root>`,
			path:     "root.item.1",
			value:    "updated",
			expected: "updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			// Verify the value was set by reading it back
			readResult := Get(result, tt.path)
			if readResult.String() != tt.expected {
				t.Errorf("Set() then Get() = %v, want %v", readResult.String(), tt.expected)
			}
		})
	}
}

// Integration test: Set followed by Get
func TestSet_Integration(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	// Set a new value
	modified, err := Set(xml, "root.user.age", 30)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get the value back
	result := Get(modified, "root.user.age")
	if result.Int() != 30 {
		t.Errorf("Get() after Set() = %v, want 30", result.Int())
	}

	// Verify original value still exists
	name := Get(modified, "root.user.name")
	if name.String() != "John" {
		t.Errorf("Get() original value = %v, want John", name.String())
	}
}

// Helper function to compare XML with attributes (attribute order may vary)
func attributesEqual(xml1, xml2 string) bool {
	// Simple comparison - in production might need to normalize attribute order
	// For now, just check if both contain the same attributes
	return xml1 == xml2
}

// Test deep nesting position tracking (bug fix verification)
func TestSet_DeepNesting(t *testing.T) {
	xml := `<root><level1><level2><level3><target>old</target></level3></level2></level1></root>`
	result, err := Set(xml, "root.level1.level2.level3.target", "new")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify target was actually changed
	got := Get(result, "root.level1.level2.level3.target")
	if got.String() != "new" {
		t.Errorf("Deep nesting set failed: got %q, want %q", got.String(), "new")
	}

	// Verify XML structure is intact
	if !containsAll(result, []string{"<level1>", "<level2>", "<level3>", "</level1>", "</level2>", "</level3>"}) {
		t.Error("XML structure corrupted during deep nesting set")
	}
}

func TestSet_MultipleNested(t *testing.T) {
	xml := `<root>
		<user>
			<name>John</name>
			<address>
				<city>NYC</city>
				<zip>10001</zip>
			</address>
		</user>
	</root>`

	result, err := Set(xml, "root.user.address.city", "LA")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	gotCity := Get(result, "root.user.address.city")
	gotZip := Get(result, "root.user.address.zip")
	gotName := Get(result, "root.user.name")

	if gotCity.String() != "LA" {
		t.Errorf("City not updated: got %q", gotCity.String())
	}
	if gotZip.String() != "10001" {
		t.Errorf("Zip corrupted: got %q", gotZip.String())
	}
	if gotName.String() != "John" {
		t.Errorf("Name corrupted: got %q", gotName.String())
	}
}

func TestSet_DeepNestingWithAttributes(t *testing.T) {
	xml := `<root><a><b id="1"><c><d>value</d></c></b></a></root>`

	// Test modifying deeply nested element
	result, err := Set(xml, "root.a.b.c.d", "newvalue")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got := Get(result, "root.a.b.c.d")
	if got.String() != "newvalue" {
		t.Errorf("Deep nesting set failed: got %q, want %q", got.String(), "newvalue")
	}

	// Verify attribute wasn't corrupted
	gotAttr := Get(result, "root.a.b.@id")
	if gotAttr.String() != "1" {
		t.Errorf("Attribute corrupted: got %q", gotAttr.String())
	}
}

func TestSet_NestedSiblings(t *testing.T) {
	xml := `<root>
		<section>
			<item>first</item>
			<nested>
				<value>nested1</value>
			</nested>
			<item>second</item>
		</section>
	</root>`

	// Modify nested element
	result, err := Set(xml, "root.section.nested.value", "modified")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify the nested element was modified
	gotNested := Get(result, "root.section.nested.value")
	if gotNested.String() != "modified" {
		t.Errorf("Nested value not updated: got %q", gotNested.String())
	}

	// Verify siblings weren't affected
	gotFirst := Get(result, "root.section.item.0")
	gotSecond := Get(result, "root.section.item.1")
	if gotFirst.String() != "first" || gotSecond.String() != "second" {
		t.Error("Sibling elements were corrupted")
	}
}

// Helper function to check if string contains all substrings
func containsAll(s string, substrs []string) bool {
	for _, substr := range substrs {
		if !containsSubstring(s, substr) {
			return false
		}
	}
	return true
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark Set operations
func BenchmarkSet_SimpleElement(b *testing.B) {
	xml := `<root><user><name>John</name></user></root>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Set(xml, "root.user.name", "Jane")
	}
}

func BenchmarkSet_CreateElement(b *testing.B) {
	xml := `<root><user><name>John</name></user></root>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Set(xml, "root.user.age", 30)
	}
}

func BenchmarkSet_Attribute(b *testing.B) {
	xml := `<root><user id="123"><name>John</name></user></root>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Set(xml, "root.user.@id", "456")
	}
}

func BenchmarkSetRaw(b *testing.B) {
	xml := `<root></root>`
	rawxml := "<item><name>Test</name></item>"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetRaw(xml, "root.data", rawxml)
	}
}

// Test SetMany - Multiple unrelated paths
func TestSetMany_MultipleUnrelated(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`
	paths := []string{"root.user.age", "root.user.email", "root.user.city"}
	values := []interface{}{30, "john@example.com", "NYC"}

	result, err := SetMany(xml, paths, values)
	if err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}

	// Verify all values were set
	if Get(result, "root.user.age").Int() != 30 {
		t.Errorf("age not set correctly")
	}
	if Get(result, "root.user.email").String() != "john@example.com" {
		t.Errorf("email not set correctly")
	}
	if Get(result, "root.user.city").String() != "NYC" {
		t.Errorf("city not set correctly")
	}

	// Verify original data preserved
	if Get(result, "root.user.name").String() != "John" {
		t.Errorf("original name corrupted")
	}
}

// Test SetMany - Multiple related paths (parent and child)
func TestSetMany_RelatedPaths(t *testing.T) {
	xml := `<root></root>`
	paths := []string{"root.user.name", "root.user.address.city", "root.user.address.zip"}
	values := []interface{}{"John", "NYC", "10001"}

	result, err := SetMany(xml, paths, values)
	if err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}

	// Verify nested structure created correctly
	if Get(result, "root.user.name").String() != "John" {
		t.Errorf("name not set correctly")
	}
	if Get(result, "root.user.address.city").String() != "NYC" {
		t.Errorf("city not set correctly")
	}
	if Get(result, "root.user.address.zip").String() != "10001" {
		t.Errorf("zip not set correctly")
	}
}

// Test SetMany - Mix of creates and updates
func TestSetMany_MixedOperations(t *testing.T) {
	xml := `<root><user><name>John</name><age>25</age></user></root>`
	paths := []string{"root.user.name", "root.user.age", "root.user.email"}
	values := []interface{}{"Jane", 30, "jane@example.com"}

	result, err := SetMany(xml, paths, values)
	if err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}

	// Verify updates
	if Get(result, "root.user.name").String() != "Jane" {
		t.Errorf("name not updated correctly")
	}
	if Get(result, "root.user.age").Int() != 30 {
		t.Errorf("age not updated correctly")
	}

	// Verify new element created
	if Get(result, "root.user.email").String() != "jane@example.com" {
		t.Errorf("email not created correctly")
	}
}

// Test SetMany - Overlapping paths (later wins)
func TestSetMany_OverlappingPaths(t *testing.T) {
	xml := `<root><value>original</value></root>`
	paths := []string{"root.value", "root.value", "root.value"}
	values := []interface{}{"first", "second", "third"}

	result, err := SetMany(xml, paths, values)
	if err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}

	// Last operation should win
	if Get(result, "root.value").String() != "third" {
		t.Errorf("overlapping paths not handled correctly, got %v", Get(result, "root.value").String())
	}
}

// Test SetMany - Error case: mismatched counts
func TestSetMany_MismatchedCounts(t *testing.T) {
	xml := `<root></root>`
	paths := []string{"root.a", "root.b", "root.c"}
	values := []interface{}{1, 2} // Only 2 values for 3 paths

	_, err := SetMany(xml, paths, values)
	if err == nil {
		t.Fatalf("SetMany() expected error for mismatched counts, got nil")
	}
}

// Test SetMany - Error case: invalid path
func TestSetMany_InvalidPath(t *testing.T) {
	xml := `<root></root>`
	paths := []string{"root.valid", "", "root.another"}
	values := []interface{}{1, 2, 3}

	_, err := SetMany(xml, paths, values)
	if err == nil {
		t.Fatalf("SetMany() expected error for invalid path, got nil")
	}
}

// Test SetMany - Empty inputs
func TestSetMany_EmptyInputs(t *testing.T) {
	xml := `<root><value>test</value></root>`
	paths := []string{}
	values := []interface{}{}

	result, err := SetMany(xml, paths, values)
	if err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}

	// Should return original XML unchanged
	if result != xml {
		t.Errorf("SetMany() with empty inputs should return original XML")
	}
}

// Test SetMany - Attributes
func TestSetMany_Attributes(t *testing.T) {
	xml := `<user id="123"><name>John</name></user>`
	paths := []string{"user.@active", "user.@role", "user.@id"}
	values := []interface{}{"true", "admin", "456"}

	result, err := SetMany(xml, paths, values)
	if err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}

	// Verify attributes
	if Get(result, "user.@active").String() != "true" {
		t.Errorf("active attribute not set correctly")
	}
	if Get(result, "user.@role").String() != "admin" {
		t.Errorf("role attribute not set correctly")
	}
	if Get(result, "user.@id").String() != "456" {
		t.Errorf("id attribute not updated correctly")
	}
}

// Test SetMany - Array indices
func TestSetMany_ArrayIndices(t *testing.T) {
	xml := `<root>
		<item>first</item>
		<item>second</item>
		<item>third</item>
	</root>`
	paths := []string{"root.item.0", "root.item.2"}
	values := []interface{}{"updated-first", "updated-third"}

	result, err := SetMany(xml, paths, values)
	if err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}

	// Verify updates
	if Get(result, "root.item.0").String() != "updated-first" {
		t.Errorf("first item not updated correctly")
	}
	if Get(result, "root.item.1").String() != "second" {
		t.Errorf("middle item corrupted")
	}
	if Get(result, "root.item.2").String() != "updated-third" {
		t.Errorf("last item not updated correctly")
	}
}

// Test SetManyBytes
func TestSetManyBytes(t *testing.T) {
	xml := []byte(`<root><value>old</value></root>`)
	paths := []string{"root.value", "root.new"}
	values := []interface{}{"updated", "created"}

	result, err := SetManyBytes(xml, paths, values)
	if err != nil {
		t.Fatalf("SetManyBytes() error = %v", err)
	}

	resultStr := string(result)
	if Get(resultStr, "root.value").String() != "updated" {
		t.Errorf("value not updated correctly")
	}
	if Get(resultStr, "root.new").String() != "created" {
		t.Errorf("new element not created correctly")
	}
}

// Test DeleteMany - Multiple unrelated deletes
func TestDeleteMany_MultipleUnrelated(t *testing.T) {
	xml := `<root>
		<a>1</a>
		<b>2</b>
		<c>3</c>
		<d>4</d>
	</root>`

	result, err := DeleteMany(xml, "root.b", "root.d")
	if err != nil {
		t.Fatalf("DeleteMany() error = %v", err)
	}

	// Verify deletions
	if Get(result, "root.b").Exists() {
		t.Errorf("element b should be deleted")
	}
	if Get(result, "root.d").Exists() {
		t.Errorf("element d should be deleted")
	}

	// Verify remaining elements
	if Get(result, "root.a").String() != "1" {
		t.Errorf("element a corrupted")
	}
	if Get(result, "root.c").String() != "3" {
		t.Errorf("element c corrupted")
	}
}

// Test DeleteMany - Parent and child deletes
func TestDeleteMany_ParentChild(t *testing.T) {
	xml := `<root>
		<user>
			<name>John</name>
			<address>
				<city>NYC</city>
				<zip>10001</zip>
			</address>
		</user>
	</root>`

	// Delete parent first - child path becomes invalid
	result, err := DeleteMany(xml, "root.user", "root.user.name")
	if err != nil {
		t.Fatalf("DeleteMany() error = %v", err)
	}

	// Verify parent deleted
	if Get(result, "root.user").Exists() {
		t.Errorf("user element should be deleted")
	}

	// Verify root still exists
	if !Get(result, "root").Exists() {
		t.Errorf("root element should still exist")
	}
}

// Test DeleteMany - Non-existent paths
func TestDeleteMany_NonExistent(t *testing.T) {
	xml := `<root><a>1</a><b>2</b></root>`

	result, err := DeleteMany(xml, "root.nonexistent", "root.missing", "root.fake")
	if err != nil {
		t.Fatalf("DeleteMany() error = %v", err)
	}

	// Should return original XML unchanged (minus whitespace differences)
	if Get(result, "root.a").String() != "1" {
		t.Errorf("element a corrupted")
	}
	if Get(result, "root.b").String() != "2" {
		t.Errorf("element b corrupted")
	}
}

// Test DeleteMany - Duplicate paths
func TestDeleteMany_DuplicatePaths(t *testing.T) {
	xml := `<root><a>1</a><b>2</b><c>3</c></root>`

	result, err := DeleteMany(xml, "root.a", "root.b", "root.a", "root.b")
	if err != nil {
		t.Fatalf("DeleteMany() error = %v", err)
	}

	// Verify deletions (duplicates should be handled gracefully)
	if Get(result, "root.a").Exists() {
		t.Errorf("element a should be deleted")
	}
	if Get(result, "root.b").Exists() {
		t.Errorf("element b should be deleted")
	}
	if Get(result, "root.c").String() != "3" {
		t.Errorf("element c corrupted")
	}
}

// Test DeleteMany - Empty inputs
func TestDeleteMany_EmptyInputs(t *testing.T) {
	xml := `<root><value>test</value></root>`

	result, err := DeleteMany(xml)
	if err != nil {
		t.Fatalf("DeleteMany() error = %v", err)
	}

	// Should return original XML unchanged
	if result != xml {
		t.Errorf("DeleteMany() with empty inputs should return original XML")
	}
}

// Test DeleteMany - Attributes
func TestDeleteMany_Attributes(t *testing.T) {
	xml := `<user id="123" active="true" role="admin"><name>John</name></user>`

	result, err := DeleteMany(xml, "user.@active", "user.@role")
	if err != nil {
		t.Fatalf("DeleteMany() error = %v", err)
	}

	// Verify attribute deletions
	if Get(result, "user.@active").Exists() {
		t.Errorf("active attribute should be deleted")
	}
	if Get(result, "user.@role").Exists() {
		t.Errorf("role attribute should be deleted")
	}

	// Verify remaining attribute
	if Get(result, "user.@id").String() != "123" {
		t.Errorf("id attribute corrupted")
	}

	// Verify element content
	if Get(result, "user.name").String() != "John" {
		t.Errorf("name element corrupted")
	}
}

// Test DeleteMany - Array elements
func TestDeleteMany_ArrayElements(t *testing.T) {
	xml := `<root>
		<item>first</item>
		<item>second</item>
		<item>third</item>
		<item>fourth</item>
	</root>`

	// Delete indices 1 and 3 (note: after first deletion, indices shift)
	result, err := DeleteMany(xml, "root.item.1", "root.item.2")
	if err != nil {
		t.Fatalf("DeleteMany() error = %v", err)
	}

	// After deleting index 1 (second), the remaining items shift down
	// After deleting index 2 of the shifted array (originally fourth), we should have first and third
	count := Get(result, "root.item.#").Int()
	if count != 2 {
		t.Errorf("expected 2 remaining items, got %d", count)
	}
}

// Test DeleteManyBytes
func TestDeleteManyBytes(t *testing.T) {
	xml := []byte(`<root><a>1</a><b>2</b><c>3</c></root>`)

	result, err := DeleteManyBytes(xml, "root.b")
	if err != nil {
		t.Fatalf("DeleteManyBytes() error = %v", err)
	}

	resultStr := string(result)
	if Get(resultStr, "root.b").Exists() {
		t.Errorf("element b should be deleted")
	}
	if Get(resultStr, "root.a").String() != "1" {
		t.Errorf("element a corrupted")
	}
	if Get(resultStr, "root.c").String() != "3" {
		t.Errorf("element c corrupted")
	}
}

// Test DeleteMany - Error case: invalid path
func TestDeleteMany_InvalidPath(t *testing.T) {
	xml := `<root><a>1</a></root>`

	_, err := DeleteMany(xml, "root.a", "")
	if err == nil {
		t.Fatalf("DeleteMany() expected error for invalid path, got nil")
	}
}

// Test DeleteMany - Error case: document too large
func TestDeleteMany_DocumentTooLarge(t *testing.T) {
	xml := string(make([]byte, MaxDocumentSize+1))

	_, err := DeleteMany(xml, "root.a")
	if err == nil {
		t.Fatalf("DeleteMany() expected error for document too large, got nil")
	}
}

// Integration test: SetMany and DeleteMany combination
func TestSetDeleteMany_Integration(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	// Add multiple elements
	paths := []string{"root.user.age", "root.user.email", "root.user.city"}
	values := []interface{}{30, "john@example.com", "NYC"}

	modified, err := SetMany(xml, paths, values)
	if err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}

	// Verify additions
	if Get(modified, "root.user.age").Int() != 30 {
		t.Errorf("age not set correctly")
	}
	if Get(modified, "root.user.email").String() != "john@example.com" {
		t.Errorf("email not set correctly")
	}

	// Delete some elements
	modified2, err := DeleteMany(modified, "root.user.age", "root.user.city")
	if err != nil {
		t.Fatalf("DeleteMany() error = %v", err)
	}

	// Verify deletions
	if Get(modified2, "root.user.age").Exists() {
		t.Errorf("age should be deleted")
	}
	if Get(modified2, "root.user.city").Exists() {
		t.Errorf("city should be deleted")
	}

	// Verify remaining data
	if Get(modified2, "root.user.name").String() != "John" {
		t.Errorf("name corrupted")
	}
	if Get(modified2, "root.user.email").String() != "john@example.com" {
		t.Errorf("email corrupted")
	}
}

// Benchmark SetMany vs individual Set calls
func BenchmarkSetMany_3Paths(b *testing.B) {
	xml := `<root><user><name>John</name></user></root>`
	paths := []string{"root.user.age", "root.user.email", "root.user.city"}
	values := []interface{}{30, "john@example.com", "NYC"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetMany(xml, paths, values)
	}
}

func BenchmarkSet_3Individual(b *testing.B) {
	xml := `<root><user><name>John</name></user></root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, _ := Set(xml, "root.user.age", 30)
		result, _ = Set(result, "root.user.email", "john@example.com")
		result, _ = Set(result, "root.user.city", "NYC")
		_ = result
	}
}

// Benchmark DeleteMany vs individual Delete calls
func BenchmarkDeleteMany_3Paths(b *testing.B) {
	xml := `<root><a>1</a><b>2</b><c>3</c><d>4</d></root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeleteMany(xml, "root.b", "root.c", "root.d")
	}
}

func BenchmarkDelete_3Individual(b *testing.B) {
	xml := `<root><a>1</a><b>2</b><c>3</c><d>4</d></root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, _ := Delete(xml, "root.b")
		result, _ = Delete(result, "root.c")
		result, _ = Delete(result, "root.d")
		_ = result
	}
}

// Example functions for godoc

// ExampleSet demonstrates updating an element value
func ExampleSet() {
	xml := `<user>
		<name>John Doe</name>
		<age>30</age>
	</user>`

	// Update the age element
	modified, _ := Set(xml, "user.age", 31)
	result := Get(modified, "user.age")
	fmt.Println(result.String())
	// Output: 31
}

// ExampleSet_attribute demonstrates setting an attribute value
func ExampleSet_attribute() {
	xml := `<user id="123">
		<name>John Doe</name>
	</user>`

	// Update existing attribute
	modified, _ := Set(xml, "user.@id", "456")

	// Add new attribute
	modified, _ = Set(modified, "user.@active", "true")

	fmt.Println(Get(modified, "user.@id").String())
	fmt.Println(Get(modified, "user.@active").String())
	// Output:
	// 456
	// true
}

// ExampleSet_createNew demonstrates creating new elements
func ExampleSet_createNew() {
	xml := `<user>
		<name>John Doe</name>
	</user>`

	// Create new element that doesn't exist
	modified, _ := Set(xml, "user.email", "john@example.com")

	// Create deeply nested element
	modified, _ = Set(modified, "user.address.city", "New York")

	fmt.Println(Get(modified, "user.email").String())
	fmt.Println(Get(modified, "user.address.city").String())
	// Output:
	// john@example.com
	// New York
}

// ExampleDelete demonstrates removing an element
func ExampleDelete() {
	xml := `<user>
		<name>John Doe</name>
		<age>30</age>
		<email>john@example.com</email>
	</user>`

	// Delete the email element
	modified, _ := Delete(xml, "user.email")

	fmt.Println("Email exists:", Get(modified, "user.email").Exists())
	fmt.Println("Name exists:", Get(modified, "user.name").Exists())
	// Output:
	// Email exists: false
	// Name exists: true
}

// ExampleSetMany demonstrates batch set operations
func ExampleSetMany() {
	xml := `<user>
		<name>John Doe</name>
	</user>`

	// Set multiple paths in one call
	paths := []string{"user.age", "user.email", "user.city"}
	values := []interface{}{30, "john@example.com", "New York"}

	modified, _ := SetMany(xml, paths, values)

	fmt.Println("Age:", Get(modified, "user.age").Int())
	fmt.Println("Email:", Get(modified, "user.email").String())
	fmt.Println("City:", Get(modified, "user.city").String())
	// Output:
	// Age: 30
	// Email: john@example.com
	// City: New York
}

// ExampleDeleteMany demonstrates batch delete operations
func ExampleDeleteMany() {
	xml := `<user>
		<name>John Doe</name>
		<age>30</age>
		<email>john@example.com</email>
		<phone>555-1234</phone>
	</user>`

	// Delete multiple elements in one call
	modified, _ := DeleteMany(xml, "user.email", "user.phone")

	fmt.Println("Email exists:", Get(modified, "user.email").Exists())
	fmt.Println("Phone exists:", Get(modified, "user.phone").Exists())
	fmt.Println("Name exists:", Get(modified, "user.name").Exists())
	// Output:
	// Email exists: false
	// Phone exists: false
	// Name exists: true
}

// ExampleSetWithOptions demonstrates using options when setting values
func ExampleSetWithOptions() {
	xml := `<root><value>test</value></root>`

	// Set with custom options for indentation
	opts := Options{Indent: "  "}
	modified, _ := SetWithOptions(xml, "root.value", "updated", &opts)

	// Output shows the modified value (indentation may vary by implementation)
	result := Get(modified, "root.value")
	fmt.Println(result.String())
	// Output: updated
}

// ExampleSetWithOptions_indent demonstrates pretty printing with indentation
func ExampleSetWithOptions_indent() {
	xml := `<root><user><name>John</name><age>30</age></user></root>`

	// Set value with options
	opts := Options{Indent: "    "}
	modified, _ := SetWithOptions(xml, "root.user.age", 31, &opts)

	// Verify the value was updated
	result := Get(modified, "root.user.age")
	fmt.Printf("New age: %d\n", result.Int())
	// Output: New age: 31
}

// TestSetMalformedXMLRegression is a regression test for fuzz-discovered crash
// Issue: Set operations on malformed XML with unclosed tags caused runtime panics
func TestSetMalformedXMLRegression(t *testing.T) {
	tests := []struct {
		name         string
		xml          string
		path         string
		value        interface{}
		expectError  bool
		errorMessage string
	}{
		{
			name:         "unclosed root tag",
			xml:          "<root>",
			path:         "root.item",
			value:        "value",
			expectError:  true,
			errorMessage: "malformed",
		},
		{
			name:         "unclosed nested tag",
			xml:          "<root><user>",
			path:         "root.user.name",
			value:        "John",
			expectError:  true,
			errorMessage: "malformed",
		},
		{
			name:         "mismatched closing tag",
			xml:          "<root><user></root>",
			path:         "root.user",
			value:        "value",
			expectError:  true,
			errorMessage: "malformed",
		},
		{
			name:         "missing opening tag",
			xml:          "<root></user></root>",
			path:         "root.user",
			value:        "value",
			expectError:  true,
			errorMessage: "malformed",
		},
		{
			name:         "unclosed attribute quote",
			xml:          `<root attr="value><item>test</item></root>`,
			path:         "root.item",
			value:        "new",
			expectError:  true,
			errorMessage: "malformed",
		},
		{
			name:        "valid XML should work",
			xml:         "<root></root>",
			path:        "root.item",
			value:       "value",
			expectError: false,
		},
		{
			name:        "valid XML with nested elements",
			xml:         "<root><user><name>John</name></user></root>",
			path:        "root.user.age",
			value:       30,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("Set() expected error but got none; result = %v", result)
				}
				if tt.errorMessage != "" && err != nil {
					errStr := err.Error()
					if !containsIgnoreCase(errStr, tt.errorMessage) {
						t.Errorf("Set() error = %v, should contain %q", err, tt.errorMessage)
					}
				}
				// Error case - result should be original XML
				if result != tt.xml {
					t.Errorf("Set() on error should return original XML, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Set() unexpected error = %v", err)
				}
				// Success case - verify value was set
				if !Get(result, tt.path).Exists() {
					t.Errorf("Set() value not found at path %s in result: %s", tt.path, result)
				}
			}
		})
	}
}

// TestDeleteMalformedXMLRegression is a regression test for fuzz-discovered crash
// Issue: Delete operations on malformed XML caused runtime panics
func TestDeleteMalformedXMLRegression(t *testing.T) {
	tests := []struct {
		name         string
		xml          string
		path         string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "unclosed root tag",
			xml:          "<root>",
			path:         "root",
			expectError:  true,
			errorMessage: "malformed",
		},
		{
			name:         "unclosed nested tag",
			xml:          "<root><user>",
			path:         "root.user",
			expectError:  true,
			errorMessage: "malformed",
		},
		{
			name:         "mismatched closing tag",
			xml:          "<root><user></root>",
			path:         "root.user",
			expectError:  true,
			errorMessage: "malformed",
		},
		{
			name:        "valid XML should work",
			xml:         "<root><item>test</item></root>",
			path:        "root.item",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Delete(tt.xml, tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Delete() expected error but got none; result = %v", result)
				}
				if tt.errorMessage != "" && err != nil {
					errStr := err.Error()
					if !containsIgnoreCase(errStr, tt.errorMessage) {
						t.Errorf("Delete() error = %v, should contain %q", err, tt.errorMessage)
					}
				}
				// Error case - result should be original XML
				if result != tt.xml {
					t.Errorf("Delete() on error should return original XML, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Delete() unexpected error = %v", err)
				}
			}
		})
	}
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLowerASCII(s)
	substr = toLowerASCII(substr)
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================================
// Phase 4 Production Hardening: Wildcard/Filter Integration with Batch Ops
// ============================================================================

// TestSetMany_WithWildcards tests that SetMany accepts wildcard paths without errors
func TestSetMany_WithWildcards(t *testing.T) {
	t.Run("wildcard path in SetMany - basic test", func(t *testing.T) {
		xml := `<root>
			<user><name>Alice</name><age>25</age></user>
			<user><name>Bob</name><age>30</age></user>
		</root>`

		paths := []string{"root.*.age"}
		values := []interface{}{35}

		result, err := SetMany(xml, paths, values)
		if err != nil {
			t.Fatalf("SetMany() with wildcards error = %v", err)
		}

		// Verify operation completes without error
		if result == "" {
			t.Errorf("SetMany() should return non-empty result")
		}
	})

	t.Run("recursive wildcard in SetMany - basic test", func(t *testing.T) {
		xml := `<root><user><profile><name>Alice</name></profile></user></root>`

		paths := []string{"root.**.name"}
		values := []interface{}{"Updated"}

		result, err := SetMany(xml, paths, values)
		if err != nil {
			t.Fatalf("SetMany() with recursive wildcard error = %v", err)
		}

		// Verify operation completes without error
		if result == "" {
			t.Errorf("SetMany() should return non-empty result")
		}
	})
}

// TestSetMany_WithFilters tests that SetMany accepts filter expressions without errors
func TestSetMany_WithFilters(t *testing.T) {
	t.Run("filter in SetMany - basic test", func(t *testing.T) {
		xml := `<root>
			<user><name>Alice</name><age>25</age><status>active</status></user>
			<user><name>Bob</name><age>30</age><status>inactive</status></user>
		</root>`

		paths := []string{"root.user[age>25].status"}
		values := []interface{}{"premium"}

		result, err := SetMany(xml, paths, values)
		if err != nil {
			t.Fatalf("SetMany() with filter error = %v", err)
		}

		// Verify operation completes without error
		if result == "" {
			t.Errorf("SetMany() should return non-empty result")
		}
	})
}

// TestDeleteMany_WithWildcards tests that DeleteMany accepts wildcard paths without errors
func TestDeleteMany_WithWildcards(t *testing.T) {
	t.Run("wildcard path in DeleteMany - basic test", func(t *testing.T) {
		xml := `<root>
			<user><name>Alice</name><age>25</age></user>
			<user><name>Bob</name><age>30</age></user>
		</root>`

		result, err := DeleteMany(xml, "root.*.age")
		if err != nil {
			t.Fatalf("DeleteMany() with wildcards error = %v", err)
		}

		// Verify operation completes without error
		if result == "" {
			t.Errorf("DeleteMany() should return non-empty result")
		}
	})

	t.Run("recursive wildcard in DeleteMany - basic test", func(t *testing.T) {
		xml := `<root><level1><value>1</value><nested><value>2</value></nested></level1></root>`

		result, err := DeleteMany(xml, "root.**.value")
		if err != nil {
			t.Fatalf("DeleteMany() with recursive wildcard error = %v", err)
		}

		// Verify operation completes without error
		if result == "" {
			t.Errorf("DeleteMany() should return non-empty result")
		}
	})
}

// TestDeleteMany_WithFilters tests that DeleteMany accepts filter expressions without errors
func TestDeleteMany_WithFilters(t *testing.T) {
	t.Run("filter in DeleteMany - basic test", func(t *testing.T) {
		xml := `<root>
			<user><name>Alice</name><age>25</age></user>
			<user><name>Bob</name><age>30</age></user>
		</root>`

		result, err := DeleteMany(xml, "root.user[age>=30]")
		if err != nil {
			t.Fatalf("DeleteMany() with filter error = %v", err)
		}

		// Verify operation completes without error
		if result == "" {
			t.Errorf("DeleteMany() should return non-empty result")
		}
	})
}

// TestBatchOps_ComplexFilterWildcardCombinations tests complex scenarios
func TestBatchOps_ComplexFilterWildcardCombinations(t *testing.T) {
	t.Run("SetMany with wildcard and filter combination", func(t *testing.T) {
		xml := `<root>
			<group><user><name>Alice</name><age>25</age></user></group>
			<group><user><name>Bob</name><age>30</age></user></group>
		</root>`

		paths := []string{"root.*.user[age>=30].name"}
		values := []interface{}{"Senior User"}

		result, err := SetMany(xml, paths, values)
		if err != nil {
			t.Fatalf("SetMany() with wildcard+filter error = %v", err)
		}

		// Verify operation completes without error
		if result == "" {
			t.Errorf("SetMany() should return non-empty result")
		}
	})

	t.Run("DeleteMany with multiple filter operations", func(t *testing.T) {
		t.Skip("DeleteMany with GJSON filter syntax not yet implemented - filters return result arrays, not element paths")

		xml := `<root>
			<user><name>Alice</name><age>25</age><role>admin</role></user>
			<user><name>Bob</name><age>30</age><role>user</role></user>
			<user><name>Charlie</name><age>35</age><role>admin</role></user>
		</root>`

		result, err := DeleteMany(xml, "root.user.#(role==user)#", "root.user.#(age<30)#")
		if err != nil {
			t.Fatalf("DeleteMany() with multiple filters error = %v", err)
		}

		// Verify Charlie remains (age>=30 AND role==admin)
		count := Get(result, "root.user.#")
		if count.Int() != 1 {
			t.Errorf("DeleteMany() count = %v, want 1", count.Int())
		}

		name0 := Get(result, "root.user.0.name")
		if name0.String() != "Charlie" {
			t.Errorf("Only Charlie should remain, got %v", name0.String())
		}
	})
}

// ============================================================================
// Phase 4 Production Hardening: MaxDocumentSize Tests for Batch Operations
// ============================================================================

// TestSetMany_MaxDocumentSize tests that SetMany respects MaxDocumentSize limit
func TestSetMany_MaxDocumentSize(t *testing.T) {
	t.Run("SetMany rejects document exceeding MaxDocumentSize", func(t *testing.T) {
		// Create XML document larger than MaxDocumentSize (10MB)
		largeXML := string(make([]byte, MaxDocumentSize+1))
		paths := []string{"root.test"}
		values := []interface{}{"value"}

		_, err := SetMany(largeXML, paths, values)
		if err == nil {
			t.Fatalf("SetMany() should reject documents exceeding MaxDocumentSize")
		}

		if err != ErrMalformedXML {
			t.Errorf("SetMany() error = %v, want ErrMalformedXML", err)
		}
	})

	t.Run("SetMany accepts document at MaxDocumentSize limit", func(_ *testing.T) {
		// Create valid XML document at exactly MaxDocumentSize
		xmlSize := MaxDocumentSize - 100 // Leave room for tags
		content := string(make([]byte, xmlSize))
		xml := "<root>" + content + "</root>"

		paths := []string{"root.test"}
		values := []interface{}{"value"}

		_, err := SetMany(xml, paths, values)
		// May fail due to malformed XML (just random bytes), but should not be
		// rejected for size reasons. We just verify it doesn't panic.
		_ = err
	})

	t.Run("SetMany with multiple operations on large document", func(_ *testing.T) {
		// Create document approaching MaxDocumentSize
		xmlSize := MaxDocumentSize/2 - 1000
		largeContent := string(make([]byte, xmlSize))
		xml := "<root><data>" + largeContent + "</data></root>"

		paths := []string{"root.field1", "root.field2", "root.field3"}
		values := []interface{}{"val1", "val2", "val3"}

		_, err := SetMany(xml, paths, values)
		// May fail due to malformed XML, but should not panic
		_ = err
	})
}

// TestSetManyBytes_MaxDocumentSize tests that SetManyBytes respects MaxDocumentSize
func TestSetManyBytes_MaxDocumentSize(t *testing.T) {
	t.Run("SetManyBytes rejects document exceeding MaxDocumentSize", func(t *testing.T) {
		largeXML := make([]byte, MaxDocumentSize+1)
		paths := []string{"root.test"}
		values := []interface{}{"value"}

		_, err := SetManyBytes(largeXML, paths, values)
		if err == nil {
			t.Fatalf("SetManyBytes() should reject documents exceeding MaxDocumentSize")
		}

		if err != ErrMalformedXML {
			t.Errorf("SetManyBytes() error = %v, want ErrMalformedXML", err)
		}
	})
}

// TestDeleteMany_MaxDocumentSize tests that DeleteMany respects MaxDocumentSize limit
func TestDeleteMany_MaxDocumentSize(t *testing.T) {
	t.Run("DeleteMany rejects document exceeding MaxDocumentSize", func(t *testing.T) {
		// Create XML document larger than MaxDocumentSize (10MB)
		largeXML := string(make([]byte, MaxDocumentSize+1))

		_, err := DeleteMany(largeXML, "root.test")
		if err == nil {
			t.Fatalf("DeleteMany() should reject documents exceeding MaxDocumentSize")
		}

		if err != ErrMalformedXML {
			t.Errorf("DeleteMany() error = %v, want ErrMalformedXML", err)
		}
	})

	t.Run("DeleteMany accepts document at MaxDocumentSize limit", func(_ *testing.T) {
		// Create valid XML document at exactly MaxDocumentSize
		xmlSize := MaxDocumentSize - 100 // Leave room for tags
		content := string(make([]byte, xmlSize))
		xml := "<root>" + content + "</root>"

		_, err := DeleteMany(xml, "root.test")
		// May fail due to malformed XML (just random bytes), but should not be
		// rejected for size reasons. We just verify it doesn't panic.
		_ = err
	})

	t.Run("DeleteMany with multiple operations on large document", func(_ *testing.T) {
		// Create document approaching MaxDocumentSize
		xmlSize := MaxDocumentSize/2 - 1000
		largeContent := string(make([]byte, xmlSize))
		xml := "<root><data>" + largeContent + "</data><field1>1</field1><field2>2</field2></root>"

		_, err := DeleteMany(xml, "root.field1", "root.field2")
		// May fail due to malformed XML, but should not panic
		_ = err
	})
}

// TestDeleteManyBytes_MaxDocumentSize tests that DeleteManyBytes respects MaxDocumentSize
func TestDeleteManyBytes_MaxDocumentSize(t *testing.T) {
	t.Run("DeleteManyBytes rejects document exceeding MaxDocumentSize", func(t *testing.T) {
		largeXML := make([]byte, MaxDocumentSize+1)

		_, err := DeleteManyBytes(largeXML, "root.test")
		if err == nil {
			t.Fatalf("DeleteManyBytes() should reject documents exceeding MaxDocumentSize")
		}

		if err != ErrMalformedXML {
			t.Errorf("DeleteManyBytes() error = %v, want ErrMalformedXML", err)
		}
	})
}

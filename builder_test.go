// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"testing"
)

// Test valueToXML conversion
func TestValueToXML(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
		isRaw    bool
		wantErr  bool
	}{
		{
			name:     "string value",
			value:    "test",
			expected: "test",
			isRaw:    false,
			wantErr:  false,
		},
		{
			name:     "string with special characters",
			value:    `<test>&"'`,
			expected: `&lt;test&gt;&amp;&quot;&apos;`,
			isRaw:    false,
			wantErr:  false,
		},
		{
			name:     "int value",
			value:    42,
			expected: "42",
			isRaw:    false,
			wantErr:  false,
		},
		{
			name:     "int64 value",
			value:    int64(9223372036854775807),
			expected: "9223372036854775807",
			isRaw:    false,
			wantErr:  false,
		},
		{
			name:     "float64 value",
			value:    3.14159,
			expected: "3.14159",
			isRaw:    false,
			wantErr:  false,
		},
		{
			name:     "float32 value",
			value:    float32(2.718),
			expected: "2.718",
			isRaw:    false,
			wantErr:  false,
		},
		{
			name:     "bool true",
			value:    true,
			expected: "true",
			isRaw:    false,
			wantErr:  false,
		},
		{
			name:     "bool false",
			value:    false,
			expected: "false",
			isRaw:    false,
			wantErr:  false,
		},
		{
			name:     "byte slice as raw XML",
			value:    []byte("<item>test</item>"),
			expected: "<item>test</item>",
			isRaw:    true,
			wantErr:  false,
		},
		{
			name:     "nil value",
			value:    nil,
			expected: "",
			isRaw:    false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, isRaw, err := valueToXML(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("valueToXML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if result != tt.expected {
					t.Errorf("valueToXML() = %v, want %v", result, tt.expected)
				}
				if isRaw != tt.isRaw {
					t.Errorf("valueToXML() isRaw = %v, want %v", isRaw, tt.isRaw)
				}
			}
		})
	}
}

// Test validateRawXML
func TestValidateRawXML(t *testing.T) {
	tests := []struct {
		name    string
		rawxml  string
		wantErr bool
	}{
		{
			name:    "valid simple XML",
			rawxml:  "<item>test</item>",
			wantErr: false,
		},
		{
			name:    "valid nested XML",
			rawxml:  "<item><name>test</name><value>123</value></item>",
			wantErr: false,
		},
		{
			name:    "valid self-closing tag",
			rawxml:  "<item/>",
			wantErr: false,
		},
		{
			name:    "valid multiple elements",
			rawxml:  "<item>1</item><item>2</item>",
			wantErr: false,
		},
		{
			name:    "invalid - unmatched opening tag",
			rawxml:  "<item>test",
			wantErr: true,
		},
		{
			name:    "invalid - unmatched closing tag",
			rawxml:  "test</item>",
			wantErr: true,
		},
		{
			name:    "invalid - mismatched tags",
			rawxml:  "<item>test</other>",
			wantErr: true,
		},
		{
			name:    "valid with comment",
			rawxml:  "<item><!-- comment --></item>",
			wantErr: false,
		},
		{
			name:    "valid with CDATA",
			rawxml:  "<item><![CDATA[data]]></item>",
			wantErr: false,
		},
		{
			name:    "valid with processing instruction",
			rawxml:  "<?xml version=\"1.0\"?><item>test</item>",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRawXML(tt.rawxml)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRawXML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test findElementLocation
func TestBuilder_FindElementLocation(t *testing.T) {
	tests := []struct {
		name      string
		xml       string
		path      string
		wantFound bool
	}{
		{
			name:      "find simple element",
			xml:       `<root><user><name>John</name></user></root>`,
			path:      "root.user.name",
			wantFound: true,
		},
		{
			name:      "find nested element",
			xml:       `<root><level1><level2><value>test</value></level2></level1></root>`,
			path:      "root.level1.level2.value",
			wantFound: true,
		},
		{
			name:      "find element with attributes",
			xml:       `<root><user id="123"><name>John</name></user></root>`,
			path:      "root.user",
			wantFound: true,
		},
		{
			name:      "find self-closing element",
			xml:       `<root><item/></root>`,
			path:      "root.item",
			wantFound: true,
		},
		{
			name:      "element not found",
			xml:       `<root><user><name>John</name></user></root>`,
			path:      "root.user.age",
			wantFound: false,
		},
		{
			name:      "find first of array",
			xml:       `<root><item>1</item><item>2</item><item>3</item></root>`,
			path:      "root.item.0",
			wantFound: true,
		},
		{
			name:      "find middle of array",
			xml:       `<root><item>1</item><item>2</item><item>3</item></root>`,
			path:      "root.item.1",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := newXMLBuilder([]byte(tt.xml))
			parser := newXMLParser(builder.data)
			segments := parsePath(tt.path)

			location, found := builder.findElementLocation(parser, segments, 0, 0)
			if found != tt.wantFound {
				t.Errorf("findElementLocation() found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantFound && location == nil {
				t.Errorf("findElementLocation() location should not be nil when found")
			}
			if !tt.wantFound && location != nil {
				t.Errorf("findElementLocation() location should be nil when not found")
			}
		})
	}
}

// Test buildElementPath
func TestBuilder_BuildElementPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		value    string
		isRaw    bool
		expected string
	}{
		{
			name:     "single element",
			path:     "item",
			value:    "test",
			isRaw:    false,
			expected: "<item>test</item>",
		},
		{
			name:     "nested elements",
			path:     "user.name",
			value:    "John",
			isRaw:    false,
			expected: "<user><name>John</name></user>",
		},
		{
			name:     "deeply nested",
			path:     "a.b.c.d",
			value:    "deep",
			isRaw:    false,
			expected: "<a><b><c><d>deep</d></c></b></a>",
		},
		{
			name:     "with raw XML",
			path:     "data",
			value:    "<item>test</item>",
			isRaw:    true,
			expected: "<data><item>test</item></data>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := newXMLBuilder([]byte{})
			segments := parsePath(tt.path)

			builder.buildElementPath(segments, tt.value, tt.isRaw)
			result := builder.result.String()

			if result != tt.expected {
				t.Errorf("buildElementPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test element replacement
func TestBuilder_ReplaceElement(t *testing.T) {
	xml := `<root><value>old</value></root>`
	builder := newXMLBuilder([]byte(xml))
	parser := newXMLParser(builder.data)
	segments := parsePath("root.value")

	location, found := builder.findElementLocation(parser, segments, 0, 0)
	if !found {
		t.Fatal("Element not found")
	}

	err := builder.replaceElement(location, segments[len(segments)-1], "new")
	if err != nil {
		t.Fatalf("replaceElement() error = %v", err)
	}

	result := builder.getResult()
	expected := `<root><value>new</value></root>`
	if result != expected {
		t.Errorf("replaceElement() result = %v, want %v", result, expected)
	}
}

// Test attribute replacement
func TestBuilder_ReplaceAttribute(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		attrName string
		value    string
		expected string
	}{
		{
			name:     "update existing attribute",
			xml:      `<user id="123"><name>John</name></user>`,
			path:     "user",
			attrName: "id",
			value:    "456",
			expected: `<user id="456"><name>John</name></user>`,
		},
		{
			name:     "add new attribute",
			xml:      `<user id="123"><name>John</name></user>`,
			path:     "user",
			attrName: "active",
			value:    "true",
			expected: `<user active="true" id="123"><name>John</name></user>`, // Attributes are sorted alphabetically
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := newXMLBuilder([]byte(tt.xml))
			parser := newXMLParser(builder.data)
			segments := parsePath(tt.path)

			location, found := builder.findElementLocation(parser, segments, 0, 0)
			if !found {
				t.Fatal("Element not found")
			}

			err := builder.replaceAttribute(location, tt.attrName, tt.value)
			if err != nil {
				t.Fatalf("replaceAttribute() error = %v", err)
			}

			result := builder.getResult()
			// Attributes may be in different order, so check both contain same attributes
			if !attributesEqual(result, tt.expected) {
				t.Errorf("replaceAttribute() result = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test delete attribute
func TestBuilder_DeleteAttribute(t *testing.T) {
	xml := `<user id="123" active="true"><name>John</name></user>`
	builder := newXMLBuilder([]byte(xml))
	parser := newXMLParser(builder.data)
	segments := parsePath("user")

	location, found := builder.findElementLocation(parser, segments, 0, 0)
	if !found {
		t.Fatal("Element not found")
	}

	err := builder.deleteAttribute(location, "active")
	if err != nil {
		t.Fatalf("deleteAttribute() error = %v", err)
	}

	result := builder.getResult()
	expected := `<user id="123"><name>John</name></user>`
	if !attributesEqual(result, expected) {
		t.Errorf("deleteAttribute() result = %v, want %v", result, expected)
	}
}

// Test element creation
func TestBuilder_CreateElement(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		value    string
		expected string
	}{
		{
			name:     "create in empty root",
			xml:      `<root></root>`,
			path:     "root.item",
			value:    "test",
			expected: `<root><item>test</item></root>`,
		},
		{
			name:     "create nested path",
			xml:      `<root></root>`,
			path:     "root.user.name",
			value:    "John",
			expected: `<root><user><name>John</name></user></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := newXMLBuilder([]byte(tt.xml))
			segments := parsePath(tt.path)

			err := builder.createElement(segments, tt.value, false)
			if err != nil {
				t.Fatalf("createElement() error = %v", err)
			}

			result := builder.getResult()
			if result != tt.expected {
				t.Errorf("createElement() result = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Benchmark builder operations
func BenchmarkBuilder_FindElement(b *testing.B) {
	xml := []byte(`<root><user><name>John</name><age>30</age></user></root>`)
	segments := parsePath("root.user.name")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := newXMLBuilder(xml)
		parser := newXMLParser(builder.data)
		builder.findElementLocation(parser, segments, 0, 0)
	}
}

func BenchmarkBuilder_ReplaceElement(b *testing.B) {
	xml := []byte(`<root><value>old</value></root>`)
	segments := parsePath("root.value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := newXMLBuilder(xml)
		parser := newXMLParser(builder.data)
		location, _ := builder.findElementLocation(parser, segments, 0, 0)
		_ = builder.replaceElement(location, segments[len(segments)-1], "new")
	}
}

func BenchmarkBuilder_CreateElement(b *testing.B) {
	xml := []byte(`<root></root>`)
	segments := parsePath("root.item")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := newXMLBuilder(xml)
		_ = builder.createElement(segments, "test", false)
	}
}

// ============================================================================
// Coverage Tests for Missing Functions
// ============================================================================

// TestCreateInRoot tests the createInRoot function for self-closing root elements
func TestCreateInRoot(t *testing.T) {
	tests := []struct {
		name      string
		xml       string
		path      string
		value     string
		wantErr   bool
		wantMatch string
	}{
		{
			name:      "create in self-closing root",
			xml:       `<root/>`,
			path:      "root.child",
			value:     "test",
			wantErr:   false,
			wantMatch: "<child>test</child>",
		},
		{
			name:      "create nested in self-closing root",
			xml:       `<root/>`,
			path:      "root.level1.level2",
			value:     "nested",
			wantErr:   false,
			wantMatch: "<level2>nested</level2>",
		},
		{
			name:      "create in self-closing root with attributes",
			xml:       `<root id="123"/>`,
			path:      "root.child",
			value:     "test",
			wantErr:   false,
			wantMatch: "<child>test</child>",
		},
		{
			name:      "path matches root exactly",
			xml:       `<root/>`,
			path:      "root",
			value:     "test",
			wantErr:   false,
			wantMatch: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !contains(result, tt.wantMatch) {
					t.Errorf("Set() result = %q, want match %q", result, tt.wantMatch)
				}
			}
		})
	}
}

// TestGetResultUnchanged tests the getResult function when XML is unchanged
func TestGetResultUnchanged(t *testing.T) {
	xml := `<root><child>value</child></root>`

	// Delete non-existent element (no changes)
	result, err := Delete(xml, "root.nonexistent")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Should return unchanged XML
	if result != xml {
		t.Errorf("Expected unchanged XML when deleting non-existent element")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"testing"
)

// Basic Validation Tests

func TestValidSimpleXML(t *testing.T) {
	xml := "<root>value</root>"
	if !Valid(xml) {
		t.Error("Simple valid XML should pass validation")
	}
	if err := ValidateWithError(xml); err != nil {
		t.Errorf("Simple valid XML should pass validation: %v", err)
	}
}

func TestValidComplexXML(t *testing.T) {
	xml := `<root>
		<child1>value1</child1>
		<child2>
			<nested>value2</nested>
		</child2>
	</root>`
	if !Valid(xml) {
		t.Error("Complex nested XML should pass validation")
	}
}

func TestValidWithAttributes(t *testing.T) {
	xml := `<root attr="value"><child id="123">text</child></root>`
	if !Valid(xml) {
		t.Error("XML with attributes should pass validation")
	}
}

func TestValidEmptyDocument(t *testing.T) {
	xml := ""
	if Valid(xml) {
		t.Error("Empty document should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Empty document should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "empty document") {
		t.Errorf("Expected 'empty document' error, got: %s", err.Message)
	}
}

func TestValidWhitespaceOnly(t *testing.T) {
	xml := "   \n\t  "
	if Valid(xml) {
		t.Error("Whitespace-only document should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Whitespace-only document should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "no root element") {
		t.Errorf("Expected 'no root element' error, got: %s", err.Message)
	}
}

func TestValidRootOnly(t *testing.T) {
	xml := "<root/>"
	if !Valid(xml) {
		t.Error("Single self-closing root element should be valid")
	}

	xml2 := "<root></root>"
	if !Valid(xml2) {
		t.Error("Single root element with closing tag should be valid")
	}
}

// Unclosed Tag Tests

func TestInvalidUnclosedRootTag(t *testing.T) {
	xml := "<root><child></root>"
	if Valid(xml) {
		t.Error("XML with unclosed child tag should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Unclosed child tag should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "mismatched") {
		t.Errorf("Expected 'mismatched' error, got: %s", err.Message)
	}
}

func TestInvalidMissingClosingTag(t *testing.T) {
	xml := "<root><child>"
	if Valid(xml) {
		t.Error("XML with missing closing tags should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Missing closing tags should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "unclosed") {
		t.Errorf("Expected 'unclosed' error, got: %s", err.Message)
	}
}

func TestInvalidUnclosedSelfClosing(t *testing.T) {
	// Verify self-closing tags ARE valid
	xml := "<root/>"
	if !Valid(xml) {
		t.Error("Self-closing root tag should be valid")
	}

	xml2 := "<root><child/></root>"
	if !Valid(xml2) {
		t.Error("Self-closing child tag should be valid")
	}
}

func TestInvalidNestedUnclosed(t *testing.T) {
	xml := "<root><level1><level2><level3></level2></level1></root>"
	if Valid(xml) {
		t.Error("XML with unclosed nested tag should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Unclosed nested tag should return validation error")
	}
}

// Mismatched Tag Tests

func TestInvalidMismatchedTag(t *testing.T) {
	xml := "<a></b>"
	if Valid(xml) {
		t.Error("XML with mismatched tags should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Mismatched tags should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "mismatched") {
		t.Errorf("Expected 'mismatched' error, got: %s", err.Message)
	}
}

func TestInvalidCaseMismatch(t *testing.T) {
	xml := "<Root></root>"
	if Valid(xml) {
		t.Error("XML is case-sensitive, Root != root should fail")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Case mismatch should return validation error")
	}
}

func TestInvalidNestedMismatch(t *testing.T) {
	xml := "<a><b></a></b>"
	if Valid(xml) {
		t.Error("XML with overlapping tags should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Overlapping tags should return validation error")
	}
}

func TestInvalidOrderMismatch(t *testing.T) {
	xml := "<a><b></b></c>"
	if Valid(xml) {
		t.Error("XML with wrong closing tag should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Wrong closing tag should return validation error")
	}
}

// Invalid Name Tests

func TestInvalidElementNameSpace(t *testing.T) {
	xml := "<element name>value</element name>"
	if Valid(xml) {
		t.Error("Element name with space should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Element name with space should return validation error")
	}
}

func TestInvalidElementNameSpecialChar(t *testing.T) {
	xml := "<element@name>value</element@name>"
	if Valid(xml) {
		t.Error("Element name with @ should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Element name with @ should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "invalid") {
		t.Errorf("Expected 'invalid' error, got: %s", err.Message)
	}
}

func TestInvalidElementNameStartsDigit(t *testing.T) {
	xml := "<1element>value</1element>"
	if Valid(xml) {
		t.Error("Element name starting with digit should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Element name starting with digit should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "digit") {
		t.Errorf("Expected 'digit' error, got: %s", err.Message)
	}
}

func TestInvalidEmptyElementName(t *testing.T) {
	xml := "<>value</>"
	if Valid(xml) {
		t.Error("Empty element name should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Empty element name should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "empty") {
		t.Errorf("Expected 'empty' error, got: %s", err.Message)
	}
}

func TestInvalidAttributeNameSpecialChar(t *testing.T) {
	xml := `<element attr@name="value">text</element>`
	if Valid(xml) {
		t.Error("Attribute name with @ should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Attribute name with @ should return validation error")
	}
}

// Error Location Tests

func TestValidateErrorLineNumber(t *testing.T) {
	xml := "<root>\n<child>\n</root>"
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Mismatched tags should return validation error")
	}
	if err != nil && err.Line != 3 {
		t.Errorf("Expected error on line 3, got line %d", err.Line)
	}
}

func TestValidateErrorColumnNumber(t *testing.T) {
	xml := "<root><child></root>"
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Mismatched tags should return validation error")
	}
	if err != nil {
		// Error should be on the closing tag
		if err.Line != 1 {
			t.Errorf("Expected error on line 1, got line %d", err.Line)
		}
		// Column should be at the closing tag position
		if err.Column < 13 {
			t.Errorf("Expected error column >= 13, got column %d", err.Column)
		}
	}
}

func TestValidateErrorMultiline(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("<root>\n")
	for i := 0; i < 8; i++ {
		sb.WriteString("  <valid>content</valid>\n")
	}
	sb.WriteString("  <broken>\n") // Line 10
	sb.WriteString("</root>")

	xml := sb.String()
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Unclosed tag should return validation error")
	}
	if err != nil && err.Line != 11 {
		t.Errorf("Expected error on line 11, got line %d: %s", err.Line, err.Message)
	}
}

func TestValidateErrorDeepNesting(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 10; i++ {
		sb.WriteString("<level>")
	}
	sb.WriteString("<broken>")
	for i := 0; i < 10; i++ {
		sb.WriteString("</level>")
	}
	sb.WriteString("</root>")

	xml := sb.String()
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Unclosed deeply nested tag should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "mismatched") {
		t.Errorf("Expected 'mismatched' error for deeply nested structure, got: %s", err.Message)
	}
}

func TestValidateErrorMessage(t *testing.T) {
	xml := "<root><child></root>"
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Mismatched tags should return validation error")
	}
	if err != nil {
		// Error message should be descriptive
		if !strings.Contains(err.Message, "mismatched") {
			t.Errorf("Error message should contain 'mismatched': %s", err.Message)
		}
		// Error message should mention both tag names
		if !strings.Contains(err.Message, "root") || !strings.Contains(err.Message, "child") {
			t.Errorf("Error message should mention both 'root' and 'child': %s", err.Message)
		}
	}
}

// Edge Cases

func TestValidXMLWithCDATA(t *testing.T) {
	xml := "<root><![CDATA[<tag>not parsed</tag>]]></root>"
	if !Valid(xml) {
		t.Error("XML with CDATA should be valid")
	}
}

func TestValidXMLWithComments(t *testing.T) {
	xml := "<root><!-- comment --><child>value</child></root>"
	if !Valid(xml) {
		t.Error("XML with comments should be valid")
	}
}

func TestValidXMLWithProcessingInstructions(t *testing.T) {
	xml := `<?xml version="1.0"?><root>value</root>`
	if !Valid(xml) {
		t.Error("XML with processing instructions should be valid")
	}
}

func TestValidateMultipleErrors(t *testing.T) {
	// XML with multiple errors - should report the first one
	xml := "<root><a><b></root>"
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("XML with multiple errors should return validation error")
	}
	// Should report first error encountered (mismatched closing tag)
	if err != nil && !strings.Contains(err.Message, "mismatched") {
		t.Errorf("Expected first error to be reported, got: %s", err.Message)
	}
}

// Security Validation Tests

func TestValidateRespectsSecurityLimits(t *testing.T) {
	// Test MaxDocumentSize
	tooLarge := make([]byte, MaxDocumentSize+1)
	for i := range tooLarge {
		tooLarge[i] = 'x'
	}
	xml := string(tooLarge)
	if Valid(xml) {
		t.Error("Document exceeding MaxDocumentSize should fail validation")
	}

	// Test MaxNestingDepth
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < MaxNestingDepth+10; i++ {
		sb.WriteString("<level>")
	}
	sb.WriteString("value")
	for i := 0; i < MaxNestingDepth+10; i++ {
		sb.WriteString("</level>")
	}
	sb.WriteString("</root>")
	deepXML := sb.String()

	err := ValidateWithError(deepXML)
	if err == nil {
		t.Error("XML exceeding MaxNestingDepth should fail validation")
	}
	if err != nil && !strings.Contains(err.Message, "nesting depth") {
		t.Errorf("Expected 'nesting depth' error, got: %s", err.Message)
	}
}

func TestValidateLargeDocument(t *testing.T) {
	// Create a 1MB valid document
	var sb strings.Builder
	sb.WriteString("<root>")
	elementSize := 50 // Approximate size per element
	numElements := (1024 * 1024) / elementSize

	for i := 0; i < numElements; i++ {
		sb.WriteString("<item>data</item>")
	}
	sb.WriteString("</root>")

	xml := sb.String()

	// Should validate successfully (under 10MB limit)
	if !Valid(xml) {
		t.Error("1MB valid document should pass validation")
	}
}

// XML Fragment Support Tests (Multiple Root Elements)

func TestValidFragments(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{
			name: "two root elements",
			xml:  "<root1></root1><root2></root2>",
		},
		{
			name: "three root elements with whitespace",
			xml:  "<root1>A</root1>  <root2>B</root2>  <root3>C</root3>",
		},
		{
			name: "fragment with prolog",
			xml:  `<?xml version="1.0"?><root1>A</root1><root2>B</root2>`,
		},
		{
			name: "fragment with comments",
			xml:  `<root1>A</root1><!-- comment --><root2>B</root2>`,
		},
		{
			name: "nested elements in fragment",
			xml:  `<root1><child>A</child></root1><root2><child>B</child></root2>`,
		},
		{
			name: "self-closing roots",
			xml:  `<root1/><root2/><root3/>`,
		},
		{
			name: "mixed self-closing and paired",
			xml:  `<root1/><root2>X</root2><root3/>`,
		},
		{
			name: "empty roots",
			xml:  `<root1></root1><root2></root2>`,
		},
		{
			name: "fragment with attributes",
			xml:  `<user id="1">Alice</user><user id="2">Bob</user>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !Valid(tt.xml) {
				t.Errorf("Fragment should be valid: %s", tt.xml)
			}
			err := ValidateWithError(tt.xml)
			if err != nil {
				t.Errorf("Fragment should pass validation: %v", err)
			}
		})
	}
}

func TestInvalidFragmentWithTextBetweenRoots(t *testing.T) {
	xml := `<root1>A</root1>invalid text<root2>B</root2>`
	if Valid(xml) {
		t.Error("Text content between root elements should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Text between roots should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "outside root element") {
		t.Errorf("Expected 'outside root element' error, got: %s", err.Message)
	}
}

// Content Outside Root Element Test

func TestInvalidContentOutsideRoot(t *testing.T) {
	xml := "<root></root>text outside"
	if Valid(xml) {
		t.Error("Content outside root element should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Content outside root should return validation error")
	}
	if err != nil && !strings.Contains(err.Message, "outside root") {
		t.Errorf("Expected 'outside root' error, got: %s", err.Message)
	}
}

// Attribute Validation Tests

func TestInvalidAttributeWithoutEquals(t *testing.T) {
	xml := `<root attr"value"></root>`
	if Valid(xml) {
		t.Error("Attribute without = should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Attribute without = should return validation error")
	}
}

func TestInvalidAttributeWithoutQuotes(t *testing.T) {
	xml := `<root attr=value></root>`
	if Valid(xml) {
		t.Error("Unquoted attribute value should fail validation")
	}
	err := ValidateWithError(xml)
	if err == nil {
		t.Error("Unquoted attribute should return validation error")
	}
}

func TestValidAttributeWithSingleQuotes(t *testing.T) {
	xml := `<root attr='value'></root>`
	if !Valid(xml) {
		t.Error("Attribute with single quotes should be valid")
	}
}

// Namespace Validation Tests (basic)

func TestValidXMLWithNamespaces(t *testing.T) {
	xml := `<root xmlns="http://example.com"><child>value</child></root>`
	if !Valid(xml) {
		t.Error("XML with namespace should be valid")
	}
}

func TestValidXMLWithPrefixedNamespaces(t *testing.T) {
	xml := `<ns:root xmlns:ns="http://example.com"><ns:child>value</ns:child></ns:root>`
	if !Valid(xml) {
		t.Error("XML with prefixed namespaces should be valid")
	}
}

// ValidBytes tests

func TestValidBytes(t *testing.T) {
	xml := []byte("<root>value</root>")
	if !ValidBytes(xml) {
		t.Error("ValidBytes should work with byte slice")
	}
}

func TestValidateBytesWithError(t *testing.T) {
	xml := []byte("<root><child></root>")
	err := ValidateBytesWithError(xml)
	if err == nil {
		t.Error("ValidateBytesWithError should detect errors")
	}
}

// Benchmark Tests

func BenchmarkValid10KB(b *testing.B) {
	// Create 10KB valid XML document
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 200; i++ {
		sb.WriteString("<item>")
		sb.WriteString("<name>Product Name</name>")
		sb.WriteString("<price>99.99</price>")
		sb.WriteString("</item>")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Valid(xml)
	}
}

func BenchmarkValid100KB(b *testing.B) {
	// Create 100KB valid XML document
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 2000; i++ {
		sb.WriteString("<item>")
		sb.WriteString("<name>Product Name</name>")
		sb.WriteString("<price>99.99</price>")
		sb.WriteString("</item>")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Valid(xml)
	}
}

func BenchmarkValid1MB(b *testing.B) {
	// Create 1MB valid XML document
	var sb strings.Builder
	sb.WriteString("<root>")
	elementSize := 50
	numElements := (1024 * 1024) / elementSize
	for i := 0; i < numElements; i++ {
		sb.WriteString("<item>data</item>")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Valid(xml)
	}
}

func BenchmarkValidateWithError(b *testing.B) {
	xml := "<root><child>value</child></root>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateWithError(xml)
	}
}

func BenchmarkValidBytes(b *testing.B) {
	xml := []byte("<root><child>value</child></root>")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidBytes(xml)
	}
}

func BenchmarkValidVsValidateWithError(b *testing.B) {
	// Create a large valid XML document (10KB)
	largeXML := "<root>"
	for i := 0; i < 100; i++ {
		largeXML += "<item><name>Item " + itoa(i) + "</name><value>" + itoa(i*100) + "</value></item>"
	}
	largeXML += "</root>"

	b.Run("Valid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Valid(largeXML)
		}
	})

	b.Run("ValidateWithError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ValidateWithError(largeXML)
		}
	})
}

// Example functions for godoc

// ExampleValid demonstrates quick XML validation
func ExampleValid() {
	// Valid XML
	validXML := "<root><child>value</child></root>"
	fmt.Println("Valid:", Valid(validXML))

	// Invalid XML (unclosed tag)
	invalidXML := "<root><child></root>"
	fmt.Println("Invalid:", Valid(invalidXML))
	// Output:
	// Valid: true
	// Invalid: false
}

// ExampleValidateWithError demonstrates detailed error reporting
func ExampleValidateWithError() {
	// Invalid XML with unclosed tag
	xml := `<root><person><name>John</name></root>`

	err := ValidateWithError(xml)
	if err != nil {
		fmt.Printf("Validation error: %s\n", err.Message)
	}
	// Output: Validation error: mismatched closing tag 'root' (expected 'person' opened at line 1, column 6)
}

// ExampleValid_fragment demonstrates validation of XML fragments with multiple roots
func ExampleValid_fragment() {
	// XML fragment with multiple root elements
	fragment := `<user id="1">Alice</user><user id="2">Bob</user>`

	if Valid(fragment) {
		fmt.Println("Fragment is valid")
	}
	// Output: Fragment is valid
}

// ExampleValid_fragmentWithProlog demonstrates fragment validation with XML declaration
func ExampleValid_fragmentWithProlog() {
	fragment := `<?xml version="1.0"?>
<item>first</item>
<item>second</item>`

	if Valid(fragment) {
		fmt.Println("Fragment with prolog is valid")
	}
	// Output: Fragment with prolog is valid
}

// ExampleValidateWithError_fragmentWithText demonstrates that text between roots is invalid
func ExampleValidateWithError_fragmentWithText() {
	// Text content between root elements is not allowed
	fragment := `<item>first</item>invalid text<item>second</item>`

	err := ValidateWithError(fragment)
	if err != nil {
		fmt.Println("Text between roots is not allowed")
	}
	// Output: Text between roots is not allowed
}

// ============================================================================
// Coverage Tests for Missing Functions
// ============================================================================

// TestValidateErrorError tests the ValidateError.Error() method
func TestValidateErrorError(t *testing.T) {
	err := &ValidateError{
		Line:    5,
		Column:  12,
		Message: "unclosed tag 'root'",
	}

	expected := "XML validation error at line 5, column 12: unclosed tag 'root'"
	if err.Error() != expected {
		t.Errorf("Error() = %q, expected %q", err.Error(), expected)
	}
}

// TestValidateDOCTYPE tests validation with DOCTYPE declarations
func TestValidateDOCTYPE(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		wantErr bool
	}{
		{
			name:    "valid DOCTYPE simple",
			xml:     `<!DOCTYPE root SYSTEM "root.dtd"><root></root>`,
			wantErr: false,
		},
		{
			name:    "valid DOCTYPE with internal subset",
			xml:     `<!DOCTYPE root [<!ELEMENT root (#PCDATA)>]><root>text</root>`,
			wantErr: false,
		},
		{
			name: "valid DOCTYPE with multiple entities",
			xml: `<!DOCTYPE root [
				<!ELEMENT root (#PCDATA)>
				<!ENTITY test "value">
			]><root>text</root>`,
			wantErr: false,
		},
		{
			name:    "unclosed DOCTYPE",
			xml:     `<!DOCTYPE root SYSTEM "root.dtd"<root></root>`,
			wantErr: true,
		},
		{
			name:    "unclosed DOCTYPE internal subset",
			xml:     `<!DOCTYPE root [<!ELEMENT root (#PCDATA)><root></root>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := Valid(tt.xml)
			if tt.wantErr && valid {
				t.Error("Expected validation error but got valid=true")
			}
			if !tt.wantErr && !valid {
				t.Error("Expected valid XML but got valid=false")
			}
		})
	}
}

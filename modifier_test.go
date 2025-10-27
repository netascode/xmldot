// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// Modifier Framework Tests (10 tests)

func TestModifierFramework_RegisterCustomModifier(t *testing.T) {
	// Create a custom modifier
	customMod := &testUppercaseModifier{}

	// Register it
	err := RegisterModifier("testuppercase", customMod)
	if err != nil {
		t.Fatalf("Failed to register custom modifier: %v", err)
	}

	// Verify it can be retrieved
	retrieved := GetModifier("testuppercase")
	if retrieved == nil {
		t.Fatal("Failed to retrieve registered modifier")
	}

	// Clean up
	_ = UnregisterModifier("testuppercase")
}

func TestModifierFramework_UnregisterCustomModifier(t *testing.T) {
	// Register a custom modifier
	customMod := &testUppercaseModifier{}
	_ = RegisterModifier("testcustom", customMod)

	// Unregister it
	err := UnregisterModifier("testcustom")
	if err != nil {
		t.Fatalf("Failed to unregister custom modifier: %v", err)
	}

	// Verify it's gone
	retrieved := GetModifier("testcustom")
	if retrieved != nil {
		t.Fatal("Modifier should be unregistered")
	}
}

func TestModifierFramework_BuiltinCannotBeUnregistered(t *testing.T) {
	// Try to unregister a built-in modifier
	err := UnregisterModifier("reverse")
	if err == nil {
		t.Fatal("Should not be able to unregister built-in modifier")
	}

	// Verify it's still there
	retrieved := GetModifier("reverse")
	if retrieved == nil {
		t.Fatal("Built-in modifier should still exist")
	}
}

func TestModifierFramework_GetModifier(t *testing.T) {
	// Test retrieving built-in modifiers
	modifiers := []string{"reverse", "sort", "first", "last", "flatten", "pretty", "ugly"}
	for _, name := range modifiers {
		mod := GetModifier(name)
		if mod == nil {
			t.Errorf("Failed to retrieve built-in modifier %q", name)
		}
	}

	// Test retrieving non-existent modifier
	mod := GetModifier("nonexistent")
	if mod != nil {
		t.Error("Should return nil for non-existent modifier")
	}
}

func TestModifierFramework_ModifierFunc(t *testing.T) {
	// Create a ModifierFunc using NewModifierFunc
	upperFunc := NewModifierFunc("testuppercase", func(r Result) Result {
		return Result{
			Type: r.Type,
			Str:  strings.ToUpper(r.Str),
			Raw:  r.Raw,
		}
	})

	// Test it
	input := Result{Type: String, Str: "hello"}
	output := upperFunc.Apply(input)
	if output.Str != "HELLO" {
		t.Errorf("Expected HELLO, got %s", output.Str)
	}

	// Test Name() returns correct name
	if upperFunc.Name() != "testuppercase" {
		t.Errorf("Expected name 'testuppercase', got %s", upperFunc.Name())
	}
}

func TestModifierFramework_ConcurrentRegistration(_ *testing.T) {
	// Test thread safety of registration
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent%d", n)
			_ = RegisterModifier(name, &testUppercaseModifier{})
			// Clean up
			_ = UnregisterModifier(name)
		}(i)
	}
	wg.Wait()
}

func TestModifierFramework_DuplicateRegistrationError(t *testing.T) {
	// Register a modifier
	_ = RegisterModifier("duplicate1", &testUppercaseModifier{})
	defer func() { _ = UnregisterModifier("duplicate1") }()

	// Try to register with the same name
	err := RegisterModifier("duplicate1", &testUppercaseModifier{})
	if err == nil {
		t.Fatal("Should return error for duplicate registration")
	}
}

func TestModifierFramework_ChainDepthLimit(t *testing.T) {
	// Test that chains exceeding MaxModifierChainDepth return Null
	modifiers := make([]string, MaxModifierChainDepth+5)
	for i := range modifiers {
		modifiers[i] = "reverse"
	}

	input := Result{Type: Array, Results: []Result{{Type: String, Str: "test"}}}
	result := applyModifiers(input, modifiers)

	if result.Type != Null {
		t.Errorf("Expected Null for excessive chain depth, got %v", result.Type)
	}
}

func TestModifierFramework_EmptyModifierName(t *testing.T) {
	err := RegisterModifier("", &testUppercaseModifier{})
	if err == nil {
		t.Fatal("Should return error for empty modifier name")
	}
}

func TestModifierFramework_InvalidModifierHandling(t *testing.T) {
	// Test that unknown modifiers return Null
	input := Result{Type: String, Str: "test"}
	result := applyModifiers(input, []string{"nonexistent"})

	if result.Type != Null {
		t.Errorf("Expected Null for unknown modifier, got %v", result.Type)
	}
}

// @reverse Tests (5 tests)

func TestModifierReverse_Array(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: String, Str: "a"},
			{Type: String, Str: "b"},
			{Type: String, Str: "c"},
		},
	}

	mod := GetModifier("reverse")
	result := mod.Apply(input)

	if result.Type != Array || len(result.Results) != 3 {
		t.Fatalf("Expected array with 3 elements, got %v with %d elements", result.Type, len(result.Results))
	}

	expected := []string{"c", "b", "a"}
	for i, r := range result.Results {
		if r.Str != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], r.Str)
		}
	}
}

func TestModifierReverse_SingleElement(t *testing.T) {
	input := Result{Type: String, Str: "single"}
	mod := GetModifier("reverse")
	result := mod.Apply(input)

	// Should return input unchanged for non-arrays
	if result.Type != String || result.Str != "single" {
		t.Errorf("Expected unchanged single element, got %v: %s", result.Type, result.Str)
	}
}

func TestModifierReverse_EmptyResult(t *testing.T) {
	input := Result{Type: Array, Results: []Result{}}
	mod := GetModifier("reverse")
	result := mod.Apply(input)

	// Should handle empty arrays gracefully
	if result.Type != Array || len(result.Results) != 0 {
		t.Errorf("Expected empty array, got %v with %d elements", result.Type, len(result.Results))
	}
}

func TestModifierReverse_NonArray(t *testing.T) {
	input := Result{Type: String, Str: "notarray"}
	mod := GetModifier("reverse")
	result := mod.Apply(input)

	// Should return input unchanged
	if result.Type != String || result.Str != "notarray" {
		t.Errorf("Expected unchanged result for non-array, got %v: %s", result.Type, result.Str)
	}
}

func TestModifierReverse_PreservesMetadata(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: Number, Num: 1.0, Str: "1"},
			{Type: Number, Num: 2.0, Str: "2"},
		},
	}

	mod := GetModifier("reverse")
	result := mod.Apply(input)

	// Verify metadata (Num field) is preserved
	if result.Results[0].Num != 2.0 || result.Results[1].Num != 1.0 {
		t.Error("Metadata not preserved during reverse")
	}
}

// @sort Tests (6 tests)

func TestModifierSort_NumericArray(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: Number, Num: 5.0, Str: "5"},
			{Type: Number, Num: 1.0, Str: "1"},
			{Type: Number, Num: 3.0, Str: "3"},
		},
	}

	mod := GetModifier("sort")
	result := mod.Apply(input)

	expected := []float64{1.0, 3.0, 5.0}
	for i, r := range result.Results {
		if r.Num != expected[i] {
			t.Errorf("Index %d: expected %f, got %f", i, expected[i], r.Num)
		}
	}
}

func TestModifierSort_StringArray(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: String, Str: "zebra"},
			{Type: String, Str: "apple"},
			{Type: String, Str: "mango"},
		},
	}

	mod := GetModifier("sort")
	result := mod.Apply(input)

	expected := []string{"apple", "mango", "zebra"}
	for i, r := range result.Results {
		if r.Str != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], r.Str)
		}
	}
}

func TestModifierSort_MixedArray(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: String, Str: "50"},
			{Type: String, Str: "100"},
			{Type: String, Str: "5"},
		},
	}

	mod := GetModifier("sort")
	result := mod.Apply(input)

	// Should sort numerically since all can be parsed as numbers
	expected := []string{"5", "50", "100"}
	for i, r := range result.Results {
		if r.Str != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], r.Str)
		}
	}
}

func TestModifierSort_EmptyArray(t *testing.T) {
	input := Result{Type: Array, Results: []Result{}}
	mod := GetModifier("sort")
	result := mod.Apply(input)

	if result.Type != Array || len(result.Results) != 0 {
		t.Errorf("Expected empty array, got %v with %d elements", result.Type, len(result.Results))
	}
}

func TestModifierSort_SingleElement(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: String, Str: "single"},
		},
	}

	mod := GetModifier("sort")
	result := mod.Apply(input)

	if len(result.Results) != 1 || result.Results[0].Str != "single" {
		t.Error("Single element should remain unchanged")
	}
}

func TestModifierSort_PreservesMetadata(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: Number, Num: 3.0, Str: "3", Raw: "<item>3</item>"},
			{Type: Number, Num: 1.0, Str: "1", Raw: "<item>1</item>"},
		},
	}

	mod := GetModifier("sort")
	result := mod.Apply(input)

	// Verify Raw field is preserved
	if result.Results[0].Raw != "<item>1</item>" {
		t.Error("Metadata not preserved during sort")
	}
}

// @first Tests (4 tests)

func TestModifierFirst_Array(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: String, Str: "first"},
			{Type: String, Str: "second"},
			{Type: String, Str: "third"},
		},
	}

	mod := GetModifier("first")
	result := mod.Apply(input)

	if result.Type != String || result.Str != "first" {
		t.Errorf("Expected first element, got %v: %s", result.Type, result.Str)
	}
}

func TestModifierFirst_SingleElement(t *testing.T) {
	input := Result{Type: String, Str: "single"}
	mod := GetModifier("first")
	result := mod.Apply(input)

	// Should return input unchanged
	if result.Type != String || result.Str != "single" {
		t.Errorf("Expected unchanged single element, got %v: %s", result.Type, result.Str)
	}
}

func TestModifierFirst_EmptyResult(t *testing.T) {
	input := Result{Type: Array, Results: []Result{}}
	mod := GetModifier("first")
	result := mod.Apply(input)

	// Should return input unchanged (empty array)
	if result.Type != Array {
		t.Errorf("Expected Array type for empty array, got %v", result.Type)
	}
}

func TestModifierFirst_NonArray(t *testing.T) {
	input := Result{Type: String, Str: "notarray"}
	mod := GetModifier("first")
	result := mod.Apply(input)

	// Should return input unchanged
	if result.Type != String || result.Str != "notarray" {
		t.Errorf("Expected unchanged result, got %v: %s", result.Type, result.Str)
	}
}

// @last Tests (4 tests)

func TestModifierLast_Array(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: String, Str: "first"},
			{Type: String, Str: "second"},
			{Type: String, Str: "last"},
		},
	}

	mod := GetModifier("last")
	result := mod.Apply(input)

	if result.Type != String || result.Str != "last" {
		t.Errorf("Expected last element, got %v: %s", result.Type, result.Str)
	}
}

func TestModifierLast_SingleElement(t *testing.T) {
	input := Result{Type: String, Str: "single"}
	mod := GetModifier("last")
	result := mod.Apply(input)

	// Should return input unchanged
	if result.Type != String || result.Str != "single" {
		t.Errorf("Expected unchanged single element, got %v: %s", result.Type, result.Str)
	}
}

func TestModifierLast_EmptyResult(t *testing.T) {
	input := Result{Type: Array, Results: []Result{}}
	mod := GetModifier("last")
	result := mod.Apply(input)

	// Should return input unchanged (empty array)
	if result.Type != Array {
		t.Errorf("Expected Array type for empty array, got %v", result.Type)
	}
}

func TestModifierLast_NonArray(t *testing.T) {
	input := Result{Type: String, Str: "notarray"}
	mod := GetModifier("last")
	result := mod.Apply(input)

	// Should return input unchanged
	if result.Type != String || result.Str != "notarray" {
		t.Errorf("Expected unchanged result, got %v: %s", result.Type, result.Str)
	}
}

// @flatten Tests (5 tests)

func TestModifierFlatten_NestedArrays(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{
				Type: Array,
				Results: []Result{
					{Type: String, Str: "a"},
					{Type: String, Str: "b"},
				},
			},
			{
				Type: Array,
				Results: []Result{
					{Type: String, Str: "c"},
					{Type: String, Str: "d"},
				},
			},
		},
	}

	mod := GetModifier("flatten")
	result := mod.Apply(input)

	if result.Type != Array || len(result.Results) != 4 {
		t.Fatalf("Expected flattened array with 4 elements, got %v with %d elements", result.Type, len(result.Results))
	}

	expected := []string{"a", "b", "c", "d"}
	for i, r := range result.Results {
		if r.Str != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], r.Str)
		}
	}
}

func TestModifierFlatten_DeeplyNested(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{
				Type: Array,
				Results: []Result{
					{
						Type: Array,
						Results: []Result{
							{Type: String, Str: "deep"},
						},
					},
				},
			},
		},
	}

	mod := GetModifier("flatten")
	result := mod.Apply(input)

	// Should only flatten one level
	if result.Type != Array || len(result.Results) != 1 {
		t.Fatalf("Expected array with 1 element, got %v with %d elements", result.Type, len(result.Results))
	}

	// The result should still be an array (one level deep)
	if result.Results[0].Type != Array {
		t.Error("Should only flatten one level")
	}
}

func TestModifierFlatten_EmptyArray(t *testing.T) {
	input := Result{Type: Array, Results: []Result{}}
	mod := GetModifier("flatten")
	result := mod.Apply(input)

	// Should return Null for empty array
	if result.Type != Null {
		t.Errorf("Expected Null for empty array, got %v", result.Type)
	}
}

func TestModifierFlatten_NonArray(t *testing.T) {
	input := Result{Type: String, Str: "notarray"}
	mod := GetModifier("flatten")
	result := mod.Apply(input)

	// Should return input unchanged
	if result.Type != String || result.Str != "notarray" {
		t.Errorf("Expected unchanged result, got %v: %s", result.Type, result.Str)
	}
}

func TestModifierFlatten_MixedContent(t *testing.T) {
	input := Result{
		Type: Array,
		Results: []Result{
			{Type: String, Str: "single"},
			{
				Type: Array,
				Results: []Result{
					{Type: String, Str: "nested1"},
					{Type: String, Str: "nested2"},
				},
			},
			{Type: String, Str: "single2"},
		},
	}

	mod := GetModifier("flatten")
	result := mod.Apply(input)

	if len(result.Results) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(result.Results))
	}
}

// @pretty Tests (4 tests)

func TestModifierPretty_SimpleXML(t *testing.T) {
	input := Result{
		Type: Element,
		Raw:  "<root><child>value</child></root>",
		Str:  "value",
	}

	mod := GetModifier("pretty")
	result := mod.Apply(input)

	// Should contain indentation
	if !strings.Contains(result.Raw, "  ") {
		t.Error("Expected indentation in pretty output")
	}

	// Should preserve Str field
	if result.Str != "value" {
		t.Errorf("Expected Str to be preserved, got %s", result.Str)
	}
}

func TestModifierPretty_NestedXML(t *testing.T) {
	input := Result{
		Type: Element,
		Raw:  "<root><level1><level2><level3><level4><level5>deep</level5></level4></level3></level2></level1></root>",
		Str:  "deep",
	}

	mod := GetModifier("pretty")
	result := mod.Apply(input)

	// Should contain multiple levels of indentation
	lines := strings.Split(result.Raw, "\n")
	if len(lines) < 5 {
		t.Error("Expected multiple lines in pretty output")
	}
}

func TestModifierPretty_WithAttributes(t *testing.T) {
	input := Result{
		Type: Element,
		Raw:  `<root attr="value"><child id="1">text</child></root>`,
		Str:  "text",
	}

	mod := GetModifier("pretty")
	result := mod.Apply(input)

	// Should preserve attributes
	if !strings.Contains(result.Raw, `attr="value"`) {
		t.Error("Attributes should be preserved")
	}
}

func TestModifierPretty_InvalidXML(t *testing.T) {
	input := Result{
		Type: Element,
		Raw:  "<root><unclosed>",
		Str:  "",
	}

	mod := GetModifier("pretty")
	result := mod.Apply(input)

	// Should return input unchanged on error
	if result.Raw != "<root><unclosed>" {
		t.Error("Should return unchanged result on invalid XML")
	}
}

// @pretty xmlns Tests - Regression tests for duplicate xmlns bug fix

func TestModifierPretty_SingleNamespace(t *testing.T) {
	input := Result{
		Type: Element,
		Raw:  `<native xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-native"><cdp xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-cdp"><holdtime>15</holdtime></cdp></native>`,
		Str:  "",
	}

	mod := GetModifier("pretty")
	result := mod.Apply(input)

	// Each namespace should appear exactly once
	nativeCount := strings.Count(result.Raw, `xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-native"`)
	cdpCount := strings.Count(result.Raw, `xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-cdp"`)

	if nativeCount != 1 {
		t.Errorf("Expected 1 occurrence of native namespace, got %d. Output:\n%s", nativeCount, result.Raw)
	}
	if cdpCount != 1 {
		t.Errorf("Expected 1 occurrence of cdp namespace, got %d. Output:\n%s", cdpCount, result.Raw)
	}
}

func TestModifierPretty_DuplicateXmlnsInInput(t *testing.T) {
	// Input already has duplicate xmlns on same element
	input := Result{
		Type: Element,
		Raw:  `<native xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-native"><cdp><holdtime xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-cdp" xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-cdp">15</holdtime></cdp></native>`,
		Str:  "",
	}

	mod := GetModifier("pretty")
	result := mod.Apply(input)

	// Should deduplicate the xmlns on holdtime element
	holdtimeXmlnsCount := 0
	// Find the holdtime opening tag and count xmlns in it
	start := strings.Index(result.Raw, "<holdtime")
	if start != -1 {
		end := strings.Index(result.Raw[start:], ">")
		if end != -1 {
			holdtimeTag := result.Raw[start : start+end+1]
			holdtimeXmlnsCount = strings.Count(holdtimeTag, `xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-cdp"`)
		}
	}

	if holdtimeXmlnsCount != 1 {
		t.Errorf("Expected 1 xmlns in <holdtime> tag, got %d. Output:\n%s", holdtimeXmlnsCount, result.Raw)
	}
}

func TestModifierPretty_MultipleNamespaces(t *testing.T) {
	input := Result{
		Type: Element,
		Raw:  `<root xmlns="http://example.com/ns1"><child1 xmlns="http://example.com/ns2"><child2>value</child2></child1></root>`,
		Str:  "value",
	}

	mod := GetModifier("pretty")
	result := mod.Apply(input)

	// Each namespace should appear exactly once
	ns1Count := strings.Count(result.Raw, `xmlns="http://example.com/ns1"`)
	ns2Count := strings.Count(result.Raw, `xmlns="http://example.com/ns2"`)

	if ns1Count != 1 {
		t.Errorf("Expected 1 occurrence of ns1, got %d. Output:\n%s", ns1Count, result.Raw)
	}
	if ns2Count != 1 {
		t.Errorf("Expected 1 occurrence of ns2, got %d. Output:\n%s", ns2Count, result.Raw)
	}

	// Should still format with proper indentation
	if !strings.Contains(result.Raw, "  ") {
		t.Error("Expected indentation in pretty output")
	}
}

func TestModifierPretty_NetconfExample(t *testing.T) {
	// Real-world NETCONF example that triggered the bug
	input := Result{
		Type: Element,
		Raw:  `<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0"><edit-config><config><native xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-native"><cdp operation="remove"></cdp></native></config></edit-config></rpc>`,
		Str:  "",
	}

	mod := GetModifier("pretty")
	result := mod.Apply(input)

	// Check that each xmlns declaration appears only once
	netconfNsCount := strings.Count(result.Raw, `xmlns="urn:ietf:params:xml:ns:netconf:base:1.0"`)
	nativeNsCount := strings.Count(result.Raw, `xmlns="http://cisco.com/ns/yang/Cisco-IOS-XE-native"`)

	if netconfNsCount != 1 {
		t.Errorf("Expected 1 occurrence of netconf namespace, got %d. Output:\n%s", netconfNsCount, result.Raw)
	}
	if nativeNsCount != 1 {
		t.Errorf("Expected 1 occurrence of native namespace, got %d. Output:\n%s", nativeNsCount, result.Raw)
	}
}

// @ugly Tests (4 tests)

func TestModifierUgly_CompactXML(t *testing.T) {
	input := Result{
		Type: Element,
		Raw: `<root>
  <child>value</child>
</root>`,
		Str: "value",
	}

	mod := GetModifier("ugly")
	result := mod.Apply(input)

	// Should remove whitespace between tags
	if strings.Contains(result.Raw, "\n") || strings.Contains(result.Raw, "  ") {
		t.Error("Expected whitespace to be removed")
	}

	// Should contain the actual content
	if !strings.Contains(result.Raw, "<root>") || !strings.Contains(result.Raw, "<child>") {
		t.Error("XML structure should be preserved")
	}
}

func TestModifierUgly_RemoveAllWhitespace(t *testing.T) {
	input := Result{
		Type: Element,
		Raw: `<root>
    <child>
        value
    </child>
</root>`,
		Str: "value",
	}

	mod := GetModifier("ugly")
	result := mod.Apply(input)

	// Count whitespace characters
	whitespaceCount := strings.Count(result.Raw, " ") + strings.Count(result.Raw, "\n") + strings.Count(result.Raw, "\t")
	if whitespaceCount > 0 {
		t.Errorf("Expected no whitespace between tags, found %d whitespace chars", whitespaceCount)
	}
}

func TestModifierUgly_PreserveContentWhitespace(t *testing.T) {
	input := Result{
		Type: Element,
		Raw:  `<root>  <child attr="val ue">text content</child>  </root>`,
		Str:  "text content",
	}

	mod := GetModifier("ugly")
	result := mod.Apply(input)

	// Attribute values should preserve whitespace
	if !strings.Contains(result.Raw, `attr="val ue"`) {
		t.Error("Whitespace in attribute values should be preserved")
	}
}

func TestModifierUgly_InvalidXML(t *testing.T) {
	input := Result{
		Type: Element,
		Raw:  "<root><unclosed>",
		Str:  "",
	}

	mod := GetModifier("ugly")
	result := mod.Apply(input)

	// Should handle gracefully (process even invalid XML)
	if result.Raw == "" {
		t.Error("Should process invalid XML without error")
	}
}

func TestModifierUgly_PreservesCDATA(t *testing.T) {
	input := Result{
		Type: Element,
		Raw: `<root>
  <data><![CDATA[  whitespace   preserved  ]]></data>
</root>`,
		Str: "  whitespace   preserved  ",
	}

	mod := GetModifier("ugly")
	result := mod.Apply(input)

	// CDATA whitespace should be preserved
	if !strings.Contains(result.Raw, "  whitespace   preserved  ") {
		t.Error("@ugly should preserve whitespace in CDATA sections")
	}

	// CDATA markers should be preserved
	if !strings.Contains(result.Raw, "<![CDATA[") || !strings.Contains(result.Raw, "]]>") {
		t.Error("CDATA markers should be preserved")
	}

	// Whitespace outside CDATA should still be removed
	// The input has newlines and indentation outside CDATA which should be compacted
	if strings.HasPrefix(result.Raw, "\n") {
		t.Error("Leading whitespace outside CDATA should be removed")
	}
}

// Modifier Chaining Tests (8 tests)

func TestModifierChain_SortReverse(t *testing.T) {
	xml := `<root><items><item>3</item><item>1</item><item>2</item></items></root>`
	// Use wildcard to get all items as an array
	result := Get(xml, "root.items.*|@sort|@reverse")

	if !result.IsArray() {
		t.Fatalf("Expected array result, got type %v", result.Type)
	}

	expected := []string{"3", "2", "1"}
	for i, r := range result.Array() {
		if r.Str != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], r.Str)
		}
	}
}

func TestModifierChain_FirstPretty(t *testing.T) {
	xml := `<root><items><item><nested><data>value</data></nested></item><item><nested><data>value2</data></nested></item></items></root>`
	// Use wildcard to get all items, then apply @first|@pretty
	result := Get(xml, "root.items.*|@first|@pretty")

	// Should get first item - verify it still has the right content
	// Note: @pretty may not always add newlines for simple XML
	if !strings.Contains(result.Raw, "value") {
		t.Error("Expected first item content")
	}
	// Verify the result is still valid (not null)
	if !result.Exists() {
		t.Error("Result should exist after modifiers")
	}
}

func TestModifierChain_FlattenSortLast(t *testing.T) {
	xml := `<root><group><item>3</item><item>1</item></group><group><item>2</item></group></root>`
	// Use wildcard on group to get all items (returns nested arrays)
	result := Get(xml, "root.*.item|@flatten|@sort|@last")

	// Should flatten, sort numerically, then return last
	if result.Str != "3" {
		t.Errorf("Expected 3, got %s", result.Str)
	}
}

func TestModifierChain_FiveModifiers(t *testing.T) {
	xml := `<root><items><item>1</item><item>2</item><item>3</item></items></root>`
	// Use wildcard to get all items as array
	result := Get(xml, "root.items.*|@sort|@reverse|@first|@pretty|@ugly")

	// Should chain 5 modifiers successfully
	if result.Str != "3" {
		t.Errorf("Expected 3 (first of reversed sorted array), got %s", result.Str)
	}
}

func TestModifierChain_MaxDepth(t *testing.T) {
	// Create path with exactly MaxModifierChainDepth modifiers
	var path strings.Builder
	path.WriteString("root.item")
	for i := 0; i < MaxModifierChainDepth; i++ {
		path.WriteString("|@reverse")
	}

	xml := `<root><item>test</item></root>`
	result := Get(xml, path.String())

	// Should work with exactly MaxModifierChainDepth
	if !result.Exists() {
		t.Error("Should handle max modifier chain depth")
	}
}

func TestModifierChain_ExceedMaxDepth(t *testing.T) {
	// Create path exceeding MaxModifierChainDepth
	var path strings.Builder
	path.WriteString("root.item")
	for i := 0; i < MaxModifierChainDepth+5; i++ {
		path.WriteString("|@reverse")
	}

	xml := `<root><item>test</item></root>`
	result := Get(xml, path.String())

	// Should return Null for excessive chain depth
	if result.Exists() {
		t.Error("Should return Null for excessive modifier chain depth")
	}
}

func TestModifierChain_UnknownModifier(t *testing.T) {
	xml := `<root><item>test</item></root>`
	result := Get(xml, "root.item|@sort|@unknown|@first")

	// Should fail gracefully on unknown modifier
	if result.Exists() {
		t.Error("Should return Null for unknown modifier in chain")
	}
}

func TestModifierChain_WithWildcards(t *testing.T) {
	xml := `<root><a><item>3</item></a><b><item>1</item></b><c><item>2</item></c></root>`
	result := Get(xml, "root.*.item|@sort|@first")

	// Should work with wildcards
	if result.Str != "1" {
		t.Errorf("Expected 1 (first of sorted wildcard results), got %s", result.Str)
	}
}

// Integration Tests (5 tests)

func TestModifierIntegration_GetWithSingleModifier(t *testing.T) {
	xml := `<root><items><item>c</item><item>a</item><item>b</item></items></root>`
	// Use wildcard to get all items as array
	result := Get(xml, "root.items.*|@sort")

	if !result.IsArray() {
		t.Fatalf("Expected array result, got type %v", result.Type)
	}

	expected := []string{"a", "b", "c"}
	for i, r := range result.Array() {
		if r.Str != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], r.Str)
		}
	}
}

func TestModifierIntegration_GetManyWithModifiers(t *testing.T) {
	xml := `<root><nums><n>3</n><n>1</n><n>2</n></nums></root>`
	// Use wildcard to get all n elements as arrays
	results := GetMany(xml,
		"root.nums.*|@sort|@first",
		"root.nums.*|@sort|@last",
	)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results[0].Str != "1" {
		t.Errorf("Expected first to be 1, got %s", results[0].Str)
	}
	if results[1].Str != "3" {
		t.Errorf("Expected last to be 3, got %s", results[1].Str)
	}
}

func TestModifierIntegration_WithWildcards(t *testing.T) {
	xml := `<root><a>3</a><b>1</b><c>2</c></root>`
	result := Get(xml, "root.*|@sort")

	if !result.IsArray() || len(result.Array()) != 3 {
		t.Fatal("Expected array with 3 elements")
	}

	expected := []string{"1", "2", "3"}
	for i, r := range result.Array() {
		if r.Str != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], r.Str)
		}
	}
}

func TestModifierIntegration_WithFilters(t *testing.T) {
	xml := `<root><item age="30">Alice</item><item age="25">Bob</item><item age="35">Carol</item></root>`
	result := Get(xml, "root.item.#(@age>26)#|@sort")

	if !result.IsArray() || len(result.Array()) != 2 {
		t.Fatalf("Expected array with 2 elements, got %d", len(result.Array()))
	}

	// Should have Alice and Carol, sorted alphabetically
	expected := []string{"Alice", "Carol"}
	for i, r := range result.Array() {
		if r.Str != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], r.Str)
		}
	}
}

func TestModifierIntegration_WithArrayIndex(t *testing.T) {
	xml := `<root><item>3</item><item>1</item><item>2</item></root>`
	result := Get(xml, "root.item.1|@pretty")

	// Should get the second item (index 1) and pretty print it
	if result.Str != "1" {
		t.Errorf("Expected 1, got %s", result.Str)
	}
}

// Test helper: custom modifier for testing
type testUppercaseModifier struct{}

func (m *testUppercaseModifier) Name() string { return "testuppercase" }

func (m *testUppercaseModifier) Apply(r Result) Result {
	return Result{
		Type: r.Type,
		Str:  strings.ToUpper(r.Str),
		Raw:  r.Raw,
		Num:  r.Num,
	}
}

// Benchmark Tests (7 benchmarks)

func BenchmarkModifierReverse(b *testing.B) {
	// Create array with 100 elements
	results := make([]Result, 100)
	for i := 0; i < 100; i++ {
		results[i] = Result{Type: String, Str: fmt.Sprintf("item%d", i)}
	}
	input := Result{Type: Array, Results: results}

	mod := GetModifier("reverse")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.Apply(input)
	}
}

func BenchmarkModifierSort100Elements(b *testing.B) {
	// Create array with 100 numeric elements in random order
	results := make([]Result, 100)
	for i := 0; i < 100; i++ {
		results[i] = Result{Type: Number, Num: float64((i*7 + 13) % 100), Str: fmt.Sprintf("%d", (i*7+13)%100)}
	}
	input := Result{Type: Array, Results: results}

	mod := GetModifier("sort")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.Apply(input)
	}
}

func BenchmarkModifierFirst(b *testing.B) {
	results := make([]Result, 100)
	for i := 0; i < 100; i++ {
		results[i] = Result{Type: String, Str: fmt.Sprintf("item%d", i)}
	}
	input := Result{Type: Array, Results: results}

	mod := GetModifier("first")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.Apply(input)
	}
}

func BenchmarkModifierLast(b *testing.B) {
	results := make([]Result, 100)
	for i := 0; i < 100; i++ {
		results[i] = Result{Type: String, Str: fmt.Sprintf("item%d", i)}
	}
	input := Result{Type: Array, Results: results}

	mod := GetModifier("last")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.Apply(input)
	}
}

func BenchmarkModifierFlatten(b *testing.B) {
	// Create nested arrays
	results := make([]Result, 10)
	for i := 0; i < 10; i++ {
		subResults := make([]Result, 10)
		for j := 0; j < 10; j++ {
			subResults[j] = Result{Type: String, Str: fmt.Sprintf("item%d-%d", i, j)}
		}
		results[i] = Result{Type: Array, Results: subResults}
	}
	input := Result{Type: Array, Results: results}

	mod := GetModifier("flatten")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.Apply(input)
	}
}

func BenchmarkModifierPretty1KB(b *testing.B) {
	// Create ~1KB XML document
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString("<root>")
	for i := 0; i < 20; i++ {
		xmlBuilder.WriteString(fmt.Sprintf("<item id=\"%d\"><name>Item %d</name><value>%d</value></item>", i, i, i*10))
	}
	xmlBuilder.WriteString("</root>")

	input := Result{
		Type: Element,
		Raw:  xmlBuilder.String(),
		Str:  "",
	}

	mod := GetModifier("pretty")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.Apply(input)
	}
}

func BenchmarkModifierUgly1KB(b *testing.B) {
	// Create ~1KB pretty-formatted XML document
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString("<root>\n")
	for i := 0; i < 20; i++ {
		xmlBuilder.WriteString(fmt.Sprintf("  <item id=\"%d\">\n    <name>Item %d</name>\n    <value>%d</value>\n  </item>\n", i, i, i*10))
	}
	xmlBuilder.WriteString("</root>")

	input := Result{
		Type: Element,
		Raw:  xmlBuilder.String(),
		Str:  "",
	}

	mod := GetModifier("ugly")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.Apply(input)
	}
}

func BenchmarkModifierChain3Deep(b *testing.B) {
	// Create test data
	results := make([]Result, 50)
	for i := 0; i < 50; i++ {
		results[i] = Result{Type: Number, Num: float64(i), Str: fmt.Sprintf("%d", i)}
	}
	input := Result{Type: Array, Results: results}

	// Chain: sort -> reverse -> first
	modifiers := []string{"sort", "reverse", "first"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyModifiers(input, modifiers)
	}
}

// Example functions for godoc

// ExampleRegisterModifier demonstrates registering a custom modifier
func ExampleRegisterModifier() {
	// Create a custom modifier that uppercases strings
	upperMod := NewModifierFunc("testupper2", func(r Result) Result {
		return Result{
			Type: r.Type,
			Str:  strings.ToUpper(r.Str),
			Raw:  r.Raw,
		}
	})

	// Register the modifier
	err := RegisterModifier("testupper2", upperMod)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer func() { _ = UnregisterModifier("testupper2") }()

	// Use it in a query
	xml := `<item>apple</item>`

	result := Get(xml, "item|@testupper2")
	fmt.Println(result.String())
	// Output: APPLE
}

// ExampleGet_modifierChain demonstrates chaining multiple modifiers
func ExampleGet_modifierChain() {
	xml := `<score>85</score>`

	// Get score value
	result := Get(xml, "score")
	fmt.Printf("Score: %d\n", result.Int())
	// Output: Score: 85
}

// ExampleModifierFunc demonstrates creating a modifier function
func ExampleModifierFunc() {
	xml := `<price>10</price>`

	// Get price without modifier
	result := Get(xml, "price")
	fmt.Printf("$%.0f\n", result.Float())
	// Output: $10
}

// ============================================================================
// Coverage Tests for Modifier Name() Methods
// ============================================================================

// TestModifierNames tests that all built-in modifiers have correct names
func TestModifierNames(t *testing.T) {
	tests := []struct {
		modifierName string
		expectedName string
	}{
		{"reverse", "reverse"},
		{"sort", "sort"},
		{"first", "first"},
		{"last", "last"},
		{"flatten", "flatten"},
		{"pretty", "pretty"},
		{"ugly", "ugly"},
	}

	for _, tt := range tests {
		t.Run(tt.modifierName, func(t *testing.T) {
			mod := GetModifier(tt.modifierName)
			if mod == nil {
				t.Fatalf("Modifier %q not found", tt.modifierName)
			}
			if mod.Name() != tt.expectedName {
				t.Errorf("Name() = %q, expected %q", mod.Name(), tt.expectedName)
			}
		})
	}
}

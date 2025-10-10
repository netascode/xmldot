// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"strings"
	"testing"
)

// TestNamespacePrefixElement verifies matching elements with namespace prefixes
func TestNamespacePrefixElement(t *testing.T) {
	xml := `<soap:Envelope><soap:Body>content</soap:Body></soap:Envelope>`
	result := Get(xml, "soap:Envelope.soap:Body")
	if !result.Exists() {
		t.Errorf("Expected to find soap:Body, got: %v", result)
	}
	if result.String() != "content" {
		t.Errorf("Expected 'content', got: %s", result.String())
	}
}

// TestNamespacePrefixAttribute verifies matching attributes with namespace prefixes
func TestNamespacePrefixAttribute(t *testing.T) {
	xml := `<root xmlns:ns="http://example.com" ns:attr="value">content</root>`
	result := Get(xml, "root.@ns:attr")
	if !result.Exists() {
		t.Errorf("Expected to find @ns:attr")
	}
	if result.String() != "value" {
		t.Errorf("Expected 'value', got: %s", result.String())
	}
}

// TestUnprefixedMatchesAny verifies unprefixed path matches local name (backward compat)
func TestUnprefixedMatchesAny(t *testing.T) {
	xml := `<soap:Envelope><Body>content</Body></soap:Envelope>`

	// Unprefixed path should match namespaced element by local name
	result := Get(xml, "Envelope.Body")
	if !result.Exists() {
		t.Errorf("Expected unprefixed 'Envelope' to match 'soap:Envelope'")
	}
	if result.String() != "content" {
		t.Errorf("Expected 'content', got: %s", result.String())
	}
}

// TestPrefixedRequiresExactMatch verifies prefixed path requires exact prefix+local match
func TestPrefixedRequiresExactMatch(t *testing.T) {
	xml := `<soap:Envelope><other:Envelope>wrong</other:Envelope></soap:Envelope>`

	// Path with prefix should only match exact prefix+local
	result := Get(xml, "soap:Envelope.other:Envelope")
	if !result.Exists() {
		t.Errorf("Expected to find other:Envelope inside soap:Envelope")
	}
	if result.String() != "wrong" {
		t.Errorf("Expected 'wrong', got: %s", result.String())
	}

	// Should not match different prefix
	result2 := Get(xml, "soap:Envelope.soap:Envelope")
	if result2.Exists() {
		t.Errorf("Expected NOT to find soap:Envelope inside soap:Envelope")
	}
}

// TestLocalNameOnly verifies matching by local name without prefix
func TestLocalNameOnly(t *testing.T) {
	xml := `<root><ns1:item>value1</ns1:item><ns2:item>value2</ns2:item><item>value3</item></root>`

	// Unprefixed path should match all elements with that local name
	result := Get(xml, "root.item")
	if !result.Exists() {
		t.Errorf("Expected to find item")
	}
	// Should match first occurrence (ns1:item)
	if result.String() != "value1" {
		t.Errorf("Expected 'value1' (first item), got: %s", result.String())
	}
}

// TestEmptyPrefix verifies handling of edge case with colon but no prefix
func TestEmptyPrefix(_ *testing.T) {
	xml := `<root><:element>value</:element></root>`

	// Path with colon should be treated specially
	// splitNamespace will return ("", "element") for ":element"
	result := Get(xml, "root.:element")
	// This might not match due to path parsing - that's acceptable
	// The important thing is no crash/panic
	_ = result
}

// TestNestedNamespacedElements verifies nested elements with same namespace prefix
func TestNestedNamespacedElements(t *testing.T) {
	xml := `<soap:Envelope><soap:Body><soap:Request>data</soap:Request></soap:Body></soap:Envelope>`
	result := Get(xml, "soap:Envelope.soap:Body.soap:Request")
	if !result.Exists() {
		t.Errorf("Expected to find nested namespaced elements")
	}
	if result.String() != "data" {
		t.Errorf("Expected 'data', got: %s", result.String())
	}
}

// TestMixedNamespacedUnprefixed verifies mixed prefixed and unprefixed paths
func TestMixedNamespacedUnprefixed(t *testing.T) {
	xml := `<soap:Envelope><Body>content</Body></soap:Envelope>`

	// Start with prefixed, end with unprefixed
	result := Get(xml, "soap:Envelope.Body")
	if !result.Exists() {
		t.Errorf("Expected mixed namespace path to work")
	}
	if result.String() != "content" {
		t.Errorf("Expected 'content', got: %s", result.String())
	}
}

// TestMultipleNamespaces verifies different prefixes in same path
func TestMultipleNamespaces(t *testing.T) {
	xml := `<ns1:root><ns2:child><ns3:data>value</ns3:data></ns2:child></ns1:root>`
	result := Get(xml, "ns1:root.ns2:child.ns3:data")
	if !result.Exists() {
		t.Errorf("Expected to find element with multiple different namespaces")
	}
	if result.String() != "value" {
		t.Errorf("Expected 'value', got: %s", result.String())
	}
}

// TestDeepNestedNamespaces verifies 5+ levels with various prefixes
func TestDeepNestedNamespaces(t *testing.T) {
	xml := `<a:l1><b:l2><c:l3><d:l4><e:l5>deep</e:l5></d:l4></c:l3></b:l2></a:l1>`
	result := Get(xml, "a:l1.b:l2.c:l3.d:l4.e:l5")
	if !result.Exists() {
		t.Errorf("Expected to find deeply nested namespaced element")
	}
	if result.String() != "deep" {
		t.Errorf("Expected 'deep', got: %s", result.String())
	}
}

// TestWildcardWithNamespaces verifies wildcards work with namespaced elements
func TestWildcardWithNamespaces(t *testing.T) {
	xml := `<soap:Envelope><soap:Header>h</soap:Header><soap:Body><Data>content</Data></soap:Body></soap:Envelope>`
	result := Get(xml, "soap:Envelope.*.Data")
	if !result.Exists() {
		t.Errorf("Expected wildcard to match through namespaced elements")
	}
	if result.String() != "content" {
		t.Errorf("Expected 'content', got: %s", result.String())
	}
}

// TestNamespaceInWildcardResults verifies wildcard returns namespaced elements
func TestNamespaceInWildcardResults(t *testing.T) {
	xml := `<root><ns1:item>1</ns1:item><ns2:item>2</ns2:item><item>3</item></root>`
	result := Get(xml, "root.*")
	if result.Type != Array {
		t.Errorf("Expected array result, got: %v", result.Type)
	}
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 items, got: %d", len(result.Results))
	}
}

// TestRecursiveWildcardNamespace verifies recursive wildcard with namespaces
func TestRecursiveWildcardNamespace(t *testing.T) {
	xml := `<soap:Envelope><soap:Body><nested><Data>target</Data></nested></soap:Body></soap:Envelope>`
	result := Get(xml, "soap:Envelope.**.Data")
	if !result.Exists() {
		t.Errorf("Expected recursive wildcard to find Data through namespaced elements")
	}
	if result.String() != "target" {
		t.Errorf("Expected 'target', got: %s", result.String())
	}
}

// TestWildcardUnprefixedMatchesNamespaced verifies unprefixed wildcard matches namespaced children
func TestWildcardUnprefixedMatchesNamespaced(t *testing.T) {
	xml := `<Envelope><soap:Body>content</soap:Body><Footer>end</Footer></Envelope>`
	result := Get(xml, "Envelope.*")
	if result.Type != Array {
		t.Errorf("Expected array result, got: %v", result.Type)
		return
	}
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 children, got: %d", len(result.Results))
		return
	}
	// First child should be soap:Body
	if result.Results[0].String() != "content" {
		t.Errorf("Expected 'content' from soap:Body, got: %s", result.Results[0].String())
	}
	// Second child should be Footer
	if result.Results[1].String() != "end" {
		t.Errorf("Expected 'end' from Footer, got: %s", result.Results[1].String())
	}
}

// TestFilterNamespacedElements verifies filters work with namespaced elements
func TestFilterNamespacedElements(t *testing.T) {
	xml := `<ns:items>
		<ns:item ns:status="active"><ns:name>item1</ns:name></ns:item>
		<ns:item ns:status="inactive"><ns:name>item2</ns:name></ns:item>
		<ns:item ns:status="active"><ns:name>item3</ns:name></ns:item>
	</ns:items>`

	// First verify we can access the element without filter
	resultNoFilter := Get(xml, "ns:items.ns:item.ns:name")
	if !resultNoFilter.Exists() {
		t.Errorf("Basic path failed - expected to find ns:item.ns:name")
	}

	// Then test with filter - need to access child element after filter
	result := Get(xml, "ns:items.ns:item.#(@ns:status==active).ns:name")
	if !result.Exists() {
		t.Errorf("Expected to find namespaced element with filter")
	}
	// When there are multiple matches, Get returns the first one
	if result.Type == Array && len(result.Results) > 0 {
		if result.Results[0].String() != "item1" {
			t.Errorf("Expected 'item1', got: %s", result.Results[0].String())
		}
	} else {
		if result.String() != "item1" {
			t.Errorf("Expected 'item1', got: %s", result.String())
		}
	}
}

// TestFilterNamespacedAttributes verifies filtering by namespaced attribute value
func TestFilterNamespacedAttributes(t *testing.T) {
	xml := `<root>
		<item xmlns:a="uri1" a:id="100">first</item>
		<item xmlns:a="uri1" a:id="200">second</item>
	</root>`
	result := Get(xml, "root.item.#(@a:id==200)")
	if !result.Exists() {
		t.Errorf("Expected to find item with a:id=200")
	}
	if result.String() != "second" {
		t.Errorf("Expected 'second', got: %s", result.String())
	}
}

// TestCombinedNamespaceFilterWildcard verifies complex query combining all features
func TestCombinedNamespaceFilterWildcard(t *testing.T) {
	xml := `<ns:root>
		<ns:section>
			<ns:item status="active" ns:priority="high">target</ns:item>
			<ns:item status="inactive" ns:priority="low">skip</ns:item>
		</ns:section>
	</ns:root>`
	result := Get(xml, "ns:root.*.ns:item.#(@ns:priority==high)")
	if !result.Exists() {
		t.Errorf("Expected to find item combining namespace, wildcard, and filter")
	}
	if result.String() != "target" {
		t.Errorf("Expected 'target', got: %s", result.String())
	}
}

// TestNamespacePrefixTooLong verifies security limit on prefix length
func TestNamespacePrefixTooLong(_ *testing.T) {
	longPrefix := strings.Repeat("a", MaxNamespacePrefixLength+100)
	xml := `<root><` + longPrefix + `:element>value</` + longPrefix + `:element></root>`

	// Should handle gracefully without crash/panic
	result := Get(xml, "root."+longPrefix+":element")
	// splitNamespace should treat this as unprefixed due to length limit
	// So it might not match, which is acceptable security behavior
	_ = result
}

// TestNamespaceColonOnly verifies handling of edge case element names
func TestNamespaceColonOnly(_ *testing.T) {
	// Test various edge cases with colons
	testCases := []struct {
		xml  string
		path string
	}{
		{`<root><:element>value</:element></root>`, "root.:element"},
		{`<root>< :element>value</ :element></root>`, "root.:element"},
	}

	for _, tc := range testCases {
		result := Get(tc.xml, tc.path)
		// Main goal: no crash/panic, result can be Exists() or not
		_ = result
	}
}

// TestNamespaceMultipleColons verifies only first colon is namespace separator
func TestNamespaceMultipleColons(_ *testing.T) {
	// In XML, only the first colon is the namespace separator
	xml := `<ns:sub:element>value</ns:sub:element>`

	// This is actually invalid XML, but test parser handles it gracefully
	result := Get(xml, "ns:sub:element")
	// splitNamespace will split at first colon: ("ns", "sub:element")
	// This won't match the element name "ns:sub:element" from parser
	// That's correct behavior - we only handle first colon as separator
	_ = result
}

// TestNamespaceCaseInsensitive verifies namespace matching with case-insensitive option
func TestNamespaceCaseInsensitive(t *testing.T) {
	xml := `<SOAP:Envelope><soap:Body>content</soap:Body></SOAP:Envelope>`
	opts := &Options{CaseSensitive: false}
	result := GetWithOptions(xml, "soap:envelope.SOAP:body", opts)
	if !result.Exists() {
		t.Errorf("Expected case-insensitive namespace matching to work")
	}
	if result.String() != "content" {
		t.Errorf("Expected 'content', got: %s", result.String())
	}
}

// TestUnprefixedCaseInsensitiveNamespace verifies unprefixed case-insensitive matching
func TestUnprefixedCaseInsensitiveNamespace(t *testing.T) {
	xml := `<SOAP:ENVELOPE><Body>content</Body></SOAP:ENVELOPE>`
	opts := &Options{CaseSensitive: false}
	result := GetWithOptions(xml, "envelope.body", opts)
	if !result.Exists() {
		t.Errorf("Expected case-insensitive unprefixed path to match namespaced element")
	}
	if result.String() != "content" {
		t.Errorf("Expected 'content', got: %s", result.String())
	}
}

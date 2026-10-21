// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"testing"
)

// TestArrayLimit_BeyondWildcardLimit verifies that non-recursive array operations
// on sibling lists larger than MaxWildcardResults return every element.
func TestArrayLimit_BeyondWildcardLimit(t *testing.T) {
	const n = 4094

	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&sb, "<vlan><id>%d</id><priority>4096</priority></vlan>", i)
	}
	sb.WriteString("</root>")
	xml := sb.String()

	if got := Get(xml, "root.vlan.#").Int(); got != n {
		t.Errorf("count: got %d, want %d", got, n)
	}
	if got := Get(xml, fmt.Sprintf("root.vlan.%d.id", n-1)).String(); got != fmt.Sprint(n) {
		t.Errorf("last index: got %q, want %q", got, fmt.Sprint(n))
	}
	if got := len(Get(xml, "root.vlan.#.id").Array()); got != n {
		t.Errorf("field extraction: got %d, want %d", got, n)
	}
	if got := len(Get(xml, "root.vlan.#(priority==4096)#").Array()); got != n {
		t.Errorf("filter: got %d, want %d", got, n)
	}
	if got := len(Get(xml, "root.*").Array()); got != n {
		t.Errorf("single-level wildcard: got %d, want %d", got, n)
	}

	count := 0
	Get(xml, "root.vlan").ForEach(func(_ int, _ Result) bool {
		count++
		return true
	})
	if count != 1 {
		t.Errorf("ForEach on non-array path: got %d, want 1", count)
	}
	count = 0
	Get(xml, "root.vlan.#.id").ForEach(func(_ int, _ Result) bool {
		count++
		return true
	})
	if count != n {
		t.Errorf("ForEach on array: got %d, want %d", count, n)
	}

	root := Get(xml, "root")
	if got := len(root.Map()["vlan"].Array()); got != n {
		t.Errorf("Map children: got %d, want %d", got, n)
	}

	// Multi-root fragment (e.g. Result.Raw of a parent element)
	fragment := root.Raw
	if got := Get(fragment, "vlan.#").Int(); got != n {
		t.Errorf("fragment roots count: got %d, want %d", got, n)
	}
	if got := Get(fragment, fmt.Sprintf("vlan.%d.id", n-1)).String(); got != fmt.Sprint(n) {
		t.Errorf("fragment roots last index: got %q, want %q", got, fmt.Sprint(n))
	}

	// Indexed Set past the old limit
	modified, err := Set(xml, fmt.Sprintf("root.vlan.%d.@xmlns", n-1), "urn:test")
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := Get(modified, fmt.Sprintf("root.vlan.%d.@xmlns", n-1)).String(); got != "urn:test" {
		t.Errorf("indexed Set: got %q, want %q", got, "urn:test")
	}
}

// TestArrayLimit_RecursiveWildcardLimit verifies that recursive wildcard queries
// fail once they exceed MaxWildcardResults.
func TestArrayLimit_RecursiveWildcardLimit(t *testing.T) {
	defer withLimit(&MaxWildcardResults, 1000)()

	var sb strings.Builder
	sb.WriteString("<root><data>")
	for i := 0; i < MaxWildcardResults+500; i++ {
		sb.WriteString("<items>x</items>")
	}
	sb.WriteString("</data></root>")

	result := Get(sb.String(), "root.**.items")
	if result.Exists() {
		t.Errorf("recursive wildcard: expected Null result, got %d results", len(result.Array()))
	}
	requireLimitError(t, result.Err, "MaxWildcardResults")
}

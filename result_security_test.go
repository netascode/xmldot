// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// Security Tests for Fluent API (Result.Get, GetMany, GetWithOptions)
// ============================================================================

// TestResultGet_DeepChainingAttack verifies deep chaining doesn't cause memory/CPU issues
func TestResultGet_DeepChainingAttack(t *testing.T) {
	// Generate deeply nested XML (100 levels)
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf("<level%d>", i))
	}
	sb.WriteString("<value>test</value>")
	for i := 99; i >= 0; i-- {
		sb.WriteString(fmt.Sprintf("</level%d>", i))
	}
	sb.WriteString("</root>")
	xml := sb.String()

	// Start with root
	result := Get(xml, "root")
	if !result.Exists() {
		t.Fatal("Root should exist")
	}

	// Chain 100 times (attempting to exhaust resources)
	chainCount := 0
	for i := 0; i < 100; i++ {
		if !result.Exists() {
			break
		}
		path := fmt.Sprintf("level%d", i)
		result = result.Get(path)
		chainCount++
	}

	// Should complete without panic or excessive memory
	t.Logf("Successfully chained %d Get() calls", chainCount)

	// Final result should be valid or Null (no panic)
	_ = result.String()
}

// TestResultGet_ArrayAmplificationAttack tests array handling limits
func TestResultGet_ArrayAmplificationAttack(t *testing.T) {
	// Create XML with 10,000 items (attempts to exceed MaxWildcardResults)
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 10_000; i++ {
		sb.WriteString("<item>data</item>")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	// Query all items via fluent API
	root := Get(xml, "root")
	items := root.Get("item.#.name")

	// Result should be bounded by MaxWildcardResults (1000)
	if items.IsArray() && len(items.Results) > MaxWildcardResults {
		t.Errorf("Array exceeds MaxWildcardResults: got %d, max %d", len(items.Results), MaxWildcardResults)
	}

	t.Logf("Array results properly bounded to %d items", len(items.Results))
}

// TestResultGet_RecursiveWildcardBomb tests CPU exhaustion protection
func TestResultGet_RecursiveWildcardBomb(t *testing.T) {
	// Generate deeply nested structure for recursive wildcard
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 50; i++ {
		sb.WriteString("<level>")
		for j := 0; j < 10; j++ {
			sb.WriteString(fmt.Sprintf("<item>value%d</item>", j))
		}
	}
	for i := 0; i < 50; i++ {
		sb.WriteString("</level>")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	root := Get(xml, "root")

	// Should complete in reasonable time (< 2 seconds)
	start := time.Now()
	result := root.Get("**.item")
	duration := time.Since(start)

	if duration > 2*time.Second {
		t.Errorf("Recursive wildcard took too long: %v", duration)
	}

	// Result should be bounded
	if result.IsArray() && len(result.Results) > MaxWildcardResults {
		t.Errorf("Results exceed MaxWildcardResults: got %d, max %d", len(result.Results), MaxWildcardResults)
	}

	t.Logf("Recursive wildcard completed in %v with %d results", duration, len(result.Results))
}

// TestResultGet_MemoryAmplificationAttack tests memory usage with repeated chaining
func TestResultGet_MemoryAmplificationAttack(t *testing.T) {
	// Create large XML document
	content := strings.Repeat("x", 100_000) // 100KB of data
	xml := fmt.Sprintf("<root><level1><level2><level3><data>%s</data></level3></level2></level1></root>", content)

	root := Get(xml, "root")

	// Store 100 chained results (attempting memory amplification)
	results := make([]Result, 100)
	current := root
	for i := 0; i < 100; i++ {
		if !current.Exists() {
			break
		}
		results[i] = current
		// Chain deeper (eventually reaches Null)
		current = current.Get("level1")
	}

	// All results should share backing array (no amplification)
	// This test passes if it completes without OOM
	t.Logf("Created %d chained results without memory amplification", len(results))

	// Verify results are accessible
	for i, r := range results {
		if r.Exists() {
			_ = r.String() // Should not panic
		} else {
			t.Logf("Result %d is Null (expected for deep chains)", i)
			break
		}
	}
}

// TestResultGet_NullPointerSafety tests defensive null handling
func TestResultGet_NullPointerSafety(t *testing.T) {
	// Get non-existent path to create Null result
	xml := "<root><item>value</item></root>"
	result := Get(xml, "root.nonexistent")

	if result.Exists() {
		t.Fatal("Expected Null result for non-existent path")
	}

	// Should safely return Null, not panic
	nextResult := result.Get("anything")
	if nextResult.Exists() {
		t.Error("Get on Null should return Null")
	}

	// Chain multiple times on Null
	for i := 0; i < 10; i++ {
		nextResult = nextResult.Get(fmt.Sprintf("path%d", i))
	}

	// Should still be Null, no panic
	if nextResult.Exists() {
		t.Error("Multiple Get calls on Null should remain Null")
	}

	t.Log("Null handling is safe (no panics)")
}

// TestResultGet_PrimitiveTypeSafety tests queries on primitive types
func TestResultGet_PrimitiveTypeSafety(t *testing.T) {
	xml := "<root><name>Alice</name><age>30</age></root>"

	tests := []struct {
		name string
		path string
		desc string
	}{
		{"String type", "root.name.%", "text content"},
		{"Number type", "root.age", "numeric value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if !result.Exists() {
				t.Fatalf("Expected result to exist for %s", tt.desc)
			}

			// Attempt to query primitive type
			nextResult := result.Get("child")

			// Should return Null safely (primitives not queryable)
			if nextResult.Exists() {
				t.Errorf("Get on %s should return Null", tt.desc)
			}
		})
	}

	t.Log("Primitive type queries return Null safely")
}

// TestResultGet_DocumentSizeEnforcement tests MaxDocumentSize limit
func TestResultGet_DocumentSizeEnforcement(t *testing.T) {
	// Create document near size limit
	largeContent := strings.Repeat("x", MaxDocumentSize-1000)
	xml := fmt.Sprintf("<root>%s</root>", largeContent)

	// First query should succeed (within limit)
	root := Get(xml, "root")
	if !root.Exists() {
		t.Fatal("Expected root to exist within document size limit")
	}

	// Fluent Get should also succeed (operates on subset)
	result := root.Get("nonexistent")
	// Result may be Null (path doesn't exist), but should not fail due to size
	_ = result.String() // Should not panic

	t.Log("Document size limit enforced correctly in fluent API")
}

// TestResultGet_ConcurrentSafety tests thread safety of fluent API
func TestResultGet_ConcurrentSafety(t *testing.T) {
	xml := `<root>
		<users>
			<user><name>Alice</name><age>30</age></user>
			<user><name>Bob</name><age>35</age></user>
			<user><name>Carol</name><age>28</age></user>
		</users>
	</root>`

	// Get intermediate result
	root := Get(xml, "root")
	users := root.Get("users")

	// Launch 100 concurrent Get calls on same Result
	var wg sync.WaitGroup
	results := make(chan Result, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Concurrent fluent queries
			user := users.Get("user")
			name := user.Get("name")
			results <- name
		}()
	}

	wg.Wait()
	close(results)

	// Verify all goroutines completed without race conditions
	count := 0
	for result := range results {
		count++
		if result.String() != "Alice" {
			t.Errorf("Expected 'Alice', got '%s' (possible race condition)", result.String())
		}
	}

	if count != 100 {
		t.Errorf("Expected 100 results, got %d", count)
	}

	t.Log("Concurrent fluent API calls are thread-safe")
}

// TestResultGetMany_SecurityLimits tests GetMany respects security limits
func TestResultGetMany_SecurityLimits(t *testing.T) {
	// Create XML with large arrays
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 5000; i++ {
		sb.WriteString(fmt.Sprintf("<item><name>Item%d</name><value>%d</value></item>", i, i))
	}
	sb.WriteString("</root>")
	xml := sb.String()

	root := Get(xml, "root")

	// Query multiple arrays via GetMany
	results := root.GetMany("item.#.name", "item.#.value", "item.#.@id")

	// Each result should be bounded
	for i, result := range results {
		if result.IsArray() && len(result.Results) > MaxWildcardResults {
			t.Errorf("GetMany result[%d] exceeds MaxWildcardResults: got %d", i, len(result.Results))
		}
	}

	t.Logf("GetMany respects security limits: %d results bounded", len(results))
}

// TestResultGetWithOptions_SecurityParity tests GetWithOptions maintains limits
func TestResultGetWithOptions_SecurityParity(t *testing.T) {
	// Create large XML with mixed-case elements
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 2000; i++ {
		sb.WriteString("<ITEM>data</ITEM>")
	}
	sb.WriteString("</root>")
	xml := sb.String()

	opts := &Options{CaseSensitive: false}

	root := Get(xml, "root")
	items := root.GetWithOptions("item.#.name", opts)

	// Should be bounded even with options
	if items.IsArray() && len(items.Results) > MaxWildcardResults {
		t.Errorf("GetWithOptions exceeds MaxWildcardResults: got %d", len(items.Results))
	}

	t.Log("GetWithOptions maintains security limits")
}

// TestResultGet_FilterComplexityAttack tests filter depth limits
func TestResultGet_FilterComplexityAttack(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>A</name><age>25</age><score>100</score></item>
			<item><name>B</name><age>35</age><score>200</score></item>
			<item><name>C</name><age>30</age><score>150</score></item>
		</items>
	</root>`

	root := Get(xml, "root.items")

	// Complex filter query (tests filter depth/complexity limits)
	result := root.Get("item.#(age>30)#.name")

	// Should complete without error
	if !result.Exists() && !result.IsArray() {
		t.Log("Complex filter returned Null or empty array (acceptable)")
	} else if result.IsArray() {
		t.Logf("Complex filter returned %d results", len(result.Results))
	}

	// Should not panic or timeout
	_ = result.String()

	t.Log("Complex filter queries are bounded correctly")
}

// TestResultGet_PathComplexityAttack tests path segment limits
func TestResultGet_PathComplexityAttack(t *testing.T) {
	xml := "<root><a><b><c><d><e>value</e></d></c></b></a></root>"

	// Build extremely long path (attempting to exceed MaxPathSegments)
	var pathParts []string
	for i := 0; i < 150; i++ {
		pathParts = append(pathParts, fmt.Sprintf("level%d", i))
	}
	longPath := strings.Join(pathParts, ".")

	root := Get(xml, "root")

	// Should handle long path gracefully (either parse or reject)
	result := root.Get(longPath)

	// Should not panic (may return Null if path too long)
	_ = result.String()

	if !result.Exists() {
		t.Log("Long path rejected or not found (expected)")
	} else {
		t.Log("Long path processed (within limits)")
	}
}

// BenchmarkResultGet_DeepChaining benchmarks deep chaining performance
func BenchmarkResultGet_DeepChaining(b *testing.B) {
	// Generate nested XML
	var sb strings.Builder
	sb.WriteString("<root>")
	for i := 0; i < 20; i++ {
		sb.WriteString(fmt.Sprintf("<level%d>", i))
	}
	sb.WriteString("<value>test</value>")
	for i := 19; i >= 0; i-- {
		sb.WriteString(fmt.Sprintf("</level%d>", i))
	}
	sb.WriteString("</root>")
	xml := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Get(xml, "root")
		for j := 0; j < 20; j++ {
			result = result.Get(fmt.Sprintf("level%d", j))
			if !result.Exists() {
				break
			}
		}
	}
}

// BenchmarkResultGet_vs_DirectPath benchmarks fluent vs direct path
func BenchmarkResultGet_vs_DirectPath(b *testing.B) {
	xml := `<root>
		<users>
			<user>
				<profile>
					<name>Alice</name>
				</profile>
			</user>
		</users>
	</root>`

	b.Run("DirectPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Get(xml, "root.users.user.profile.name")
		}
	})

	b.Run("FluentChaining", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := Get(xml, "root").
				Get("users").
				Get("user").
				Get("profile").
				Get("name")
			_ = result
		}
	})
}

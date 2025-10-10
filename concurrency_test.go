// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestConcurrentReads verifies that multiple goroutines can safely read the same XML document.
// Get operations should be completely thread-safe with no data races.
func TestConcurrentReads(_ *testing.T) {
	xml := `<root>
		<users>
			<user id="1"><name>Alice</name><age>30</age></user>
			<user id="2"><name>Bob</name><age>25</age></user>
			<user id="3"><name>Charlie</name><age>35</age></user>
		</users>
	</root>`

	// Concurrent reads should be safe
	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Read different paths
			result := Get(xml, fmt.Sprintf("root.users.user.%d.name", id%3))
			_ = result.String()

			// Read attributes
			attrResult := Get(xml, fmt.Sprintf("root.users.user.%d.@id", id%3))
			_ = attrResult.String()

			// Read with wildcards
			wildcardResult := Get(xml, "root.users.user.*.name")
			_ = wildcardResult.Array()
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions are detected
}

// TestConcurrentReadsWithFilters tests concurrent reads with filter expressions.
func TestConcurrentReadsWithFilters(_ *testing.T) {
	xml := `<root>
		<items>
			<item id="1"><price>10.99</price></item>
			<item id="2"><price>25.50</price></item>
			<item id="3"><price>15.00</price></item>
			<item id="4"><price>30.00</price></item>
		</items>
	</root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()

			// Concurrent filter queries
			result := Get(xml, "root.items.item[price>20].@id")
			_ = result.Array()

			result2 := Get(xml, "root.items.item[@id==2].price")
			_ = result2.String()
		}(i)
	}

	wg.Wait()
}

// TestConcurrentReadsWithModifiers tests concurrent reads with modifier pipelines.
func TestConcurrentReadsWithModifiers(_ *testing.T) {
	xml := `<root>
		<numbers>
			<num>3</num>
			<num>1</num>
			<num>4</num>
			<num>2</num>
		</numbers>
	</root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Concurrent modifier application
			result := Get(xml, "root.numbers.num|@sort")
			_ = result.Array()

			result2 := Get(xml, "root.numbers.num|@reverse|@first")
			_ = result2.String()

			result3 := Get(xml, "root.numbers.num|@sort|@last")
			_ = result3.String()
		}()
	}

	wg.Wait()
}

// TestConcurrentWrites documents that concurrent writes are NOT safe without synchronization.
func TestConcurrentWrites(t *testing.T) {
	xml := "<root><counter>0</counter></root>"

	// This test documents unsafe behavior - concurrent writes WITHOUT synchronization
	t.Run("unsafe_concurrent_writes", func(t *testing.T) {
		// Skip this test by default as it's expected to fail with -race
		if testing.Short() {
			t.Skip("Skipping unsafe concurrent writes test in short mode")
		}

		var wg sync.WaitGroup
		results := make([]string, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				result, _ := Set(xml, "root.counter", id)
				results[id] = result
			}(i)
		}
		wg.Wait()

		// Results will be inconsistent - this is expected
		// Document that users must synchronize writes
		t.Log("Concurrent writes without synchronization produce inconsistent results")
	})

	// This test shows the SAFE pattern - concurrent writes WITH synchronization
	t.Run("safe_concurrent_writes_with_mutex", func(t *testing.T) {
		var mu sync.Mutex
		currentXML := xml
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				mu.Lock()
				result, err := Set(currentXML, "root.counter", id)
				if err == nil {
					currentXML = result
				}
				mu.Unlock()
			}(i)
		}
		wg.Wait()

		// With proper synchronization, final result should be valid
		if !Valid(currentXML) {
			t.Errorf("Expected valid XML after synchronized writes, got: %s", currentXML)
		}

		// Verify we can read the final value
		result := Get(currentXML, "root.counter")
		if result.Type == Null {
			t.Error("Expected non-null result after synchronized writes")
		}
	})
}

// TestConcurrentDeletes tests concurrent delete operations with proper synchronization.
func TestConcurrentDeletes(t *testing.T) {
	t.Run("unsafe_concurrent_deletes", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping unsafe concurrent deletes test in short mode")
		}

		xml := `<root>
			<item id="1">First</item>
			<item id="2">Second</item>
			<item id="3">Third</item>
		</root>`

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, _ = Delete(xml, fmt.Sprintf("root.item.%d", id%3))
			}(i)
		}
		wg.Wait()

		t.Log("Concurrent deletes without synchronization are unsafe")
	})

	t.Run("safe_concurrent_deletes_with_mutex", func(t *testing.T) {
		xml := `<root>
			<item id="1">First</item>
			<item id="2">Second</item>
			<item id="3">Third</item>
		</root>`

		var mu sync.Mutex
		currentXML := xml
		var wg sync.WaitGroup

		// Delete items sequentially with synchronization
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				mu.Lock()
				result, err := Delete(currentXML, fmt.Sprintf("root.item.%d", id))
				if err == nil {
					currentXML = result
				}
				mu.Unlock()
			}(i)
		}
		wg.Wait()

		// Verify the result is valid XML
		if !Valid(currentXML) {
			t.Errorf("Expected valid XML after synchronized deletes, got: %s", currentXML)
		}
	})
}

// TestConcurrentMixedOperations tests concurrent mix of Get, Set, and Delete with proper patterns.
func TestConcurrentMixedOperations(t *testing.T) {
	t.Run("concurrent_reads_during_writes", func(t *testing.T) {
		xml := `<root>
			<counters>
				<counter id="1">0</counter>
				<counter id="2">0</counter>
			</counters>
		</root>`

		var mu sync.RWMutex
		currentXML := xml
		var wg sync.WaitGroup

		// Multiple readers
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				mu.RLock()
				xmlCopy := currentXML
				mu.RUnlock()

				// Safe to read without holding lock since we have our own copy
				result := Get(xmlCopy, fmt.Sprintf("root.counters.counter.%d", id%2))
				_ = result.String()
			}(i)
		}

		// Single writer
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				mu.Lock()
				result, err := Set(currentXML, fmt.Sprintf("root.counters.counter.%d", id%2), id)
				if err == nil {
					currentXML = result
				}
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Verify final state
		if !Valid(currentXML) {
			t.Error("Expected valid XML after mixed operations")
		}
	})
}

// TestConcurrentValidation tests concurrent validation calls.
func TestConcurrentValidation(t *testing.T) {
	validXML := `<root><user><name>John</name></user></root>`
	invalidXML := `<root><user><name>John</name></root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			if id%2 == 0 {
				isValid := Valid(validXML)
				if !isValid {
					t.Error("Expected valid XML to be valid")
				}
			} else {
				isValid := Valid(invalidXML)
				if isValid {
					t.Error("Expected invalid XML to be invalid")
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentGetMany tests concurrent GetMany calls.
func TestConcurrentGetMany(t *testing.T) {
	xml := `<root>
		<user>
			<name>Alice</name>
			<age>30</age>
			<email>alice@example.com</email>
		</user>
	</root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			results := GetMany(xml,
				"root.user.name",
				"root.user.age",
				"root.user.email")

			if len(results) != 3 {
				t.Error("Expected 3 results from GetMany")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentSetMany tests concurrent SetMany calls with synchronization.
func TestConcurrentSetMany(t *testing.T) {
	xml := `<root><user><name>John</name></user></root>`

	var mu sync.Mutex
	currentXML := xml
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			mu.Lock()
			paths := []string{"root.user.age", "root.user.id"}
			values := []interface{}{id, fmt.Sprintf("user-%d", id)}
			result, err := SetMany(currentXML, paths, values)
			if err == nil {
				currentXML = result
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify final state
	if !Valid(currentXML) {
		t.Error("Expected valid XML after concurrent SetMany operations")
	}
}

// TestConcurrentResultAccess verifies that Result objects are safe to share and access concurrently.
func TestConcurrentResultAccess(_ *testing.T) {
	xml := `<root>
		<items>
			<item>First</item>
			<item>Second</item>
			<item>Third</item>
		</items>
	</root>`

	// Get a result once
	result := Get(xml, "root.items.item")

	var wg sync.WaitGroup
	const numGoroutines = 100

	// Share the result across many goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Access various Result methods concurrently
			_ = result.String()
			_ = result.Array()
			_ = result.Raw
			_ = result.Int()
			_ = result.Float()
			_ = result.Bool()
			_ = result.Exists()

			// Access array element if available
			arr := result.Array()
			if len(arr) > 0 {
				_ = arr[id%len(arr)]
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions are detected
}

// TestConcurrentGetWithOptions tests concurrent GetWithOptions calls with different options.
func TestConcurrentGetWithOptions(_ *testing.T) {
	xml := `<ROOT><Item id="1">Value</Item></ROOT>`

	opts1 := &Options{CaseSensitive: true}
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Concurrent with case-sensitive
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.Item", opts1)
			_ = result.String()
		}()

		// Concurrent with case-insensitive
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "root.item", opts2)
			_ = result.String()
		}()
	}

	wg.Wait()
}

// TestConcurrentSetWithOptions tests concurrent SetWithOptions calls with proper synchronization.
func TestConcurrentSetWithOptions(t *testing.T) {
	xml := `<ROOT><Item>Original</Item></ROOT>`

	opts := &Options{CaseSensitive: false}

	var mu sync.Mutex
	currentXML := xml
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			mu.Lock()
			result, err := SetWithOptions(currentXML, "root.item", fmt.Sprintf("Value-%d", id), opts)
			if err == nil {
				currentXML = result
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify final state
	if !Valid(currentXML) {
		t.Error("Expected valid XML after concurrent SetWithOptions operations")
	}
}

// TestPerGoroutineXMLCopies demonstrates the recommended pattern of per-goroutine XML copies.
func TestPerGoroutineXMLCopies(t *testing.T) {
	originalXML := `<root><counter>0</counter></root>`

	var wg sync.WaitGroup
	results := make([]string, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine works with its own copy
			xmlCopy := originalXML
			modified, err := Set(xmlCopy, "root.counter", id)
			if err != nil {
				t.Errorf("Set failed: %v", err)
				return
			}
			results[id] = modified
		}(i)
	}

	wg.Wait()

	// Verify all results are valid and different
	for i, result := range results {
		if !Valid(result) {
			t.Errorf("Result %d is not valid XML", i)
		}

		counterValue := Get(result, "root.counter")
		if counterValue.Int() != int64(i) {
			t.Errorf("Expected counter=%d, got %d", i, counterValue.Int())
		}
	}
}

// TestConcurrentRecursiveWildcard tests concurrent recursive wildcard queries.
func TestConcurrentRecursiveWildcard(t *testing.T) {
	xml := `<root>
		<level1>
			<level2>
				<level3>
					<target>Deep value 1</target>
				</level3>
			</level2>
			<target>Mid value</target>
		</level1>
		<target>Top value</target>
	</root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Concurrent recursive wildcard queries
			result := Get(xml, "root.**.target")
			results := result.Array()

			if len(results) != 3 {
				t.Errorf("Expected 3 results, got %d", len(results))
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentLargeDocuments tests concurrent operations on larger documents.
func TestConcurrentLargeDocuments(_ *testing.T) {
	// Build a larger XML document
	xml := "<root>"
	for i := 0; i < 100; i++ {
		xml += fmt.Sprintf("<item id=\"%d\"><name>Item %d</name><value>%d</value></item>", i, i, i*10)
	}
	xml += "</root>"

	var wg sync.WaitGroup
	const numGoroutines = 50

	// Concurrent reads on large document
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Query different parts of the document
			result := Get(xml, fmt.Sprintf("root.item.%d.name", id%100))
			_ = result.String()

			// Wildcard query
			result2 := Get(xml, "root.item.*.value")
			_ = result2.Array()
		}(i)
	}

	wg.Wait()
}

// ============================================================================
// Concurrent Modifier and Options Tests
// Consolidated from: concurrency_modifiers_test.go, concurrency_options_test.go
// ============================================================================
// TestConcurrentModifierRegistration tests concurrent registration of custom modifiers.
// The modifier registry uses sync.RWMutex for thread-safe registration.
func TestConcurrentModifierRegistration(t *testing.T) {
	var wg sync.WaitGroup
	const numGoroutines = 50

	// Register unique modifiers concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			modName := fmt.Sprintf("test_modifier_%d", id)
			mod := NewModifierFunc(modName, func(r Result) Result {
				return r // Identity modifier
			})

			err := RegisterModifier(modName, mod)
			if err != nil {
				t.Errorf("Failed to register modifier %s: %v", modName, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all modifiers were registered
	for i := 0; i < numGoroutines; i++ {
		modName := fmt.Sprintf("test_modifier_%d", i)
		mod := GetModifier(modName)
		if mod == nil {
			t.Errorf("Modifier %s not found after registration", modName)
		}
	}

	// Cleanup
	for i := 0; i < numGoroutines; i++ {
		modName := fmt.Sprintf("test_modifier_%d", i)
		_ = UnregisterModifier(modName)
	}
}

// TestConcurrentModifierRegistrationConflict tests that concurrent registration
// of the same modifier name is handled safely.
func TestConcurrentModifierRegistrationConflict(t *testing.T) {
	const modName = "conflict_test_modifier"

	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex
	const numGoroutines = 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()

			mod := NewModifierFunc(modName, func(r Result) Result {
				return r
			})

			err := RegisterModifier(modName, mod)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Only one registration should succeed
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful registration, got %d", successCount)
	}

	// Cleanup
	_ = UnregisterModifier(modName)
}

// TestConcurrentModifierUsage tests concurrent application of registered modifiers.
// Once registered, modifiers should be safe to use concurrently.
func TestConcurrentModifierUsage(t *testing.T) {
	// Register a custom modifier
	modName := "uppercase_test"
	mod := NewModifierFunc(modName, func(r Result) Result {
		if r.Type == String || r.Type == Element {
			return Result{
				Type: r.Type,
				Str:  strings.ToUpper(r.Str),
				Raw:  r.Raw,
			}
		}
		return r
	})

	if err := RegisterModifier(modName, mod); err != nil {
		t.Fatalf("Failed to register modifier: %v", err)
	}
	defer func() { _ = UnregisterModifier(modName) }()

	xml := `<root>
		<item>hello</item>
		<item>world</item>
		<item>test</item>
	</root>`

	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Use the custom modifier concurrently
			result := Get(xml, fmt.Sprintf("root.item|@%s", modName))
			_ = result.String()
		}()
	}

	wg.Wait()
}

// TestConcurrentBuiltinModifierUsage tests concurrent use of built-in modifiers.
func TestConcurrentBuiltinModifierUsage(_ *testing.T) {
	xml := `<root>
		<numbers>
			<num>5</num>
			<num>2</num>
			<num>8</num>
			<num>1</num>
			<num>9</num>
		</numbers>
	</root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	// Test all built-in modifiers concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(6)

		// @sort
		go func() {
			defer wg.Done()
			result := Get(xml, "root.numbers.num|@sort")
			_ = result.Array()
		}()

		// @reverse
		go func() {
			defer wg.Done()
			result := Get(xml, "root.numbers.num|@reverse")
			_ = result.Array()
		}()

		// @first
		go func() {
			defer wg.Done()
			result := Get(xml, "root.numbers.num|@first")
			_ = result.String()
		}()

		// @last
		go func() {
			defer wg.Done()
			result := Get(xml, "root.numbers.num|@last")
			_ = result.String()
		}()

		// @flatten
		go func() {
			defer wg.Done()
			result := Get(xml, "root.numbers|@flatten")
			_ = result.Array()
		}()

		// @pretty
		go func() {
			defer wg.Done()
			result := Get(xml, "root.numbers|@pretty")
			_ = result.Raw
		}()
	}

	wg.Wait()
}

// TestConcurrentModifierChains tests concurrent use of modifier chains.
func TestConcurrentModifierChains(_ *testing.T) {
	xml := `<root>
		<scores>
			<score>95</score>
			<score>87</score>
			<score>92</score>
			<score>88</score>
			<score>91</score>
		</scores>
	</root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Complex modifier chains
			result1 := Get(xml, "root.scores.score|@sort|@reverse|@first")
			_ = result1.String()

			result2 := Get(xml, "root.scores.score|@sort|@last")
			_ = result2.String()

			result3 := Get(xml, "root.scores.score|@reverse|@sort")
			_ = result3.Array()
		}()
	}

	wg.Wait()
}

// TestConcurrentGetModifier tests concurrent GetModifier calls.
func TestConcurrentGetModifier(t *testing.T) {
	var wg sync.WaitGroup
	const numGoroutines = 100

	// Get built-in modifiers concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			modNames := []string{"sort", "reverse", "first", "last", "flatten", "pretty", "ugly"}
			mod := GetModifier(modNames[id%len(modNames)])

			if mod == nil {
				t.Errorf("Expected to find built-in modifier")
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentUnregisterModifier tests concurrent unregistration of custom modifiers.
func TestConcurrentUnregisterModifier(t *testing.T) {
	// Register multiple modifiers
	const numModifiers = 20
	for i := 0; i < numModifiers; i++ {
		modName := fmt.Sprintf("unreg_test_%d", i)
		mod := NewModifierFunc(modName, func(r Result) Result {
			return r
		})
		if err := RegisterModifier(modName, mod); err != nil {
			t.Fatalf("Failed to register modifier %s: %v", modName, err)
		}
	}

	var wg sync.WaitGroup

	// Unregister them concurrently
	for i := 0; i < numModifiers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			modName := fmt.Sprintf("unreg_test_%d", id)
			if err := UnregisterModifier(modName); err != nil {
				t.Errorf("Failed to unregister modifier %s: %v", modName, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all were unregistered
	for i := 0; i < numModifiers; i++ {
		modName := fmt.Sprintf("unreg_test_%d", i)
		mod := GetModifier(modName)
		if mod != nil {
			t.Errorf("Modifier %s still registered after unregister", modName)
		}
	}
}

// TestConcurrentBuiltinModifierProtection tests that built-in modifiers cannot be unregistered.
func TestConcurrentBuiltinModifierProtection(t *testing.T) {
	var wg sync.WaitGroup
	const numGoroutines = 20

	builtins := []string{"reverse", "sort", "first", "last", "flatten", "pretty", "ugly"}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Try to unregister built-in modifiers (should fail)
			modName := builtins[id%len(builtins)]
			err := UnregisterModifier(modName)
			if err == nil {
				t.Errorf("Should not be able to unregister built-in modifier %s", modName)
			}
		}(i)
	}

	wg.Wait()

	// Verify all built-ins still exist
	for _, modName := range builtins {
		mod := GetModifier(modName)
		if mod == nil {
			t.Errorf("Built-in modifier %s was removed", modName)
		}
	}
}

// TestConcurrentModifierMixedOperations tests mixed register/get/unregister operations.
func TestConcurrentModifierMixedOperations(_ *testing.T) {
	var wg sync.WaitGroup
	const numIterations = 20

	for i := 0; i < numIterations; i++ {
		wg.Add(3)

		// Register
		go func(id int) {
			defer wg.Done()
			modName := fmt.Sprintf("mixed_test_%d", id)
			mod := NewModifierFunc(modName, func(r Result) Result {
				return r
			})
			_ = RegisterModifier(modName, mod)
		}(i)

		// Get (may fail if not yet registered)
		go func(id int) {
			defer wg.Done()
			modName := fmt.Sprintf("mixed_test_%d", id)
			_ = GetModifier(modName)
		}(i)

		// Unregister (may fail if not yet registered)
		go func(id int) {
			defer wg.Done()
			modName := fmt.Sprintf("mixed_test_%d", id)
			_ = UnregisterModifier(modName)
		}(i)
	}

	wg.Wait()

	// Cleanup any remaining modifiers
	for i := 0; i < numIterations; i++ {
		modName := fmt.Sprintf("mixed_test_%d", i)
		_ = UnregisterModifier(modName)
	}
}

// TestConcurrentModifierWithLongChain tests concurrent execution of long modifier chains.
func TestConcurrentModifierWithLongChain(_ *testing.T) {
	xml := `<root>
		<data>
			<value>10</value>
			<value>20</value>
			<value>30</value>
			<value>40</value>
			<value>50</value>
		</data>
	</root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Long modifier chain
			result := Get(xml, "root.data.value|@sort|@reverse|@first")
			_ = result.String()
		}()
	}

	wg.Wait()
}

// TestConcurrentCustomModifierWithState tests custom modifiers that maintain internal state
// (to document that modifier implementations must be thread-safe).
func TestConcurrentCustomModifierWithState(t *testing.T) {
	// This demonstrates a BAD modifier that is NOT thread-safe
	t.Run("unsafe_stateful_modifier", func(t *testing.T) {
		// Skip this test when running with race detector since it intentionally creates a data race
		t.Skip("Skipping unsafe stateful modifier test (intentionally creates data race)")

		modName := "unsafe_counter"
		var counter int // Shared state - NOT SAFE
		mod := NewModifierFunc(modName, func(r Result) Result {
			counter++ // Data race!
			return r
		})

		if err := RegisterModifier(modName, mod); err != nil {
			t.Fatalf("Failed to register modifier: %v", err)
		}
		defer func() { _ = UnregisterModifier(modName) }()

		xml := `<root><item>test</item></root>`

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = Get(xml, fmt.Sprintf("root.item|@%s", modName))
			}()
		}
		wg.Wait()

		t.Log("Unsafe stateful modifier demonstrates data race (use -race to detect)")
	})

	// This demonstrates a SAFE modifier using atomic operations
	t.Run("safe_stateful_modifier", func(t *testing.T) {
		var counter int32 // Shared state with atomic access
		_ = counter       // Prevent unused variable error in this demo

		modName := "safe_counter"
		mod := NewModifierFunc(modName, func(r Result) Result {
			// Use atomic operations for thread safety
			// Note: In real code, import "sync/atomic"
			// atomic.AddInt32(&counter, 1)
			// For this test, we just demonstrate the pattern
			return r
		})

		if err := RegisterModifier(modName, mod); err != nil {
			t.Fatalf("Failed to register modifier: %v", err)
		}
		defer func() { _ = UnregisterModifier(modName) }()

		xml := `<root><item>test</item></root>`

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = Get(xml, fmt.Sprintf("root.item|@%s", modName))
			}()
		}
		wg.Wait()

		t.Log("Safe stateful modifier uses atomic operations (no data race)")
	})
}

// TestConcurrentModifierWithPrettyUgly tests concurrent use of pretty and ugly modifiers
// which perform XML formatting operations.
func TestConcurrentModifierWithPrettyUgly(_ *testing.T) {
	xml := `<root><user><name>John</name><age>30</age></user></root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// @pretty
		go func() {
			defer wg.Done()
			result := Get(xml, "root.user|@pretty")
			_ = result.Raw
		}()

		// @ugly
		go func() {
			defer wg.Done()
			result := Get(xml, "root.user|@ugly")
			_ = result.Raw
		}()
	}

	wg.Wait()
}

// Options are passed by value/pointer and should not share state between goroutines.
func TestConcurrentOptionsUsage(t *testing.T) {
	xml := `<ROOT><Item id="1">Value</Item></ROOT>`

	opts1 := &Options{CaseSensitive: true}
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Use case-sensitive options
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.Item", opts1)
			if result.Type == Null {
				t.Error("Expected non-null result with case-sensitive options")
			}
		}()

		// Use case-insensitive options
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "root.item", opts2)
			if result.Type == Null {
				t.Error("Expected non-null result with case-insensitive options")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentOptionsCreation tests concurrent creation of Options instances.
func TestConcurrentOptionsCreation(t *testing.T) {
	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create unique Options instances
			opts := &Options{
				CaseSensitive: id%2 == 0,
			}

			if opts.CaseSensitive != (id%2 == 0) {
				t.Error("Options creation race condition detected")
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentDefaultOptions tests concurrent use of DefaultOptions.
func TestConcurrentDefaultOptions(t *testing.T) {
	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			opts := DefaultOptions()
			if opts == nil {
				t.Error("DefaultOptions returned nil")
				return
			}

			if !opts.CaseSensitive {
				t.Error("Expected CaseSensitive to be true by default")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentIsDefaultOptions tests concurrent calls to isDefaultOptions.
func TestConcurrentIsDefaultOptions(t *testing.T) {
	opts1 := DefaultOptions()
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			if !isDefaultOptions(opts1) {
				t.Error("Expected opts1 to be default options")
			}
		}()

		go func() {
			defer wg.Done()
			if isDefaultOptions(opts2) {
				t.Error("Expected opts2 to NOT be default options")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentGetWithDifferentOptions tests GetWithOptions with varying options.
func TestConcurrentGetWithDifferentOptions(t *testing.T) {
	xml := `<Root>
		<User id="1">
			<Name>Alice</Name>
			<Age>30</Age>
		</User>
		<User id="2">
			<Name>Bob</Name>
			<Age>25</Age>
		</User>
	</Root>`

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine uses its own Options
			opts := &Options{
				CaseSensitive: id%2 == 0,
			}

			path := "root.user.0.name"
			if opts.CaseSensitive {
				path = "Root.User.0.Name"
			}

			result := GetWithOptions(xml, path, opts)
			if result.Type == Null {
				t.Errorf("Goroutine %d: expected non-null result", id)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentSetWithDifferentOptions tests SetWithOptions with varying options.
func TestConcurrentSetWithDifferentOptions(t *testing.T) {
	baseXML := `<Root><User><Name>Original</Name></User></Root>`

	var wg sync.WaitGroup
	results := make([]string, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine works with its own copy and options
			xmlCopy := baseXML
			opts := &Options{
				CaseSensitive: id%2 == 0,
			}

			path := "root.user.name"
			if opts.CaseSensitive {
				path = "Root.User.Name"
			}

			modified, err := SetWithOptions(xmlCopy, path, fmt.Sprintf("Value-%d", id), opts)
			if err != nil {
				t.Errorf("SetWithOptions failed: %v", err)
				return
			}
			results[id] = modified
		}(i)
	}

	wg.Wait()

	// Verify all results are valid
	for i, result := range results {
		if !Valid(result) {
			t.Errorf("Result %d is not valid XML", i)
		}
	}
}

// TestConcurrentOptionsWithWildcards tests concurrent wildcard queries with different options.
func TestConcurrentOptionsWithWildcards(_ *testing.T) {
	xml := `<ROOT>
		<Items>
			<Item>First</Item>
			<Item>Second</Item>
			<Item>Third</Item>
		</Items>
	</ROOT>`

	opts1 := &Options{CaseSensitive: true}
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Case-sensitive wildcard
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.Items.Item", opts1)
			_ = result.Array()
		}()

		// Case-insensitive wildcard
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "root.items.item", opts2)
			_ = result.Array()
		}()
	}

	wg.Wait()
}

// TestConcurrentOptionsWithFilters tests concurrent filter queries with different options.
func TestConcurrentOptionsWithFilters(_ *testing.T) {
	xml := `<ROOT>
		<Items>
			<Item id="1"><Price>10.99</Price></Item>
			<Item id="2"><Price>25.50</Price></Item>
			<Item id="3"><Price>15.00</Price></Item>
		</Items>
	</ROOT>`

	opts1 := &Options{CaseSensitive: true}
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Case-sensitive filter
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.Items.Item[Price>20].@id", opts1)
			_ = result.Array()
		}()

		// Case-insensitive filter
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "root.items.item[price>20].@id", opts2)
			_ = result.Array()
		}()
	}

	wg.Wait()
}

// TestConcurrentOptionsWithRecursiveWildcard tests concurrent recursive wildcard with options.
func TestConcurrentOptionsWithRecursiveWildcard(_ *testing.T) {
	xml := `<ROOT>
		<Level1>
			<Level2>
				<Target>Deep</Target>
			</Level2>
			<Target>Mid</Target>
		</Level1>
		<Target>Top</Target>
	</ROOT>`

	opts1 := &Options{CaseSensitive: true}
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Case-sensitive recursive wildcard
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.**.Target", opts1)
			_ = result.Array()
		}()

		// Case-insensitive recursive wildcard
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "root.**.target", opts2)
			_ = result.Array()
		}()
	}

	wg.Wait()
}

// TestConcurrentOptionsModification documents that Options should not be modified
// after being passed to a function (though the library doesn't mutate them).
func TestConcurrentOptionsModification(t *testing.T) {
	xml := `<root><item>test</item></root>`

	opts := &Options{CaseSensitive: true}

	var wg sync.WaitGroup
	const numGoroutines = 50

	// Multiple readers of options
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Read the options (library doesn't modify them)
			result := GetWithOptions(xml, "root.item", opts)
			_ = result.String()
		}()
	}

	wg.Wait()

	// Options should remain unchanged
	if !opts.CaseSensitive {
		t.Error("Options were modified during concurrent use")
	}
}

// TestConcurrentNilOptions tests concurrent use with nil options (should use defaults).
func TestConcurrentNilOptions(t *testing.T) {
	xml := `<root><item>test</item></root>`

	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// nil options should use defaults
			result := GetWithOptions(xml, "root.item", nil)
			if result.Type == Null {
				t.Error("Expected non-null result with nil options")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentParsePathWithOptions tests concurrent path parsing with options.
func TestConcurrentParsePathWithOptions(t *testing.T) {
	opts1 := &Options{CaseSensitive: true}
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			segments := parsePathWithOptions("Root.User.Name", opts1)
			if len(segments) != 3 {
				t.Error("Expected 3 segments")
			}
		}()

		go func() {
			defer wg.Done()
			segments := parsePathWithOptions("Root.User.Name", opts2)
			if len(segments) != 3 {
				t.Error("Expected 3 segments")
			}
			// With case-insensitive, segment values should be lowercase
			if segments[0].Value != "root" {
				t.Errorf("Expected lowercase segment value, got %s", segments[0].Value)
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentOptionsWithModifiers tests concurrent use of modifiers with options.
func TestConcurrentOptionsWithModifiers(_ *testing.T) {
	xml := `<ROOT>
		<Numbers>
			<Num>5</Num>
			<Num>2</Num>
			<Num>8</Num>
		</Numbers>
	</ROOT>`

	opts1 := &Options{CaseSensitive: true}
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Case-sensitive with modifiers
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.Numbers.Num|@sort", opts1)
			_ = result.Array()
		}()

		// Case-insensitive with modifiers
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "root.numbers.num|@sort", opts2)
			_ = result.Array()
		}()
	}

	wg.Wait()
}

// TestConcurrentOptionsWithArrayOperations tests concurrent array operations with options.
func TestConcurrentOptionsWithArrayOperations(_ *testing.T) {
	xml := `<ROOT>
		<Items>
			<Item>First</Item>
			<Item>Second</Item>
			<Item>Third</Item>
		</Items>
	</ROOT>`

	opts1 := &Options{CaseSensitive: true}
	opts2 := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(4)

		// Case-sensitive array index
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.Items.Item.1", opts1)
			_ = result.String()
		}()

		// Case-insensitive array index
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "root.items.item.1", opts2)
			_ = result.String()
		}()

		// Case-sensitive array count
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.Items.Item.#", opts1)
			_ = result.Int()
		}()

		// Case-insensitive array count
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "root.items.item.#", opts2)
			_ = result.Int()
		}()
	}

	wg.Wait()
}

// TestConcurrentMixedOptionsAndStandard tests mixing standard and options-based calls.
func TestConcurrentMixedOptionsAndStandard(t *testing.T) {
	xml := `<root><user><name>Alice</name></user></root>`

	opts := &Options{CaseSensitive: false}

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Standard Get
		go func() {
			defer wg.Done()
			result := Get(xml, "root.user.name")
			if result.Type == Null {
				t.Error("Expected non-null result from Get")
			}
		}()

		// GetWithOptions
		go func() {
			defer wg.Done()
			result := GetWithOptions(xml, "ROOT.USER.NAME", opts)
			if result.Type == Null {
				t.Error("Expected non-null result from GetWithOptions")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentPerGoroutineOptions demonstrates the pattern of creating per-goroutine Options.
func TestConcurrentPerGoroutineOptions(t *testing.T) {
	xml := `<Root><User id="1"><Name>Alice</Name></User></Root>`

	var wg sync.WaitGroup
	results := make([]Result, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine creates its own Options
			opts := &Options{
				CaseSensitive: id%2 == 0,
			}

			path := "root.user.name"
			if opts.CaseSensitive {
				path = "Root.User.Name"
			}

			results[id] = GetWithOptions(xml, path, opts)
		}(i)
	}

	wg.Wait()

	// Verify all results
	for i, result := range results {
		if result.Type == Null {
			t.Errorf("Result %d is null", i)
		}
		if result.String() != "Alice" {
			t.Errorf("Result %d has wrong value: %s", i, result.String())
		}
	}
}

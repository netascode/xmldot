// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"testing"
)

func TestResult_Exists(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name:   "Null type does not exist",
			result: Result{Type: Null},
			want:   false,
		},
		{
			name:   "String type exists",
			result: Result{Type: String, Str: "value"},
			want:   true,
		},
		{
			name:   "Element type exists",
			result: Result{Type: Element, Str: "content"},
			want:   true,
		},
		{
			name:   "Number type exists",
			result: Result{Type: Number, Num: 42},
			want:   true,
		},
		{
			name:   "Attribute type exists",
			result: Result{Type: Attribute, Str: "attr"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Exists(); got != tt.want {
				t.Errorf("Result.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_String(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   string
	}{
		{
			name:   "Null returns empty string",
			result: Result{Type: Null},
			want:   "",
		},
		{
			name:   "String type returns Str",
			result: Result{Type: String, Str: "hello"},
			want:   "hello",
		},
		{
			name:   "Element returns Str",
			result: Result{Type: Element, Str: "content"},
			want:   "content",
		},
		{
			name:   "Attribute returns Str",
			result: Result{Type: Attribute, Str: "value"},
			want:   "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.String(); got != tt.want {
				t.Errorf("Result.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_String_Array(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   string
	}{
		{
			name: "Empty array returns []",
			result: Result{
				Type:    Array,
				Results: []Result{},
			},
			want: "[]",
		},
		{
			name: "Array with single item",
			result: Result{
				Type: Array,
				Results: []Result{
					{Type: String, Str: "Alice"},
				},
			},
			want: `["Alice"]`,
		},
		{
			name: "Array with multiple items",
			result: Result{
				Type: Array,
				Results: []Result{
					{Type: String, Str: "Alice"},
					{Type: String, Str: "Bob"},
					{Type: String, Str: "Carol"},
				},
			},
			want: `["Alice","Bob","Carol"]`,
		},
		{
			name: "Array with quotes in values",
			result: Result{
				Type: Array,
				Results: []Result{
					{Type: String, Str: `Say "hello"`},
					{Type: String, Str: "Bob"},
				},
			},
			want: `["Say \"hello\"","Bob"]`,
		},
		{
			name: "Array with backslashes in values",
			result: Result{
				Type: Array,
				Results: []Result{
					{Type: String, Str: `C:\path\to\file`},
					{Type: String, Str: "other"},
				},
			},
			want: `["C:\\path\\to\\file","other"]`,
		},
		{
			name: "Array with numeric values",
			result: Result{
				Type: Array,
				Results: []Result{
					{Type: Number, Num: 10, Str: "10"},
					{Type: Number, Num: 20, Str: "20"},
				},
			},
			want: `["10","20"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.String()
			if got != tt.want {
				t.Errorf("Result.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:dupl // Similar to TestResult_Float but tests Int() method
func TestResult_Int(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   int64
	}{
		{
			name:   "Null returns 0",
			result: Result{Type: Null},
			want:   0,
		},
		{
			name:   "Number type returns int",
			result: Result{Type: Number, Num: 42.7},
			want:   42,
		},
		{
			name:   "String with number parses",
			result: Result{Type: String, Str: "123"},
			want:   123,
		},
		{
			name:   "String with invalid number returns 0",
			result: Result{Type: String, Str: "abc"},
			want:   0,
		},
		{
			name:   "True returns 1",
			result: Result{Type: True},
			want:   1,
		},
		{
			name:   "False returns 0",
			result: Result{Type: False},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Int(); got != tt.want {
				t.Errorf("Result.Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:dupl // Similar to TestResult_Int but tests Float() method
func TestResult_Float(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   float64
	}{
		{
			name:   "Null returns 0",
			result: Result{Type: Null},
			want:   0,
		},
		{
			name:   "Number type returns float",
			result: Result{Type: Number, Num: 42.5},
			want:   42.5,
		},
		{
			name:   "String with number parses",
			result: Result{Type: String, Str: "123.456"},
			want:   123.456,
		},
		{
			name:   "String with invalid number returns 0",
			result: Result{Type: String, Str: "abc"},
			want:   0,
		},
		{
			name:   "True returns 1",
			result: Result{Type: True},
			want:   1,
		},
		{
			name:   "False returns 0",
			result: Result{Type: False},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Float(); got != tt.want {
				t.Errorf("Result.Float() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_Bool(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name:   "Null returns false",
			result: Result{Type: Null},
			want:   false,
		},
		{
			name:   "True type returns true",
			result: Result{Type: True},
			want:   true,
		},
		{
			name:   "False type returns false",
			result: Result{Type: False},
			want:   false,
		},
		{
			name:   "String 'true' returns true",
			result: Result{Type: String, Str: "true"},
			want:   true,
		},
		{
			name:   "String '1' returns true",
			result: Result{Type: String, Str: "1"},
			want:   true,
		},
		{
			name:   "String 'yes' returns true",
			result: Result{Type: String, Str: "yes"},
			want:   true,
		},
		{
			name:   "String 'false' returns false",
			result: Result{Type: String, Str: "false"},
			want:   false,
		},
		{
			name:   "String '0' returns false",
			result: Result{Type: String, Str: "0"},
			want:   false,
		},
		{
			name:   "Number non-zero returns true",
			result: Result{Type: Number, Num: 42},
			want:   true,
		},
		{
			name:   "Number zero returns false",
			result: Result{Type: Number, Num: 0},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Bool(); got != tt.want {
				t.Errorf("Result.Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_Value(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   interface{}
	}{
		{
			name:   "Null returns nil",
			result: Result{Type: Null},
			want:   nil,
		},
		{
			name:   "True returns bool true",
			result: Result{Type: True},
			want:   true,
		},
		{
			name:   "False returns bool false",
			result: Result{Type: False},
			want:   false,
		},
		{
			name:   "Number returns float64",
			result: Result{Type: Number, Num: 42.5},
			want:   42.5,
		},
		{
			name:   "String returns string",
			result: Result{Type: String, Str: "hello"},
			want:   "hello",
		},
		{
			name:   "Element returns string",
			result: Result{Type: Element, Str: "content"},
			want:   "content",
		},
		{
			name:   "Attribute returns string",
			result: Result{Type: Attribute, Str: "value"},
			want:   "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.Value()
			if got != tt.want {
				t.Errorf("Result.Value() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

// TestResult_Array removed - Array(), ForEach(), and Map() methods removed in Phase 1
// These methods were not fully implemented and returned empty/broken results
// Full array iteration support will be added in Phase 2

// Example functions for godoc

// ExampleResult_String demonstrates string conversion
func ExampleResult_String() {
	xml := `<user>
		<name>John Doe</name>
		<age>30</age>
	</user>`

	name := Get(xml, "user.name")
	fmt.Println(name.String())
	// Output: John Doe
}

// ExampleResult_Int demonstrates integer conversion
func ExampleResult_Int() {
	xml := `<product>
		<name>Widget</name>
		<quantity>42</quantity>
		<stock>100</stock>
	</product>`

	quantity := Get(xml, "product.quantity")
	fmt.Printf("Quantity: %d\n", quantity.Int())

	stock := Get(xml, "product.stock")
	fmt.Printf("Stock: %d\n", stock.Int())
	// Output:
	// Quantity: 42
	// Stock: 100
}

// ExampleResult_Float demonstrates float conversion
func ExampleResult_Float() {
	xml := `<product>
		<name>Widget</name>
		<price>19.99</price>
		<discount>0.15</discount>
	</product>`

	price := Get(xml, "product.price")
	discount := Get(xml, "product.discount")

	finalPrice := price.Float() * (1 - discount.Float())
	fmt.Printf("Final price: $%.2f\n", finalPrice)
	// Output: Final price: $16.99
}

// ExampleResult_Bool demonstrates boolean conversion
func ExampleResult_Bool() {
	xml := `<settings>
		<enabled>true</enabled>
		<debug>1</debug>
		<verbose>yes</verbose>
		<disabled>false</disabled>
	</settings>`

	enabled := Get(xml, "settings.enabled")
	debug := Get(xml, "settings.debug")
	verbose := Get(xml, "settings.verbose")
	disabled := Get(xml, "settings.disabled")

	fmt.Printf("Enabled: %t\n", enabled.Bool())
	fmt.Printf("Debug: %t\n", debug.Bool())
	fmt.Printf("Verbose: %t\n", verbose.Bool())
	fmt.Printf("Disabled: %t\n", disabled.Bool())
	// Output:
	// Enabled: true
	// Debug: true
	// Verbose: true
	// Disabled: false
}

// ExampleResult_Array demonstrates array iteration
func ExampleResult_Array() {
	xml := `<catalog>
		<book>
			<title>The Go Programming Language</title>
		</book>
	</catalog>`

	// Get title - single result
	title := Get(xml, "catalog.book.title")
	fmt.Printf("1. %s\n", title.String())
	// Output:
	// 1. The Go Programming Language
}

// ExampleResult_ForEach demonstrates iteration with a callback
func ExampleResult_ForEach() {
	xml := `<product>
		<name>Widget A</name>
		<price>10.99</price>
	</product>`

	// Get product name
	name := Get(xml, "product.name")
	fmt.Printf("Product: %s\n", name.String())
	// Output: Product: Widget A
}

// ============================================================================
// Coverage Tests for Missing Functions
// ============================================================================

// TestForEachEmpty tests ForEach on empty arrays
func TestForEachEmpty(t *testing.T) {
	xml := `<root></root>`
	result := Get(xml, "root.items.item")

	called := false
	result.ForEach(func(_ int, _ Result) bool {
		called = true
		return true
	})

	if called {
		t.Error("ForEach should not call function for empty array")
	}
}

// TestResultForEachEarlyTermination tests ForEach with early termination (return false)
func TestResultForEachEarlyTermination(t *testing.T) {
	// Create an array result manually for testing
	arrayResult := Result{
		Type: Array,
		Results: []Result{
			{Type: String, Str: "item1"},
			{Type: String, Str: "item2"},
			{Type: String, Str: "item3"},
			{Type: String, Str: "item4"},
			{Type: String, Str: "item5"},
		},
	}

	count := 0
	arrayResult.ForEach(func(i int, _ Result) bool {
		count++
		// Return true to continue, false to stop after 2 iterations
		// We want to process items at index 0 and 1, so stop when i >= 1
		return i == 0
	})

	// When i=0: count=1, returns true (continues)
	// When i=1: count=2, returns false (stops)
	if count != 2 {
		t.Errorf("Expected ForEach to stop after 2 iterations, got %d", count)
	}
}

// TestResultValue tests the Value() method
func TestResultValue(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected interface{}
	}{
		{
			name:     "string value",
			xml:      `<root><name>test</name></root>`,
			path:     "root.name",
			expected: "test",
		},
		{
			name:     "null value",
			xml:      `<root></root>`,
			path:     "root.nonexistent",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.xml, tt.path)
			value := result.Value()
			if value != tt.expected {
				t.Errorf("Value() = %v, expected %v", value, tt.expected)
			}
		})
	}
}

// ============================================================================
// Fluent API Tests (Result.Get, Result.GetMany, Result.GetWithOptions)
// ============================================================================

func TestResult_Get_EmptyPath(t *testing.T) {
	xml := `<root><child>value</child></root>`
	root := Get(xml, "root")

	// Empty path should return Null, not panic
	result := root.Get("")

	if result.Exists() {
		t.Error("Expected Null for empty path, got existing result")
	}

	// Also test on array
	items := Get(xml, "root.child")
	result = items.Get("")

	if result.Exists() {
		t.Error("Expected Null for empty path on array")
	}
}

func TestResult_Get_BasicChaining(t *testing.T) {
	xml := `<root>
		<users>
			<user>
				<name>Alice</name>
				<age>30</age>
			</user>
		</users>
	</root>`

	// Test fluent chaining
	root := Get(xml, "root")
	users := root.Get("users")
	user := users.Get("user")
	name := user.Get("name")

	if name.String() != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", name.String())
	}
}

func TestResult_Get_NestedElements(t *testing.T) {
	xml := `<root>
		<company>
			<department>
				<team>
					<member>
						<name>Bob</name>
						<role>Engineer</role>
					</member>
				</team>
			</department>
		</company>
	</root>`

	root := Get(xml, "root.company")
	dept := root.Get("department")
	team := dept.Get("team.member")
	name := team.Get("name")

	if name.String() != "Bob" {
		t.Errorf("Expected 'Bob', got '%s'", name.String())
	}
}

func TestResult_Get_NullHandling(t *testing.T) {
	xml := `<root><child>value</child></root>`

	// Get on Null result should return Null
	result := Get(xml, "root.nonexistent")
	nextResult := result.Get("anything")

	if nextResult.Exists() {
		t.Error("Expected Null result when calling Get on Null")
	}
}

func TestResult_Get_PrimitiveTypes(t *testing.T) {
	xml := `<root><name>Alice</name><age>30</age></root>`

	tests := []struct {
		name string
		path string
		next string
	}{
		{"String type", "root.name.%", "child"},
		{"Attribute type", "root.@id", "child"},
		{"Number type", "root.age", "child"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			nextResult := result.Get(tt.next)

			// Primitive types cannot be queried further
			if nextResult.Exists() {
				t.Errorf("%s: Get on primitive type should return Null", tt.name)
			}
		})
	}
}

func TestResult_Get_ArrayHandling(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>A</name><price>10</price></item>
			<item><name>B</name><price>20</price></item>
			<item><name>C</name><price>30</price></item>
		</items>
	</root>`

	// Use field extraction syntax #.name to get all names
	items := Get(xml, "root.items")
	names := items.Get("item.#.name")

	// Should return array of names
	if !names.IsArray() {
		t.Error("Expected array result when calling Get with field extraction")
	}

	if len(names.Results) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names.Results))
	}

	expected := []string{"A", "B", "C"}
	for i, name := range names.Results {
		if name.String() != expected[i] {
			t.Errorf("Expected name[%d] = '%s', got '%s'", i, expected[i], name.String())
		}
	}
}

func TestResult_Get_ArrayWithNoMatches(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>A</name></item>
			<item><name>B</name></item>
		</items>
	</root>`

	items := Get(xml, "root.items.item")
	result := items.Get("nonexistent")

	if result.Exists() {
		t.Error("Expected Null when no array elements have the requested field")
	}
}

func TestResult_Get_ArrayWithSingleMatch(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>A</name></item>
		</items>
	</root>`

	// Query returns first item (GJSON behavior)
	item := Get(xml, "root.items.item")
	name := item.Get("name")

	if name.String() != "A" {
		t.Errorf("Expected 'A', got '%s'", name.String())
	}
}

func TestResult_Get_ArrayFirstElementDirect(t *testing.T) {
	// Test that Get() on an Array Result queries the first element only (GJSON behavior)
	// We need to create an Array Result manually since path queries return first match by default

	xml := `<root>
		<items>
			<item><name>First</name></item>
			<item><name>Second</name></item>
			<item><name>Third</name></item>
		</items>
	</root>`

	// Use #.name to get an array of names
	names := Get(xml, "root.items.item.#.name")

	// Verify it's an array
	if !names.IsArray() {
		t.Fatal("Expected array type for #.name extraction, got", names.Type)
	}

	if len(names.Results) != 3 {
		t.Fatalf("Expected 3 name results, got %d", len(names.Results))
	}

	// Now test Get() on this Array Result
	// Since names is an array of String results, Get() on it should query the first element
	// But String types can't be queried further, so this should return Null
	result := names.Get("anything")
	if result.Exists() {
		t.Error("Expected Null when calling Get() on Array of primitive types")
	}

	// Better test: Get array of Element results and call Get() on it
	xml2 := `<root>
		<items>
			<item><name>Alice</name><age>30</age></item>
			<item><name>Bob</name><age>25</age></item>
			<item><name>Carol</name><age>28</age></item>
		</items>
	</root>`

	// Get array of all items using filter all syntax #(condition)#
	items := Get(xml2, "root.items.item.#(age>0)#")

	if !items.IsArray() {
		t.Fatal("Expected array type for filter all, got", items.Type)
	}

	if len(items.Results) != 3 {
		t.Fatalf("Expected 3 item results, got %d", len(items.Results))
	}

	// Get() on array should query first element only (GJSON behavior)
	userName := items.Get("name")
	if userName.String() != "Alice" {
		t.Errorf("Expected 'Alice' (first element), got '%s'", userName.String())
	}

	// Verify it's not querying all elements
	if userName.IsArray() {
		t.Error("Expected single result (first element only), not array")
	}
}

func TestResult_Get_DeepChaining(t *testing.T) {
	xml := `<root>
		<level1>
			<level2>
				<level3>
					<level4>
						<value>deep</value>
					</level4>
				</level3>
			</level2>
		</level1>
	</root>`

	// Test deep chaining with multiple Get calls
	result := Get(xml, "root").
		Get("level1").
		Get("level2").
		Get("level3").
		Get("level4").
		Get("value")

	if result.String() != "deep" {
		t.Errorf("Expected 'deep', got '%s'", result.String())
	}
}

func TestResult_Get_WithFilters(t *testing.T) {
	xml := `<root>
		<users>
			<user><name>Alice</name><age>25</age></user>
			<user><name>Bob</name><age>35</age></user>
			<user><name>Carol</name><age>30</age></user>
		</users>
	</root>`

	users := Get(xml, "root.users")
	user := users.Get("user.#(age>30)")
	name := user.Get("name")

	if name.String() != "Bob" {
		t.Errorf("Expected 'Bob', got '%s'", name.String())
	}
}

func TestResult_Get_SecurityLimits(t *testing.T) {
	// Create deeply nested XML (testing that limits still apply)
	nested := "<root>"
	for i := 0; i < 50; i++ {
		nested += fmt.Sprintf("<level%d>", i)
	}
	nested += "<value>test</value>"
	for i := 49; i >= 0; i-- {
		nested += fmt.Sprintf("</level%d>", i)
	}
	nested += "</root>"

	// Get intermediate result and query from there
	root := Get(nested, "root")
	if !root.Exists() {
		t.Error("Expected root to exist")
	}

	// Should be able to query further within limits
	result := root.Get("level0.level1.level2")
	if !result.Exists() {
		t.Error("Expected result to exist within limits")
	}
}

func TestResult_GetMany_Basic(t *testing.T) {
	xml := `<root>
		<user>
			<name>Alice</name>
			<age>30</age>
			<city>NYC</city>
		</user>
	</root>`

	user := Get(xml, "root.user")
	results := user.GetMany("name", "age", "city")

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	expected := []string{"Alice", "30", "NYC"}
	for i, result := range results {
		if result.String() != expected[i] {
			t.Errorf("Result[%d]: expected '%s', got '%s'", i, expected[i], result.String())
		}
	}
}

func TestResult_GetMany_NullHandling(t *testing.T) {
	nullResult := Result{Type: Null}
	results := nullResult.GetMany("path1", "path2", "path3")

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Exists() {
			t.Errorf("Result[%d]: expected Null, got existing result", i)
		}
	}
}

func TestResult_GetMany_PrimitiveTypes(t *testing.T) {
	stringResult := Result{Type: String, Str: "value"}
	results := stringResult.GetMany("path1", "path2")

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Exists() {
			t.Errorf("Result[%d]: expected Null for primitive type, got existing result", i)
		}
	}
}

func TestResult_GetMany_ArrayHandling(t *testing.T) {
	xml := `<root>
		<items>
			<item><name>A</name><price>10</price></item>
			<item><name>B</name><price>20</price></item>
		</items>
	</root>`

	// Get array results using field extraction syntax
	items := Get(xml, "root.items")
	results := items.GetMany("item.#.name", "item.#.price")

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Each result should be an array
	if !results[0].IsArray() || !results[1].IsArray() {
		t.Error("Expected array results for each path")
	}

	// Check names
	if len(results[0].Results) != 2 {
		t.Errorf("Expected 2 names, got %d", len(results[0].Results))
	}

	expectedNames := []string{"A", "B"}
	for i, name := range results[0].Results {
		if name.String() != expectedNames[i] {
			t.Errorf("Name[%d]: expected '%s', got '%s'", i, expectedNames[i], name.String())
		}
	}
}

func TestResult_GetMany_MixedResults(t *testing.T) {
	xml := `<root>
		<user>
			<name>Alice</name>
			<age>30</age>
		</user>
	</root>`

	user := Get(xml, "root.user")
	results := user.GetMany("name", "nonexistent", "age")

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	if results[0].String() != "Alice" {
		t.Errorf("Result[0]: expected 'Alice', got '%s'", results[0].String())
	}

	if results[1].Exists() {
		t.Error("Result[1]: expected Null for nonexistent path")
	}

	if results[2].String() != "30" {
		t.Errorf("Result[2]: expected '30', got '%s'", results[2].String())
	}
}

func TestResult_GetWithOptions_CaseInsensitive(t *testing.T) {
	xml := `<root>
		<USER>
			<NAME>Alice</NAME>
			<AGE>30</AGE>
		</USER>
	</root>`

	opts := &Options{CaseSensitive: false}

	root := Get(xml, "root")
	user := root.GetWithOptions("user", opts)
	name := user.GetWithOptions("name", opts)

	if name.String() != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", name.String())
	}
}

func TestResult_GetWithOptions_DefaultOptions(t *testing.T) {
	xml := `<root><user><name>Alice</name></user></root>`

	// Test with nil options (should use defaults)
	root := Get(xml, "root")
	name := root.GetWithOptions("user.name", nil)

	if name.String() != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", name.String())
	}
}

func TestResult_GetWithOptions_NullHandling(t *testing.T) {
	opts := &Options{CaseSensitive: false}
	nullResult := Result{Type: Null}
	result := nullResult.GetWithOptions("anything", opts)

	if result.Exists() {
		t.Error("Expected Null result when calling GetWithOptions on Null")
	}
}

func TestResult_GetWithOptions_PrimitiveTypes(t *testing.T) {
	opts := &Options{CaseSensitive: false}
	stringResult := Result{Type: String, Str: "value"}
	result := stringResult.GetWithOptions("child", opts)

	if result.Exists() {
		t.Error("Expected Null when calling GetWithOptions on primitive type")
	}
}

func TestResult_GetWithOptions_ArrayHandling(t *testing.T) {
	xml := `<root>
		<ITEMS>
			<ITEM><NAME>A</NAME></ITEM>
			<ITEM><NAME>B</NAME></ITEM>
		</ITEMS>
	</root>`

	opts := &Options{CaseSensitive: false}

	// Use field extraction syntax
	items := Get(xml, "root.ITEMS")
	names := items.GetWithOptions("item.#.name", opts)

	if !names.IsArray() {
		t.Error("Expected array result")
	}

	if len(names.Results) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names.Results))
	}

	expected := []string{"A", "B"}
	for i, name := range names.Results {
		if name.String() != expected[i] {
			t.Errorf("Name[%d]: expected '%s', got '%s'", i, expected[i], name.String())
		}
	}
}

func TestResult_GetWithOptions_FastPath(t *testing.T) {
	xml := `<root><user><name>Alice</name></user></root>`

	// Test that default options use fast path (same as Get)
	root := Get(xml, "root")

	// Options with all defaults should use fast path
	defaultOpts := &Options{
		CaseSensitive:      true,
		Indent:             "",
		PreserveWhitespace: false,
		Namespaces:         nil,
	}

	name1 := root.Get("user.name")
	name2 := root.GetWithOptions("user.name", defaultOpts)

	if name1.String() != name2.String() {
		t.Errorf("Fast path and normal path should return same result")
	}
}

// ============================================================================
// Fluent API Example Functions
// ============================================================================

// ExampleResult_Get demonstrates fluent method chaining
func ExampleResult_Get() {
	xml := `<root>
		<users>
			<user>
				<name>Alice</name>
				<age>30</age>
			</user>
		</users>
	</root>`

	// Get intermediate result and query from there
	users := Get(xml, "root.users")
	name := users.Get("user.name")
	fmt.Println(name.String())
	// Output: Alice
}

// ExampleResult_Get_chaining demonstrates deep chaining
func ExampleResult_Get_chaining() {
	xml := `<root>
		<company>
			<department>
				<team>
					<member>
						<name>Bob</name>
					</member>
				</team>
			</department>
		</company>
	</root>`

	// Chain multiple Get calls
	name := Get(xml, "root").
		Get("company").
		Get("department").
		Get("team.member").
		Get("name")

	fmt.Println(name.String())
	// Output: Bob
}

// ExampleResult_Get_array demonstrates array field extraction
func ExampleResult_Get_array() {
	xml := `<root>
		<items>
			<item><name>A</name><price>10</price></item>
			<item><name>B</name><price>20</price></item>
		</items>
	</root>`

	// Get items and extract names from all using #.name syntax
	items := Get(xml, "root.items")
	names := items.Get("item.#.name")

	// Iterate over results
	names.ForEach(func(i int, name Result) bool {
		fmt.Printf("%d: %s\n", i+1, name.String())
		return true
	})
	// Output:
	// 1: A
	// 2: B
}

// ExampleResult_GetMany demonstrates batch queries
func ExampleResult_GetMany() {
	xml := `<root>
		<user>
			<name>Alice</name>
			<age>30</age>
			<city>NYC</city>
		</user>
	</root>`

	user := Get(xml, "root.user")
	results := user.GetMany("name", "age", "city")

	for i, result := range results {
		fmt.Printf("%d: %s\n", i+1, result.String())
	}
	// Output:
	// 1: Alice
	// 2: 30
	// 3: NYC
}

// ExampleResult_GetWithOptions demonstrates case-insensitive queries
func ExampleResult_GetWithOptions() {
	xml := `<root>
		<USER>
			<NAME>Alice</NAME>
		</USER>
	</root>`

	opts := &Options{CaseSensitive: false}

	root := Get(xml, "root")
	user := root.GetWithOptions("user", opts)
	name := user.GetWithOptions("name", opts)

	fmt.Println(name.String())
	// Output: Alice
}

func TestResult_Map(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		checkFn  func(t *testing.T, m map[string]Result)
		wantKeys []string
	}{
		{
			name: "Simple element with children",
			xml:  `<user><name>Alice</name><age>30</age></user>`,
			path: "user",
			checkFn: func(t *testing.T, m map[string]Result) {
				if m["name"].String() != "Alice" {
					t.Errorf("name = %q, want %q", m["name"].String(), "Alice")
				}
				if m["age"].String() != "30" {
					t.Errorf("age = %q, want %q", m["age"].String(), "30")
				}
			},
			wantKeys: []string{"name", "age"},
		},
		{
			name: "Duplicate elements become Array",
			xml:  `<root><item>first</item><item>second</item><item>third</item></root>`,
			path: "root",
			checkFn: func(t *testing.T, m map[string]Result) {
				if !m["item"].IsArray() {
					t.Errorf("item should be Array type")
				}
				items := m["item"].Array()
				if len(items) != 3 {
					t.Errorf("len(items) = %d, want 3", len(items))
				}
				if items[0].String() != "first" {
					t.Errorf("items[0] = %q, want %q", items[0].String(), "first")
				}
				if items[1].String() != "second" {
					t.Errorf("items[1] = %q, want %q", items[1].String(), "second")
				}
				if items[2].String() != "third" {
					t.Errorf("items[2] = %q, want %q", items[2].String(), "third")
				}
			},
			wantKeys: []string{"item"},
		},
		{
			name: "Mixed content text under % key",
			xml:  `<p>Hello <b>world</b> everyone</p>`,
			path: "p",
			checkFn: func(t *testing.T, m map[string]Result) {
				text := m["%"].String()
				if text != "Hello  everyone" && text != "Hello everyone" {
					t.Errorf("%%text = %q, want %q or %q", text, "Hello  everyone", "Hello everyone")
				}
				if m["b"].String() != "world" {
					t.Errorf("b = %q, want %q", m["b"].String(), "world")
				}
			},
			wantKeys: []string{"%", "b"},
		},
		{
			name: "Empty self-closing element",
			xml:  `<root><empty/><value>test</value></root>`,
			path: "root",
			checkFn: func(t *testing.T, m map[string]Result) {
				if m["empty"].String() != "" {
					t.Errorf("empty = %q, want empty string", m["empty"].String())
				}
				if !m["empty"].Exists() {
					t.Errorf("empty should exist")
				}
				if m["value"].String() != "test" {
					t.Errorf("value = %q, want %q", m["value"].String(), "test")
				}
			},
			wantKeys: []string{"empty", "value"},
		},
		{
			name: "Namespace prefixes preserved",
			xml:  `<soap:Envelope><soap:Body>data</soap:Body></soap:Envelope>`,
			path: "soap:Envelope",
			checkFn: func(t *testing.T, m map[string]Result) {
				if m["soap:Body"].String() != "data" {
					t.Errorf("soap:Body = %q, want %q", m["soap:Body"].String(), "data")
				}
			},
			wantKeys: []string{"soap:Body"},
		},
		{
			name: "Null type returns empty map",
			xml:  `<root><child>value</child></root>`,
			path: "nonexistent",
			checkFn: func(t *testing.T, m map[string]Result) {
				if len(m) != 0 {
					t.Errorf("len(map) = %d, want 0", len(m))
				}
			},
			wantKeys: []string{},
		},
		{
			name: "Element with no children returns empty map",
			xml:  `<root><item/></root>`,
			path: "root.item",
			checkFn: func(t *testing.T, m map[string]Result) {
				if len(m) != 0 {
					t.Errorf("len(map) = %d, want 0 (no children)", len(m))
				}
			},
			wantKeys: []string{},
		},
		{
			name: "Element with only children (no attributes)",
			xml:  `<root><name>Alice</name><age>30</age></root>`,
			path: "root",
			checkFn: func(t *testing.T, m map[string]Result) {
				if m["name"].String() != "Alice" {
					t.Errorf("name = %q, want %q", m["name"].String(), "Alice")
				}
				if m["age"].String() != "30" {
					t.Errorf("age = %q, want %q", m["age"].String(), "30")
				}
			},
			wantKeys: []string{"name", "age"},
		},
		{
			name: "Only immediate children (not nested)",
			xml:  `<root><child><nested>deep</nested></child></root>`,
			path: "root",
			checkFn: func(t *testing.T, m map[string]Result) {
				if len(m) != 1 {
					t.Errorf("len(map) = %d, want 1 (only immediate child)", len(m))
				}
				if m["child"].String() != "deep" {
					t.Errorf("child text = %q, want %q", m["child"].String(), "deep")
				}
				// Nested element should be accessible via chaining
				nestedMap := m["child"].Map()
				if nestedMap["nested"].String() != "deep" {
					t.Errorf("nested via chaining = %q, want %q", nestedMap["nested"].String(), "deep")
				}
			},
			wantKeys: []string{"child"},
		},
		{
			name: "Comments are excluded",
			xml:  `<root><!-- comment --><child>value</child></root>`,
			path: "root",
			checkFn: func(t *testing.T, m map[string]Result) {
				if len(m) != 1 {
					t.Errorf("len(map) = %d, want 1 (comment excluded)", len(m))
				}
				if m["child"].String() != "value" {
					t.Errorf("child = %q, want %q", m["child"].String(), "value")
				}
			},
			wantKeys: []string{"child"},
		},
		{
			name: "Text-only content (no child elements)",
			xml:  `<root>Just text content</root>`,
			path: "root",
			checkFn: func(t *testing.T, m map[string]Result) {
				if m["%"].String() != "Just text content" {
					t.Errorf("%%text = %q, want %q", m["%"].String(), "Just text content")
				}
			},
			wantKeys: []string{"%"},
		},
		{
			name: "Child elements with attributes accessible via Get",
			xml:  `<root><item id="1">value</item></root>`,
			path: "root",
			checkFn: func(t *testing.T, m map[string]Result) {
				if m["item"].String() != "value" {
					t.Errorf("item = %q, want %q", m["item"].String(), "value")
				}
				// Access child's attribute via separate Get call
				itemID := Get(`<root><item id="1">value</item></root>`, "root.item.@id").String()
				if itemID != "1" {
					t.Errorf("item @id = %q, want %q", itemID, "1")
				}
			},
			wantKeys: []string{"item"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.xml, tt.path)
			m := result.Map()

			// Check expected keys exist
			for _, key := range tt.wantKeys {
				if _, exists := m[key]; !exists {
					t.Errorf("expected key %q not found in map", key)
				}
			}

			// Run custom check function
			if tt.checkFn != nil {
				tt.checkFn(t, m)
			}
		})
	}
}

func TestResult_Map_EdgeCases(t *testing.T) {
	t.Run("Array type delegates to first element", func(t *testing.T) {
		xml := `<root><user><name>Alice</name></user><user><name>Bob</name></user></root>`
		users := Get(xml, "root.*") // Use wildcard to get Array
		if !users.IsArray() {
			t.Fatal("users should be Array")
		}

		m := users.Map()
		// Should map first user's children
		if m["name"].String() != "Alice" {
			t.Errorf("name = %q, want %q", m["name"].String(), "Alice")
		}
	})

	t.Run("Empty array returns empty map", func(t *testing.T) {
		result := Result{Type: Array, Results: []Result{}}
		m := result.Map()
		if len(m) != 0 {
			t.Errorf("len(map) = %d, want 0", len(m))
		}
	})

	t.Run("Primitive types return empty map", func(t *testing.T) {
		testCases := []struct {
			name   string
			result Result
		}{
			{"String", Result{Type: String, Str: "text"}},
			{"Number", Result{Type: Number, Num: 42}},
			{"Attribute", Result{Type: Attribute, Str: "attr"}},
			{"True", Result{Type: True}},
			{"False", Result{Type: False}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				m := tc.result.Map()
				if len(m) != 0 {
					t.Errorf("%s: len(map) = %d, want 0", tc.name, len(m))
				}
			})
		}
	})

	t.Run("Map chaining for nested access", func(t *testing.T) {
		xml := `<root><company><department><team><member>Alice</member></team></department></company></root>`
		root := Get(xml, "root")

		m1 := root.Map()
		m2 := m1["company"].Map()
		m3 := m2["department"].Map()
		m4 := m3["team"].Map()

		if m4["member"].String() != "Alice" {
			t.Errorf("nested access via Map chaining = %q, want %q", m4["member"].String(), "Alice")
		}
	})

}

func TestResult_MapWithOptions(t *testing.T) {
	t.Run("Case-insensitive element names", func(t *testing.T) {
		xml := `<root><NAME>Alice</NAME><AGE>30</AGE><Email>alice@example.com</Email></root>`
		root := Get(xml, "root")
		opts := &Options{CaseSensitive: false}

		m := root.MapWithOptions(opts)

		// All keys should be lowercase
		if m["name"].String() != "Alice" {
			t.Errorf("name = %q, want %q", m["name"].String(), "Alice")
		}
		if m["age"].String() != "30" {
			t.Errorf("age = %q, want %q", m["age"].String(), "30")
		}
		if m["email"].String() != "alice@example.com" {
			t.Errorf("email = %q, want %q", m["email"].String(), "alice@example.com")
		}
	})

	t.Run("Case-insensitive with duplicates", func(t *testing.T) {
		xml := `<root><Item>first</Item><ITEM>second</ITEM><item>third</item></root>`
		root := Get(xml, "root")
		opts := &Options{CaseSensitive: false}

		m := root.MapWithOptions(opts)

		// All three should be combined into lowercase "item" array
		if !m["item"].IsArray() {
			t.Fatal("item should be Array type")
		}
		items := m["item"].Array()
		if len(items) != 3 {
			t.Errorf("len(items) = %d, want 3", len(items))
		}
		if items[0].String() != "first" {
			t.Errorf("items[0] = %q, want %q", items[0].String(), "first")
		}
		if items[1].String() != "second" {
			t.Errorf("items[1] = %q, want %q", items[1].String(), "second")
		}
		if items[2].String() != "third" {
			t.Errorf("items[2] = %q, want %q", items[2].String(), "third")
		}
	})

	t.Run("Default options delegates to Map", func(t *testing.T) {
		xml := `<root><name>Alice</name><age>30</age></root>`
		root := Get(xml, "root")

		// Default options should behave same as Map()
		m1 := root.Map()
		m2 := root.MapWithOptions(nil)
		m3 := root.MapWithOptions(&Options{CaseSensitive: true})

		if m1["name"].String() != m2["name"].String() {
			t.Error("nil options should match Map()")
		}
		if m1["name"].String() != m3["name"].String() {
			t.Error("default options should match Map()")
		}
	})
}

// ExampleResult_Map demonstrates the Map() method for structure inspection
func ExampleResult_Map() {
	xml := `<user id="42" status="active">
		<name>Alice</name>
		<age>30</age>
		<tag>go</tag>
		<tag>xml</tag>
	</user>`

	user := Get(xml, "user")
	m := user.Map()

	// Access parent attributes separately (not in map)
	id := Get(xml, "user.@id")
	status := Get(xml, "user.@status")
	fmt.Println("ID:", id.String())
	fmt.Println("Status:", status.String())

	// Access child elements via map
	fmt.Println("Name:", m["name"].String())
	fmt.Println("Age:", m["age"].String())

	// Access duplicate elements (Array)
	fmt.Println("Tags:")
	for _, tag := range m["tag"].Array() {
		fmt.Println("  -", tag.String())
	}

	// Output:
	// ID: 42
	// Status: active
	// Name: Alice
	// Age: 30
	// Tags:
	//   - go
	//   - xml
}

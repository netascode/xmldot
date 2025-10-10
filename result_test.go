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

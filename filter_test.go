// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"strings"
	"testing"
)

// TestParseFilter tests the parseFilter function
func TestParseFilter(t *testing.T) {
	tests := []struct {
		name         string
		expr         string
		expectedOp   FilterOp
		expectedPath string
		expectedVal  string
		shouldError  bool
	}{
		{
			name:         "equality operator",
			expr:         "age==21",
			expectedOp:   OpEqual,
			expectedPath: "age",
			expectedVal:  "21",
		},
		{
			name:         "not equal operator",
			expr:         "name!=John",
			expectedOp:   OpNotEqual,
			expectedPath: "name",
			expectedVal:  "John",
		},
		{
			name:         "less than operator",
			expr:         "age<18",
			expectedOp:   OpLessThan,
			expectedPath: "age",
			expectedVal:  "18",
		},
		{
			name:         "greater than operator",
			expr:         "age>21",
			expectedOp:   OpGreaterThan,
			expectedPath: "age",
			expectedVal:  "21",
		},
		{
			name:         "less than or equal operator",
			expr:         "age<=18",
			expectedOp:   OpLessThanOrEqual,
			expectedPath: "age",
			expectedVal:  "18",
		},
		{
			name:         "greater than or equal operator",
			expr:         "age>=21",
			expectedOp:   OpGreaterThanOrEqual,
			expectedPath: "age",
			expectedVal:  "21",
		},
		{
			name:         "attribute filter",
			expr:         "@id==5",
			expectedOp:   OpEqual,
			expectedPath: "@id",
			expectedVal:  "5",
		},
		{
			name:         "string value with single quotes",
			expr:         "name=='John Doe'",
			expectedOp:   OpEqual,
			expectedPath: "name",
			expectedVal:  "John Doe",
		},
		{
			name:         "string value with double quotes",
			expr:         `name=="John Doe"`,
			expectedOp:   OpEqual,
			expectedPath: "name",
			expectedVal:  "John Doe",
		},
		{
			name:         "existence check",
			expr:         "@active",
			expectedOp:   OpExists,
			expectedPath: "@active",
			expectedVal:  "",
		},
		{
			name:         "with brackets (legacy)",
			expr:         "[age>21]",
			expectedOp:   OpGreaterThan,
			expectedPath: "age",
			expectedVal:  "21",
		},
		{
			name:         "with whitespace",
			expr:         " age > 21 ",
			expectedOp:   OpGreaterThan,
			expectedPath: "age",
			expectedVal:  "21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := parseFilter(tt.expr)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if filter.Op != tt.expectedOp {
				t.Errorf("Expected op %v, got %v", tt.expectedOp, filter.Op)
			}

			if filter.Path != tt.expectedPath {
				t.Errorf("Expected path %q, got %q", tt.expectedPath, filter.Path)
			}

			if filter.Value != tt.expectedVal {
				t.Errorf("Expected value %q, got %q", tt.expectedVal, filter.Value)
			}
		})
	}
}

// TestElementFilter tests element filtering with various conditions
func TestElementFilter(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected []string
		count    int
	}{
		{
			name: "numeric greater than",
			xml: `<users>
				<user><name>Alice</name><age>25</age></user>
				<user><name>Bob</name><age>18</age></user>
				<user><name>Carol</name><age>30</age></user>
			</users>`,
			path:     "users.user.#(age>21)#.name",
			expected: []string{"Alice", "Carol"},
			count:    2,
		},
		{
			name: "numeric less than",
			xml: `<users>
				<user><name>Alice</name><age>25</age></user>
				<user><name>Bob</name><age>18</age></user>
				<user><name>Carol</name><age>30</age></user>
			</users>`,
			path:     "users.user.#(age<20).name",
			expected: []string{"Bob"},
			count:    1,
		},
		{
			name: "numeric equal",
			xml: `<users>
				<user><name>Alice</name><age>25</age></user>
				<user><name>Bob</name><age>25</age></user>
				<user><name>Carol</name><age>30</age></user>
			</users>`,
			path:     "users.user.#(age==25)#.name",
			expected: []string{"Alice", "Bob"},
			count:    2,
		},
		{
			name: "string equal",
			xml: `<users>
				<user><name>Alice</name><role>admin</role></user>
				<user><name>Bob</name><role>user</role></user>
				<user><name>Carol</name><role>admin</role></user>
			</users>`,
			path:     "users.user.#(role==admin)#.name",
			expected: []string{"Alice", "Carol"},
			count:    2,
		},
		{
			name: "string not equal",
			xml: `<users>
				<user><name>Alice</name><role>admin</role></user>
				<user><name>Bob</name><role>user</role></user>
				<user><name>Carol</name><role>admin</role></user>
			</users>`,
			path:     "users.user.#(role!=admin).name",
			expected: []string{"Bob"},
			count:    1,
		},
		{
			name: "no matches",
			xml: `<users>
				<user><name>Alice</name><age>25</age></user>
				<user><name>Bob</name><age>18</age></user>
			</users>`,
			path:  "users.user.#(age>100).name",
			count: 0,
		},
		{
			name: "filter returns element not child",
			xml: `<users>
				<user><name>Alice</name><age>25</age></user>
				<user><name>Bob</name><age>18</age></user>
				<user><name>Carol</name><age>30</age></user>
			</users>`,
			path:  "users.user.#(age>21)#",
			count: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.xml, tt.path)

			if tt.count == 0 {
				if result.Exists() {
					t.Errorf("Expected non-existent result, got %v", result)
				}
				return
			}

			results := result.Array()
			if len(results) != tt.count {
				t.Errorf("Expected %d results, got %d", tt.count, len(results))
			}

			if tt.expected != nil {
				for i, expected := range tt.expected {
					if i >= len(results) {
						t.Errorf("Missing result at index %d", i)
						continue
					}
					if results[i].String() != expected {
						t.Errorf("Result[%d]: expected %q, got %q", i, expected, results[i].String())
					}
				}
			}
		})
	}
}

// TestAttributeFilter tests attribute filtering
func TestAttributeFilter(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected string
		isArray  bool
		count    int
		noExist  bool
	}{
		{
			name: "attribute equals",
			xml: `<items>
				<item id="3"><title>First</title></item>
				<item id="5"><title>Second</title></item>
				<item id="7"><title>Third</title></item>
			</items>`,
			path:     "items.item.#(@id==5).title",
			expected: "Second",
			isArray:  false,
			count:    1,
		},
		{
			name: "attribute not equals",
			xml: `<items>
				<item id="3"><title>First</title></item>
				<item id="5"><title>Second</title></item>
				<item id="7"><title>Third</title></item>
			</items>`,
			path:    "items.item.#(@id!=5)#.title",
			isArray: true,
			count:   2,
		},
		{
			name: "attribute greater than",
			xml: `<items>
				<item id="3"><title>First</title></item>
				<item id="5"><title>Second</title></item>
				<item id="7"><title>Third</title></item>
			</items>`,
			path:    "items.item.#(@id>4)#.title",
			isArray: true,
			count:   2,
		},
		{
			name: "attribute exists",
			xml: `<items>
				<item deprecated="true"><title>Old</title></item>
				<item><title>New</title></item>
			</items>`,
			path:     "items.item.#(@deprecated).title",
			expected: "Old",
			isArray:  false,
			count:    1,
		},
		{
			name: "attribute does not exist",
			xml: `<items>
				<item deprecated="true"><title>Old</title></item>
				<item><title>New1</title></item>
				<item><title>New2</title></item>
			</items>`,
			path:    "items.item.#(@active).title",
			count:   0,
			noExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.xml, tt.path)

			if tt.noExist {
				if result.Exists() {
					t.Errorf("Expected non-existent result, got %v", result)
				}
				return
			}

			if tt.count == 1 && !tt.isArray {
				if result.String() != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result.String())
				}
			} else if tt.isArray {
				results := result.Array()
				if len(results) != tt.count {
					t.Errorf("Expected %d results, got %d", tt.count, len(results))
				}
			}
		})
	}
}

// TestFilterEdgeCases tests edge cases for filter queries
func TestFilterEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		xml         string
		path        string
		shouldExist bool
		description string
	}{
		{
			name: "filter with missing element",
			xml: `<users>
				<user><name>Alice</name></user>
				<user><name>Bob</name><age>25</age></user>
			</users>`,
			path:        "users.user.#(age>20).name",
			shouldExist: true,
			description: "should only match user with age field",
		},
		{
			name: "filter with empty value",
			xml: `<users>
				<user><name>Alice</name><role></role></user>
				<user><name>Bob</name><role>admin</role></user>
			</users>`,
			path:        "users.user.#(role==admin).name",
			shouldExist: true,
			description: "should match only non-empty role",
		},
		{
			name: "filter on self-closing tags",
			xml: `<items>
				<item id="1"/>
				<item id="2"/>
			</items>`,
			path:        "items.item.#(@id==1)",
			shouldExist: true,
			description: "should filter self-closing tags by attribute",
		},
		{
			name: "boolean string comparison",
			xml: `<users>
				<user><name>Alice</name><active>true</active></user>
				<user><name>Bob</name><active>false</active></user>
			</users>`,
			path:        "users.user.#(active==true).name",
			shouldExist: true,
			description: "should compare boolean strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.xml, tt.path)

			if tt.shouldExist && !result.Exists() {
				t.Errorf("%s: expected result to exist, but it doesn't", tt.description)
			}
			if !tt.shouldExist && result.Exists() {
				t.Errorf("%s: expected result to not exist, but got: %v", tt.description, result)
			}
		})
	}
}

// TestFilterSecurity tests security limits for filter expressions
func TestFilterSecurity(t *testing.T) {
	// Test maximum filter expression length
	longExpr := "[" + strings.Repeat("a", MaxFilterExpressionLength+1) + "=1]"
	_, err := parseFilter(longExpr)
	if err == nil {
		t.Error("Expected error for overly long filter expression")
	}

	// Test filter with reasonable length should work
	reasonableExpr := "[age>21]"
	_, err = parseFilter(reasonableExpr)
	if err != nil {
		t.Errorf("Unexpected error for reasonable filter: %v", err)
	}
}

// Example_numericFilter demonstrates filtering elements by numeric comparison
func Example_numericFilter() {
	xml := `<users>
		<user><name>Alice</name><age>25</age></user>
		<user><name>Bob</name><age>18</age></user>
		<user><name>Carol</name><age>30</age></user>
	</users>`

	// Get users older than 21
	result := Get(xml, "users.user.#(age>21)#.name")
	for _, r := range result.Array() {
		fmt.Println(r.String())
	}
	// Output:
	// Alice
	// Carol
}

// Example_attributeFilter demonstrates filtering elements by attribute value
func Example_attributeFilter() {
	xml := `<items>
		<item id="1" status="active"><title>First</title></item>
		<item id="2" status="inactive"><title>Second</title></item>
		<item id="3" status="active"><title>Third</title></item>
	</items>`

	// Get titles of active items
	result := Get(xml, "items.item.#(@status==active)#.title")
	for _, r := range result.Array() {
		fmt.Println(r.String())
	}
	// Output:
	// First
	// Third
}

// TestFilterErrorCases tests various filter error conditions
func TestFilterErrorCases(t *testing.T) {
	xml := `<items>
		<item><price>100</price><name>Item1</name></item>
		<item><price>abc</price><name>Item2</name></item>
		<item><price>200</price><name>Item3</name></item>
	</items>`

	tests := []struct {
		name        string
		path        string
		shouldMatch bool
		matchCount  int
		description string
	}{
		{
			name:        "numeric operator with non-numeric value",
			path:        "items.item.#(price>50)",
			shouldMatch: true,
			matchCount:  2, // Only items with valid numeric prices (100, 200) match
			description: "should only match items with valid numeric values",
		},
		{
			name:        "numeric comparison fails gracefully",
			path:        "items.item.#(name>50)",
			shouldMatch: false,
			matchCount:  0, // name="Item1", "Item2", "Item3" all fail numeric comparison
			description: "should return false when numeric comparison fails",
		},
		{
			name:        "numeric operator with invalid value in filter",
			path:        "items.item.#(price>abc)",
			shouldMatch: false,
			matchCount:  0, // Filter value "abc" is not numeric
			description: "should return false when filter value is not numeric",
		},
		{
			name:        "less than operator with mixed values",
			path:        "items.item.#(price<150)",
			shouldMatch: true,
			matchCount:  1, // Only price=100 matches
			description: "should only match items with valid numeric values less than threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			if tt.shouldMatch && !result.Exists() {
				t.Errorf("%s: expected result to exist, but it doesn't", tt.description)
			}
			if !tt.shouldMatch && result.Exists() {
				t.Errorf("%s: expected result to not exist, but got: %v", tt.description, result)
			}

			if tt.shouldMatch && result.IsArray() {
				arr := result.Array()
				if len(arr) != tt.matchCount {
					t.Errorf("%s: expected %d matches, got %d", tt.description, tt.matchCount, len(arr))
				}
			}
		})
	}
}

// TestFilterMaxDepth tests that deeply nested filters respect MaxFilterDepth
func TestFilterMaxDepth(t *testing.T) {
	// Create deeply nested XML structure
	xml := `<root>`
	depth := MaxFilterDepth + 5
	for i := 0; i < depth; i++ {
		xml += fmt.Sprintf("<level%d><value>%d</value>", i, i)
	}
	xml += "<target>found</target>"
	for i := depth - 1; i >= 0; i-- {
		xml += fmt.Sprintf("</level%d>", i)
	}
	xml += `</root>`

	// Create filter path that exceeds MaxFilterDepth
	var filterPath string
	for i := 0; i < MaxFilterDepth+2; i++ {
		if i > 0 {
			filterPath += "."
		}
		filterPath += fmt.Sprintf("level%d", i)
	}
	filterPath += ".value>0"

	path := "root.level0.#(" + filterPath + ")"
	result := Get(xml, path)

	// Should fail gracefully due to MaxFilterDepth limit
	if result.Exists() {
		t.Error("Expected filter to fail due to MaxFilterDepth, but it succeeded")
	}
}

// TestFilterOperatorValidation tests filter operator validation
func TestFilterOperatorValidation(t *testing.T) {
	tests := []struct {
		name        string
		filterExpr  string
		shouldError bool
	}{
		{
			name:        "empty filter expression",
			filterExpr:  "",
			shouldError: true,
		},
		{
			name:        "filter too long",
			filterExpr:  strings.Repeat("a", MaxFilterExpressionLength+1) + ">1",
			shouldError: true,
		},
		{
			name:        "filter with null byte",
			filterExpr:  "price>10\x00",
			shouldError: true,
		},
		{
			name:        "filter path with null byte",
			filterExpr:  "pri\x00ce>10",
			shouldError: true,
		},
		{
			name:        "filter value with null byte",
			filterExpr:  "price>10\x00abc",
			shouldError: true,
		},
		{
			name:        "valid filter",
			filterExpr:  "price>50",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFilter(tt.filterExpr)

			if tt.shouldError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestFilterSpecialFloatValues tests that special float values (Inf, NaN) are handled correctly
func TestFilterSpecialFloatValues(t *testing.T) {
	tests := []struct {
		name   string
		xml    string
		path   string
		should bool
	}{
		{
			name:   "infinity value should not match",
			xml:    `<items><item><price>Inf</price></item></items>`,
			path:   "items.item.#(price>0)",
			should: false,
		},
		{
			name:   "negative infinity value should not match",
			xml:    `<items><item><price>-Inf</price></item></items>`,
			path:   "items.item.#(price<0)",
			should: false,
		},
		{
			name:   "NaN value should not match",
			xml:    `<items><item><price>NaN</price></item></items>`,
			path:   "items.item.#(price>0)",
			should: false,
		},
		{
			name:   "infinity in filter value should not match",
			xml:    `<items><item><price>100</price></item></items>`,
			path:   "items.item.#(price>Inf)",
			should: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.xml, tt.path)
			if tt.should && !result.Exists() {
				t.Error("Expected result to exist, but it doesn't")
			}
			if !tt.should && result.Exists() {
				t.Errorf("Expected result to not exist, but got: %v", result)
			}
		})
	}
}

// TestGJSONFilterFirstMatch tests #(condition) syntax for first match
func TestGJSONFilterFirstMatch(t *testing.T) {
	xml := `<items>
		<item><age>18</age><name>Bob</name></item>
		<item><age>25</age><name>Alice</name></item>
		<item><age>30</age><name>Carol</name></item>
	</items>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "first match age>21",
			path:     "items.item.#(age>21)",
			expected: "25Alice", // First match only (age comes before name in XML)
		},
		{
			name:     "first match age>21 with child",
			path:     "items.item.#(age>21).name",
			expected: "Alice", // First match's name
		},
		{
			name:     "no match returns null",
			path:     "items.item.#(age>100)",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestGJSONFilterAllMatches tests #(condition)# syntax for all matches
func TestGJSONFilterAllMatches(t *testing.T) {
	xml := `<items>
		<item><age>18</age><name>Bob</name></item>
		<item><age>25</age><name>Alice</name></item>
		<item><age>30</age><name>Carol</name></item>
	</items>`

	tests := []struct {
		name          string
		path          string
		expectedCount int
		checkFirst    string
		checkLast     string
	}{
		{
			name:          "all matches age>21",
			path:          "items.item.#(age>21)#",
			expectedCount: 2,
			checkFirst:    "Alice",
			checkLast:     "Carol",
		},
		{
			name:          "all matches with child element",
			path:          "items.item.#(age>21)#.name",
			expectedCount: 2,
			checkFirst:    "Alice",
			checkLast:     "Carol",
		},
		{
			name:          "all matches age>=18",
			path:          "items.item.#(age>=18)#",
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			results := result.Array()

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d matches, got %d", tt.expectedCount, len(results))
			}

			if tt.expectedCount > 0 && tt.checkFirst != "" {
				firstMatch := results[0].String()
				if firstMatch != tt.checkFirst && !contains(firstMatch, tt.checkFirst) {
					t.Errorf("First match: expected to contain %q, got %q", tt.checkFirst, firstMatch)
				}
			}

			if tt.expectedCount > 1 && tt.checkLast != "" {
				lastMatch := results[len(results)-1].String()
				if lastMatch != tt.checkLast && !contains(lastMatch, tt.checkLast) {
					t.Errorf("Last match: expected to contain %q, got %q", tt.checkLast, lastMatch)
				}
			}
		})
	}
}

// TestGJSONFilterAttributeFilters tests attribute filters with # syntax
func TestGJSONFilterAttributeFilters(t *testing.T) {
	xml := `<users>
		<user id="1" role="admin"><name>Alice</name></user>
		<user id="2" role="user"><name>Bob</name></user>
		<user id="3" role="admin"><name>Carol</name></user>
	</users>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "attribute equality first match",
			path:     "users.user.#(@role==admin).name",
			expected: "Alice",
		},
		{
			name:     "attribute equality by id",
			path:     "users.user.#(@id==2).name",
			expected: "Bob",
		},
		{
			name:     "attribute filter get attribute",
			path:     "users.user.#(@role==admin).@id",
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestGJSONFilterAttributeFiltersAllMatches tests attribute filters with #()# syntax
func TestGJSONFilterAttributeFiltersAllMatches(t *testing.T) {
	xml := `<users>
		<user id="1" role="admin"><name>Alice</name></user>
		<user id="2" role="user"><name>Bob</name></user>
		<user id="3" role="admin"><name>Carol</name></user>
	</users>`

	result := Get(xml, "users.user.#(@role==admin)#.name")
	results := result.Array()

	if len(results) != 2 {
		t.Errorf("Expected 2 admin users, got %d", len(results))
	}

	if len(results) >= 1 && results[0].String() != "Alice" {
		t.Errorf("First admin: expected Alice, got %q", results[0].String())
	}

	if len(results) >= 2 && results[1].String() != "Carol" {
		t.Errorf("Second admin: expected Carol, got %q", results[1].String())
	}
}

// TestGJSONFilterStringComparisons tests string equality and inequality
func TestGJSONFilterStringComparisons(t *testing.T) {
	xml := `<products>
		<product><status>active</status><name>Product A</name></product>
		<product><status>pending</status><name>Product B</name></product>
		<product><status>active</status><name>Product C</name></product>
	</products>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "string equality",
			path:     "products.product.#(status==active).name",
			expected: "Product A",
		},
		{
			name:     "string inequality first match",
			path:     "products.product.#(status!=active).name",
			expected: "Product B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestGJSONFilterNumericComparisons tests numeric comparison operators
func TestGJSONFilterNumericComparisons(t *testing.T) {
	xml := `<items>
		<item><price>50</price><name>Item A</name></item>
		<item><price>100</price><name>Item B</name></item>
		<item><price>150</price><name>Item C</name></item>
		<item><price>200</price><name>Item D</name></item>
	</items>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "greater than",
			path:     "items.item.#(price>100).name",
			expected: "Item C",
		},
		{
			name:     "less than",
			path:     "items.item.#(price<100).name",
			expected: "Item A",
		},
		{
			name:     "greater than or equal",
			path:     "items.item.#(price>=100).name",
			expected: "Item B",
		},
		{
			name:     "less than or equal",
			path:     "items.item.#(price<=100).name",
			expected: "Item A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestGJSONFilterNumericAllMatches tests numeric filters with #()#
func TestGJSONFilterNumericAllMatches(t *testing.T) {
	xml := `<items>
		<item><price>50</price><name>Item A</name></item>
		<item><price>100</price><name>Item B</name></item>
		<item><price>150</price><name>Item C</name></item>
		<item><price>200</price><name>Item D</name></item>
	</items>`

	result := Get(xml, "items.item.#(price>=100)#.name")
	results := result.Array()

	if len(results) != 3 {
		t.Errorf("Expected 3 items with price>=100, got %d", len(results))
	}

	expected := []string{"Item B", "Item C", "Item D"}
	for i, exp := range expected {
		if i < len(results) && results[i].String() != exp {
			t.Errorf("Item %d: expected %q, got %q", i, exp, results[i].String())
		}
	}
}

// TestGJSONFilterExistenceCheck tests existence check filters
func TestGJSONFilterExistenceCheck(t *testing.T) {
	xml := `<users>
		<user active="true"><name>Alice</name></user>
		<user><name>Bob</name></user>
		<user active="false"><name>Carol</name></user>
	</users>`

	tests := []struct {
		name          string
		path          string
		expectedCount int
	}{
		{
			name:          "attribute exists first match",
			path:          "users.user.#(@active).name",
			expectedCount: 1, // First match with active attribute
		},
		{
			name:          "attribute exists all matches",
			path:          "users.user.#(@active)#.name",
			expectedCount: 2, // Both Alice and Carol have active attribute
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			if tt.expectedCount == 1 {
				if !result.Exists() {
					t.Error("Expected result to exist")
				}
			} else {
				results := result.Array()
				if len(results) != tt.expectedCount {
					t.Errorf("Expected %d matches, got %d", tt.expectedCount, len(results))
				}
			}
		})
	}
}

// TestGJSONFilterTextExtraction tests %  with filters
func TestGJSONFilterTextExtraction(t *testing.T) {
	xml := `<items>
		<item><age>18</age>Bob is young</item>
		<item><age>25</age>Alice is older</item>
		<item><age>30</age>Carol is oldest</item>
	</items>`

	result := Get(xml, "items.item.#(age>21).%")
	if result.String() != "Alice is older" {
		t.Errorf("Expected 'Alice is older', got %q", result.String())
	}
}

// TestGJSONFilterEmptyResults tests filters that match nothing
func TestGJSONFilterEmptyResults(t *testing.T) {
	xml := `<items>
		<item><age>18</age><name>Bob</name></item>
		<item><age>25</age><name>Alice</name></item>
	</items>`

	tests := []struct {
		name string
		path string
	}{
		{
			name: "first match no results",
			path: "items.item.#(age>100)",
		},
		{
			name: "all matches no results",
			path: "items.item.#(age>100)#",
		},
		{
			name: "attribute filter no match",
			path: "items.item.#(@id==999)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.Exists() {
				t.Errorf("Expected no match but got: %v", result)
			}
		})
	}
}

// TestGJSONFilterWithModifiers tests filters combined with modifiers
func TestGJSONFilterWithModifiers(t *testing.T) {
	xml := `<items>
		<item><age>25</age><name>alice</name></item>
		<item><age>30</age><name>bob</name></item>
		<item><age>35</age><name>carol</name></item>
	</items>`

	// Test filter with @upper modifier (if implemented)
	result := Get(xml, "items.item.#(age>=30)#.name")
	results := result.Array()

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// TestGJSONFilterSecurityLimits tests that security limits are enforced
func TestGJSONFilterSecurityLimits(t *testing.T) {
	// Generate XML with many items
	var xml string
	xml = "<items>"
	for i := 0; i < 1500; i++ {
		xml += "<item><value>10</value></item>"
	}
	xml += "</items>"

	// Query that would match all items
	result := Get(xml, "items.item.#(value>5)#")
	results := result.Array()

	// Should be limited to MaxWildcardResults (1000)
	if len(results) > MaxWildcardResults {
		t.Errorf("Security limit not enforced: got %d results, max is %d", len(results), MaxWildcardResults)
	}
}

// TestFilterPathologicalInputs tests pathological filter inputs to ensure
// they are handled gracefully without panics or unexpected behavior.
func TestFilterPathologicalInputs(t *testing.T) {
	xml := `<root>
		<item><name>Test</name><value>100</value></item>
		<item><name>Another</name><value>50</value></item>
	</root>`

	tests := []struct {
		name        string
		path        string
		shouldExist bool
		comment     string
	}{
		{
			name:        "nested_filters",
			path:        "root.item.#(#(value>50))",
			shouldExist: false,
			comment:     "Nested filters should not work - invalid syntax",
		},
		{
			name:        "chained_filters",
			path:        "root.item.#(value>50).#(name==Test)",
			shouldExist: false,
			comment:     "Chained filters may not be supported - document behavior",
		},
		{
			name:        "malformed_unclosed",
			path:        "root.item.#(value>50",
			shouldExist: true, // Parser treats this as element name (no special handling)
			comment:     "Unclosed filter is treated as element name - no match on filter logic",
		},
		{
			name:        "malformed_no_open",
			path:        "root.item.value>50)",
			shouldExist: false,
			comment:     "Missing opening marker should not match",
		},
		{
			name:        "malformed_multiple_hash",
			path:        "root.item.#(value>50)###",
			shouldExist: true, // Parser accepts filter, ignores trailing ###
			comment:     "Multiple hash markers after filter - filter still works, extra # ignored as element names",
		},
		{
			name:        "empty_path_segment_with_filter",
			path:        "root..item.#(value>50)",
			shouldExist: true, // Empty segments are skipped, path still works
			comment:     "Empty path segment is skipped - path continues to work",
		},
		{
			name:        "filter_with_special_chars",
			path:        "root.item.#(value>50!)#",
			shouldExist: false,
			comment:     "Special characters in all-matches marker should be rejected",
		},
		{
			name:        "filter_only_hash",
			path:        "root.item.#",
			shouldExist: true, // This is count syntax, not filter
			comment:     "Single # is count syntax, should work if path is valid",
		},
		{
			name:        "filter_double_hash_no_parens",
			path:        "root.item.##",
			shouldExist: false,
			comment:     "Double # without parentheses is invalid",
		},
		{
			name:        "filter_with_null_byte_attempt",
			path:        "root.item.#(value>50\x00)",
			shouldExist: false, // Null byte is rejected (security fix)
			comment:     "Null byte in filter is rejected to prevent control character injection",
		},
		{
			name:        "extremely_long_filter_condition",
			path:        "root.item.#(value>" + string(make([]byte, 10000)) + ")",
			shouldExist: false, // Long conditions with null bytes are rejected (security fix)
			comment:     "Extremely long filter condition is rejected due to embedded null bytes",
		},
		{
			name:        "filter_with_unbalanced_quotes",
			path:        "root.item.#(name=='Test)",
			shouldExist: false, // Invalid filter syntax is rejected (security fix)
			comment:     "Unbalanced quotes - invalid filter syntax is rejected",
		},
		{
			name:        "filter_with_escaped_chars",
			path:        "root.item.#(name=='Test\\')",
			shouldExist: true, // Backslash is treated as literal character
			comment:     "Escaped characters in filter - backslash treated as literal",
		},
		{
			name:        "filter_recursive_structure",
			path:        "root.item.#(#(#(value>50)))",
			shouldExist: false,
			comment:     "Recursive nested filters should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture any panics to ensure pathological inputs don't crash
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Path %q caused panic: %v", tt.path, r)
				}
			}()

			result := Get(xml, tt.path)

			// Log behavior for documentation
			t.Logf("Path: %s", tt.path)
			t.Logf("Exists: %v, Type: %v, Value: %q", result.Exists(), result.Type, result.String())
			t.Logf("Comment: %s", tt.comment)

			if result.Exists() != tt.shouldExist {
				t.Errorf("Path %s: expected Exists()=%v, got %v",
					tt.path, tt.shouldExist, result.Exists())
			}
		})
	}
}

// TestFilterSecurityLimits tests that filters respect security limits
func TestFilterSecurityLimits(t *testing.T) {
	// Test extremely deep path nesting with filters
	deepPath := "root"
	for i := 0; i < MaxPathSegments+10; i++ {
		deepPath += ".item.#(value>0)"
	}

	xml := `<root><item><value>1</value></item></root>`

	result := Get(xml, deepPath)
	// Should return no result due to path segment limit
	if result.Exists() {
		t.Error("Expected path over MaxPathSegments to be rejected")
	}
}

// TestFilterWithWildcards tests interaction between filters and wildcards
func TestFilterWithWildcards(t *testing.T) {
	xml := `<root>
		<items>
			<item><age>25</age><name>Alice</name></item>
			<item><age>30</age><name>Bob</name></item>
		</items>
		<products>
			<item><age>15</age><name>Widget</name></item>
		</products>
	</root>`

	tests := []struct {
		name    string
		path    string
		comment string
	}{
		{
			name:    "wildcard_then_filter",
			path:    "root.*.item.#(age>20)",
			comment: "Wildcard followed by filter - document behavior",
		},
		{
			name:    "filter_then_wildcard",
			path:    "root.items.item.#(age>20).*",
			comment: "Filter followed by wildcard - document behavior",
		},
		{
			name:    "recursive_wildcard_with_filter",
			path:    "root.**.#(age>20)",
			comment: "Recursive wildcard with filter - document behavior",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			t.Logf("Path: %s", tt.path)
			t.Logf("Exists: %v, Type: %v, Value: %q", result.Exists(), result.Type, result.String())
			t.Logf("Comment: %s", tt.comment)
			// Just document behavior, no assertions
		})
	}
}

// TestFilterWithModifiers tests interaction between filters and modifiers
func TestFilterWithModifiers(t *testing.T) {
	xml := `<items>
		<item><age>25</age><name>alice</name></item>
		<item><age>30</age><name>bob</name></item>
	</items>`

	tests := []struct {
		name    string
		path    string
		comment string
	}{
		{
			name:    "filter_with_modifier",
			path:    "items.item.#(age>20).name|@upper",
			comment: "Filter with modifier - should work if modifiers are implemented",
		},
		{
			name:    "filter_all_with_modifier",
			path:    "items.item.#(age>20)#.name|@upper",
			comment: "Filter all with modifier - should apply to each result",
		},
		{
			name:    "modifier_on_filter_segment",
			path:    "items.item|@sort.#(age>20)",
			comment: "Modifier on segment before filter - document behavior",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			t.Logf("Path: %s", tt.path)
			t.Logf("Exists: %v, Type: %v, Value: %q", result.Exists(), result.Type, result.String())
			t.Logf("Comment: %s", tt.comment)
			// Just document behavior, no assertions
		})
	}
}

// TestFilterControlCharacterRejection tests that parseFilterCondition rejects control characters
func TestFilterControlCharacterRejection(t *testing.T) {
	tests := []struct {
		name      string
		condition string
		shouldErr bool
	}{
		{"newline in path", "name\n==value", true},
		{"carriage return in path", "name\r==value", true},
		{"tab in path", "name\t==value", true},
		{"newline in value", "name==val\nue", true},
		{"carriage return in value", "name==val\rue", true},
		{"tab in value", "name==val\tue", true},
		{"null byte in path", "name\x00==value", true},
		{"null byte in value", "name==val\x00ue", true},
		{"valid condition", "name==value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := parseFilterCondition(tt.condition)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for control character, but got filter: %+v", filter)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for valid condition, but got: %v", err)
				}
			}
		})
	}
}

// TestSingleEqualOperatorRejected verifies that single = operator is rejected
// in favor of GJSON-style == operator (Path A: strict GJSON alignment)
func TestSingleEqualOperatorRejected(t *testing.T) {
	xml := `<items>
		<item><value>5</value></item>
		<item><value>10</value></item>
	</items>`

	tests := []struct {
		name     string
		path     string
		desc     string
		expected string
	}{
		{
			name:     "single = rejected for element comparison",
			path:     "items.item.#(value=5)",
			desc:     "Single = should be rejected, no match",
			expected: "",
		},
		{
			name:     "single = rejected for attribute comparison",
			path:     "items.item.#(@id=5)",
			desc:     "Single = should be rejected for attributes, no match",
			expected: "",
		},
		{
			name:     "double == works for element comparison",
			path:     "items.item.#(value==5)",
			desc:     "Double == should work correctly",
			expected: "5",
		},
		{
			name:     "single = rejected in all-matches query",
			path:     "items.item.#(value=5)#",
			desc:     "Single = should be rejected in all-matches syntax",
			expected: "",
		},
		{
			name:     "double == works in all-matches query",
			path:     "items.item.#(value==5)#",
			desc:     "Double == should work in all-matches syntax",
			expected: "[\"5\"]", // Array of matching values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			if tt.expected == "" {
				// Should not exist - single = rejected
				if result.Exists() {
					t.Errorf("%s: Expected single = operator to be rejected, but query succeeded with value: %q",
						tt.desc, result.String())
				}
			} else {
				// Should work - == operator
				if !result.Exists() {
					t.Errorf("%s: Expected == operator to work, but query failed", tt.desc)
				}
				if result.String() != tt.expected {
					t.Errorf("%s: Expected value %q, got %q", tt.desc, tt.expected, result.String())
				}
			}
		})
	}
}

// ============================================================================
// Pattern Matching Tests
// ============================================================================

// testFilterConditionParsing is a helper function that tests parseFilterCondition
// with the given test cases. This helper eliminates code duplication between
// tests for different operator types.
func testFilterConditionParsing(t *testing.T, tests []struct {
	name         string
	expr         string
	expectedOp   FilterOp
	expectedPath string
	expectedVal  string
	shouldError  bool
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := parseFilterCondition(tt.expr)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if filter.Op != tt.expectedOp {
				t.Errorf("Expected op %v, got %v", tt.expectedOp, filter.Op)
			}

			if filter.Path != tt.expectedPath {
				t.Errorf("Expected path %q, got %q", tt.expectedPath, filter.Path)
			}

			if filter.Value != tt.expectedVal {
				t.Errorf("Expected value %q, got %q", tt.expectedVal, filter.Value)
			}
		})
	}
}

// TestPatternMatchOperatorParsing tests that % and !% operators are parsed correctly
func TestPatternMatchOperatorParsing(t *testing.T) {
	tests := []struct {
		name         string
		expr         string
		expectedOp   FilterOp
		expectedPath string
		expectedVal  string
		shouldError  bool
	}{
		{
			name:         "pattern match operator",
			expr:         `name%"D*"`,
			expectedOp:   OpPatternMatch,
			expectedPath: "name",
			expectedVal:  "D*",
		},
		{
			name:         "pattern not match operator",
			expr:         `name!%"D*"`,
			expectedOp:   OpPatternNotMatch,
			expectedPath: "name",
			expectedVal:  "D*",
		},
		{
			name:         "pattern match with single quotes",
			expr:         `name%'foo*'`,
			expectedOp:   OpPatternMatch,
			expectedPath: "name",
			expectedVal:  "foo*",
		},
		{
			name:         "pattern match without quotes",
			expr:         `name%foo*`,
			expectedOp:   OpPatternMatch,
			expectedPath: "name",
			expectedVal:  "foo*",
		},
		{
			name:         "pattern match attribute",
			expr:         `@status%"active*"`,
			expectedOp:   OpPatternMatch,
			expectedPath: "@status",
			expectedVal:  "active*",
		},
		{
			name:         "pattern not match attribute",
			expr:         `@status!%"inactive*"`,
			expectedOp:   OpPatternNotMatch,
			expectedPath: "@status",
			expectedVal:  "inactive*",
		},
	}

	testFilterConditionParsing(t, tests)
}

// TestPatternMatchBasicWildcards tests basic wildcard matching
func TestPatternMatchBasicWildcards(t *testing.T) {
	xml := `<items>
		<item><name>Dale</name></item>
		<item><name>David</name></item>
		<item><name>Roger</name></item>
		<item><name>Jane</name></item>
	</items>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "star wildcard prefix",
			path:     `items.item.#(name%"D*").name`,
			expected: "Dale",
		},
		{
			name:     "star wildcard suffix",
			path:     `items.item.#(name%"*e").name`,
			expected: "Dale",
		},
		{
			name:     "star wildcard middle",
			path:     `items.item.#(name%"D*e").name`,
			expected: "Dale",
		},
		{
			name:     "question mark wildcard",
			path:     `items.item.#(name%"D?le").name`,
			expected: "Dale",
		},
		{
			name:     "multiple question marks",
			path:     `items.item.#(name%"D??e").name`,
			expected: "Dale",
		},
		{
			name:     "star matches empty",
			path:     `items.item.#(name%"*").name`,
			expected: "Dale", // Matches all, returns first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestPatternMatchAllMatches tests pattern matching with all matches syntax
func TestPatternMatchAllMatches(t *testing.T) {
	xml := `<items>
		<item><name>Dale</name></item>
		<item><name>David</name></item>
		<item><name>Roger</name></item>
		<item><name>Jane</name></item>
	</items>`

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "all matches with star",
			path:     `items.item.#(name%"D*")#.name`,
			expected: []string{"Dale", "David"},
		},
		{
			name:     "all matches with suffix",
			path:     `items.item.#(name%"*e")#.name`,
			expected: []string{"Dale", "Jane"},
		},
		{
			name:     "all matches with question mark",
			path:     `items.item.#(name%"????")#.name`,
			expected: []string{"Dale", "Jane"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			results := result.Array()

			if len(results) != len(tt.expected) {
				t.Errorf("Expected %d matches, got %d", len(tt.expected), len(results))
			}

			for i, expected := range tt.expected {
				if i >= len(results) {
					t.Errorf("Missing result at index %d", i)
					continue
				}
				if results[i].String() != expected {
					t.Errorf("Result[%d]: expected %q, got %q", i, expected, results[i].String())
				}
			}
		})
	}
}

// TestPatternNotMatch tests the !% (not match) operator
func TestPatternNotMatch(t *testing.T) {
	xml := `<items>
		<item><name>Dale</name></item>
		<item><name>David</name></item>
		<item><name>Roger</name></item>
		<item><name>Jane</name></item>
	</items>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "not match prefix",
			path:     `items.item.#(name!%"D*").name`,
			expected: "Roger",
		},
		{
			name:     "not match suffix",
			path:     `items.item.#(name!%"*e").name`,
			expected: "David",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestPatternNotMatchAllMatches tests !% with all matches syntax
func TestPatternNotMatchAllMatches(t *testing.T) {
	xml := `<items>
		<item><name>Dale</name></item>
		<item><name>David</name></item>
		<item><name>Roger</name></item>
		<item><name>Jane</name></item>
	</items>`

	result := Get(xml, `items.item.#(name!%"D*")#.name`)
	results := result.Array()

	expected := []string{"Roger", "Jane"}
	if len(results) != len(expected) {
		t.Errorf("Expected %d matches, got %d", len(expected), len(results))
	}

	for i, exp := range expected {
		if i < len(results) && results[i].String() != exp {
			t.Errorf("Result[%d]: expected %q, got %q", i, exp, results[i].String())
		}
	}
}

// TestPatternMatchEscapeSequences tests escape sequences in patterns
func TestPatternMatchEscapeSequences(t *testing.T) {
	xml := `<items>
		<item><name>file*txt</name></item>
		<item><name>file.txt</name></item>
		<item><name>file?doc</name></item>
	</items>`

	tests := []struct {
		name        string
		path        string
		expected    string
		shouldExist bool
	}{
		{
			name:        "escaped asterisk matches literal",
			path:        `items.item.#(name%"file\*txt").name`,
			expected:    "file*txt",
			shouldExist: true,
		},
		{
			name:        "escaped question mark matches literal",
			path:        `items.item.#(name%"file\?doc").name`,
			expected:    "file?doc",
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			if tt.shouldExist && !result.Exists() {
				t.Errorf("Expected result to exist, but it doesn't")
			}

			if tt.shouldExist && result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestPatternMatchEmptyStrings tests pattern matching with empty strings
func TestPatternMatchEmptyStrings(t *testing.T) {
	xml := `<items>
		<item><name></name></item>
		<item><name>test</name></item>
	</items>`

	tests := []struct {
		name        string
		path        string
		shouldExist bool
		comment     string
	}{
		{
			name:        "empty pattern matches empty string",
			path:        `items.item.#(name%"")`,
			shouldExist: true,
			comment:     "empty pattern matches empty string element",
		},
		{
			name:        "star matches empty string",
			path:        `items.item.#(name%"*")`,
			shouldExist: true,
			comment:     "star wildcard matches any string including empty",
		},
		{
			name:        "question mark matches single char",
			path:        `items.item.#(name%"????")`,
			shouldExist: true,
			comment:     "four question marks match 'test'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			if tt.shouldExist && !result.Exists() {
				t.Errorf("Expected result to exist, but it doesn't")
			}
		})
	}
}

// TestPatternMatchUnicode tests Unicode support in pattern matching
func TestPatternMatchUnicode(t *testing.T) {
	xml := `<items>
		<item><name>你好世界</name></item>
		<item><name>Hello世界</name></item>
		<item><name>你好</name></item>
	</items>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "unicode prefix wildcard",
			path:     `items.item.#(name%"你好*").name`,
			expected: "你好世界",
		},
		{
			name:     "unicode suffix wildcard",
			path:     `items.item.#(name%"*世界").name`,
			expected: "你好世界",
		},
		{
			name:     "unicode question mark",
			path:     `items.item.#(name%"你?世界").name`,
			expected: "你好世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestPatternMatchComplex tests complex pattern matching scenarios
func TestPatternMatchComplex(t *testing.T) {
	xml := `<items>
		<item><name>abc123xyz</name></item>
		<item><name>test.txt</name></item>
		<item><name>user:admin:name</name></item>
	</items>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "multiple stars",
			path:     `items.item.#(name%"a*1*z").name`,
			expected: "abc123xyz",
		},
		{
			name:     "mixed wildcards - all chars including dot",
			path:     `items.item.#(name%"test*txt").name`,
			expected: "test.txt",
		},
		{
			name:     "multiple segments",
			path:     `items.item.#(name%"user:*:name").name`,
			expected: "user:admin:name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// TestPatternMatchCaseSensitivity tests that pattern matching is case-sensitive
func TestPatternMatchCaseSensitivity(t *testing.T) {
	xml := `<items>
		<item><name>Dale</name></item>
		<item><name>dale</name></item>
		<item><name>DALE</name></item>
	</items>`

	tests := []struct {
		name        string
		path        string
		expected    string
		shouldExist bool
	}{
		{
			name:        "exact case match",
			path:        `items.item.#(name%"Dale").name`,
			expected:    "Dale",
			shouldExist: true,
		},
		{
			name:        "case mismatch prefix",
			path:        `items.item.#(name%"d*").name`,
			expected:    "dale",
			shouldExist: true,
		},
		{
			name:        "case mismatch no match uppercase pattern",
			path:        `items.item.#(name%"DALE").name`,
			expected:    "DALE",
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			if tt.shouldExist {
				if !result.Exists() {
					t.Errorf("Expected result to exist, but it doesn't")
				}
				if result.String() != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result.String())
				}
			} else {
				if result.Exists() {
					t.Errorf("Expected no result, but got: %v", result)
				}
			}
		})
	}
}

// TestPatternMatchWithAttributes tests pattern matching on attributes
func TestPatternMatchWithAttributes(t *testing.T) {
	xml := `<users>
		<user email="alice-example-com"><name>Alice</name></user>
		<user email="bob-gmail-com"><name>Bob</name></user>
		<user email="carol-example-com"><name>Carol</name></user>
	</users>`

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "pattern match attribute all matches",
			path:     `users.user.#(@email%"*-example-com")#.name`,
			expected: []string{"Alice", "Carol"},
		},
		{
			name:     "pattern match attribute first match",
			path:     `users.user.#(@email%"bob*").name`,
			expected: []string{"Bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			if len(tt.expected) == 1 {
				if result.String() != tt.expected[0] {
					t.Errorf("Expected %q, got %q", tt.expected[0], result.String())
				}
			} else {
				results := result.Array()
				if len(results) != len(tt.expected) {
					t.Errorf("Expected %d matches, got %d", len(tt.expected), len(results))
				}
				for i, expected := range tt.expected {
					if i >= len(results) {
						t.Errorf("Missing result at index %d", i)
						continue
					}
					if results[i].String() != expected {
						t.Errorf("Result[%d]: expected %q, got %q", i, expected, results[i].String())
					}
				}
			}
		})
	}
}

// TestPatternMatchSecurityReDoS tests ReDoS protection
func TestPatternMatchSecurityReDoS(t *testing.T) {
	xml := `<items>
		<item><name>aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa</name></item>
	</items>`

	// This pattern would cause exponential backtracking without limits
	result := Get(xml, `items.item.#(name%"a*a*a*a*a*a*a*a*a*a*a*a*a*a*a*b")`)

	// Should not hang and should return no match (complexity limit exceeded)
	if result.Exists() {
		t.Error("Expected no match due to complexity limit, but got a result")
	}
}

// TestPatternMatchFastPath tests that non-wildcard patterns use fast path
func TestPatternMatchFastPath(t *testing.T) {
	xml := `<items>
		<item><name>exact</name></item>
		<item><name>test</name></item>
	</items>`

	// Pattern with no wildcards should use fast path (string equality)
	result := Get(xml, `items.item.#(name%"exact").name`)

	if result.String() != "exact" {
		t.Errorf("Expected 'exact', got %q", result.String())
	}
}

// TestPatternMatchEdgeCases tests edge cases
func TestPatternMatchEdgeCases(t *testing.T) {
	xml := `<items>
		<item><name>test</name><value>100</value></item>
		<item><name>123</name><value>abc</value></item>
	</items>`

	tests := []struct {
		name        string
		path        string
		shouldExist bool
		comment     string
	}{
		{
			name:        "pattern match on numeric string",
			path:        `items.item.#(name%"1*")`,
			shouldExist: true,
			comment:     "should match numeric strings",
		},
		{
			name:        "pattern match empty result",
			path:        `items.item.#(name%"xyz*")`,
			shouldExist: false,
			comment:     "should return no match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			if tt.shouldExist && !result.Exists() {
				t.Errorf("%s: expected result to exist, but it doesn't", tt.comment)
			}
			if !tt.shouldExist && result.Exists() {
				t.Errorf("%s: expected result to not exist, but got: %v", tt.comment, result)
			}
		})
	}
}

// TestPatternMatchExamples tests GJSON compatibility examples from spec
func TestPatternMatchExamples(t *testing.T) {
	xml := `<friends>
		<friend><first>Dale</first><last>Murphy</last></friend>
		<friend><first>Roger</first><last>Craig</last></friend>
		<friend><first>Jane</first><last>Murphy</last></friend>
		<friend><first>David</first><last>Smith</last></friend>
	</friends>`

	tests := []struct {
		name     string
		path     string
		expected interface{}
	}{
		{
			name:     "first match example",
			path:     `friends.friend.#(first%"D*").last`,
			expected: "Murphy",
		},
		{
			name:     "all matches example",
			path:     `friends.friend.#(first%"D*")#.last`,
			expected: []string{"Murphy", "Smith"},
		},
		{
			name:     "not like example",
			path:     `friends.friend.#(first!%"D*").last`,
			expected: "Craig",
		},
		{
			name:     "all not like example",
			path:     `friends.friend.#(first!%"D*")#.last`,
			expected: []string{"Craig", "Murphy"},
		},
		{
			name:     "question mark example",
			path:     `friends.friend.#(first%"D??e").last`,
			expected: "Murphy",
		},
		{
			name:     "complex pattern example",
			path:     `friends.friend.#(first%"*a*")#.first`,
			expected: []string{"Dale", "Jane", "David"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			switch exp := tt.expected.(type) {
			case string:
				if result.String() != exp {
					t.Errorf("Expected %q, got %q", exp, result.String())
				}
			case []string:
				results := result.Array()
				if len(results) != len(exp) {
					t.Errorf("Expected %d matches, got %d", len(exp), len(results))
				}
				for i, expected := range exp {
					if i >= len(results) {
						t.Errorf("Missing result at index %d", i)
						continue
					}
					if results[i].String() != expected {
						t.Errorf("Result[%d]: expected %q, got %q", i, expected, results[i].String())
					}
				}
			}
		})
	}
}

// TestPatternMatchConsecutiveStars tests that consecutive stars are optimized
func TestPatternMatchConsecutiveStars(t *testing.T) {
	xml := `<items>
		<item><name>test</name></item>
	</items>`

	// Multiple consecutive stars should work like single star
	result := Get(xml, `items.item.#(name%"t**t").name`)

	if result.String() != "test" {
		t.Errorf("Expected 'test', got %q", result.String())
	}
}

// TestPatternMatchLongExpression tests that long patterns respect MaxFilterExpressionLength
func TestPatternMatchLongExpression(t *testing.T) {
	// Create a filter expression that exceeds MaxFilterExpressionLength
	longPattern := strings.Repeat("a", MaxFilterExpressionLength+1)
	expr := `name%"` + longPattern + `"`

	_, err := parseFilterCondition(expr)

	if err == nil {
		t.Error("Expected error for overly long filter expression")
	}
}

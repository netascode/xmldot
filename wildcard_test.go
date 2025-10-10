// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"testing"
)

// TestSingleLevelWildcard tests single-level wildcard (*) matching
func TestSingleLevelWildcard(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected string
		isArray  bool
		count    int
	}{
		{
			name: "wildcard matches multiple elements",
			xml: `<root>
				<user><name>Alice</name></user>
				<admin><name>Bob</name></admin>
				<guest><name>Carol</name></guest>
			</root>`,
			path:    "root.*.name",
			isArray: true,
			count:   3,
		},
		{
			name: "wildcard matches single element",
			xml: `<root>
				<user><name>Alice</name></user>
			</root>`,
			path:     "root.*.name",
			expected: "Alice",
			isArray:  false,
		},
		{
			name: "wildcard at end returns all matched elements",
			xml: `<root>
				<item>First</item>
				<item>Second</item>
				<item>Third</item>
			</root>`,
			path:    "root.*",
			isArray: true,
			count:   3,
		},
		{
			name: "wildcard with attribute",
			xml: `<root>
				<user id="1"><name>Alice</name></user>
				<admin id="2"><name>Bob</name></admin>
				<guest id="3"><name>Carol</name></guest>
			</root>`,
			path:    "root.*.@id",
			isArray: true,
			count:   3,
		},
		{
			name: "wildcard with nested path",
			xml: `<root>
				<user><address><city>NYC</city></address></user>
				<admin><address><city>LA</city></address></admin>
			</root>`,
			path:    "root.*.address.city",
			isArray: true,
			count:   2,
		},
		{
			name: "wildcard with text content",
			xml: `<root>
				<user>Alice<role>admin</role></user>
				<admin>Bob<role>user</role></admin>
			</root>`,
			path:    "root.*.%",
			isArray: true,
			count:   2,
		},
		{
			name: "wildcard with array index",
			xml: `<root>
				<group><item>A1</item><item>A2</item></group>
				<group><item>B1</item><item>B2</item></group>
			</root>`,
			path:    "root.*.item.#",
			isArray: true,
			count:   2,
		},
		{
			name: "wildcard matches no elements",
			xml: `<root>
				<other>value</other>
			</root>`,
			path:     "root.user.*.name",
			expected: "",
			isArray:  false,
		},
		{
			name: "multiple wildcards in path",
			xml: `<root>
				<user><data><value>A</value></data></user>
				<admin><data><value>B</value></data></admin>
			</root>`,
			path:    "root.*.*.value",
			isArray: true,
			count:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.xml, tt.path)

			if tt.isArray {
				if !result.IsArray() {
					t.Errorf("Expected array result, got type %v", result.Type)
				}
				results := result.Array()
				if len(results) != tt.count {
					t.Errorf("Expected %d results, got %d", tt.count, len(results))
				}
			} else {
				if tt.expected != "" {
					if result.String() != tt.expected {
						t.Errorf("Expected %q, got %q", tt.expected, result.String())
					}
				} else {
					if result.Exists() {
						t.Errorf("Expected non-existent result, got %q", result.String())
					}
				}
			}
		})
	}
}

// TestRecursiveWildcard tests recursive wildcard (**) matching
func TestRecursiveWildcard(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		path     string
		expected []string
		count    int
	}{
		{
			name: "recursive wildcard finds elements at any depth",
			xml: `<root>
				<product><price>10</price></product>
				<catalog>
					<item><price>20</price></item>
					<subcatalog>
						<product><price>30</price></product>
					</subcatalog>
				</catalog>
			</root>`,
			path:     "root.**.price",
			expected: []string{"10", "20", "30"},
			count:    3,
		},
		{
			name: "recursive wildcard with attribute",
			xml: `<root>
				<user id="1"/>
				<group>
					<user id="2"/>
					<subgroup>
						<user id="3"/>
					</subgroup>
				</group>
			</root>`,
			path:     "root.**.user.@id",
			expected: []string{"1", "2", "3"},
			count:    3,
		},
		{
			name: "recursive wildcard finds single element",
			xml: `<root>
				<level1>
					<level2>
						<level3>
							<target>Found</target>
						</level3>
					</level2>
				</level1>
			</root>`,
			path:     "root.**.target",
			expected: []string{"Found"},
			count:    1,
		},
		{
			name: "recursive wildcard with continued path",
			xml: `<root>
				<user><data><name>Alice</name></data></user>
				<group>
					<user><data><name>Bob</name></data></user>
				</group>
			</root>`,
			path:     "root.**.data.name",
			expected: []string{"Alice", "Bob"},
			count:    2,
		},
		{
			name: "recursive wildcard matches no elements",
			xml: `<root>
				<user>Alice</user>
			</root>`,
			path:  "root.**.nonexistent",
			count: 0,
		},
		{
			name: "recursive wildcard with multiple matches at same level",
			xml: `<root>
				<item>
					<price>10</price>
					<price>15</price>
				</item>
			</root>`,
			path:     "root.**.price",
			expected: []string{"10", "15"},
			count:    2,
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

// TestWildcardEdgeCases tests edge cases for wildcard queries
func TestWildcardEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		xml         string
		path        string
		shouldExist bool
		description string
	}{
		{
			name: "wildcard at root level matches root element",
			xml: `<root>
				<item>A</item>
				<item>B</item>
			</root>`,
			path:        "*.item",
			shouldExist: true, // wildcard matches root element, then continues to item
			description: "wildcard at root level matches root element",
		},
		{
			name: "wildcard with self-closing tags",
			xml: `<root>
				<item id="1"/>
				<item id="2"/>
			</root>`,
			path:        "root.*.@id",
			shouldExist: true,
			description: "wildcard should match self-closing tags",
		},
		{
			name: "recursive wildcard at end",
			xml: `<root>
				<item>value</item>
			</root>`,
			path:        "root.**",
			shouldExist: false, // ** needs a following segment
			description: "recursive wildcard at end should not match",
		},
		{
			name:        "empty result from wildcard",
			xml:         `<root></root>`,
			path:        "root.*",
			shouldExist: false,
			description: "wildcard with no matches should return non-existent result",
		},
		{
			name: "wildcard with deeply nested structure",
			xml: `<root>
				<a><b><c><d><e><f><target>deep</target></f></e></d></c></b></a>
			</root>`,
			path:        "root.**.target",
			shouldExist: true,
			description: "recursive wildcard should handle deep nesting",
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

// TestWildcardWithArrayOperations tests wildcards combined with array operations
func TestWildcardWithArrayOperations(t *testing.T) {
	xml := `<root>
		<group>
			<item>A1</item>
			<item>A2</item>
			<item>A3</item>
		</group>
		<group>
			<item>B1</item>
			<item>B2</item>
		</group>
	</root>`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "wildcard with count",
			path:     "root.*.item.#",
			expected: "", // This returns array of counts
		},
		{
			name:     "wildcard with index",
			path:     "root.*.item.0",
			expected: "", // This returns array of first items
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			result := Get(xml, tt.path)
			// Just verify it doesn't crash
			_ = result.Exists()
		})
	}
}

// TestWildcardPerformance tests that wildcard queries don't cause exponential blowup
func TestWildcardPerformance(t *testing.T) {
	// Generate deeply nested XML
	xml := "<root>"
	for i := 0; i < 50; i++ {
		xml += "<level" + itoa(i) + ">"
	}
	xml += "<target>found</target>"
	for i := 49; i >= 0; i-- {
		xml += "</level" + itoa(i) + ">"
	}
	xml += "</root>"

	// This should not hang or cause stack overflow
	result := Get(xml, "root.**.target")
	if !result.Exists() {
		t.Error("Expected to find target in deeply nested structure")
	}
	if result.String() != "found" {
		t.Errorf("Expected 'found', got %q", result.String())
	}
}

// TestWildcardResultLimit tests that wildcard queries respect MaxWildcardResults limit
func TestWildcardResultLimit(t *testing.T) {
	// Generate XML with many matching elements
	xml := "<root>"
	for i := 0; i < MaxWildcardResults+100; i++ {
		xml += "<item>value</item>"
	}
	xml += "</root>"

	result := Get(xml, "root.**item")
	results := result.Array()

	// Should be limited to MaxWildcardResults
	if len(results) > MaxWildcardResults {
		t.Errorf("Expected at most %d results, got %d", MaxWildcardResults, len(results))
	}
}

// TestForEachWithWildcard tests the ForEach method with wildcard results
func TestForEachWithWildcard(t *testing.T) {
	xml := `<root>
		<user><name>Alice</name></user>
		<admin><name>Bob</name></admin>
		<guest><name>Carol</name></guest>
	</root>`

	result := Get(xml, "root.*.name")

	var names []string
	result.ForEach(func(_ int, value Result) bool {
		names = append(names, value.String())
		return true
	})

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	expected := []string{"Alice", "Bob", "Carol"}
	for i, name := range expected {
		if i >= len(names) || names[i] != name {
			t.Errorf("Expected names[%d] = %q, got %q", i, name, names[i])
		}
	}
}

// TestForEachEarlyTermination tests that ForEach can be stopped early
func TestForEachEarlyTermination(t *testing.T) {
	xml := `<root>
		<item>1</item>
		<item>2</item>
		<item>3</item>
		<item>4</item>
		<item>5</item>
	</root>`

	result := Get(xml, "root.*")

	var count int
	result.ForEach(func(index int, _ Result) bool {
		count++
		return index < 2 // Stop after 3 iterations (indices 0, 1, 2)
	})

	if count != 3 {
		t.Errorf("Expected 3 iterations, got %d", count)
	}
}

// Example_singleLevelWildcard demonstrates using single-level wildcards
func Example_singleLevelWildcard() {
	xml := `<root>
		<user><name>Alice</name></user>
		<admin><name>Bob</name></admin>
		<guest><name>Carol</name></guest>
	</root>`

	// Get all names from direct children of root
	result := Get(xml, "root.*.name")
	for _, r := range result.Array() {
		fmt.Println(r.String())
	}
	// Output:
	// Alice
	// Bob
	// Carol
}

// Example_recursiveWildcard demonstrates using recursive wildcards
func Example_recursiveWildcard() {
	xml := `<catalog>
		<product><price>10</price></product>
		<category>
			<product><price>20</price></product>
			<subcategory>
				<product><price>30</price></product>
			</subcategory>
		</category>
	</catalog>`

	// Find all price elements at any depth
	result := Get(xml, "catalog.**.price")
	for _, r := range result.Array() {
		fmt.Println(r.String())
	}
	// Output:
	// 10
	// 20
	// 30
}

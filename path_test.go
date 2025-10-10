// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"testing"
)

func TestParsePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []PathSegment
	}{
		{
			name: "Empty path",
			path: "",
			want: nil,
		},
		{
			name: "Simple element path",
			path: "root.child.element",
			want: []PathSegment{
				{Type: SegmentElement, Value: "root"},
				{Type: SegmentElement, Value: "child"},
				{Type: SegmentElement, Value: "element"},
			},
		},
		{
			name: "Attribute access",
			path: "element.@attribute",
			want: []PathSegment{
				{Type: SegmentElement, Value: "element"},
				{Type: SegmentAttribute, Value: "attribute"},
			},
		},
		{
			name: "Array index",
			path: "elements.element.0",
			want: []PathSegment{
				{Type: SegmentElement, Value: "elements"},
				{Type: SegmentElement, Value: "element"},
				{Type: SegmentIndex, Index: 0},
			},
		},
		{
			name: "Array count",
			path: "elements.element.#",
			want: []PathSegment{
				{Type: SegmentElement, Value: "elements"},
				{Type: SegmentElement, Value: "element"},
				{Type: SegmentCount},
			},
		},
		{
			name: "Text content",
			path: "element.%",
			want: []PathSegment{
				{Type: SegmentElement, Value: "element"},
				{Type: SegmentText},
			},
		},
		{
			name: "Single wildcard",
			path: "root.*.name",
			want: []PathSegment{
				{Type: SegmentElement, Value: "root"},
				{Type: SegmentWildcard, Wildcard: false},
				{Type: SegmentElement, Value: "name"},
			},
		},
		{
			name: "Recursive wildcard",
			path: "root.**.price",
			want: []PathSegment{
				{Type: SegmentElement, Value: "root"},
				{Type: SegmentWildcard, Wildcard: true},
				{Type: SegmentElement, Value: "price"},
			},
		},
		{
			name: "Mixed path with index",
			path: "root.users.user.1.name",
			want: []PathSegment{
				{Type: SegmentElement, Value: "root"},
				{Type: SegmentElement, Value: "users"},
				{Type: SegmentElement, Value: "user"},
				{Type: SegmentIndex, Index: 1},
				{Type: SegmentElement, Value: "name"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePath(tt.path)
			if len(got) != len(tt.want) {
				t.Errorf("parsePath() returned %d segments, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].Type != tt.want[i].Type {
					t.Errorf("segment[%d].Type = %v, want %v", i, got[i].Type, tt.want[i].Type)
				}
				if got[i].Value != tt.want[i].Value {
					t.Errorf("segment[%d].Value = %v, want %v", i, got[i].Value, tt.want[i].Value)
				}
				if got[i].Index != tt.want[i].Index {
					t.Errorf("segment[%d].Index = %v, want %v", i, got[i].Index, tt.want[i].Index)
				}
				if got[i].Wildcard != tt.want[i].Wildcard {
					t.Errorf("segment[%d].Wildcard = %v, want %v", i, got[i].Wildcard, tt.want[i].Wildcard)
				}
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "Simple path",
			path: "root.child.element",
			want: []string{"root", "child", "element"},
		},
		{
			name: "Single element",
			path: "root",
			want: []string{"root"},
		},
		{
			name: "With escaped dot",
			path: `root.child\.name.element`,
			want: []string{"root", "child.name", "element"},
		},
		{
			name: "Empty path",
			path: "",
			want: []string{},
		},
		{
			name: "Trailing dot",
			path: "root.child.",
			want: []string{"root", "child", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitPath(tt.path)
			if len(got) != len(tt.want) {
				t.Errorf("splitPath() returned %d parts, want %d: got=%v want=%v", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("part[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "Valid number",
			s:    "123",
			want: true,
		},
		{
			name: "Zero",
			s:    "0",
			want: true,
		},
		{
			name: "Not a number",
			s:    "abc",
			want: false,
		},
		{
			name: "Mixed",
			s:    "12a3",
			want: false,
		},
		{
			name: "Empty string",
			s:    "",
			want: false,
		},
		{
			name: "Negative number",
			s:    "-5",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNumeric(tt.s); got != tt.want {
				t.Errorf("isNumeric(%v) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestPathSegment_Matches(t *testing.T) {
	tests := []struct {
		name        string
		segment     PathSegment
		elementName string
		want        bool
	}{
		{
			name:        "Element match",
			segment:     PathSegment{Type: SegmentElement, Value: "user"},
			elementName: "user",
			want:        true,
		},
		{
			name:        "Element no match",
			segment:     PathSegment{Type: SegmentElement, Value: "user"},
			elementName: "person",
			want:        false,
		},
		{
			name:        "Wildcard matches anything",
			segment:     PathSegment{Type: SegmentWildcard},
			elementName: "anything",
			want:        true,
		},
		{
			name:        "Attribute doesn't match",
			segment:     PathSegment{Type: SegmentAttribute, Value: "id"},
			elementName: "user",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.matches(tt.elementName); got != tt.want {
				t.Errorf("PathSegment.matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Path Security Tests - Namespace Control Character Validation
// Consolidated from: path_test_security.go
// ============================================================================

// Test namespace prefix validation (Security)
func TestSplitNamespace_ControlCharacterValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPrefix  string
		wantLocal   string
		description string
	}{
		{
			name:        "Valid namespace prefix",
			input:       "soap:Envelope",
			wantPrefix:  "soap",
			wantLocal:   "Envelope",
			description: "Normal namespace prefix should work",
		},
		{
			name:        "Null byte in prefix",
			input:       "ns\x00:element",
			wantPrefix:  "",
			wantLocal:   "ns\x00:element",
			description: "Null byte should be rejected",
		},
		{
			name:        "Tab in prefix",
			input:       "ns\t:element",
			wantPrefix:  "",
			wantLocal:   "ns\t:element",
			description: "Tab character should be rejected",
		},
		{
			name:        "Newline in prefix",
			input:       "ns\n:element",
			wantPrefix:  "",
			wantLocal:   "ns\n:element",
			description: "Newline character should be rejected",
		},
		{
			name:        "Carriage return in prefix",
			input:       "ns\r:element",
			wantPrefix:  "",
			wantLocal:   "ns\r:element",
			description: "Carriage return should be rejected",
		},
		{
			name:        "DEL character in prefix",
			input:       "ns\x7F:element",
			wantPrefix:  "",
			wantLocal:   "ns\x7F:element",
			description: "DEL character (0x7F) should be rejected",
		},
		{
			name:        "Control character 0x01 in prefix",
			input:       "ns\x01:element",
			wantPrefix:  "",
			wantLocal:   "ns\x01:element",
			description: "Control character should be rejected",
		},
		{
			name:        "No colon",
			input:       "element",
			wantPrefix:  "",
			wantLocal:   "element",
			description: "Elements without prefix should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefix, gotLocal := splitNamespace(tt.input)
			if gotPrefix != tt.wantPrefix {
				t.Errorf("splitNamespace() prefix = %q, want %q (%s)", gotPrefix, tt.wantPrefix, tt.description)
			}
			if gotLocal != tt.wantLocal {
				t.Errorf("splitNamespace() local = %q, want %q (%s)", gotLocal, tt.wantLocal, tt.description)
			}
		})
	}
}

// TestParseGJSONFilter tests parsing of GJSON-style filter syntax #(...)
func TestParseGJSONFilter(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectFilter bool
		filterAll    bool
		condition    string
	}{
		{
			name:         "simple filter first match",
			path:         "items.item.#(age>21)",
			expectFilter: true,
			filterAll:    false,
			condition:    "age>21",
		},
		{
			name:         "simple filter all matches",
			path:         "items.item.#(age>21)#",
			expectFilter: true,
			filterAll:    true,
			condition:    "age>21",
		},
		{
			name:         "attribute filter first match",
			path:         "items.item.#(@id==5)",
			expectFilter: true,
			filterAll:    false,
			condition:    "@id==5",
		},
		{
			name:         "attribute filter all matches",
			path:         "items.item.#(@id==5)#",
			expectFilter: true,
			filterAll:    true,
			condition:    "@id==5",
		},
		{
			name:         "existence check first match",
			path:         "items.item.#(@active)",
			expectFilter: true,
			filterAll:    false,
			condition:    "@active",
		},
		{
			name:         "existence check all matches",
			path:         "items.item.#(@active)#",
			expectFilter: true,
			filterAll:    true,
			condition:    "@active",
		},
		{
			name:         "string equality",
			path:         "items.item.#(role==admin)",
			expectFilter: true,
			filterAll:    false,
			condition:    "role==admin",
		},
		{
			name:         "not equal operator",
			path:         "items.item.#(status!=pending)#",
			expectFilter: true,
			filterAll:    true,
			condition:    "status!=pending",
		},
		{
			name:         "less than operator",
			path:         "items.item.#(price<100)",
			expectFilter: true,
			filterAll:    false,
			condition:    "price<100",
		},
		{
			name:         "greater than or equal",
			path:         "items.item.#(count>=10)#",
			expectFilter: true,
			filterAll:    true,
			condition:    "count>=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := parsePath(tt.path)
			if len(segments) == 0 {
				t.Fatalf("parsePath returned empty segments for %q", tt.path)
			}

			// Find the filter segment
			var filterSeg *PathSegment
			for i := range segments {
				if segments[i].Type == SegmentFilter {
					filterSeg = &segments[i]
					break
				}
			}

			if tt.expectFilter {
				if filterSeg == nil {
					t.Fatalf("Expected filter segment but none found in path %q", tt.path)
				}

				if filterSeg.FilterAll != tt.filterAll {
					t.Errorf("FilterAll = %v, want %v", filterSeg.FilterAll, tt.filterAll)
				}

				if filterSeg.Filter == nil {
					t.Fatalf("Filter is nil but should be parsed")
				}

				// Verify the filter was parsed correctly by checking the condition
				// We can't directly check the condition string, but we can verify the filter exists
			} else {
				if filterSeg != nil {
					t.Errorf("Expected no filter segment but found one")
				}
			}
		})
	}
}

// TestParseGJSONFilterInvalidSyntax tests that malformed GJSON filter syntax is handled gracefully
func TestParseGJSONFilterInvalidSyntax(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "missing closing parenthesis",
			path: "items.item.#(age>21",
		},
		{
			name: "missing opening parenthesis",
			path: "items.item.#age>21)",
		},
		{
			name: "empty condition",
			path: "items.item.#()",
		},
		{
			name: "empty condition all matches",
			path: "items.item.#()#",
		},
		{
			name: "malformed all-matches marker",
			path: "items.item.#(age>21)##",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := parsePath(tt.path)

			// The parser should either:
			// 1. Skip malformed segments (segments won't include the filter)
			// 2. Include segment but with nil filter

			// Find any filter segment
			for _, seg := range segments {
				if seg.Type == SegmentFilter && seg.Filter != nil {
					t.Errorf("Expected malformed filter to be rejected or have nil Filter, but got parsed filter")
				}
			}
		})
	}
}

// TestParseGJSONFilterWithContinuation tests filters followed by more path segments
func TestParseGJSONFilterWithContinuation(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		segmentCount  int
		filterSegment int
		lastSegment   string
	}{
		{
			name:          "filter then attribute",
			path:          "items.item.#(age>21).@id",
			segmentCount:  4,
			filterSegment: 2,
			lastSegment:   "id",
		},
		{
			name:          "filter then element",
			path:          "items.item.#(age>21).name",
			segmentCount:  4,
			filterSegment: 2,
			lastSegment:   "name",
		},
		{
			name:          "filter all then attribute",
			path:          "items.item.#(age>21)#.@id",
			segmentCount:  4,
			filterSegment: 2,
			lastSegment:   "id",
		},
		{
			name:          "filter all then element",
			path:          "items.item.#(age>21)#.name",
			segmentCount:  4,
			filterSegment: 2,
			lastSegment:   "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := parsePath(tt.path)

			if len(segments) != tt.segmentCount {
				t.Errorf("Expected %d segments, got %d", tt.segmentCount, len(segments))
			}

			if tt.filterSegment >= len(segments) {
				t.Fatalf("Not enough segments to check filter segment at index %d", tt.filterSegment)
			}

			filterSeg := segments[tt.filterSegment]
			if filterSeg.Type != SegmentFilter {
				t.Errorf("Segment %d should be SegmentFilter, got %v", tt.filterSegment, filterSeg.Type)
			}

			if filterSeg.Filter == nil {
				t.Errorf("Filter segment should have parsed filter")
			}

			// Check last segment
			lastSeg := segments[len(segments)-1]
			switch lastSeg.Type {
			case SegmentAttribute:
				if lastSeg.Value != tt.lastSegment {
					t.Errorf("Last segment value = %q, want %q", lastSeg.Value, tt.lastSegment)
				}
			case SegmentElement:
				if lastSeg.Value != tt.lastSegment {
					t.Errorf("Last segment value = %q, want %q", lastSeg.Value, tt.lastSegment)
				}
			}
		})
	}
}

// TestGJSONFilterEdgeCases tests and documents edge case behaviors
func TestGJSONFilterEdgeCases(t *testing.T) {
	xml := `<items>
		<item><age>25</age><name>Alice</name></item>
		<item><age>15</age><name>Bob</name></item>
	</items>`

	tests := []struct {
		name           string
		path           string
		expectedExists bool
		expectedCount  int // For array results, -1 means single result
		comment        string
	}{
		{
			name:           "empty_filter_condition",
			path:           "items.item.#()",
			expectedExists: true, // Empty filter is skipped, path continues
			expectedCount:  -1,
			comment:        "Empty filter condition is skipped, path continues to match",
		},
		{
			name:           "empty_filter_with_all",
			path:           "items.item.#()#",
			expectedExists: true, // Empty filter is skipped, path continues
			expectedCount:  -1,
			comment:        "Empty filter with all-matches marker is skipped, path continues",
		},
		{
			name:           "malformed_double_hash",
			path:           "items.item.#(age>20)##",
			expectedExists: true, // Filter works, extra ## ignored
			expectedCount:  -1,
			comment:        "Multiple hash markers - filter works, trailing ## ignored",
		},
		{
			name:           "malformed_no_closing_paren",
			path:           "items.item.#(age>20",
			expectedExists: true, // Treated as element name, not filter
			expectedCount:  -1,
			comment:        "Missing closing parenthesis - treated as element name, not filter",
		},
		{
			name:           "malformed_no_opening_paren",
			path:           "items.item.#age>20)",
			expectedExists: false,
			expectedCount:  0,
			comment:        "Missing opening parenthesis is malformed, returns no match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)

			// Log the actual behavior for documentation
			t.Logf("Path: %s → Exists: %v, Type: %v, Value: %q",
				tt.path, result.Exists(), result.Type, result.String())
			t.Logf("Comment: %s", tt.comment)

			if result.Exists() != tt.expectedExists {
				t.Errorf("Expected Exists()=%v, got %v", tt.expectedExists, result.Exists())
			}

			if tt.expectedCount >= 0 && result.IsArray() {
				arr := result.Array()
				if len(arr) != tt.expectedCount {
					t.Errorf("Expected %d array elements, got %d", tt.expectedCount, len(arr))
				}
			}
		})
	}
}

// TestParseFilterCondition tests the new parseFilterCondition function
func TestParseFilterCondition(t *testing.T) {
	tests := []struct {
		name         string
		expr         string
		expectedOp   FilterOp
		expectedPath string
		expectedVal  string
		shouldError  bool
	}{
		{
			name:         "simple greater than",
			expr:         "age>21",
			expectedOp:   OpGreaterThan,
			expectedPath: "age",
			expectedVal:  "21",
		},
		{
			name:         "attribute equality",
			expr:         "@id==5",
			expectedOp:   OpEqual,
			expectedPath: "@id",
			expectedVal:  "5",
		},
		{
			name:         "string with spaces",
			expr:         "name=='John Doe'",
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
			name:         "not equal",
			expr:         "status!=pending",
			expectedOp:   OpNotEqual,
			expectedPath: "status",
			expectedVal:  "pending",
		},
		{
			name:         "less than or equal",
			expr:         "count<=100",
			expectedOp:   OpLessThanOrEqual,
			expectedPath: "count",
			expectedVal:  "100",
		},
	}

	// Use the shared helper function from filter_test.go
	testFilterConditionParsing(t, tests)
}

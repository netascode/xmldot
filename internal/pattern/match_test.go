// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package pattern

import (
	"strings"
	"testing"
)

// TestMatchBasicWildcards tests basic wildcard matching
func TestMatchBasicWildcards(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		pattern  string
		expected bool
	}{
		// Star wildcard tests
		{"star matches empty", "hello", "hello*", true},
		{"star matches prefix", "hello world", "hello*", true},
		{"star matches suffix", "hello world", "*world", true},
		{"star matches middle", "hello world", "hello*world", true},
		{"star matches all", "anything", "*", true},
		{"star empty string", "", "*", true},
		{"multiple stars", "abc123xyz", "a*1*z", true},

		// Question mark wildcard tests
		{"question mark single char", "hello", "hell?", true},
		{"question mark exact count", "Dale", "D?le", true},
		{"multiple question marks", "Dale", "D??e", true},
		{"question mark wrong count", "Dale", "D???e", false},

		// Exact matches
		{"exact match", "hello", "hello", true},
		{"exact no match", "hello", "world", false},
		{"partial match fails", "hello", "hel", false},

		// Empty patterns
		{"empty matches empty", "", "", true},
		{"empty pattern no match", "hello", "", false},

		// Case sensitivity
		{"case sensitive match", "Hello", "Hello", true},
		{"case sensitive no match", "Hello", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, stopped := Match(tt.str, tt.pattern, 10000)
			if stopped {
				t.Error("Unexpected complexity limit")
			}
			if matched != tt.expected {
				t.Errorf("Match(%q, %q) = %v, expected %v", tt.str, tt.pattern, matched, tt.expected)
			}
		})
	}
}

// TestMatchEscapeSequences tests escape sequence handling
func TestMatchEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		pattern  string
		expected bool
	}{
		{"escaped asterisk matches literal", "file*txt", `file\*txt`, true},
		{"escaped asterisk no match", "filetxt", `file\*txt`, false},
		{"escaped question mark matches literal", "file?doc", `file\?doc`, true},
		{"escaped question mark no match", "filedoc", `file\?doc`, false},
		{"escaped backslash", `file\txt`, `file\\txt`, true},
		{"multiple escapes", `a*b?c\d`, `a\*b\?c\\d`, true},
		{"escaped at end", "test*", `test\*`, true},
		{"unescaped vs escaped", "test123", "test*", true},
		{"unescaped matches literal star", "test*", "test*", true},
		{"wildcard no match", "testxyz", `test\*`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, stopped := Match(tt.str, tt.pattern, 10000)
			if stopped {
				t.Error("Unexpected complexity limit")
			}
			if matched != tt.expected {
				t.Errorf("Match(%q, %q) = %v, expected %v", tt.str, tt.pattern, matched, tt.expected)
			}
		})
	}
}

// TestMatchUnicode tests Unicode support
func TestMatchUnicode(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		pattern  string
		expected bool
	}{
		{"unicode exact match", "你好世界", "你好世界", true},
		{"unicode prefix wildcard", "你好世界", "你好*", true},
		{"unicode suffix wildcard", "你好世界", "*世界", true},
		{"unicode middle wildcard", "你好世界", "你*界", true},
		{"unicode question mark", "你好世界", "你?世界", true},
		{"unicode multiple question marks", "你好世界", "你??界", true},
		{"unicode mixed with ascii", "Hello世界", "Hello*", true},
		{"emoji support", "Hello 👋 World", "Hello*World", true},
		{"emoji question mark", "👋", "?", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, stopped := Match(tt.str, tt.pattern, 10000)
			if stopped {
				t.Error("Unexpected complexity limit")
			}
			if matched != tt.expected {
				t.Errorf("Match(%q, %q) = %v, expected %v", tt.str, tt.pattern, matched, tt.expected)
			}
		})
	}
}

// TestMatchComplexPatterns tests complex pattern matching scenarios
func TestMatchComplexPatterns(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		pattern  string
		expected bool
	}{
		{"multiple stars", "abc123xyz", "a*1*z", true},
		{"mixed wildcards", "test.txt", "test?txt", true},
		{"multiple segments", "user:admin:name", "user:*:name", true},
		{"consecutive stars", "test", "t**t", true},
		{"star question mix", "hello", "h*?o", true},
		{"complex pattern", "abc123def456ghi", "a*3*6*i", true},
		{"pattern no match", "abc123", "a*z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, stopped := Match(tt.str, tt.pattern, 10000)
			if stopped {
				t.Error("Unexpected complexity limit")
			}
			if matched != tt.expected {
				t.Errorf("Match(%q, %q) = %v, expected %v", tt.str, tt.pattern, matched, tt.expected)
			}
		})
	}
}

// TestMatchReDoSProtection tests ReDoS protection
func TestMatchReDoSProtection(t *testing.T) {
	tests := []struct {
		name        string
		str         string
		pattern     string
		maxIter     int
		expectStop  bool
		expectMatch bool
	}{
		{
			name:        "pathological pattern exceeds limit",
			str:         strings.Repeat("a", 100),
			pattern:     "a*a*a*a*a*a*a*a*a*a*a*a*a*a*a*b",
			maxIter:     1000,
			expectStop:  true,
			expectMatch: false,
		},
		{
			name:        "simple pattern within limit",
			str:         "hello world",
			pattern:     "hello*",
			maxIter:     100,
			expectStop:  false,
			expectMatch: true,
		},
		{
			name:        "complex pattern within limit",
			str:         strings.Repeat("a", 50),
			pattern:     "a*a*a*b",
			maxIter:     10000,
			expectStop:  false,
			expectMatch: false,
		},
		{
			name:        "long string short pattern ok",
			str:         strings.Repeat("a", 1000),
			pattern:     "*a",
			maxIter:     10000,
			expectStop:  false,
			expectMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, stopped := Match(tt.str, tt.pattern, tt.maxIter)
			if stopped != tt.expectStop {
				t.Errorf("stopped = %v, expected %v", stopped, tt.expectStop)
			}
			if matched != tt.expectMatch {
				t.Errorf("matched = %v, expected %v", matched, tt.expectMatch)
			}
		})
	}
}

// TestMatchEdgeCases tests edge cases
func TestMatchEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		pattern  string
		expected bool
	}{
		// Empty strings
		{"both empty", "", "", true},
		{"empty str with star", "", "*", true},
		{"empty str with question", "", "?", false},
		{"empty pattern", "hello", "", false},

		// Only wildcards
		{"only stars", "anything", "***", true},
		{"only questions", "abc", "???", true},
		{"only questions wrong count", "abc", "????", false},

		// Special characters
		{"dots in string", "file.txt", "file*txt", true},
		{"colons in string", "a:b:c", "a:*:c", true},
		{"spaces in string", "hello world", "hello*world", true},

		// Numeric strings
		{"numeric exact", "12345", "12345", true},
		{"numeric wildcard", "12345", "1*5", true},

		// Leading/trailing wildcards
		{"leading stars", "hello", "***hello", true},
		{"trailing stars", "hello", "hello***", true},
		{"both ends stars", "hello", "***hello***", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, stopped := Match(tt.str, tt.pattern, 10000)
			if stopped {
				t.Error("Unexpected complexity limit")
			}
			if matched != tt.expected {
				t.Errorf("Match(%q, %q) = %v, expected %v", tt.str, tt.pattern, matched, tt.expected)
			}
		})
	}
}

// TestProcessEscapes tests the escape sequence processor
func TestProcessEscapes(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedPattern string
		expectedEscapes []bool
	}{
		{
			name:            "no escapes",
			input:           "hello*world",
			expectedPattern: "hello*world",
			expectedEscapes: []bool{false, false, false, false, false, false, false, false, false, false, false},
		},
		{
			name:            "escaped asterisk",
			input:           `file\*txt`,
			expectedPattern: "file*txt",
			expectedEscapes: []bool{false, false, false, false, true, false, false, false},
		},
		{
			name:            "escaped question mark",
			input:           `file\?doc`,
			expectedPattern: "file?doc",
			expectedEscapes: []bool{false, false, false, false, true, false, false, false},
		},
		{
			name:            "escaped backslash",
			input:           `file\\txt`,
			expectedPattern: `file\txt`,
			expectedEscapes: []bool{false, false, false, false, true, false, false, false},
		},
		{
			name:            "multiple escapes",
			input:           `a\*b\?c`,
			expectedPattern: "a*b?c",
			expectedEscapes: []bool{false, true, false, true, false},
		},
		{
			name:            "trailing backslash ignored",
			input:           `test\`,
			expectedPattern: `test\`,
			expectedEscapes: []bool{false, false, false, false, false},
		},
		{
			name:            "empty",
			input:           "",
			expectedPattern: "",
			expectedEscapes: []bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, escapes := processEscapes([]rune(tt.input))
			resultPattern := string(pattern)

			if resultPattern != tt.expectedPattern {
				t.Errorf("pattern = %q, expected %q", resultPattern, tt.expectedPattern)
			}

			if len(escapes) != len(tt.expectedEscapes) {
				t.Errorf("escapes length = %d, expected %d", len(escapes), len(tt.expectedEscapes))
				return
			}

			for i, esc := range escapes {
				if esc != tt.expectedEscapes[i] {
					t.Errorf("escapes[%d] = %v, expected %v", i, esc, tt.expectedEscapes[i])
				}
			}
		})
	}
}

// BenchmarkMatch benchmarks pattern matching performance
func BenchmarkMatch(b *testing.B) {
	benchmarks := []struct {
		name    string
		str     string
		pattern string
	}{
		{"simple prefix", "hello world", "hello*"},
		{"simple suffix", "hello world", "*world"},
		{"exact match", "hello", "hello"},
		{"multiple wildcards", "abc123xyz", "a*1*z"},
		{"long string", strings.Repeat("a", 1000) + "b", "*b"},
		{"unicode", "你好世界", "你好*"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Match(bm.str, bm.pattern, 10000)
			}
		})
	}
}

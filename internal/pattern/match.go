// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

// Package pattern provides wildcard pattern matching with ReDoS protection.
//
// This package implements secure wildcard matching using dynamic programming
// with configurable iteration limits to prevent Regular Expression Denial
// of Service (ReDoS) attacks from pathological patterns.
package pattern

// Match performs wildcard pattern matching with iteration limiting.
//
// Supports:
//   - '*' matches zero or more characters
//   - '?' matches exactly one character
//   - '\' escapes the next character (\*, \?, \\)
//
// Returns:
//
//	matched: true if pattern matches string
//	stopped: true if complexity limit (maxIterations) was exceeded
//
// When stopped=true, matched will be false. This prevents ReDoS attacks
// from pathological patterns like "a*a*a*a*b" on long strings.
//
// Algorithm: Uses iterative dynamic programming (DP) approach.
// DP table: dp[i][j] = whether str[0:i] matches pattern[0:j]
//
// Time complexity: O(m*n) where m=len(str), n=len(pattern)
// Space complexity: O(m*n) for DP table
//
// Example usage:
//
//	matched, stopped := Match("hello world", "hello*", 10000)
//	if stopped {
//	    // Complexity limit exceeded
//	}
//	if matched {
//	    // Pattern matched
//	}
func Match(str, pattern string, maxIterations int) (matched, stopped bool) {
	// Convert to runes for proper Unicode support
	s := []rune(str)
	p := []rune(pattern)

	sLen := len(s)

	// Handle escape sequences in pattern
	p, escapes := processEscapes(p)
	pLen := len(p)

	// DP table: dp[i][j] = whether s[0:i] matches p[0:j]
	dp := make([][]bool, sLen+1)
	for i := range dp {
		dp[i] = make([]bool, pLen+1)
	}

	// Empty pattern matches empty string
	dp[0][0] = true

	// Leading stars match empty string
	// Example: "***hello" should match "hello"
	for j := 1; j <= pLen; j++ {
		if p[j-1] == '*' && !escapes[j-1] {
			dp[0][j] = dp[0][j-1]
		}
	}

	iterations := 0

	// Fill DP table
	for i := 1; i <= sLen; i++ {
		for j := 1; j <= pLen; j++ {
			iterations++
			if iterations > maxIterations {
				// Security: Complexity limit exceeded, stop processing
				return false, true
			}

			pChar := p[j-1]
			escaped := escapes[j-1]

			if pChar == '*' && !escaped {
				// Star matches zero or more chars
				// dp[i][j-1]: star matches zero chars (skip star)
				// dp[i-1][j]: star matches one or more chars (consume char from string)
				dp[i][j] = dp[i][j-1] || dp[i-1][j]
			} else if pChar == '?' && !escaped {
				// Question mark matches exactly one char
				dp[i][j] = dp[i-1][j-1]
			} else if pChar == s[i-1] {
				// Exact match (including escaped wildcards treated as literals)
				dp[i][j] = dp[i-1][j-1]
			}
			// else: characters don't match, dp[i][j] remains false
		}
	}

	return dp[sLen][pLen], false
}

// processEscapes handles escape sequences in pattern.
// Returns modified pattern and boolean slice indicating escaped positions.
//
// Escape sequences:
//   - \* → literal asterisk
//   - \? → literal question mark
//   - \\ → literal backslash
//
// Example:
//
//	input:  "file\*txt"
//	output: "file*txt", [false, false, false, false, true, false, false, false]
func processEscapes(pattern []rune) ([]rune, []bool) {
	result := make([]rune, 0, len(pattern))
	escaped := make([]bool, 0, len(pattern))

	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			// Next char is escaped - treat it as literal
			i++
			result = append(result, pattern[i])
			escaped = append(escaped, true)
		} else {
			// Normal character
			result = append(result, pattern[i])
			escaped = append(escaped, false)
		}
	}

	return result, escaped
}

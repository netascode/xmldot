// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

// Package pattern provides wildcard pattern matching with bounded complexity.
//
// This package implements wildcard matching with a linear-space, iterative
// algorithm and a configurable step limit, so that pathological patterns can
// neither exhaust memory nor run unbounded.
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
// When stopped=true, matched will be false and the caller must treat the
// result as undetermined.
//
// Algorithm: Greedy matching that backtracks only to the most recent '*'.
// This is correct for '*' and '?' globbing because an earlier '*' never needs
// to consume more input once a later '*' has matched.
//
// Time complexity: O(m*n) worst case where m=len(str), n=len(pattern), O(m+n) typical.
// A trailing '*' ends matching immediately.
// Space complexity: O(m+n) for the rune slices
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
	p, escapes := processEscapes([]rune(pattern))

	isStar := func(i int) bool { return p[i] == '*' && !escapes[i] }

	// Start of the trailing run of stars, which matches any remaining input
	trailingStars := len(p)
	for trailingStars > 0 && isStar(trailingStars-1) {
		trailingStars--
	}

	si, pi := 0, 0
	starPi, starSi := -1, 0
	iterations := 0

	for si < len(s) {
		iterations++
		if iterations > maxIterations {
			return false, true
		}

		switch {
		case pi < len(p) && !isStar(pi) && ((p[pi] == '?' && !escapes[pi]) || p[pi] == s[si]):
			// Literal or '?' match: advance both
			si++
			pi++
		case pi < len(p) && isStar(pi):
			// A trailing star matches the rest of the input
			if pi >= trailingStars {
				return true, false
			}
			// Remember the star position and initially match zero characters
			starPi, starSi = pi, si
			pi++
		case starPi != -1:
			// Mismatch: let the most recent star consume one more character
			starSi++
			si = starSi
			pi = starPi + 1
		default:
			return false, false
		}
	}

	// Remaining pattern must consist only of stars
	for pi < len(p) && isStar(pi) {
		pi++
	}

	return pi == len(p), false
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

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"math"
	"strconv"
	"strings"

	"github.com/netascode/xmldot/internal/pattern"
)

const (
	// MaxFilterExpressionLength is the maximum allowed length for filter expressions.
	// This prevents DoS attacks with extremely long filter strings.
	MaxFilterExpressionLength = 256

	// MaxFilterDepth is the maximum recursion depth for filter evaluation.
	// This prevents stack overflow attacks with deeply nested filter paths.
	MaxFilterDepth = 10

	// MaxPatternIterations is the maximum complexity for pattern matching operations.
	// This prevents ReDoS (Regular Expression Denial of Service) attacks with
	// exponential backtracking patterns. Value of 10000 matches GJSON's default.
	MaxPatternIterations = 10000
)

// FilterOp represents a filter comparison operator.
type FilterOp int

const (
	// OpEqual represents the == operator.
	OpEqual FilterOp = iota
	// OpNotEqual represents the != operator.
	OpNotEqual
	// OpLessThan represents the < operator.
	OpLessThan
	// OpGreaterThan represents the > operator.
	OpGreaterThan
	// OpLessThanOrEqual represents the <= operator.
	OpLessThanOrEqual
	// OpGreaterThanOrEqual represents the >= operator.
	OpGreaterThanOrEqual
	// OpPatternMatch represents the % operator (GJSON-style pattern matching).
	OpPatternMatch
	// OpPatternNotMatch represents the !% operator (GJSON-style negated pattern matching).
	OpPatternNotMatch
	// OpExists checks if an attribute/element exists.
	OpExists
)

// parseFilter parses a filter expression like "[age>21]" into a Filter.
// Supported operators: ==, !=, <, >, <=, >=, %, !%
// Supported operands: element paths, attribute paths (@attr), numeric values, string values
//
// Examples:
//   - "age>21" → {Path: "age", Op: OpGreaterThan, Value: "21"}
//   - "@id==5" → {Path: "@id", Op: OpEqual, Value: "5"}
//   - "name=='John'" → {Path: "name", Op: OpEqual, Value: "John"}
//   - "@active" → {Path: "@active", Op: OpExists, Value: ""}
//   - "name%'*Go*'" → {Path: "name", Op: OpPatternMatch, Value: "*Go*"}
//   - "status!%'temp*'" → {Path: "status", Op: OpPatternNotMatch, Value: "temp*"}
//
// Security: Expressions longer than MaxFilterExpressionLength are rejected.
// Security: Null bytes and operator characters in paths are rejected.
//
// DEPRECATED: This function supports legacy bracket syntax. Use parseFilterCondition instead.
func parseFilter(expr string) (*Filter, error) {
	// Security check: limit filter expression length
	if len(expr) > MaxFilterExpressionLength {
		return nil, ErrInvalidPath
	}

	// Security check: reject null bytes in expression
	if strings.ContainsRune(expr, '\x00') {
		return nil, ErrInvalidPath
	}

	// Remove brackets if present
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "[") && strings.HasSuffix(expr, "]") {
		expr = expr[1 : len(expr)-1]
	}
	expr = strings.TrimSpace(expr)

	return parseFilterCondition(expr)
}

// parseFilterCondition parses a filter condition (without brackets) into a Filter.
// This is the GJSON-style filter parser that doesn't expect bracket markers.
// Supported operators: ==, !=, <, >, <=, >=, %, !%
// Supported operands: element paths, attribute paths (@attr), numeric values, string values
//
// Examples:
//   - "age>21" → {Path: "age", Op: OpGreaterThan, Value: "21"}
//   - "@id==5" → {Path: "@id", Op: OpEqual, Value: "5"}
//   - "name=='John'" → {Path: "name", Op: OpEqual, Value: "John"}
//   - "@active" → {Path: "@active", Op: OpExists, Value: ""}
//   - "name%'*Go*'" → {Path: "name", Op: OpPatternMatch, Value: "*Go*"}
//   - "status!%'temp*'" → {Path: "status", Op: OpPatternNotMatch, Value: "temp*"}
//
// Security: Expressions longer than MaxFilterExpressionLength are rejected.
// Security: Null bytes and operator characters in paths are rejected.
func parseFilterCondition(expr string) (*Filter, error) {
	// Security check: limit filter expression length
	if len(expr) > MaxFilterExpressionLength {
		return nil, ErrInvalidPath
	}

	// Security check: reject null bytes in expression
	if strings.ContainsRune(expr, '\x00') {
		return nil, ErrInvalidPath
	}

	expr = strings.TrimSpace(expr)

	if expr == "" {
		return nil, ErrInvalidPath
	}

	// Check for existence filter (just a path with no operator)
	// e.g., [@active] or [name]
	if !strings.ContainsAny(expr, "=!<>%") {
		// Security check: validate path doesn't contain null bytes
		if strings.ContainsRune(expr, '\x00') {
			return nil, ErrInvalidPath
		}
		return &Filter{
			Path:  expr,
			Op:    OpExists,
			Value: "",
		}, nil
	}

	// Find the operator
	var op FilterOp
	var opStr string
	var opPos = -1

	// Check for two-character operators first (==, <=, >=, !=, !%)
	for i := 0; i < len(expr)-1; i++ {
		twoChar := expr[i : i+2]
		switch twoChar {
		case "==":
			op = OpEqual
			opStr = "=="
			opPos = i
		case "<=":
			op = OpLessThanOrEqual
			opStr = "<="
			opPos = i
		case ">=":
			op = OpGreaterThanOrEqual
			opStr = ">="
			opPos = i
		case "!=":
			op = OpNotEqual
			opStr = "!="
			opPos = i
		case "!%":
			op = OpPatternNotMatch
			opStr = "!%"
			opPos = i
		}
		if opPos >= 0 {
			break
		}
	}

	// If no two-character operator found, check for single-character operators
	if opPos < 0 {
		for i := 0; i < len(expr); i++ {
			c := expr[i]
			switch c {
			case '<':
				op = OpLessThan
				opStr = "<"
				opPos = i
			case '>':
				op = OpGreaterThan
				opStr = ">"
				opPos = i
			case '%':
				op = OpPatternMatch
				opStr = "%"
				opPos = i
			}
			if opPos >= 0 {
				break
			}
		}
	}

	if opPos < 0 {
		return nil, ErrInvalidPath
	}

	// Extract path and value (before trimming)
	pathRaw := expr[:opPos]
	valueRaw := expr[opPos+len(opStr):]

	// Security check: validate path doesn't contain control characters BEFORE trimming
	// This prevents log injection attacks where control characters are hidden by TrimSpace
	if strings.ContainsAny(pathRaw, "\x00\n\r\t") {
		return nil, ErrInvalidPath
	}

	// Now safe to trim
	path := strings.TrimSpace(pathRaw)
	value := strings.TrimSpace(valueRaw)

	if path == "" || value == "" {
		return nil, ErrInvalidPath
	}

	// Security check: validate path doesn't contain operator characters
	// Path should only contain element names, dots, and @ for attributes
	if strings.ContainsAny(path, "=!<>%") {
		return nil, ErrInvalidPath
	}

	// Remove quotes from string values
	if (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) ||
		(strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) {
		value = value[1 : len(value)-1]
	}

	// Security check: validate value doesn't contain control characters AFTER quote removal
	if strings.ContainsAny(value, "\x00\n\r\t") {
		return nil, ErrInvalidPath
	}

	return &Filter{
		Path:  path,
		Op:    op,
		Value: value,
	}, nil
}

// evaluateFilterWithDepth evaluates a filter with recursion depth tracking.
// Optimized: Fast paths for common filter patterns to avoid parsing overhead.
func evaluateFilterWithDepth(filter *Filter, content string, attrs map[string]string, depth int) bool {
	if filter == nil {
		return true
	}

	// Security check: enforce maximum filter recursion depth
	if depth >= MaxFilterDepth {
		return false
	}

	// Get the value to compare
	var actualValue string
	var exists bool

	if strings.HasPrefix(filter.Path, "@") {
		// Fast path: Attribute filter - direct map lookup, no parsing
		attrName := filter.Path[1:]
		actualValue, exists = attrs[attrName]
	} else {
		// Element filter - extract text from specific child element
		parser := newXMLParser([]byte(content))
		parser.filterDepth = depth + 1
		result := executeQuery(parser, parsePath(filter.Path), 0)
		exists = result.Exists()
		actualValue = result.String()
	}

	// Handle existence check (fast path)
	if filter.Op == OpExists {
		return exists
	}

	// If value doesn't exist, filter doesn't match
	if !exists {
		return false
	}

	// Perform comparison based on operator
	switch filter.Op {
	case OpEqual:
		// Fast path: Direct string comparison
		return actualValue == filter.Value

	case OpNotEqual:
		// Fast path: Direct string inequality
		return actualValue != filter.Value

	case OpLessThan, OpGreaterThan, OpLessThanOrEqual, OpGreaterThanOrEqual:
		// Numeric operators ONLY work with valid numbers
		// Fast path: Check if values are numeric before parsing
		if !isNumericValue(actualValue) || !isNumericValue(filter.Value) {
			return false
		}

		actualNum, actualErr := strconv.ParseFloat(actualValue, 64)
		filterNum, filterErr := strconv.ParseFloat(filter.Value, 64)

		if actualErr == nil && filterErr == nil {
			// Security check: detect special float values (Inf, NaN) and reject
			if math.IsInf(actualNum, 0) || math.IsNaN(actualNum) || math.IsInf(filterNum, 0) || math.IsNaN(filterNum) {
				return false
			}

			// Numeric comparison
			switch filter.Op {
			case OpLessThan:
				return actualNum < filterNum
			case OpGreaterThan:
				return actualNum > filterNum
			case OpLessThanOrEqual:
				return actualNum <= filterNum
			case OpGreaterThanOrEqual:
				return actualNum >= filterNum
			}
		}

		// For numeric operators, ONLY accept valid numbers
		// Return false instead of confusing string comparison
		return false

	case OpPatternMatch, OpPatternNotMatch:
		// Pattern matching operators use string matching with wildcards
		// * matches any sequence of characters (zero or more)
		// ? matches exactly one character
		// \ escapes the next character

		// Fast path: if pattern contains no wildcards, use simple string comparison
		patternStr := filter.Value
		if !strings.ContainsAny(patternStr, "*?\\") {
			matched := actualValue == patternStr
			if filter.Op == OpPatternMatch {
				return matched
			}
			return !matched
		}

		// Use internal pattern matcher with complexity limiting for DoS protection
		matched, stopped := pattern.Match(actualValue, patternStr, MaxPatternIterations)

		// Security: If complexity limit exceeded, treat as no match.
		// Returning false (instead of an error) is intentional - this prevents
		// attackers from using pathological patterns to cause DoS by triggering
		// expensive error handling. The pattern simply doesn't match.
		if stopped {
			return false
		}

		if filter.Op == OpPatternMatch {
			return matched
		}
		return !matched
	}

	return false
}

// evaluateFilterOnMatch evaluates a filter against an elementMatch.
func evaluateFilterOnMatch(filter *Filter, match elementMatch) bool {
	return evaluateFilterWithDepth(filter, match.content, match.attrs, 0)
}

// isNumericValue checks if a string contains a valid numeric value (int or float).
// Fast path helper to avoid expensive ParseFloat on non-numeric strings.
func isNumericValue(s string) bool {
	if len(s) == 0 {
		return false
	}
	hasDigit := false
	hasDot := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if c == '.' {
			if hasDot {
				return false // Multiple dots
			}
			hasDot = true
		} else if c == '-' || c == '+' {
			if i != 0 {
				return false // Sign must be at start
			}
		} else if c == 'e' || c == 'E' {
			// Scientific notation - just return true and let ParseFloat handle it
			return hasDigit
		} else {
			return false
		}
	}
	return hasDigit
}

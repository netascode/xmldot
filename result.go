// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"strconv"
	"strings"
)

// Type represents the type of a Result value.
type Type int

const (
	// Null represents a non-existent path.
	Null Type = iota
	// String represents a text value.
	String
	// Number represents a numeric value.
	Number
	// True represents a boolean true value.
	True
	// False represents a boolean false value.
	False
	// Element represents an XML element.
	Element
	// Attribute represents an XML attribute.
	Attribute
	// Array represents multiple XML elements with the same name.
	Array
)

// Result represents the result of a Get operation. It contains the matched
// value and provides methods for type conversion.
type Result struct {
	// Type is the type of the result value.
	Type Type
	// Raw is the raw XML segment that was matched.
	Raw string
	// Str is the parsed string value.
	Str string
	// Index is the array index if this result is part of an array.
	Index int
	// Num is the cached numeric value if the result is a number.
	Num float64
	// Results holds child results for Array type (Phase 3+)
	Results []Result
}

// Exists returns true if the result represents an existing value in the XML.
func (r Result) Exists() bool {
	return r.Type != Null
}

// String returns the string representation of the result.
// For Null types, it returns an empty string.
// For Array types, it returns a JSON-like array representation.
// This implements the fmt.Stringer interface.
func (r Result) String() string {
	if r.Type == Null {
		return ""
	}
	if r.Type == Array {
		// Return JSON-like array representation: ["item1","item2",...]
		if len(r.Results) == 0 {
			return "[]"
		}
		var sb strings.Builder
		sb.WriteString("[")
		for i, item := range r.Results {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(`"`)
			// Escape quotes and backslashes in the value
			itemStr := item.String()
			for _, ch := range itemStr {
				switch ch {
				case '"':
					sb.WriteString(`\"`)
				case '\\':
					sb.WriteString(`\\`)
				default:
					sb.WriteRune(ch)
				}
			}
			sb.WriteString(`"`)
		}
		sb.WriteString("]")
		return sb.String()
	}
	return r.Str
}

// Int returns the result as an int64. If the result cannot be converted,
// it returns 0.
func (r Result) Int() int64 {
	switch r.Type {
	case Number:
		return int64(r.Num)
	case String, Element, Attribute:
		// Try parsing the string value
		if val, err := parseInt64(r.Str); err == nil {
			return val
		}
	case True:
		return 1
	case False:
		return 0
	}
	return 0
}

// Float returns the result as a float64. If the result cannot be converted,
// it returns 0.
func (r Result) Float() float64 {
	switch r.Type {
	case Number:
		return r.Num
	case String, Element, Attribute:
		// Try parsing the string value
		if val, err := parseFloat64(r.Str); err == nil {
			return val
		}
	case True:
		return 1
	case False:
		return 0
	}
	return 0
}

// Bool returns the result as a bool. The strings "true", "1", "yes" are
// considered true. Everything else is false.
func (r Result) Bool() bool {
	switch r.Type {
	case True:
		return true
	case False:
		return false
	case Number:
		return r.Num != 0
	case String, Element, Attribute:
		s := r.Str
		return s == "true" || s == "1" || s == "yes" || s == "True" || s == "YES" || s == "t" || s == "T"
	}
	return false
}

// Value returns the result as an interface{} with the appropriate Go type.
func (r Result) Value() interface{} {
	switch r.Type {
	case Null:
		return nil
	case True:
		return true
	case False:
		return false
	case Number:
		return r.Num
	case String, Element, Attribute:
		return r.Str
	case Array:
		// Return slice of Result values
		return r.Results
	}
	return nil
}

// IsArray returns true if the Result represents an array (multiple elements).
func (r Result) IsArray() bool {
	return r.Type == Array
}

// Array returns the Result as a slice of Results for array types.
// For non-array types, returns a single-element slice containing the result.
func (r Result) Array() []Result {
	if r.Type == Array {
		return r.Results
	}
	if r.Type == Null {
		return []Result{}
	}
	return []Result{r}
}

// ForEach iterates over array elements, calling the iterator function for each.
// The iterator receives the index and value. Return false to stop iteration.
// For non-array types, the iterator is called once with index 0.
func (r Result) ForEach(iterator func(index int, value Result) bool) {
	if r.Type == Array {
		for i, result := range r.Results {
			if !iterator(i, result) {
				return
			}
		}
	} else if r.Type != Null {
		iterator(0, r)
	}
}

// Helper functions for type conversion

// parseInt64 parses a string to int64, handling various formats
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// parseFloat64 parses a string to float64
func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// itoa converts int to string
func itoa(i int) string {
	return strconv.Itoa(i)
}

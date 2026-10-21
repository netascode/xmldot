// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"errors"
	"fmt"
)

// Security limits to protect against malicious XML and pathological queries.
//
// The limits are variables so that applications processing trusted documents can
// adjust them. They are read without synchronization, so set them once during
// program initialization, before any concurrent use of the package.
//
// When a limit is exceeded, the operation fails instead of returning a partial
// result: Get and related functions return a Null Result whose Err field holds a
// *LimitError, and Set, Delete and related functions return the *LimitError.
// Use errors.Is(err, ErrLimitExceeded) to detect any limit violation and
// errors.As to inspect which limit was hit.
var (
	// MaxDocumentSize is the maximum allowed size for an XML document in bytes.
	MaxDocumentSize = 100 * 1024 * 1024

	// MaxValueSize is the maximum allowed size in bytes for a single value in Set operations.
	MaxValueSize = 100 * 1024 * 1024

	// MaxNestingDepth is the maximum allowed nesting depth for XML elements.
	// This prevents stack exhaustion with deeply nested structures.
	MaxNestingDepth = 100

	// MaxAttributes is the maximum number of attributes allowed per element.
	MaxAttributes = 1000

	// MaxTokenSize is the maximum size in bytes for a single element name, attribute
	// name, attribute value, closing tag or processing instruction.
	MaxTokenSize = 1024 * 1024

	// MaxNamespacePrefixLength is the maximum allowed length for namespace prefixes in paths.
	MaxNamespacePrefixLength = 256

	// MaxArrayResults is the maximum number of sibling elements collected by
	// non-recursive array operations: count (#), index (.N), field extraction (#.field),
	// filters, single-level wildcards, fragment roots and Map(). These scan a single
	// level of the document, so their cost is linear and already bounded by
	// MaxDocumentSize. The limit only guards against per-match memory overhead.
	MaxArrayResults = 100000

	// MaxWildcardResults is the maximum number of results from recursive wildcard
	// queries ("**").
	MaxWildcardResults = 100000

	// MaxRecursiveOperations is the minimum operation budget for a recursive wildcard
	// query ("**"). The effective budget is the larger of this value and the size of
	// the searched document in bytes, so a single recursive search over any document
	// always completes, while chained recursive searches cannot grow without bound.
	MaxRecursiveOperations = 10000

	// MaxFilterExpressionLength is the maximum allowed length for filter expressions.
	MaxFilterExpressionLength = 4096

	// MaxPatternIterations is the step budget for a single pattern match (% and !%
	// filter operators), on top of one step per byte of the matched value. Typical
	// patterns match in linear time and always complete; only pathological
	// value/pattern combinations, which take up to len(value)*len(pattern) steps,
	// can exceed it.
	MaxPatternIterations = 10000000

	// MaxModifierChainDepth is the maximum number of chained modifiers in a path.
	MaxModifierChainDepth = 20

	// MaxPathSegments is the maximum number of path segments allowed in a query path.
	MaxPathSegments = 100

	// MaxFieldNameLength is the maximum length of a field name in #.field syntax.
	MaxFieldNameLength = 256
)

// ErrLimitExceeded is the sentinel error wrapped by every *LimitError.
var ErrLimitExceeded = errors.New("security limit exceeded")

// LimitError reports that an operation exceeded one of the package security limits.
type LimitError struct {
	// Limit is the name of the exceeded limit variable, e.g. "MaxDocumentSize".
	Limit string
	// Value is the value observed when the check failed. Checks that count while
	// scanning stop at the first value over the limit, so Value is typically Max+1
	// rather than the total size of the input.
	Value int
	// Max is the effective limit at the time of the check. For MaxRecursiveOperations
	// and MaxPatternIterations this is the budget derived from the limit and the input.
	Max int
}

// Error implements the error interface.
func (e *LimitError) Error() string {
	return fmt.Sprintf("%s: %s is %d, observed %d", ErrLimitExceeded, e.Limit, e.Max, e.Value)
}

// Unwrap returns ErrLimitExceeded so that errors.Is(err, ErrLimitExceeded) reports true.
func (e *LimitError) Unwrap() error {
	return ErrLimitExceeded
}

// limitPanic carries a *LimitError from deep inside the parser or query engine
// to the nearest public entry point, where it is recovered and returned.
type limitPanic struct {
	err *LimitError
}

// exceedLimit aborts the current operation with a *LimitError. It must only be
// called below a public entry point that defers one of the recover helpers
// (recoverLimitResult, recoverLimitBytes or recoverMapLimit).
func exceedLimit(limit string, value, maxValue int) {
	panic(limitPanic{err: &LimitError{Limit: limit, Value: value, Max: maxValue}})
}

// recoverLimitResult converts a limit violation raised by exceedLimit into a
// Null Result carrying the error. Usage: defer recoverLimitResult(&result).
// Other panics are re-raised unchanged.
func recoverLimitResult(result *Result) {
	if r := recover(); r != nil {
		lp, ok := r.(limitPanic)
		if !ok {
			panic(r)
		}
		*result = Result{Type: Null, Err: lp.err}
	}
}

// recoverLimitBytes converts a limit violation into an error, returning the original
// document unchanged. Usage: defer recoverLimitBytes(&out, &err, xml).
// Other panics are re-raised unchanged.
func recoverLimitBytes(out *[]byte, err *error, orig []byte) {
	if r := recover(); r != nil {
		lp, ok := r.(limitPanic)
		if !ok {
			panic(r)
		}
		*out = orig
		*err = lp.err
	}
}

// recoverMapLimit converts a limit violation into a nil map.
// Usage: defer recoverMapLimit(&result). Other panics are re-raised unchanged.
func recoverMapLimit(result *map[string]Result) {
	if r := recover(); r != nil {
		if _, ok := r.(limitPanic); !ok {
			panic(r)
		}
		*result = nil
	}
}

// checkDocumentSize returns a *LimitError if a document exceeds MaxDocumentSize.
func checkDocumentSize(size int) error {
	if size > MaxDocumentSize {
		return &LimitError{Limit: "MaxDocumentSize", Value: size, Max: MaxDocumentSize}
	}
	return nil
}

// checkValueSize returns a *LimitError if a value exceeds MaxValueSize.
func checkValueSize(size int) error {
	if size > MaxValueSize {
		return &LimitError{Limit: "MaxValueSize", Value: size, Max: MaxValueSize}
	}
	return nil
}

// checkArrayResults aborts if adding one more match would exceed MaxArrayResults.
func checkArrayResults(current int) {
	if current >= MaxArrayResults {
		exceedLimit("MaxArrayResults", current+1, MaxArrayResults)
	}
}

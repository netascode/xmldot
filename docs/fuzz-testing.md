# Fuzz Testing for XMLDOT

This document describes the comprehensive fuzz testing suite for the XMLDOT library.

## Overview

Fuzz testing is a software testing technique that involves providing invalid, unexpected, or random data as inputs to a program. The XMLDOT library includes an extensive fuzz testing suite designed to find crashes, panics, and edge cases in critical operations.

**Purpose**: Find robustness issues, not correctness bugs. Fuzz tests ensure the library never panics or crashes regardless of input.

## Fuzz Test Files

### 1. Parser Fuzzing (`parser_fuzz_test.go`)

Tests the core XML parser for robustness with arbitrary input.

**Tests**:
- `FuzzParser` - General parser robustness with malformed XML
- `FuzzParserWithPaths` - Parser with various path combinations
- `FuzzParserNesting` - Deeply nested structures (tests MaxNestingDepth)
- `FuzzParserAttributes` - Attribute parsing edge cases
- `FuzzParserSpecialContent` - CDATA, comments, processing instructions
- `FuzzParserLargeInput` - Size limits and security boundaries
- `FuzzParserEscaping` - XML entity escaping/unescaping
- `FuzzParserNamespaces` - Namespace prefix handling

**Key findings**: Parser is robust against malformed input and respects security limits.

### 2. Get Operation Fuzzing (`get_fuzz_test.go`)

Tests the Get operation for crashes with arbitrary XML and paths.

**Tests**:
- `FuzzGet` - General Get operation robustness
- `FuzzGetBytes` - Zero-copy byte slice parsing
- `FuzzGetWithFilters` - Filter expression parsing and evaluation
- `FuzzGetWithWildcards` - Wildcard matching (* and **)
- `FuzzGetWithArrayAccess` - Array indexing edge cases
- `FuzzGetWithAttributes` - Attribute access patterns
- `FuzzGetWithTextContent` - Text extraction (% operator)
- `FuzzGetWithNamespaces` - Namespace-prefixed elements
- `FuzzGetWithModifiers` - Modifier chains
- `FuzzGetComplexPaths` - Complex path combinations

**Key findings**: Get operation handles most edge cases gracefully. Minor type constant issues identified (non-critical).

### 3. Set Operation Fuzzing (`set_fuzz_test.go`)

Tests the Set operation for crashes with arbitrary modifications.

**Tests**:
- `FuzzSet` - General Set operation robustness
- `FuzzSetBytes` - Zero-copy modification
- `FuzzSetWithTypes` - Type conversion and escaping
- `FuzzSetRaw` - Raw XML insertion validation
- `FuzzSetAttributes` - Attribute setting
- `FuzzSetNested` - Deep path creation
- `FuzzSetArrays` - Array manipulation
- `FuzzSetDelete` - Delete operations (Set with nil)
- `FuzzSetMalformed` - Malformed XML handling
- `FuzzSetSizeLimit` - Size limit enforcement

**Key findings**:
- **CRASH FOUND**: Set panics on malformed XML `<root>` with slice bounds error
- See `testdata/fuzz/crash-set-malformed.txt` for details

### 4. Path Parser Fuzzing (`path_fuzz_test.go`)

Tests path parsing for crashes with arbitrary path strings.

**Tests**:
- `FuzzPathParser` - General path parsing robustness
- `FuzzPathSplitting` - Path splitting edge cases
- `FuzzPathWithFilters` - Filter expression syntax
- `FuzzPathWithModifiers` - Modifier syntax
- `FuzzPathWithArrayAccess` - Array index parsing
- `FuzzPathWithSpecialOperators` - Special operators (%, #, *, **)
- `FuzzPathWithAttributes` - Attribute access (@) syntax
- `FuzzPathWithNamespaces` - Namespace prefix syntax
- `FuzzPathCacheConsistency` - Cache consistency verification
- `FuzzPathSegmentLimit` - MaxPathSegments security limit
- `FuzzPathMatchWithOptions` - Case sensitivity and options
- `FuzzIsNumeric` - Numeric detection edge cases

**Key findings**: Path parser is robust and cache is consistent.

### 5. Validation Fuzzing (`validate_fuzz_test.go`)

Tests XML validation for crashes and consistency.

**Tests**:
- `FuzzValid` - Basic validation robustness
- `FuzzValidateWithError` - Detailed error reporting
- `FuzzValidateNesting` - Nested structure validation
- `FuzzValidateAttributes` - Attribute syntax validation
- `FuzzValidateSpecialContent` - Special content validation
- `FuzzValidateTagMatching` - Opening/closing tag matching
- `FuzzValidateSelfClosing` - Self-closing tag syntax
- `FuzzValidateNamespaces` - Namespace syntax validation
- `FuzzValidateWhitespace` - Whitespace handling
- `FuzzValidateEmptyAndMinimal` - Edge case inputs
- `FuzzValidateEscaping` - Entity reference validation
- `FuzzValidateSizeLimit` - Size limit enforcement
- `FuzzValidateMultiRoot` - Multiple root element detection
- `FuzzValidateConsistency` - Valid/ValidateWithError consistency

**Key findings**: Validation is robust and consistent between Valid() and ValidateWithError().

## Running Fuzz Tests

### Quick Validation (30 seconds per test)

Run a quick fuzz test to verify robustness:

```bash
go test -fuzz=FuzzParser -fuzztime=30s
go test -fuzz=FuzzGet -fuzztime=30s
go test -fuzz=FuzzSet -fuzztime=30s
go test -fuzz=FuzzPathParser -fuzztime=30s
go test -fuzz=FuzzValid -fuzztime=30s
```

### Extended Fuzzing (recommended for CI/CD)

Run longer fuzz tests to find deeper issues:

```bash
# 5 minutes per test
go test -fuzz=FuzzParser -fuzztime=5m
go test -fuzz=FuzzGet -fuzztime=5m
go test -fuzz=FuzzSet -fuzztime=5m
go test -fuzz=FuzzPathParser -fuzztime=5m
go test -fuzz=FuzzValid -fuzztime=5m
```

### Continuous Fuzzing (background)

Run fuzz tests continuously in the background:

```bash
# Run until crash found or manually stopped
go test -fuzz=FuzzParser -fuzztime=0
```

### Run All Fuzz Tests

```bash
# List all fuzz tests
go test -list 'Fuzz.*'

# Run all fuzz tests for 30s each (sequential)
for test in $(go test -list 'Fuzz.*' | grep Fuzz); do
    echo "Running $test..."
    go test -fuzz="$test" -fuzztime=30s
done
```

## Crash Corpus

Crash inputs are saved in `testdata/fuzz/` for regression testing and bug analysis.

**Current crashes**:
- `crash-set-malformed.xml` - Set operation panic with unclosed tag

**Bugs identified**:
- `bug-attribute-type.txt` - Attribute type constant inconsistency (non-critical)

## Interpreting Results

### Success
```
fuzz: elapsed: 30s, execs: 125432 (4181/sec), new interesting: 87 (3/sec)
PASS
```

### Crash Found
```
--- FAIL: FuzzSet (0.01s)
    --- FAIL: FuzzSet/4a8b3c2d1e0f (0.00s)
        set_fuzz_test.go:25: Set panicked: panic=runtime error: slice bounds out of range
```

When a crash is found:
1. The crashing input is saved to `testdata/fuzz/FuzzTestName/`
2. Create a regression test for the crash
3. Fix the underlying issue
4. Re-run the fuzz test to verify fix

## Fuzz Coverage

The fuzz test suite covers:

### Parser Coverage
- ✅ Malformed XML (unclosed tags, missing delimiters)
- ✅ Nested structures (up to MaxNestingDepth)
- ✅ Attributes (quoted, unquoted, malformed)
- ✅ Special content (CDATA, comments, PIs)
- ✅ Large inputs (size limits)
- ✅ Entity escaping
- ✅ Namespace prefixes

### Get Coverage
- ✅ Element access
- ✅ Attribute access
- ✅ Array indexing
- ✅ Filters
- ✅ Wildcards (* and **)
- ✅ Text content (%)
- ✅ Count (#)
- ✅ Modifiers
- ✅ Namespaces

### Set Coverage
- ✅ Element creation
- ✅ Element modification
- ✅ Attribute setting
- ✅ Raw XML insertion
- ✅ Nested path creation
- ✅ Array operations
- ✅ Delete operations
- ✅ Malformed input handling
- ✅ Size limits

### Path Coverage
- ✅ Path parsing
- ✅ Path splitting
- ✅ Filter expressions
- ✅ Modifiers
- ✅ Array access
- ✅ Special operators
- ✅ Attributes
- ✅ Namespaces
- ✅ Cache consistency
- ✅ Segment limits

### Validation Coverage
- ✅ Well-formedness checking
- ✅ Tag matching
- ✅ Nesting validation
- ✅ Attribute syntax
- ✅ Special content
- ✅ Self-closing tags
- ✅ Namespaces
- ✅ Whitespace handling
- ✅ Error reporting

## Known Limitations

Fuzz tests focus on **robustness** (no panics), not **correctness** (correct behavior):

1. **Type constants**: Minor inconsistencies in Result.Type values are not tested
2. **Semantic correctness**: Fuzz tests don't verify correct XML semantics
3. **Output validation**: Set results are validated for well-formedness but not correctness
4. **Edge case behavior**: Some edge cases may return unexpected results without panicking

For correctness testing, see the standard unit test suite (`*_test.go` files).

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Fuzz Tests

on:
  schedule:
    - cron: '0 0 * * *'  # Daily
  workflow_dispatch:

jobs:
  fuzz:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run fuzz tests
        run: |
          for test in $(go test -list 'Fuzz.*' | grep Fuzz); do
            echo "Fuzzing $test..."
            go test -fuzz="$test" -fuzztime=5m || exit 1
          done
```

## Regression Testing

After fixing crashes, add regression tests:

```go
func TestSetMalformedXMLRegression(t *testing.T) {
    // Regression test for crash found in FuzzSet
    xml := "<root>"  // Unclosed tag
    _, err := Set(xml, "root.item", "value")

    // Should return error, not panic
    if err == nil {
        t.Error("Expected error for malformed XML")
    }
}
```

## Best Practices

1. **Run fuzz tests regularly**: Include in CI/CD pipeline
2. **Save crash inputs**: Commit crash cases to `testdata/fuzz/`
3. **Create regression tests**: Add tests for each crash found
4. **Monitor coverage**: Track new interesting inputs to ensure coverage growth
5. **Focus on panics**: Fuzz tests prevent crashes, unit tests verify correctness

## Statistics (2025-10-08)

- **Total fuzz tests**: 49
- **Parser tests**: 8
- **Get tests**: 10
- **Set tests**: 10
- **Path tests**: 11
- **Validation tests**: 13
- **Crashes found**: 1 (Set with malformed XML)
- **Bugs identified**: 1 (Attribute type constant)
- **Test execution time**: ~30 seconds per test
- **Total seed corpus size**: 200+ inputs

## Future Improvements

1. **Extended fuzzing**: Run longer fuzz campaigns (hours/days)
2. **Coverage-guided fuzzing**: Monitor code coverage during fuzzing
3. **Mutation strategies**: Improve seed corpus for better coverage
4. **Parallel fuzzing**: Run multiple fuzz tests simultaneously
5. **Crash deduplication**: Identify unique vs. duplicate crashes

## References

- [Go Fuzzing Tutorial](https://go.dev/doc/tutorial/fuzz)
- [OSS-Fuzz Integration](https://github.com/google/oss-fuzz)
- [Fuzzing Best Practices](https://github.com/google/fuzzing/blob/master/docs/good-fuzz-target.md)

# Error Handling Guide

This guide covers error handling in XMLDOT, including error types, error detection, and best practices for robust XML processing.

## Table of Contents

- [Error Types](#error-types)
- [Error Detection](#error-detection)
- [Operation-Specific Errors](#operation-specific-errors)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Error Types

XMLDOT defines three sentinel error types that can be checked using `errors.Is()`:

### ErrMalformedXML

Returned when XML input is not well-formed.

**Common causes:**
- Unclosed tags: `<root><item>`
- Mismatched tags: `<root><item>value</wrong></root>`
- Invalid XML structure: `<<<>>>`
- Document exceeds `MaxDocumentSize` (10MB)
- Empty documents

**Example:**
```go
xml := "<root><unclosed>"
_, err := Set(xml, "root.item", "value")
if errors.Is(err, ErrMalformedXML) {
    fmt.Println("XML is malformed")
}
```

### ErrInvalidPath

Returned when path syntax is invalid or unsupported.

**Common causes:**
- Empty paths: `""`
- Invalid SetMany/DeleteMany parameters (mismatched array lengths)

**Example:**
```go
_, err := Set("<root/>", "", "value")
if errors.Is(err, ErrInvalidPath) {
    fmt.Println("Path is invalid")
}
```

### ErrInvalidValue

Returned when a value cannot be converted to XML or is inappropriate for the operation.

**Common causes:**
- Raw XML with unclosed tags in `SetRaw()`
- Raw XML with mismatched tags
- DOCTYPE declarations in `SetRaw()` (security)
- ENTITY declarations in `SetRaw()` (security)
- Nested CDATA sections in `SetRaw()`

**Example:**
```go
_, err := SetRaw("<root/>", "root.data", "<unclosed>")
if errors.Is(err, ErrInvalidValue) {
    fmt.Println("Raw XML value is invalid")
}
```

## Error Detection

### Set and Delete Operations Return Errors

Operations that modify XML return `(string, error)`:

```go
result, err := Set(xml, path, value)
if err != nil {
    // Handle error
    // Original XML is returned unchanged
    fmt.Println("Error:", err)
    return
}
// Use result...
```

### Get Operations Never Return Errors

`Get` operations return a `Result` instead of an error. Check `Exists()` to determine if the path was found:

```go
result := Get(xml, "root.item")
if !result.Exists() {
    // Path not found or XML is malformed
    fmt.Println("Path not found")
}
```

**Why Get doesn't return errors:**
- Follows gjson/sjson pattern for simplicity
- Missing paths are not errors - they return empty results
- Malformed XML returns empty results gracefully
- No panic on invalid input

### Validation Functions

Use validation functions to check XML well-formedness before processing:

```go
// Quick validation
if !Valid(xml) {
    return errors.New("invalid XML")
}

// Detailed error reporting
if err := ValidateWithError(xml); err != nil {
    fmt.Printf("Invalid XML at line %d, column %d: %s\n",
        err.Line, err.Column, err.Message)
    return
}
```

## Operation-Specific Errors

### Set Operations

```go
// Returns ErrMalformedXML
_, err := Set("<root>", "root.item", "value")

// Returns ErrInvalidPath
_, err := Set("<root/>", "", "value")

// Returns original XML on error
result, err := Set("<root>", "root.item", "value")
if err != nil {
    // result == "<root>" (unchanged)
}
```

### Delete Operations

```go
// Returns ErrMalformedXML
_, err := Delete("<root>", "root.item")

// Returns ErrInvalidPath
_, err := Delete("<root/>", "")

// Non-existent paths do NOT error
_, err := Delete("<root/>", "root.nonexistent")
// err == nil, returns unchanged XML
```

### Batch Operations

```go
// SetMany returns ErrInvalidPath on length mismatch
_, err := SetMany("<root/>",
    []string{"a", "b"},
    []interface{}{"value"}) // Mismatched lengths

// SetMany validates ALL paths before applying changes
// On error, original XML is returned unchanged
_, err := SetMany("<root/>",
    []string{"root.a", "", "root.c"}, // Invalid middle path
    []interface{}{1, 2, 3})
// Returns ErrInvalidPath, XML unchanged

// DeleteMany succeeds even if some paths don't exist
_, err := DeleteMany("<root><a>1</a></root>",
    "root.a", "root.nonexistent", "root.alsomissing")
// err == nil
```

### SetRaw Security Validation

`SetRaw` validates raw XML to prevent security issues:

```go
// Rejects DOCTYPE (XXE protection)
_, err := SetRaw("<root/>", "root.data",
    "<!DOCTYPE test><test/>")
// Returns ErrInvalidValue

// Rejects ENTITY declarations
_, err := SetRaw("<root/>", "root.data",
    "<!ENTITY test 'value'><test/>")
// Returns ErrInvalidValue

// Rejects nested CDATA
_, err := SetRaw("<root/>", "root.data",
    "<![CDATA[<![CDATA[nested]]>]]>")
// Returns ErrInvalidValue

// Rejects unbalanced tags
_, err := SetRaw("<root/>", "root.data", "<unclosed>")
// Returns ErrInvalidValue
```

## Best Practices

### 1. Always Check Errors from Set/Delete

```go
// Bad
result, _ := Set(xml, path, value)

// Good
result, err := Set(xml, path, value)
if err != nil {
    return fmt.Errorf("failed to set value: %w", err)
}
```

### 2. Use Validation Before Expensive Operations

```go
// Validate before processing large documents
if !Valid(largeXML) {
    return errors.New("invalid XML document")
}

// Then proceed with operations
for _, item := range items {
    result, err := Set(result, item.Path, item.Value)
    if err != nil {
        return err
    }
}
```

### 3. Check Result.Exists() for Get Operations

```go
result := Get(xml, "root.optional.field")
if !result.Exists() {
    // Use default value
    value = defaultValue
} else {
    value = result.String()
}
```

### 4. Use errors.Is() for Error Type Checking

```go
_, err := Set(xml, path, value)
if err != nil {
    if errors.Is(err, ErrMalformedXML) {
        // Handle malformed XML specifically
        log.Error("XML validation failed")
    } else if errors.Is(err, ErrInvalidPath) {
        // Handle invalid path
        log.Error("Path syntax error")
    }
    return err
}
```

### 5. Handle Original XML on Error

```go
result, err := Set(xml, path, value)
if err != nil {
    // result contains original XML unchanged
    log.Printf("Set failed, XML unchanged: %v", err)
    // Can safely use result as original document
    return result, err
}
```

### 6. Batch Operations for Efficiency

```go
// Instead of multiple Set calls that might partially fail
paths := []string{"root.a", "root.b", "root.c"}
values := []interface{}{1, 2, 3}

result, err := SetMany(xml, paths, values)
if err != nil {
    // No partial changes - original XML returned
    log.Error("Batch set failed:", err)
    return xml, err
}
```

## Examples

### Complete Error Handling Example

```go
package main

import (
    "errors"
    "fmt"
    "log"

    "github.com/netascode/xmldot"
)

func updateXML(xml string, updates map[string]interface{}) (string, error) {
    // 1. Validate XML first
    if !xmldot.Valid(xml) {
        return xml, fmt.Errorf("invalid XML document")
    }

    // 2. Prepare batch operations
    paths := make([]string, 0, len(updates))
    values := make([]interface{}, 0, len(updates))

    for path, value := range updates {
        paths = append(paths, path)
        values = append(values, value)
    }

    // 3. Apply changes
    result, err := xmldot.SetMany(xml, paths, values)
    if err != nil {
        // 4. Handle specific error types
        if errors.Is(err, xmldot.ErrMalformedXML) {
            return xml, fmt.Errorf("XML malformed: %w", err)
        }
        if errors.Is(err, xmldot.ErrInvalidPath) {
            return xml, fmt.Errorf("invalid path in updates: %w", err)
        }
        return xml, fmt.Errorf("update failed: %w", err)
    }

    // 5. Validate result
    if !xmldot.Valid(result) {
        return xml, fmt.Errorf("result validation failed")
    }

    return result, nil
}

func main() {
    xml := `<config><server><host>localhost</host><port>8080</port></server></config>`

    updates := map[string]interface{}{
        "config.server.host": "example.com",
        "config.server.port": 443,
        "config.server.tls":  true,
    }

    result, err := updateXML(xml, updates)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Updated XML:", result)
}
```

### Defensive Get with Error Checking

```go
func getConfigValue(xml, path, defaultValue string) string {
    // Validate XML first
    if !xmldot.Valid(xml) {
        log.Printf("Invalid XML, using default: %s", defaultValue)
        return defaultValue
    }

    // Get value
    result := xmldot.Get(xml, path)
    if !result.Exists() {
        log.Printf("Path %s not found, using default: %s", path, defaultValue)
        return defaultValue
    }

    return result.String()
}
```

### Safe Raw XML Insertion

```go
func insertRawXML(xml, path, rawXML string) (string, error) {
    // Validate input XML
    if !xmldot.Valid(xml) {
        return xml, xmldot.ErrMalformedXML
    }

    // SetRaw validates raw XML for security
    result, err := xmldot.SetRaw(xml, path, rawXML)
    if err != nil {
        if errors.Is(err, xmldot.ErrInvalidValue) {
            return xml, fmt.Errorf("raw XML validation failed: %w", err)
        }
        return xml, err
    }

    return result, nil
}
```

## Error Recovery

### Graceful Degradation

```go
func safeGet(xml, path string) string {
    // Get never panics, even with malformed XML
    result := xmldot.Get(xml, path)
    if !result.Exists() {
        return ""
    }
    return result.String()
}
```

### Transaction-like Batch Operations

```go
func atomicUpdates(xml string, operations []Operation) (string, error) {
    // Collect all changes
    paths := make([]string, len(operations))
    values := make([]interface{}, len(operations))

    for i, op := range operations {
        paths[i] = op.Path
        values[i] = op.Value
    }

    // Apply all at once - succeeds completely or fails completely
    result, err := xmldot.SetMany(xml, paths, values)
    if err != nil {
        // No partial changes - original XML unchanged
        log.Printf("Atomic update failed: %v", err)
        return xml, err
    }

    return result, nil
}
```

## No Panics Guarantee

All XMLDOT operations are designed to never panic on invalid input:

```go
// None of these will panic
_ = xmldot.Get("<<<invalid>>>", "root.item")
_ = xmldot.Get("<root>", "")
_ = xmldot.Get(strings.Repeat("x", 100*1024*1024), "root.item")

// All return errors gracefully
_, _ = xmldot.Set("<root>", "root.item", "value")
_, _ = xmldot.Set("<root/>", "", "value")
_, _ = xmldot.Delete("invalid", "path")
```

All operations handle:
- Malformed XML gracefully
- Invalid paths without panicking
- Document size limits safely
- Concurrent access without races
- Invalid UTF-8 and control characters

## Security Considerations

Error handling includes security protections:

1. **Document Size Limits**: Documents >10MB return `ErrMalformedXML`
2. **Nesting Depth Limits**: Depth >100 levels is truncated/rejected
3. **SetRaw Validation**: Rejects DOCTYPE, ENTITY, nested CDATA
4. **No XXE Attacks**: External entities never processed
5. **No Code Injection**: All values are properly escaped

When security limits are exceeded:
```go
// Document too large
largeXML := strings.Repeat("<item>x</item>", 1000000)
_, err := xmldot.Set(largeXML, "root.new", "value")
// Returns ErrMalformedXML

// Validation provides detailed error
err := xmldot.ValidateWithError(largeXML)
// err.Message contains "maximum size" information
```

## Summary

### Key Points

1. **Set/Delete return errors** - always check them
2. **Get never returns errors** - check `Result.Exists()` instead
3. **Errors return original XML** - unchanged on failure
4. **Use errors.Is()** - for type checking
5. **Validate early** - before expensive operations
6. **No panics** - all operations handle invalid input gracefully
7. **Security built-in** - SetRaw validates for XXE and injection

### Quick Reference

| Operation | Returns Error | Original XML on Error | Notes |
|-----------|---------------|----------------------|-------|
| `Get()` | No | N/A | Returns empty `Result` on error |
| `Set()` | Yes | Yes | `ErrMalformedXML`, `ErrInvalidPath` |
| `Delete()` | Yes | Yes | Non-existent paths don't error |
| `SetMany()` | Yes | Yes | All-or-nothing |
| `DeleteMany()` | Yes | Yes | Non-existent paths skipped |
| `SetRaw()` | Yes | Yes | `ErrInvalidValue` for security issues |
| `Valid()` | No | N/A | Returns bool |
| `ValidateWithError()` | Yes | N/A | Returns `*ValidateError` |

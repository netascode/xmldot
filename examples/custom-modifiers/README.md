# Custom Modifiers Example

This example demonstrates how to create and use custom modifiers with xmldot to transform query results in powerful ways.

## What Are Custom Modifiers?

Modifiers are transformations applied to query results after path resolution. xmldot includes built-in modifiers like `@sort`, `@reverse`, `@first`, and `@last`, but you can also create your own custom modifiers to extend functionality for your specific use cases.

## Quick Start

Run the example:

```bash
cd examples/custom-modifiers
go run main.go
```

## Basic Example: Uppercase Modifier

The simplest custom modifier converts text to uppercase:

```go
// uppercaseModifier converts Result string to uppercase
type uppercaseModifier struct{}

func (m *uppercaseModifier) Name() string {
    return "uppercase"
}

func (m *uppercaseModifier) Apply(r xmldot.Result) xmldot.Result {
    if r.Type == xmldot.Null {
        return r
    }

    return xmldot.Result{
        Type:    r.Type,
        Str:     strings.ToUpper(r.Str),
        Raw:     r.Raw,
        Num:     r.Num,
        Index:   r.Index,
        Results: r.Results,
    }
}

func init() {
    xmldot.RegisterModifier("uppercase", &uppercaseModifier{})
}
```

Usage:

```go
result := xmldot.Get(xml, "book.title|@first|@uppercase")
fmt.Println(result.String()) // "THE GO PROGRAMMING LANGUAGE"
```

## Advanced Examples

### Count Modifier (Returns Number)

Count array elements and return a numeric Result:

```go
type countModifier struct{}

func (m *countModifier) Name() string {
    return "count"
}

func (m *countModifier) Apply(r xmldot.Result) xmldot.Result {
    if r.Type == xmldot.Null {
        return xmldot.Result{Type: xmldot.Number, Num: 0, Str: "0"}
    }

    if r.Type == xmldot.Array {
        count := len(r.Results)
        return xmldot.Result{
            Type: xmldot.Number,
            Num:  float64(count),
            Str:  fmt.Sprintf("%d", count),
        }
    }

    return xmldot.Result{Type: xmldot.Number, Num: 1, Str: "1"}
}
```

Usage:

```go
result := xmldot.Get(xml, "catalog.books.book|@count")
fmt.Printf("Total books: %d\n", result.Int()) // "Total books: 3"
```

### Join Modifier (Array to String)

Combine array elements with a separator:

```go
type joinModifier struct{}

func (m *joinModifier) Name() string {
    return "join"
}

func (m *joinModifier) Apply(r xmldot.Result) xmldot.Result {
    if r.Type != xmldot.Array {
        return r
    }

    var parts []string
    for _, elem := range r.Results {
        parts = append(parts, elem.String())
    }

    return xmldot.Result{
        Type: xmldot.String,
        Str:  strings.Join(parts, ", "),
    }
}
```

Usage:

```go
result := xmldot.Get(xml, "catalog.books.book.title|@join")
fmt.Println(result.String())
// "The Go Programming Language, Learning Go, Concurrency in Go"
```

### Lowercase Modifier (Array-Aware)

Apply transformations to all array elements:

```go
type lowercaseModifier struct{}

func (m *lowercaseModifier) Name() string {
    return "lowercase"
}

func (m *lowercaseModifier) Apply(r xmldot.Result) xmldot.Result {
    if r.Type == xmldot.Null {
        return r
    }

    // Handle arrays: apply to all elements
    if r.Type == xmldot.Array {
        results := make([]xmldot.Result, len(r.Results))
        for i, elem := range r.Results {
            results[i] = xmldot.Result{
                Type:  elem.Type,
                Str:   strings.ToLower(elem.Str),
                Raw:   elem.Raw,
                Num:   elem.Num,
                Index: elem.Index,
            }
        }
        return xmldot.Result{Type: xmldot.Array, Results: results}
    }

    // Single element
    return xmldot.Result{
        Type: r.Type,
        Str:  strings.ToLower(r.Str),
        Raw:  r.Raw,
        Num:  r.Num,
    }
}
```

## API Reference

### Modifier Interface

All custom modifiers must implement the `Modifier` interface:

```go
type Modifier interface {
    // Apply transforms the input Result and returns a new Result
    Apply(r Result) Result

    // Name returns the modifier name (without @ prefix)
    Name() string
}
```

### Registration Functions

```go
// RegisterModifier registers a custom modifier globally
func RegisterModifier(name string, m Modifier) error

// UnregisterModifier removes a custom modifier
// Built-in modifiers (@reverse, @sort, etc.) cannot be unregistered
func UnregisterModifier(name string) error

// GetModifier retrieves a registered modifier by name
func GetModifier(name string) Modifier
```

### ModifierFunc Adapter

For simple modifiers, use the `ModifierFunc` adapter:

```go
upperFunc := xmldot.ModifierFunc(func(r xmldot.Result) xmldot.Result {
    return xmldot.Result{
        Type: r.Type,
        Str:  strings.ToUpper(r.Str),
        Raw:  r.Raw,
    }
})

xmldot.RegisterModifier("upper", upperFunc)
```

## Best Practices

### 1. Return Null on Invalid Input

Don't panic - return a Null Result for invalid input:

```go
func (m *myModifier) Apply(r xmldot.Result) xmldot.Result {
    if r.Type == xmldot.Null {
        return r // Gracefully handle Null
    }

    if !canProcess(r) {
        return xmldot.Result{Type: xmldot.Null} // Return Null for invalid input
    }

    // Process result...
}
```

### 2. Preserve Result Metadata

Keep Raw, Num, and Index fields when possible:

```go
return xmldot.Result{
    Type:    newType,
    Str:     newString,
    Raw:     r.Raw,     // Preserve original XML
    Num:     r.Num,     // Preserve numeric value
    Index:   r.Index,   // Preserve array index
    Results: r.Results, // Preserve child results
}
```

### 3. Use Copy-On-Write Semantics

Never modify the input Result - always return a new Result:

```go
// DON'T: Modifying input
func (m *badModifier) Apply(r xmldot.Result) xmldot.Result {
    r.Str = strings.ToUpper(r.Str) // WRONG: Mutates input
    return r
}

// DO: Return new Result
func (m *goodModifier) Apply(r xmldot.Result) xmldot.Result {
    return xmldot.Result{
        Type: r.Type,
        Str:  strings.ToUpper(r.Str), // New string, input unchanged
        Raw:  r.Raw,
    }
}
```

### 4. Handle Both Single Results and Arrays

Consider whether your modifier should work on arrays, single elements, or both:

```go
func (m *myModifier) Apply(r xmldot.Result) xmldot.Result {
    // Option 1: Apply to arrays element-by-element
    if r.Type == xmldot.Array {
        results := make([]xmldot.Result, len(r.Results))
        for i, elem := range r.Results {
            results[i] = m.applyToElement(elem)
        }
        return xmldot.Result{Type: xmldot.Array, Results: results}
    }

    // Option 2: Transform array into single value (like @count, @join)
    if r.Type == xmldot.Array {
        return m.aggregateArray(r)
    }

    // Single element processing
    return m.applyToElement(r)
}
```

### 5. Consider Performance

Modifiers are called frequently - optimize hot paths:

```go
// Pre-allocate slices when creating arrays
results := make([]xmldot.Result, len(r.Results))

// Use strings.Builder for string concatenation
var sb strings.Builder
sb.Grow(estimatedSize)
for _, elem := range r.Results {
    sb.WriteString(elem.String())
}
return xmldot.Result{Type: xmldot.String, Str: sb.String()}
```

## Security Considerations

### Custom Modifiers Run in Your Process

Custom modifiers execute in your application's process space with full privileges:

- Validate all input to prevent DoS attacks (e.g., infinite loops, excessive memory)
- Be cautious with modifiers that make network calls or access files
- Test error paths thoroughly

### Respect MaxModifierChainDepth

The framework limits modifier chains to 20 by default to prevent stack overflow:

```go
const MaxModifierChainDepth = 20
```

Design modifiers to be composable but efficient. Avoid modifiers that internally chain many operations.

### Built-In Modifiers Cannot Be Unregistered

Built-in modifiers (`@reverse`, `@sort`, `@first`, `@last`, `@flatten`, `@pretty`, `@ugly`) are protected from unregistration to ensure API stability.

## Testing Custom Modifiers

Write comprehensive tests for your modifiers:

```go
func TestUppercaseModifier(t *testing.T) {
    mod := &uppercaseModifier{}

    // Test single element
    input := xmldot.Result{Type: xmldot.String, Str: "hello"}
    result := mod.Apply(input)
    if result.Str != "HELLO" {
        t.Errorf("Expected HELLO, got %s", result.Str)
    }

    // Test Null handling
    input = xmldot.Result{Type: xmldot.Null}
    result = mod.Apply(input)
    if result.Type != xmldot.Null {
        t.Error("Should preserve Null type")
    }

    // Test metadata preservation
    input = xmldot.Result{
        Type: xmldot.Element,
        Str:  "text",
        Raw:  "<elem>text</elem>",
        Num:  42,
    }
    result = mod.Apply(input)
    if result.Raw != input.Raw || result.Num != input.Num {
        t.Error("Should preserve metadata")
    }
}
```

## Performance Tips

### Benchmark Your Modifiers

Use Go's benchmarking tools to measure performance:

```go
func BenchmarkUppercaseModifier(b *testing.B) {
    mod := &uppercaseModifier{}
    input := xmldot.Result{Type: xmldot.String, Str: "hello world"}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mod.Apply(input)
    }
}
```

### Optimize for Common Cases

Add fast paths for common scenarios:

```go
func (m *myModifier) Apply(r xmldot.Result) xmldot.Result {
    // Fast path: empty or Null
    if r.Type == xmldot.Null || r.Str == "" {
        return r
    }

    // Fast path: already processed
    if alreadyTransformed(r) {
        return r
    }

    // Slow path: full transformation
    return transform(r)
}
```

### Pre-Allocate Memory

Avoid repeated allocations in loops:

```go
// Pre-allocate result slice
results := make([]xmldot.Result, 0, len(r.Results))

// Pre-size string builder
var sb strings.Builder
sb.Grow(estimatedSize)
```

## Chaining Modifiers

Custom modifiers work seamlessly with built-in modifiers:

```go
// Chain custom and built-in modifiers
result := xmldot.Get(xml, "books.book.title|@sort|@reverse|@first|@uppercase")

// Combine multiple custom modifiers
result := xmldot.Get(xml, "items.item|@sort|@lowercase|@join")
```

Modifiers execute left-to-right (pipeline order):

```
Query result → @sort → @reverse → @first → @uppercase → Final result
```

## Example Output

Running the example produces:

```
Example 1: Uppercase first title
Result: THE GO PROGRAMMING LANGUAGE

Example 2: Lowercase all titles
  - the go programming language
  - learning go
  - concurrency in go

Example 3: Count books
Total books: 3

Example 4: Join titles
All titles: The Go Programming Language, Learning Go, Concurrency in Go

Example 5: Chain modifiers
Last book (sorted): THE GO PROGRAMMING LANGUAGE
```

## Further Reading

- [Main README](../../README.md) - Complete documentation
- [Built-in Modifiers](../../README.md#modifiers) - Reference for built-in modifiers
- [Path Syntax](../../README.md#path-syntax) - Query path documentation
- [API Documentation](https://pkg.go.dev/github.com/netascode/xmldot) - Full API reference

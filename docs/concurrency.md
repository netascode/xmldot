# Concurrency Safety Guide

## Overview

The XMLDOT library is designed with concurrency safety in mind. This document describes thread-safety guarantees, safe usage patterns, and operations that require external synchronization.

## Thread-Safe Operations

### ✅ Safe for Concurrent Use

The following operations are **completely thread-safe** and can be called from multiple goroutines without any external synchronization:

#### 1. Read Operations

All read operations are safe for concurrent use:

- **`Get(xml, path)`** - Query XML documents
- **`GetBytes(xml, path)`** - Zero-copy XML queries
- **`GetMany(xml, paths...)`** - Multiple path queries
- **`GetWithOptions(xml, path, opts)`** - Options-aware queries
- **`Valid(xml)`** - XML validation
- **`ValidBytes(xml)`** - Zero-copy validation

**Example: Concurrent Reads**
```go
var wg sync.WaitGroup
xml := loadXMLDocument()

for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        result := xmldot.Get(xml, fmt.Sprintf("users.user.%d.name", id))
        process(result)
    }(i)
}
wg.Wait()
```

#### 2. Result Methods

`Result` objects are **immutable** and safe to share across goroutines:

- `result.String()` - Get string value
- `result.Int()` - Get integer value
- `result.Float()` - Get float value
- `result.Bool()` - Get boolean value
- `result.Array()` - Get array of results
- `result.Raw()` - Get raw XML
- `result.Exists()` - Check existence
- `result.Index(i)` - Array indexing
- All other Result methods

**Example: Shared Result**
```go
xml := `<root><items><item>A</item><item>B</item><item>C</item></items></root>`
result := xmldot.Get(xml, "root.items.item")

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // Safe to access result concurrently
        _ = result.String()
        _ = result.Array()
        _ = result.Index(id % 3)
    }(i)
}
wg.Wait()
```

#### 3. Modifier Registry

The modifier registry is thread-safe for both registration and lookup:

- **`RegisterModifier(name, modifier)`** - Register custom modifiers (thread-safe)
- **`GetModifier(name)`** - Retrieve modifiers (thread-safe)
- **`UnregisterModifier(name)`** - Remove custom modifiers (thread-safe)
- **Modifier execution** - Registered modifiers can be called concurrently

**Example: Concurrent Modifier Registration**
```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()

        mod := xmldot.NewModifierFunc(fmt.Sprintf("custom%d", id), func(r xmldot.Result) xmldot.Result {
            return r
        })

        xmldot.RegisterModifier(fmt.Sprintf("custom%d", id), mod)
    }(i)
}
wg.Wait()
```

#### 4. Options Usage

`Options` instances can be safely used concurrently when passed to functions:

- **`GetWithOptions(xml, path, opts)`** - Thread-safe
- **`SetWithOptions(xml, path, value, opts)`** - Thread-safe (but see write safety below)
- **`DefaultOptions()`** - Thread-safe

**Note**: Options are not mutated by the library, so the same Options instance can be safely shared across goroutines.

**Example: Shared Options**
```go
xml := `<ROOT><Item>Value</Item></ROOT>`
opts := &xmldot.Options{CaseSensitive: false}

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // Safe: opts is not modified by GetWithOptions
        result := xmldot.GetWithOptions(xml, "root.item", opts)
        _ = result.String()
    }()
}
wg.Wait()
```

## NOT Thread-Safe Operations

### ❌ Requires External Synchronization

The following operations modify XML and are **NOT thread-safe** when operating on the same XML document:

#### 1. Write Operations

- **`Set(xml, path, value)`** - Modify XML
- **`SetBytes(xml, path, value)`** - Zero-copy modify
- **`SetMany(xml, paths, values)`** - Batch modifications
- **`SetWithOptions(xml, path, value, opts)`** - Options-aware modify
- **`SetRaw(xml, path, rawxml)`** - Raw XML insertion

#### 2. Delete Operations

- **`Delete(xml, path)`** - Remove elements/attributes
- **`DeleteBytes(xml, path)`** - Zero-copy delete
- **`DeleteMany(xml, paths...)`** - Batch deletions

### Why Write Operations Are Not Thread-Safe

Write operations return **new XML strings** rather than modifying in place. However, concurrent writes to track the "current" XML state requires synchronization.

**Unsafe Example** (Don't do this):
```go
// ❌ UNSAFE: Concurrent writes without synchronization
xml := "<root><counter>0</counter></root>"

for i := 0; i < 10; i++ {
    go func(id int) {
        // Data race: reading and writing xml concurrently
        xml, _ = xmldot.Set(xml, "root.counter", id)
    }(i)
}
```

## Safe Concurrent Patterns

### Pattern 1: Read-Only Concurrent Access

Multiple goroutines reading the same XML document:

```go
xml := loadXMLDocument()
var wg sync.WaitGroup

for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()

        // All reads are safe
        name := xmldot.Get(xml, fmt.Sprintf("users.user.%d.name", id))
        age := xmldot.Get(xml, fmt.Sprintf("users.user.%d.age", id))
        email := xmldot.Get(xml, fmt.Sprintf("users.user.%d.email", id))

        processUser(name, age, email)
    }(i)
}
wg.Wait()
```

### Pattern 2: Synchronized Writes with Mutex

Use `sync.Mutex` for exclusive write access:

```go
var mu sync.Mutex
currentXML := "<root><counter>0</counter></root>"

func updateXML(path string, value interface{}) error {
    mu.Lock()
    defer mu.Unlock()

    result, err := xmldot.Set(currentXML, path, value)
    if err != nil {
        return err
    }
    currentXML = result
    return nil
}

// Multiple goroutines can safely call updateXML
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        updateXML("root.counter", id)
    }(i)
}
wg.Wait()
```

### Pattern 3: Read-Write Lock for Mixed Operations

Use `sync.RWMutex` for concurrent reads with occasional writes:

```go
var mu sync.RWMutex
currentXML := `<root><counters><c1>0</c1><c2>0</c2></counters></root>`

// Many readers
func readCounter(id int) int {
    mu.RLock()
    xmlCopy := currentXML // Make a copy while holding read lock
    mu.RUnlock()

    // Safe to parse without lock (we have our own copy)
    result := xmldot.Get(xmlCopy, fmt.Sprintf("root.counters.c%d", id))
    return result.Int()
}

// Few writers
func updateCounter(id, value int) error {
    mu.Lock()
    defer mu.Unlock()

    result, err := xmldot.Set(currentXML, fmt.Sprintf("root.counters.c%d", id), value)
    if err != nil {
        return err
    }
    currentXML = result
    return nil
}

var wg sync.WaitGroup

// 100 concurrent readers
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        count := readCounter(id % 2 + 1)
        _ = count
    }(i)
}

// 10 concurrent writers
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        updateCounter(id%2+1, id)
    }(i)
}

wg.Wait()
```

### Pattern 4: Per-Goroutine XML Copies

Each goroutine works with its own copy of the XML:

```go
originalXML := "<root><counter>0</counter></root>"
results := make([]string, 100)
var wg sync.WaitGroup

for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()

        // Each goroutine has its own copy
        xmlCopy := originalXML

        // Safe: no shared state
        modified, _ := xmldot.Set(xmlCopy, "root.counter", id)
        results[id] = modified
    }(i)
}
wg.Wait()

// All results are independent and valid
for i, result := range results {
    counter := xmldot.Get(result, "root.counter")
    fmt.Printf("Result %d: counter=%d\n", i, counter.Int())
}
```

### Pattern 5: Channel-Based Sequential Processing

Use channels to serialize write operations:

```go
type writeRequest struct {
    path  string
    value interface{}
    resp  chan error
}

func startXMLWriter(initialXML string) chan writeRequest {
    requests := make(chan writeRequest)

    go func() {
        currentXML := initialXML
        for req := range requests {
            result, err := xmldot.Set(currentXML, req.path, req.value)
            if err == nil {
                currentXML = result
            }
            req.resp <- err
        }
    }()

    return requests
}

// Usage
xmlWriter := startXMLWriter("<root></root>")
defer close(xmlWriter)

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()

        resp := make(chan error, 1)
        xmlWriter <- writeRequest{
            path:  fmt.Sprintf("root.item%d", id),
            value: id,
            resp:  resp,
        }

        err := <-resp
        if err != nil {
            log.Printf("Write failed: %v", err)
        }
    }(i)
}
wg.Wait()
```

## Custom Modifier Thread-Safety

### Requirement: Modifiers Must Be Thread-Safe

Custom modifiers must be safe for concurrent execution. The same modifier instance may be called from multiple goroutines simultaneously.

**✅ Safe Modifier** (Stateless):
```go
uppercase := xmldot.NewModifierFunc("uppercase", func(r xmldot.Result) xmldot.Result {
    // Safe: No shared state, only transforms input
    return xmldot.Result{
        Type: r.Type,
        Str:  strings.ToUpper(r.Str),
        Raw:  r.Raw,
    }
})
xmldot.RegisterModifier("uppercase", uppercase)
```

**❌ Unsafe Modifier** (Mutable State):
```go
// DON'T DO THIS: Shared mutable state causes data races
counter := 0 // Shared state!

unsafe := xmldot.NewModifierFunc("counter", func(r xmldot.Result) xmldot.Result {
    counter++ // DATA RACE!
    return r
})
xmldot.RegisterModifier("counter", unsafe)
```

**✅ Safe Stateful Modifier** (Atomic Operations):
```go
import "sync/atomic"

var counter int64

safe := xmldot.NewModifierFunc("counter", func(r xmldot.Result) xmldot.Result {
    atomic.AddInt64(&counter, 1) // Safe: atomic operation
    return r
})
xmldot.RegisterModifier("counter", safe)
```

**✅ Safe Stateful Modifier** (Mutex):
```go
type StatefulModifier struct {
    mu      sync.Mutex
    counter int
}

func (m *StatefulModifier) Apply(r xmldot.Result) xmldot.Result {
    m.mu.Lock()
    m.counter++
    count := m.counter
    m.mu.Unlock()

    // Use count for transformation
    return r
}

func (m *StatefulModifier) Name() string { return "stateful" }

// Register
xmldot.RegisterModifier("stateful", &StatefulModifier{})
```

## Testing for Race Conditions

### Using Go's Race Detector

Run tests with the `-race` flag to detect data races:

```bash
# Test all packages
go test -race ./...

# Test specific concurrency tests
go test -race -run Concurrent

# Run with high iteration count
go test -race -count=100 -run Concurrent

# Test with benchmarks
go test -race -bench=. -benchtime=1s
```

### Using the Provided Race Test Script

The library includes a comprehensive race detection script:

```bash
# Run all race detection tests
./scripts/race-test.sh

# The script will:
# 1. Run all concurrency tests with -race
# 2. Run tests multiple times (count=10)
# 3. Run benchmarks with race detector
# 4. Report any detected race conditions
```

### Expected Behavior

- **No race conditions** should be detected when following safe patterns
- **Race conditions** will be detected in "unsafe" test cases (for documentation purposes)
- All tests marked as "safe" should pass with `-race`

## Performance Implications

### Read Operations

- **Zero overhead**: Read operations have no synchronization overhead
- **Highly parallel**: Scale linearly with CPU cores
- **No contention**: Each goroutine operates independently

### Write Operations with Synchronization

- **Mutex overhead**: Acquiring/releasing locks adds overhead
- **Serialization**: Writes are serialized by locks
- **Contention**: High write contention may reduce parallelism

### Recommendations

1. **Favor reads**: Design systems to maximize concurrent reads
2. **Batch writes**: Use `SetMany`/`DeleteMany` to reduce lock acquisitions
3. **Per-goroutine copies**: Avoid shared state when possible
4. **RWMutex for read-heavy**: Use `sync.RWMutex` when reads dominate

## Common Pitfalls

### ❌ Pitfall 1: Concurrent Writes to Shared Variable

```go
// DON'T DO THIS
xml := "<root></root>"

for i := 0; i < 10; i++ {
    go func(id int) {
        // DATA RACE: reading and writing xml concurrently
        xml, _ = xmldot.Set(xml, "item", id)
    }(i)
}
```

**Fix**: Use mutex or per-goroutine copies.

### ❌ Pitfall 2: Modifier Registration During Use

```go
// DON'T DO THIS
func handler(w http.ResponseWriter, r *http.Request) {
    // Registering modifiers in handlers causes contention
    xmldot.RegisterModifier("custom", myModifier)

    result := xmldot.Get(xml, "path|@custom")
    // ...
}
```

**Fix**: Register modifiers during initialization (init() or main()).

### ❌ Pitfall 3: Sharing Options Across Writes

```go
// DON'T DO THIS
opts := &xmldot.Options{CaseSensitive: false}
xml := "<root></root>"

for i := 0; i < 10; i++ {
    go func(id int) {
        // DATA RACE: xml is shared without synchronization
        xml, _ = xmldot.SetWithOptions(xml, "item", id, opts)
    }(i)
}
```

**Fix**: Options are safe to share, but xml writes need synchronization.

## Summary

| Operation | Thread-Safe? | Notes |
|-----------|-------------|-------|
| `Get()`, `GetBytes()`, `GetMany()` | ✅ Yes | Safe for concurrent reads |
| `GetWithOptions()` | ✅ Yes | Safe for concurrent reads |
| `Valid()`, `ValidBytes()` | ✅ Yes | Safe for concurrent validation |
| `Result` methods | ✅ Yes | Results are immutable |
| `RegisterModifier()` | ✅ Yes | Registry is thread-safe |
| `GetModifier()` | ✅ Yes | Lookup is thread-safe |
| Modifier execution | ✅ Yes | If modifier is thread-safe |
| `Options` usage | ✅ Yes | Options not mutated |
| `Set()`, `SetBytes()`, `SetMany()` | ❌ No | Requires synchronization |
| `Delete()`, `DeleteBytes()`, `DeleteMany()` | ❌ No | Requires synchronization |
| `SetWithOptions()` | ❌ No | Requires synchronization |

### Key Takeaways

1. **Read operations are fully thread-safe** - no synchronization needed
2. **Write operations require external synchronization** - use mutexes or per-goroutine copies
3. **Result objects are immutable** - safe to share across goroutines
4. **Modifier registry is thread-safe** - safe to register/use from multiple goroutines
5. **Custom modifiers must be thread-safe** - avoid mutable state or use synchronization
6. **Test with `-race`** - always test concurrent code with Go's race detector

## Additional Resources

- [Examples](../examples/) - See concurrency examples
- [Testing](../concurrency_test.go) - Concurrency test suite
- [Go Race Detector](https://go.dev/doc/articles/race_detector) - Official documentation
- [Go Concurrency Patterns](https://go.dev/blog/pipelines) - Advanced patterns

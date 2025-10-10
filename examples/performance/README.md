# Performance Optimization

**Complexity Level**: Advanced
**Estimated Time**: 25 minutes
**Prerequisites**: Understanding of all basic operations

## What You'll Learn

- Use batch operations (GetMany, SetMany) for better performance
- Enable Optimistic mode for trusted input
- Reduce allocations with byte-based APIs
- Understand automatic path caching benefits
- Choose efficient wildcard strategies
- Optimize filter usage
- Profile and benchmark your usage

## Running the Example

```bash
cd examples/performance
go run main.go
```

## Expected Output

```
Performance Optimization Example
=================================

Example 1: Baseline - Individual Get operations
100 iterations × 3 queries: [baseline time]

Example 2: GetMany for batch queries
100 iterations × 1 GetMany: [batch time]
Speedup: [X]x faster

Example 3: Optimistic mode for Set operations
Normal mode (100 Sets): [normal time]
Optimistic mode (100 Sets): [optimistic time]
Speedup: [X]x faster

Example 4: GetBytes/SetBytes for reduced allocations
String-based (1000 Gets): [string time]
Bytes-based (1000 Gets): [bytes time]
Allocation savings: ~[X]%

Example 5: Path caching benefits (automatic)
First 10000 queries: [cold time]
Repeat 10000 queries: [warm time]
Cache benefit: [X]% faster

Example 6: Wildcard performance comparison
Single-level wildcard: [single time]
Recursive wildcard: [recursive time]
Recursive overhead: [X]x slower

Example 7: Filter optimization
Numeric filter (100 queries): [numeric time]
String filter (100 queries): [string time]

Example 8: SetMany vs sequential Set operations
Sequential Set (5 operations): [sequential time]
Batch SetMany (5 operations): [batch time]
Speedup: [X]x faster

Example 9: Performance Optimization Checklist
✓ Use GetMany/SetMany for multiple paths
✓ Use Optimistic mode for trusted input (2-3x faster)
✓ Use GetBytes/SetBytes to reduce allocations
✓ Path caching is automatic (85-91% faster on repeat queries)
✓ Prefer single-level wildcards (*) over recursive (**)
✓ Use numeric filters when possible (fast path optimization)
✓ Profile your specific workload before optimizing
```

## Key Concepts

### Batch Operations

Process multiple paths in a single operation:
```go
// Instead of:
r1 := xmldot.Get(xml, "path1")
r2 := xmldot.Get(xml, "path2")
r3 := xmldot.Get(xml, "path3")

// Use:
results := xmldot.GetMany(xml, "path1", "path2", "path3")
```

Benefits: Single parse pass, reduced overhead

### Optimistic Mode

Skip validation for trusted input:
```go
xml, err := xmldot.SetWithOptions(xml, path, value,
    &xmldot.Options{Optimistic: true})
```

**Speedup**: 2-3x faster for Set operations
**Use when**: Input is trusted and well-formed
**Avoid when**: Processing untrusted user input

### Byte-Based APIs

Reduce string allocations:
```go
xmlBytes := []byte(xml)
result := xmldot.GetBytes(xmlBytes, path)
newXML, err := xmldot.SetBytes(xmlBytes, path, value)
```

Benefits: Fewer allocations, better GC performance

### Automatic Path Caching

xmldot caches parsed paths automatically using thread-safe LRU cache:
- **85-91% faster** on repeated queries
- **Zero configuration** required
- **Thread-safe** for concurrent access
- **LRU eviction** prevents unbounded memory growth

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Get (simple) | O(n) | Linear scan, incremental parsing |
| Get (wildcard *) | O(n) | May match multiple elements |
| Get (recursive **) | O(n²) | Expensive, avoid on large docs |
| Get (filter) | O(n) | Fast path for numeric comparisons |
| Set | O(n) | Incremental parsing + rebuild |
| Delete | O(n) | Incremental parsing + rebuild |

### Benchmark Results (Phase 4)

From Phase 4 optimization results:

**Path Parsing Cache**:
- Before: 2,500 ns/op
- After: 220-350 ns/op
- **Improvement: 85-91%**

**Filter Evaluation**:
- Before: 3,200 ns/op
- After: 2,500-2,700 ns/op
- **Improvement: 16-21%**

**Write Operations**:
- SetSimple: 7% faster
- SetComplex: 25% faster
- Delete: 13% faster

## Code Walkthrough

The example demonstrates practical optimization techniques:

1. **Baseline**: Measure individual Get operations
2. **Batch Queries**: Use GetMany for multiple paths
3. **Optimistic Mode**: Skip validation for trusted input
4. **Byte APIs**: Reduce allocations with GetBytes/SetBytes
5. **Path Cache**: Automatic caching on repeated queries
6. **Wildcard Cost**: Compare single-level vs recursive
7. **Filter Types**: Numeric vs string filter performance
8. **Batch Writes**: SetMany vs sequential Set calls
9. **Checklist**: Summary of optimization techniques

## Common Pitfalls

- **Pitfall**: Using Optimistic mode with untrusted input
  - **Solution**: Only use for validated/trusted XML sources

- **Pitfall**: Over-optimizing before profiling
  - **Solution**: Measure first, optimize where it matters

- **Pitfall**: Converting between string/bytes repeatedly
  - **Solution**: Pick one format and stick with it

- **Pitfall**: Using recursive wildcards unnecessarily
  - **Solution**: Use specific paths or single-level wildcards

## Optimization Strategies

### For Query-Heavy Workloads

1. Use GetMany for multiple paths
2. Rely on automatic path caching
3. Use GetBytes if already working with []byte
4. Prefer specific paths over wildcards

### For Write-Heavy Workloads

1. Use SetMany/DeleteMany for batch operations
2. Enable Optimistic mode for trusted sources
3. Use SetBytes to avoid string conversions
4. Consider pre-validating with Valid() once

### For Large Documents

1. Avoid recursive wildcards (**)
2. Use filters to narrow results early
3. Consider streaming alternatives (future Phase 9)
4. Profile memory usage and GC pressure

## Profiling Your Usage

Use Go's built-in profiling tools:

```go
import _ "net/http/pprof"
import "net/http"

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    // Your xmldot code here
}
```

Then profile:
```bash
go tool pprof http://localhost:6060/debug/pprof/profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Real-World Performance

Typical performance characteristics on modern hardware:

- **Simple Get**: 400-600 ns/op
- **Complex Get (filter)**: 1,200-2,500 ns/op
- **Set**: 2,000-4,000 ns/op
- **Large doc (1MB)**: 10-50ms depending on query

Compare with encoding/xml:
- xmldot: 2-5x faster for query operations
- encoding/xml: Better for full document parsing to structs

## Next Steps

- [Real-World Examples](../real-world/) - See optimizations in practice
- [Main README](../../README.md) - Complete documentation
- Benchmark your specific workload

## See Also

- [Phase 4 Results](../../project/PHASE4_RESULTS.md) - Detailed optimization data
- [Benchmark Tests](../../benchmark_test.go) - Full benchmark suite
- [API Reference](https://pkg.go.dev/github.com/netascode/xmldot) - Full API documentation

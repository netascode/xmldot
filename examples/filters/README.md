# Filter Operations

**Complexity Level**: Intermediate
**Estimated Time**: 20 minutes
**Prerequisites**: Understanding of arrays and basic queries

## What You'll Learn

- Filter elements using comparison operators
- Use numeric comparisons (>, <, >=, <=)
- Filter by string equality and inequality
- Filter by XML attributes
- Manual filtering for complex queries (chained filters not supported)
- Combine filters with wildcards and modifiers
- Handle empty filter results
- Count filtered elements

## Running the Example

```bash
cd examples/filters
go run main.go
```

## Expected Output

```
Filter Operations Example
=========================

Example 1: Employees older than 30
  - Bob
  - Carol
  - David

Example 2: Employees in Engineering
  - Alice
  - Carol
  - Eve

Example 3: Active employees
  - Alice
  - Bob
  - David
  - Eve

Example 4: Salary ranges
High earners (>= $85,000):
  - Alice
  - Carol
  - Eve

Low to mid earners (< $80,000):
  - Bob
  - David

Example 5: Non-Sales employees
  - Alice
  - Carol
  - David
  - Eve

Example 6: Active Engineering employees over 30
  (no output - no matches)

Example 7: All active employees' details
  - Alice (Engineering)
  - Bob (Sales)
  - David (Marketing)
  - Eve (Engineering)

Example 8: Employees under 25 (no matches)
  No employees found

Example 9: Top 2 earners (sorted)
Highest salary: $95000

Example 10: Count active employees
Active employees: 4
```

## Key Concepts

### Filter Syntax

Filters use GJSON-style syntax with comparison operators:
```
element.#(condition)     # First match
element.#(condition)#    # All matches
```

### Comparison Operators

- `==` - Equality (string or numeric)
- `!=` - Inequality
- `<` - Less than (numeric)
- `>` - Greater than (numeric)
- `<=` - Less than or equal (numeric)
- `>=` - Greater than or equal (numeric)

### Filter Types

1. **Element filters**: `employee.#(age>30)`
2. **Attribute filters**: `employee.#(@status==active)`

### Type Coercion

Filters automatically coerce types:
- Numeric comparisons: strings converted to numbers
- String comparisons: case-sensitive exact match
- Boolean: "true"/"false" strings

## Code Walkthrough

The example demonstrates advanced filtering techniques:

1. **Numeric Comparison**: Filter employees by age using `>`
2. **String Equality**: Filter by department name
3. **Attribute Filter**: Use `@status` to filter by attribute value
4. **Range Queries**: Use `>=` and `<` for salary ranges
5. **Inequality**: Use `!=` to exclude departments
6. **Manual Filtering**: Iterate and filter manually (chained filters not supported)
7. **Filter with Iteration**: Process filtered results
8. **Empty Results**: Handle queries with no matches
9. **Filter with Modifiers**: Combine filters with sorting
10. **Count Filtered**: Count elements matching filter

## Common Pitfalls

- **Pitfall**: Forgetting `@` prefix for attribute filters
  - **Solution**: Use `.#(@attribute==value)`, not `.#(attribute==value)`

- **Pitfall**: Trying to chain filters like `.#(...)#(...)#`
  - **Solution**: Chained filters NOT supported; use manual iteration with Get() on Result.Raw

- **Pitfall**: Case sensitivity in string comparisons
  - **Solution**: Filters are case-sensitive; normalize data if needed

- **Pitfall**: Numeric comparison on non-numeric strings
  - **Solution**: Non-numeric strings evaluate as 0; validate data structure

## Performance Considerations

Filter operations are optimized but have some costs:

- **Numeric filters**: Fast path optimization (16-21% improvement)
- **String filters**: Require full string comparison
- **Multiple filters**: Evaluated sequentially (short-circuit possible)
- **Large datasets**: Consider alternative approaches for very large XML

Benchmark from Phase 4 optimization:
- Simple filter: ~1,200 ns/op
- Complex filter: ~2,500 ns/op

## Next Steps

- [Modifiers](../modifiers/) - Transform filtered results
- [Wildcards](../basic-get/) - Combine with wildcard queries
- [Performance](../performance/) - Optimize filter operations

## See Also

- [Arrays](../arrays/) - Array operations and iteration
- [Main README](../../README.md) - Complete documentation
- [API Reference](https://pkg.go.dev/github.com/netascode/xmldot) - Full API documentation

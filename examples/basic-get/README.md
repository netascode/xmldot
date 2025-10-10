# Basic Get Operations

**Complexity Level**: Beginner
**Estimated Time**: 10 minutes
**Prerequisites**: Basic Go knowledge, understanding of XML structure

## What You'll Learn

- Access XML elements using simple path syntax
- Retrieve XML attributes with the `@` prefix
- Navigate nested elements
- Work with arrays using indexing and count operations
- Convert results to different types (string, int, float)
- Check for element existence
- Query multiple paths with `GetMany`

## Running the Example

```bash
cd examples/basic-get
go run main.go
```

## Expected Output

```
Basic Get Operations Example
=============================

Example 1: Simple element access
First product name: Laptop

Example 2: Attribute access
First product ID: 101

Example 3: Nested element access
CPU: Intel i7

Example 4: Array indexing
Second product: Mouse

Example 5: Array count
Total products: 3

Example 6: Type conversion
Price as float: 999.99
Price as string: 999.99

Example 7: GetMany for multiple paths
Product: Laptop (ID: 101) - $999.99

Example 8: Non-existent paths
Exists: false
Value:

Example 9: Checking existence
Specs element exists
Second product has no specs
```

## Key Concepts

### Path Syntax

xmldot uses dot notation for navigating XML elements:
- `element.child` - Access child elements
- `element.@attribute` - Access attributes
- `element.0` - Array indexing (first element)
- `element.#` - Array count

### Result Methods

The `Result` type provides several methods for accessing values:
- `String()` - Get string representation
- `Int()` - Convert to integer
- `Float()` - Convert to float64
- `Bool()` - Convert to boolean
- `Exists()` - Check if element was found

### Array Detection

Repeated elements with the same name are automatically detected as arrays:
```xml
<catalog>
    <product>...</product>  <!-- Index 0 -->
    <product>...</product>  <!-- Index 1 -->
    <product>...</product>  <!-- Index 2 -->
</catalog>
```

## Code Walkthrough

The example demonstrates fundamental operations:

1. **Simple Element Access**: Retrieve the first product's name using basic dot notation
2. **Attribute Access**: Use `@id` to access XML attributes
3. **Nested Elements**: Navigate multiple levels with `catalog.product.specs.cpu`
4. **Array Indexing**: Access specific array elements using numeric indices
5. **Array Count**: Use `#` to count repeated elements
6. **Type Conversion**: Convert string values to numeric types
7. **Batch Queries**: Query multiple paths efficiently with `GetMany`
8. **Missing Elements**: Handle non-existent paths gracefully
9. **Existence Checks**: Verify elements exist before processing

## Common Pitfalls

- **Pitfall**: Forgetting the `@` prefix for attributes
  - **Solution**: Always use `@` for attributes: `element.@attr`, not `element.attr`

- **Pitfall**: Expecting errors for missing elements
  - **Solution**: Use `Exists()` to check before accessing values

- **Pitfall**: Assuming single elements are arrays
  - **Solution**: Arrays require multiple elements with the same name

## Next Steps

- [Basic Set Operations](../basic-set/) - Learn to modify XML
- [Arrays](../arrays/) - Advanced array operations and iteration
- [Path Syntax Guide](../../docs/path-syntax.md) - Complete path syntax reference

## See Also

- [Main README](../../README.md) - Complete documentation
- [API Reference](https://pkg.go.dev/github.com/netascode/xmldot) - Full API documentation
- [Filters Example](../filters/) - Advanced queries with conditions

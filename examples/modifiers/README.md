# Built-in Modifiers

**Complexity Level**: Intermediate
**Estimated Time**: 20 minutes
**Prerequisites**: Understanding of arrays and filters

## What You'll Learn

- Transform query results with built-in modifiers
- Use array modifiers (@reverse, @sort, @first, @last, @flatten)
- Format XML output (@pretty, @ugly, @raw)
- Extract structure information (@keys, @values)
- Chain multiple modifiers for complex transformations
- Combine modifiers with filters and wildcards

## Running the Example

```bash
cd examples/modifiers
go run main.go
```

## Expected Output

```
Built-in Modifiers Example
==========================

Example 1: @reverse - Reverse array order
  - Best Practices
  - Advanced Topics
  - Getting Started

Example 2: @sort - Sort titles alphabetically
  - Advanced Topics
  - Best Practices
  - Getting Started

Example 3: @first - Get first post title
First title: Getting Started

Example 4: @last - Get last post title
Last title: Best Practices

Example 5: @flatten - Flatten nested structure
Flattened: [A B C D]

Example 6: @pretty - Format XML
<root>
	<item>value</item>
	<nested>
		<child>data</child>
	</nested>
</root>

Example 7: @ugly - Compact XML
Compact: <root><item>value</item><nested><child>data</child></nested></root>

Example 8: @raw - Get raw XML
<post id="3"><title>Getting Started</title><views>150</views></post>

Example 9: @keys - Extract keys
Keys: id, title, views

Example 10: @values - Extract values
Values: 3, Getting Started, 150

Example 11: Chaining - @sort|@reverse|@first
Last title alphabetically: Getting Started

Example 12: Combining filters and modifiers
Popular posts (sorted):
  - Advanced Topics
  - Best Practices
```

## Key Concepts

### Modifier Syntax

Modifiers use the pipe operator:
```
path.to.element|@modifier
```

Chain multiple modifiers:
```
path.to.element|@modifier1|@modifier2|@modifier3
```

### Built-in Modifiers

**Array Modifiers**:
- `@reverse` - Reverse array element order
- `@sort` - Sort array elements alphabetically
- `@first` - Extract first array element
- `@last` - Extract last array element
- `@flatten` - Flatten nested arrays into single-level array

**Formatting Modifiers**:
- `@pretty` - Format XML with indentation (tab-based)
- `@ugly` - Remove all whitespace formatting
- `@raw` - Return raw XML string (preserves formatting)

**Structure Modifiers**:
- `@keys` - Extract element/attribute names
- `@values` - Extract element/attribute values

## Code Walkthrough

The example demonstrates all built-in modifiers:

1. **Reverse**: Change array order from original
2. **Sort**: Alphabetical sorting of array elements
3. **First**: Extract first element from array
4. **Last**: Extract last element from array
5. **Flatten**: Convert nested arrays to single-level
6. **Pretty**: Add readable indentation to XML
7. **Ugly**: Remove formatting for compact storage
8. **Raw**: Get complete XML element with tags
9. **Keys**: List all element/attribute names
10. **Values**: List all element/attribute values
11. **Chaining**: Combine modifiers for complex transformations
12. **With Filters**: Use modifiers on filtered results

## Common Pitfalls

- **Pitfall**: Applying array modifiers to single elements
  - **Solution**: Single elements treated as 1-element array

- **Pitfall**: Expecting @sort to sort numerically
  - **Solution**: @sort uses lexicographic (string) sorting

- **Pitfall**: Chaining order matters
  - **Solution**: Modifiers execute left-to-right: `@sort|@reverse` ≠ `@reverse|@sort`

- **Pitfall**: Using @pretty/@ugly on non-XML content
  - **Solution**: These modifiers expect valid XML structure

## Performance Characteristics

Modifiers have varying performance costs:

- **Low cost**: @first, @last, @raw (O(1))
- **Medium cost**: @reverse, @flatten (O(n))
- **Higher cost**: @sort (O(n log n))
- **Formatting cost**: @pretty, @ugly (O(n) with parsing overhead)

For large arrays, consider:
- Use @first/@last instead of sorting when possible
- Apply filters before sorting to reduce data size
- Cache results if used multiple times

## Modifier Execution Order

Modifiers execute in pipeline order (left-to-right):

```
Query → Result → @filter → @sort → @reverse → @first → Final Result
```

Example execution:
```go
"posts.post[views>100].title|@sort|@reverse|@first"
// 1. Query: posts.post[views>100].title → ["Advanced Topics", "Best Practices"]
// 2. @sort → ["Advanced Topics", "Best Practices"]
// 3. @reverse → ["Best Practices", "Advanced Topics"]
// 4. @first → "Best Practices"
```

## Custom Modifiers

You can create custom modifiers for application-specific transformations.
See [Custom Modifiers Example](../custom-modifiers/) for details.

## Next Steps

- [Custom Modifiers](../custom-modifiers/) - Create your own modifiers
- [Performance](../performance/) - Optimize modifier usage
- [Filters](../filters/) - Combine with filtering

## See Also

- [Arrays](../arrays/) - Array operations
- [Main README](../../README.md) - Complete documentation
- [API Reference](https://pkg.go.dev/github.com/netascode/xmldot) - Full API documentation

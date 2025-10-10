# Array Operations

**Complexity Level**: Intermediate
**Estimated Time**: 15 minutes
**Prerequisites**: Understanding of basic Get and Set operations

## What You'll Learn

- Work with repeated XML elements (arrays)
- Access elements by index (positive and negative)
- Count array elements
- Replace specific array elements
- Append to arrays
- Delete array elements
- Iterate over arrays with `ForEach`
- Extract all array elements with `Array()`
- Batch operations on arrays with `SetMany` and `DeleteMany`

## Running the Example

```bash
cd examples/arrays
go run main.go
```

## Expected Output

```
Array Operations Example
========================

Example 1: Count array elements
Total items: 3

Example 2: Access by index
First: Book
Second: Pen
Last: Notebook

Example 3: Negative indices (last element)
Last item: Notebook

Example 4: Replace array element
Updated second item: Pencil

Example 5: Append to array
Items after append: 4

Example 6: Delete array element
Items after delete: 3

Example 7: Iterate with ForEach
Current items:
  - Pencil (qty: 5)
  - Notebook (qty: 1)
  - Eraser (qty: 3)

Example 8: Array() method for all elements
1. Pencil
2. Notebook
3. Eraser

Example 9: SetMany for multiple array updates
Updated quantities: 10, 20, 30

Example 10: DeleteMany for batch deletion
Items remaining after batch delete: 1
```

## Key Concepts

### Array Detection

Repeated elements with the same name are automatically detected as arrays:
```xml
<items>
    <item>...</item>  <!-- Index 0 -->
    <item>...</item>  <!-- Index 1 -->
    <item>...</item>  <!-- Index 2 -->
</items>
```

### Array Indexing

- **Positive indices**: `item.0`, `item.1`, `item.2` (zero-based)
- **Negative indices**: `item.-1` (last element), `item.-2` (second-to-last)
- **Count**: `item.#` returns the number of elements

### Array Modification

- **Replace**: Use positive index with `Set`
- **Append**: Use `-1` index with `Set`
- **Delete**: Use any valid index with `Delete`

### Array Iteration

Two methods for iterating:
1. **ForEach**: Callback-based iteration with early exit support
2. **Array()**: Returns `[]Result` slice for standard Go loops

## Code Walkthrough

The example demonstrates comprehensive array operations:

1. **Count Elements**: Use `#` to get array length
2. **Index Access**: Access specific elements by position
3. **Negative Indexing**: Use `-1` for last element (Python-style)
4. **Element Replacement**: Update specific array element
5. **Array Append**: Add new element at end using `-1` index
6. **Element Deletion**: Remove specific element by index
7. **ForEach Iteration**: Process each element with callback
8. **Array Extraction**: Get all elements as Go slice
9. **Batch Updates**: Update multiple elements efficiently
10. **Batch Deletion**: Delete multiple elements in one call

## Common Pitfalls

- **Pitfall**: Expecting 1-based indexing
  - **Solution**: Arrays are zero-based: first element is `item.0`

- **Pitfall**: Out-of-bounds access throws error
  - **Solution**: Out-of-bounds returns empty Result with `Exists() == false`

- **Pitfall**: Deleting elements changes indices
  - **Solution**: Delete from highest to lowest index, or use batch operations

- **Pitfall**: Appending with wrong index
  - **Solution**: Use `-1` to append, not the next sequential number

## Next Steps

- [Filters](../filters/) - Query arrays with conditions
- [Modifiers](../modifiers/) - Transform arrays with built-in modifiers
- [Performance](../performance/) - Optimize array operations

## See Also

- [Basic Get Operations](../basic-get/) - Fundamental queries
- [Basic Set Operations](../basic-set/) - Modification basics
- [Main README](../../README.md) - Complete documentation
- [API Reference](https://pkg.go.dev/github.com/netascode/xmldot) - Full API documentation

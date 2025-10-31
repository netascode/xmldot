# Basic Set Operations

**Complexity Level**: Beginner
**Estimated Time**: 15 minutes
**Prerequisites**: Understanding of basic Get operations

## What You'll Learn

- Modify existing XML elements
- Set XML attributes
- Create new elements automatically
- Create nested paths on-the-fly
- Delete elements and attributes
- Work with different value types (string, int, bool)
- Create XML from scratch (empty XML)
- Build multi-root XML fragments with sibling elements
- Verify modifications with Get operations

## Running the Example

```bash
cd examples/basic-set
go run main.go
```

## Expected Output

```
Basic Set Operations Example
============================

Example 1: Update element value
Updated name: Jane Smith

Example 2: Set attribute
Notifications enabled: false

Example 3: Create new element
Phone: 555-1234

Example 4: Create nested path
Privacy level: high

Example 5: Delete element
Phone exists: false

Example 6: Delete attribute
Enabled attribute exists: false

Example 7: Set with different types
Age: 30
Premium: true

Example 8: Create XML from scratch (empty XML)
Created hostname: router1

Example 9: Create sibling root elements (multi-root fragment)
Sequence: 10
Deny prefix: 10.0.0.0
Permit prefix: 192.168.0.0
Fragment: <sequence>10</sequence><deny><prefix>10.0.0.0</prefix></deny><permit><prefix>192.168.0.0</prefix></permit>

Example 10: Final XML state
<user>
	<name>Jane Smith</name>
	<email>john@example.com</email>
	<age>30</age>
	<premium>true</premium>
	<settings>
		<theme>light</theme>
		<notifications/>
		<privacy>
			<level>high</level>
		</privacy>
	</settings>
</user>
```

## Key Concepts

### Set Operations

The `Set` function updates or creates elements:
```go
newXML, err := xmldot.Set(xml, "path.to.element", value)
```

Key behaviors:
- Updates existing elements
- Creates missing elements automatically
- Creates entire missing paths
- Accepts empty XML to create from scratch
- Creates sibling roots when path specifies different root name
- Returns modified XML string
- Returns error only for invalid XML or paths

### Delete Operations

The `Delete` function removes elements or attributes:
```go
newXML, err := xmldot.Delete(xml, "path.to.element")
```

Deleting non-existent elements is not an error.

### Type Handling

Set accepts various Go types:
- `string` - Set as text content
- `int`, `int64` - Converted to string
- `float64` - Converted to string
- `bool` - Converted to "true"/"false"
- XML string - Set as raw XML

## Code Walkthrough

The example demonstrates core modification operations:

1. **Update Existing**: Change the user's name from "John Doe" to "Jane Smith"
2. **Set Attribute**: Modify the `enabled` attribute value
3. **Create Element**: Add a new `phone` element (auto-created)
4. **Create Path**: Create nested `settings.privacy.level` in one operation
5. **Delete Element**: Remove the `phone` element
6. **Delete Attribute**: Remove the `enabled` attribute
7. **Type Conversion**: Set numeric and boolean values
8. **Create from Empty**: Build XML from scratch starting with empty string
9. **Sibling Roots**: Create multi-root fragments by adding different root elements
10. **Verification**: Use `@pretty` modifier to view final XML

## Common Pitfalls

- **Pitfall**: Forgetting to capture the returned XML
  - **Solution**: Always assign the result: `xml, err = xmldot.Set(xml, ...)`

- **Pitfall**: Expecting Set to modify the original XML
  - **Solution**: Set returns a new XML string; the original is unchanged

- **Pitfall**: Not checking errors
  - **Solution**: Always check `err` after Set/Delete operations

- **Pitfall**: Deleting elements that don't exist
  - **Solution**: Not an error - Delete is idempotent

## Next Steps

- [Arrays](../arrays/) - Work with repeated elements
- [Filters](../filters/) - Conditional updates
- [Performance](../performance/) - Optimize batch operations

## See Also

- [Basic Get Operations](../basic-get/) - Reading XML
- [Main README](../../README.md) - Complete documentation
- [API Reference](https://pkg.go.dev/github.com/netascode/xmldot) - Full API documentation

# Path Syntax Reference

Complete reference for XMLDOT path syntax, inspired by JSONPath and adapted for XML.

## Table of Contents

1. [Introduction](#introduction)
2. [Basic Element Access](#basic-element-access)
3. [Attribute Access](#attribute-access)
4. [Array Operations](#array-operations)
5. [Text Content](#text-content)
6. [Wildcards](#wildcards)
7. [Filters](#filters)
8. [Modifiers](#modifiers)
9. [Namespace Support](#namespace-support)
10. [Escape Sequences](#escape-sequences)
11. [Path Composition](#path-composition)
12. [Performance Considerations](#performance-considerations)
13. [Fluent API](#fluent-api-v020)
14. [Quick Reference](#quick-reference)

---

## Introduction

XMLDOT uses a simple, dot-notation path syntax to query and manipulate XML documents. The syntax is inspired by GJSON for JSON, adapted to XML's unique characteristics like attributes, namespaces, and mixed content.

### Design Philosophy

- **Simplicity over completeness**: Easy-to-read paths for common cases
- **Performance-first**: Fast parsing and execution
- **Secure-by-default**: Built-in protection against DoS attacks
- **XML-aware**: Native support for attributes, namespaces, text content

### Comparison to XPath

| Feature | XPath | XMLDOT |
|---------|-------|--------|
| Element access | `/root/child/element` | `root.child.element` |
| Attributes | `/root/@id` | `root.@id` |
| Array index | `/items/item[1]` | `items.item.0` |
| Filters | `/users/user[age>21]` | `users.user.#(age>21)` |
| Wildcards | `/root/*/name` | `root.*.name` |
| Recursive | `//price` | `root.**.price` |
| Namespaces | Full support | Prefix matching only |
| Axes | Full support | Not supported |

**When to use XMLDOT**: Simple queries, high performance, untrusted XML
**When to use XPath**: Complex queries, full namespace support, standards compliance

### Reading Guide

- Code examples use `→` to show output
- All examples are runnable (see [examples/](../examples/))
- Security notes highlight DoS protections

---

## Basic Element Access

Access XML elements using dot notation, similar to accessing nested objects in JSON.

### Simple Paths

```go
xml := `<root><child><element>value</element></child></root>`
result := xmldot.Get(xml, "root.child.element")
fmt.Println(result.String()) // → "value"
```

### Nested Elements

Navigate through multiple levels:

```go
xml := `
<catalog>
    <products>
        <product>
            <name>Laptop</name>
            <specs>
                <cpu>Intel i7</cpu>
            </specs>
        </product>
    </products>
</catalog>`

result := xmldot.Get(xml, "catalog.products.product.specs.cpu")
fmt.Println(result.String()) // → "Intel i7"
```

### Root Element

The path must start from the root element:

```go
xml := `<catalog><book>Title</book></catalog>`
result := xmldot.Get(xml, "catalog.book")  // ✓ Correct
result := xmldot.Get(xml, "book")          // ✗ Won't match
```

### Case Sensitivity

Element names are case-sensitive by default:

```go
xml := `<Root><Child>value</Child></Root>`
result := xmldot.Get(xml, "Root.Child")    // ✓ Matches
result := xmldot.Get(xml, "root.child")    // ✗ Doesn't match
```

### Empty Elements

Empty elements return empty strings:

```go
xml := `<root><empty></empty></root>`
result := xmldot.Get(xml, "root.empty")
fmt.Println(result.Exists())  // → true
fmt.Println(result.String())  // → ""
```

### Self-Closing Elements

Self-closing elements are treated the same as empty elements:

```go
xml := `<root><empty/></root>`
result := xmldot.Get(xml, "root.empty")
fmt.Println(result.Exists())  // → true
fmt.Println(result.String())  // → ""
```

### Whitespace Handling

Leading and trailing whitespace in element text is preserved:

```go
xml := `<root><text>  value  </text></root>`
result := xmldot.Get(xml, "root.text")
fmt.Println(result.String())  // → "  value  "
```

### Non-Existent Paths

Paths that don't match return a null result:

```go
xml := `<root><child>value</child></root>`
result := xmldot.Get(xml, "root.missing")
fmt.Println(result.Exists())  // → false
fmt.Println(result.String())  // → ""
```

### Path Segment Limit

Paths are limited to 100 segments for security:

```go
// Security: Extremely long paths are rejected
path := "root" + strings.Repeat(".child", 150)
result := xmldot.Get(xml, path)  // Returns Null (exceeds MaxPathSegments)
```

**Example: Product Catalog**

```go
xml := `
<catalog>
    <product id="101">
        <name>Go Programming</name>
        <price>44.99</price>
    </product>
</catalog>`

name := xmldot.Get(xml, "catalog.product.name")
price := xmldot.Get(xml, "catalog.product.price")

fmt.Println(name.String())    // → "Go Programming"
fmt.Println(price.Float())    // → 44.99
```

---

## Attribute Access

Access XML attributes using the `@` prefix.

### Basic Attribute Syntax

```go
xml := `<book id="123" lang="en">Title</book>`
result := xmldot.Get(xml, "book.@id")
fmt.Println(result.String())  // → "123"
```

### Nested Element Attributes

Access attributes on nested elements:

```go
xml := `
<catalog>
    <product id="101" status="active">
        <name>Laptop</name>
    </product>
</catalog>`

id := xmldot.Get(xml, "catalog.product.@id")
status := xmldot.Get(xml, "catalog.product.@status")

fmt.Println(id.String())      // → "101"
fmt.Println(status.String())  // → "active"
```

### Multiple Attributes

Query different attributes on the same element:

```go
xml := `<user id="5" role="admin" active="true">John</user>`

results := xmldot.GetMany(xml,
    "user.@id",
    "user.@role",
    "user.@active")

fmt.Println(results[0].String())  // → "5"
fmt.Println(results[1].String())  // → "admin"
fmt.Println(results[2].Bool())    // → true
```

### Attribute vs Element Disambiguation

The `@` prefix distinguishes attributes from elements:

```go
xml := `
<product id="101">
    <id>SKU-12345</id>
</product>`

attr := xmldot.Get(xml, "product.@id")     // → "101" (attribute)
elem := xmldot.Get(xml, "product.id")      // → "SKU-12345" (element)
```

### Non-Existent Attributes

Missing attributes return null results:

```go
xml := `<book id="123">Title</book>`
result := xmldot.Get(xml, "book.@isbn")
fmt.Println(result.Exists())  // → false
```

### Empty Attribute Values

Attributes with empty values are distinguished from missing attributes:

```go
xml := `<book id="" title="Book">Content</book>`

id := xmldot.Get(xml, "book.@id")
missing := xmldot.Get(xml, "book.@isbn")

fmt.Println(id.Exists())       // → true
fmt.Println(id.String())       // → ""
fmt.Println(missing.Exists())  // → false
```

### Special Attribute Names

XML allows various characters in attribute names:

```go
xml := `<element data-id="123" xmlns:ns="uri">Content</element>`

dataId := xmldot.Get(xml, "element.@data-id")      // → "123"
xmlns := xmldot.Get(xml, "element.@xmlns:ns")      // → "uri"
```

**Example: Configuration File**

```go
xml := `
<config version="2.0">
    <database host="localhost" port="5432" ssl="true">
        <name>production</name>
    </database>
</config>`

version := xmldot.Get(xml, "config.@version")
host := xmldot.Get(xml, "config.database.@host")
port := xmldot.Get(xml, "config.database.@port")
ssl := xmldot.Get(xml, "config.database.@ssl")

fmt.Printf("Version: %s\n", version.String())  // → Version: 2.0
fmt.Printf("Host: %s:%d\n", host.String(), port.Int())  // → Host: localhost:5432
fmt.Printf("SSL: %v\n", ssl.Bool())  // → SSL: true
```

---

## Array Operations

XML arrays are represented as repeated elements with the same name. XMLDOT provides array indexing, counting, and iteration.

### Array Detection

Repeated elements are automatically treated as arrays:

```go
xml := `
<items>
    <item>First</item>
    <item>Second</item>
    <item>Third</item>
</items>`

// Access by index (0-based)
result := xmldot.Get(xml, "items.item.0")
fmt.Println(result.String())  // → "First"
```

### Array Indexing

Access specific array elements by index:

```go
xml := `
<books>
    <book><title>Book 1</title></book>
    <book><title>Book 2</title></book>
    <book><title>Book 3</title></book>
</books>`

first := xmldot.Get(xml, "books.book.0.title")
second := xmldot.Get(xml, "books.book.1.title")
third := xmldot.Get(xml, "books.book.2.title")

fmt.Println(first.String())   // → "Book 1"
fmt.Println(second.String())  // → "Book 2"
fmt.Println(third.String())   // → "Book 3"
```

### Negative Indices

Negative index `-1` is supported for `Set()` and `SetRaw()` operations to append elements (see [Modifying Arrays](#modifying-arrays)).

Note: Negative indices for `Get()` are not currently implemented. To access the last element, you need to:
1. Get the count using `#`
2. Access the element at index `count-1`

```go
xml := `
<items>
    <item>First</item>
    <item>Second</item>
    <item>Third</item>
</items>`

// Get last element by calculating index
count := xmldot.Get(xml, "items.item.#").Int()
last := xmldot.Get(xml, fmt.Sprintf("items.item.%d", count-1))
fmt.Println(last.String())  // → "Third"
```

### Array Count

Use `#` to get the number of array elements:

```go
xml := `
<items>
    <item>A</item>
    <item>B</item>
    <item>C</item>
</items>`

count := xmldot.Get(xml, "items.item.#")
fmt.Println(count.Int())  // → 3
```

### Out-of-Bounds Access

Accessing beyond array bounds returns null:

```go
xml := `<items><item>A</item><item>B</item></items>`

result := xmldot.Get(xml, "items.item.10")
fmt.Println(result.Exists())  // → false
```

### Single vs Multiple Elements

A single element is not treated as an array:

```go
xml := `<items><item>Only one</item></items>`

// Access as element (no index)
result := xmldot.Get(xml, "items.item")
fmt.Println(result.String())  // → "Only one"

// Access as array (with index 0)
result = xmldot.Get(xml, "items.item.0")
fmt.Println(result.String())  // → "Only one"

// Count still works
count := xmldot.Get(xml, "items.item.#")
fmt.Println(count.Int())  // → 1
```

### Nested Arrays

Arrays can be nested:

```go
xml := `
<catalog>
    <category>
        <product>Product A1</product>
        <product>Product A2</product>
    </category>
    <category>
        <product>Product B1</product>
        <product>Product B2</product>
    </category>
</catalog>`

// Access second category, first product
result := xmldot.Get(xml, "catalog.category.1.product.0")
fmt.Println(result.String())  // → "Product B1"

// Count products in first category
count := xmldot.Get(xml, "catalog.category.0.product.#")
fmt.Println(count.Int())  // → 2
```

### Array Iteration

Use `Result.Array()` or `Result.ForEach()` to iterate:

```go
xml := `
<items>
    <item>Alpha</item>
    <item>Beta</item>
    <item>Gamma</item>
</items>`

// Get all items without index (returns array result)
items := xmldot.Get(xml, "items.item")

// Method 1: Array()
for _, item := range items.Array() {
    fmt.Println(item.String())
}

// Method 2: ForEach()
items.ForEach(func(index int, value Result) bool {
    fmt.Printf("%d: %s\n", index, value.String())
    return true  // Continue iteration
})
```

**Example: Shopping Cart**

```go
xml := `
<cart>
    <items>
        <item id="1">
            <name>Book</name>
            <quantity>2</quantity>
            <price>15.99</price>
        </item>
        <item id="2">
            <name>Pen</name>
            <quantity>5</quantity>
            <price>2.50</price>
        </item>
        <item id="3">
            <name>Notebook</name>
            <quantity>1</quantity>
            <price>8.99</price>
        </item>
    </items>
</cart>`

// Count items
itemCount := xmldot.Get(xml, "cart.items.item.#")
fmt.Printf("Cart has %d items\n", itemCount.Int())  // → Cart has 3 items

// Get first item details
firstName := xmldot.Get(xml, "cart.items.item.0.name")
firstQty := xmldot.Get(xml, "cart.items.item.0.quantity")
firstPrice := xmldot.Get(xml, "cart.items.item.0.price")

fmt.Printf("%s (x%d) @ $%.2f\n",
    firstName.String(),
    firstQty.Int(),
    firstPrice.Float())  // → Book (x2) @ $15.99

// Calculate total
var total float64
items := xmldot.Get(xml, "cart.items.item")
items.ForEach(func(i int, item Result) bool {
    qtyPath := fmt.Sprintf("cart.items.item.%d.quantity", i)
    pricePath := fmt.Sprintf("cart.items.item.%d.price", i)

    qty := xmldot.Get(xml, qtyPath).Int()
    price := xmldot.Get(xml, pricePath).Float()

    total += float64(qty) * price
    return true
})
fmt.Printf("Total: $%.2f\n", total)  // → Total: $52.48
```

### Array Append Operations

**Available since:** v0.4.0

Use index `-1` with `Set()` or `SetRaw()` to append new elements to an array:

```go
xml := `
<cart>
    <items>
        <item><name>Book</name></item>
        <item><name>Pen</name></item>
    </items>
</cart>`

// Append a new item to the array using SetRaw for XML content
result, err := xmldot.SetRaw(xml, "cart.items.item.-1", "<name>Notebook</name>")
if err != nil {
    log.Fatal(err)
}

// Verify the append
count := xmldot.Get(result, "cart.items.item.#")
fmt.Println(count.Int())  // → 3

// Get last item's name by index
lastIndex := count.Int() - 1
lastName := xmldot.Get(result, fmt.Sprintf("cart.items.item.%d.name", lastIndex))
fmt.Println(lastName.String())  // → "Notebook"
```

**Creating First Element**: When the array is empty, `-1` creates the first element:

```go
xml := `<cart><items></items></cart>`

result, err := xmldot.SetRaw(xml, "cart.items.item.-1", "<name>First Item</name>")
if err != nil {
    log.Fatal(err)
}

count := xmldot.Get(result, "cart.items.item.#")
fmt.Println(count.Int())  // → 1
```

**Auto-Creating Parents**: If the parent path doesn't exist, it will be created automatically:

```go
xml := `<cart></cart>`

// Creates <items> parent automatically
result, err := xmldot.SetRaw(xml, "cart.items.item.-1", "<name>First Item</name>")
if err != nil {
    log.Fatal(err)
}

// Result: <cart><items><item><name>First Item</name></item></items></cart>
```

**Appending to Single Elements**: If only one element exists, `-1` treats it as a 1-element array and appends a second:

```go
xml := `<items><item>First</item></items>`

// Use Set() for simple text content, SetRaw() for XML markup
result, err := xmldot.Set(xml, "items.item.-1", "Second")
if err != nil {
    log.Fatal(err)
}

// Result: <items><item>First</item><item>Second</item></items>
```

**Limitations**:

- Only supported in `Set()` and `SetRaw()` operations
- Nested paths after `-1` are not allowed: `item.-1.child` returns an error
- Other negative indices (`-2`, `-3`, etc.) are reserved and return an error

---

## Text Content

Extract text content while ignoring child elements.

### Pure Text Extraction

Use `%` to get only text content, skipping child elements:

```go
xml := `
<article>
    Article introduction text.
    <section>Section 1</section>
    More article text.
    <section>Section 2</section>
</article>`

// Without % - gets all content including child elements
full := xmldot.Get(xml, "article")
fmt.Println(full.String())
// → "Article introduction text.\n    Section 1\n    More article text.\n    Section 2"

// With % - only direct text nodes
text := xmldot.Get(xml, "article.%")
fmt.Println(text.String())
// → "Article introduction text.\n    More article text."
```

### Text vs Element Content

```go
xml := `<book><title>The Go Programming Language</title></book>`

// Element content (default)
title := xmldot.Get(xml, "book.title")
fmt.Println(title.String())  // → "The Go Programming Language"

// Text content (same result for simple elements)
titleText := xmldot.Get(xml, "book.title.%")
fmt.Println(titleText.String())  // → "The Go Programming Language"
```

### Mixed Content

The `%` modifier is most useful for mixed content:

```go
xml := `
<paragraph>
    This is <bold>important</bold> text with <italic>formatting</italic>.
</paragraph>`

// Without % - includes child element text
full := xmldot.Get(xml, "paragraph")
fmt.Println(full.String())  // → "This is important text with formatting."

// With % - only direct text nodes
text := xmldot.Get(xml, "paragraph.%")
fmt.Println(text.String())  // → "This is  text with ."
```

### Whitespace Preservation

Text content preserves whitespace:

```go
xml := `
<poem>
    Roses are red,
    Violets are blue,
    XML is structured,
    And so are you.
</poem>`

text := xmldot.Get(xml, "poem.%")
fmt.Println(text.String())
// → "    Roses are red,\n    Violets are blue,\n    XML is structured,\n    And so are you."
```

---

## Wildcards

Wildcards enable flexible queries when you don't know exact paths or want to query multiple elements.

### Single-Level Wildcard (`*`)

The `*` wildcard matches any single element at the current level:

```go
xml := `
<products>
    <electronics>
        <laptop>Laptop 1</laptop>
        <phone>Phone 1</phone>
    </electronics>
    <books>
        <fiction>Book 1</fiction>
        <nonfiction>Book 2</nonfiction>
    </books>
</products>`

// Match any category, then access specific element
result := xmldot.Get(xml, "products.*.laptop")
fmt.Println(result.String())  // → "Laptop 1"

// Match any element in electronics
result = xmldot.Get(xml, "products.electronics.*")
// Returns array with both laptop and phone
result.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())  // → "Laptop 1", "Phone 1"
    return true
})
```

### Recursive Wildcard (`**`)

The `**` wildcard matches elements at any depth:

```go
xml := `
<catalog>
    <category name="Electronics">
        <product>
            <name>Laptop</name>
            <price>999.99</price>
        </product>
    </category>
    <category name="Books">
        <product>
            <name>Go Book</name>
            <price>44.99</price>
        </product>
    </category>
</catalog>`

// Find all 'price' elements at any depth
prices := xmldot.Get(xml, "catalog.**.price")
prices.ForEach(func(i int, r Result) bool {
    fmt.Printf("Price: $%.2f\n", r.Float())
    return true
})
// → Price: $999.99
// → Price: $44.99
```

### Wildcard First Match Semantics

Without modifiers, wildcards return the first match:

```go
xml := `
<items>
    <item>First</item>
    <item>Second</item>
    <item>Third</item>
</items>`

// Single-level wildcard returns first match
result := xmldot.Get(xml, "items.*")
fmt.Println(result.String())  // → "First"

// Use modifiers to get all matches
all := xmldot.Get(xml, "items.*|@flatten")
all.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())  // → "First", "Second", "Third"
    return true
})
```

### Combining Wildcards and Filters

Wildcards work with filters to create powerful queries:

```go
xml := `
<employees>
    <department name="Engineering">
        <employee><name>Alice</name><salary>85000</salary></employee>
        <employee><name>Bob</name><salary>75000</salary></employee>
    </department>
    <department name="Sales">
        <employee><name>Carol</name><salary>95000</salary></employee>
        <employee><name>Dave</name><salary>65000</salary></employee>
    </department>
</employees>`

// Find high earners in any department
highEarners := xmldot.Get(xml, "employees.*.employee.#(salary>80000)#.name")
highEarners.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())  // → "Alice", "Carol"
    return true
})
```

### Performance Characteristics

Wildcard performance varies by type:

- **Single-level (`*`)**: Fast - O(n) where n = children at that level
- **Recursive (`**`)**: Slower - O(n) where n = total descendants
- **With filters**: Slower - filter evaluated on each match

```go
// Fast: Single-level wildcard
result := xmldot.Get(xml, "root.*.element")  // ~ 500-1000ns

// Slower: Recursive wildcard
result := xmldot.Get(xml, "root.**.element")  // ~ 1-5µs depending on depth

// Slowest: Recursive wildcard with filter
result := xmldot.Get(xml, "root.**.#(price>100)#")  // ~ 2-10µs depending on matches
```

### Wildcard Result Limits

Recursive wildcards are capped at 1000 results for security:

```go
// Security: Large result sets are limited
const MaxWildcardResults = 1000

// If document has >1000 matching elements, only first 1000 are returned
result := xmldot.Get(largeXML, "root.**.item")
count := len(result.Array())  // Maximum 1000
```

### Wildcard with Attributes

Wildcards can be followed by attribute access:

```go
xml := `
<products>
    <product id="1">Laptop</product>
    <product id="2">Mouse</product>
    <product id="3">Keyboard</product>
</products>`

// Get all product IDs
result := xmldot.Get(xml, "products.*.@id")
result.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())  // → "1", "2", "3"
    return true
})
```

**Example: Find All Prices in Complex Document**

```go
xml := `
<store>
    <electronics>
        <computers>
            <laptop><name>Pro 15</name><price>1299.99</price></laptop>
            <desktop><name>Tower</name><price>899.99</price></desktop>
        </computers>
        <accessories>
            <mouse><name>Wireless</name><price>29.99</price></mouse>
        </accessories>
    </electronics>
    <books>
        <fiction>
            <book><title>Novel 1</title><price>15.99</price></book>
        </fiction>
    </books>
</store>`

// Find all prices regardless of location
prices := xmldot.Get(xml, "store.**.price")

var total float64
prices.ForEach(func(i int, r Result) bool {
    price := r.Float()
    fmt.Printf("Item %d: $%.2f\n", i+1, price)
    total += price
    return true
})
fmt.Printf("Total value: $%.2f\n", total)

// Output:
// Item 1: $1299.99
// Item 2: $899.99
// Item 3: $29.99
// Item 4: $15.99
// Total value: $2245.96
```

---

## Filters

Filters enable conditional queries based on element values or attributes using GJSON-style syntax.

### Filter Syntax

XMLDOT uses GJSON-style filter syntax with `#(condition)` for first match or `#(condition)#` for all matches:

```go
xml := `
<users>
    <user><name>Alice</name><age>28</age></user>
    <user><name>Bob</name><age>35</age></user>
    <user><name>Carol</name><age>42</age></user>
</users>`

// Find first user older than 30
result := xmldot.Get(xml, "users.user.#(age>30).name")
fmt.Println(result.String())  // → "Bob" (first match)

// Find all users older than 30
result = xmldot.Get(xml, "users.user.#(age>30)#.name")
results := result.Array()
// → ["Bob", "Carol"]
```

### Numeric Comparison Operators

Supported operators: `==`, `!=`, `<`, `>`, `<=`, `>=`

```go
xml := `
<products>
    <product><name>Laptop</name><price>999.99</price><stock>5</stock></product>
    <product><name>Mouse</name><price>29.99</price><stock>50</stock></product>
    <product><name>Monitor</name><price>299.99</price><stock>0</stock></product>
</products>`

// Equal
result := xmldot.Get(xml, "products.product.#(stock==50).name")
fmt.Println(result.String())  // → "Mouse"

// Not equal
result = xmldot.Get(xml, "products.product.#(stock!=0).name")
fmt.Println(result.String())  // → "Laptop"

// Greater than
result = xmldot.Get(xml, "products.product.#(price>500).name")
fmt.Println(result.String())  // → "Laptop"

// Less than or equal
result = xmldot.Get(xml, "products.product.#(price<=100).name")
fmt.Println(result.String())  // → "Mouse"

// Greater than or equal
result = xmldot.Get(xml, "products.product.#(stock>=50).name")
fmt.Println(result.String())  // → "Mouse"
```

### String Comparison

The `==` and `!=` operators work with strings:

```go
xml := `
<employees>
    <employee><name>Alice</name><dept>Engineering</dept></employee>
    <employee><name>Bob</name><dept>Sales</dept></employee>
    <employee><name>Carol</name><dept>Engineering</dept></employee>
</employees>`

// String equality
engrs := xmldot.Get(xml, "employees.employee.#(dept==Engineering).name")
fmt.Println(engrs.String())  // → "Alice"

// String inequality
nonEngrs := xmldot.Get(xml, "employees.employee.#(dept!=Engineering).name")
fmt.Println(nonEngrs.String())  // → "Bob"
```

### Attribute Filters

Filter by attribute values using `@` prefix:

```go
xml := `
<items>
    <item id="1" status="active">Item A</item>
    <item id="2" status="inactive">Item B</item>
    <item id="3" status="active">Item C</item>
</items>`

// Filter by attribute
active := xmldot.Get(xml, "items.item.#(@status==active)")
fmt.Println(active.String())  // → "Item A"

// Numeric attribute filter
item2 := xmldot.Get(xml, "items.item.#(@id==2)")
fmt.Println(item2.String())  // → "Item B"
```

### Existence Checks

Check if an attribute or element exists:

```go
xml := `
<products>
    <product featured="true"><name>Laptop</name></product>
    <product><name>Mouse</name></product>
    <product featured="false"><name>Keyboard</name></product>
</products>`

// Find products with 'featured' attribute (any value)
result := xmldot.Get(xml, "products.product.#(@featured)#.name")
result.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())  // → "Laptop", "Keyboard"
    return true
})
```

### ⚠️ Chained Filters Limitation

**Chained filters (e.g., `#(condition1).#(condition2)`) are NOT currently supported.**

```go
xml := `
<employees>
    <employee>
        <name>Alice</name>
        <age>28</age>
        <dept>Engineering</dept>
        <salary>85000</salary>
    </employee>
    <employee>
        <name>Bob</name>
        <age>35</age>
        <dept>Sales</dept>
        <salary>75000</salary>
    </employee>
    <employee>
        <name>Carol</name>
        <age>42</age>
        <dept>Engineering</dept>
        <salary>95000</salary>
    </employee>
</employees>`

// ❌ This does NOT work - chained filters not supported
result := xmldot.Get(xml, "employees.employee.#(dept==Engineering).#(salary>90000).name")
fmt.Println(result.Exists())  // → false (doesn't work)

// ✅ Workaround: Use all-matches syntax and filter manually
allEngineering := xmldot.Get(xml, "employees.employee.#(dept==Engineering)#")
allEngineering.ForEach(func(i int, emp xmldot.Result) bool {
    salary := xmldot.Get(emp.Raw, "salary")
    if salary.Int() > 90000 {
        name := xmldot.Get(emp.Raw, "name")
        fmt.Println(name.String())  // → "Carol"
    }
    return true
})
```

**Why This Limitation Exists**: The current architecture applies filters to element arrays at the current nesting level. Chaining filters would require executing a second filter on the result of the first filter, which is architecturally different.

**Future Support**: Chained filter support may be added in a future version (v2.0+) if there is sufficient user demand.

### Type Coercion Rules

Filters perform type coercion for comparisons:

```go
xml := `<data><value>42</value><text>hello</text></data>`

// Numeric comparison coerces string to number
result := xmldot.Get(xml, "data.#(value>40)")
fmt.Println(result.Exists())  // → true

// Non-numeric strings with numeric operators fail
result = xmldot.Get(xml, "data.#(text>5)")
fmt.Println(result.Exists())  // → false (can't compare "hello" as number)

// String comparison works
result = xmldot.Get(xml, "data.#(text==hello)")
fmt.Println(result.Exists())  // → true
```

### Filter Performance

Filter evaluation has performance implications:

```go
// Fast: Attribute filter (direct map lookup)
result := xmldot.Get(xml, "items.item.#(@id==5)")  // ~ 1-2µs

// Medium: Numeric element filter
result := xmldot.Get(xml, "items.item.#(price<100)")  // ~ 2-3µs

// Slowest: Recursive wildcard + filter
result := xmldot.Get(xml, "root.**.#(price<100)")  // ~ 5-20µs depending on depth
```

### Filter Security Limits

Filters have security limits to prevent DoS:

```go
const MaxFilterDepth = 10              // Max recursion depth
const MaxFilterExpressionLength = 256  // Max filter string length

// Security: Complex nested filters are limited
result := xmldot.Get(xml, "root.#(a.#(b.#(c.#(d.#(e.#(f.#(g.#(h.#(i.#(j>10))))))))))
// Exceeds MaxFilterDepth, returns Null
```

**Example: E-commerce Product Search**

```go
xml := `
<catalog>
    <products>
        <product id="101" category="electronics">
            <name>Laptop Pro 15</name>
            <price>1299.99</price>
            <rating>4.5</rating>
            <stock>12</stock>
        </product>
        <product id="102" category="electronics">
            <name>Mouse Wireless</name>
            <price>29.99</price>
            <rating>4.8</rating>
            <stock>0</stock>
        </product>
        <product id="103" category="books">
            <name>Go Programming</name>
            <price>44.99</price>
            <rating>4.9</rating>
            <stock>25</stock>
        </product>
        <product id="104" category="electronics">
            <name>Keyboard RGB</name>
            <price>79.99</price>
            <rating>4.2</rating>
            <stock>8</stock>
        </product>
    </products>
</catalog>`

// Find in-stock electronics under $100
// Note: Chained filters not supported - use manual filtering
affordable := xmldot.Get(xml, "catalog.products.product.#(@category==electronics)#")
affordable.ForEach(func(i int, r xmldot.Result) bool {
    price := xmldot.Get(r.Raw, "price")
    stock := xmldot.Get(r.Raw, "stock")
    if price.Float() < 100 && stock.Int() > 0 {
        name := xmldot.Get(r.Raw, "name")
        fmt.Println(name.String())  // → "Keyboard RGB"
    }
    return true
})

// Find highly-rated products (rating >= 4.5) with price
highRated := xmldot.Get(xml, "catalog.products.product.#(rating>=4.5)#")
highRated.ForEach(func(i int, r Result) bool {
    // Extract name and price for each result
    name := xmldot.Get(r.Raw, "name")
    price := xmldot.Get(r.Raw, "price")
    rating := xmldot.Get(r.Raw, "rating")

    fmt.Printf("%s - $%.2f (★%.1f)\n",
        name.String(), price.Float(), rating.Float())
    return true
})
// Output:
// Laptop Pro 15 - $1299.99 (★4.5)
// Mouse Wireless - $29.99 (★4.8)
// Go Programming - $44.99 (★4.9)
```

---

## Modifiers

Modifiers transform query results after path resolution.

### Modifier Syntax

Modifiers use the pipe `|` followed by `@modifierName`:

```go
xml := `
<items>
    <item>Apple</item>
    <item>Banana</item>
    <item>Cherry</item>
</items>`

// Reverse array order
result := xmldot.Get(xml, "items.item|@reverse")
result.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())
    return true
})
// → "Cherry", "Banana", "Apple"
```

### Built-in Modifiers

#### `@reverse` - Reverse Array Order

```go
xml := `<nums><n>1</n><n>2</n><n>3</n></nums>`
result := xmldot.Get(xml, "nums.n|@reverse")
// Result: ["3", "2", "1"]
```

#### `@sort` - Sort Array Elements

```go
xml := `
<items>
    <item>Zebra</item>
    <item>Apple</item>
    <item>Mango</item>
</items>`

sorted := xmldot.Get(xml, "items.item|@sort")
sorted.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())
    return true
})
// → "Apple", "Mango", "Zebra"
```

#### `@first` - Get First Element

```go
xml := `<nums><n>10</n><n>20</n><n>30</n></nums>`
result := xmldot.Get(xml, "nums.n|@first")
fmt.Println(result.String())  // → "10"
```

#### `@last` - Get Last Element

```go
xml := `<nums><n>10</n><n>20</n><n>30</n></nums>`
result := xmldot.Get(xml, "nums.n|@last")
fmt.Println(result.String())  // → "30"
```

#### `@flatten` - Flatten Nested Arrays

```go
xml := `
<data>
    <group><item>A</item><item>B</item></group>
    <group><item>C</item><item>D</item></group>
</data>`

// Without flatten: nested structure
nested := xmldot.Get(xml, "data.group")
fmt.Println(len(nested.Array()))  // → 2 groups

// With flatten: flat array
flat := xmldot.Get(xml, "data.group.item|@flatten")
fmt.Println(len(flat.Array()))  // → 4 items
```

#### `@pretty` - Format XML with Indentation

```go
xml := `<root><child><item>value</item></child></root>`

pretty := xmldot.Get(xml, "root|@pretty")
fmt.Println(pretty.String())
// Output:
// <root>
//   <child>
//     <item>value</item>
//   </child>
// </root>
```

#### `@ugly` - Compact XML (Remove Whitespace)

```go
xml := `
<root>
    <child>
        <item>value</item>
    </child>
</root>`

compact := xmldot.Get(xml, "root|@ugly")
fmt.Println(compact.String())
// → "<root><child><item>value</item></child></root>"
```

#### `@raw` - Raw XML Output

```go
xml := `<product id="5"><name>Laptop</name><price>999</price></product>`

raw := xmldot.Get(xml, "product|@raw")
fmt.Println(raw.String())
// → "<product id=\"5\"><name>Laptop</name><price>999</price></product>"
```

#### `@keys` - Extract Element Names

```go
xml := `<data><name>John</name><age>30</age><city>NYC</city></data>`

keys := xmldot.Get(xml, "data|@keys")
keys.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())
    return true
})
// → "name", "age", "city"
```

#### `@values` - Extract Values Only

```go
xml := `<data><name>John</name><age>30</age><city>NYC</city></data>`

values := xmldot.Get(xml, "data|@values")
values.ForEach(func(i int, r Result) bool {
    fmt.Println(r.String())
    return true
})
// → "John", "30", "NYC"
```

### Chaining Modifiers

Combine multiple modifiers in sequence:

```go
xml := `
<scores>
    <score>85</score>
    <score>92</score>
    <score>78</score>
    <score>95</score>
    <score>88</score>
</scores>`

// Sort, then reverse, then get first (highest score)
highest := xmldot.Get(xml, "scores.score|@sort|@reverse|@first")
fmt.Println(highest.String())  // → "95"

// Sort, then get last 3 (could chain with custom modifier)
top3 := xmldot.Get(xml, "scores.score|@sort|@reverse")
count := 0
top3.ForEach(func(i int, r Result) bool {
    if count >= 3 {
        return false  // Stop iteration
    }
    fmt.Printf("%d. %s\n", i+1, r.String())
    count++
    return true
})
// → 1. 95
// → 2. 92
// → 3. 88
```

### Custom Modifiers

Register custom modifiers for application-specific transformations:

```go
// See examples/custom-modifiers/ for detailed examples

func init() {
    // Register a custom modifier to calculate sum
    xmldot.RegisterModifier("sum", func(result xmldot.Result) xmldot.Result {
        var sum float64
        result.ForEach(func(i int, r xmldot.Result) bool {
            sum += r.Float()
            return true
        })
        return xmldot.Result{
            Type: xmldot.Number,
            Str:  fmt.Sprintf("%.2f", sum),
            Num:  sum,
        }
    })
}

xml := `<nums><n>10</n><n>20</n><n>30</n></nums>`
total := xmldot.Get(xml, "nums.n|@sum")
fmt.Println(total.Float())  // → 60.00
```

### Modifier Performance

Modifiers add minimal overhead:

```go
// Base query
result := xmldot.Get(xml, "items.item")  // ~ 500ns

// With single modifier
result := xmldot.Get(xml, "items.item|@reverse")  // ~ 600ns

// With chained modifiers
result := xmldot.Get(xml, "items.item|@sort|@reverse|@first")  // ~ 800ns
```

**Example: Data Analysis with Modifiers**

```go
xml := `
<sales>
    <transaction date="2024-01-15">
        <amount>125.50</amount>
        <customer>Alice</customer>
    </transaction>
    <transaction date="2024-01-16">
        <amount>89.99</amount>
        <customer>Bob</customer>
    </transaction>
    <transaction date="2024-01-17">
        <amount>250.00</amount>
        <customer>Carol</customer>
    </transaction>
    <transaction date="2024-01-18">
        <amount>45.25</amount>
        <customer>Dave</customer>
    </transaction>
</sales>`

// Get all amounts, sorted descending
amounts := xmldot.Get(xml, "sales.transaction.amount|@sort|@reverse")
fmt.Println("Top sales amounts:")
amounts.ForEach(func(i int, r Result) bool {
    if i < 3 {  // Top 3
        fmt.Printf("%d. $%.2f\n", i+1, r.Float())
    }
    return i < 2  // Continue only for top 3
})
// Output:
// Top sales amounts:
// 1. $250.00
// 2. $125.50
// 3. $89.99

// Get first transaction customer
firstCustomer := xmldot.Get(xml, "sales.transaction.customer|@first")
fmt.Printf("First customer: %s\n", firstCustomer.String())
// → First customer: Alice

// Get last transaction customer
lastCustomer := xmldot.Get(xml, "sales.transaction.customer|@last")
fmt.Printf("Last customer: %s\n", lastCustomer.String())
// → Last customer: Dave
```

---

## Namespace Support

XMLDOT provides basic namespace prefix matching for simple XML with predictable prefixes.

### ⚠️ Important Limitations

**XMLDOT does NOT implement full XML Namespaces (xmlns) specification**:

- ❌ NO namespace URI resolution
- ❌ NO xmlns attribute processing
- ❌ NO default namespace support
- ❌ NO namespace validation
- ✓ Only prefix string matching

**Use `encoding/xml` for full namespace support.**

### Prefix Syntax

Query elements with namespace prefixes using colon notation:

```go
xml := `
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <m:GetStockPrice xmlns:m="http://www.example.org/stock">
            <m:StockName>AAPL</m:StockName>
        </m:GetStockPrice>
    </soap:Body>
</soap:Envelope>`

// Match by prefix (xmlns attributes are ignored)
result := xmldot.Get(xml, "soap:Envelope.soap:Body.m:GetStockPrice.m:StockName")
fmt.Println(result.String())  // → "AAPL"
```

### Backward Compatibility

Paths without prefixes match local names regardless of prefix:

```go
xml := `<ns:root><ns:child>value</ns:child></ns:root>`

// Matches by local name only
result := xmldot.Get(xml, "root.child")
fmt.Println(result.String())  // → "value"

// Matches by exact prefix
result = xmldot.Get(xml, "ns:root.ns:child")
fmt.Println(result.String())  // → "value"
```

### Namespace Prefix Limitations

Example demonstrating why full namespace support is needed:

```go
xml := `
<root xmlns:a="http://example.com/a" xmlns:b="http://example.com/b">
    <a:item>Value A</a:item>
    <b:item>Value B</b:item>
</root>`

// xmldot treats these as DIFFERENT elements (by prefix string)
resultA := xmldot.Get(xml, "root.a:item")
fmt.Println(resultA.String())  // → "Value A"

resultB := xmldot.Get(xml, "root.b:item")
fmt.Println(resultB.String())  // → "Value B"

// But xmldot does NOT understand that 'a' and 'b' map to different URIs
// It only matches prefix strings literally

// Without prefix, matches first occurrence
result := xmldot.Get(xml, "root.item")
fmt.Println(result.String())  // → "Value A" (first match)
```

### When Namespace Prefix Matching Works

```go
// ✓ Good: Consistent prefixes throughout document
xml := `
<soap:Envelope>
    <soap:Body>
        <soap:Request>Data</soap:Request>
    </soap:Body>
</soap:Envelope>`

result := xmldot.Get(xml, "soap:Envelope.soap:Body.soap:Request")
// Works reliably because prefixes are predictable

// ✓ Good: Internal API with controlled XML
xml := `<ns:data><ns:item>Value</ns:item></ns:data>`
result := xmldot.Get(xml, "ns:data.ns:item")
// Works because you control the prefix convention
```

### When to Avoid Namespace Prefix Matching

```go
// ✗ Bad: Dynamic prefixes from external sources
xml := `
<root xmlns:custom="http://example.com">
    <custom:item>Value</custom:item>
</root>`

// Fails if external source changes prefix from 'custom' to 'c' or 'ex'
result := xmldot.Get(xml, "root.custom:item")

// ✗ Bad: Default namespaces
xml := `
<root xmlns="http://example.com/default">
    <item>Value</item>
</root>`

// xmldot treats this as unprefixed (no xmlns support)
result := xmldot.Get(xml, "root.item")
// May work but semantically incorrect

// Use encoding/xml for proper handling
```

### Security: Prefix Length Limit

```go
const MaxNamespacePrefixLength = 256

// Security: Extremely long prefixes are rejected
xml := `<` + strings.Repeat("a", 300) + `:element>value</element>`
result := xmldot.Get(xml, "element")
// Treats oversized prefix as unprefixed element
```

**Example: SOAP Message Handling**

```go
xml := `
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Header>
        <auth:Token xmlns:auth="http://example.com/auth">SECRET123</auth:Token>
    </soap:Header>
    <soap:Body>
        <m:GetUser xmlns:m="http://example.com/api">
            <m:UserId>12345</m:UserId>
        </m:GetUser>
    </soap:Body>
</soap:Envelope>`

// Extract token
token := xmldot.Get(xml, "soap:Envelope.soap:Header.auth:Token")
fmt.Println("Auth token:", token.String())  // → Auth token: SECRET123

// Extract user ID
userId := xmldot.Get(xml, "soap:Envelope.soap:Body.m:GetUser.m:UserId")
fmt.Println("User ID:", userId.String())  // → User ID: 12345

// Without prefixes (matches local names)
userIdAlt := xmldot.Get(xml, "Envelope.Body.GetUser.UserId")
fmt.Println("User ID (alt):", userIdAlt.String())  // → User ID (alt): 12345
```

**Migration Path for Full Namespace Support**:

```go
// For full XML Namespaces support, use encoding/xml + xmldot together:

import (
    "encoding/xml"
    "github.com/netascode/xmldot"
)

// Step 1: Parse with encoding/xml for namespace resolution
type Envelope struct {
    XMLName xml.Name `xml:"Envelope"`
    Body    Body     `xml:"Body"`
}

// Step 2: Use xmldot for flexible querying after validation
var envelope Envelope
xml.Unmarshal([]byte(soapXML), &envelope)

// Step 3: Query normalized structure with xmldot
result := xmldot.Get(normalizedXML, "Envelope.Body.Request")
```

---

## Escape Sequences

Escape special characters in path strings to treat them literally.

### Escaping Dots

Use backslash to escape dots in element names:

```go
xml := `<data><file.name>document.pdf</file.name></data>`

// Without escape: splits on dot
result := xmldot.Get(xml, "data.file.name")
fmt.Println(result.Exists())  // → false (looking for data.file.name)

// With escape: treats dot as literal
result = xmldot.Get(xml, "data.file\\.name")
fmt.Println(result.String())  // → "document.pdf"
```

### Escaping Brackets

Escape brackets when element names contain them:

```go
xml := `<data><item[0]>First</item[0]></data>`

result := xmldot.Get(xml, "data.item\\[0\\]")
fmt.Println(result.String())  // → "First"
```

### Escaping At-Signs

Escape `@` to query elements starting with @:

```go
xml := `<data><@special>value</@special></data>`

result := xmldot.Get(xml, "data.\\@special")
fmt.Println(result.String())  // → "value"
```

### Escaping Pipes

Escape pipes in element names:

```go
xml := `<data><option|default>value</option|default></data>`

result := xmldot.Get(xml, "data.option\\|default")
fmt.Println(result.String())  // → "value"
```

### Escape Sequences in Filters

Escaping also works in filter expressions:

```go
xml := `<items><item name="file.txt">Content</item></items>`

result := xmldot.Get(xml, "items.item.#(@name==file\\.txt)")
fmt.Println(result.String())  // → "Content"
```

---

## Path Composition

Combine syntax elements for complex queries.

### Precedence and Evaluation Order

Path segments are evaluated left-to-right:

```go
// Evaluation order:
// 1. root
// 2. items
// 3. item (with filter)
// 4. name child element

path := "root.items.item.#(price>100).name"
```

### Complex Path Examples

#### Example 1: Nested Arrays with Filters

```go
xml := `
<store>
    <departments>
        <department name="Electronics">
            <products>
                <product><name>Laptop</name><price>999</price><rating>4.5</rating></product>
                <product><name>Mouse</name><price>29</price><rating>4.2</rating></product>
            </products>
        </department>
        <department name="Books">
            <products>
                <product><name>Go Book</name><price>45</price><rating>4.9</rating></product>
            </products>
        </department>
    </departments>
</store>`

// Find highly-rated products in first department
path := "store.departments.department.0.products.product.#(rating>=4.5).name"
result := xmldot.Get(xml, path)
fmt.Println(result.String())  // → "Laptop"
```

#### Example 2: Wildcards with Filters and Modifiers

```go
// Find all products under $50, sorted by price
path := "store.departments.*.products.product.#(price<50)#.name|@sort"
```

#### Example 3: Recursive Wildcard with Attribute Filter

```go
// Find all active items anywhere in document
path := "root.**.#(@status==active)#"
```

#### Example 4: Filter with Manual Iteration and Modifiers

```go
// Note: Chained filters not supported - use manual filtering
// Find Engineering employees earning over $80k, sorted by salary descending
allEngineering := xmldot.Get(xml, "company.employees.employee.#(dept==Engineering)#")
var names []string
allEngineering.ForEach(func(i int, emp xmldot.Result) bool {
    salary := xmldot.Get(emp.Raw, "salary")
    if salary.Int() > 80000 {
        name := xmldot.Get(emp.Raw, "name")
        names = append(names, name.String())
    }
    return true
})
// Then sort names as needed
```

### Performance Optimization Tips

Optimize complex paths for better performance:

```go
// Slow: Recursive wildcard early in path
slowPath := "root.**.item.#(price<100)#"  // Searches entire tree

// Fast: Specific path with wildcard at end
fastPath := "root.catalog.*.item.#(price<100)#"  // Only searches catalog children

// Slow: Multiple nested filters
slowPath := "items.item.#(category.#(type==electronics))"

// Fast: Flatten filter conditions
fastPath := "items.item.#(@category==electronics)"  // Direct attribute check
```

### Path Length Considerations

```go
// Maximum 100 segments
const MaxPathSegments = 100

// Security: Paths exceeding limit are rejected
longPath := "root" + strings.Repeat(".child", 150)
result := xmldot.Get(xml, longPath)
fmt.Println(result.Exists())  // → false (exceeds limit)
```

---

## Performance Considerations

### Path Parsing Cache

XMLDOT automatically caches parsed paths for performance:

```go
// First call: parses path and caches it
result := xmldot.Get(xml, "root.child.element")  // ~500ns

// Subsequent calls: uses cached parse
result = xmldot.Get(xml, "root.child.element")   // ~200ns (85-91% faster)
```

Cache characteristics:
- Thread-safe LRU cache
- 256 path limit (oldest evicted when full)
- Automatic cache management
- No manual cache invalidation needed

### Expensive Operations

Relative performance of different operations:

| Operation | Typical Time | Notes |
|-----------|--------------|-------|
| Simple path | 200-500ns | `root.child.element` |
| Attribute access | 300-600ns | `element.@attr` |
| Array index | 400-800ns | `items.item.0` |
| Single wildcard | 500-1000ns | `root.*.name` |
| Array count | 500-1000ns | `items.item.#` |
| Numeric filter | 1-2µs | `item.#(price>100)` |
| String filter | 1-3µs | `item.#(name==value)` |
| Recursive wildcard | 2-10µs | `root.**.element` |
| Complex filter | 3-10µs | Multiple conditions |
| Modifier | +100-300ns | Per modifier |

### Best Practices for Performance

#### 1. Use Specific Paths

```go
// Prefer specific paths
xmldot.Get(xml, "catalog.products.product.0.name")  // Fast

// Avoid recursive wildcards when possible
xmldot.Get(xml, "catalog.**.name")  // Slower
```

#### 2. Use Batch Operations

```go
// Slow: Multiple Get calls
name := xmldot.Get(xml, "product.name")
price := xmldot.Get(xml, "product.price")
stock := xmldot.Get(xml, "product.stock")

// Fast: Single GetMany call
results := xmldot.GetMany(xml,
    "product.name",
    "product.price",
    "product.stock")
```

#### 3. Minimize Filter Complexity

```go
// Fast: Attribute filter (direct map lookup)
xmldot.Get(xml, "items.item.#(@id==5)")

// Slower: Element filter (requires parsing)
xmldot.Get(xml, "items.item.#(id==5)")
```

#### 4. Use Modifiers Efficiently

```go
// Efficient: Single modifier
xmldot.Get(xml, "items.item|@sort")

// Less efficient: Unnecessary chaining
xmldot.Get(xml, "items.item|@sort|@first")
// Better: Use array index
xmldot.Get(xml, "items.item.0")  // After knowing it's sorted
```

### Security Limits Summary

All security limits with performance impact:

| Limit | Default Value | Performance Impact |
|-------|---------------|-------------------|
| MaxDocumentSize | 10MB | Rejects large docs |
| MaxNestingDepth | 100 levels | Truncates deep nesting |
| MaxAttributes | 100 per element | Ignores excess attrs |
| MaxTokenSize | 1MB | Truncates large tokens |
| MaxPathSegments | 100 segments | Rejects long paths |
| MaxFilterDepth | 10 levels | Limits filter recursion |
| MaxFilterExpressionLength | 256 bytes | Rejects long filters |
| MaxWildcardResults | 1000 results | Caps wildcard matches |
| MaxRecursiveOperations | 10000 ops | Prevents CPU exhaustion |

For more details, see [docs/performance.md](performance.md) and [docs/security.md](security.md).

---

## Fluent API (v0.2.0+)

Starting in v0.2.0, Result objects support method chaining for cleaner code.

### Basic Chaining

```go
root := xmldot.Get(xml, "root")
user := root.Get("user")
name := user.Get("name").String()
```

Equivalent to:
```go
name := xmldot.Get(xml, "root.user.name").String()
```

### Deep Chaining

```go
fullPath := xmldot.Get(xml, "root").
    Get("company").
    Get("department").
    Get("team.member").
    Get("name").
    String()
```

### Batch Queries

```go
user := xmldot.Get(xml, "root.user")
results := user.GetMany("name", "age", "email")
name := results[0].String()
age := results[1].Int()
email := results[2].String()
```

### Array Handling

When querying arrays with fluent API, use explicit field extraction syntax:

```go
items := xmldot.Get(xml, "catalog.items")
prices := items.Get("item.#.price")  // Extract all prices
```

**Note**: The `Get()` method on Array types delegates to the first element.

### Performance Considerations

Fluent chaining adds overhead per call:
- 1-level chain: ~27.9% overhead
- 3-level chain: ~276.7% overhead

**Recommendation**: Use fluent API for readability, use full paths for performance-critical loops.

### Options Support

```go
opts := &xmldot.Options{CaseSensitive: false}
result := root.GetWithOptions("user.name", opts)
```

See the main documentation for Options details.

---

## Quick Reference

### Syntax Summary Table

| Syntax | Description | Example | Result |
|--------|-------------|---------|--------|
| `root.child` | Element path | `<root><child>val</child></root>` | "val" |
| `element.@attr` | Attribute | `<element attr="val"/>` | "val" |
| `items.item.0` | Array index | `<items><item>A</item><item>B</item></items>` | "A" |
| `items.item.#` | Array count | `<items><item>A</item><item>B</item></items>` | 2 |
| `element.%` | Text only | `<element>text<child/>more</element>` | "textmore" |
| `root.*` | Single wildcard | `<root><a>1</a><b>2</b></root>` | "1" |
| `root.**` | Recursive wildcard | Matches at any depth | First match |
| `item.#(price>100)` | Numeric filter | `<item><price>150</price></item>` | Element |
| `item.#(@id==5)` | Attribute filter | `<item id="5">val</item>` | Element |
| `item.#(@status)` | Exists check | `<item status="ok">val</item>` | Element |
| `path\|@reverse` | Modifier | Array of items | Reversed |
| `ns:element` | Namespace prefix | `<ns:element>val</ns:element>` | "val" |
| `element\\.name` | Escaped dot | `<element.name>val</element.name>` | "val" |

### Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal | `#(price==100)`, `#(name==value)` |
| `!=` | Not equal | `#(status!=inactive)` |
| `<` | Less than | `#(age<30)` |
| `>` | Greater than | `#(price>50)` |
| `<=` | Less than or equal | `#(quantity<=10)` |
| `>=` | Greater than or equal | `#(rating>=4.5)` |
| `%` | Pattern match | `#(name%"D*")` |
| `!%` | Pattern not match | `#(name!%"D*")` |

### Built-in Modifiers

| Modifier | Description | Example Result |
|----------|-------------|----------------|
| `@reverse` | Reverse array | [3, 2, 1] |
| `@sort` | Sort ascending | [1, 2, 3] |
| `@first` | First element | 1 |
| `@last` | Last element | 3 |
| `@flatten` | Flatten nested | [1, 2, 3, 4] |
| `@pretty` | Format XML | Indented XML |
| `@ugly` | Compact XML | Minified XML |
| `@raw` | Raw XML | Full element XML |
| `@keys` | Element names | ["name", "age"] |
| `@values` | Values only | ["John", "30"] |

### Common Patterns

```go
// Get first item in array
xmldot.Get(xml, "items.item.0")

// Get last item in array (calculate index)
count := xmldot.Get(xml, "items.item.#").Int()
xmldot.Get(xml, fmt.Sprintf("items.item.%d", count-1))

// Count items
xmldot.Get(xml, "items.item.#")

// Find by attribute (first match)
xmldot.Get(xml, "items.item.#(@id==5)")

// Find by condition (first match)
xmldot.Get(xml, "items.item.#(price<100)")

// Find all matches
xmldot.Get(xml, "items.item.#(price<100)#")

// Get all items (iterate)
items := xmldot.Get(xml, "items.item")
items.ForEach(func(i int, r Result) bool {
    // Process each item
    return true
})

// Multiple queries efficiently
results := xmldot.GetMany(xml, "path1", "path2", "path3")

// Complex query with modifiers
xmldot.Get(xml, "root.**.item.#(price>50)#.name|@sort|@first")
```

---

## See Also

- [Performance Guide](performance.md) - Optimization techniques and benchmarks
- [Migration Guide](migration.md) - Migrating from encoding/xml or GJSON/SJSON
- [Security Documentation](security.md) - Security best practices and limits
- [API Reference](https://godoc.org/github.com/netascode/xmldot) - Complete API documentation
- [Examples](../examples/) - Runnable code examples

---

**Document Version**: 1.1
**Last Updated**: 2025-10-18
**Status**: Complete

# Migration Guide

Guide for migrating to XMLDOT from encoding/xml or GJSON/SJSON.

## Table of Contents

1. [Introduction](#introduction)
2. [Migrating from encoding/xml](#migrating-from-encodingxml)
3. [GJSON/SJSON Users Guide](#gjsonsjson-users-guide)
4. [Common Pitfalls](#common-pitfalls)
5. [Feature Parity Matrix](#feature-parity-matrix)
6. [Gradual Migration Strategy](#gradual-migration-strategy)
7. [Getting Help](#getting-help)

---

## Introduction

This guide helps you migrate to XMLDOT from other XML or JSON processing libraries. We provide side-by-side comparisons, migration recipes, and common pitfall solutions.

### Who Is This Guide For?

- **encoding/xml users**: Go developers looking for simpler XML querying
- **GJSON/SJSON users**: JSON developers working with XML for the first time
- **XPath users**: Developers familiar with XPath-style querying

### Migration Strategies

**Complete Migration**:
- Replace all encoding/xml or GJSON usage with XMLDOT
- Best for new projects or simple XML processing

**Gradual Migration**:
- Migrate read operations first, keep write operations
- Best for large codebases with extensive XML processing

**Hybrid Approach**:
- Use both libraries for their strengths
- Best when both validation and flexible querying are needed

---

## Migrating from encoding/xml

### Feature Parity Matrix

| Feature | encoding/xml | xmldot | Notes |
|---------|--------------|--------|-------|
| Parse to struct | ✓ Unmarshal | ~ Manual | xmldot returns Result, not structs |
| Serialize from struct | ✓ Marshal | ~ SetRaw | xmldot focuses on modification |
| Query elements | ~ Requires struct | ✓ Path syntax | xmldot is simpler for ad-hoc queries |
| Modify XML | ~ Unmarshal/modify/Marshal | ✓ Set/Delete | xmldot is much simpler |
| Attributes | ✓ Full support | ✓ @attribute syntax | Both support attributes |
| Namespaces | ✓ Full xmlns support | ⚠️ Prefix only | encoding/xml has full support |
| Validation | ~ Struct tags | ✓ Valid() | Both provide basic validation |
| Streaming | ✓ Decoder/Encoder | ✗ Not yet | Use encoding/xml for streaming |
| Type safety | ✓ Compile-time | ~ Runtime | encoding/xml provides type safety |
| Performance (queries) | ~ Slower | ✓ Faster | xmldot optimized for queries |

### Common Patterns

#### Pattern 1: Extract Single Value

**encoding/xml**:
```go
type Root struct {
    User struct {
        Name string `xml:"name"`
        Age  int    `xml:"age"`
    } `xml:"user"`
}

xml := `<root><user><name>John</name><age>30</age></user></root>`

var root Root
if err := xml.Unmarshal([]byte(xml), &root); err != nil {
    return err
}
name := root.User.Name  // "John"
```

**xmldot**:
```go
xml := `<root><user><name>John</name><age>30</age></user></root>`

name := xmldot.Get(xml, "root.user.name")
fmt.Println(name.String())  // "John"
```

**Migration tip**: Use xmldot when you need one or two values, encoding/xml when you need the entire structure.

#### Pattern 2: Modify Existing Element

**encoding/xml**:
```go
type User struct {
    Name string `xml:"name"`
    Age  int    `xml:"age"`
}

xml := `<user><name>John</name><age>30</age></user>`

var user User
xml.Unmarshal([]byte(xml), &user)

user.Age = 31  // Modify

output, _ := xml.Marshal(user)
newXML := string(output)
// <user><name>John</name><age>31</age></user>
```

**xmldot**:
```go
xml := `<user><name>John</name><age>30</age></user>`

newXML, _ := xmldot.Set(xml, "user.age", 31)
// <user><name>John</name><age>31</age></user>
```

**Migration tip**: xmldot is much simpler for targeted modifications.

#### Pattern 3: Add New Element

**encoding/xml**:
```go
type User struct {
    Name  string `xml:"name"`
    Age   int    `xml:"age"`
    Email string `xml:"email,omitempty"`
}

xml := `<user><name>John</name><age>30</age></user>`

var user User
xml.Unmarshal([]byte(xml), &user)

user.Email = "john@example.com"  // Add email

output, _ := xml.Marshal(user)
newXML := string(output)
```

**xmldot**:
```go
xml := `<user><name>John</name><age>30</age></user>`

newXML, _ := xmldot.Set(xml, "user.email", "john@example.com")
// <user><name>John</name><age>30</age><email>john@example.com</email></user>
```

**Migration tip**: xmldot automatically creates missing elements.

#### Pattern 4: Iterate Over Array

**encoding/xml**:
```go
type Items struct {
    Items []string `xml:"item"`
}

xml := `<items><item>A</item><item>B</item><item>C</item></items>`

var items Items
xml.Unmarshal([]byte(xml), &items)

for _, item := range items.Items {
    fmt.Println(item)
}
```

**xmldot**:
```go
xml := `<items><item>A</item><item>B</item><item>C</item></items>`

items := xmldot.Get(xml, "items.item")
items.ForEach(func(i int, item Result) bool {
    fmt.Println(item.String())
    return true
})

// Or use Array()
for _, item := range items.Array() {
    fmt.Println(item.String())
}
```

**Migration tip**: Both approaches work well; choose based on whether you need struct type safety.

#### Pattern 5: Extract Attribute

**encoding/xml**:
```go
type Book struct {
    ID    int    `xml:"id,attr"`
    Title string `xml:",chardata"`
}

xml := `<book id="123">Go Programming</book>`

var book Book
xml.Unmarshal([]byte(xml), &book)
id := book.ID  // 123
```

**xmldot**:
```go
xml := `<book id="123">Go Programming</book>`

id := xmldot.Get(xml, "book.@id")
fmt.Println(id.Int())  // 123
```

**Migration tip**: xmldot uses `@` prefix for attributes, simpler syntax than struct tags.

#### Pattern 6: Conditional Extraction

**encoding/xml**:
```go
type Users struct {
    Users []struct {
        Name string `xml:"name"`
        Age  int    `xml:"age"`
    } `xml:"user"`
}

xml := `
<users>
    <user><name>Alice</name><age>28</age></user>
    <user><name>Bob</name><age>35</age></user>
    <user><name>Carol</name><age>42</age></user>
</users>`

var users Users
xml.Unmarshal([]byte(xml), &users)

for _, user := range users.Users {
    if user.Age > 30 {
        fmt.Println(user.Name)
    }
}
```

**xmldot**:
```go
xml := `
<users>
    <user><name>Alice</name><age>28</age></user>
    <user><name>Bob</name><age>35</age></user>
    <user><name>Carol</name><age>42</age></user>
</users>`

// Direct filtering in query (first match)
result := xmldot.Get(xml, "users.user.#(age>30).name")
fmt.Println(result.String())  // "Bob"

// Or iterate all matches
results := xmldot.Get(xml, "users.user.#(age>30)#")
results.ForEach(func(i int, user Result) bool {
    name := xmldot.Get(user.Raw, "name")
    fmt.Println(name.String())
    return true
})
```

**Migration tip**: xmldot filters are much more expressive than encoding/xml.

### Migration Recipes

#### Recipe 1: Convert Unmarshal to Get

**Before**:
```go
type Config struct {
    Database struct {
        Host string `xml:"host,attr"`
        Port int    `xml:"port,attr"`
        Name string `xml:"name"`
    } `xml:"database"`
}

var config Config
err := xml.Unmarshal(configXML, &config)
if err != nil {
    return err
}

dbHost := config.Database.Host
dbPort := config.Database.Port
dbName := config.Database.Name
```

**After**:
```go
results := xmldot.GetMany(configXML,
    "config.database.@host",
    "config.database.@port",
    "config.database.name")

dbHost := results[0].String()
dbPort := results[1].Int()
dbName := results[2].String()

// Or individual Gets
dbHost := xmldot.Get(configXML, "config.database.@host").String()
dbPort := xmldot.Get(configXML, "config.database.@port").Int()
dbName := xmldot.Get(configXML, "config.database.name").String()
```

#### Recipe 2: Convert Marshal to Set

**Before**:
```go
type Product struct {
    Name  string  `xml:"name"`
    Price float64 `xml:"price"`
    Stock int     `xml:"stock"`
}

product := Product{
    Name:  "Laptop",
    Price: 999.99,
    Stock: 10,
}

output, err := xml.Marshal(product)
if err != nil {
    return err
}
productXML := string(output)
```

**After**:
```go
// Start with empty or template XML
productXML := `<product></product>`

// Set values
productXML, _ = xmldot.Set(productXML, "product.name", "Laptop")
productXML, _ = xmldot.Set(productXML, "product.price", 999.99)
productXML, _ = xmldot.Set(productXML, "product.stock", 10)

// Or use SetMany
productXML, _ = xmldot.SetMany(productXML,
    []string{"product.name", "product.price", "product.stock"},
    []interface{}{"Laptop", 999.99, 10})
```

#### Recipe 3: Replace Decoder Loops with Wildcards

**Before**:
```go
decoder := xml.NewDecoder(reader)
for {
    token, err := decoder.Token()
    if err == io.EOF {
        break
    }
    if se, ok := token.(xml.StartElement); ok {
        if se.Name.Local == "item" {
            var item Item
            decoder.DecodeElement(&item, &se)
            processItem(item)
        }
    }
}
```

**After**:
```go
// Read entire document (for documents <10MB)
xmlData, _ := io.ReadAll(reader)

// Query all items
items := xmldot.Get(string(xmlData), "root.items.item")
items.ForEach(func(i int, item Result) bool {
    // Process each item
    name := xmldot.Get(item.Raw, "name").String()
    price := xmldot.Get(item.Raw, "price").Float()
    processItem(name, price)
    return true
})
```

**Note**: For files >10MB, continue using encoding/xml's streaming decoder.

#### Recipe 4: Hybrid Approach

**Combine both libraries for validation + flexible querying**:

```go
import (
    "encoding/xml"
    "github.com/netascode/xmldot"
)

// Step 1: Validate structure with encoding/xml
type Config struct {
    XMLName xml.Name `xml:"config"`
    Version string   `xml:"version,attr"`
    // ... other required fields
}

var config Config
if err := xml.Unmarshal(xmlData, &config); err != nil {
    return fmt.Errorf("invalid config structure: %w", err)
}

// Step 2: Use xmldot for flexible querying
xmlStr := string(xmlData)

// Query optional fields that may not be in struct
theme := xmldot.Get(xmlStr, "config.ui.theme").String()
if theme == "" {
    theme = "light"  // Default
}

// Query dynamic arrays
plugins := xmldot.Get(xmlStr, "config.plugins.plugin")
plugins.ForEach(func(i int, plugin Result) bool {
    name := xmldot.Get(plugin.Raw, "name").String()
    enabled := xmldot.Get(plugin.Raw.@enabled).Bool()
    if enabled {
        loadPlugin(name)
    }
    return true
})
```

### Limitations

#### No Struct Tags

**encoding/xml** struct tags are not supported:

```go
// encoding/xml supports:
type User struct {
    Name  string `xml:"name"`
    Age   int    `xml:"age,omitempty"`
    Email string `xml:"email,attr"`
}

// xmldot alternative: manual conversion
xml := `<user name="John"><age>30</age></user>`

name := xmldot.Get(xml, "user.name").String()
age := xmldot.Get(xml, "user.age").Int()
email := xmldot.Get(xml, "user.@email").String()

// Construct your struct manually
user := User{
    Name:  name,
    Age:   age,
    Email: email,
}
```

#### No Type Marshaling

**encoding/xml** automatically handles types:

```go
// encoding/xml converts types automatically:
type Data struct {
    Value int `xml:"value"`
}
xml.Marshal(Data{Value: 42})  // <Data><value>42</value></Data>

// xmldot uses interface{} with manual type conversion:
xml := `<data></data>`
xml, _ = xmldot.Set(xml, "data.value", 42)  // Converts int to string "42"

// When reading back:
value := xmldot.Get(xml, "data.value").Int()  // Converts back to int
```

#### No Schema Validation

Neither library provides schema validation:

```go
// For schema validation, use external tools:
// - libxml2 with Go bindings
// - xmlschema package
// - External schema validators

// Both libraries provide well-formedness validation:
if !xmldot.Valid(xml) {
    return errors.New("malformed XML")
}
```

#### No Full Namespace Support

**encoding/xml** has full xmlns support, **xmldot** only does prefix matching:

```go
xml := `
<root xmlns="http://example.com/default" xmlns:ns="http://example.com/ns">
    <item>Default NS</item>
    <ns:item>Explicit NS</ns:item>
</root>`

// encoding/xml understands namespaces semantically
type Root struct {
    Items []string `xml:"http://example.com/default item"`
}

// xmldot only matches prefixes literally
result := xmldot.Get(xml, "root.ns:item")  // Matches "ns:item" prefix
result := xmldot.Get(xml, "root.item")     // Matches unprefixed "item"

// Use encoding/xml for full namespace resolution
```

---

## GJSON/SJSON Users Guide

If you're coming from GJSON/SJSON for JSON, XMLDOT will feel familiar with XML-specific adaptations.

### JSON → XML Syntax Translation

| GJSON/SJSON | xmldot | Notes |
|-------------|--------|-------|
| `object.key` | `element.child` | Same dot notation |
| `array.0` | `items.item.0` | Array indexing (XML arrays are repeated elements) |
| `array.#` | `items.item.#` | Array count |
| `array.#.key` | `items.item.#` then iterate | XML arrays are different structure |
| `object.*.key` | `root.*.child` | Single-level wildcard |
| `object.**.key` | `root.**.child` | Recursive wildcard |
| `array.#(age>21)` | `array.item.#(age>21)` | Same filter syntax |
| `key\\.escaped` | `element\\.name` | Same escape syntax |
| `@reverse` | `\|@reverse` | Modifier syntax (pipe vs direct) |
| N/A | `element.@attribute` | XML-specific: attribute access |
| N/A | `element.%` | XML-specific: text content only |
| N/A | `ns:element` | XML-specific: namespace prefix |

### Key Differences

#### 1. No JSON Types

**GJSON** preserves JSON types (string, number, boolean, null, object, array):

```go
// GJSON
json := `{"age": 30, "active": true, "price": 99.99}`
age := gjson.Get(json, "age")
fmt.Println(age.Type)  // Number

// xmldot
xml := `<data><age>30</age><active>true</active><price>99.99</price></data>`
age := xmldot.Get(xml, "data.age")
fmt.Println(age.Type)  // String (XML has no numeric type)

// Use type conversion methods
fmt.Println(age.Int())    // 30
fmt.Println(age.String()) // "30"
```

#### 2. Attributes Are XML-Specific

**GJSON** has no concept of attributes:

```go
// JSON has no attributes
json := `{"id": "123", "name": "Item"}`
id := gjson.Get(json, "id")  // Just another key

// XML distinguishes attributes
xml := `<item id="123"><name>Item</name></item>`
id := xmldot.Get(xml, "item.@id")      // Attribute
name := xmldot.Get(xml, "item.name")   // Element
```

#### 3. Array Detection Differences

**GJSON** arrays are explicit `[...]`:

```go
// JSON arrays
json := `{"items": ["A", "B", "C"]}`
item := gjson.Get(json, "items.0")  // "A"

// XML arrays are repeated elements
xml := `<items><item>A</item><item>B</item><item>C</item></items>`
item := xmldot.Get(xml, "items.item.0")  // "A"
```

#### 4. Text Content vs Elements

**GJSON** doesn't have mixed content:

```go
// JSON: text is always a value
json := `{"paragraph": "This is text with no child elements."}`

// XML: mixed content exists
xml := `<paragraph>This is <bold>bold</bold> text.</paragraph>`

// Get full content (including child element text)
full := xmldot.Get(xml, "paragraph")
fmt.Println(full.String())  // "This is bold text."

// Get only direct text nodes (% modifier)
text := xmldot.Get(xml, "paragraph.%")
fmt.Println(text.String())  // "This is  text."
```

#### 5. Namespace Prefixes (XML-Specific)

**GJSON** has no namespaces:

```go
// JSON: no namespaces
json := `{"soap:Envelope": {"soap:Body": "data"}}`  // Just key names

// XML: namespace prefixes
xml := `
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>data</soap:Body>
</soap:Envelope>`

result := xmldot.Get(xml, "soap:Envelope.soap:Body")
```

### Translation Examples

#### Example 1: Simple Query

**GJSON**:
```go
json := `{"user": {"name": "John", "age": 30}}`
name := gjson.Get(json, "user.name")
fmt.Println(name.String())  // "John"
```

**xmldot**:
```go
xml := `<user><name>John</name><age>30</age></user>`
name := xmldot.Get(xml, "user.name")
fmt.Println(name.String())  // "John"
```

#### Example 2: Array Access

**GJSON**:
```go
json := `{"items": ["A", "B", "C"]}`
second := gjson.Get(json, "items.1")
count := gjson.Get(json, "items.#")
fmt.Println(second.String())  // "B"
fmt.Println(count.Int())      // 3
```

**xmldot**:
```go
xml := `<items><item>A</item><item>B</item><item>C</item></items>`
second := xmldot.Get(xml, "items.item.1")
count := xmldot.Get(xml, "items.item.#")
fmt.Println(second.String())  // "B"
fmt.Println(count.Int())      // 3
```

#### Example 3: Wildcard Query

**GJSON**:
```go
json := `{"users": {"alice": {"age": 28}, "bob": {"age": 35}}}`
ages := gjson.Get(json, "users.*.age")
// Returns array of ages
```

**xmldot**:
```go
xml := `
<users>
    <alice><age>28</age></alice>
    <bob><age>35</age></bob>
</users>`
ages := xmldot.Get(xml, "users.*.age")
ages.ForEach(func(i int, age Result) bool {
    fmt.Println(age.Int())
    return true
})
```

#### Example 4: Filter Query

**GJSON**:
```go
json := `{
    "users": [
        {"name": "Alice", "age": 28},
        {"name": "Bob", "age": 35}
    ]
}`
result := gjson.Get(json, "users.#(age>30).name")
fmt.Println(result.String())  // "Bob"
```

**xmldot**:
```go
xml := `
<users>
    <user><name>Alice</name><age>28</age></user>
    <user><name>Bob</name><age>35</age></user>
</users>`
result := xmldot.Get(xml, "users.user.#(age>30).name")
fmt.Println(result.String())  // "Bob"
```

#### Example 5: Modifier Usage

**GJSON**:
```go
json := `{"items": ["C", "A", "B"]}`
sorted := gjson.Get(json, "items.@sort")
// Returns sorted array
```

**xmldot**:
```go
xml := `<items><item>C</item><item>A</item><item>B</item></items>`
sorted := xmldot.Get(xml, "items.item|@sort")
sorted.ForEach(func(i int, item Result) bool {
    fmt.Println(item.String())
    return true
})
// "A", "B", "C"
```

#### Example 6: Set Operation

**SJSON**:
```go
json := `{"user": {"name": "John"}}`
updated := sjson.Set(json, "user.age", 30)
// {"user": {"name": "John", "age": 30}}
```

**xmldot**:
```go
xml := `<user><name>John</name></user>`
updated, _ := xmldot.Set(xml, "user.age", 30)
// <user><name>John</name><age>30</age></user>
```

#### Example 7: Delete Operation

**SJSON**:
```go
json := `{"user": {"name": "John", "age": 30}}`
updated := sjson.Delete(json, "user.age")
// {"user": {"name": "John"}}
```

**xmldot**:
```go
xml := `<user><name>John</name><age>30</age></user>`
updated, _ := xmldot.Delete(xml, "user.age")
// <user><name>John</name></user>
```

#### Example 8: Batch Operations

**SJSON** doesn't have batch operations, **xmldot** does:

```go
// xmldot provides efficient batch operations
xml := `<user><name>John</name></user>`

updated, _ := xmldot.SetMany(xml,
    []string{"user.age", "user.email", "user.active"},
    []interface{}{30, "john@example.com", true})

// <user>
//   <name>John</name>
//   <age>30</age>
//   <email>john@example.com</email>
//   <active>true</active>
// </user>
```

---

## Common Pitfalls

### Pitfall 1: Expecting JSON Types in Result

**Problem**:
```go
xml := `<data><value>42</value></data>`
result := xmldot.Get(xml, "data.value")
// result.Type is String, not Number
```

**Solution**:
```go
// Use type conversion methods
value := result.Int()     // 42
value := result.Float()   // 42.0
value := result.String()  // "42"
```

### Pitfall 2: Forgetting Attribute `@` Prefix

**Problem**:
```go
xml := `<book id="123">Title</book>`
id := xmldot.Get(xml, "book.id")
fmt.Println(id.Exists())  // false - looking for <id> element
```

**Solution**:
```go
id := xmldot.Get(xml, "book.@id")  // Use @ prefix for attributes
fmt.Println(id.String())  // "123"
```

### Pitfall 3: Array Detection Differences

**Problem**:
```go
// Single element is not an array in JSON
json := `{"items": ["A"]}`  // Still an array

// Single element in XML is not automatically an array
xml := `<items><item>A</item></items>`
item := xmldot.Get(xml, "items.item")
fmt.Println(item.IsArray())  // false
```

**Solution**:
```go
// Access with index still works for single element
item := xmldot.Get(xml, "items.item.0")
fmt.Println(item.String())  // "A"

// Count works
count := xmldot.Get(xml, "items.item.#")
fmt.Println(count.Int())  // 1

// Use Array() method for consistent handling
items := xmldot.Get(xml, "items.item")
for _, item := range items.Array() {
    fmt.Println(item.String())  // Works for single or multiple
}
```

### Pitfall 4: Text Content with Child Elements

**Problem**:
```go
xml := `<para>Text <bold>bold</bold> more</para>`
text := xmldot.Get(xml, "para")
fmt.Println(text.String())  // "Text bold more" (includes child text)
```

**Solution**:
```go
// Use % to get only direct text nodes
text := xmldot.Get(xml, "para.%")
fmt.Println(text.String())  // "Text  more" (excludes child text)
```

### Pitfall 5: Namespace Handling

**Problem**:
```go
xml := `<ns:root xmlns:ns="http://example.com"><ns:item>Value</ns:item></ns:root>`

// Forgetting namespace prefix
result := xmldot.Get(xml, "root.item")
fmt.Println(result.Exists())  // false or unexpected
```

**Solution**:
```go
// Include namespace prefix in path
result := xmldot.Get(xml, "ns:root.ns:item")
fmt.Println(result.String())  // "Value"

// Or query by local name (backward compatible)
result := xmldot.Get(xml, "root.item")  // May work if no conflicts
```

**Warning**: XMLDOT only does prefix matching, not semantic namespace resolution. For full xmlns support, use encoding/xml.

### Pitfall 6: Path Escaping Requirements

**Problem**:
```go
xml := `<data><file.name>document.pdf</file.name></data>`
result := xmldot.Get(xml, "data.file.name")
fmt.Println(result.Exists())  // false - split on dot
```

**Solution**:
```go
// Escape dots in element names
result := xmldot.Get(xml, "data.file\\.name")
fmt.Println(result.String())  // "document.pdf"
```

### Pitfall 7: Security Limits in Untrusted Input

**Problem**:
```go
// User provides large document
userXML := generateLargeXML(20 * 1024 * 1024)  // 20MB
result := xmldot.Get(userXML, path)
// Returns Null - exceeds MaxDocumentSize (10MB)
```

**Solution**:
```go
// Check size before processing
if len(userXML) > xmldot.MaxDocumentSize {
    return errors.New("document too large")
}

// Or validate first
if !xmldot.Valid(userXML) {
    return errors.New("invalid XML")
}

result := xmldot.Get(userXML, path)
```

### Pitfall 8: Optimistic Mode Risks

**Problem**:
```go
// Using Optimistic mode with untrusted input
opts := &xmldot.Options{Optimistic: true}
result, _ := xmldot.SetWithOptions(userXML, path, value, opts)
// Skips validation - may produce invalid XML
```

**Solution**:
```go
// Only use Optimistic mode with trusted input
if isTrustedSource(xmlSource) {
    opts := &xmldot.Options{Optimistic: true}
    result, _ := xmldot.SetWithOptions(xml, path, value, opts)
} else {
    // Default mode validates
    result, _ := xmldot.Set(xml, path, value)
}
```

---

## Feature Parity Matrix

Comprehensive comparison of features across libraries:

| Feature | xmldot | encoding/xml | GJSON | SJSON |
|---------|--------|--------------|-------|-------|
| **Core Operations** |
| Query elements | ✅ | ⚠️ Via structs | ✅ | N/A |
| Modify values | ✅ | ⚠️ Via structs | N/A | ✅ |
| Delete elements | ✅ | ⚠️ Via structs | N/A | ✅ |
| Batch operations | ✅ GetMany/SetMany | ❌ | ❌ | ❌ |
| **Path Syntax** |
| Dot notation | ✅ | ❌ | ✅ | ✅ |
| Array indexing | ✅ | ⚠️ Via slice | ✅ | ✅ |
| Wildcards (*) | ✅ | ❌ | ✅ | ❌ |
| Recursive (**) | ✅ | ❌ | ✅ | ❌ |
| Filters | ✅ | ❌ | ✅ | ❌ |
| Modifiers | ✅ | ❌ | ✅ | ❌ |
| **XML Features** |
| Attributes | ✅ @attr | ✅ Tags | N/A | N/A |
| Namespaces | ⚠️ Prefix only | ✅ Full | N/A | N/A |
| Mixed content | ✅ % syntax | ✅ | N/A | N/A |
| CDATA | ✅ | ✅ | N/A | N/A |
| Comments | ⚠️ Skipped | ✅ | N/A | N/A |
| Processing instructions | ⚠️ Skipped | ✅ | N/A | N/A |
| **Type System** |
| Type safety | ⚠️ Runtime | ✅ Compile-time | ⚠️ Runtime | ⚠️ Runtime |
| Type conversion | ✅ Int/Float/Bool | ✅ Struct tags | ✅ | ✅ |
| Custom types | ⚠️ Manual | ✅ Unmarshal | N/A | N/A |
| **Performance** |
| Query speed | ✅ Fast | ⚠️ Moderate | ✅ Fast | ✅ Fast |
| Modification speed | ✅ Fast | ⚠️ Slow | N/A | ✅ Fast |
| Memory efficiency | ✅ Low alloc | ⚠️ Moderate | ✅ Low alloc | ✅ Low alloc |
| Streaming | ❌ Not yet | ✅ Decoder | ❌ | ❌ |
| **Security** |
| XXE protection | ✅ | ⚠️ Manual | N/A | N/A |
| Entity expansion | ✅ Disabled | ⚠️ Manual | N/A | N/A |
| DoS limits | ✅ Built-in | ⚠️ Manual | ⚠️ Manual | ⚠️ Manual |
| **Validation** |
| Well-formedness | ✅ Valid() | ✅ | ❌ | ❌ |
| Schema validation | ❌ | ❌ | ❌ | ❌ |

Legend:
- ✅ Full support
- ⚠️ Partial support or manual implementation required
- ❌ Not supported
- N/A Not applicable

---

## Gradual Migration Strategy

### Phase 1: Add XMLDOT Alongside Existing Code

**Goal**: Introduce XMLDOT without breaking existing functionality.

```go
import (
    "encoding/xml"
    "github.com/netascode/xmldot"
)

// Keep existing encoding/xml code
var config Config
xml.Unmarshal(xmlData, &config)

// Add xmldot for new queries
theme := xmldot.Get(string(xmlData), "config.ui.theme").String()
```

**Testing**: Parallel runs comparing results to ensure correctness.

### Phase 2: Migrate Read-Heavy Code Paths

**Goal**: Replace encoding/xml Unmarshal with XMLDOT Get for read operations.

```go
// Before: encoding/xml
type User struct {
    Name  string `xml:"name"`
    Email string `xml:"email"`
}
var user User
xml.Unmarshal(userData, &user)

// After: xmldot
userStr := string(userData)
name := xmldot.Get(userStr, "user.name").String()
email := xmldot.Get(userStr, "user.email").String()
```

**Benefits**: Simpler code, better performance for single-value extraction.

### Phase 3: Migrate Write Operations

**Goal**: Replace encoding/xml Marshal with XMLDOT Set/Delete.

```go
// Before: encoding/xml
var user User
xml.Unmarshal(userData, &user)
user.Email = "new@example.com"
newData, _ := xml.Marshal(user)

// After: xmldot
userStr := string(userData)
newUserStr, _ := xmldot.Set(userStr, "user.email", "new@example.com")
newData := []byte(newUserStr)
```

**Benefits**: Much simpler for targeted modifications.

### Phase 4: Remove Old Library Dependencies

**Goal**: Complete migration, remove encoding/xml where no longer needed.

```go
// Remove encoding/xml import
// import "encoding/xml"  // No longer needed

import "github.com/netascode/xmldot"

// All operations use xmldot
name := xmldot.Get(xml, "user.name").String()
xml, _ = xmldot.Set(xml, "user.age", 31)
xml, _ = xmldot.Delete(xml, "user.deprecated")
```

**Benefits**: Cleaner dependency tree, consistent API.

### Testing Strategies During Migration

#### 1. Parallel Testing

```go
func TestMigrationCompatibility(t *testing.T) {
    xml := loadTestXML()

    // encoding/xml result
    var oldResult OldStruct
    xml.Unmarshal([]byte(xml), &oldResult)

    // xmldot result
    newName := xmldot.Get(xml, "path.to.name").String()
    newAge := xmldot.Get(xml, "path.to.age").Int()

    // Compare results
    if newName != oldResult.Name {
        t.Errorf("name mismatch: %v != %v", newName, oldResult.Name)
    }
    if newAge != oldResult.Age {
        t.Errorf("age mismatch: %v != %v", newAge, oldResult.Age)
    }
}
```

#### 2. Incremental Rollout

```go
// Use feature flag for gradual rollout
func processXML(xml string) Result {
    if featureFlags.UseGoXML {
        return processWithGoXML(xml)
    }
    return processWithEncodingXML(xml)
}

// Monitor metrics: performance, error rates, correctness
```

#### 3. Rollback Plan

Keep both implementations available during migration:

```go
func processXML(xml string) (Result, error) {
    result, err := processWithGoXML(xml)
    if err != nil {
        log.Warn("xmldot failed, falling back to encoding/xml")
        return processWithEncodingXML(xml)
    }
    return result, nil
}
```

---

## Getting Help

### Documentation Resources

- **Path Syntax**: [docs/path-syntax.md](path-syntax.md)
- **Performance Guide**: [docs/performance.md](performance.md)
- **Security Documentation**: [docs/security.md](security.md)
- **API Reference**: [godoc](https://godoc.org/github.com/netascode/xmldot)
- **Examples**: [examples/](../examples/)

### Community Support

- **GitHub Issues**: [github.com/netascode/xmldot/issues](https://github.com/netascode/xmldot/issues)
- **GitHub Discussions**: [github.com/netascode/xmldot/discussions](https://github.com/netascode/xmldot/discussions)

---

## Summary

XMLDOT provides a simpler, faster alternative to encoding/xml for ad-hoc querying and modification:

**Choose xmldot when:**
- ✓ Quick value extraction from XML
- ✓ Simple XML modifications
- ✓ Dynamic/unknown XML structures
- ✓ Performance-critical read operations
- ✓ Untrusted XML input (secure-by-default)

**Keep encoding/xml when:**
- ✓ Struct type safety is critical
- ✓ Full namespace support needed
- ✓ Streaming large files (>10MB)
- ✓ Complex schema validation required

**Best approach**: Use both libraries for their strengths in a hybrid approach.

---

**Document Version**: 1.0
**Last Updated**: 2025-10-08
**Status**: Complete

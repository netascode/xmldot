# Namespace Handling

**Complexity Level**: Advanced
**Estimated Time**: 15 minutes
**Prerequisites**: Understanding of XML namespaces

## What You'll Learn

- Query elements with namespace prefixes
- Understand prefix matching behavior
- Work with multiple namespace prefixes
- Recognize namespace handling limitations
- Decide when to use xmldot vs encoding/xml
- Implement hybrid approaches for complex namespace scenarios

## Running the Example

```bash
cd examples/namespaces
go run main.go
```

## Expected Output

```
Namespace Handling Example
===========================

Example 1: Query elements with namespace prefix
Found soap:Body element

Example 2: Query nested namespace elements
Stock name: AAPL

Example 3: Query authentication token
Auth token: abc123

Example 4: RSS with mixed namespaces
Feed title: Example Feed
Atom link: http://example.org/feed

Example 5: Namespace limitations
IMPORTANT: xmldot uses prefix matching, not full namespace resolution

✓ Prefix matching works: USD
✗ Namespace URI resolution not supported
  (use encoding/xml for full namespace support)

Example 6: When xmldot namespace support is sufficient
✓ SOAP APIs with consistent prefixes (soap:, m:, etc.)
✓ RSS/Atom feeds with standard prefixes
✓ Configuration files with simple namespaces
✓ Documents where you control namespace prefixes

Example 7: When to use encoding/xml instead
✗ Default namespaces (xmlns="..." without prefix)
✗ Dynamic/unknown namespace prefixes
✗ Namespace validation requirements
✗ Complex XPath queries with namespace axes

Example 8: Hybrid approach (encoding/xml + xmldot)
1. Use encoding/xml to unmarshal and normalize namespaces
2. Use xmldot for querying and manipulation
3. Use encoding/xml to marshal back if needed

See real-world/soap-client example for hybrid implementation
```

## Key Concepts

### Prefix Matching

xmldot uses **prefix matching**, not full namespace resolution:

```go
// This works (prefix matching):
xmldot.Get(xml, "soap:Envelope.soap:Body")

// This does NOT work (namespace URI resolution):
// xmldot.Get(xml, "{http://schemas.xmlsoap.org/soap/envelope/}Body")
```

### Namespace Syntax

Include prefixes in path segments:
```
prefix:element.prefix:child.prefix:grandchild
```

Mixed prefixes:
```
soap:Envelope.soap:Body.m:GetStockPrice.m:StockName
```

### Limitations

**What xmldot supports**:
- Namespace prefixes in queries
- Multiple different prefixes
- Mixed prefixed/non-prefixed elements

**What xmldot does NOT support**:
- Default namespaces (xmlns="...")
- Namespace URI resolution
- Namespace validation
- Dynamic prefix discovery
- XPath namespace axes

## Code Walkthrough

The example demonstrates namespace capabilities and limitations:

1. **Basic Prefix Query**: Access SOAP body using `soap:` prefix
2. **Nested Prefixes**: Navigate through multiple namespace levels
3. **Multiple Namespaces**: Handle different prefixes in single query
4. **Mixed Content**: Combine prefixed and non-prefixed elements
5. **Limitations**: Demonstrate what prefix matching cannot do
6. **Sufficient Cases**: List scenarios where xmldot works well
7. **Insufficient Cases**: List scenarios requiring encoding/xml
8. **Hybrid Approach**: Suggest combining libraries for complex cases

## Common Pitfalls

- **Pitfall**: Expecting default namespace handling
  - **Solution**: Use encoding/xml for default namespaces

- **Pitfall**: Using namespace URIs in queries
  - **Solution**: Use prefixes only: `soap:Body`, not `{uri}Body`

- **Pitfall**: Dynamic prefix scenarios
  - **Solution**: Normalize prefixes first or use encoding/xml

- **Pitfall**: Namespace validation requirements
  - **Solution**: Use encoding/xml for schema validation

## When to Use xmldot

xmldot namespace support is sufficient when:

1. **Predictable Prefixes**: SOAP APIs with standard prefixes
2. **Simple Namespaces**: RSS/Atom feeds with atom:link patterns
3. **Configuration Files**: Internal configs with known structure
4. **Controlled Documents**: You generate the XML and control prefixes

**Example scenarios**:
- SOAP web service clients with fixed WSDL
- RSS feed parsing with standard atom: prefix
- Configuration files with xmlns:app="..." prefixes
- Internal APIs with consistent namespace conventions

## When to Use encoding/xml

Use encoding/xml instead when you need:

1. **Default Namespaces**: `xmlns="http://..."` without prefix
2. **Dynamic Prefixes**: Prefixes vary between documents
3. **Namespace Validation**: Verify namespace URIs match schema
4. **Complex XPath**: Namespace axes, URI-based selection
5. **Strict Compliance**: W3C namespace specification compliance

**Example scenarios**:
- XML documents with default namespaces
- Third-party XML with unpredictable prefixes
- Schema validation requirements
- Full XPath 2.0/3.0 query support

## Hybrid Approach

Combine both libraries for best results:

```go
// Step 1: Unmarshal with encoding/xml
type Envelope struct {
    XMLName xml.Name `xml:"Envelope"`
    Body    Body     `xml:"Body"`
}

var env Envelope
xml.Unmarshal([]byte(soapXML), &env)

// Step 2: Marshal to normalized XML
normalized, _ := xml.Marshal(env)

// Step 3: Use xmldot for queries
result := xmldot.Get(string(normalized), "Envelope.Body.GetStockPrice.StockName")
```

This approach provides:
- Full namespace resolution (encoding/xml)
- Fast querying (xmldot)
- Type safety (encoding/xml structs)
- Flexibility (xmldot path syntax)

## Next Steps

- [Real-World SOAP Client](../real-world/soap-client/) - Namespace handling in practice
- [Performance](../performance/) - Optimize namespace queries
- [encoding/xml docs](https://pkg.go.dev/encoding/xml) - Full namespace support

## See Also

- [Main README](../../README.md) - Complete documentation
- [API Reference](https://pkg.go.dev/github.com/netascode/xmldot) - Full API documentation
- [W3C Namespaces Spec](https://www.w3.org/TR/xml-names/) - XML namespace specification

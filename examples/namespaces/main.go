package main

import (
	"fmt"

	xmldot "github.com/netascode/xmldot"
)

// Sample XML: SOAP envelope with namespaces
const soapXML = `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	<soap:Header>
		<auth:Authentication xmlns:auth="http://example.org/auth">
			<auth:Token>abc123</auth:Token>
		</auth:Authentication>
	</soap:Header>
	<soap:Body>
		<m:GetStockPrice xmlns:m="http://www.example.org/stock">
			<m:StockName>AAPL</m:StockName>
			<m:Currency>USD</m:Currency>
		</m:GetStockPrice>
	</soap:Body>
</soap:Envelope>`

// Sample XML: RSS feed with default namespace
const rssXML = `<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
	<channel>
		<title>Example Feed</title>
		<atom:link href="http://example.org/feed" rel="self" type="application/rss+xml"/>
		<item>
			<title>Post 1</title>
		</item>
	</channel>
</rss>`

func main() {
	fmt.Println("Namespace Handling Example")
	fmt.Println("===========================\n")

	// Example 1: Query with prefix (soap:Body)
	fmt.Println("Example 1: Query elements with namespace prefix")
	result := xmldot.Get(soapXML, "soap:Envelope.soap:Body")
	if result.Exists() {
		fmt.Println("Found soap:Body element")
	}
	fmt.Println()

	// Example 2: Nested prefixes (m:StockName)
	fmt.Println("Example 2: Query nested namespace elements")
	result = xmldot.Get(soapXML, "soap:Envelope.soap:Body.m:GetStockPrice.m:StockName")
	fmt.Printf("Stock name: %s\n\n", result.String())

	// Example 3: Multiple namespace prefixes
	fmt.Println("Example 3: Query authentication token")
	result = xmldot.Get(soapXML, "soap:Envelope.soap:Header.auth:Authentication.auth:Token")
	fmt.Printf("Auth token: %s\n\n", result.String())

	// Example 4: Mixing prefixed and non-prefixed elements
	fmt.Println("Example 4: RSS with mixed namespaces")
	result = xmldot.Get(rssXML, "rss.channel.title")
	fmt.Printf("Feed title: %s\n", result.String())

	result = xmldot.Get(rssXML, "rss.channel.atom:link.@href")
	fmt.Printf("Atom link: %s\n\n", result.String())

	// Example 5: Limitations demonstration
	fmt.Println("Example 5: Namespace limitations")
	fmt.Println("IMPORTANT: xmldot uses prefix matching, not full namespace resolution")
	fmt.Println()

	// This works (prefix matching):
	result = xmldot.Get(soapXML, "soap:Body.m:GetStockPrice.m:Currency")
	fmt.Printf("✓ Prefix matching works: %s\n", result.String())

	// This would NOT work (namespace URI resolution):
	// Cannot query by namespace URI like: {http://www.example.org/stock}StockName
	fmt.Println("✗ Namespace URI resolution not supported")
	fmt.Println("  (use encoding/xml for full namespace support)")
	fmt.Println()

	// Example 6: When prefixes are predictable
	fmt.Println("Example 6: When xmldot namespace support is sufficient")
	fmt.Println("✓ SOAP APIs with consistent prefixes (soap:, m:, etc.)")
	fmt.Println("✓ RSS/Atom feeds with standard prefixes")
	fmt.Println("✓ Configuration files with simple namespaces")
	fmt.Println("✓ Documents where you control namespace prefixes")
	fmt.Println()

	// Example 7: When to avoid xmldot
	fmt.Println("Example 7: When to use encoding/xml instead")
	fmt.Println("✗ Default namespaces (xmlns=\"...\" without prefix)")
	fmt.Println("✗ Dynamic/unknown namespace prefixes")
	fmt.Println("✗ Namespace validation requirements")
	fmt.Println("✗ Complex XPath queries with namespace axes")
	fmt.Println()

	// Example 8: Hybrid approach
	fmt.Println("Example 8: Hybrid approach (encoding/xml + xmldot)")
	fmt.Println("1. Use encoding/xml to unmarshal and normalize namespaces")
	fmt.Println("2. Use xmldot for querying and manipulation")
	fmt.Println("3. Use encoding/xml to marshal back if needed")
	fmt.Println("\nSee real-world/soap-client example for hybrid implementation")
}

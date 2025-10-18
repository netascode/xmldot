// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"testing"
)

// Test data of varying complexity

var smallXML = `<root>
	<user>
		<name>John Doe</name>
		<age>30</age>
		<email>john@example.com</email>
	</user>
</root>`

var mediumXML = `<root>
	<users>
		<user id="1">
			<name>John Doe</name>
			<age>30</age>
			<email>john@example.com</email>
			<address>
				<street>123 Main St</street>
				<city>Springfield</city>
				<zip>12345</zip>
			</address>
		</user>
		<user id="2">
			<name>Jane Smith</name>
			<age>25</age>
			<email>jane@example.com</email>
			<address>
				<street>456 Oak Ave</street>
				<city>Springfield</city>
				<zip>12345</zip>
			</address>
		</user>
		<user id="3">
			<name>Bob Johnson</name>
			<age>35</age>
			<email>bob@example.com</email>
			<address>
				<street>789 Elm St</street>
				<city>Springfield</city>
				<zip>12345</zip>
			</address>
		</user>
	</users>
	<products>
		<product id="101">
			<name>Widget</name>
			<price>19.99</price>
			<stock>100</stock>
		</product>
		<product id="102">
			<name>Gadget</name>
			<price>29.99</price>
			<stock>50</stock>
		</product>
	</products>
</root>`

var largeXML string

func init() {
	// Generate large XML with 100 items
	var builder string
	builder = "<root><items>"
	for i := 0; i < 100; i++ {
		builder += fmt.Sprintf(`
		<item id="%d">
			<name>Item %d</name>
			<description>This is a detailed description of item number %d with lots of text to simulate real-world data</description>
			<price>%d.99</price>
			<category>Category %d</category>
			<tags>
				<tag>tag1</tag>
				<tag>tag2</tag>
				<tag>tag3</tag>
			</tags>
			<metadata>
				<created>2024-01-01</created>
				<modified>2024-01-15</modified>
				<author>System</author>
			</metadata>
		</item>`, i, i, i, 10+(i%100), i%10)
	}
	builder += "</items></root>"
	largeXML = builder
}

var deepXML string

func init() {
	// Generate deeply nested XML (20 levels)
	builder := "<root>"
	for i := 0; i < 20; i++ {
		builder += fmt.Sprintf("<level%d>", i)
	}
	builder += "<value>deep content</value>"
	for i := 19; i >= 0; i-- {
		builder += fmt.Sprintf("</level%d>", i)
	}
	builder += "</root>"
	deepXML = builder
}

// ============================================================================
// Basic Operations Benchmarks
// ============================================================================

func BenchmarkGet_SimpleElement(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(smallXML, "root.user.name")
	}
}

func BenchmarkGet_DeeplyNested(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(deepXML, "root.level0.level1.level2.level3.level4.level5.level6.level7.level8.level9.level10.level11.level12.level13.level14.level15.level16.level17.level18.level19.value")
	}
}

func BenchmarkGet_Attribute(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.users.user.@id")
	}
}

func BenchmarkGet_ArrayIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.users.user.1.name")
	}
}

func BenchmarkGet_ArrayCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.users.user.#")
	}
}

func BenchmarkGet_TextContent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.users.user.name.%")
	}
}

func BenchmarkGetMany(b *testing.B) {
	paths := []string{
		"root.user.name",
		"root.user.age",
		"root.user.email",
	}
	for i := 0; i < b.N; i++ {
		_ = GetMany(smallXML, paths...)
	}
}

// ============================================================================
// Advanced Operations Benchmarks
// ============================================================================

func BenchmarkGet_SingleWildcard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.users.*.name")
	}
}

func BenchmarkGet_RecursiveWildcard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.**.price")
	}
}

func BenchmarkGet_RecursiveWildcard_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(largeXML, "root.**.tag")
	}
}

func BenchmarkGet_FilterNumeric(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.users.user[age>25]")
	}
}

func BenchmarkGet_FilterString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.users.user[name==Jane Smith]")
	}
}

func BenchmarkGet_FilterAttribute(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.products.product[@id==101]")
	}
}

func BenchmarkGet_WildcardWithFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.*.user[age>25].name")
	}
}

func BenchmarkGet_FilterExists(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.users.user[@id]")
	}
}

// ============================================================================
// Write Operations Benchmarks
// ============================================================================

func BenchmarkSet_Element(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Set(smallXML, "root.user.name", "Jane Doe")
	}
}

func BenchmarkSet_AttributeValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Set(mediumXML, "root.users.user.@id", "999")
	}
}

func BenchmarkSet_CreateNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Set(smallXML, "root.user.newfield", "new value")
	}
}

func BenchmarkSet_ComplexValue(b *testing.B) {
	complexValue := "<address><street>New St</street><city>New City</city></address>"
	for i := 0; i < b.N; i++ {
		_, _ = SetRaw(mediumXML, "root.users.user.address", complexValue)
	}
}

func BenchmarkDelete_ElementNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Delete(smallXML, "root.user.email")
	}
}

func BenchmarkDelete_AttributeValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Delete(mediumXML, "root.users.user.@id")
	}
}

// ============================================================================
// Parser Benchmarks
// ============================================================================

func BenchmarkParser_SkipWhitespace(b *testing.B) {
	data := []byte("    \t\n\r    <element>")
	for i := 0; i < b.N; i++ {
		p := newXMLParser(data)
		p.skipWhitespace()
	}
}

func BenchmarkParser_ParseAttributes(b *testing.B) {
	data := []byte(`id="123" name="test" value="data" enabled="true">`)
	for i := 0; i < b.N; i++ {
		p := newXMLParser(data)
		_ = p.parseAttributes()
	}
}

func BenchmarkParser_ParseElement(b *testing.B) {
	data := []byte(`<user id="1"><name>John</name><age>30</age></user>`)
	for i := 0; i < b.N; i++ {
		p := newXMLParser(data)
		p.next() // skip '<'
		_, _, _ = p.parseElementName()
	}
}

func BenchmarkParser_ParseElementContent(b *testing.B) {
	data := []byte(`<name>John Doe</name><age>30</age></user>`)
	for i := 0; i < b.N; i++ {
		p := newXMLParser(data)
		_ = p.parseElementContent("user")
	}
}

func BenchmarkParser_ExtractTextContent(b *testing.B) {
	content := "<name>John</name><age>30</age><email>john@example.com</email>"
	for i := 0; i < b.N; i++ {
		_ = extractTextContent(content)
	}
}

func BenchmarkParser_EscapeXML(b *testing.B) {
	text := `This is a "test" with <special> characters & symbols`
	for i := 0; i < b.N; i++ {
		_ = escapeXML(text)
	}
}

func BenchmarkParser_UnescapeXML(b *testing.B) {
	text := `This is a &quot;test&quot; with &lt;special&gt; characters &amp; symbols`
	for i := 0; i < b.N; i++ {
		_ = unescapeXML(text)
	}
}

// ============================================================================
// Path Parsing Benchmarks
// ============================================================================

func BenchmarkPath_ParseSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = parsePath("root.user.name")
	}
}

func BenchmarkPath_ParseComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = parsePath("root.users.user[age>25].address.@id")
	}
}

func BenchmarkPath_ParseWildcard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = parsePath("root.**.price")
	}
}

func BenchmarkPath_ParseFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = parsePath("root.items.item[price>19.99]")
	}
}

// ============================================================================
// Filter Benchmarks
// ============================================================================

func BenchmarkFilter_ParseSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parseFilter("[age>25]")
	}
}

func BenchmarkFilter_ParseAttribute(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parseFilter("[@id==123]")
	}
}

func BenchmarkFilter_EvaluateNumeric(b *testing.B) {
	filter := &Filter{Path: "age", Op: OpGreaterThan, Value: "25"}
	content := "<age>30</age>"
	attrs := make(map[string]string)
	for i := 0; i < b.N; i++ {
		_ = evaluateFilterWithDepth(filter, content, attrs, 0)
	}
}

func BenchmarkFilter_EvaluateString(b *testing.B) {
	filter := &Filter{Path: "name", Op: OpEqual, Value: "John"}
	content := "<name>John</name>"
	attrs := make(map[string]string)
	for i := 0; i < b.N; i++ {
		_ = evaluateFilterWithDepth(filter, content, attrs, 0)
	}
}

func BenchmarkFilter_EvaluateAttribute(b *testing.B) {
	filter := &Filter{Path: "@id", Op: OpEqual, Value: "123"}
	content := ""
	attrs := map[string]string{"id": "123"}
	for i := 0; i < b.N; i++ {
		_ = evaluateFilterWithDepth(filter, content, attrs, 0)
	}
}

// ============================================================================
// Real-World Scenario Benchmarks
// ============================================================================

func BenchmarkScenario_ParseUserData(b *testing.B) {
	// Simulate parsing multiple fields from a user record
	paths := []string{
		"root.users.user.0.name",
		"root.users.user.0.age",
		"root.users.user.0.email",
		"root.users.user.0.address.city",
	}
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = Get(mediumXML, path)
		}
	}
}

func BenchmarkScenario_SearchProducts(b *testing.B) {
	// Simulate searching for products matching criteria
	for i := 0; i < b.N; i++ {
		_ = Get(mediumXML, "root.products.product[price<25].name")
	}
}

func BenchmarkScenario_UpdateMultipleFields(b *testing.B) {
	// Simulate updating multiple fields
	for i := 0; i < b.N; i++ {
		xml := mediumXML
		xml, _ = Set(xml, "root.users.user.0.age", "31")
		xml, _ = Set(xml, "root.users.user.0.email", "newemail@example.com")
		_ = xml
	}
}

func BenchmarkScenario_CollectAllPrices(b *testing.B) {
	// Simulate collecting all prices from a document
	for i := 0; i < b.N; i++ {
		_ = Get(largeXML, "root.**.price")
	}
}

// ============================================================================
// Result Type Conversion Benchmarks
// ============================================================================

func BenchmarkResult_String(b *testing.B) {
	result := Get(smallXML, "root.user.name")
	for i := 0; i < b.N; i++ {
		_ = result.String()
	}
}

func BenchmarkResult_Int(b *testing.B) {
	result := Get(smallXML, "root.user.age")
	for i := 0; i < b.N; i++ {
		_ = result.Int()
	}
}

func BenchmarkResult_Float(b *testing.B) {
	result := Get(mediumXML, "root.products.product.price")
	for i := 0; i < b.N; i++ {
		_ = result.Float()
	}
}

func BenchmarkResult_Bool(b *testing.B) {
	result := Get(smallXML, "root.user.age")
	for i := 0; i < b.N; i++ {
		_ = result.Bool()
	}
}

// Phase 6: Namespace benchmarks

var namespacedXML = `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	<soap:Header>
		<auth:Token xmlns:auth="http://example.com/auth">secret123</auth:Token>
	</soap:Header>
	<soap:Body>
		<m:GetUser xmlns:m="http://example.com/methods">
			<m:UserId>123</m:UserId>
		</m:GetUser>
	</soap:Body>
</soap:Envelope>`

// BenchmarkNamespacePrefixMatch measures overhead of exact namespace prefix matching
func BenchmarkNamespacePrefixMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := Get(namespacedXML, "soap:Envelope.soap:Body.m:GetUser.m:UserId")
		_ = result.String()
	}
}

// BenchmarkNamespaceUnprefixedMatch measures backward compatible matching (unprefixed path)
func BenchmarkNamespaceUnprefixedMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := Get(namespacedXML, "Envelope.Body.GetUser.UserId")
		_ = result.String()
	}
}

// BenchmarkNamespaceSplitting measures performance of splitNamespace function
func BenchmarkNamespaceSplitting(b *testing.B) {
	testNames := []string{
		"soap:Envelope",
		"Element",
		"ns1:ns2:complex",
		"verylongnamespaceprefix:element",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range testNames {
			prefix, local := splitNamespace(name)
			_ = prefix
			_ = local
		}
	}
}

// ============================================================================
// Batch Operations Scaling Benchmarks (Phase 4 Production Hardening)
// ============================================================================

// BenchmarkSetMany_Scale10 measures SetMany performance with 10 operations
func BenchmarkSetMany_Scale10(b *testing.B) {
	xml := `<root></root>`
	paths := make([]string, 10)
	values := make([]interface{}, 10)
	for i := 0; i < 10; i++ {
		paths[i] = fmt.Sprintf("root.item%d", i)
		values[i] = fmt.Sprintf("value%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetMany(xml, paths, values)
	}
}

// BenchmarkSetMany_Scale100 measures SetMany performance with 100 operations
func BenchmarkSetMany_Scale100(b *testing.B) {
	xml := `<root></root>`
	paths := make([]string, 100)
	values := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		paths[i] = fmt.Sprintf("root.item%d", i)
		values[i] = fmt.Sprintf("value%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetMany(xml, paths, values)
	}
}

// BenchmarkSetMany_Scale1000 measures SetMany performance with 1000 operations
func BenchmarkSetMany_Scale1000(b *testing.B) {
	xml := `<root></root>`
	paths := make([]string, 1000)
	values := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		paths[i] = fmt.Sprintf("root.item%d", i)
		values[i] = fmt.Sprintf("value%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetMany(xml, paths, values)
	}
}

// BenchmarkSetMany_Scale10K measures SetMany performance with 10,000 operations
func BenchmarkSetMany_Scale10K(b *testing.B) {
	xml := `<root></root>`
	paths := make([]string, 10000)
	values := make([]interface{}, 10000)
	for i := 0; i < 10000; i++ {
		paths[i] = fmt.Sprintf("root.item%d", i)
		values[i] = fmt.Sprintf("value%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetMany(xml, paths, values)
	}
}

// BenchmarkDeleteMany_Scale10 measures DeleteMany performance with 10 operations
func BenchmarkDeleteMany_Scale10(b *testing.B) {
	// Pre-populate XML with items to delete
	xml := `<root>`
	for i := 0; i < 10; i++ {
		xml += fmt.Sprintf("<item%d>value%d</item%d>", i, i, i)
	}
	xml += `</root>`

	paths := make([]string, 10)
	for i := 0; i < 10; i++ {
		paths[i] = fmt.Sprintf("root.item%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeleteMany(xml, paths...)
	}
}

// BenchmarkDeleteMany_Scale100 measures DeleteMany performance with 100 operations
func BenchmarkDeleteMany_Scale100(b *testing.B) {
	// Pre-populate XML with items to delete
	xml := `<root>`
	for i := 0; i < 100; i++ {
		xml += fmt.Sprintf("<item%d>value%d</item%d>", i, i, i)
	}
	xml += `</root>`

	paths := make([]string, 100)
	for i := 0; i < 100; i++ {
		paths[i] = fmt.Sprintf("root.item%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeleteMany(xml, paths...)
	}
}

// BenchmarkDeleteMany_Scale1000 measures DeleteMany performance with 1000 operations
func BenchmarkDeleteMany_Scale1000(b *testing.B) {
	// Pre-populate XML with items to delete
	xml := `<root>`
	for i := 0; i < 1000; i++ {
		xml += fmt.Sprintf("<item%d>value%d</item%d>", i, i, i)
	}
	xml += `</root>`

	paths := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		paths[i] = fmt.Sprintf("root.item%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeleteMany(xml, paths...)
	}
}

// BenchmarkDeleteMany_Scale10K measures DeleteMany performance with 10,000 operations
func BenchmarkDeleteMany_Scale10K(b *testing.B) {
	// Pre-populate XML with items to delete
	xml := `<root>`
	for i := 0; i < 10000; i++ {
		xml += fmt.Sprintf("<item%d>value%d</item%d>", i, i, i)
	}
	xml += `</root>`

	paths := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		paths[i] = fmt.Sprintf("root.item%d", i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeleteMany(xml, paths...)
	}
}

// ============================================================================
// Fluent API Benchmarks (Result.Get, Result.GetMany, Result.GetWithOptions)
// ============================================================================

// BenchmarkResultGet_Simple benchmarks simple fluent Get chaining
func BenchmarkResultGet_Simple(b *testing.B) {
	xml := `<root>
		<user>
			<name>Alice</name>
			<age>30</age>
		</user>
	</root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root := Get(xml, "root")
		user := root.Get("user")
		_ = user.Get("name")
	}
}

// BenchmarkResultGet_Deep benchmarks deep chaining (5 levels)
func BenchmarkResultGet_Deep(b *testing.B) {
	xml := `<root>
		<level1>
			<level2>
				<level3>
					<level4>
						<level5>
							<value>test</value>
						</level5>
					</level4>
				</level3>
			</level2>
		</level1>
	</root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Get(xml, "root").
			Get("level1").
			Get("level2").
			Get("level3").
			Get("level4").
			Get("level5").
			Get("value")
		_ = result.String()
	}
}

// BenchmarkResultGet_FieldExtraction benchmarks fluent field extraction
func BenchmarkResultGet_FieldExtraction(b *testing.B) {
	xml := `<root>
		<items>
			<item><name>A</name></item>
			<item><name>B</name></item>
			<item><name>C</name></item>
		</items>
	</root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		items := Get(xml, "root.items")
		names := items.Get("item.#.name")
		_ = names.IsArray()
	}
}

// BenchmarkResultGetMany benchmarks GetMany batch queries
func BenchmarkResultGetMany(b *testing.B) {
	xml := `<root>
		<user>
			<name>Alice</name>
			<age>30</age>
			<city>NYC</city>
			<country>USA</country>
			<email>alice@example.com</email>
		</user>
	</root>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := Get(xml, "root.user")
		results := user.GetMany("name", "age", "city", "country", "email")
		_ = results[0].String()
	}
}

// BenchmarkResultGetWithOptions benchmarks GetWithOptions with case-insensitive
func BenchmarkResultGetWithOptions(b *testing.B) {
	xml := `<root>
		<USER>
			<NAME>Alice</NAME>
			<AGE>30</AGE>
		</USER>
	</root>`

	opts := &Options{CaseSensitive: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root := Get(xml, "root")
		user := root.GetWithOptions("user", opts)
		_ = user.GetWithOptions("name", opts)
	}
}

// BenchmarkResultGet_vs_DirectGet compares fluent vs direct Get performance
func BenchmarkResultGet_vs_DirectGet(b *testing.B) {
	xml := `<root>
		<user>
			<profile>
				<name>Alice</name>
			</profile>
		</user>
	</root>`

	b.Run("Fluent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			root := Get(xml, "root")
			user := root.Get("user")
			profile := user.Get("profile")
			_ = profile.Get("name")
		}
	})

	b.Run("Direct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Get(xml, "root.user.profile.name")
		}
	})
}

// BenchmarkResultGet_LargeXML benchmarks fluent API on large XML
func BenchmarkResultGet_LargeXML(b *testing.B) {
	// largeXML has 100 items
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root := Get(largeXML, "root")
		items := root.Get("items")
		item := items.Get("item.0")
		_ = item.Get("name")
	}
}

package main

import (
	"fmt"

	xmldot "github.com/netascode/xmldot"
)

// Sample XML: Simple product catalog
const catalogXML = `<catalog>
	<product id="101">
		<name>Laptop</name>
		<price>999.99</price>
		<specs>
			<cpu>Intel i7</cpu>
			<ram>16GB</ram>
		</specs>
	</product>
	<product id="102">
		<name>Mouse</name>
		<price>29.99</price>
	</product>
	<product id="103">
		<name>Keyboard</name>
		<price>79.99</price>
	</product>
</catalog>`

func main() {
	fmt.Println("Basic Get Operations Example")
	fmt.Println("=============================\n")

	// Example 1: Simple element access
	fmt.Println("Example 1: Simple element access")
	result := xmldot.Get(catalogXML, "catalog.product.name")
	fmt.Printf("First product name: %s\n\n", result.String())

	// Example 2: Attribute access
	fmt.Println("Example 2: Attribute access")
	result = xmldot.Get(catalogXML, "catalog.product.@id")
	fmt.Printf("First product ID: %s\n\n", result.String())

	// Example 3: Nested element access
	fmt.Println("Example 3: Nested element access")
	result = xmldot.Get(catalogXML, "catalog.product.specs.cpu")
	fmt.Printf("CPU: %s\n\n", result.String())

	// Example 4: Array indexing
	fmt.Println("Example 4: Array indexing")
	result = xmldot.Get(catalogXML, "catalog.product.1.name")
	fmt.Printf("Second product: %s\n\n", result.String())

	// Example 5: Array count
	fmt.Println("Example 5: Array count")
	result = xmldot.Get(catalogXML, "catalog.product.#")
	fmt.Printf("Total products: %d\n\n", result.Int())

	// Example 6: Type conversion
	fmt.Println("Example 6: Type conversion")
	result = xmldot.Get(catalogXML, "catalog.product.price")
	fmt.Printf("Price as float: %.2f\n", result.Float())
	fmt.Printf("Price as string: %s\n\n", result.String())

	// Example 7: GetMany for multiple paths
	fmt.Println("Example 7: GetMany for multiple paths")
	results := xmldot.GetMany(catalogXML,
		"catalog.product.0.name",
		"catalog.product.0.@id",
		"catalog.product.0.price",
	)
	fmt.Printf("Product: %s (ID: %s) - $%s\n\n", results[0].String(), results[1].String(), results[2].String())

	// Example 8: Non-existent paths
	fmt.Println("Example 8: Non-existent paths")
	result = xmldot.Get(catalogXML, "catalog.product.nonexistent")
	fmt.Printf("Exists: %v\n", result.Exists())
	fmt.Printf("Value: %s\n\n", result.String())

	// Example 9: Checking existence
	fmt.Println("Example 9: Checking existence")
	result = xmldot.Get(catalogXML, "catalog.product.specs")
	if result.Exists() {
		fmt.Println("Specs element exists")
	}
	result = xmldot.Get(catalogXML, "catalog.product.1.specs")
	if !result.Exists() {
		fmt.Println("Second product has no specs")
	}
}

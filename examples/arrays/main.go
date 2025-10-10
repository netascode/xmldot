package main

import (
	"fmt"
	"log"

	xmldot "github.com/netascode/xmldot"
)

// Sample XML: Shopping cart
const cartXML = `<cart>
	<items>
		<item><name>Book</name><quantity>2</quantity></item>
		<item><name>Pen</name><quantity>5</quantity></item>
		<item><name>Notebook</name><quantity>1</quantity></item>
	</items>
</cart>`

func main() {
	fmt.Println("Array Operations Example")
	fmt.Println("========================\n")

	// Example 1: Count array elements
	fmt.Println("Example 1: Count array elements")
	count := xmldot.Get(cartXML, "cart.items.item.#")
	fmt.Printf("Total items: %d\n\n", count.Int())

	// Example 2: Access by index
	fmt.Println("Example 2: Access by index")
	first := xmldot.Get(cartXML, "cart.items.item.0.name")
	second := xmldot.Get(cartXML, "cart.items.item.1.name")
	last := xmldot.Get(cartXML, "cart.items.item.2.name")
	fmt.Printf("First: %s\n", first.String())
	fmt.Printf("Second: %s\n", second.String())
	fmt.Printf("Last: %s\n\n", last.String())

	// Example 3: Negative indices
	fmt.Println("Example 3: Negative indices (last element)")
	lastItem := xmldot.Get(cartXML, "cart.items.item.-1.name")
	fmt.Printf("Last item: %s\n\n", lastItem.String())

	// Example 4: Replace array element
	fmt.Println("Example 4: Replace array element")
	xml, err := xmldot.Set(cartXML, "cart.items.item.1.name", "Pencil")
	if err != nil {
		log.Fatal(err)
	}
	updated := xmldot.Get(xml, "cart.items.item.1.name")
	fmt.Printf("Updated second item: %s\n\n", updated.String())

	// Example 5: Count after modification
	fmt.Println("Example 5: Count after modification")
	currentCount := xmldot.Get(xml, "cart.items.item.#")
	fmt.Printf("Current item count: %d\n\n", currentCount.Int())

	// Example 6: Delete array element (use fresh XML)
	fmt.Println("Example 6: Delete array element")
	xml2, err := xmldot.Delete(cartXML, "cart.items.item.0")
	if err != nil {
		log.Fatal(err)
	}
	remaining := xmldot.Get(xml2, "cart.items.item.#")
	fmt.Printf("Items after delete: %d\n", remaining.Int())
	firstItem := xmldot.Get(xml2, "cart.items.item.0.name")
	fmt.Printf("New first item: %s\n\n", firstItem.String())

	// Example 7: Iterate with ForEach
	fmt.Println("Example 7: Iterate with ForEach")
	result := xmldot.Get(cartXML, "cart.items.item")
	fmt.Println("All items:")
	result.ForEach(func(index int, value xmldot.Result) bool {
		name := xmldot.Get(value.Raw, "name")
		qty := xmldot.Get(value.Raw, "quantity")
		fmt.Printf("  %d. %s (qty: %s)\n", index+1, name.String(), qty.String())
		return true // Continue iteration
	})
	fmt.Println()

	// Example 8: Array() method for all elements
	fmt.Println("Example 8: Array() method for all elements")
	items := xmldot.Get(cartXML, "cart.items.item.name")
	fmt.Println("Item names:")
	for i, item := range items.Array() {
		fmt.Printf("  %d. %s\n", i+1, item.String())
	}
	fmt.Println()

	// Example 9: SetMany for multiple array updates
	fmt.Println("Example 9: SetMany for multiple array updates")
	xml, err = xmldot.SetMany(cartXML,
		[]string{
			"cart.items.item.0.quantity",
			"cart.items.item.1.quantity",
			"cart.items.item.2.quantity",
		},
		[]interface{}{10, 20, 30},
	)
	if err != nil {
		log.Fatal(err)
	}
	quantities := xmldot.Get(xml, "cart.items.item.quantity")
	fmt.Print("Updated quantities: ")
	for i, q := range quantities.Array() {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(q.String())
	}
	fmt.Println("\n")

	// Example 10: DeleteMany for batch deletion
	fmt.Println("Example 10: DeleteMany for batch deletion")
	xml, err = xmldot.DeleteMany(cartXML,
		"cart.items.item.0",
		"cart.items.item.1",
	)
	if err != nil {
		log.Fatal(err)
	}
	finalCount := xmldot.Get(xml, "cart.items.item.#")
	fmt.Printf("Items remaining after batch delete: %d\n", finalCount.Int())
}

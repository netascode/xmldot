package main

import (
	"fmt"

	xmldot "github.com/netascode/xmldot"
)

// Sample XML: Blog posts
const blogXML = `<blog>
	<posts>
		<post id="3"><title>Getting Started</title><views>150</views></post>
		<post id="1"><title>Advanced Topics</title><views>420</views></post>
		<post id="2"><title>Best Practices</title><views>300</views></post>
	</posts>
</blog>`

func main() {
	fmt.Println("Built-in Modifiers Example")
	fmt.Println("==========================\n")

	// Example 1: @reverse - Reverse array
	fmt.Println("Example 1: @reverse - Reverse array order")
	result := xmldot.Get(blogXML, "blog.posts.post.title|@reverse")
	for _, title := range result.Array() {
		fmt.Printf("  - %s\n", title.String())
	}
	fmt.Println()

	// Example 2: @sort - Sort array
	fmt.Println("Example 2: @sort - Sort titles alphabetically")
	result = xmldot.Get(blogXML, "blog.posts.post.title|@sort")
	for _, title := range result.Array() {
		fmt.Printf("  - %s\n", title.String())
	}
	fmt.Println()

	// Example 3: @first - First element
	fmt.Println("Example 3: @first - Get first post title")
	result = xmldot.Get(blogXML, "blog.posts.post.title|@first")
	fmt.Printf("First title: %s\n\n", result.String())

	// Example 4: @last - Last element
	fmt.Println("Example 4: @last - Get last post title")
	result = xmldot.Get(blogXML, "blog.posts.post.title|@last")
	fmt.Printf("Last title: %s\n\n", result.String())

	// Example 5: @flatten - Flatten nested arrays
	fmt.Println("Example 5: @flatten - Flatten nested structure")
	nestedXML := `<data>
		<group><item>A</item><item>B</item></group>
		<group><item>C</item><item>D</item></group>
	</data>`
	result = xmldot.Get(nestedXML, "data.group.item|@flatten")
	fmt.Printf("Flattened: %v\n\n", result.Array())

	// Example 6: @pretty - Format XML with indentation
	fmt.Println("Example 6: @pretty - Format XML")
	compactXML := `<root><item>value</item><nested><child>data</child></nested></root>`
	result = xmldot.Get(compactXML, "root|@pretty")
	fmt.Println(result.String())
	fmt.Println()

	// Example 7: @ugly - Compact XML (remove formatting)
	fmt.Println("Example 7: @ugly - Compact XML")
	prettyXML := `<root>
	<item>value</item>
	<nested>
		<child>data</child>
	</nested>
</root>`
	result = xmldot.Get(prettyXML, "root|@ugly")
	fmt.Printf("Compact: %s\n\n", result.String())

	// Example 8: @raw - Raw XML output
	fmt.Println("Example 8: @raw - Get raw XML")
	result = xmldot.Get(blogXML, "blog.posts.post.0|@raw")
	fmt.Println(result.String())
	fmt.Println()

	// Example 9: @keys - Extract element names
	fmt.Println("Example 9: @keys - Extract keys")
	result = xmldot.Get(blogXML, "blog.posts.post.0|@keys")
	fmt.Print("Keys: ")
	for i, key := range result.Array() {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(key.String())
	}
	fmt.Println("\n")

	// Example 10: @values - Extract values only
	fmt.Println("Example 10: @values - Extract values")
	result = xmldot.Get(blogXML, "blog.posts.post.0|@values")
	fmt.Print("Values: ")
	for i, val := range result.Array() {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(val.String())
	}
	fmt.Println("\n")

	// Example 11: Chaining modifiers
	fmt.Println("Example 11: Chaining - @sort|@reverse|@first")
	result = xmldot.Get(blogXML, "blog.posts.post.title|@sort|@reverse|@first")
	fmt.Printf("Last title alphabetically: %s\n\n", result.String())

	// Example 12: Modifiers with filters
	fmt.Println("Example 12: Combining filters and modifiers")
	result = xmldot.Get(blogXML, "blog.posts.post.#(views>200)#.title|@sort")
	fmt.Println("Popular posts (sorted):")
	for _, title := range result.Array() {
		fmt.Printf("  - %s\n", title.String())
	}
}

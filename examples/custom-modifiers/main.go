package main

import (
	"fmt"
	"strings"

	xmldot "github.com/netascode/xmldot"
)

// uppercaseModifier converts Result string to uppercase
type uppercaseModifier struct{}

func (m *uppercaseModifier) Name() string {
	return "uppercase"
}

func (m *uppercaseModifier) Apply(r xmldot.Result) xmldot.Result {
	// Handle different Result types
	if r.Type == xmldot.Null {
		return r
	}

	// Convert the string representation to uppercase
	return xmldot.Result{
		Type:    r.Type,
		Str:     strings.ToUpper(r.Str),
		Raw:     r.Raw,
		Num:     r.Num,
		Index:   r.Index,
		Results: r.Results,
	}
}

// lowercaseModifier converts Result string to lowercase
type lowercaseModifier struct{}

func (m *lowercaseModifier) Name() string {
	return "lowercase"
}

func (m *lowercaseModifier) Apply(r xmldot.Result) xmldot.Result {
	if r.Type == xmldot.Null {
		return r
	}

	// Handle arrays: apply to all elements
	if r.Type == xmldot.Array {
		results := make([]xmldot.Result, len(r.Results))
		for i, elem := range r.Results {
			results[i] = xmldot.Result{
				Type:  elem.Type,
				Str:   strings.ToLower(elem.Str),
				Raw:   elem.Raw,
				Num:   elem.Num,
				Index: elem.Index,
			}
		}
		return xmldot.Result{
			Type:    xmldot.Array,
			Results: results,
		}
	}

	// Convert single element to lowercase
	return xmldot.Result{
		Type:  r.Type,
		Str:   strings.ToLower(r.Str),
		Raw:   r.Raw,
		Num:   r.Num,
		Index: r.Index,
	}
}

// countModifier counts array elements and returns a Number Result
type countModifier struct{}

func (m *countModifier) Name() string {
	return "count"
}

func (m *countModifier) Apply(r xmldot.Result) xmldot.Result {
	// Return 0 for Null
	if r.Type == xmldot.Null {
		return xmldot.Result{
			Type: xmldot.Number,
			Num:  0,
			Str:  "0",
		}
	}

	// Count array elements
	if r.Type == xmldot.Array {
		count := len(r.Results)
		return xmldot.Result{
			Type: xmldot.Number,
			Num:  float64(count),
			Str:  fmt.Sprintf("%d", count),
		}
	}

	// Single elements count as 1
	return xmldot.Result{
		Type: xmldot.Number,
		Num:  1,
		Str:  "1",
	}
}

// joinModifier joins array elements with a separator (default: ", ")
type joinModifier struct{}

func (m *joinModifier) Name() string {
	return "join"
}

func (m *joinModifier) Apply(r xmldot.Result) xmldot.Result {
	// Return empty for Null
	if r.Type == xmldot.Null {
		return r
	}

	// For non-arrays, return as-is
	if r.Type != xmldot.Array {
		return r
	}

	// Join array elements with ", "
	var parts []string
	for _, elem := range r.Results {
		parts = append(parts, elem.String())
	}

	joined := strings.Join(parts, ", ")
	return xmldot.Result{
		Type: xmldot.String,
		Str:  joined,
	}
}

func init() {
	// Register custom modifiers
	if err := xmldot.RegisterModifier("uppercase", &uppercaseModifier{}); err != nil {
		panic(fmt.Sprintf("Failed to register uppercase modifier: %v", err))
	}
	if err := xmldot.RegisterModifier("lowercase", &lowercaseModifier{}); err != nil {
		panic(fmt.Sprintf("Failed to register lowercase modifier: %v", err))
	}
	if err := xmldot.RegisterModifier("count", &countModifier{}); err != nil {
		panic(fmt.Sprintf("Failed to register count modifier: %v", err))
	}
	if err := xmldot.RegisterModifier("join", &joinModifier{}); err != nil {
		panic(fmt.Sprintf("Failed to register join modifier: %v", err))
	}
}

func main() {
	xml := `<catalog>
		<books>
			<book><title>The Go Programming Language</title></book>
			<book><title>Learning Go</title></book>
			<book><title>Concurrency in Go</title></book>
		</books>
	</catalog>`

	// Example 1: Uppercase first title
	fmt.Println("Example 1: Uppercase first title")
	result := xmldot.Get(xml, "catalog.books.*.title|@first|@uppercase")
	fmt.Printf("Result: %s\n\n", result.String())

	// Example 2: Lowercase all titles
	fmt.Println("Example 2: Lowercase all titles")
	result = xmldot.Get(xml, "catalog.books.*.title|@lowercase")
	for _, title := range result.Array() {
		fmt.Printf("  - %s\n", title.String())
	}
	fmt.Println()

	// Example 3: Count books
	fmt.Println("Example 3: Count books")
	result = xmldot.Get(xml, "catalog.books.*|@count")
	fmt.Printf("Total books: %d\n\n", result.Int())

	// Example 4: Join titles
	fmt.Println("Example 4: Join titles")
	result = xmldot.Get(xml, "catalog.books.*.title|@join")
	fmt.Printf("All titles: %s\n\n", result.String())

	// Example 5: Chain custom and built-in modifiers
	fmt.Println("Example 5: Chain modifiers")
	result = xmldot.Get(xml, "catalog.books.*.title|@sort|@reverse|@first|@uppercase")
	fmt.Printf("Last book (sorted): %s\n", result.String())
}

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	xmldot "github.com/netascode/xmldot"
)

// generateLargeXML creates a sample XML document with N items
func generateLargeXML(items int) string {
	var sb strings.Builder
	sb.WriteString("<catalog>\n")
	for i := 0; i < items; i++ {
		sb.WriteString(fmt.Sprintf(`	<product id="%d">
		<name>Product %d</name>
		<price>%.2f</price>
		<category>Category %d</category>
	</product>
`, i, i, float64(i)*1.5+10.0, i%5))
	}
	sb.WriteString("</catalog>")
	return sb.String()
}

func main() {
	fmt.Println("Performance Optimization Example")
	fmt.Println("=================================\n")

	// Generate test data
	smallXML := generateLargeXML(10)
	largeXML := generateLargeXML(1000)

	// Example 1: Baseline - Normal Get operations
	fmt.Println("Example 1: Baseline - Individual Get operations")
	start := time.Now()
	for i := 0; i < 100; i++ {
		_ = xmldot.Get(smallXML, "catalog.product.0.name")
		_ = xmldot.Get(smallXML, "catalog.product.0.price")
		_ = xmldot.Get(smallXML, "catalog.product.0.@id")
	}
	baseline := time.Since(start)
	fmt.Printf("100 iterations × 3 queries: %v\n\n", baseline)

	// Example 2: GetMany for batch queries
	fmt.Println("Example 2: GetMany for batch queries")
	start = time.Now()
	for i := 0; i < 100; i++ {
		_ = xmldot.GetMany(smallXML,
			"catalog.product.0.name",
			"catalog.product.0.price",
			"catalog.product.0.@id",
		)
	}
	batchTime := time.Since(start)
	speedup := float64(baseline) / float64(batchTime)
	fmt.Printf("100 iterations × 1 GetMany: %v\n", batchTime)
	fmt.Printf("Speedup: %.2fx faster\n\n", speedup)

	// Example 3: SetBytes for Set operations
	fmt.Println("Example 3: SetBytes for Set operations")

	// Normal mode (with validation)
	start = time.Now()
	xml := smallXML
	for i := 0; i < 100; i++ {
		var err error
		xml, err = xmldot.Set(xml, "catalog.product.0.price", fmt.Sprintf("%.2f", float64(i)))
		if err != nil {
			log.Fatal(err)
		}
	}
	normalTime := time.Since(start)
	fmt.Printf("Normal mode (100 Sets): %v\n", normalTime)

	// Using SetBytes for better performance
	start = time.Now()
	xmlBytes := []byte(smallXML)
	for i := 0; i < 100; i++ {
		var err error
		xmlBytes, err = xmldot.SetBytes(xmlBytes, "catalog.product.0.price", fmt.Sprintf("%.2f", float64(i)))
		if err != nil {
			log.Fatal(err)
		}
	}
	optimisticTime := time.Since(start)
	speedup = float64(normalTime) / float64(optimisticTime)
	fmt.Printf("SetBytes mode (100 Sets): %v\n", optimisticTime)
	fmt.Printf("Speedup: %.2fx faster\n\n", speedup)

	// Example 4: GetBytes for reduced allocations
	fmt.Println("Example 4: GetBytes for reduced allocations")
	xmlBytesStatic := []byte(smallXML)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		_ = xmldot.Get(smallXML, "catalog.product.0.name")
	}
	stringTime := time.Since(start)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		_ = xmldot.GetBytes(xmlBytesStatic, "catalog.product.0.name")
	}
	bytesTime := time.Since(start)
	fmt.Printf("String-based (1000 Gets): %v\n", stringTime)
	fmt.Printf("Bytes-based (1000 Gets): %v\n", bytesTime)
	fmt.Printf("Allocation savings: ~%.1f%%\n\n", (1.0-float64(bytesTime)/float64(stringTime))*100)

	// Example 5: Path caching benefits
	fmt.Println("Example 5: Path caching benefits (automatic)")
	const iterations = 10000

	// First run (cold cache)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = xmldot.Get(smallXML, "catalog.product.0.name")
	}
	coldTime := time.Since(start)

	// Second run (warm cache)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = xmldot.Get(smallXML, "catalog.product.0.name")
	}
	warmTime := time.Since(start)

	fmt.Printf("First %d queries: %v\n", iterations, coldTime)
	fmt.Printf("Repeat %d queries: %v\n", iterations, warmTime)
	fmt.Printf("Cache benefit: %.1f%% faster\n\n", (1.0-float64(warmTime)/float64(coldTime))*100)

	// Example 6: Avoid recursive wildcards on large documents
	fmt.Println("Example 6: Wildcard performance comparison")

	// Single-level wildcard
	start = time.Now()
	_ = xmldot.Get(largeXML, "catalog.*.name")
	singleLevel := time.Since(start)

	// Recursive wildcard (expensive)
	start = time.Now()
	_ = xmldot.Get(largeXML, "catalog.**.name")
	recursive := time.Since(start)

	fmt.Printf("Single-level wildcard: %v\n", singleLevel)
	fmt.Printf("Recursive wildcard: %v\n", recursive)
	fmt.Printf("Recursive overhead: %.2fx slower\n\n", float64(recursive)/float64(singleLevel))

	// Example 7: Filter optimization (numeric vs string)
	fmt.Println("Example 7: Filter optimization")

	// Numeric filter (fast path)
	start = time.Now()
	for i := 0; i < 100; i++ {
		_ = xmldot.Get(largeXML, "catalog.product.#(price>50)#.name")
	}
	numericFilter := time.Since(start)

	// String filter
	start = time.Now()
	for i := 0; i < 100; i++ {
		_ = xmldot.Get(largeXML, "catalog.product.#(category==Category 1)#.name")
	}
	stringFilter := time.Since(start)

	fmt.Printf("Numeric filter (100 queries): %v\n", numericFilter)
	fmt.Printf("String filter (100 queries): %v\n", stringFilter)
	fmt.Println()

	// Example 8: SetMany vs sequential Set
	fmt.Println("Example 8: SetMany vs sequential Set operations")

	// Sequential Set
	start = time.Now()
	xml = smallXML
	for i := 0; i < 5; i++ {
		var err error
		xml, err = xmldot.Set(xml, fmt.Sprintf("catalog.product.%d.price", i), "99.99")
		if err != nil {
			log.Fatal(err)
		}
	}
	sequentialTime := time.Since(start)

	// Batch SetMany
	start = time.Now()
	paths := make([]string, 5)
	values := make([]interface{}, 5)
	for i := 0; i < 5; i++ {
		paths[i] = fmt.Sprintf("catalog.product.%d.price", i)
		values[i] = "99.99"
	}
	xml, err := xmldot.SetMany(smallXML, paths, values)
	if err != nil {
		log.Fatal(err)
	}
	batchSetTime := time.Since(start)

	fmt.Printf("Sequential Set (5 operations): %v\n", sequentialTime)
	fmt.Printf("Batch SetMany (5 operations): %v\n", batchSetTime)
	fmt.Printf("Speedup: %.2fx faster\n\n", float64(sequentialTime)/float64(batchSetTime))

	// Example 9: Performance checklist
	fmt.Println("Example 9: Performance Optimization Checklist")
	fmt.Println("✓ Use GetMany/SetMany for multiple paths")
	fmt.Println("✓ Use GetBytes/SetBytes to reduce allocations")
	fmt.Println("✓ Path caching is automatic (85-91% faster on repeat queries)")
	fmt.Println("✓ Prefer single-level wildcards (*) over recursive (**)")
	fmt.Println("✓ Use numeric filters when possible (fast path optimization)")
	fmt.Println("✓ Profile your specific workload before optimizing")
}

// nolint // Example programs don't need package comments
package main

import (
	"fmt"
	"sync"

	xmldot "github.com/netascode/xmldot"
)

func main() {
	fmt.Println("Concurrency Examples")
	fmt.Println("=============================")
	fmt.Println()

	// Example 1: Concurrent reads (safe)
	example1ConcurrentReads()

	// Example 2: Synchronized writes
	example2SynchronizedWrites()

	// Example 3: Per-goroutine copies
	example3PerGoroutineCopies()

	// Example 4: RWMutex for read-heavy workloads
	example4RWMutex()
}

// Example 1: Concurrent reads are safe without synchronization
func example1ConcurrentReads() {
	fmt.Println("Example 1: Concurrent Reads (Safe)")
	fmt.Println("-----------------------------------")

	xml := `<store>
		<products>
			<product id="1"><name>Widget</name><price>19.99</price></product>
			<product id="2"><name>Gadget</name><price>29.99</price></product>
			<product id="3"><name>Doohickey</name><price>39.99</price></product>
		</products>
	</store>`

	var wg sync.WaitGroup

	// Multiple goroutines reading the same XML
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// All read operations are safe
			productID := id % 3
			name := xmldot.Get(xml, fmt.Sprintf("store.products.product.%d.name", productID))
			price := xmldot.Get(xml, fmt.Sprintf("store.products.product.%d.price", productID))

			fmt.Printf("  Goroutine %d: Product=%s, Price=$%s\n", id, name.String(), price.String())
		}(i)
	}

	wg.Wait()
	fmt.Println()
}

// Example 2: Synchronized writes using mutex
func example2SynchronizedWrites() {
	fmt.Println("Example 2: Synchronized Writes")
	fmt.Println("--------------------------------")

	var mu sync.Mutex
	currentXML := "<inventory><items></items></inventory>"

	var wg sync.WaitGroup

	// Multiple goroutines adding items with synchronization
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			mu.Lock()
			result, err := xmldot.Set(currentXML, fmt.Sprintf("inventory.items.item%d", id), fmt.Sprintf("Item %d", id))
			if err == nil {
				currentXML = result
			}
			mu.Unlock()

			fmt.Printf("  Goroutine %d: Added item\n", id)
		}(i)
	}

	wg.Wait()

	// Verify final result
	if xmldot.Valid(currentXML) {
		fmt.Println("  Final XML is valid")
		fmt.Printf("  Items: %s\n", xmldot.Get(currentXML, "inventory.items").Raw)
	}
	fmt.Println()
}

// Example 3: Per-goroutine copies (no synchronization needed)
func example3PerGoroutineCopies() {
	fmt.Println("Example 3: Per-Goroutine Copies")
	fmt.Println("--------------------------------")

	originalXML := "<config><setting>default</setting></config>"
	results := make([]string, 5)

	var wg sync.WaitGroup

	// Each goroutine works with its own copy
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine has its own copy - no synchronization needed
			xmlCopy := originalXML
			modified, _ := xmldot.Set(xmlCopy, "config.setting", fmt.Sprintf("value-%d", id))
			results[id] = modified

			fmt.Printf("  Goroutine %d: Created independent copy\n", id)
		}(i)
	}

	wg.Wait()

	// Verify all results are independent
	for i, result := range results {
		setting := xmldot.Get(result, "config.setting")
		fmt.Printf("  Result %d: setting=%s\n", i, setting.String())
	}
	fmt.Println()
}

// Example 4: RWMutex for read-heavy workloads
func example4RWMutex() {
	fmt.Println("Example 4: RWMutex for Read-Heavy Workloads")
	fmt.Println("--------------------------------------------")

	var mu sync.RWMutex
	currentXML := `<cache>
		<entry key="user1">Alice</entry>
		<entry key="user2">Bob</entry>
	</cache>`

	var wg sync.WaitGroup

	// Many concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			mu.RLock()
			xmlCopy := currentXML
			mu.RUnlock()

			// Safe to read without lock (we have our own copy)
			key := fmt.Sprintf("user%d", (id%2)+1)
			value := xmldot.Get(xmlCopy, fmt.Sprintf("cache.entry.#(@key==%s)", key))
			fmt.Printf("  Reader %d: %s=%s\n", id, key, value.String())
		}(i)
	}

	// A few concurrent writers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			mu.Lock()
			key := fmt.Sprintf("user%d", id+3)
			result, err := xmldot.Set(currentXML, fmt.Sprintf("cache.entry%d.@key", id+2), key)
			if err == nil {
				currentXML = result
				fmt.Printf("  Writer %d: Added %s\n", id, key)
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	fmt.Println()
}

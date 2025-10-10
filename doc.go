// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

// Package xmldot provides high-performance XML querying and manipulation
// using GJSON-inspired path syntax.
//
// xmldot brings the simplicity and power of GJSON's dot-notation paths to XML,
// enabling intuitive queries and modifications with built-in security protections.
//
// # Features
//
//   - Simple dot-notation paths: "root.user.name"
//   - Array operations: indexing, counting, filtering
//   - GJSON-style queries: #(condition) and #(condition)# filters
//   - Field extraction: #.field to extract fields from all array elements
//   - Wildcards: * (single-level) and ** (recursive)
//   - Modifiers: @reverse, @sort, @first, @last, @flatten, etc.
//   - Built-in security: DoS protection, resource limits, XXE prevention
//   - Zero dependencies: uses only Go standard library
//   - High performance: 400-1000ns/op for basic queries
//
// # Basic Usage
//
// Query XML documents using intuitive path syntax:
//
//	xml := `<root><user><name>John</name><age>30</age></user></root>`
//	result := xmldot.Get(xml, "root.user.name")
//	fmt.Println(result.String()) // "John"
//
// Access attributes using @ prefix:
//
//	xml := `<user id="123"><name>John</name></user>`
//	id := xmldot.Get(xml, "user.@id")
//	fmt.Println(id.String()) // "123"
//
// Work with arrays using indexing and counting:
//
//	xml := `<users><user>Alice</user><user>Bob</user></users>`
//	count := xmldot.Get(xml, "users.user.#")
//	fmt.Println(count.Int()) // 2
//	first := xmldot.Get(xml, "users.user.0")
//	fmt.Println(first.String()) // "Alice"
//
// Filter arrays using GJSON-style queries:
//
//	xml := `<users>
//	  <user><name>Alice</name><age>30</age></user>
//	  <user><name>Bob</name><age>25</age></user>
//	</users>`
//	result := xmldot.Get(xml, "users.user.#(age>28)#.name")
//	fmt.Println(result.Array()[0].String()) // "Alice"
//
// Modify XML documents:
//
//	xml := `<root><value>old</value></root>`
//	updated, _ := xmldot.Set(xml, "root.value", "new")
//	result := xmldot.Get(updated, "root.value")
//	fmt.Println(result.String()) // "new"
//
// # Security
//
// This package implements multiple security protections to prevent common
// XML vulnerabilities and denial-of-service attacks:
//
//   - Document size limits (MaxDocumentSize: 10MB)
//   - Nesting depth limits (MaxNestingDepth: 100 levels)
//   - Attribute limits (MaxAttributes: 100 per element)
//   - Token size limits (MaxTokenSize: 1MB)
//   - Wildcard result limits (MaxWildcardResults: 1000)
//   - CPU exhaustion protection (MaxRecursiveOperations: 10000)
//   - XXE attack prevention (DOCTYPE declarations skipped)
//   - No entity expansion (entities not processed)
//
// All limits are enforced automatically and can be adjusted by modifying
// the constants in parser.go if needed for trusted documents.
//
// # Path Syntax
//
// The path syntax is inspired by GJSON and supports:
//
//   - Dot notation: "root.child.element"
//   - Attributes: "element.@attribute"
//   - Array indexing: "items.item.0" (first), "items.item.-1" (last)
//   - Array counting: "items.item.#"
//   - Text content: "element.%" (direct text only)
//   - Wildcards: "root.*.name" (any child), "root.**.price" (any depth)
//   - Filters: "items.item.#(price>10)#.name"
//   - Field extraction: "items.item.#.name" (extract name from all items)
//   - Modifiers: "items.item.#.price|@sort|@reverse"
//
// For complete path syntax documentation, see docs/path-syntax.md
//
// # Performance
//
// xmldot is designed for high performance with typical operations completing
// in sub-microsecond time:
//
//   - Simple paths: 200-500ns
//   - Path with caching: ~200ns
//   - Filter queries: 1-2µs
//   - Recursive wildcards: 2-5µs
//
// The parser uses an incremental parsing approach that stops as soon as
// the requested path is found, avoiding unnecessary parsing of large documents.
//
// # Concurrency
//
// All query operations (Get, GetMany, GetWithOptions) are safe for concurrent
// use from multiple goroutines. Each call creates its own parser instance
// with no shared state.
//
// Modification operations (Set, Delete) return new XML strings and are also
// safe for concurrent use, though modifications to the same document should
// be coordinated by the caller to avoid lost updates.
//
// # For More Information
//
// Complete documentation: https://github.com/netascode/xmldot
//
// Documentation files:
//   - docs/path-syntax.md - Complete path syntax reference
//   - docs/performance.md - Performance guide and benchmarks
//   - docs/security.md - Security features and best practices
//   - docs/error-handling.md - Error handling patterns
//   - docs/migration.md - Migration from other XML libraries
//
// Examples: See the examples/ directory for runnable code samples.
package xmldot

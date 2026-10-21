# Security Documentation

Comprehensive guide to XMLDOT security features, safe usage patterns, and protection mechanisms.

## Table of Contents

1. [Introduction](#introduction)
2. [Built-in Protections](#built-in-protections)
3. [Security Limits Reference](#security-limits-reference)
4. [Safe Usage Patterns](#safe-usage-patterns)
5. [Unsafe Patterns to Avoid](#unsafe-patterns-to-avoid)
6. [Concurrency Safety](#concurrency-safety)
7. [Reporting Security Issues](#reporting-security-issues)

---

## Introduction

XMLDOT is designed with a **secure-by-default** philosophy, implementing multiple layers of protection against common XML attacks and resource exhaustion vulnerabilities.

### Security Philosophy

- **Secure by default**: All security protections enabled automatically
- **Defense in depth**: Multiple overlapping protections
- **Fail loudly**: Exceeding a limit returns an error, never a partial result
- **No external entities**: XXE protection built-in
- **Resource limits**: Protection against DoS attacks

### Threat Model

XMLDOT protects against:
- **XXE (XML External Entity)** attacks
- **Billion Laughs** (entity expansion) attacks
- **Resource exhaustion** (memory, CPU)
- **Stack overflow** (deeply nested XML)
- **Buffer overflow** (extremely large tokens)
- **Attribute flooding** attacks
- **Query complexity** attacks (recursive wildcards)
- **XML injection** attacks (Set operations)

### Target Audience

This guide is for:
- Developers processing untrusted XML input
- Security engineers reviewing XML processing code
- DevOps teams deploying XML services
- Compliance teams requiring security documentation

---

## Built-in Protections

### 1. XXE (XML External Entity) Prevention

**Threat**: External entity expansion can read arbitrary files or make network requests.

**Protection**:
```go
// xmldot does NOT process external entities or DOCTYPEs
xml := `
<!DOCTYPE foo [
  <!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<root>&xxe;</root>`

// External entity is NOT expanded - returns empty
result := xmldot.Get(xml, "root")
fmt.Println(result.String())  // Empty (entity reference not expanded)

// DOCTYPE declarations are skipped entirely
// No external entity resolution
// No file system access
// No network requests
```

**Technical Details**:
- DOCTYPE declarations are detected and skipped (case-insensitive)
- Entity references (`&entity;`) are not expanded
- No external entity resolution
- No file system or network access
- Built-in protection, cannot be disabled

**Attack Example (Prevented)**:
```go
// Attack attempt: Read /etc/passwd via XXE
maliciousXML := `
<!DOCTYPE foo [
  <!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<data><value>&xxe;</value></data>`

// xmldot protection: Entity not expanded, returns empty
value := xmldot.Get(maliciousXML, "data.value")
fmt.Println(value.Exists())  // true (element exists)
fmt.Println(value.String())  // "" (empty - entity not expanded)

// No file access occurs
// No sensitive data leaked
```

### 2. Entity Expansion Protection (Billion Laughs)

**Threat**: Recursive entity expansion can cause exponential memory consumption.

**Protection**:
```go
// Billion Laughs attack attempt
billionLaughs := `
<!DOCTYPE lolz [
  <!ENTITY lol "lol">
  <!ENTITY lol2 "&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;">
  <!ENTITY lol3 "&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;">
  <!ENTITY lol4 "&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;">
]>
<root>&lol4;</root>`

// xmldot protection: Entities not expanded
result := xmldot.Get(billionLaughs, "root")
// No exponential memory allocation
// No DoS condition
```

**Technical Details**:
- Entity references are never expanded
- No recursive entity processing
- Constant memory usage regardless of entity depth
- Protection is always active

### 3. Failing on Limit Violations

Every security limit fails the operation instead of returning a partial result.
A query that exceeds a limit never looks like "element not found":

- `Get` and related functions return a Null `Result` whose `Err` field holds a `*LimitError`
- `Set`, `Delete` and related functions return the `*LimitError` and the original document
- `Result.Map` returns `nil` (an element without children returns an empty, non-nil map)
- `ValidateWithError` returns a `*ValidateError` that wraps the `*LimitError`

```go
result := xmldot.Get(xml, "root.items.item.#")
if result.Err != nil {
    var limitErr *xmldot.LimitError
    if errors.As(result.Err, &limitErr) {
        log.Printf("limit %s exceeded: %d > %d", limitErr.Limit, limitErr.Value, limitErr.Max)
    }
    return result.Err
}
```

Use `errors.Is(err, xmldot.ErrLimitExceeded)` to detect any limit violation. A Null
result derived from a failed result (for example via `Result.Get`) carries the same `Err`.

### 4. Document and Value Size Limits

**Threat**: Extremely large documents or values can exhaust memory.

**Protection**:
```go
xmldot.MaxDocumentSize // 100 MB default
xmldot.MaxValueSize    // 100 MB default (single value in Set operations)

result := xmldot.Get(hugeXML, "root.element")
// result.Exists() == false, result.Err is a *LimitError for MaxDocumentSize

_, err := xmldot.Set(xml, "root.element", hugeValue)
// err is a *LimitError for MaxValueSize, or MaxDocumentSize if the result grows too large
```

**Technical Details**:
- xmldot parses the input in place, so the cost of a query is linear in the document size
- The limit bounds the memory used for intermediate copies, not the input itself; enforce
  a size limit while reading untrusted input (see Safe Usage Patterns)

### 5. Nesting Depth Limits

**Threat**: Deeply nested XML can exhaust the stack in recursive processing.

**Protection**:
```go
xmldot.MaxNestingDepth // 100 levels default

result := xmldot.Get(deeplyNestedXML, "root.item")
// Any query that scans the over-deep part of the document fails:
// result.Err is a *LimitError for MaxNestingDepth

err := xmldot.ValidateWithError(deeplyNestedXML)
// errors.Is(err, xmldot.ErrLimitExceeded) == true

_, err2 := xmldot.Set(deeplyNestedXML, "root.item", "v")
// err2 is a *LimitError for MaxNestingDepth
```

Get, Set and validation apply the same absolute depth, counting the root element.

### 6. Attribute and Token Limits

**Threat**: Elements with huge numbers of attributes, or huge names and attribute
values, can exhaust memory during parsing.

**Protection**:
```go
xmldot.MaxAttributes // 1000 attributes per element
xmldot.MaxTokenSize  // 1 MB per element name, attribute name or attribute value

result := xmldot.Get(attributeFlood, "item.@attr0")
// result.Err is a *LimitError for MaxAttributes
```

### 7. Control Character Rejection

**Threat**: Control characters in filter paths and values can enable log injection attacks and output corruption.

**Protection**:
```go
// Control characters automatically rejected in filter expressions
xml := `<items><item><name>value</name></item></items>`

// ❌ Rejected: newline in filter path
result := xmldot.Get(xml, "items.item.#(field\nname==value)")
// Returns Null (control character rejected)

// ❌ Rejected: tab in filter value
result := xmldot.Get(xml, "items.item.#(name==val\tue)")
// Returns Null (control character rejected)

// ✅ Accepted: valid filter
result := xmldot.Get(xml, "items.item.#(name==value)")
// Works correctly
```

**Technical Details**:
- ASCII control characters rejected: `\x00` (null), `\n` (newline), `\r` (carriage return), `\t` (tab)
- Validation occurs BEFORE string trimming (prevents hidden control characters)
- Applies to both filter paths and filter values
- Protection cannot be disabled (always active)
- Prevents log injection, output corruption, and downstream parser confusion

**Attack Example (Prevented)**:
```go
// Attack: Log injection via newline in filter
maliciousFilter := "items.item.#(status==active\nINJECTED LOG LINE)"
result := xmldot.Get(xml, maliciousFilter)
// Returns Null - attack prevented
// No log injection occurs
```

### 8. Filter Limits

**Threat**: Long filter expressions and pathological patterns can cause excessive
CPU usage.

**Protection**:
```go
xmldot.MaxFilterExpressionLength // 4096 bytes per filter expression
xmldot.MaxPatternIterations      // 10,000,000 steps per % or !% match, plus one per value byte

result := xmldot.Get(xml, `items.item.#(name%"*`+strings.Repeat("a", 50)+`b")#`)
// Pathological patterns take up to O(len(value) * len(pattern)) steps; if a single
// match exceeds its budget, result.Err is a *LimitError instead of a guessed answer
```

**Technical Details**:
- Pattern matching uses a linear-space glob matcher that backtracks only to the most
  recent `*`, so it never needs memory proportional to `len(value) * len(pattern)`
- Typical patterns match in linear time and always complete, because the budget
  includes one step per byte of the value; a trailing `*` ends matching immediately

### 9. Array and Wildcard Result Limits

**Threat**: Queries that collect many elements can consume excessive memory.

**Protection**:
```go
xmldot.MaxArrayResults    // 100,000 siblings for #, .N, #.field, filters, *, Map()
xmldot.MaxWildcardResults // 100,000 results for recursive ** queries

result := xmldot.Get(xml, "root.item.#")
// Counts every <item>; with more than MaxArrayResults items,
// result.Err is a *LimitError instead of a truncated count
```

### 10. Path Limits

**Threat**: Extremely long or complex paths can exhaust memory during parsing.

**Protection**:
```go
xmldot.MaxPathSegments          // 100 segments
xmldot.MaxFieldNameLength       // 256 bytes for #.field names
xmldot.MaxNamespacePrefixLength // 256 bytes for namespace prefixes
xmldot.MaxModifierChainDepth    // 20 chained modifiers

longPath := "root" + strings.Repeat(".child", 150)
result := xmldot.Get(xml, longPath)
// result.Err is a *LimitError for MaxPathSegments
```

### 11. Recursive Operation Limits

**Threat**: Chained recursive wildcard queries can cause CPU exhaustion.

**Protection**:
```go
xmldot.MaxRecursiveOperations // 10,000 minimum budget

result := xmldot.Get(xml, "root.**.item.**.name")
// All "**" segments of one query share a budget that is the larger of
// MaxRecursiveOperations and the document size in bytes. A single recursive
// search over any document completes; chained searches that would revisit the
// document many times exceed the budget and set result.Err to a *LimitError
```

### 12. Set Operation Limits

Set and Delete check the input document, the value and the resulting document against
`MaxDocumentSize` and `MaxValueSize`, and return a `*LimitError` together with the
unchanged input when a limit is exceeded. Validation failures caused by a limit are
reported as the `*LimitError`, not as `ErrMalformedXML`.

### 13. XML Injection Prevention (Set Operations)

**Threat**: Unescaped user input in Set operations can inject malicious XML.

**Protection**:
```go
// User input with XML special characters
userInput := `<script>alert('xss')</script>`

// xmldot protection: Automatic escaping
xml := `<data></data>`
newXML, _ := xmldot.Set(xml, "data.value", userInput)
fmt.Println(newXML)
// <data><value>&lt;script&gt;alert('xss')&lt;/script&gt;</value></data>

// Special characters escaped:
// < → &lt;
// > → &gt;
// & → &amp;
// " → &quot;
// ' → &apos;
```

**Technical Details**:
- All Set operation values are automatically escaped
- XML special characters converted to entities
- Injection attacks prevented
- Raw XML via SetRaw has basic validation

---

## Security Limits Reference

All limits are package variables. Exceeding any of them fails the operation with a
`*LimitError` (see [Failing on Limit Violations](#3-failing-on-limit-violations)).

| Variable | Default | Purpose |
|----------|---------|---------|
| **MaxDocumentSize** | 100 MB | Maximum XML document size |
| **MaxValueSize** | 100 MB | Maximum value in Set operations |
| **MaxNestingDepth** | 100 levels | Maximum element nesting |
| **MaxAttributes** | 1000 per element | Maximum attributes per element |
| **MaxTokenSize** | 1 MB | Maximum element name, attribute name or attribute value |
| **MaxNamespacePrefixLength** | 256 bytes | Maximum namespace prefix length in paths (longer prefixes in documents are treated as unprefixed) |
| **MaxArrayResults** | 100,000 | Maximum siblings collected by #, .N, #.field, filters, * and Map() |
| **MaxWildcardResults** | 100,000 | Maximum results from recursive ** queries |
| **MaxRecursiveOperations** | 10,000 | Minimum operation budget shared by the ** segments of a query (scales with document size) |
| **MaxFilterExpressionLength** | 4096 bytes | Maximum filter expression length |
| **MaxPatternIterations** | 10,000,000 | Step budget for a single pattern match, plus one step per value byte |
| **MaxModifierChainDepth** | 20 | Maximum chained modifiers |
| **MaxPathSegments** | 100 segments | Maximum path length |
| **MaxFieldNameLength** | 256 bytes | Maximum #.field name length |

### When to Adjust Limits

The defaults are sized so that legitimate documents rarely hit them, while bounding
work and memory that grow faster than the input. Applications that process trusted
documents may raise them; applications with tighter memory budgets may lower them.

Set limits once during program initialization, before any concurrent use of the
package. They are read without synchronization.

```go
func init() {
    // Trusted network device configurations can exceed the default size
    xmldot.MaxDocumentSize = 512 * 1024 * 1024
}
```

## Safe Usage Patterns

### 1. Validate Untrusted Input

**Always validate XML from untrusted sources:**

```go
func processUserXML(userXML string) error {
    // Step 1: Check size
    if len(userXML) > xmldot.MaxDocumentSize {
        return errors.New("document too large")
    }

    // Step 2: Validate well-formedness
    if !xmldot.Valid(userXML) {
        return errors.New("invalid XML")
    }

    // Step 3: Safe to process
    result := xmldot.Get(userXML, expectedPath)
    // Process result...

    return nil
}
```

### 2. Use ValidateWithError for Detailed Diagnostics

**Get detailed error information:**

```go
func processXMLWithErrorReporting(xml string) error {
    if err := xmldot.ValidateWithError(xml); err != nil {
        return fmt.Errorf("invalid XML at line %d, column %d: %s",
            err.Line, err.Column, err.Message)
    }

    // Safe to process
    result := xmldot.Get(xml, path)
    return nil
}
```

### 3. Avoid Optimistic Mode with Untrusted Input

**Use default validation for untrusted XML:**

```go
// Safe: Default mode validates
newXML, err := xmldot.Set(untrustedXML, path, value)
if err != nil {
    return fmt.Errorf("set failed: %w", err)
}

// Unsafe: Optimistic mode skips validation
opts := &xmldot.Options{Optimistic: true}
newXML, _ := xmldot.SetWithOptions(untrustedXML, path, value, opts)
// May produce invalid XML from malformed input
```

**When Optimistic mode is safe:**

```go
// Safe: XML from trusted database
dbXML := loadFromDatabase(id)

// Optimistic mode OK for trusted source
opts := &xmldot.Options{Optimistic: true}
newXML, _ := xmldot.SetWithOptions(dbXML, path, value, opts)
// 2-3x faster, safe because source is trusted
```

### 4. Sanitize User Input Before Set Operations

**Escape user data before inserting:**

```go
func updateConfig(xml, userValue string) (string, error) {
    // User input is automatically escaped in Set()
    newXML, err := xmldot.Set(xml, "config.value", userValue)
    if err != nil {
        return "", err
    }

    // Validate result
    if !xmldot.Valid(newXML) {
        return "", errors.New("resulting XML is invalid")
    }

    return newXML, nil
}
```

### 5. Limit Query Complexity

**Use specific paths, not unbounded wildcards:**

```go
// Safe: Specific path
result := xmldot.Get(userXML, "data.users.user.name")

// Risky: Recursive wildcard on untrusted input
result := xmldot.Get(userXML, "data.**.name")
// Potential CPU/memory exhaustion if document is huge

// Better: Limit wildcard scope
result := xmldot.Get(userXML, "data.users.*.name")  // Single-level only
```

### 6. Implement Request Size Limits

**Enforce size limits before processing:**

```go
func handleXMLUpload(w http.ResponseWriter, r *http.Request) {
    // Limit request body size
    r.Body = http.MaxBytesReader(w, r.Body, int64(xmldot.MaxDocumentSize))

    xmlData, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
        return
    }

    // Safe to process (already size-limited)
    result := xmldot.Get(string(xmlData), path)
    // ...
}
```

### 7. Use GetMany for Multiple Trusted Paths

**Batch operations are safe with trusted paths:**

```go
// Safe: Trusted, known paths
results := xmldot.GetMany(xml,
    "config.database.host",
    "config.database.port",
    "config.database.name")

// Unsafe: User-controlled paths (path injection risk)
userPaths := []string{req.GetString("path1"), req.GetString("path2")}
results := xmldot.GetMany(xml, userPaths...)  // Don't do this!
```

### 8. Path Injection Prevention

**Validate user-provided paths:**

```go
// Unsafe: User controls path
userPath := r.FormValue("path")
result := xmldot.Get(xml, userPath)  // Path injection risk

// Safe: Whitelist allowed paths
allowedPaths := map[string]bool{
    "user.name":  true,
    "user.email": true,
    "user.age":   true,
}

userPath := r.FormValue("path")
if !allowedPaths[userPath] {
    return errors.New("invalid path")
}

result := xmldot.Get(xml, userPath)  // Safe (whitelisted)
```

### 9. Output Sanitization

**Sanitize output for different contexts:**

```go
// HTML context: escape XML output
xmlContent := xmldot.Get(xml, path).String()
safeHTML := html.EscapeString(xmlContent)

// JavaScript context: JSON encode
xmlContent := xmldot.Get(xml, path).String()
safeJS, _ := json.Marshal(xmlContent)

// SQL context: use prepared statements
xmlContent := xmldot.Get(xml, path).String()
_, err := db.Exec("INSERT INTO data (value) VALUES (?)", xmlContent)
```

---

## Unsafe Patterns to Avoid

### 1. User-Controlled Paths Without Validation

**Unsafe**:
```go
// Attack: User injects recursive wildcard
userPath := r.FormValue("path")  // "root.**"
result := xmldot.Get(hugeXML, userPath)
// CPU/memory exhaustion
```

**Safe**:
```go
// Whitelist approach
allowedPaths := []string{"user.name", "user.email"}
if !contains(allowedPaths, userPath) {
    return errors.New("path not allowed")
}
result := xmldot.Get(xml, userPath)
```

### 2. Optimistic Mode with Untrusted Input

**Unsafe**:
```go
// Skips validation on untrusted input
opts := &xmldot.Options{Optimistic: true}
newXML, _ := xmldot.SetWithOptions(userXML, path, userValue, opts)
// May produce invalid XML
```

**Safe**:
```go
// Validate first, then use optimistic mode
if !xmldot.Valid(userXML) {
    return errors.New("invalid XML")
}

opts := &xmldot.Options{Optimistic: true}
newXML, _ := xmldot.SetWithOptions(userXML, path, trustedValue, opts)
```

### 3. Ignoring ValidateWithError Details

**Unsafe**:
```go
// Ignoring validation errors
xml := processUserInput()
xmldot.Get(xml, path)  // May fail silently on malformed XML
```

**Safe**:
```go
// Check validation errors
xml := processUserInput()
if err := xmldot.ValidateWithError(xml); err != nil {
    log.Errorf("XML validation failed at line %d: %s", err.Line, err.Message)
    return errors.New("invalid XML")
}
result := xmldot.Get(xml, path)
```

### 4. Unbounded Document Sizes from User Input

**Unsafe**:
```go
// No size check on user upload
uploadedXML, _ := io.ReadAll(r.Body)  // Could be 1GB+
result := xmldot.Get(string(uploadedXML), path)
// Memory exhaustion
```

**Safe**:
```go
// Limit read size
limitedReader := io.LimitReader(r.Body, int64(xmldot.MaxDocumentSize))
uploadedXML, _ := io.ReadAll(limitedReader)
result := xmldot.Get(string(uploadedXML), path)
```

### 5. SetRaw with Unvalidated User Input

**Unsafe**:
```go
// User provides raw XML without validation
userXML := r.FormValue("xml")
newXML, _ := xmldot.SetRaw(xml, path, userXML)
// Potential XML injection
```

**Safe**:
```go
// Validate user XML first
userXML := r.FormValue("xml")
if !xmldot.Valid("<root>" + userXML + "</root>") {
    return errors.New("invalid XML fragment")
}

// Additional validation: Check for script tags, etc.
if containsDangerousTags(userXML) {
    return errors.New("XML contains dangerous content")
}

newXML, _ := xmldot.SetRaw(xml, path, userXML)
```

---

## Concurrency Safety

### Thread-Safe Operations

**All XMLDOT functions are safe for concurrent use:**

```go
// Safe: Concurrent Get operations
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        result := xmldot.Get(xml, path)  // Thread-safe
        process(result)
    }(i)
}
wg.Wait()
```

**Path cache is thread-safe:**

```go
// Safe: Concurrent path parsing uses sync.RWMutex
go func() {
    result := xmldot.Get(xml, "root.element")  // Cache read
}()
go func() {
    result := xmldot.Get(xml, "root.other")    // Cache write
}()
// No data races
```

### Unsafe Patterns: Concurrent Writes

**Unsafe: Concurrent modifications to same XML:**

```go
// Unsafe: Data race on shared XML variable
var xml string = loadXML()

go func() {
    xml, _ = xmldot.Set(xml, "path1", value1)  // Race condition
}()
go func() {
    xml, _ = xmldot.Set(xml, "path2", value2)  // Race condition
}()
// Lost updates, corrupted XML
```

**Safe: Synchronize writes:**

```go
var (
    xml   string
    xmlMu sync.Mutex
)

go func() {
    xmlMu.Lock()
    xml, _ = xmldot.Set(xml, "path1", value1)
    xmlMu.Unlock()
}()
go func() {
    xmlMu.Lock()
    xml, _ = xmldot.Set(xml, "path2", value2)
    xmlMu.Unlock()
}()
```

**Better: Batch operations:**

```go
// Better: Use SetMany for atomic batch updates
xml, _ := xmldot.SetMany(xml,
    []string{"path1", "path2"},
    []interface{}{value1, value2})
// Single atomic operation, no synchronization needed
```

---

## Summary

XMLDOT provides comprehensive security protection:

**Built-in Protections**:
- ✅ XXE prevention (no external entity resolution)
- ✅ Entity expansion protection (no entity processing)
- ✅ Document size limits (100 MB default)
- ✅ Nesting depth limits (100 levels)
- ✅ Attribute flood protection (1000 attributes)
- ✅ Token size limits (1 MB tokens)
- ✅ Filter limits (4096-byte expressions, 10 levels, bounded pattern matching)
- ✅ Array and wildcard result limits (100,000 results)
- ✅ Path segment limits (100 segments)
- ✅ Recursive operation budget (scales with document size)
- ✅ Set value size limits (100 MB values)
- ✅ Automatic XML escaping (injection prevention)

**Safe Usage Checklist**:
- [ ] Validate untrusted XML with Valid() or ValidateWithError()
- [ ] Check Result.Err when a missing result matters
- [ ] Check document size before processing
- [ ] Avoid Optimistic mode with untrusted input
- [ ] Whitelist allowed paths if user-controlled
- [ ] Use specific paths instead of recursive wildcards
- [ ] Implement request size limits in HTTP handlers
- [ ] Sanitize output for target context (HTML, JS, SQL)
- [ ] Synchronize concurrent writes to shared XML

**Security Limits**:
- All limits enabled by default
- Adjustable package variables, set once at initialization
- Exceeding a limit returns a *LimitError, never a truncated result
- Predictable resource usage

For more information:
- [Path Syntax Reference](path-syntax.md)
- [Performance Guide](performance.md)
- [Migration Guide](migration.md)
- [Security Policy](../SECURITY.md)

---

**Document Version**: 1.0
**Last Updated**: 2025-10-08
**Status**: Complete

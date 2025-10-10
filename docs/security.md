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
- **Fail-safe behavior**: Graceful degradation, not crashes
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

### 3. Document Size Limits

**Threat**: Extremely large documents can exhaust memory.

**Protection**:
```go
const MaxDocumentSize = 10 * 1024 * 1024  // 10MB default

// Large document rejected
largeXML := make([]byte, 20*1024*1024)  // 20MB
result := xmldot.Get(string(largeXML), "root.element")
fmt.Println(result.Exists())  // false (document too large)

// Memory exhaustion prevented
// Predictable memory usage
```

**Safe Usage**:
```go
// Check size before processing
xmlData := loadFromNetwork()

if len(xmlData) > xmldot.MaxDocumentSize {
    return errors.New("document exceeds size limit")
}

// Safe to process
result := xmldot.Get(string(xmlData), path)
```

### 4. Nesting Depth Limits

**Threat**: Deeply nested XML can cause stack overflow.

**Protection**:
```go
const MaxNestingDepth = 100  // 100 levels default

// Attack: Deeply nested XML
nested := "<root>"
for i := 0; i < 200; i++ {
    nested += "<level" + strconv.Itoa(i) + ">"
}
nested += "value"
for i := 199; i >= 0; i-- {
    nested += "</level" + strconv.Itoa(i) + ">"
}
nested += "</root>"

// xmldot protection: Parsing stops at depth 100
// Deeper elements ignored (fail-safe truncation)
// No stack overflow
// Predictable behavior
```

**Technical Details**:
- Parser tracks nesting depth
- Elements beyond MaxNestingDepth are skipped
- Fail-safe behavior: truncation, not error
- Stack overflow prevented

### 5. Attribute Flood Protection

**Threat**: Elements with thousands of attributes can exhaust memory during parsing.

**Protection**:
```go
const MaxAttributes = 100  // 100 attributes per element

// Attack: Attribute flooding
attack := "<element"
for i := 0; i < 1000; i++ {
    attack += fmt.Sprintf(" attr%d=\"value%d\"", i, i)
}
attack += ">content</element>"

// xmldot protection: Only first 100 attributes parsed
// Excess attributes silently ignored (fail-safe)
// Memory usage bounded
// Attack mitigated
```

**Example**:
```go
// Element with 200 attributes (attack)
xml := `<item ` + generateAttributes(200) + `>value</item>`

// Only first 100 attributes accessible
for i := 0; i < 150; i++ {
    attrPath := fmt.Sprintf("item.@attr%d", i)
    result := xmldot.Get(xml, attrPath)
    if i < 100 {
        fmt.Println(result.Exists())  // true
    } else {
        fmt.Println(result.Exists())  // false (excess ignored)
    }
}
```

### 6. Token Size Limits

**Threat**: Extremely large element names or attribute values can cause buffer overflows.

**Protection**:
```go
const MaxTokenSize = 1024 * 1024  // 1MB per token

// Attack: Huge element name
hugeName := strings.Repeat("a", 2*1024*1024)  // 2MB name
attack := "<" + hugeName + ">value</" + hugeName + ">"

// xmldot protection: Token truncated at 1MB
// Parser continues safely (fail-safe)
// No buffer overflow
// Memory bounded
```

**Technical Details**:
- Element names, attribute names, attribute values limited
- Tokens exceeding MaxTokenSize are truncated
- Fail-safe behavior: truncation or empty result
- Buffer overflow prevented

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

### 8. Filter Depth Limits

**Threat**: Deeply nested filter expressions can cause stack overflow.

**Protection**:
```go
const MaxFilterDepth = 10  // 10 levels of nested filters

// Attack: Deeply nested filter
deepFilter := "root.#(a.#(b.#(c.#(d.#(e.#(f.#(g.#(h.#(i.#(j.#(k>10)))))))))))"
result := xmldot.Get(xml, deepFilter)
// Returns Null (exceeds MaxFilterDepth)
// No stack overflow
// Attack prevented
```

**Safe Usage**:
```go
// Simple filter (safe)
result := xmldot.Get(xml, "items.item.#(price>100)")  // Depth 1

// Multiple filters (safe)
result := xmldot.Get(xml, "items.item.#(price>100).#(@active==true)")  // Depth 1

// Nested element filter (moderate depth)
result := xmldot.Get(xml, "items.item.#(category.type==electronics)")  // Depth 2
```

### 9. Wildcard Result Limits

**Threat**: Recursive wildcards on large documents can consume excessive memory.

**Protection**:
```go
const MaxWildcardResults = 1000  // 1000 results maximum

// Attack: Recursive wildcard on huge document with 10,000 matches
largeXML := generateXMLWithManyElements(10000)
result := xmldot.Get(largeXML, "root.**.item")

// xmldot protection: Only first 1000 matches returned
matches := result.Array()
fmt.Println(len(matches))  // Maximum 1000
// Memory usage bounded
// DoS prevented
```

**Safe Usage**:
```go
// Use specific paths when possible (no limit)
result := xmldot.Get(xml, "root.catalog.items.item")  // No wildcard

// Limit wildcard scope
result := xmldot.Get(xml, "root.catalog.*.item")  // Single-level (faster)
```

### 10. Path Segment Limits

**Threat**: Extremely long paths can exhaust memory during parsing.

**Protection**:
```go
const MaxPathSegments = 100  // 100 segments maximum

// Attack: Extremely long path
longPath := "root" + strings.Repeat(".child", 150)  // 150 segments
result := xmldot.Get(xml, longPath)
// Returns Null (exceeds MaxPathSegments)
// Memory bounded
```

### 11. Recursive Operation Limits

**Threat**: Recursive wildcard queries can cause CPU exhaustion.

**Protection**:
```go
const MaxRecursiveOperations = 10000  // 10,000 operations max

// Attack: Recursive wildcard on deeply nested document
deepXML := generateDeeplyNestedXML(1000)  // 1000 levels
result := xmldot.Get(deepXML, "root.**.item")

// xmldot protection: Search stops after 10,000 operations
// CPU usage bounded
// DoS prevented
```

### 12. Set Operation Value Limits

**Threat**: Setting extremely large values can cause memory exhaustion.

**Protection**:
```go
const MaxValueSize = 5 * 1024 * 1024  // 5MB maximum value

// Attack: Set huge value
hugeValue := strings.Repeat("x", 10*1024*1024)  // 10MB
newXML, err := xmldot.Set(xml, "root.element", hugeValue)
// Returns error: value too large
// Memory exhaustion prevented
```

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

Complete reference of all security limits with rationale:

| Constant | Default Value | Purpose | Rationale |
|----------|---------------|---------|-----------|
| **MaxDocumentSize** | 10 MB | Maximum XML document size | Prevents memory exhaustion from huge documents |
| **MaxValueSize** | 5 MB | Maximum value in Set operations | Prevents memory exhaustion via large values |
| **MaxNestingDepth** | 100 levels | Maximum element nesting | Prevents stack overflow from deeply nested XML |
| **MaxAttributes** | 100 per element | Maximum attributes per element | Prevents memory exhaustion from attribute flooding |
| **MaxTokenSize** | 1 MB | Maximum token size (name, value) | Prevents buffer overflow from huge tokens |
| **MaxNamespacePrefixLength** | 256 bytes | Maximum namespace prefix length | Prevents memory exhaustion from long prefixes |
| **MaxFilterDepth** | 10 levels | Maximum filter nesting depth | Prevents stack overflow from nested filters |
| **MaxFilterExpressionLength** | 256 bytes | Maximum filter expression length | Prevents parsing overhead from long filters |
| **MaxWildcardResults** | 1000 results | Maximum wildcard match results | Prevents memory exhaustion from many matches |
| **MaxRecursiveOperations** | 10,000 ops | Maximum recursive search operations | Prevents CPU exhaustion from complex wildcards |
| **MaxPathSegments** | 100 segments | Maximum path length | Prevents memory exhaustion from long paths |

### When to Adjust Limits

**Increase limits when**:
- Processing trusted internal documents
- Documents naturally exceed limits (e.g., 20MB configs)
- Performance testing shows safe headroom
- Running in controlled, isolated environments

**Example**:
```go
// Adjust limits for trusted environment (NOT recommended for untrusted input)
// Note: Limits are package constants, cannot be changed at runtime

// For documents >10MB, validate and chunk:
if len(trustedXML) > xmldot.MaxDocumentSize {
    // Process in sections
    sections := splitXMLIntoSections(trustedXML)
    for _, section := range sections {
        result := xmldot.Get(section, path)
        // Process section
    }
}
```

**Never adjust limits when**:
- Processing untrusted user input
- Accepting XML from external APIs
- Handling uploaded files
- Operating in shared/multi-tenant environments

### Performance vs Security Trade-offs

| Limit | Performance Impact | Security Benefit | Recommendation |
|-------|-------------------|------------------|----------------|
| MaxDocumentSize | High (rejects large docs) | High (prevents memory exhaustion) | Keep default |
| MaxNestingDepth | Low (rarely hit) | High (prevents stack overflow) | Keep default |
| MaxAttributes | Low (rarely hit) | High (prevents DoS) | Keep default |
| MaxTokenSize | Low (rarely hit) | High (prevents buffer overflow) | Keep default |
| MaxWildcardResults | Medium (limits matches) | High (prevents memory exhaustion) | Keep default |
| MaxRecursiveOperations | Medium (limits recursion) | High (prevents CPU exhaustion) | Keep default |
| MaxPathSegments | Low (rarely hit) | Medium (prevents complex paths) | Can increase for deep paths |

---

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
    r.Body = http.MaxBytesReader(w, r.Body, xmldot.MaxDocumentSize)

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
limitedReader := io.LimitReader(r.Body, xmldot.MaxDocumentSize)
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
- ✅ Document size limits (10MB default)
- ✅ Nesting depth limits (100 levels)
- ✅ Attribute flood protection (100 attributes)
- ✅ Token size limits (1MB tokens)
- ✅ Filter depth limits (10 levels)
- ✅ Wildcard result limits (1000 results)
- ✅ Path segment limits (100 segments)
- ✅ Recursive operation limits (10,000 ops)
- ✅ Set value size limits (5MB values)
- ✅ Automatic XML escaping (injection prevention)

**Safe Usage Checklist**:
- [ ] Validate untrusted XML with Valid() or ValidateWithError()
- [ ] Check document size before processing
- [ ] Avoid Optimistic mode with untrusted input
- [ ] Whitelist allowed paths if user-controlled
- [ ] Use specific paths instead of recursive wildcards
- [ ] Implement request size limits in HTTP handlers
- [ ] Sanitize output for target context (HTML, JS, SQL)
- [ ] Synchronize concurrent writes to shared XML

**Security Limits**:
- All limits enabled by default
- No configuration required
- Fail-safe behavior (truncation, not crashes)
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

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **BREAKING: Security limits fail instead of truncating**: Exceeding any security limit now fails the operation instead of returning a partial, truncated or empty result that is indistinguishable from a missing element.
  - `Get` and related functions return a Null `Result` with the new `Result.Err` field set to a `*LimitError`. `Result.Get`, `GetWithOptions` and `GetMany` propagate the error. `Result.Err` is excluded from JSON encoding.
  - `Set`, `Delete` and related functions return a `*LimitError` instead of `ErrMalformedXML`, `ErrInvalidValue` or `ErrInvalidPath` (for example for paths longer than `MaxPathSegments`), including when validation fails because of a limit.
  - `Set` now also checks the resulting document against all limits (size, nesting depth, attributes, token size), so it never produces a document that later operations reject. This adds a validation pass to every `Set`.
  - `Result.Map()` and `MapWithOptions()` return `nil` instead of an incomplete map, and also when called on a failed `Result`.
  - `ValidateError` has a new `Err` field and `Unwrap()` method, so `errors.Is(err, ErrLimitExceeded)` reports limit violations found during validation. Malformed input is still reported as malformed, also beyond `MaxAttributes`.
  - New `ErrLimitExceeded` sentinel and `LimitError` type reporting the limit name, the value observed when the check failed and the effective maximum.
  - `MaxNestingDepth` is now enforced by all queries, not only by recursive `**` searches and validation, so Get, Set and validation apply the same depth limit.
  - Query paths with namespace prefixes longer than `MaxNamespacePrefixLength` fail with a `*LimitError`. Element names in documents with such prefixes are still matched as unprefixed names.
- **BREAKING: Security limits are variables**: All `Max*` limits except `MaxPositionOffset` are now `int` package variables in `limits.go` instead of untyped constants, so applications processing trusted documents can adjust them during initialization. Code using them as constants or as other integer types must be updated, for example `io.LimitReader(r, int64(xmldot.MaxDocumentSize))`. Cached path parses are invalidated when path-related limits change.
- **Security limit defaults revised**: Limits that only capped input already bounded by the document size were raised; limits that guard against superlinear work were kept or made proportional to the input.
  - `MaxDocumentSize`: 10 MB → 100 MB
  - `MaxValueSize`: 5 MB → 100 MB
  - `MaxAttributes`: 100 → 1000
  - `MaxFilterExpressionLength`: 256 → 4096
  - `MaxWildcardResults`: 1000 → 100,000
  - `MaxRecursiveOperations` is now the minimum budget shared by all `**` segments of a query; the effective budget is the larger of this value and the document size in bytes. Previously each recursive search silently stopped after about 10,000 elements, while chained searches each got a new budget.
  - `MaxPatternIterations`: 10,000 → 10,000,000, plus one step per byte of the matched value.
  - New `MaxArrayResults` (100,000) for non-recursive array operations. Indexed access (`.N`) and first-match filters (`#(...)`) only collect siblings up to the answer, and all-match filters (`#(...)#`) count only matching siblings.
- **Linear-space pattern matching**: The `%` and `!%` filter operators use a greedy glob matcher with O(len(value) + len(pattern)) memory and an immediate match for a trailing `*`. The previous implementation allocated a len(value) × len(pattern) table before applying its iteration limit. Match results are unchanged.
- **`MaxTokenSize` counting aligned with validation**: Processing instructions and closing tags are counted from after `<?` and `</`, as the validator does.

### Removed

- **BREAKING: `MaxFilterDepth`**: Filters cannot be nested, so filter evaluation never recursed and the limit had no effect. Code referencing it must drop the reference.

### Fixed

- **Markup inside comments and CDATA sections**: Element content is now scanned with comments and CDATA sections copied verbatim. Previously, tag-like text inside them was parsed as elements, which could rewrite the content, end an element early (for example `<![CDATA[</a>]]>`), or trip `MaxNestingDepth` and `MaxTokenSize` on valid documents.
- **Filters containing dots**: Dots inside `#(...)` were split as path separators, so filters with a nested condition path (`#(specs.cpu==M4)`), a decimal value (`#(price==19.99)`) or a quoted value containing a dot never matched. Backslash escapes inside filters are now resolved by the filter parser: `%` and `!%` patterns keep `\*`, `\?` and `\\` as literal-character escapes, and an escaped closing quote no longer terminates a quoted value.
- **Panic on filters with a lone quote value**: Paths such as `#(name%")` or `#(name==')` panicked with a slice bounds error in `Get`, `GetWithOptions` and `Set`. Such filters are now rejected as invalid paths.
- **Non-recursive array operations no longer truncate at 1000 elements**: Counts (`#`), indexed access (`.N`), field extraction (`#.field`), filters (`#(...)#`), single-level wildcards (`*`), multi-root fragments and `Map()` silently dropped every sibling past the 1000th. These operations now use the new `MaxArrayResults` limit.

## [0.5.1] - 2025-12-18

### Fixed

- **`Result.Get()` array operations now work on multi-root fragments**: Fixed bug where array operations (`#`, `#.field`, indexed access) failed when called via `Result.Get()` on Element results containing multiple sibling elements.

## [0.5.0] - 2025-11-08

### Added

- **`Map()` and `MapWithOptions()` methods for structure inspection**: New fluent API methods return immediate children as `map[string]Result` for dynamic field access and introspection.
  - `result.Map()` returns child elements with duplicate names combined into Array Results
  - `result.MapWithOptions()` supports case-insensitive element name matching
  - Mixed content text stored under `"%"` key
  - Security limits enforced (max 1000 children, silent truncation)

## [0.4.3] - 2025-11-07

### Fixed

- **Append to empty array now inserts at end of parent content**: When appending to an empty array using `-1` index (no matching elements exist), new elements are now correctly inserted at the end of the parent's content instead of the beginning. This preserves document order and matches user expectations for "append" semantics.
  - Before: `<root><other>data</other></root>` + `root.item.-1` → `<root><item>NEW</item><other>data</other></root>`
  - After: `<root><other>data</other></root>` + `root.item.-1` → `<root><other>data</other><item>NEW</item></root>`
  - Non-empty arrays unchanged: Existing behavior of inserting after the last matching element is preserved

## [0.4.2] - 2025-11-07

### Fixed

- **Root-level append with `-1` index now creates siblings instead of nesting**: Previously, appending to root level (e.g., `Set("<user>Alice</user>", "item.-1", "first")`) incorrectly nested the new element inside the first root element. Now it correctly creates sibling root elements, enabling proper multi-root fragment building: `<user>Alice</user><item>first</item>`

## [0.4.1] - 2025-10-31

### Fixed

- **`Set()` now works with empty XML and creates sibling roots in fragments**: Previously, `Set()` would fail on empty XML or incorrectly try to nest paths under existing roots even when a different root was specified. Now:
  - Empty XML is accepted: `Set("", "root.child", "value")` creates valid XML from scratch
  - Different root names create siblings
  - Matching root names still nest inside existing roots (preserves expected behavior)
  - Enables building multi-root XML fragments incrementally

## [0.4.0] - 2025-10-31

### Added

- **XML fragment support**: xmldot now supports XML fragments with multiple root elements. Fragments with matching root names can be treated as arrays using standard array syntax. This is useful for processing log entries, streaming data, and partial XML documents.
  - Validation accepts multiple root elements: `Valid("<item>A</item><item>B</item>")` → true
  - Get/Set/Delete operations work with fragments (operate on first matching root element)
  - Array operations on matching roots: `Get(fragment, "user.#")`, `Get(fragment, "user.0.name")`, `Get(fragment, "user.#.name")`
  - Whitespace and comments between roots are preserved
  - Text content between roots is still rejected (maintains well-formedness)
  - All security limits remain enforced

- **Array append operations using `-1` index**: `Set()` and `SetRaw()` now support using index `-1` to append elements to arrays. This provides an intuitive way to add items without calculating array length.
  - `SetRaw(xml, "items.item.-1", "<name>New Item</name>")` appends to existing array
  - Creates first element when array is empty: `SetRaw(xml, "root.item.-1", "<value>First</value>")`
  - Auto-creates parent paths when needed
  - Treats single elements as 1-element arrays and appends a second
  - Limitations: Only supported in Set/SetRaw operations (not Get/Delete), nested paths after `-1` are not allowed

### Fixed

- **Array iteration examples**: Updated examples to use proper field extraction syntax (`#.field`) for iterating over array elements, fixing previous incorrect usage patterns

## [0.3.1] - 2025-10-27

### Fixed

- **@pretty modifier no longer duplicates xmlns declarations**: The `@pretty` modifier previously added duplicate `xmlns` namespace declarations when pretty-printing XML with namespaces.

## [0.3.0] - 2025-10-27

### Changed

- **Auto-create parent elements when setting attributes**: `Set()` now automatically creates missing parent elements when setting attributes on non-existent paths. Previously, this would return an error "parent element not found for attribute". This is a **breaking behavior change** for error handling but makes the API more consistent with element creation behavior.
  - Example: `Set("<root></root>", "root.user.@id", "123")` now succeeds and creates `<root><user id="123"></user></root>`
  - Before v0.3.0: This would return an error
  - After v0.3.0: Element is automatically created with the attribute

## [0.2.0] - 2025-10-18

### Added

- **Fluent API methods on Result type** for method chaining:
  - `Result.Get(path string) Result` - Query Result's XML content with path
  - `Result.GetMany(paths ...string) []Result` - Batch query multiple paths
  - `Result.GetWithOptions(path string, opts *Options) Result` - Query with custom options

## [0.1.0] - 2025-10-16

### Added

- Initial release

---

[Unreleased]: https://github.com/netascode/xmldot/compare/v0.5.1...HEAD
[0.5.1]: https://github.com/netascode/xmldot/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/netascode/xmldot/compare/v0.4.3...v0.5.0
[0.4.3]: https://github.com/netascode/xmldot/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/netascode/xmldot/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/netascode/xmldot/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/netascode/xmldot/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/netascode/xmldot/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/netascode/xmldot/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/netascode/xmldot/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/netascode/xmldot/releases/tag/v0.1.0

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/netascode/xmldot/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/netascode/xmldot/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/netascode/xmldot/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/netascode/xmldot/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/netascode/xmldot/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/netascode/xmldot/releases/tag/v0.1.0

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[0.3.0]: https://github.com/netascode/xmldot/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/netascode/xmldot/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/netascode/xmldot/releases/tag/v0.1.0

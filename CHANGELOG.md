# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.2.0]: https://github.com/netascode/xmldot/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/netascode/xmldot/releases/tag/v0.1.0

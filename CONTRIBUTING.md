# Contributing to xmldot

Thank you for your interest in contributing to xmldot! This document provides guidelines for contributing to the project.

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the problem
- **Expected behavior** and **actual behavior**
- **Code samples** demonstrating the issue
- **Go version** and **OS information**

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).

### Suggesting Features

Feature suggestions are welcome! Please:

- **Check existing feature requests** to avoid duplicates
- **Provide clear use cases** for the feature
- **Explain why this feature would be useful** to most users
- **Consider implementation complexity** and performance impact

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md).

### Pull Requests

1. **Fork the repository** and create a branch from `main`
2. **Make your changes** following the coding guidelines below
3. **Add tests** for any new functionality
4. **Ensure all tests pass**: `make test`
5. **Run the linter**: `make lint`
6. **Update documentation** as needed
7. **Write a clear commit message** following conventional commits
8. **Submit a pull request** using the PR template

## Development Setup

### Prerequisites

- Go 1.24 or later
- Make (optional but recommended)
- golangci-lint (for linting)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/netascode/xmldot.git
cd xmldot

# Install dependencies
go mod download

# Run tests
make test

# Run linter
make lint

# Run benchmarks
make bench
```

## Coding Guidelines

### License Headers

The project uses [Google's addlicense](https://github.com/google/addlicense) tool for license header management.

**Install addlicense:**
```bash
go install github.com/google/addlicense@v1.1.1
# Or use: make tools
```

**Version Note**: Use the exact version shown above (`@v1.1.1`) to match the CI environment. Using `@latest` may cause version mismatches between local development and CI checks.

**Verify headers:**
```bash
make check-license
```

**Important Notes:**
- **For Maintainers**: Use `make license` to add headers with project default copyright holder (Daniel Schmidt)
- **For Contributors**: Do NOT use `make license` - it will overwrite copyright holder with the project default
- **For Contributors**: Manually add headers to new files with YOUR organization's copyright:
  ```go
  // SPDX-License-Identifier: MIT
  // Copyright (c) 2025 Your Organization Name

  package xmldot
  ```
- Then run `make check-license` to verify format compliance (accepts any copyright holder)
- Both SPDX-first and Copyright-first formats are accepted by the check command

### Go Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `go fmt` for formatting
- Use meaningful variable and function names
- Write clear comments for exported functions

### Testing

- Write unit tests for all new functionality
- Aim for >80% code coverage
- Include edge cases and error conditions
- Use table-driven tests where appropriate
- Add benchmarks for performance-critical code
- Run race detector: `go test -race`

Example test:

```go
func TestGet(t *testing.T) {
    tests := []struct {
        name     string
        xml      string
        path     string
        expected string
    }{
        {
            name:     "simple element",
            xml:      `<root><child>value</child></root>`,
            path:     "root.child",
            expected: "value",
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Get(tt.xml, tt.path)
            if result.String() != tt.expected {
                t.Errorf("got %q, want %q", result.String(), tt.expected)
            }
        })
    }
}
```

### Documentation

- Document all exported functions, types, and constants
- Include usage examples in godoc comments
- Update README.md for significant changes
- Add examples to `examples/` for new features

Example documentation:

```go
// Get searches xml for the specified path and returns a Result containing
// the value found. If the path is not found, an empty Result is returned.
//
// The path syntax supports:
//   - Element access: "root.child.element"
//   - Attribute access: "element.@attribute"
//   - Array indexing: "elements.element.0"
//
// Example:
//
//	xml := `<root><user><name>John</name></user></root>`
//	name := Get(xml, "root.user.name")
//	fmt.Println(name.String()) // "John"
func Get(xml, path string) Result {
    // Implementation...
}
```

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test changes
- `perf`: Performance improvements
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Examples:

```
feat: add support for recursive wildcards in paths
fix: handle empty XML documents correctly
docs: update path syntax examples in README
test: add edge cases for array indexing
perf: optimize path parsing with zero allocations
```

## Review Process

1. All PRs require at least one approval
2. CI must pass (tests, lint, coverage)
3. Code must follow style guidelines
4. Documentation must be updated
5. Breaking changes require discussion

## Questions?

Feel free to:

- Open an issue for questions
- Start a discussion in GitHub Discussions
- Reach out to maintainers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

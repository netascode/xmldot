---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## Steps to Reproduce

1. Step 1
2. Step 2
3. Step 3
4. See error

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

A clear and concise description of what actually happened.

## Code Sample

```go
// Paste your code here
package main

import "github.com/netascode/xmldot"

func main() {
    xml := `<root><child>value</child></root>`
    result := xmldot.Get(xml, "root.child")
    // Expected: "value"
    // Actual: something else
}
```

## Environment

- **Go version**: (e.g., 1.21.0)
- **xmldot version**: (e.g., v0.1.0 or commit hash)
- **Operating System**: (e.g., macOS 13.0, Ubuntu 22.04)
- **Architecture**: (e.g., amd64, arm64)

## Additional Context

Add any other context about the problem here (e.g., XML characteristics, performance observations, etc.).

## Possible Solution

If you have suggestions on how to fix the bug, please describe them here.

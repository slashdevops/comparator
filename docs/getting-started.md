# 🚀 Getting Started

## 📦 Installation

```bash
go get github.com/slashdevops/comparator
```

Update to the latest available version:

```bash
go get -u github.com/slashdevops/comparator
```

`comparator` requires **Go 1.26 or newer** and has **zero third-party
dependencies** — only the Go standard library.

## 👋 Your First Comparison

```go
package main

import (
    "fmt"

    "github.com/slashdevops/comparator"
)

func main() {
    a := map[string]int{"cpu": 2, "mem": 8}
    b := map[string]int{"cpu": 2, "mem": 8}

    if comparator.Equal(a, b) {
        fmt.Println("values are equal")
    }
}
```

`comparator.Equal` is a package-level convenience function. It builds a
comparator, runs a single comparison, and returns a `bool`. For repeated
comparisons with the same configuration, create a comparator once (see below).

## 🧱 Core Types

### The `Comparator` interface

Returned by [`New`](configuration.md) and
[`NewWithOptions`](configuration.md). Use it for equality checks:

```go
comp := comparator.New()

comp.Equal(a, b)                 // bool
comp.Diff(a, b)                  // ([]Difference, error)
comp.CompareWithDiff(a, b)       // *DiffResult
```

### The `DiffComparer` interface

Returned by [`NewDiffComparer`](diffing.md). It extends `Comparator` with the
richer diff and formatting operations:

```go
comp := comparator.NewDiffComparer()

comp.CompareWithDiff(a, b)       // *DiffResult
comp.GetUnifiedDiff(a, b)        // (*UnifiedDiff, error)
comp.GetJSONPatch(a, b)          // ([]JSONPatchOperation, error)
comp.GetVisualDiff(a, b)         // (*VisualDiff, error)
comp.FormatDiff(result, format)  // (string, error)
```

## 🧩 Package-level Convenience Functions

When you only need a single call, skip the constructor entirely:

| Function | Returns | Use for |
| -------- | ------- | ------- |
| `Equal(a, b, opts...)` | `bool` | One-off equality check. |
| `DeepEqual(a, b, opts...)` | `bool` | Same as `Equal`; familiar name for `reflect.DeepEqual` users. |
| `CompareWithDiff(a, b, opts...)` | `*DiffResult` | One-off detailed diff. |
| `GetJSONPatch(a, b, opts...)` | `([]JSONPatchOperation, error)` | One-off RFC 6902 patch. |

All of them accept the same [functional options](configuration.md) as the
constructors.

## ✅ Supported Types

The comparator handles every Go type:

- **Primitives**: `bool`, `int*`, `uint*`, `float*`, `complex*`, `string`
- **Composite**: arrays, slices, maps, structs
- **Reference**: pointers, interfaces, channels, functions
- **Special**: `time.Time` (compared with its native `Equal` method)

Recursive structures are handled safely via cycle detection.

## ➡️ Next Steps

- Tune behavior with [Configuration & Options](configuration.md).
- Understand diff output in [Diffing](diffing.md).
- Present diffs with [Output Formats](output-formats.md).

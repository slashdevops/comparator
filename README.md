# 🔬 comparator

[![main branch](https://github.com/slashdevops/comparator/actions/workflows/main.yml/badge.svg)](https://github.com/slashdevops/comparator/actions/workflows/main.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/slashdevops/comparator?style=plastic)
[![Go Reference](https://pkg.go.dev/badge/github.com/slashdevops/comparator.svg)](https://pkg.go.dev/github.com/slashdevops/comparator)
[![Go Report Card](https://goreportcard.com/badge/github.com/slashdevops/comparator)](https://goreportcard.com/report/github.com/slashdevops/comparator)
[![license](https://img.shields.io/github/license/slashdevops/comparator.svg)](https://github.com/slashdevops/comparator/blob/main/LICENSE)
[![Release](https://github.com/slashdevops/comparator/actions/workflows/release.yml/badge.svg)](https://github.com/slashdevops/comparator/actions/workflows/release.yml)

`comparator` is a small, dependency-free Go package for **deep comparison** and
**rich diffing** of arbitrary Go values — primitives, structs, slices, maps,
pointers, and complex nested structures. It answers two questions: *are these two
values equal?* and *how exactly do they differ?* — with output ready for
terminals, web UIs, APIs, and version control.

Only the Go standard library is used. There are no third-party dependencies.

## ✨ Features

- 🔁 **Deep recursive comparison** with cycle detection
- 🔍 **Difference detection** — detailed, path-addressed reports of what changed
- 🎨 **Multiple output formats** — text (with ANSI color), JSON, Markdown, HTML
- 📐 **Unified diff** — Unix `diff`-style output
- 🩹 **JSON Patch** — generate *and apply* RFC 6902 patch documents
- 🌳 **Visual tree diff** — hierarchical representation for viewers
- 🧠 **Custom equality** — per-type comparators plus auto-detected
  `Equatable`/`Comparable` interfaces
- 🔤 **Type-safe generics** — `EqualT[T]`, `DiffT[T]`, and friends
- ⏱️ **Context-aware** — cancel or bound long comparisons with `context.Context`
- 📡 **Streaming reporter** — observe each difference as it is found
- 🎨 **Pluggable formatters** — register custom output formats (CSV, JUnit, …)
- 🧪 **Testing helpers** — `AssertEqual` / `RequireEqual` for test suites
- ⚙️ **Configurable behavior** — float precision, slice-order insensitivity,
  field/path/map-key ignoring, depth limits, NaN handling, and more
- 🚫 **Zero third-party dependencies** — standard library only
- 📄 **Apache-2.0 licensed**

## 📦 Installation

```bash
go get github.com/slashdevops/comparator
```

### 🔄 Update

Update to the latest available version:

```bash
go get -u github.com/slashdevops/comparator
```

## 🧰 Requirements

- Go **1.26** or newer
- No external Go modules

## 🚀 Quick Start

### Simple equality check

```go
package main

import (
    "fmt"

    "github.com/slashdevops/comparator"
)

func main() {
    comp := comparator.New()
    if comp.Equal(obj1, obj2) {
        fmt.Println("objects are equal")
    }

    // Or use the package-level convenience function:
    if comparator.Equal(obj1, obj2) {
        fmt.Println("objects are equal")
    }
}
```

### Comparison with options

```go
comp := comparator.NewWithOptions(
    comparator.WithFloatPrecision(1e-6),
    comparator.IgnoreSliceOrder(),
    comparator.IgnoreStructFields("ID", "CreatedAt", "UpdatedAt"),
)

if comp.Equal(user1, user2) {
    fmt.Println("users are equal (ignoring ID and timestamps)")
}
```

### Detailed difference analysis

```go
diffComp := comparator.NewDiffComparer(
    comparator.WithOutputFormat("markdown"),
    comparator.WithMaxDiffs(100),
)

result := diffComp.CompareWithDiff(expected, actual)
if !result.Equal {
    fmt.Printf("found %d differences\n", len(result.Differences))
    fmt.Println(result.Summary)

    for _, diff := range result.Differences {
        fmt.Printf("[%s] %s: %s\n", diff.Severity, diff.Path, diff.Message)
    }
}
```

### JSON Patch (RFC 6902)

```go
patch, err := comparator.GetJSONPatch(oldDoc, newDoc)
if err != nil {
    log.Fatal(err)
}

out, _ := json.MarshalIndent(patch, "", "  ")
fmt.Println(string(out))
```

## 📖 Documentation

Full documentation lives in the [`docs/`](docs/) folder, and the API reference is
on [pkg.go.dev](https://pkg.go.dev/github.com/slashdevops/comparator).

| Guide | What it covers |
| ----- | -------------- |
| 🚀 [Getting Started](docs/getting-started.md) | Installation, first comparison, core types. |
| ⚙️ [Configuration & Options](docs/configuration.md) | Every option, its default, and when to use it. |
| 🔍 [Diffing](docs/diffing.md) | `DiffResult`, `Difference`, diff modes. |
| 🎨 [Output Formats](docs/output-formats.md) | Text/color, JSON, Markdown, HTML, unified, visual. |
| 🩹 [JSON Patch](docs/json-patch.md) | Generating and applying RFC 6902 patches. |
| 🧠 [Custom Comparators](docs/custom-comparators.md) | Custom equality + `Equatable`/`Comparable`. |
| 🧩 [Extensibility](docs/extensibility.md) | Generics, reporters, context, pluggable formatters. |
| 🧪 [Testing](docs/testing.md) | `AssertEqual` and friends for test suites. |
| ⚡ [Performance](docs/performance.md) | Cost model, benchmarks, tuning. |
| ❓ [FAQ](docs/faq.md) | Gotchas, thread-safety, common questions. |

## 🎛️ Configuration Options

| Option | Default | Description |
| ------ | ------- | ----------- |
| `WithFloatPrecision(float64)` | `1e-9` | Tolerance for float equality. |
| `IgnoreSliceOrder()` | `false` | Treat slices as sets (ignore order). |
| `WithMaxDepth(int)` | `0` (unlimited) | Limit recursion depth. |
| `IgnoreUnexported()` | `false` | Skip unexported struct fields. |
| `WithEquateEmpty(bool)` | `true` | Toggle nil/empty-container equivalence (both directions). |
| `EquateNaNs()` | `false` | Treat NaN values as equal. |
| `IgnoreStructFields(...string)` | none | Skip specific struct fields by name. |
| `WithIgnorePaths(...string)` | none | Skip struct fields at exact canonical paths. |
| `WithIgnorePathPatterns(...string)` | none | Skip struct fields whose path matches a regexp. |
| `WithIgnoreMapKeys(...string)` | none | Skip map entries by key. |
| `WithFieldNaming(FieldNaming)` | `GoFieldNaming` | Use Go field names or `json` tag names in paths/pointers. |
| `WithTimeLayout(string)` | `time.RFC3339Nano` | Layout for rendering `time.Time`. |
| `WithCustomComparator[T](func(T, T) bool)` | none | Register custom equality for a type. |
| `WithReporter(func(Difference))` | none | Stream each difference to a callback. |
| `WithDiffMode(DiffMode)` | `DiffModeSimple` | Diff detail level. |
| `WithMaxDiffs(int)` | `1000` | Cap the number of differences collected. |
| `WithOutputFormat(string)` | `"text"` | `text`, `json`, `markdown`, `html`, or a registered format. |
| `WithColorize(bool)` | `false` | ANSI colors in text output. |
| `WithIncludeEqual(bool)` | `false` | Include equal values in reports. |

Struct fields tagged `comparator:"-"` are always skipped. See
[Configuration & Options](docs/configuration.md) and
[Extensibility](docs/extensibility.md) for the full set.

See [Configuration & Options](docs/configuration.md) for details.

## 🧵 Thread Safety

Comparator instances are **not** thread-safe. Create separate instances for
concurrent use, or synchronize access with a mutex. The package-level helpers
build a fresh instance per call and are safe to call concurrently.

## 🧪 Testing

Run the test suite:

```bash
go test ./...
```

Run the runnable documentation examples:

```bash
go test -run Example ./...
```

Run benchmarks:

```bash
go test -run '^$' -bench . ./...
```

Run the same local quality checks used by CI:

```bash
go fmt ./...
go vet ./...
go test -race -coverprofile=/tmp/comparator-coverage.txt -covermode=atomic ./...
go build ./...
```

## 🗂️ Project Structure

```text
.
|-- .github/                            GitHub Actions, CodeQL, Dependabot, release metadata
|-- .golangci.yaml                      Optional local golangci-lint configuration
|-- docs/                               Extensive usage documentation
|-- AGENTS.md                           AI assistant guidance (symlink to copilot-instructions)
|-- doc.go                              Package documentation rendered by pkg.go.dev
|-- comparator.go                       Core comparison and diffing engine
|-- errors.go                           Sentinel errors
|-- generics.go                         Type-safe generic API (EqualT, DiffT, ...)
|-- context.go                          Context-aware comparison
|-- patch.go                            ApplyJSONPatch (RFC 6902 applier)
|-- formatters.go                       Pluggable output formatter registry
|-- options_ext.go                      Additional functional options
|-- stringer.go                         String() helpers for diff types
|-- assert.go                           Testing helpers (AssertEqual, ...)
|-- *_test.go                           Unit tests and benchmarks
|-- *_examples_test.go                  Executable examples
|-- go.mod                              Module definition with no external requirements
|-- LICENSE                             Apache License 2.0
|-- README.md                           Project overview and usage guide
`-- SECURITY.md                         Vulnerability reporting policy
```

## 🔐 Security

Security scanning is handled by GitHub CodeQL. See [SECURITY.md](SECURITY.md) for
the vulnerability reporting policy.

## 📄 License

`comparator` is licensed under the [Apache License 2.0](LICENSE).

## 🤝 Contributing

Issues and pull requests are welcome at
[github.com/slashdevops/comparator](https://github.com/slashdevops/comparator).
Please keep changes small, idiomatic, tested, documented, and dependency-free
unless there is a clear reason to expand the project scope.

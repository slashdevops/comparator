# Comparator Package

[![Go Reference](https://pkg.go.dev/badge/github.com/aizon-shared/ds-utils/pkg/comparator.svg)](https://pkg.go.dev/github.com/aizon-shared/ds-utils/pkg/comparator)

The `comparator` package provides advanced comparison and diffing capabilities for Go data structures. It offers a comprehensive solution for deep equality checking and difference detection across any Go values, including primitives, structs, slices, maps, pointers, and complex nested structures.

## Features

### Core Capabilities

- **Deep Recursive Comparison** - Handles complex nested structures with cycle detection
- **Difference Detection** - Detailed reports of what differs between two values
- **Multiple Output Formats** - Text, JSON, Markdown, HTML, and Unified Diff
- **JSON Patch Generation** - RFC 6902 compliant patch documents
- **Visual Tree Diff** - Hierarchical representation of differences
- **Custom Comparators** - Register custom comparison logic for specific types
- **Statistics & Suggestions** - Detailed metrics and actionable recommendations

### Configuration Options

- **Float Precision** - Configurable tolerance for floating-point comparisons
- **Slice Order** - Option to ignore element ordering (treat as sets)
- **Field Ignoring** - Skip specific struct fields by name
- **Depth Limiting** - Prevent stack overflow with deeply nested structures
- **Unexported Fields** - Option to include/exclude private fields
- **Empty Values** - Treat nil and empty as equal
- **NaN Handling** - Option to consider NaN values as equal
- **Colorized Output** - ANSI color-coded terminal output for better readability
- **Include Equal Values** - Show both differences and matches in reports

## Installation

```bash
go get github.com/aizon-shared/ds-utils/pkg/comparator
```

## Update

```bash
go get -u github.com/aizon-shared/ds-utils/pkg/comparator
```

## Quick Start

### Simple Equality Check

```go
import "github.com/aizon-shared/ds-utils/pkg/comparator"

comp := comparator.New()
if comp.Equal(obj1, obj2) {
    fmt.Println("Objects are equal")
}

// Or use the convenience function
if comparator.Equal(obj1, obj2) {
    fmt.Println("Objects are equal")
}
```

### Comparison with Options

```go
comp := comparator.NewWithOptions(
    comparator.WithFloatPrecision(1e-6),
    comparator.IgnoreSliceOrder(),
    comparator.IgnoreStructFields("ID", "CreatedAt", "UpdatedAt"),
)

if comp.Equal(user1, user2) {
    fmt.Println("Users are equal (ignoring ID and timestamps)")
}
```

### Detailed Difference Analysis

```go
diffComp := comparator.NewDiffComparer(
    comparator.WithOutputFormat("markdown"),
    comparator.WithMaxDiffs(100),
)

result := diffComp.CompareWithDiff(expected, actual)
if !result.Equal {
    fmt.Printf("Found %d differences\n", len(result.Differences))
    fmt.Println(result.Summary)

    for _, diff := range result.Differences {
        fmt.Printf("[%s] %s: %s\n", diff.Severity, diff.Path, diff.Message)
        for _, suggestion := range diff.Suggestions {
            fmt.Printf("  → %s\n", suggestion)
        }
    }
}
```

## Advanced Features

### Colorized Terminal Output

Enable ANSI color-coded output for better readability in terminals:

```go
comp := comparator.NewDiffComparer(
    comparator.WithColorize(true),
    comparator.WithOutputFormat("text"),
)

result := comp.CompareWithDiff(config1, config2)
formatted, _ := comp.FormatDiff(result, "text")
fmt.Println(formatted) // Displays with colors

// Colors used:
// - Cyan: Paths and section headers
// - Green: Expected values and equal fields
// - Red: Actual values and errors
// - Yellow: Warnings
// - Bold: Section headers
```

**Use Cases:**

- Terminal-based tools and CLI applications
- Development and debugging sessions
- Interactive diff viewers
- CI/CD pipeline outputs

### Include Equal Values

Show both differences and matches in comparison reports:

```go
comp := comparator.NewDiffComparer(
    comparator.WithIncludeEqual(true),
)

result := comp.CompareWithDiff(user1, user2)

// result.Differences now contains both:
// - Differences (severity: "error")
// - Equal values (severity: "info", type: "equal")

for _, diff := range result.Differences {
    if diff.Detail.Type == "equal" {
        fmt.Printf("✓ %s: values match\n", diff.Path)
    } else {
        fmt.Printf("✗ %s: %s\n", diff.Path, diff.Message)
    }
}
```

**Use Cases:**

- Comprehensive audit trails
- Configuration validation reports
- Debugging comparison logic
- Understanding what hasn't changed between versions

### Combined Features

```go
comp := comparator.NewDiffComparer(
    comparator.WithColorize(true),
    comparator.WithIncludeEqual(true),
    comparator.WithOutputFormat("text"),
)

result := comp.CompareWithDiff(oldConfig, newConfig)
formatted, _ := comp.FormatDiff(result, "text")

// Output shows:
// - Equal fields in green with "info" severity
// - Different fields in red with "error" severity
// - All paths highlighted in cyan
fmt.Println(formatted)
```

### JSON Patch Generation

Generate RFC 6902 JSON Patch documents:

```go
patch, err := comparator.GetJSONPatch(oldDoc, newDoc)
if err != nil {
    log.Fatal(err)
}

patchJSON, _ := json.MarshalIndent(patch, "", "  ")
fmt.Println(string(patchJSON))
// Output:
// [
//   {"op": "replace", "path": "/name", "value": "John"},
//   {"op": "add", "path": "/email", "value": "john@example.com"}
// ]
```

Supported operations:

- `add` - Add a new value at a path
- `remove` - Remove the value at a path
- `replace` - Replace the value at a path
- `move` - Move a value from one path to another
- `copy` - Copy a value from one path to another
- `test` - Test that a value at a path equals a specified value

### Unified Diff Format

Generate Unix diff-style output:

```go
diffComp := comparator.NewDiffComparer()
unifiedDiff, err := diffComp.GetUnifiedDiff(oldConfig, newConfig)
if err != nil {
    log.Fatal(err)
}

fmt.Println(unifiedDiff.Header)
for _, chunk := range unifiedDiff.Chunks {
    fmt.Println(chunk.Context)
    for _, change := range chunk.Changes {
        symbol := " "
        if change.Type == "add" {
            symbol = "+"
        } else if change.Type == "remove" {
            symbol = "-"
        }
        fmt.Printf("%s%s\n", symbol, change.Content)
    }
}
```

### Visual Tree Diff

Generate hierarchical tree representation:

```go
comp := comparator.NewDiffComparer()
visualDiff, err := comp.GetVisualDiff(obj1, obj2)
if err != nil {
    log.Fatal(err)
}

printNode(visualDiff.Root, 0)

func printNode(node *comparator.VisualNode, indent int) {
    prefix := strings.Repeat("  ", indent)
    symbol := ""
    switch node.Status {
    case "added": symbol = "+ "
    case "removed": symbol = "- "
    case "different": symbol = "~ "
    default: symbol = "  "
    }
    fmt.Printf("%s%s%s: %s\n", prefix, symbol, node.Path, node.Value)
    for _, child := range node.Children {
        printNode(child, indent+1)
    }
}
```

### Custom Comparators

Register custom comparison logic for specific types:

```go
type User struct {
    ID   int
    Name string
    Age  int
}

comp := comparator.NewWithOptions(
    comparator.WithCustomComparator(func(a, b User) bool {
        return a.ID == b.ID // Compare users by ID only
    }),
)

if comp.Equal(user1, user2) {
    fmt.Println("Users have the same ID")
}
```

## Configuration Options Reference

### Comparison Behavior

- `WithFloatPrecision(float64)`: Set precision threshold for float comparisons. Default: `1e-9`.
- `IgnoreSliceOrder()`: Treat slices as sets and ignore element order. Default: `false`.
- `WithMaxDepth(int)`: Limit recursion depth where `0` means unlimited. Default: `0`.
- `IgnoreUnexported()`: Skip unexported struct fields. Default: `false`.
- `EquateEmpty()`: Treat nil and empty values as equal. Default: `true`.
- `EquateNaNs()`: Treat all NaN values as equal. Default: `false`.
- `IgnoreStructFields(...string)`: Skip specific struct fields by name. Default: none.

### Diff Output

- `WithDiffMode(DiffMode)`: Set diff reporting mode. Default: `DiffModeSimple`.
- `WithMaxDiffs(int)`: Limit the number of differences collected. Default: `1000`.
- `WithOutputFormat(string)`: Set output format to `text`, `json`, `markdown`, or `html`. Default: `"text"`.
- `WithColorize(bool)`: Enable ANSI color codes in text output. Default: `false`.
- `WithIncludeEqual(bool)`: Include equal values in diff reports. Default: `false`.

### Other

- `WithTimeLayout(string)`: Set the time format used for display. Default: `time.RFC3339Nano`.
- `WithCustomComparator[T](func(T, T) bool)`: Register a custom comparator for type `T`. Default: none.

## Diff Modes

- **`DiffModeSimple`** - Basic difference reporting
- **`DiffModeFull`** - Comprehensive details with nested differences
- **`DiffModeUnified`** - Unix diff-style output
- **`DiffModeJSONPatch`** - RFC 6902 JSON Patch operations
- **`DiffModeVisual`** - Tree-based visual representation

## Output Formats

- **`text`** - Plain text with optional ANSI colors
- **`json`** - JSON representation of diff result
- **`markdown`** - Markdown-formatted report
- **`html`** - HTML with styling

## Performance Notes

The package already includes benchmarks for primitive values, structs, deep nesting, large slices, large maps, and diff generation.

- plain equality checks are cheaper than full diff generation
- large slices are among the more expensive cases, especially when additional diff detail is required
- reusing comparator instances is preferable when the same configuration is used repeatedly
- options such as `IgnoreSliceOrder`, `WithIncludeEqual`, and rich output formatting can materially increase work and output size

You can rerun the package benchmarks with:

```bash
go test ./pkg/comparator -run '^$' -bench .
```

## Supported Types

The comparator handles all Go types:

- **Primitives**: `bool`, `int*`, `uint*`, `float*`, `complex*`, `string`
- **Composite**: arrays, slices, maps, structs
- **Reference**: pointers, interfaces, channels, functions
- **Special**: `time.Time` (uses native Equal method)

## Performance Considerations

1. **Reuse Comparator Instances** - For repeated comparisons with the same configuration
2. **Limit Depth** - Use `WithMaxDepth` for deeply nested structures
3. **Limit Diffs** - Use `WithMaxDiffs` to control memory usage
4. **Skip Expensive Fields** - Use `IgnoreStructFields` for expensive comparisons
5. **Avoid IgnoreSliceOrder** - For large slices when possible (use pre-sorted slices)

## Thread Safety

Comparator instances are **not thread-safe**. Create separate instances for concurrent use, or synchronize access with a mutex.

## Error Handling

Most methods return errors for future compatibility. Current implementations rarely return errors, but always check error returns for forward compatibility.

## Examples

See [comparator_examples_test.go](comparator_examples_test.go) for comprehensive examples including:

- String list comparisons
- JSON structure diffing
- Configuration validation
- Colorized output
- Including equal values
- Performance testing with large structures

## License

This package is part of the ds-utils project.

## Contributing

Contributions are welcome! Please ensure all tests pass and add appropriate test coverage for new features.

```bash
make test
```

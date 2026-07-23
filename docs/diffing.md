# 🔍 Diffing

Equality tells you *whether* two values differ. Diffing tells you *how*. Use a
`DiffComparer` (from `NewDiffComparer`) or the package-level `CompareWithDiff`.

```go
result := comparator.CompareWithDiff(expected, actual)
if !result.Equal {
    fmt.Println(result.Summary)
    for _, diff := range result.Differences {
        fmt.Printf("[%s] %s: %s\n", diff.Severity, diff.Path, diff.Message)
    }
}
```

## 🧾 The `DiffResult`

```go
type DiffResult struct {
    Summary     string        // human-readable summary
    Differences []Difference  // all differences (empty when Equal)
    PathStats   PathStats     // traversal statistics
    Equal       bool          // deep-equality verdict
}
```

### `PathStats`

Useful for understanding scope and cost of a comparison:

| Field | Meaning |
| ----- | ------- |
| `TotalNodes` | Nodes encountered during traversal. |
| `ComparedNodes` | Nodes actually compared. |
| `DifferentNodes` | Nodes found different. |
| `IgnoredNodes` | Nodes skipped (e.g. via `IgnoreStructFields`). |

## 🧩 The `Difference`

```go
type Difference struct {
    Path        string            // dot-notation location, e.g. "User.Address.City"
    Message     string            // human-readable description
    Severity    string            // "error", "warning", or "info"
    Detail      DifferenceDetail  // granular, type-specific detail
    Suggestions []string          // optional remediation hints
    Level       int               // depth in the structure (0 = root)
}
```

### `DifferenceDetail.Type`

The `Detail.Type` field classifies what kind of difference was found:

| Type | Meaning |
| ---- | ------- |
| `value_different` | Same type, different value. |
| `type_mismatch` | The two values have different types. |
| `missing` | Present on one side only. |
| `extra_element` | Extra slice/array element. |
| `length_mismatch` | Slices/arrays of different length. |
| `missing_key` / `extra_key` | Map key present on one side only. |
| `nil_mismatch` | One side is nil, the other is not. |
| `equal` | Emitted only with `WithIncludeEqual(true)`. |

`Detail` also carries `ExpectedType`, `ActualType`, `ExpectedValue`,
`ActualValue`, and any nested `Children`.

> 🧭 **Convention:** the *first* argument is treated as **expected** and the
> *second* as **actual**. Keep the order consistent
> (`CompareWithDiff(expected, actual)`) so the report reads correctly.

## ➕ Including Equal Values

By default only differences are reported. Enable `WithIncludeEqual(true)` to also
emit matched values (with `Severity: "info"` and `Detail.Type: "equal"`) — handy
for audit trails and validation reports.

```go
comp := comparator.NewDiffComparer(comparator.WithIncludeEqual(true))
result := comp.CompareWithDiff(user1, user2)

for _, d := range result.Differences {
    if d.Detail.Type == "equal" {
        fmt.Printf("✓ %s\n", d.Path)
    } else {
        fmt.Printf("✗ %s: %s\n", d.Path, d.Message)
    }
}
```

## 🎚️ Bounding Work

- `WithMaxDiffs(n)` caps how many differences are collected (default `1000`).
- `WithMaxDepth(n)` caps recursion depth (default `0`, unlimited).

Both are important safety valves for very large or deeply nested inputs — see
[Performance](performance.md).

## ➡️ Next Steps

- Render results in different formats: [Output Formats](output-formats.md).
- Produce machine-applicable patches: [JSON Patch](json-patch.md).

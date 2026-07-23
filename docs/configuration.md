# ⚙️ Configuration & Options

All behavior is configured through **functional options** of type
`comparator.Option`. The same options work with every constructor and every
package-level helper.

```go
comp := comparator.NewWithOptions(
    comparator.WithFloatPrecision(1e-6),
    comparator.IgnoreSliceOrder(),
    comparator.IgnoreStructFields("ID", "CreatedAt", "UpdatedAt"),
)
```

## 🏗️ Constructors

| Constructor | Returns | Notes |
| ----------- | ------- | ----- |
| `New()` | `Comparator` | Default configuration. |
| `NewWithOptions(opts ...Option)` | `Comparator` | Equality with custom options. |
| `NewDiffComparer(opts ...Option)` | `DiffComparer` | Full diffing + formatting. |

> ♻️ **Reuse instances.** Constructing a comparator is cheap, but reusing one
> across many comparisons with the same configuration avoids repeated setup.

## 🔧 Comparison Behavior

| Option | Default | Description |
| ------ | ------- | ----------- |
| `WithFloatPrecision(float64)` | `1e-9` | Two floats are equal when their absolute difference is `<=` this value. |
| `IgnoreSliceOrder()` | `false` | Treat slices/arrays as sets; element order is ignored. |
| `WithMaxDepth(int)` | `0` (unlimited) | Stop recursing beyond this depth; guards against very deep structures. |
| `IgnoreUnexported()` | `false` | Skip unexported struct fields entirely. |
| `EquateEmpty()` | `true` | Treat `nil` and empty containers (slice/map) as equal. |
| `EquateNaNs()` | `false` | Treat `NaN == NaN` as `true` (IEEE-754 says `false`). |
| `IgnoreStructFields(...string)` | none | Skip struct fields by name, at any nesting level. |
| `WithTimeLayout(string)` | `time.RFC3339Nano` | Layout used to render `time.Time` in diff output. Equality still uses `time.Time.Equal`. |

> ℹ️ `EquateEmpty` defaults to **true** in this package. Pass it explicitly if
> you want to document the intent; there is currently no dedicated option to turn
> it back off, so rely on strict typing when nil-vs-empty must differ.

## 🧾 Diff Output

| Option | Default | Description |
| ------ | ------- | ----------- |
| `WithDiffMode(DiffMode)` | `DiffModeSimple` | Level of detail in a diff report (see below). |
| `WithMaxDiffs(int)` | `1000` | Maximum number of differences collected before stopping. |
| `WithOutputFormat(string)` | `"text"` | `"text"`, `"json"`, `"markdown"`, or `"html"`. |
| `WithColorize(bool)` | `false` | ANSI color codes in `text` output. |
| `WithIncludeEqual(bool)` | `false` | Also report values that are equal (severity `info`). |

### Diff modes

| Mode | Meaning |
| ---- | ------- |
| `DiffModeSimple` | Basic difference reporting (default). |
| `DiffModeFull` | Comprehensive detail with nested differences and statistics. |
| `DiffModeUnified` | Unix `diff`-style output. |
| `DiffModeJSONPatch` | RFC 6902 JSON Patch operations. |
| `DiffModeVisual` | Tree-based visual representation. |

## 🧪 Custom Logic

| Option | Description |
| ------ | ----------- |
| `WithCustomComparator[T](func(a, b T) bool)` | Register domain-specific equality for a concrete type `T`. See [Custom Comparators](custom-comparators.md). |

## 🎛️ Putting It Together

```go
comp := comparator.NewDiffComparer(
    comparator.WithOutputFormat("markdown"),
    comparator.WithMaxDiffs(100),
    comparator.WithIncludeEqual(true),
    comparator.IgnoreStructFields("ID", "UpdatedAt"),
    comparator.WithFloatPrecision(1e-6),
)

result := comp.CompareWithDiff(expected, actual)
report, _ := comp.FormatDiff(result, "markdown")
fmt.Println(report)
```

Continue with [Diffing](diffing.md) to learn how to read a `DiffResult`.

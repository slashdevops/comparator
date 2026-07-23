# ⚡ Performance

`comparator` uses reflection to walk arbitrary values. That flexibility has a
cost, but the cost is predictable and controllable. This guide explains the cost
model and how to tune it.

## 🧮 Cost Model

Work scales with the **number of nodes** traversed:

- **Plain equality** (`Equal`) is the cheapest path — it stops at the first
  difference and collects no report.
- **Diff generation** (`CompareWithDiff`, `GetUnifiedDiff`, `GetJSONPatch`,
  `GetVisualDiff`) is more expensive because it visits nodes to build a report.
- **Large slices and maps** dominate cost, especially with diff detail enabled.
- **`IgnoreSliceOrder`** turns slice comparison into set-style matching, which is
  materially more expensive than positional comparison — avoid it on large,
  already-ordered slices.
- **Rich output** (`WithIncludeEqual`, colorized/markdown/html formatting)
  increases both work and output size.

## 🎛️ Tuning Levers

| Lever | Effect |
| ----- | ------ |
| Reuse a comparator instance | Avoids repeated setup for same-config comparisons. |
| `WithMaxDepth(n)` | Bounds recursion for deeply nested structures. |
| `WithMaxDiffs(n)` | Caps memory when many differences exist. |
| `IgnoreStructFields(...)` | Skips expensive or irrelevant fields entirely. |
| Prefer `Equal` over diffing | Use the cheap path when you only need a verdict. |
| Pre-sort instead of `IgnoreSliceOrder` | Positional comparison is much cheaper. |

## 🏁 Benchmarks

The test suite includes benchmarks for primitives, structs, deep nesting, large
slices, large maps, and diff generation. Run them with:

```bash
go test -run '^$' -bench . ./...
```

Add memory allocation stats:

```bash
go test -run '^$' -bench . -benchmem ./...
```

Benchmarks use the Go 1.24+ `for b.Loop()` form, so timing excludes per-benchmark
setup automatically.

## 🧩 Practical Guidance

1. **Reuse comparators** for repeated comparisons with the same configuration.
2. **Limit depth** with `WithMaxDepth` for untrusted or unbounded input.
3. **Limit diffs** with `WithMaxDiffs` to keep reports (and memory) bounded.
4. **Skip expensive fields** with `IgnoreStructFields`.
5. **Avoid `IgnoreSliceOrder`** on large slices; sort once and compare positionally.

## ➡️ Next Steps

- Common questions and gotchas: [FAQ](faq.md).

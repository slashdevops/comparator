# 🧩 Extensibility

The package is designed to be extended without forking. This guide covers the
type-safe generic API, streaming reporters, context-aware comparison, and
pluggable output formatters.

## 🔤 Type-safe generic API

The generic helpers give compile-time type checking at the call site and
delegate to the same engine:

```go
comparator.EqualT(user1, user2, comparator.IgnoreStructFields("ID")) // bool
comparator.DeepEqualT(a, b)                                          // bool
comparator.DiffT(a, b)                                              // ([]Difference, error)
comparator.CompareWithDiffT(a, b)                                    // *DiffResult
comparator.GetJSONPatchT(a, b)                                       // ([]JSONPatchOperation, error)
```

Because `EqualT[T]` requires both arguments to share type `T`, mixing types is a
compile error rather than a silent `false`.

## 📡 Streaming with a reporter

`WithReporter` registers a callback invoked for every difference as it is
discovered, in traversal order:

```go
comp := comparator.NewDiffComparer(comparator.WithReporter(func(d comparator.Difference) {
    log.Printf("%s: %s", d.Path, d.Message)
}))
comp.CompareWithDiff(a, b)
```

Use it for logging, progress reporting, or feeding another system without
waiting for the full `DiffResult`.

## ⏱️ Context-aware comparison

For large or untrusted inputs, use the context-aware entry points so a
comparison can be canceled or bounded by a deadline:

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

equal, err := comparator.EqualCtx(ctx, a, b)
if errors.Is(err, comparator.ErrCanceled) {
    // canceled or deadline exceeded; err also matches context.Canceled /
    // context.DeadlineExceeded
}
```

`DiffCtx` and `CompareWithDiffCtx` are the diffing counterparts. The comparators
returned by `New`, `NewWithOptions`, and `NewDiffComparer` also satisfy the
`ContextComparator` interface if you prefer method calls.

## 🎨 Custom output formatters

Register a `Formatter` to teach `FormatDiff` a new output format — CSV, JUnit,
SARIF, or anything else — without changing the package:

```go
comparator.RegisterFormatter("summary", func(r *comparator.DiffResult) (string, error) {
    if r.Equal {
        return "no differences", nil
    }
    return fmt.Sprintf("%d differences", len(r.Differences)), nil
})

comp := comparator.NewDiffComparer()
out, _ := comp.FormatDiff(comp.CompareWithDiff(a, b), "summary")
```

Notes:

- Names are case-insensitive and cannot override the built-in formats
  (`text`, `json`, `markdown`, `html`).
- `RegisterFormatter` and `UnregisterFormatter` are safe for concurrent use.
- Requesting an unknown format returns an error wrapping `ErrUnknownFormat`.

## ➡️ Next Steps

- Use the comparator in tests: [Testing](testing.md).
- Review every option: [Configuration & Options](configuration.md).

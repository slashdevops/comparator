# ❓ FAQ

## Is it thread-safe?

**No.** Comparator instances are **not** safe for concurrent use. Create a
separate instance per goroutine, or protect a shared instance with a mutex. The
package-level helpers (`Equal`, `CompareWithDiff`, …) build a fresh instance on
each call, so they are safe to call concurrently.

## Does it have any dependencies?

No. `comparator` depends only on the Go standard library. There is no `go.sum`
because there are no external modules.

## What Go version do I need?

Go **1.26** or newer.

## How does it compare to `reflect.DeepEqual`?

`reflect.DeepEqual` gives you a single boolean with no configuration and no
report. `comparator` adds configurable behavior (float precision, slice-order
insensitivity, field ignoring, custom comparators, NaN handling) and produces
detailed, presentable diffs. Use `reflect.DeepEqual` for the simplest checks and
`comparator` when you need control or a report. `DeepEqual` in this package is a
familiar-named alias for `Equal`.

## How are `time.Time` values compared?

With the native `time.Time.Equal` method, so two instants that represent the same
moment in different locations compare as equal. `WithTimeLayout` only affects how
times are *rendered* in diff output.

## Why do JSON Patch paths use Go field names, not JSON tags?

By default paths are derived from Go field names (e.g. field `Name` → `/Name`).
Pass `WithFieldNaming(JSONTagNaming)` to use `json` tag names instead, which makes
the pointers line up with JSON documents and with `ApplyJSONPatch`. See
[JSON Patch](json-patch.md).

## Are `nil` and empty slices/maps equal?

By default, **yes** — nil/empty equivalence is on. Pass `WithEquateEmpty(false)`
to make `nil` and empty containers compare as different.

## Is `NaN` equal to `NaN`?

By default, **no**, matching IEEE-754. Enable `EquateNaNs()` to treat all NaN
values as equal.

## Do the comparison methods return errors?

Most methods return an `error` for forward compatibility, though current
implementations rarely produce one. Always check the error for future-proofing.

## How do I limit work on huge inputs?

Use `WithMaxDepth` to bound recursion and `WithMaxDiffs` to bound the number of
collected differences. See [Performance](performance.md).

## Where are the runnable examples?

- [`comparator_examples_test.go`](../comparator_examples_test.go)
- [`comparator_api_examples_test.go`](../comparator_api_examples_test.go)
- [`comparator_options_examples_test.go`](../comparator_options_examples_test.go)

Run them with:

```bash
go test -run Example ./...
```

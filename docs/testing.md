# 🧪 Testing Helpers

The package ships assertion helpers that fail a test and print a readable diff,
making `comparator` a drop-in equality checker for your test suite.

## ✅ AssertEqual / AssertNotEqual

```go
func TestUser(t *testing.T) {
    got := buildUser()
    want := User{Name: "Alice", Role: "admin"}

    comparator.AssertEqual(t, want, got, comparator.IgnoreStructFields("ID", "CreatedAt"))
}
```

On failure the test is marked failed (via `t.Errorf`) with a message like:

```text
values are not equal:
Found 1 difference(s)
  .Role: user != admin
```

- `AssertEqual(t, want, got, opts...)` — fail unless the values are equal.
  Returns `true` when equal.
- `AssertNotEqual(t, want, got, opts...)` — fail if the values are equal.
- `RequireEqual(t, want, got, opts...)` — like `AssertEqual` but stops the test
  immediately (uses `t.Fatalf` when available).

All accept the same [functional options](configuration.md) as the rest of the
package, so you can ignore fields, relax float precision, or ignore slice order
in assertions.

## 🧷 The `TestingT` interface

The helpers accept a small interface rather than `*testing.T` directly, so the
package does not import `testing` into production builds:

```go
type TestingT interface {
    Helper()
    Errorf(format string, args ...any)
}
```

`*testing.T` and `*testing.B` satisfy it. `RequireEqual` additionally uses
`Fatalf` when the supplied value implements it.

## ➡️ Next Steps

- Extend output and behavior: [Extensibility](extensibility.md).
- Common questions: [FAQ](faq.md).

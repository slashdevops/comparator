# 🧠 Custom Comparators

Sometimes structural equality is not what you want. Two records may be
"the same" when they share an identifier, even if other fields differ. Register
a **custom comparator** to encode that rule for a specific type.

## 🔧 Registering

`WithCustomComparator[T]` is a generic option. Provide a function that decides
equality for values of type `T`:

```go
type Account struct {
    ID   int
    Name string
}

comp := comparator.NewWithOptions(
    comparator.WithCustomComparator(func(a, b Account) bool {
        return a.ID == b.ID // equal when IDs match
    }),
)

a1 := Account{ID: 7, Name: "prod"}
a2 := Account{ID: 7, Name: "production"}

fmt.Println(comp.Equal(a1, a2)) // true
```

## 🎯 How It Works

- The comparator is keyed by the concrete type `T`. Whenever two values of that
  exact type are compared — at the root or nested inside a larger structure —
  your function is used instead of the default structural comparison.
- Register comparators for as many distinct types as you need by passing multiple
  `WithCustomComparator` options.

## 💡 Common Use Cases

| Scenario | Rule |
| -------- | ---- |
| Entities with surrogate keys | Compare by `ID` only. |
| Values with derived/cached fields | Ignore the cache; compare source fields. |
| Domain-specific equivalence | e.g. case-insensitive names, normalized URLs. |
| Third-party types | Compare by the fields that actually matter. |

## ⚖️ Custom Comparators vs. `IgnoreStructFields`

Both let you loosen equality, but they differ in intent:

- Use **`IgnoreStructFields`** when a few fields (IDs, timestamps) should simply
  be skipped everywhere.
- Use **`WithCustomComparator`** when equality for a type is a genuine domain
  decision that you want to express as code.

## 🧵 A Note on Diffs

Custom comparators answer *equal or not*. When you need a detailed diff and a
custom comparator reports two values as equal, they will not appear as a
difference. Combine custom comparators with `WithIncludeEqual(true)` if you want
those matches surfaced in reports.

## ➡️ Next Steps

- Review all options in [Configuration & Options](configuration.md).
- Tune large comparisons in [Performance](performance.md).

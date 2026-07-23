# 🩹 JSON Patch (RFC 6902)

`comparator` can express the difference between two values as an
[RFC 6902](https://datatracker.ietf.org/doc/html/rfc6902) JSON Patch — a list of
operations that transforms the first value into the second. This is ideal for
sending partial updates over the wire or persisting change sets.

## 🚀 Quick Start

```go
patch, err := comparator.GetJSONPatch(oldDoc, newDoc)
if err != nil {
    log.Fatal(err)
}

out, _ := json.MarshalIndent(patch, "", "  ")
fmt.Println(string(out))
```

Example output:

```json
[
  { "op": "replace", "path": "/name", "value": "John" },
  { "op": "add", "path": "/email", "value": "john@example.com" }
]
```

## 🧩 The `JSONPatchOperation`

```go
type JSONPatchOperation struct {
    Op    string // "add", "remove", "replace", "move", "copy", "test"
    Path  string // JSON Pointer (RFC 6901) to the target
    Value any    // value for add/replace/test (omitted otherwise)
    From  string // source path for move/copy
}
```

| Operation | Purpose |
| --------- | ------- |
| `add` | Add a value at a path. |
| `remove` | Remove the value at a path. |
| `replace` | Replace the value at a path. |
| `move` | Move a value from `From` to `Path`. |
| `copy` | Copy a value from `From` to `Path`. |
| `test` | Assert the value at a path. |

## 🧭 Path Format

Paths are [JSON Pointers](https://datatracker.ietf.org/doc/html/rfc6901). By
default, struct fields appear by their **Go field name** (e.g. a field `Name`
produces `/Name`), and nested paths are slash-separated (e.g. `/Server/Port`).

To make pointers line up with the JSON representation of your values, generate
the patch with `WithFieldNaming(JSONTagNaming)`:

```go
type Profile struct {
    FullName string `json:"full_name"`
}

patch, _ := comparator.GetJSONPatch(
    Profile{FullName: "Jon"},
    Profile{FullName: "John"},
    comparator.WithFieldNaming(comparator.JSONTagNaming),
)
// patch[0].Path == "/full_name"
```

## 🛠️ Applying a Patch

`ApplyJSONPatch` applies a patch and returns the modified document. It supports
`add`, `remove`, `replace`, `move`, `copy`, and `test`:

```go
doc := map[string]any{"name": "Jon"}
patch := []comparator.JSONPatchOperation{
    {Op: "replace", Path: "/name", Value: "John"},
    {Op: "add", Path: "/email", Value: "john@example.com"},
}

updated, err := comparator.ApplyJSONPatch(doc, patch)
if err != nil {
    // errors wrap comparator.ErrInvalidPatch (including a failed "test" op)
    log.Fatal(err)
}
```

The document is normalized to JSON values (objects → `map[string]any`, arrays →
`[]any`) by round-tripping through `encoding/json`, so it must be
JSON-serializable. For a clean generate-then-apply round trip, generate the patch
with `JSONTagNaming` so its pointers match the object keys.

A failed `test` operation, an unreachable path, or an unsupported op returns an
error wrapping `ErrInvalidPatch`.

## ➡️ Next Steps

- Explore other renderings in [Output Formats](output-formats.md).
- Register domain equality in [Custom Comparators](custom-comparators.md).

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

Paths are [JSON Pointers](https://datatracker.ietf.org/doc/html/rfc6901). Struct
fields appear by their **Go field name** (e.g. a field `Name` produces `/Name`),
and nested paths are slash-separated (e.g. `/Server/Port`).

> 🔎 If you need the JSON-tag name instead of the Go field name in the pointer,
> transform the paths after generation, or marshal your inputs to
> `map[string]any` (via `encoding/json`) before comparing.

## 🛠️ Applying a Patch

This package **generates** patches; it does not apply them. To apply a patch,
marshal it to JSON and use any RFC 6902-compliant applier, or interpret the
operations yourself. Because the type has standard JSON tags, it round-trips
cleanly through `encoding/json`.

## ➡️ Next Steps

- Explore other renderings in [Output Formats](output-formats.md).
- Register domain equality in [Custom Comparators](custom-comparators.md).

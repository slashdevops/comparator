# 🎨 Output Formats

A `DiffResult` can be rendered in several formats, and the package also offers
two specialized diff structures (unified and visual). Pick the one that matches
your surface — terminal, web UI, API, or version control.

## 🖨️ `FormatDiff`

`FormatDiff` turns a `*DiffResult` into a string in one of four formats:

```go
comp := comparator.NewDiffComparer(comparator.WithOutputFormat("markdown"))
result := comp.CompareWithDiff(expected, actual)

report, err := comp.FormatDiff(result, "markdown")
if err != nil {
    log.Fatal(err)
}
fmt.Println(report)
```

| Format | Best for |
| ------ | -------- |
| `"text"` | Terminals, logs, CLI tools (optionally colorized). |
| `"json"` | APIs, storage, further programmatic processing. |
| `"markdown"` | PRs, issues, documentation, chat. |
| `"html"` | Web dashboards and diff viewers. |

You can set a default format with `WithOutputFormat(...)` and still override it
per call by passing the format argument to `FormatDiff`.

## 🌈 Colorized Text

Enable ANSI colors for terminal output:

```go
comp := comparator.NewDiffComparer(
    comparator.WithColorize(true),
    comparator.WithOutputFormat("text"),
)
result := comp.CompareWithDiff(config1, config2)
out, _ := comp.FormatDiff(result, "text")
fmt.Println(out)
```

Color legend:

- 🟦 **Cyan** — paths and section headers
- 🟩 **Green** — expected values and equal fields
- 🟥 **Red** — actual values and errors
- 🟨 **Yellow** — warnings

> 💡 Disable color when writing to a file or a non-TTY (e.g. CI logs) to avoid
> escape sequences in stored output.

## 📐 Unified Diff

`GetUnifiedDiff` returns a Unix `diff`-style structure:

```go
comp := comparator.NewDiffComparer()
ud, err := comp.GetUnifiedDiff(oldConfig, newConfig)
if err != nil {
    log.Fatal(err)
}

fmt.Println(ud.Header)
for _, chunk := range ud.Chunks {
    fmt.Println(chunk.Context)
    for _, ch := range chunk.Changes {
        symbol := " "
        switch ch.Type {
        case "add":
            symbol = "+"
        case "remove":
            symbol = "-"
        }
        fmt.Printf("%s%s\n", symbol, ch.Content)
    }
}
```

Shape:

- `UnifiedDiff{ Header string, Chunks []Chunk }`
- `Chunk{ Context string, Changes []Change }`
- `Change{ Type string /* add|remove|context */, Content string, Line int }`

## 🌳 Visual Tree

`GetVisualDiff` returns a tree suitable for building graphical viewers:

```go
comp := comparator.NewDiffComparer()
vd, err := comp.GetVisualDiff(obj1, obj2)
if err != nil {
    log.Fatal(err)
}
printNode(vd.Root, 0)

func printNode(n *comparator.VisualNode, indent int) {
    prefix := strings.Repeat("  ", indent)
    symbol := "  "
    switch n.Status {
    case "added":
        symbol = "+ "
    case "removed":
        symbol = "- "
    case "different":
        symbol = "~ "
    }
    fmt.Printf("%s%s%s: %s\n", prefix, symbol, n.Path, n.Value)
    for _, child := range n.Children {
        printNode(child, indent+1)
    }
}
```

Each `VisualNode` has a `Path`, `Value`, `Status` (`same`, `different`,
`added`, `removed`), and `Children`.

## ➡️ Next Steps

- Generate machine-applicable patches: [JSON Patch](json-patch.md).

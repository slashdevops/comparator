package comparator

import "strings"

// String renders a unified diff in the familiar text form: the header followed
// by each hunk's context header and its change lines. Change contents already
// carry their "+", "-", or space prefix.
func (u *UnifiedDiff) String() string {
	if u == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(u.Header)
	for _, chunk := range u.Chunks {
		b.WriteString(chunk.Context)
		b.WriteByte('\n')
		for _, ch := range chunk.Changes {
			b.WriteString(ch.Content)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// String renders a visual diff tree as an indented, one-node-per-line string.
// Each node is prefixed by a status symbol: "+" added, "-" removed, "~"
// different, and a space for unchanged nodes.
func (v *VisualDiff) String() string {
	if v == nil || v.Root == nil {
		return ""
	}
	var b strings.Builder
	writeVisualNode(&b, v.Root, 0)
	return b.String()
}

func writeVisualNode(b *strings.Builder, node *VisualNode, depth int) {
	if node == nil {
		return
	}

	symbol := " "
	switch node.Status {
	case "added":
		symbol = "+"
	case "removed":
		symbol = "-"
	case "different":
		symbol = "~"
	}

	b.WriteString(strings.Repeat("  ", depth))
	b.WriteString(symbol)
	b.WriteByte(' ')
	b.WriteString(node.Path)
	if node.Value != "" {
		b.WriteString(": ")
		b.WriteString(node.Value)
	}
	b.WriteByte('\n')

	for _, child := range node.Children {
		writeVisualNode(b, child, depth+1)
	}
}

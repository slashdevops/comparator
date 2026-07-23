# 📚 Comparator Documentation

Welcome to the full documentation for the `comparator` package — a dependency-free
Go library for deep comparison and rich diffing of arbitrary Go values.

Everything here is built on the Go standard library only. If you are just getting
started, read the guides in order; otherwise jump straight to the topic you need.

## 🗂️ Table of Contents

| Guide | What it covers |
| ----- | -------------- |
| [Getting Started](getting-started.md) | Installation, your first comparison, and the core types. |
| [Configuration & Options](configuration.md) | Every functional option, its default, and when to use it. |
| [Diffing](diffing.md) | `DiffResult`, `Difference`, diff modes, and how to read a diff. |
| [Output Formats](output-formats.md) | Text (with color), JSON, Markdown, HTML, unified diff, and visual tree. |
| [JSON Patch](json-patch.md) | Generating and applying RFC 6902 patch documents. |
| [Custom Comparators](custom-comparators.md) | Custom equality, plus the `Equatable`/`Comparable` interfaces. |
| [Extensibility](extensibility.md) | Generics, streaming reporters, context, and pluggable formatters. |
| [Testing](testing.md) | `AssertEqual` and friends for use in test suites. |
| [Performance](performance.md) | Benchmarks, cost model, and tuning tips. |
| [FAQ](faq.md) | Common questions, gotchas, and thread-safety. |

## ⚡ Quick Links

- Package reference: [pkg.go.dev/github.com/slashdevops/comparator](https://pkg.go.dev/github.com/slashdevops/comparator)
- Runnable examples: [`comparator_examples_test.go`](../comparator_examples_test.go),
  [`comparator_api_examples_test.go`](../comparator_api_examples_test.go), and
  [`comparator_options_examples_test.go`](../comparator_options_examples_test.go)
- Source: [`comparator.go`](../comparator.go)

## 🧭 Mental Model

The package exposes two layers:

1. **Equality** — the [`Comparator`](getting-started.md#the-comparator-interface)
   interface answers a single question: *are these two values equal under this
   configuration?*
2. **Diffing** — the [`DiffComparer`](diffing.md) interface answers a richer
   question: *how exactly do these two values differ, and how do I present that?*

Both are configured with the same set of [functional options](configuration.md),
so you learn the options once and reuse them everywhere.

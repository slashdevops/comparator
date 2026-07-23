# Development Guidelines

This document contains the critical information about working with the project codebase.
Follows these guidelines precisely to ensure consistency and maintainability of the code.

## Stack

- Language: Go (Go 1.26)
- Framework: Go standard library
- Testing: Go's built-in testing package
- Dependency Management: Go modules
- Version Control: Git
- Documentation: go doc
- Code Review: Pull requests on GitHub
- CI/CD: GitHub Actions

## Project Structure

Since this is a library built in native Go, the files are organized following the
standard Go single-package layout in the repository root.

- Library files are located in the root directory.
- `*_examples_test.go` files contain executable Go examples rendered by pkg.go.dev.
- .github/ contains GitHub-specific files such as workflows for CI/CD.
- .gitignore specifies files and directories to be ignored by Git.
- .vscode/ contains Visual Studio Code configuration files.
- LICENSE is the license file for the project.
- README.md provides an overview of the project, installation instructions, usage examples, and other relevant information.
- go.mod declares the module and Go toolchain version.
- go.sum should not exist unless external dependencies are intentionally introduced.
- \*.go files contain the main source code of the library.
- \*\_test.go files contain the test cases for the library.

## Code Style

- Follow Go's idiomatic style defined in
  - <https://google.github.io/styleguide/go/guide>
  - <https://google.github.io/styleguide/go/decisions>
  - <https://google.github.io/styleguide/go/best-practices>
  - <https://golang.org/doc/effective_go.html>
- Use meaningful names for variables, functions, and packages.
- Keep functions small and focused on a single task.
- Use comments to explain complex logic or decisions.
- Prefer `for b.Loop()` over `for i := 0; i < b.N; i++` in benchmarks (Go 1.24+).
- don't use `interface{}` instead use `any` for better readability.

## Post-Change Checklist

Use standard Go commands after making changes:

```bash
go fix ./...
go fmt ./...
go vet ./...
go test -race -coverprofile=/tmp/comparator-coverage.txt -covermode=atomic ./...
go build ./...
```

Do not add external module dependencies without explicit approval; this project is intentionally standard-library-only.

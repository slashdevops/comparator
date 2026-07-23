// Package comparator provides advanced comparison and diffing capabilities for Go data structures.
//
// This package offers a comprehensive solution for deep equality checking and difference
// detection across any Go values, including primitives, structs, slices, maps, pointers,
// and complex nested structures. It supports cycle detection, custom comparators, and
// multiple output formats.
//
// # Key Features
//
//   - Deep recursive comparison with cycle detection
//   - Configurable comparison behavior (float precision, slice ordering, field ignoring)
//   - Multiple output formats: text, JSON, Markdown, HTML
//   - Unified diff format (Unix diff-style)
//   - JSON Patch generation (RFC 6902)
//   - Visual tree diff representation
//   - Custom comparators for specific types
//   - Detailed statistics and suggestions
//
// # Basic Usage
//
// Simple equality check:
//
//	comp := comparator.New()
//	if comp.Equal(obj1, obj2) {
//	    fmt.Println("Objects are equal")
//	}
//
// Or use the convenience function:
//
//	if comparator.Equal(obj1, obj2) {
//	    fmt.Println("Objects are equal")
//	}
//
// # Comparison with Options
//
// Configure comparison behavior using functional options:
//
//	comp := comparator.NewWithOptions(
//	    comparator.WithFloatPrecision(1e-6),
//	    comparator.IgnoreSliceOrder(),
//	    comparator.IgnoreStructFields("ID", "CreatedAt", "UpdatedAt"),
//	)
//
//	if comp.Equal(user1, user2) {
//	    fmt.Println("Users are equal (ignoring ID and timestamps)")
//	}
//
// # Detailed Difference Analysis
//
// Get comprehensive diff information:
//
//	diffComp := comparator.NewDiffComparer(
//	    comparator.WithOutputFormat("markdown"),
//	    comparator.WithMaxDiffs(100),
//	)
//
//	result := diffComp.CompareWithDiff(expected, actual)
//	if !result.Equal {
//	    fmt.Printf("Found %d differences\n", len(result.Differences))
//	    fmt.Println(result.Summary)
//
//	    for _, diff := range result.Differences {
//	        fmt.Printf("[%s] %s: %s\n", diff.Severity, diff.Path, diff.Message)
//	        for _, suggestion := range diff.Suggestions {
//	            fmt.Printf("  → %s\n", suggestion)
//	        }
//	    }
//	}
//
// # JSON Patch Generation
//
// Generate RFC 6902 JSON Patch documents:
//
//	patch, err := comparator.GetJSONPatch(oldDoc, newDoc)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	patchJSON, _ := json.MarshalIndent(patch, "", "  ")
//	fmt.Println(string(patchJSON))
//	// Output:
//	// [
//	//   {"op": "replace", "path": "/name", "value": "John"},
//	//   {"op": "add", "path": "/email", "value": "john@example.com"}
//	// ]
//
// # Unified Diff Format
//
// Generate Unix diff-style output:
//
//	diffComp := comparator.NewDiffComparer()
//	unifiedDiff, err := diffComp.GetUnifiedDiff(oldConfig, newConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println(unifiedDiff.Header)
//	for _, chunk := range unifiedDiff.Chunks {
//	    fmt.Println(chunk.Context)
//	    for _, change := range chunk.Changes {
//	        symbol := " "
//	        if change.Type == "add" {
//	            symbol = "+"
//	        } else if change.Type == "remove" {
//	            symbol = "-"
//	        }
//	        fmt.Printf("%s%s\n", symbol, change.Content)
//	    }
//	}
//
// # Custom Comparators
//
// Register custom comparison logic for specific types:
//
//	type User struct {
//	    ID   int
//	    Name string
//	    Age  int
//	}
//
//	comp := comparator.NewWithOptions(
//	    comparator.WithCustomComparator(func(a, b User) bool {
//	        return a.ID == b.ID // Compare users by ID only
//	    }),
//	)
//
//	if comp.Equal(user1, user2) {
//	    fmt.Println("Users have the same ID")
//	}
//
// # Configuration Options
//
// Available options for customizing comparison behavior:
//
//   - WithFloatPrecision(float64): Set precision for float comparisons
//   - IgnoreSliceOrder(): Treat slices as sets (ignore element order)
//   - WithMaxDepth(int): Limit recursion depth
//   - IgnoreUnexported(): Skip unexported struct fields
//   - EquateEmpty(): Treat nil and empty values as equal
//   - EquateNaNs(): Treat all NaN values as equal
//   - IgnoreStructFields(...string): Skip specific struct fields by name
//   - WithTimeLayout(string): Set time format for display
//   - WithCustomComparator[T](func(T, T) bool): Register custom comparator
//   - WithDiffMode(DiffMode): Set diff reporting mode
//   - WithMaxDiffs(int): Limit number of differences collected
//   - WithOutputFormat(string): Set output format (text/json/markdown/html)
//   - WithColorize(bool): Enable ANSI color codes in text output
//   - WithIncludeEqual(bool): Include equal values in diff reports
//
// # Colorized Output
//
// Enable ANSI color-coded terminal output for better readability:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithColorize(true),
//	    comparator.WithOutputFormat("text"),
//	)
//
//	result := comp.CompareWithDiff(config1, config2)
//	formatted, _ := comp.FormatDiff(result, "text")
//	fmt.Println(formatted) // Displays with colors in terminal
//
// The colorized output uses ANSI color codes:
//
//   - Cyan: Paths and section headers
//   - Green: Expected values and equal fields
//   - Red: Actual values and errors
//   - Yellow: Warnings
//   - Bold: Section headers
//
// Use cases for colorized output:
//
//   - Terminal-based tools and CLI applications
//   - Development and debugging sessions
//   - Interactive diff viewers
//   - CI/CD pipeline outputs
//
// # Include Equal Values
//
// Include both differences and matches in comparison reports:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithIncludeEqual(true),
//	)
//
//	result := comp.CompareWithDiff(user1, user2)
//	for _, diff := range result.Differences {
//	    if diff.Detail.Type == "equal" {
//	        fmt.Printf("✓ %s: values match\n", diff.Path)
//	    } else {
//	        fmt.Printf("✗ %s: %s\n", diff.Path, diff.Message)
//	    }
//	}
//
// Equal values are marked with:
//
//   - Detail.Type = "equal"
//   - Severity = "info"
//
// Use cases for including equal values:
//
//   - Comprehensive audit trails
//   - Configuration validation reports
//   - Debugging comparison logic
//   - Understanding what hasn't changed between versions
//
// # Combined Features
//
// Combine colorization and equal values for complete reports:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithColorize(true),
//	    comparator.WithIncludeEqual(true),
//	    comparator.WithOutputFormat("text"),
//	)
//
//	result := comp.CompareWithDiff(oldConfig, newConfig)
//	formatted, _ := comp.FormatDiff(result, "text")
//	fmt.Println(formatted)
//
// This produces a color-coded report showing both differences and matches,
// making it easy to see the complete picture of the comparison.
//
// # Supported Types
//
// The comparator handles all Go types including:
//
//   - Primitives: bool, int*, uint*, float*, complex*, string
//   - Composite: arrays, slices, maps, structs
//   - Reference: pointers, interfaces, channels, functions
//   - Special: time.Time (uses native Equal method)
//
// # Thread Safety
//
// Comparator instances are not thread-safe. Create separate instances for
// concurrent use, or synchronize access with a mutex.
//
// # Performance Considerations
//
//   - For repeated comparisons with the same configuration, reuse comparator instances
//   - Use WithMaxDepth to limit recursion for deeply nested structures
//   - Use WithMaxDiffs to limit memory usage when many differences exist
//   - Use IgnoreStructFields to skip expensive field comparisons
//   - For large slices, avoid IgnoreSliceOrder when possible (use sorted slices)
//
// # Error Handling
//
// Most methods return errors for future compatibility, but current implementations
// rarely return errors. Always check error returns for forward compatibility.
package comparator

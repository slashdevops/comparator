package comparator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// Comparable is a generic interface for types that can compare themselves to other instances
// of the same type. Implementations should return:
//   - negative integer if the receiver is less than other
//   - zero if the receiver equals other
//   - positive integer if the receiver is greater than other
//
// This interface is useful for implementing custom ordering and sorting logic.
type Comparable[T any] interface {
	// CompareTo compares this instance with another instance of the same type.
	// Returns -1 if this < other, 0 if this == other, 1 if this > other.
	CompareTo(other T) int
}

// Equatable is a generic interface for types that can determine equality with other instances
// of the same type. This provides a type-safe way to implement custom equality logic.
//
// Example implementation:
//
//	type Person struct {
//	    ID   int
//	    Name string
//	}
//
//	func (p Person) Equals(other Person) bool {
//	    return p.ID == other.ID
//	}
type Equatable[T any] interface {
	// Equals returns true if this instance is considered equal to the other instance.
	Equals(other T) bool
}

// Comparator is the main comparison interface that provides deep equality checking and
// difference detection for any Go values. It supports various data types including
// primitives, structs, slices, maps, pointers, and complex nested structures.
//
// The comparator performs deep recursive comparison with cycle detection to handle
// self-referential data structures safely. It can be configured with various options
// to customize comparison behavior.
//
// Example usage:
//
//	comp := comparator.New()
//	if comp.Equal(obj1, obj2) {
//	    fmt.Println("Objects are equal")
//	}
type Comparator interface {
	// Equal performs a deep equality check between two values using default configuration.
	// Returns true if the values are deeply equal, false otherwise.
	Equal(a, b any) bool

	// EqualWithConfig performs a deep equality check using a custom configuration.
	// This allows fine-grained control over comparison behavior such as float precision,
	// slice ordering, and field ignoring.
	EqualWithConfig(a, b any, config *Config) bool

	// Diff returns a list of all differences found between two values.
	// Each difference includes the path, type, values, and contextual information.
	// Returns an error if the comparison process encounters an issue.
	Diff(a, b any) ([]Difference, error)
}

// DiffComparer extends the Comparator interface with advanced diffing capabilities.
// It provides multiple output formats for differences including unified diff, JSON Patch,
// and visual tree representations. This interface is ideal for generating human-readable
// or machine-processable diff reports.
//
// Example usage:
//
//	diffComp := comparator.NewDiffComparer(
//	    comparator.WithOutputFormat("markdown"),
//	    comparator.WithMaxDiffs(100),
//	)
//	result := diffComp.CompareWithDiff(obj1, obj2)
//	fmt.Println(result.Summary)
type DiffComparer interface {
	Comparator

	// CompareWithDiff performs a comprehensive comparison and returns a detailed result
	// containing all differences, statistics, and a summary.
	CompareWithDiff(a, b any) *DiffResult

	// GetUnifiedDiff generates a unified diff format output similar to the Unix diff utility.
	// This is useful for displaying differences in a familiar, line-oriented format.
	GetUnifiedDiff(a, b any) (*UnifiedDiff, error)

	// GetJSONPatch generates a JSON Patch (RFC 6902) document describing the differences.
	// The patch can be applied to transform the first value into the second value.
	GetJSONPatch(a, b any) ([]JSONPatchOperation, error)

	// GetVisualDiff generates a tree-based visual representation of the differences.
	// This is useful for creating graphical diff viewers or hierarchical displays.
	GetVisualDiff(a, b any) (*VisualDiff, error)

	// FormatDiff formats a diff result according to the specified format.
	// Supported formats: "text", "json", "markdown", "html".
	FormatDiff(result *DiffResult, format string) (string, error)
}

// Option is a function type that configures a Config instance.
// Options follow the functional options pattern, allowing flexible and readable
// configuration of comparators.
//
// Example usage:
//
//	comp := comparator.NewWithOptions(
//	    comparator.WithFloatPrecision(1e-6),
//	    comparator.IgnoreSliceOrder(),
//	    comparator.WithMaxDepth(10),
//	)
type Option func(*Config)

// Config holds all configuration options for comparison operations.
// It controls how values are compared, what differences are reported,
// and how output is formatted. Use Option functions to create and modify
// Config instances rather than directly manipulating fields.
//
// The zero value is not safe to use; always create Config instances using
// defaultConfig() or through New/NewWithOptions functions.
type Config struct {
	ignoreStructFields map[string]bool
	customComparators  map[reflect.Type]any
	timeLayout         string
	outputFormat       string
	floatPrecision     float64
	maxDepth           int
	diffMode           DiffMode
	maxDiffs           int
	contextSize        int
	ignoreSliceOrder   bool
	ignoreUnexported   bool
	equateEmpty        bool
	equateNaNs         bool
	showContext        bool
	includeEqual       bool
	colorize           bool
}

// DiffMode specifies the format and detail level for reporting differences.
// Different modes are suitable for different use cases, from simple boolean checks
// to detailed patch operations.
type DiffMode int

const (
	// DiffModeSimple provides basic difference reporting with minimal detail.
	// Suitable for quick equality checks with simple diff messages.
	DiffModeSimple DiffMode = iota

	// DiffModeFull provides comprehensive difference reporting with full context,
	// including nested differences and detailed statistics.
	DiffModeFull

	// DiffModeUnified generates differences in unified diff format,
	// similar to the Unix diff utility output.
	DiffModeUnified

	// DiffModeJSONPatch generates differences as JSON Patch operations (RFC 6902),
	// suitable for programmatic application of changes.
	DiffModeJSONPatch

	// DiffModeVisual generates a tree-based visual representation of differences,
	// suitable for graphical diff viewers.
	DiffModeVisual
)

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		floatPrecision:     1e-9,
		maxDepth:           0,
		equateEmpty:        true,
		equateNaNs:         false,
		timeLayout:         time.RFC3339Nano,
		diffMode:           DiffModeSimple,
		maxDiffs:           1000,
		showContext:        true,
		contextSize:        3,
		outputFormat:       "text",
		ignoreStructFields: make(map[string]bool),
		customComparators:  make(map[reflect.Type]any),
	}
}

// WithFloatPrecision sets the precision threshold for floating-point comparisons.
// Two float values are considered equal if their absolute difference is less than
// or equal to the specified precision.
//
// The default precision is 1e-9. Use a larger value for less strict comparisons,
// or a smaller value for more strict comparisons.
//
// Example:
//
//	comp := comparator.NewWithOptions(
//	    comparator.WithFloatPrecision(1e-6),
//	)
//	// 1.0000001 and 1.0000002 will be considered equal
func WithFloatPrecision(precision float64) Option {
	return func(c *Config) {
		c.floatPrecision = precision
	}
}

// IgnoreSliceOrder configures the comparator to treat slices as sets,
// ignoring the order of elements. When enabled, [1, 2, 3] will be
// considered equal to [3, 1, 2].
//
// This is useful when comparing collections where order doesn't matter,
// such as sets represented as slices or unordered query results.
//
// Note: For simple types, elements are sorted before comparison.
// For complex types, a more expensive matching algorithm is used.
func IgnoreSliceOrder() Option {
	return func(c *Config) {
		c.ignoreSliceOrder = true
	}
}

// WithMaxDepth sets the maximum recursion depth for comparisons.
// When the specified depth is reached, the comparator performs a shallow
// comparison instead of recursing further.
//
// A depth of 0 (default) means unlimited depth. Use this option to prevent
// stack overflows with deeply nested structures or to improve performance
// by limiting comparison depth.
//
// Example:
//
//	comp := comparator.NewWithOptions(
//	    comparator.WithMaxDepth(5), // Only compare 5 levels deep
//	)
func WithMaxDepth(depth int) Option {
	return func(c *Config) {
		c.maxDepth = depth
	}
}

// IgnoreUnexported configures the comparator to skip unexported struct fields.
// By default, unexported fields are compared if accessible through reflection.
// Enable this option to only compare exported (public) fields.
//
// This is useful when comparing structs where unexported fields may contain
// internal state that shouldn't affect equality, or when unexported fields
// are not accessible.
func IgnoreUnexported() Option {
	return func(c *Config) {
		c.ignoreUnexported = true
	}
}

// EquateEmpty configures the comparator to treat nil and empty values as equal.
// When enabled:
//   - nil slice equals empty slice ([]int(nil) == []int{})
//   - nil map equals empty map (map[string]int(nil) == map[string]int{})
//   - nil pointer equals pointer to zero value
//   - empty string equals empty string
//
// This is enabled by default. Use this option explicitly when creating
// custom configurations to ensure consistent behavior.
func EquateEmpty() Option {
	return func(c *Config) {
		c.equateEmpty = true
	}
}

// EquateNaNs configures the comparator to treat NaN (Not-a-Number) values as equal.
// By default, NaN != NaN according to IEEE 754 standard.
// Enable this option to consider all NaN values equal to each other.
//
// This is useful in scientific computing or when NaN represents
// "missing data" or "undefined" consistently across both values.
func EquateNaNs() Option {
	return func(c *Config) {
		c.equateNaNs = true
	}
}

// IgnoreStructFields configures the comparator to skip specified struct fields
// by name. This applies to all structs encountered during comparison.
//
// Useful for ignoring fields like timestamps, IDs, or other fields that
// are expected to differ but shouldn't affect equality.
//
// Example:
//
//	comp := comparator.NewWithOptions(
//	    comparator.IgnoreStructFields("CreatedAt", "UpdatedAt", "ID"),
//	)
func IgnoreStructFields(fields ...string) Option {
	return func(c *Config) {
		for _, field := range fields {
			c.ignoreStructFields[field] = true
		}
	}
}

// WithTimeLayout sets the time format layout for time.Time comparisons.
// This is currently used for formatting but doesn't affect equality checks
// (time.Time values use the native Equal method).
//
// The default layout is time.RFC3339Nano.
//
// Example:
//
//	comp := comparator.NewWithOptions(
//	    comparator.WithTimeLayout(time.RFC3339),
//	)
func WithTimeLayout(layout string) Option {
	return func(c *Config) {
		c.timeLayout = layout
	}
}

// WithCustomComparator registers a custom comparison function for a specific type T.
// The comparator function receives two values of type T and returns true if they
// should be considered equal.
//
// Custom comparators take precedence over default comparison logic for the specified type.
// This is useful for types with special equality semantics or when you need to override
// default behavior.
//
// Example:
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
func WithCustomComparator[T any](comparator func(a, b T) bool) Option {
	return func(c *Config) {
		var zero T
		typ := reflect.TypeOf(zero)
		c.customComparators[typ] = comparator
	}
}

// WithDiffMode sets the diff reporting mode.
// See DiffMode constants for available modes and their descriptions.
//
// The default mode is DiffModeSimple.
//
// Example:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithDiffMode(comparator.DiffModeUnified),
//	)
func WithDiffMode(mode DiffMode) Option {
	return func(c *Config) {
		c.diffMode = mode
	}
}

// WithMaxDiffs sets the maximum number of differences to collect.
// Once this limit is reached, no more differences will be recorded.
// This helps prevent memory issues when comparing large structures
// with many differences.
//
// The default is 1000 differences. Set to 0 for unlimited (use with caution).
//
// Example:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithMaxDiffs(50), // Only collect first 50 differences
//	)
func WithMaxDiffs(max int) Option {
	return func(c *Config) {
		c.maxDiffs = max
	}
}

// WithOutputFormat sets the output format for formatted diffs.
// Supported formats:
//   - "text": Plain text output (default)
//   - "json": JSON format
//   - "markdown": Markdown format
//   - "html": HTML format
//
// Example:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithOutputFormat("markdown"),
//	)
func WithOutputFormat(format string) Option {
	return func(c *Config) {
		c.outputFormat = format
	}
}

// WithColorize enables colored output for text format diffs.
// When enabled, differences are highlighted using ANSI color codes:
//   - Green: added values
//   - Red: removed values
//   - Yellow: warnings
//   - Cyan: paths and labels
//
// This option only affects text format output. Other formats (JSON, HTML, Markdown)
// have their own styling mechanisms.
//
// Example:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithColorize(true),
//	    comparator.WithOutputFormat("text"),
//	)
func WithColorize(enable bool) Option {
	return func(c *Config) {
		c.colorize = enable
	}
}

// WithIncludeEqual configures the comparator to include equal values in the diff output.
// By default, only differences are reported. When enabled, the diff will also show
// paths where values are equal, providing a complete picture of the comparison.
//
// This is useful for:
//   - Generating comprehensive comparison reports
//   - Debugging comparison logic
//   - Creating detailed audit trails
//
// Note: Enabling this option may significantly increase output size for large structures.
//
// Example:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithIncludeEqual(true),
//	)
func WithIncludeEqual(enable bool) Option {
	return func(c *Config) {
		c.includeEqual = enable
	}
}

// Difference represents a single difference found during comparison.
// It includes the location (path), depth level, description, detailed information,
// severity classification, and optional suggestions for resolution.
//
// The Path field uses dot notation for nested structures:
//   - "User.Name" for struct fields
//   - "Users[0]" for slice elements
//   - "Config[key]" for map entries
type Difference struct {
	// Path is the location of the difference using dot notation (e.g., "User.Address.City")
	Path string `json:"path"`

	// Message is a human-readable description of the difference
	Message string `json:"message"`

	// Severity classifies the importance: "error", "warning", "info"
	Severity string `json:"severity"`

	// Detail contains granular information about the difference
	Detail DifferenceDetail `json:"detail"`

	// Suggestions provides optional recommendations for resolving the difference
	Suggestions []string `json:"suggestions,omitempty"`

	// Level indicates the depth in the structure where this difference was found (0 = root)
	Level int `json:"level"`
}

// DifferenceDetail provides granular, type-specific information about a difference.
// It includes the type of difference, the expected and actual types and values,
// and any nested differences for complex structures.
type DifferenceDetail struct {
	// Type describes the kind of difference:
	// "value_different", "type_mismatch", "missing", "extra_element",
	// "length_mismatch", "missing_key", "extra_key", "nil_mismatch"
	Type string `json:"type"`

	// ExpectedType is the string representation of the expected value's type
	ExpectedType string `json:"expected_type,omitempty"`

	// ActualType is the string representation of the actual value's type
	ActualType string `json:"actual_type,omitempty"`

	// ExpectedValue is the expected value (first argument to comparison)
	ExpectedValue any `json:"expected_value,omitempty"`

	// ActualValue is the actual value (second argument to comparison)
	ActualValue any `json:"actual_value,omitempty"`

	// Children contains nested differences for complex structures
	Children []Difference `json:"children,omitempty"`
}

// DiffResult holds the complete result of a comparison operation.
// It includes equality status, all differences found, statistics about
// the comparison process, and a human-readable summary.
//
// Example:
//
//	result := comp.CompareWithDiff(obj1, obj2)
//	if !result.Equal {
//	    fmt.Printf("Found %d differences\n", len(result.Differences))
//	    fmt.Println(result.Summary)
//	    for _, diff := range result.Differences {
//	        fmt.Printf("  %s: %s\n", diff.Path, diff.Message)
//	    }
//	}
type DiffResult struct {
	// Summary is a human-readable summary of the comparison result
	Summary string `json:"summary,omitempty"`

	// Differences contains all differences found (empty if Equal is true)
	Differences []Difference `json:"differences,omitempty"`

	// PathStats provides statistics about the comparison process
	PathStats PathStats `json:"stats"`

	// Equal indicates whether the two values are deeply equal
	Equal bool `json:"equal"`
}

// PathStats provides statistical information about the comparison operation.
// It tracks how many nodes were examined, compared, found different, or ignored.
//
// These statistics are useful for understanding the scope and efficiency of
// the comparison, especially for large or complex data structures.
type PathStats struct {
	// TotalNodes is the total number of nodes encountered during traversal
	TotalNodes int `json:"total_nodes"`

	// ComparedNodes is the number of nodes that were actually compared
	ComparedNodes int `json:"compared_nodes"`

	// DifferentNodes is the number of nodes found to be different
	DifferentNodes int `json:"different_nodes"`

	// IgnoredNodes is the number of nodes skipped (e.g., due to field ignore rules)
	IgnoredNodes int `json:"ignored_nodes"`
}

// UnifiedDiff represents differences in unified diff format, similar to
// the output of the Unix diff utility. This format is familiar to developers
// and integrates well with version control systems.
//
// The format uses '+' and '-' prefixes to indicate additions and removals.
type UnifiedDiff struct {
	// Header contains the diff header (e.g., "--- a\n+++ b\n")
	Header string `json:"header"`

	// Chunks contains the individual diff chunks with context
	Chunks []Chunk `json:"chunks"`
}

// Chunk represents a section of a unified diff, containing a group of
// related changes along with surrounding context lines.
type Chunk struct {
	// Context describes the location of this chunk (e.g., "@@ -1,10 +1,12 @@")
	Context string `json:"context"`

	// Changes contains the individual line changes in this chunk
	Changes []Change `json:"changes"`
}

// Change represents a single line change in a unified diff.
type Change struct {
	// Type indicates the change type: "add" (+), "remove" (-), or "context" ( )
	Type string `json:"type"`

	// Content is the line content
	Content string `json:"content"`

	// Line is the line number in the original or modified file
	Line int `json:"line,omitempty"`
}

// JSONPatchOperation represents a single operation in a JSON Patch document (RFC 6902).
// JSON Patch is a format for describing changes to a JSON document.
// It can be used to apply partial updates, avoiding the need to send the entire document.
//
// Example operations:
//   - {"op": "replace", "path": "/name", "value": "John"}
//   - {"op": "add", "path": "/tags/-", "value": "new-tag"}
//   - {"op": "remove", "path": "/deprecated"}
//   - {"op": "move", "from": "/old", "path": "/new"}
type JSONPatchOperation struct {
	// Op is the operation type: "add", "remove", "replace", "move", "copy", or "test"
	Op string `json:"op"`

	// Path is the JSON Pointer (RFC 6901) to the target location
	Path string `json:"path"`

	// Value is the value to add, replace, or test (omitted for remove and move)
	Value any `json:"value,omitempty"`

	// From is the source path for move and copy operations
	From string `json:"from,omitempty"`
}

// VisualDiff represents a tree-based visual representation of differences.
// This structure is suitable for building graphical diff viewers or
// hierarchical displays in user interfaces.
//
// Each node in the tree contains information about its path, value,
// status (same/different/added/removed), and child nodes.
type VisualDiff struct {
	// Root is the root node of the visual diff tree
	Root *VisualNode `json:"root"`
}

// VisualNode represents a single node in a visual diff tree.
type VisualNode struct {
	// Path is the full path to this node (e.g., "root.users[0].name")
	Path string `json:"path"`

	// Value is the string representation of this node's value
	Value string `json:"value"`

	// Status indicates the state: "same", "different", "added", "removed"
	Status string `json:"status"`

	// Children contains child nodes for nested structures
	Children []*VisualNode `json:"children,omitempty"`
}

type visit struct {
	typ reflect.Type
	a1  uintptr
	a2  uintptr
}

// defaultComparator implements Comparator
type defaultComparator struct {
	config       *Config
	visited      map[uintptr]visit
	pathStack    []string
	differences  []Difference
	stats        PathStats
	currentLevel int
}

// New creates a new Comparator with default configuration.
// The default comparator uses:
//   - Float precision: 1e-9
//   - Equate empty values: true
//   - Equate NaNs: false
//   - Max depth: unlimited (0)
//   - Time layout: RFC3339Nano
//
// Example:
//
//	comp := comparator.New()
//	if comp.Equal(obj1, obj2) {
//	    fmt.Println("Objects are equal")
//	}
func New() Comparator {
	return &defaultComparator{
		config:  defaultConfig(),
		visited: make(map[uintptr]visit),
	}
}

// NewWithOptions creates a new Comparator with custom options.
// Options are applied in the order provided. Later options override earlier ones.
//
// Example:
//
//	comp := comparator.NewWithOptions(
//	    comparator.WithFloatPrecision(1e-6),
//	    comparator.IgnoreSliceOrder(),
//	    comparator.IgnoreStructFields("ID", "CreatedAt"),
//	)
//
//	if comp.Equal(user1, user2) {
//	    fmt.Println("Users are equal (ignoring ID and CreatedAt)")
//	}
func NewWithOptions(opts ...Option) Comparator {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return &defaultComparator{
		config:  config,
		visited: make(map[uintptr]visit),
	}
}

// NewDiffComparer creates a new DiffComparer with custom options.
// A DiffComparer provides all Comparator functionality plus advanced
// diff reporting capabilities.
//
// Example:
//
//	diffComp := comparator.NewDiffComparer(
//	    comparator.WithOutputFormat("markdown"),
//	    comparator.WithMaxDiffs(100),
//	    comparator.WithDiffMode(comparator.DiffModeUnified),
//	)
//
//	result := diffComp.CompareWithDiff(config1, config2)
//	if !result.Equal {
//	    formatted, _ := diffComp.FormatDiff(result, "markdown")
//	    fmt.Println(formatted)
//	}
func NewDiffComparer(opts ...Option) DiffComparer {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return &defaultComparator{
		config:  config,
		visited: make(map[uintptr]visit),
	}
}

// Equal performs a deep equality check between two values.
// It returns true if the values are deeply equal according to the comparator's
// configuration, false otherwise.
//
// Equal handles:
//   - Nil values (nil == nil, nil != non-nil)
//   - Primitive types (bool, int, float, string, etc.)
//   - Complex types (complex64, complex128)
//   - Structs (including time.Time with special handling)
//   - Slices and arrays (with optional order-independent comparison)
//   - Maps (key-value pairs must match)
//   - Pointers and interfaces (dereferences and compares values)
//   - Circular references (detected and handled safely)
//
// The comparison respects all configuration options set during creation.
func (c *defaultComparator) Equal(a, b any) bool {
	c.resetState()
	return c.equal(reflect.ValueOf(a), reflect.ValueOf(b), nil)
}

// EqualWithConfig performs a deep equality check using a custom configuration.
// This allows one-off comparisons with different settings without creating
// a new comparator instance.
//
// The provided config replaces the comparator's configuration for this
// comparison only. Subsequent calls use the original configuration.
//
// Example:
//
//	comp := comparator.New()
//	customConfig := &comparator.Config{
//	    // ... custom settings
//	}
//	if comp.EqualWithConfig(obj1, obj2, customConfig) {
//	    fmt.Println("Equal with custom config")
//	}
func (c *defaultComparator) EqualWithConfig(a, b any, config *Config) bool {
	c.config = config
	c.resetState()
	return c.equal(reflect.ValueOf(a), reflect.ValueOf(b), nil)
}

// Diff returns a list of all differences found between two values.
// Each difference includes the path, level, message, detailed information,
// severity, and suggestions.
//
// This method performs a comprehensive comparison and collects all differences
// up to the configured maximum (see WithMaxDiffs). It always returns a non-nil
// slice (empty if no differences found) and nil error (error return is for
// future compatibility).
//
// Example:
//
//	comp := comparator.New()
//	diffs, err := comp.Diff(expected, actual)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, diff := range diffs {
//	    fmt.Printf("%s: %s\n", diff.Path, diff.Message)
//	}
func (c *defaultComparator) Diff(a, b any) ([]Difference, error) {
	c.resetState()
	c.differences = make([]Difference, 0)
	c.collectDifferences(reflect.ValueOf(a), reflect.ValueOf(b), "")
	return c.differences, nil
}

// CompareWithDiff performs a comprehensive comparison and returns a detailed result.
// The result includes equality status, all differences, statistics, and a summary.
//
// This is the recommended method for detailed difference analysis, as it provides
// the most complete information about the comparison.
//
// Example:
//
//	comp := comparator.NewDiffComparer()
//	result := comp.CompareWithDiff(expected, actual)
//
//	if !result.Equal {
//	    fmt.Printf("Found %d differences\n", len(result.Differences))
//	    fmt.Printf("Summary: %s\n", result.Summary)
//	    fmt.Printf("Stats: %+v\n", result.PathStats)
//
//	    for _, diff := range result.Differences {
//	        fmt.Printf("  [%s] %s: %s\n", diff.Severity, diff.Path, diff.Message)
//	        if len(diff.Suggestions) > 0 {
//	            fmt.Printf("    Suggestions: %v\n", diff.Suggestions)
//	        }
//	    }
//	}
func (c *defaultComparator) CompareWithDiff(a, b any) *DiffResult {
	c.resetState()
	c.differences = make([]Difference, 0)

	equal := c.Equal(a, b)
	c.collectDifferences(reflect.ValueOf(a), reflect.ValueOf(b), "")

	return &DiffResult{
		Equal:       equal,
		Differences: c.differences,
		PathStats:   c.stats,
		Summary:     c.generateSummary(),
	}
}

// GetUnifiedDiff generates a unified diff format output similar to the Unix diff utility.
// The output uses '+' for additions, '-' for removals, and context lines for unchanged content.
//
// This format is familiar to developers and can be easily integrated with version
// control systems or diff viewers.
//
// Example:
//
//	comp := comparator.NewDiffComparer(
//	    comparator.WithOutputFormat("text"),
//	)
//	unifiedDiff, err := comp.GetUnifiedDiff(oldConfig, newConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println(unifiedDiff.Header)
//	for _, chunk := range unifiedDiff.Chunks {
//	    fmt.Println(chunk.Context)
//	    for _, change := range chunk.Changes {
//	        fmt.Printf("%s%s\n", change.Type, change.Content)
//	    }
//	}
func (c *defaultComparator) GetUnifiedDiff(a, b any) (*UnifiedDiff, error) {
	aStr := fmt.Sprintf("%+v", a)
	bStr := fmt.Sprintf("%+v", b)

	aLines := strings.Split(aStr, "\n")
	bLines := strings.Split(bStr, "\n")

	chunks := make([]Chunk, 0)
	changes := make([]Change, 0)

	contextSize := c.config.contextSize
	if contextSize <= 0 {
		contextSize = 3
	}
	_ = contextSize // contextSize will be used for context lines in future implementation

	for i := 0; i < min(len(aLines), len(bLines)); i++ {
		if aLines[i] != bLines[i] {
			changes = append(changes, Change{
				Type:    "removed",
				Content: "- " + aLines[i],
				Line:    i + 1,
			}, Change{
				Type:    "added",
				Content: "+ " + bLines[i],
				Line:    i + 1,
			})
		} else if c.config.showContext && len(changes) > 0 {
			changes = append(changes, Change{
				Type:    "context",
				Content: "  " + aLines[i],
				Line:    i + 1,
			})
		}
	}

	if len(changes) > 0 {
		chunks = append(chunks, Chunk{
			Context: fmt.Sprintf("@@ -1,%d +1,%d @@", len(aLines), len(bLines)),
			Changes: changes,
		})
	}

	return &UnifiedDiff{
		Header: "--- a\n+++ b\n",
		Chunks: chunks,
	}, nil
}

// GetJSONPatch generates a JSON Patch (RFC 6902) document describing the differences.
// The returned operations can be applied to the first value to transform it into the second value.
//
// JSON Patch operations include:
//   - "add": Add a new value at a path
//   - "remove": Remove the value at a path
//   - "replace": Replace the value at a path
//   - "move": Move a value from one path to another
//   - "copy": Copy a value from one path to another
//   - "test": Test that a value at a path equals a specified value
//
// This format is widely supported and can be used for partial updates in REST APIs,
// database operations, and configuration management.
//
// Example:
//
//	comp := comparator.NewDiffComparer()
//	patch, err := comp.GetJSONPatch(oldDoc, newDoc)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Serialize to JSON
//	patchJSON, _ := json.MarshalIndent(patch, "", "  ")
//	fmt.Println(string(patchJSON))
//
//	// Apply patch using a JSON Patch library
//	// modified, err := jsonpatch.Apply(oldDoc, patchJSON)
func (c *defaultComparator) GetJSONPatch(a, b any) ([]JSONPatchOperation, error) {
	result := c.CompareWithDiff(a, b)
	operations := make([]JSONPatchOperation, 0)

	for _, diff := range result.Differences {
		if op := c.diffToJSONPatch(diff); op != nil {
			operations = append(operations, *op)
		}
	}

	return operations, nil
}

// GetVisualDiff generates a tree-based visual representation of the differences.
// The result is a hierarchical structure suitable for building graphical diff viewers,
// web UIs, or ASCII tree displays.
//
// Each node in the tree contains:
//   - Path: The location in the data structure
//   - Value: String representation of the value
//   - Status: "same", "different", "added", or "removed"
//   - Children: Nested nodes for complex structures
//
// Example:
//
//	comp := comparator.NewDiffComparer()
//	visualDiff, err := comp.GetVisualDiff(obj1, obj2)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Print tree structure
//	printNode(visualDiff.Root, 0)
//
//	func printNode(node *VisualNode, indent int) {
//	    prefix := strings.Repeat("  ", indent)
//	    symbol := ""
//	    switch node.Status {
//	    case "added": symbol = "+ "
//	    case "removed": symbol = "- "
//	    case "different": symbol = "~ "
//	    default: symbol = "  "
//	    }
//	    fmt.Printf("%s%s%s: %s\n", prefix, symbol, node.Path, node.Value)
//	    for _, child := range node.Children {
//	        printNode(child, indent+1)
//	    }
//	}
func (c *defaultComparator) GetVisualDiff(a, b any) (*VisualDiff, error) {
	result := c.CompareWithDiff(a, b)
	root := &VisualNode{
		Path:   "/",
		Status: "same",
	}

	c.buildVisualTree(result.Differences, root)
	return &VisualDiff{Root: root}, nil
}

// FormatDiff formats a diff result according to the specified output format.
// Supported formats:
//   - "text": Plain text output with indentation and symbols
//   - "json": JSON representation of the diff result
//   - "markdown": Markdown-formatted output suitable for documentation
//   - "html": HTML output with styling and structure
//
// If an unsupported format is specified, defaults to "text" format.
//
// Example:
//
//	comp := comparator.NewDiffComparer()
//	result := comp.CompareWithDiff(config1, config2)
//
//	// Format as markdown
//	markdown, err := comp.FormatDiff(result, "markdown")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("diff-report.md", []byte(markdown), 0644)
//
//	// Format as HTML
//	html, err := comp.FormatDiff(result, "html")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("diff-report.html", []byte(html), 0644)
func (c *defaultComparator) FormatDiff(result *DiffResult, format string) (string, error) {
	switch format {
	case "text":
		return c.formatTextDiff(result), nil
	case "json":
		return c.formatJSONDiff(result), nil
	case "markdown":
		return c.formatMarkdownDiff(result), nil
	case "html":
		return c.formatHTMLDiff(result), nil
	default:
		return c.formatTextDiff(result), nil
	}
}

// ==================== Internal Methods ====================

func (c *defaultComparator) resetState() {
	c.visited = make(map[uintptr]visit)
	c.pathStack = make([]string, 0)
	c.currentLevel = 0
	c.stats = PathStats{}
	c.differences = make([]Difference, 0)
}

func (c *defaultComparator) equal(av, bv reflect.Value, path []string) bool {
	if path == nil {
		c.visited = make(map[uintptr]visit)
	}

	if c.config.maxDepth > 0 && len(path) >= c.config.maxDepth {
		return c.shallowEqual(av, bv)
	}

	if !av.IsValid() || !bv.IsValid() {
		if c.config.equateEmpty {
			return c.isEmpty(av) && c.isEmpty(bv)
		}
		return av.IsValid() == bv.IsValid()
	}

	if c.isCyclic(av, bv) {
		return true
	}

	if comparator, ok := c.config.customComparators[av.Type()]; ok {
		return c.useCustomComparator(av, bv, comparator)
	}

	if av.Type() != bv.Type() {
		return c.equalNumeric(av, bv)
	}

	switch av.Kind() {
	case reflect.Bool:
		return av.Bool() == bv.Bool()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return av.Int() == bv.Int()

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return av.Uint() == bv.Uint()

	case reflect.Float32, reflect.Float64:
		return c.equalFloats(av.Float(), bv.Float())

	case reflect.Complex64, reflect.Complex128:
		return c.equalComplex(av.Complex(), bv.Complex())

	case reflect.String:
		return av.String() == bv.String()

	case reflect.Array, reflect.Slice:
		return c.equalSlices(av, bv, path)

	case reflect.Map:
		return c.equalMaps(av, bv, path)

	case reflect.Struct:
		return c.equalStructs(av, bv, path)

	case reflect.Pointer, reflect.Interface:
		return c.equalPointers(av, bv, path)

	case reflect.Chan:
		return av.Type() == bv.Type() && av.Cap() == bv.Cap()

	case reflect.Func:
		return av.IsNil() && bv.IsNil()

	default:
		return av.Pointer() == bv.Pointer()
	}
}

func (c *defaultComparator) equalSlices(av, bv reflect.Value, path []string) bool {
	if av.Len() != bv.Len() && !c.config.ignoreSliceOrder {
		return false
	}

	// Only check IsNil for slices (not arrays, which cannot be nil)
	if av.Kind() == reflect.Slice {
		if av.IsNil() || bv.IsNil() {
			if c.config.equateEmpty {
				return (av.IsNil() || av.Len() == 0) && (bv.IsNil() || bv.Len() == 0)
			}
			return av.IsNil() == bv.IsNil()
		}
	}

	if c.config.ignoreSliceOrder {
		return c.equalSlicesIgnoreOrder(av, bv, path)
	}

	for i := 0; i < av.Len(); i++ {
		newPath := append(path, fmt.Sprintf("[%d]", i))
		if !c.equal(av.Index(i), bv.Index(i), newPath) {
			return false
		}
	}
	return true
}

func (c *defaultComparator) equalSlicesIgnoreOrder(av, bv reflect.Value, path []string) bool {
	if av.Len() != bv.Len() {
		return false
	}

	if c.isSimpleType(av.Type().Elem()) {
		return c.equalSortedSlices(av, bv, path)
	}

	matched := make([]bool, bv.Len())
	for i := 0; i < av.Len(); i++ {
		found := false
		for j := 0; j < bv.Len(); j++ {
			if matched[j] {
				continue
			}
			newPath := append(path, fmt.Sprintf("[%d]", i))
			if c.equal(av.Index(i), bv.Index(j), newPath) {
				matched[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (c *defaultComparator) equalSortedSlices(av, bv reflect.Value, path []string) bool {
	aSlice := make([]any, av.Len())
	bSlice := make([]any, bv.Len())

	for i := 0; i < av.Len(); i++ {
		aSlice[i] = av.Index(i).Interface()
	}
	for i := 0; i < bv.Len(); i++ {
		bSlice[i] = bv.Index(i).Interface()
	}

	sort.Slice(aSlice, func(i, j int) bool {
		return c.compareValues(aSlice[i], aSlice[j]) < 0
	})
	sort.Slice(bSlice, func(i, j int) bool {
		return c.compareValues(bSlice[i], bSlice[j]) < 0
	})

	for i := range aSlice {
		newPath := append(path, fmt.Sprintf("[%d]", i))
		if !c.equal(reflect.ValueOf(aSlice[i]), reflect.ValueOf(bSlice[i]), newPath) {
			return false
		}
	}
	return true
}

func (c *defaultComparator) equalMaps(av, bv reflect.Value, path []string) bool {
	if av.Len() != bv.Len() {
		return false
	}

	if av.IsNil() || bv.IsNil() {
		if c.config.equateEmpty {
			return (av.IsNil() || av.Len() == 0) && (bv.IsNil() || bv.Len() == 0)
		}
		return av.IsNil() == bv.IsNil()
	}

	for _, key := range av.MapKeys() {
		bVal := bv.MapIndex(key)
		if !bVal.IsValid() {
			return false
		}

		newPath := append(path, fmt.Sprintf("[%v]", key.Interface()))
		if !c.equal(av.MapIndex(key), bVal, newPath) {
			return false
		}
	}
	return true
}

func (c *defaultComparator) equalStructs(av, bv reflect.Value, path []string) bool {
	if av.Type() == reflect.TypeFor[time.Time]() {
		return c.equalTimes(av, bv)
	}

	for i := 0; i < av.NumField(); i++ {
		field := av.Type().Field(i)

		if c.config.ignoreUnexported && field.PkgPath != "" {
			continue
		}

		if c.config.ignoreStructFields[field.Name] {
			continue
		}

		newPath := append(path, field.Name)
		if !c.equal(av.Field(i), bv.Field(i), newPath) {
			return false
		}
	}
	return true
}

func (c *defaultComparator) equalPointers(av, bv reflect.Value, path []string) bool {
	if av.IsNil() || bv.IsNil() {
		return av.IsNil() == bv.IsNil()
	}

	// For interfaces, just compare the elements directly
	if av.Kind() == reflect.Interface || bv.Kind() == reflect.Interface {
		return c.equal(av.Elem(), bv.Elem(), path)
	}

	// For pointers, check if they point to the same address
	if av.Pointer() == bv.Pointer() {
		return true
	}

	return c.equal(av.Elem(), bv.Elem(), path)
}

func (c *defaultComparator) equalFloats(a, b float64) bool {
	if c.config.equateNaNs && isNaN(a) && isNaN(b) {
		return true
	}

	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= c.config.floatPrecision
}

func (c *defaultComparator) equalComplex(a, b complex128) bool {
	realEqual := c.equalFloats(real(a), real(b))
	imagEqual := c.equalFloats(imag(a), imag(b))
	return realEqual && imagEqual
}

func (c *defaultComparator) equalTimes(av, bv reflect.Value) bool {
	t1 := av.Interface().(time.Time)
	t2 := bv.Interface().(time.Time)
	return t1.Equal(t2)
}

func (c *defaultComparator) equalNumeric(av, bv reflect.Value) bool {
	var aFloat, bFloat float64

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		aFloat = float64(av.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		aFloat = float64(av.Uint())
	case reflect.Float32, reflect.Float64:
		aFloat = av.Float()
	default:
		return false
	}

	switch bv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bFloat = float64(bv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bFloat = float64(bv.Uint())
	case reflect.Float32, reflect.Float64:
		bFloat = bv.Float()
	default:
		return false
	}

	return c.equalFloats(aFloat, bFloat)
}

func (c *defaultComparator) isEmpty(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	default:
		return false
	}
}

func (c *defaultComparator) isCyclic(av, bv reflect.Value) bool {
	if av.CanAddr() && bv.CanAddr() {
		addr1 := av.UnsafeAddr()
		addr2 := bv.UnsafeAddr()

		if addr1 < addr2 {
			addr1, addr2 = addr2, addr1
		}

		typ := av.Type()
		key := uintptr(addr1)<<32 | uintptr(addr2)
		if visit, ok := c.visited[key]; ok && visit.typ == typ {
			return true
		}

		c.visited[key] = visit{typ: typ, a1: addr1, a2: addr2}
	}
	return false
}

func (c *defaultComparator) useCustomComparator(av, bv reflect.Value, comparator any) bool {
	fn := reflect.ValueOf(comparator)
	result := fn.Call([]reflect.Value{av, bv})
	return result[0].Bool()
}

func (c *defaultComparator) isSimpleType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}

func (c *defaultComparator) compareValues(a, b any) int {
	return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

func (c *defaultComparator) shallowEqual(av, bv reflect.Value) bool {
	return fmt.Sprintf("%v", av.Interface()) == fmt.Sprintf("%v", bv.Interface())
}

// ==================== Diff Collection ====================

func (c *defaultComparator) collectDifferences(av, bv reflect.Value, currentPath string) {
	if c.config.maxDiffs > 0 && len(c.differences) >= c.config.maxDiffs {
		return
	}

	c.stats.TotalNodes++
	c.pathStack = append(c.pathStack, currentPath)
	c.currentLevel++
	defer func() {
		c.pathStack = c.pathStack[:len(c.pathStack)-1]
		c.currentLevel--
	}()

	if !av.IsValid() && !bv.IsValid() {
		return
	}

	if !av.IsValid() || !bv.IsValid() {
		c.addMissingDiff(av, bv, currentPath)
		return
	}

	if av.Type() != bv.Type() {
		c.addTypeMismatchDiff(av, bv, currentPath)
		return
	}

	if c.isCyclic(av, bv) {
		return
	}

	c.stats.ComparedNodes++

	switch av.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		c.comparePrimitive(av, bv, currentPath)

	case reflect.Complex64, reflect.Complex128:
		c.compareComplex(av, bv, currentPath)

	case reflect.Slice, reflect.Array:
		c.compareSlices(av, bv, currentPath)

	case reflect.Map:
		c.compareMaps(av, bv, currentPath)

	case reflect.Struct:
		c.compareStructs(av, bv, currentPath)

	case reflect.Pointer, reflect.Interface:
		c.comparePointers(av, bv, currentPath)

	default:
		if av.Interface() != bv.Interface() {
			c.addValueDiff(av, bv, currentPath, "values differ")
		}
	}
}

func (c *defaultComparator) comparePrimitive(av, bv reflect.Value, path string) {
	equal := false

	switch av.Kind() {
	case reflect.Bool:
		equal = av.Bool() == bv.Bool()
	case reflect.String:
		equal = av.String() == bv.String()
	case reflect.Float32, reflect.Float64:
		equal = c.equalFloats(av.Float(), bv.Float())
	default:
		equal = av.Interface() == bv.Interface()
	}

	if !equal {
		c.addValueDiff(av, bv, path, fmt.Sprintf("%v != %v", av.Interface(), bv.Interface()))
		c.stats.DifferentNodes++
	} else if c.config.includeEqual {
		c.addEqualDiff(av, bv, path, fmt.Sprintf("%v == %v", av.Interface(), bv.Interface()))
	}
}

func (c *defaultComparator) compareComplex(av, bv reflect.Value, path string) {
	if !c.equalComplex(av.Complex(), bv.Complex()) {
		c.addValueDiff(av, bv, path, "complex numbers differ")
		c.stats.DifferentNodes++
	} else if c.config.includeEqual {
		c.addEqualDiff(av, bv, path, "complex numbers are equal")
	}
}

func (c *defaultComparator) compareSlices(av, bv reflect.Value, path string) {
	if av.Len() != bv.Len() {
		c.addLengthDiff(av, bv, path)
		c.stats.DifferentNodes++

		if !c.config.ignoreSliceOrder {
			minLen := min(bv.Len(), av.Len())

			for i := range minLen {
				elementPath := fmt.Sprintf("%s[%d]", path, i)
				c.collectDifferences(av.Index(i), bv.Index(i), elementPath)
			}

			if av.Len() > bv.Len() {
				for i := bv.Len(); i < av.Len(); i++ {
					c.addExtraElementDiff(av.Index(i), fmt.Sprintf("%s[%d]", path, i), "extra element in first slice")
				}
			} else {
				for i := av.Len(); i < bv.Len(); i++ {
					c.addMissingElementDiff(bv.Index(i), fmt.Sprintf("%s[%d]", path, i), "missing element in first slice")
				}
			}
		}
		return
	}

	if c.config.ignoreSliceOrder {
		c.compareSlicesIgnoreOrder(av, bv, path)
	} else {
		allEqual := true
		for i := 0; i < av.Len(); i++ {
			elementPath := fmt.Sprintf("%s[%d]", path, i)

			beforeDiffs := len(c.differences)
			c.collectDifferences(av.Index(i), bv.Index(i), elementPath)

			if len(c.differences) > beforeDiffs {
				allEqual = false
				c.stats.DifferentNodes++
			}
		}

		if !allEqual {
			c.addValueDiff(av, bv, path, "slice elements differ")
		}
	}
}

func (c *defaultComparator) compareSlicesIgnoreOrder(av, bv reflect.Value, path string) {
	if av.Len() != bv.Len() {
		c.addLengthDiff(av, bv, path)
		return
	}

	matched := make([]bool, bv.Len())
	unmatchedA := make([]int, 0)

	for i := 0; i < av.Len(); i++ {
		matchedIndex := -1
		for j := 0; j < bv.Len(); j++ {
			if matched[j] {
				continue
			}

			beforeDiffs := len(c.differences)
			c.collectDifferences(av.Index(i), bv.Index(j), "")

			if len(c.differences) == beforeDiffs {
				matched[j] = true
				matchedIndex = j
				break
			}

			// Reset differences for this comparison
			c.differences = c.differences[:beforeDiffs]
		}

		if matchedIndex == -1 {
			unmatchedA = append(unmatchedA, i)
		}
	}

	if len(unmatchedA) > 0 {
		c.addValueDiff(av, bv, path, "slices have different elements when ignoring order")
		c.stats.DifferentNodes++

		for _, i := range unmatchedA {
			c.addExtraElementDiff(av.Index(i), fmt.Sprintf("%s[%d]", path, i),
				"element not found in second slice")
		}

		for j := 0; j < bv.Len(); j++ {
			if !matched[j] {
				c.addMissingElementDiff(bv.Index(j), fmt.Sprintf("%s[%d]", path, j),
					"element not found in first slice")
			}
		}
	}
}

func (c *defaultComparator) compareMaps(av, bv reflect.Value, path string) {
	if av.Len() != bv.Len() {
		c.addLengthDiff(av, bv, path)
		c.stats.DifferentNodes++
	}

	for _, key := range av.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		elementPath := fmt.Sprintf("%s[%s]", path, keyStr)

		bVal := bv.MapIndex(key)
		if !bVal.IsValid() {
			c.addMissingKeyDiff(av.MapIndex(key), elementPath, "key missing in second map")
			continue
		}

		c.collectDifferences(av.MapIndex(key), bVal, elementPath)
	}

	for _, key := range bv.MapKeys() {
		if !av.MapIndex(key).IsValid() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			elementPath := fmt.Sprintf("%s[%s]", path, keyStr)
			c.addExtraKeyDiff(bv.MapIndex(key), elementPath, "extra key in second map")
		}
	}
}

func (c *defaultComparator) compareStructs(av, bv reflect.Value, path string) {
	if av.Type() == reflect.TypeFor[time.Time]() {
		if !c.equalTimes(av, bv) {
			c.addValueDiff(av, bv, path, "time values differ")
			c.stats.DifferentNodes++
		}
		return
	}

	for i := 0; i < av.NumField(); i++ {
		field := av.Type().Field(i)
		fieldName := field.Name

		if c.config.ignoreUnexported && field.PkgPath != "" {
			c.stats.IgnoredNodes++
			continue
		}

		if c.config.ignoreStructFields[fieldName] {
			c.stats.IgnoredNodes++
			continue
		}

		fieldPath := path + "." + fieldName
		c.collectDifferences(av.Field(i), bv.Field(i), fieldPath)
	}
}

func (c *defaultComparator) comparePointers(av, bv reflect.Value, path string) {
	if av.IsNil() && bv.IsNil() {
		return
	}

	if av.IsNil() || bv.IsNil() {
		c.addNilDiff(av, bv, path)
		c.stats.DifferentNodes++
		return
	}

	c.collectDifferences(av.Elem(), bv.Elem(), path)
}

// ==================== Diff Creation Helpers ====================

func (c *defaultComparator) addValueDiff(av, bv reflect.Value, path, message string) {
	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  message,
		Severity: "error",
		Detail: DifferenceDetail{
			Type:          "value_different",
			ExpectedValue: av.Interface(),
			ActualValue:   bv.Interface(),
			ExpectedType:  av.Type().String(),
			ActualType:    bv.Type().String(),
		},
		Suggestions: c.generateSuggestions(av, bv),
	})
}

func (c *defaultComparator) addEqualDiff(av, bv reflect.Value, path, message string) {
	if c.config.maxDiffs > 0 && len(c.differences) >= c.config.maxDiffs {
		return
	}

	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  message,
		Severity: "info",
		Detail: DifferenceDetail{
			Type:          "equal",
			ExpectedValue: av.Interface(),
			ActualValue:   bv.Interface(),
			ExpectedType:  av.Type().String(),
			ActualType:    bv.Type().String(),
		},
	})
}

func (c *defaultComparator) addTypeMismatchDiff(av, bv reflect.Value, path string) {
	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  fmt.Sprintf("type mismatch: %s vs %s", av.Type(), bv.Type()),
		Severity: "error",
		Detail: DifferenceDetail{
			Type:         "type_mismatch",
			ExpectedType: av.Type().String(),
			ActualType:   bv.Type().String(),
		},
		Suggestions: []string{
			"Check that both values have the same type",
			fmt.Sprintf("Consider converting %s to %s or vice versa", av.Type(), bv.Type()),
		},
	})
	c.stats.DifferentNodes++
}

func (c *defaultComparator) addMissingDiff(av, bv reflect.Value, path string) {
	var value any
	var message string

	if av.IsValid() {
		value = av.Interface()
		message = "value missing in second structure"
	} else {
		value = bv.Interface()
		message = "value missing in first structure"
	}

	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  message,
		Severity: "error",
		Detail: DifferenceDetail{
			Type:          "missing",
			ExpectedValue: value,
		},
	})
	c.stats.DifferentNodes++
}

func (c *defaultComparator) addLengthDiff(av, bv reflect.Value, path string) {
	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  fmt.Sprintf("length mismatch: %d vs %d", av.Len(), bv.Len()),
		Severity: "error",
		Detail: DifferenceDetail{
			Type:          "length_mismatch",
			ExpectedValue: av.Len(),
			ActualValue:   bv.Len(),
		},
		Suggestions: []string{
			"Check if elements were added or removed",
			"Verify the collection sizes match",
		},
	})
}

func (c *defaultComparator) addExtraElementDiff(v reflect.Value, path, message string) {
	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  message,
		Severity: "warning",
		Detail: DifferenceDetail{
			Type:        "extra_element",
			ActualValue: v.Interface(),
		},
	})
}

func (c *defaultComparator) addMissingElementDiff(v reflect.Value, path, message string) {
	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  message,
		Severity: "warning",
		Detail: DifferenceDetail{
			Type:          "missing_element",
			ExpectedValue: v.Interface(),
		},
	})
}

func (c *defaultComparator) addMissingKeyDiff(v reflect.Value, path, message string) {
	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  message,
		Severity: "error",
		Detail: DifferenceDetail{
			Type:          "missing_key",
			ExpectedValue: v.Interface(),
		},
	})
}

func (c *defaultComparator) addExtraKeyDiff(v reflect.Value, path, message string) {
	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  message,
		Severity: "warning",
		Detail: DifferenceDetail{
			Type:        "extra_key",
			ActualValue: v.Interface(),
		},
	})
}

func (c *defaultComparator) addNilDiff(av, bv reflect.Value, path string) {
	var expected, actual any
	var message string

	if av.IsNil() {
		expected = nil
		actual = bv.Elem().Interface()
		message = "expected nil, got value"
	} else {
		expected = av.Elem().Interface()
		actual = nil
		message = "expected value, got nil"
	}

	c.differences = append(c.differences, Difference{
		Path:     path,
		Level:    c.currentLevel,
		Message:  message,
		Severity: "error",
		Detail: DifferenceDetail{
			Type:          "nil_mismatch",
			ExpectedValue: expected,
			ActualValue:   actual,
		},
		Suggestions: []string{
			"Check for nil pointer dereferences",
			"Ensure all pointers are properly initialized",
		},
	})
}

func (c *defaultComparator) generateSuggestions(av, bv reflect.Value) []string {
	suggestions := make([]string, 0)

	if av.Kind() == bv.Kind() {
		switch av.Kind() {
		case reflect.String:
			if strings.Contains(av.String(), bv.String()) || strings.Contains(bv.String(), av.String()) {
				suggestions = append(suggestions, "Strings may be substrings of each other")
			}
			if strings.EqualFold(av.String(), bv.String()) {
				suggestions = append(suggestions, "Strings differ only in case (consider case-insensitive comparison)")
			}

		case reflect.Float32, reflect.Float64:
			diff := av.Float() - bv.Float()
			if diff < 0 {
				diff = -diff
			}
			suggestions = append(suggestions,
				fmt.Sprintf("Difference is %.6f (consider increasing float precision)", diff))

		case reflect.Slice:
			if av.Len() == bv.Len() {
				suggestions = append(suggestions,
					"Slices have same length but different elements")
			}
		}
	}

	suggestions = append(suggestions,
		"Check for off-by-one errors",
		"Verify data sources are synchronized",
		"Consider if this difference is expected",
	)

	return suggestions
}

func (c *defaultComparator) generateSummary() string {
	if len(c.differences) == 0 {
		return "No differences found"
	}

	errorCount := 0
	warningCount := 0
	for _, diff := range c.differences {
		if diff.Severity == "error" {
			errorCount++
		} else {
			warningCount++
		}
	}

	return fmt.Sprintf("Found %d differences (%d errors, %d warnings) across %d nodes",
		len(c.differences), errorCount, warningCount, c.stats.TotalNodes)
}

// ==================== JSON Patch Conversion ====================

func (c *defaultComparator) diffToJSONPatch(diff Difference) *JSONPatchOperation {
	path := c.pathToJSONPointer(diff.Path)

	switch diff.Detail.Type {
	case "value_different", "type_mismatch":
		return &JSONPatchOperation{
			Op:    "replace",
			Path:  path,
			Value: diff.Detail.ActualValue,
		}
	case "missing":
		return &JSONPatchOperation{
			Op:    "add",
			Path:  path,
			Value: diff.Detail.ExpectedValue,
		}
	case "extra_element", "extra_key":
		return &JSONPatchOperation{
			Op:   "remove",
			Path: path,
		}
	case "missing_element", "missing_key":
		return &JSONPatchOperation{
			Op:    "add",
			Path:  path,
			Value: diff.Detail.ExpectedValue,
		}
	}

	return nil
}

func (c *defaultComparator) pathToJSONPointer(path string) string {
	if path == "" {
		return "/"
	}

	path = strings.ReplaceAll(path, ".", "/")
	path = strings.ReplaceAll(path, "[", "/")
	path = strings.ReplaceAll(path, "]", "")

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}

// ==================== Visual Tree Building ====================

func (c *defaultComparator) buildVisualTree(diffs []Difference, node *VisualNode) {
	childrenMap := make(map[string][]Difference)

	for _, diff := range diffs {
		parts := strings.Split(diff.Path, ".")
		if len(parts) > 1 {
			parentPath := strings.Join(parts[:len(parts)-1], ".")
			childrenMap[parentPath] = append(childrenMap[parentPath], diff)
		} else {
			childrenMap["/"] = append(childrenMap["/"], diff)
		}
	}

	for path, childDiffs := range childrenMap {
		if path == "/" || path == node.Path {
			for _, diff := range childDiffs {
				childNode := &VisualNode{
					Path:   diff.Path,
					Value:  fmt.Sprintf("%v", diff.Detail.ActualValue),
					Status: c.getVisualStatus(diff),
				}
				node.Children = append(node.Children, childNode)

				c.buildVisualTree(diffs, childNode)
			}
		}
	}
}

func (c *defaultComparator) getVisualStatus(diff Difference) string {
	switch diff.Severity {
	case "error":
		return "modified"
	case "warning":
		return "added"
	default:
		return "same"
	}
}

// ==================== Formatting ====================

func (c *defaultComparator) formatTextDiff(result *DiffResult) string {
	var sb strings.Builder

	colorize := func(text, color string) string {
		if c.config.colorize {
			return color + text + colorReset
		}
		return text
	}

	sb.WriteString(colorize("=== Comparison Result ===", colorBold) + "\n")
	fmt.Fprintf(&sb, "%s: %v\n", colorize("Equal", colorCyan), result.Equal)
	fmt.Fprintf(&sb, "%s: %s\n", colorize("Summary", colorCyan), result.Summary)
	fmt.Fprintf(&sb, "%s: %+v\n", colorize("Stats", colorCyan), result.PathStats)

	if len(result.Differences) > 0 {
		sb.WriteString("\n" + colorize("=== Differences ===", colorBold) + "\n")

		diffsByLevel := make(map[int][]Difference)
		maxLevel := 0
		for _, diff := range result.Differences {
			diffsByLevel[diff.Level] = append(diffsByLevel[diff.Level], diff)
			if diff.Level > maxLevel {
				maxLevel = diff.Level
			}
		}

		for level := 0; level <= maxLevel; level++ {
			if diffs, ok := diffsByLevel[level]; ok {
				fmt.Fprintf(&sb, "\n%s:\n", colorize(fmt.Sprintf("Level %d", level), colorCyan))
				for _, diff := range diffs {
					indent := strings.Repeat("  ", level)

					// Choose color based on severity and type
					var severityColor string
					switch diff.Severity {
					case "warning":
						severityColor = colorYellow
					case "info":
						severityColor = colorGreen
					default:
						severityColor = colorRed
					}

					fmt.Fprintf(&sb, "%s%s [%s]\n", indent,
						colorize(diff.Path, colorCyan),
						colorize(diff.Severity, severityColor))
					fmt.Fprintf(&sb, "%s  Message: %s\n", indent, diff.Message)

					if diff.Detail.ExpectedValue != nil || diff.Detail.ActualValue != nil {
						fmt.Fprintf(&sb, "%s  Expected: %s\n", indent,
							colorize(fmt.Sprintf("%v", diff.Detail.ExpectedValue), colorGreen))
						fmt.Fprintf(&sb, "%s  Actual:   %s\n", indent,
							colorize(fmt.Sprintf("%v", diff.Detail.ActualValue), colorRed))
					}

					if len(diff.Suggestions) > 0 {
						fmt.Fprintf(&sb, "%s  Suggestions:\n", indent)
						for _, suggestion := range diff.Suggestions {
							fmt.Fprintf(&sb, "%s    - %s\n", indent, suggestion)
						}
					}
					sb.WriteString("\n")
				}
			}
		}
	}

	return sb.String()
}

func (c *defaultComparator) formatJSONDiff(result *DiffResult) string {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"failed to marshal result: %v\"}", err)
	}
	return string(jsonBytes)
}

func (c *defaultComparator) formatMarkdownDiff(result *DiffResult) string {
	var sb strings.Builder

	sb.WriteString("# Comparison Report\n\n")
	fmt.Fprintf(&sb, "**Equal:** `%v`  \n", result.Equal)
	fmt.Fprintf(&sb, "**Summary:** %s  \n", result.Summary)
	sb.WriteString("**Stats:**  \n")
	fmt.Fprintf(&sb, "- Total nodes: %d  \n", result.PathStats.TotalNodes)
	fmt.Fprintf(&sb, "- Compared: %d  \n", result.PathStats.ComparedNodes)
	fmt.Fprintf(&sb, "- Different: %d  \n", result.PathStats.DifferentNodes)
	fmt.Fprintf(&sb, "- Ignored: %d  \n\n", result.PathStats.IgnoredNodes)

	if len(result.Differences) > 0 {
		sb.WriteString("## Differences\n\n")

		for _, diff := range result.Differences {
			severityIcon := "❌"
			if diff.Severity == "warning" {
				severityIcon = "⚠️"
			}

			fmt.Fprintf(&sb, "### %s %s\n\n", severityIcon, diff.Path)
			fmt.Fprintf(&sb, "**Message:** %s  \n", diff.Message)

			if diff.Detail.ExpectedValue != nil || diff.Detail.ActualValue != nil {
				sb.WriteString("| Expected | Actual |\n")
				sb.WriteString("|----------|--------|\n")
				fmt.Fprintf(&sb, "| `%v` | `%v` |\n",
					diff.Detail.ExpectedValue, diff.Detail.ActualValue)
				sb.WriteString("\n")
			}

			if len(diff.Suggestions) > 0 {
				sb.WriteString("**Suggestions:**  \n")
				for _, suggestion := range diff.Suggestions {
					fmt.Fprintf(&sb, "- %s  \n", suggestion)
				}
				sb.WriteString("\n")
			}

			sb.WriteString("---\n\n")
		}
	}

	return sb.String()
}

func (c *defaultComparator) formatHTMLDiff(result *DiffResult) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
    <title>Comparison Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .diff { border: 1px solid #ddd; padding: 10px; margin: 10px 0; }
        .error { background-color: #ffe6e6; border-left: 4px solid #ff3333; }
        .warning { background-color: #fff0cc; border-left: 4px solid #ffcc00; }
        .path { font-family: monospace; font-weight: bold; }
        .value { font-family: monospace; background-color: #f5f5f5; padding: 2px 4px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>Comparison Report</h1>
    <div class="summary">
        <p><strong>Equal:</strong> <span class="value">`)
	fmt.Fprintf(&sb, "%v", result.Equal)
	sb.WriteString(`</span></p>
        <p><strong>Summary:</strong> `)
	sb.WriteString(result.Summary)
	sb.WriteString(`</p>
    </div>`)

	if len(result.Differences) > 0 {
		sb.WriteString(`
    <h2>Differences</h2>`)

		for _, diff := range result.Differences {
			severityClass := "error"
			if diff.Severity == "warning" {
				severityClass = "warning"
			}

			fmt.Fprintf(&sb, `
    <div class="diff %s">
        <h3 class="path">%s</h3>
        <p><strong>Message:</strong> %s</p>`,
				severityClass, diff.Path, diff.Message)

			if diff.Detail.ExpectedValue != nil || diff.Detail.ActualValue != nil {
				sb.WriteString(`
        <table>
            <tr>
                <th>Expected</th>
                <th>Actual</th>
            </tr>
            <tr>
                <td><code>`)
				fmt.Fprintf(&sb, "%v", diff.Detail.ExpectedValue)
				sb.WriteString(`</code></td>
                <td><code>`)
				fmt.Fprintf(&sb, "%v", diff.Detail.ActualValue)
				sb.WriteString(`</code></td>
            </tr>
        </table>`)
			}

			if len(diff.Suggestions) > 0 {
				sb.WriteString(`
        <p><strong>Suggestions:</strong></p>
        <ul>`)
				for _, suggestion := range diff.Suggestions {
					fmt.Fprintf(&sb, `
            <li>%s</li>`, suggestion)
				}
				sb.WriteString(`
        </ul>`)
			}

			sb.WriteString(`
    </div>`)
		}
	}

	sb.WriteString(`
</body>
</html>`)

	return sb.String()
}

func isNaN(f float64) bool {
	return f != f
}

// Equal is a convenience function that creates a comparator and checks equality
// between two values. This is useful for one-off comparisons without explicitly
// creating a comparator instance.
//
// Example:
//
//	if comparator.Equal(user1, user2, comparator.IgnoreStructFields("ID")) {
//	    fmt.Println("Users are equal (ignoring ID)")
//	}
//
// For multiple comparisons with the same configuration, create a comparator
// instance using New() or NewWithOptions() for better performance.
func Equal(a, b any, opts ...Option) bool {
	comparator := NewWithOptions(opts...)
	return comparator.Equal(a, b)
}

// DeepEqual is an alias for Equal, emphasizing that the comparison is deep
// (recursive). It provides the same functionality as Equal.
//
// Example:
//
//	if comparator.DeepEqual(struct1, struct2) {
//	    fmt.Println("Structs are deeply equal")
//	}
func DeepEqual(a, b any, opts ...Option) bool {
	return Equal(a, b, opts...)
}

// CompareWithDiff is a convenience function that creates a diff comparator and
// performs a comprehensive comparison. It returns a detailed DiffResult containing
// all differences, statistics, and a summary.
//
// This is useful for one-off detailed comparisons without explicitly creating
// a comparator instance.
//
// Example:
//
//	result := comparator.CompareWithDiff(
//	    expectedConfig,
//	    actualConfig,
//	    comparator.WithOutputFormat("markdown"),
//	    comparator.WithMaxDiffs(50),
//	)
//
//	if !result.Equal {
//	    fmt.Println(result.Summary)
//	    for _, diff := range result.Differences {
//	        fmt.Printf("  %s: %s\n", diff.Path, diff.Message)
//	    }
//	}
//
// For multiple comparisons, create a DiffComparer instance using
// NewDiffComparer() for better performance.
func CompareWithDiff(a, b any, opts ...Option) *DiffResult {
	comparator := NewDiffComparer(opts...)
	return comparator.CompareWithDiff(a, b)
}

// GetJSONPatch is a convenience function that creates a diff comparator and
// generates a JSON Patch (RFC 6902) document describing the differences.
//
// This is useful for one-off patch generation without explicitly creating
// a comparator instance.
//
// Example:
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
// For multiple patch generations, create a DiffComparer instance using
// NewDiffComparer() for better performance.
func GetJSONPatch(a, b any, opts ...Option) ([]JSONPatchOperation, error) {
	comparator := NewDiffComparer(opts...)
	return comparator.GetJSONPatch(a, b)
}

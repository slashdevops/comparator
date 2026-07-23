package comparator

import (
	"encoding/json"
	"fmt"
	"time"
)

// Type definitions for examples
type exampleServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	SSL  bool   `json:"ssl"`
}

type exampleDatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type exampleRoute struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
	Handler string   `json:"handler"`
	Auth    bool     `json:"auth"`
}

type exampleCacheConfig struct {
	Type    string        `json:"type"`
	TTL     time.Duration `json:"ttl"`
	MaxSize int64         `json:"max_size"`
	Enabled bool          `json:"enabled"`
}

type exampleConfig struct {
	Version  string                `json:"version"`
	Server   exampleServerConfig   `json:"server"`
	Database exampleDatabaseConfig `json:"database"`
	Features map[string]bool       `json:"features"`
	Routes   []exampleRoute        `json:"routes"`
	Cache    exampleCacheConfig    `json:"cache"`
}

func Example_detailedDiff() {
	// Complex JSON-like structures

	// Original config
	config1 := exampleConfig{
		Version: "1.0.0",
		Server: exampleServerConfig{
			Host: "localhost",
			Port: 8080,
			SSL:  false,
		},
		Database: exampleDatabaseConfig{
			Host:     "db.local",
			Port:     5432,
			Name:     "appdb",
			Username: "admin",
			Password: "secret",
		},
		Features: map[string]bool{
			"auth":       true,
			"logging":    true,
			"debug":      false,
			"monitoring": true,
		},
		Routes: []exampleRoute{
			{
				Path:    "/api/users",
				Methods: []string{"GET", "POST"},
				Handler: "usersHandler",
				Auth:    true,
			},
			{
				Path:    "/api/products",
				Methods: []string{"GET"},
				Handler: "productsHandler",
				Auth:    false,
			},
		},
		Cache: exampleCacheConfig{
			Type:    "redis",
			TTL:     30 * time.Minute,
			MaxSize: 1024 * 1024 * 100, // 100MB
			Enabled: true,
		},
	}

	// Modified config
	config2 := exampleConfig{
		Version: "1.0.1", // Changed
		Server: exampleServerConfig{
			Host: "localhost",
			Port: 8080,
			SSL:  true, // Changed
		},
		Database: exampleDatabaseConfig{
			Host:     "db.local",
			Port:     5432,
			Name:     "appdb",
			Username: "admin",
			Password: "newsecret", // Changed
		},
		Features: map[string]bool{
			"auth":       true,
			"logging":    true,
			"debug":      true, // Changed
			"monitoring": true,
			"caching":    true, // Added
		},
		Routes: []exampleRoute{
			{
				Path:    "/api/users",
				Methods: []string{"GET", "POST", "PUT"}, // Changed
				Handler: "usersHandler",
				Auth:    true,
			},
			{
				Path:    "/api/products",
				Methods: []string{"GET"},
				Handler: "productsHandler",
				Auth:    true, // Changed
			},
			{
				Path:    "/api/orders", // Added
				Methods: []string{"GET", "POST"},
				Handler: "ordersHandler",
				Auth:    true,
			},
		},
		Cache: exampleCacheConfig{
			Type:    "redis",
			TTL:     30 * time.Minute,
			MaxSize: 1024 * 1024 * 500, // Changed
			Enabled: true,
		},
	}

	// Get detailed diff
	result := CompareWithDiff(config1, config2,
		IgnoreStructFields("Password"), // Ignore password changes
		IgnoreSliceOrder(),             // Ignore slice order in routes
		WithDiffMode(DiffModeFull),
		WithMaxDepth(10),
	)

	// Print formatted diff
	formatted, err := NewDiffComparer().FormatDiff(result, "text")
	if err != nil {
		fmt.Printf("Error formatting diff: %v\n", err)
		return
	}
	fmt.Println(formatted)

	// Or get JSON Patch
	patch, err := GetJSONPatch(config1, config2)
	if err != nil {
		fmt.Printf("Error generating patch: %v\n", err)
		return
	}
	fmt.Printf("\nJSON Patch operations needed: %d\n", len(patch))

	// Or get unified diff
	unified, err := NewDiffComparer().GetUnifiedDiff(config1, config2)
	if err != nil {
		fmt.Printf("Error generating unified diff: %v\n", err)
		return
	}
	fmt.Printf("\nUnified diff has %d chunks\n", len(unified.Chunks))
}

func Example_jsonDiff() {
	// Compare JSON strings directly
	json1 := `{
		"user": {
			"id": 12345,
			"name": "John Doe",
			"email": "john@example.com",
			"preferences": {
				"theme": "dark",
				"language": "en"
			},
			"roles": ["admin", "user"]
		}
	}`

	json2 := `{
		"user": {
			"id": 12345,
			"name": "Jane Doe",
			"email": "jane@example.com",
			"preferences": {
				"theme": "light",
				"language": "en"
			},
			"roles": ["user", "admin"],
			"active": true
		}
	}`

	var data1, data2 any
	if err := json.Unmarshal([]byte(json1), &data1); err != nil {
		fmt.Printf("Error unmarshaling json1: %v\n", err)
		return
	}
	if err := json.Unmarshal([]byte(json2), &data2); err != nil {
		fmt.Printf("Error unmarshaling json2: %v\n", err)
		return
	}

	result := CompareWithDiff(data1, data2,
		IgnoreSliceOrder(),
		WithDiffMode(DiffModeFull),
	)

	// Format as Markdown
	md, err := NewDiffComparer().FormatDiff(result, "markdown")
	if err != nil {
		fmt.Printf("Error formatting as markdown: %v\n", err)
		return
	}
	fmt.Println(md)

	// Or generate HTML report
	html, err := NewDiffComparer().FormatDiff(result, "html")
	if err != nil {
		fmt.Printf("Error formatting as HTML: %v\n", err)
		return
	}
	_ = html
	// Save to file: os.WriteFile("diff.html", []byte(html), 0644)
}

func Example_performanceWithDiff() {
	// Test with large structures
	type BigStruct struct {
		ID     int
		Data   map[string]any
		Nested []map[string][]int
	}

	// Generate large test data
	n := 1000
	big1 := &BigStruct{
		ID:     1,
		Data:   make(map[string]any),
		Nested: make([]map[string][]int, n),
	}

	big2 := &BigStruct{
		ID:     1,
		Data:   make(map[string]any),
		Nested: make([]map[string][]int, n),
	}

	for i := range n {
		key := fmt.Sprintf("key_%d", i)
		big1.Data[key] = map[string]any{
			"value": i,
			"array": make([]int, 10),
		}
		big2.Data[key] = map[string]any{
			"value": i,
			"array": make([]int, 10),
		}

		big1.Nested[i] = map[string][]int{
			"a": {i, i * 2, i * 3},
			"b": {i, i * 4, i * 5},
		}
		big2.Nested[i] = map[string][]int{
			"a": {i, i * 2, i * 3},
			"b": {i, i * 4, i * 5},
		}
	}

	// Make one change
	big2.Data["key_500"] = map[string]any{
		"value": 999, // Changed
		"array": make([]int, 10),
	}

	// Compare with diff limiting
	result := CompareWithDiff(big1, big2,
		WithMaxDepth(5),
		WithDiffMode(DiffModeFull),
		WithMaxDiffs(100), // Stop after 100 differences
	)

	fmt.Printf("Found %d differences (limited to first 100)\n", len(result.Differences))
	fmt.Printf("Total nodes examined: %d\n", result.PathStats.TotalNodes)
}

// Example_stringLists demonstrates comparing two lists of strings
// with various options like order-independent comparison.
func Example_stringLists() {
	// Basic string list comparison
	list1 := []string{"apple", "banana", "cherry", "date"}
	list2 := []string{"apple", "banana", "cherry", "date"}

	comp := New()
	if comp.Equal(list1, list2) {
		fmt.Println("Lists are equal")
	}

	// Lists with different order - strict comparison
	list3 := []string{"banana", "apple", "cherry", "date"}
	if !comp.Equal(list1, list3) {
		fmt.Println("Lists with different order are not equal (strict mode)")
	}

	// Lists with different order - order-independent comparison
	compIgnoreOrder := NewWithOptions(IgnoreSliceOrder())
	if compIgnoreOrder.Equal(list1, list3) {
		fmt.Println("Lists with different order are equal (ignore order mode)")
	}

	// Lists with differences
	list4 := []string{"apple", "banana", "cherry", "elderberry"}
	diffComp := New()
	diffs, err := diffComp.Diff(list1, list4)
	if err != nil {
		fmt.Printf("Error getting diff: %v\n", err)
		return
	}

	if len(diffs) > 0 {
		fmt.Printf("Found %d differences\n", len(diffs))
		for _, diff := range diffs {
			if diff.Path != "" { // Skip summary messages
				fmt.Printf("  - %s: %s\n", diff.Path, diff.Message)
			}
		}
	}

	// Empty and nil lists
	var nilList []string
	emptyList := []string{}
	compEquateEmpty := NewWithOptions(EquateEmpty())
	if compEquateEmpty.Equal(nilList, emptyList) {
		fmt.Println("Nil and empty lists are equal (with EquateEmpty option)")
	}

	// Output:
	// Lists are equal
	// Lists with different order are not equal (strict mode)
	// Lists with different order are equal (ignore order mode)
	// Found 2 differences
	//   - [3]: date != elderberry
	// Nil and empty lists are equal (with EquateEmpty option)
}

// Example_stringListsDiff demonstrates detailed diff generation
// for string lists showing what changed.
func Example_stringListsDiff() {
	// Original tags
	oldTags := []string{"go", "backend", "api", "rest", "database"}

	// Updated tags
	newTags := []string{"go", "backend", "api", "graphql", "microservices"}

	// Create a diff comparer
	comp := NewDiffComparer(WithOutputFormat("markdown"))

	// Get detailed differences
	result := comp.CompareWithDiff(oldTags, newTags)

	fmt.Printf("Equal: %v\n", result.Equal)
	fmt.Printf("Summary: %s\n\n", result.Summary)

	fmt.Println("Differences:")
	for _, diff := range result.Differences {
		fmt.Printf("  Path: %s\n", diff.Path)
		fmt.Printf("  Expected: %v\n", diff.Detail.ExpectedValue)
		fmt.Printf("  Actual: %v\n", diff.Detail.ActualValue)
		fmt.Println()
	}

	// Get JSON Patch for programmatic updates
	patch, err := comp.GetJSONPatch(oldTags, newTags)
	if err != nil {
		fmt.Printf("Error generating patch: %v\n", err)
		return
	}
	patchJSON, err := json.MarshalIndent(patch, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling patch: %v\n", err)
		return
	}
	fmt.Printf("JSON Patch:\n%s\n", string(patchJSON))

	// Output format will show:
	// Equal: false
	// Summary: Found X differences...
	// And detailed path-by-path changes
}

// Example_stringListsWithDuplicates shows how to compare lists
// that may contain duplicate values.
func Example_stringListsWithDuplicates() {
	list1 := []string{"apple", "banana", "apple", "cherry", "banana"}
	list2 := []string{"banana", "apple", "cherry", "apple", "banana"}

	// Strict order comparison
	comp := New()
	fmt.Printf("Equal (strict order): %v\n", comp.Equal(list1, list2))

	// Order-independent comparison (duplicates must still match)
	compIgnoreOrder := NewWithOptions(IgnoreSliceOrder())
	fmt.Printf("Equal (ignore order): %v\n", compIgnoreOrder.Equal(list1, list2))

	// Different number of duplicates
	list3 := []string{"apple", "banana", "cherry", "apple"}
	fmt.Printf("Equal with different duplicates: %v\n", compIgnoreOrder.Equal(list1, list3))

	// Output:
	// Equal (strict order): false
	// Equal (ignore order): true
	// Equal with different duplicates: false
}

// Example_colorizedOutput demonstrates using colored terminal output
// for better readability of differences.
func Example_colorizedOutput() {
	type APIConfig struct {
		Endpoint string
		Timeout  int
		Retries  int
	}

	config1 := APIConfig{
		Endpoint: "https://api.example.com",
		Timeout:  30,
		Retries:  3,
	}

	config2 := APIConfig{
		Endpoint: "https://api.example.com",
		Timeout:  60,
		Retries:  5,
	}

	// Create a comparator with colorized output enabled
	comp := NewDiffComparer(
		WithColorize(true),
		WithOutputFormat("text"),
	)

	result := comp.CompareWithDiff(config1, config2)
	formatted, err := comp.FormatDiff(result, "text")
	if err != nil {
		fmt.Printf("Error formatting diff: %v\n", err)
		return
	}

	// The output will contain ANSI color codes for:
	// - Cyan: paths and labels
	// - Red: removed/expected values
	// - Green: added/actual values
	// - Yellow: warnings
	fmt.Println(formatted)

	// Note: In a terminal, this would display with colors.
	// Without colorize, the output would be plain text.
}

// Example_includeEqualValues demonstrates including both equal
// and different values in the comparison report.
func Example_includeEqualValues() {
	type User struct {
		ID       int
		Username string
		Email    string
		Active   bool
	}

	user1 := User{
		ID:       123,
		Username: "john_doe",
		Email:    "john@example.com",
		Active:   true,
	}

	user2 := User{
		ID:       123,
		Username: "john_doe",
		Email:    "john.doe@example.com", // Changed
		Active:   true,
	}

	// Create comparator that includes equal values in the report
	comp := NewDiffComparer(
		WithIncludeEqual(true),
		WithOutputFormat("text"),
	)

	result := comp.CompareWithDiff(user1, user2)

	fmt.Printf("Total differences (including equal): %d\\n", len(result.Differences))

	// Count equal vs different
	equalCount := 0
	diffCount := 0
	for _, diff := range result.Differences {
		if diff.Detail.Type == "equal" {
			equalCount++
		} else {
			diffCount++
		}
	}

	fmt.Printf("Equal fields: %d\\n", equalCount)
	fmt.Printf("Different fields: %d\\n", diffCount)

	// This is useful for:
	// - Comprehensive audit trails
	// - Understanding what hasn't changed
	// - Debugging comparison logic
}

// Example_colorizedWithEqualValues demonstrates combining both features
// for a complete, colorized comparison report.
func Example_colorizedWithEqualValues() {
	type DatabaseConfig struct {
		Host     string
		Port     int
		Database string
		SSL      bool
	}

	oldConfig := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "myapp",
		SSL:      false,
	}

	newConfig := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "myapp_prod", // Changed
		SSL:      true,         // Changed
	}

	// Combine colorization and including equal values
	comp := NewDiffComparer(
		WithColorize(true),
		WithIncludeEqual(true),
		WithOutputFormat("text"),
	)

	result := comp.CompareWithDiff(oldConfig, newConfig)
	formatted, err := comp.FormatDiff(result, "text")
	if err != nil {
		fmt.Printf("Error formatting diff: %v\n", err)
		return
	}

	// Output will show:
	// - Equal fields (Host, Port) in green with "info" severity
	// - Different fields (Database, SSL) in red with "error" severity
	// - All paths highlighted in cyan
	fmt.Println(formatted)
}

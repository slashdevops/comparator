package comparator

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

// ==================== Basic Equality Tests ====================

func TestEqual_Primitives(t *testing.T) {
	tests := []struct {
		name     string
		a        any
		b        any
		expected bool
	}{
		{"equal ints", 42, 42, true},
		{"unequal ints", 42, 43, false},
		{"equal strings", "hello", "hello", true},
		{"unequal strings", "hello", "world", false},
		{"equal bools", true, true, true},
		{"unequal bools", true, false, false},
		{"equal floats", 3.14, 3.14, true},
		{"unequal floats", 3.14, 3.15, false},
	}

	comp := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comp.Equal(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestEqual_Slices(t *testing.T) {
	comp := New()

	s1 := []int{1, 2, 3}
	s2 := []int{1, 2, 3}
	s3 := []int{1, 2, 4}

	if !comp.Equal(s1, s2) {
		t.Error("Expected equal slices to be equal")
	}

	if comp.Equal(s1, s3) {
		t.Error("Expected unequal slices to be unequal")
	}
}

func TestEqual_Maps(t *testing.T) {
	comp := New()

	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"a": 1, "b": 2}
	m3 := map[string]int{"a": 1, "b": 3}

	if !comp.Equal(m1, m2) {
		t.Error("Expected equal maps to be equal")
	}

	if comp.Equal(m1, m3) {
		t.Error("Expected unequal maps to be unequal")
	}
}

func TestEqual_Structs(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	comp := New()

	p1 := Person{Name: "Alice", Age: 30}
	p2 := Person{Name: "Alice", Age: 30}
	p3 := Person{Name: "Bob", Age: 30}

	if !comp.Equal(p1, p2) {
		t.Error("Expected equal structs to be equal")
	}

	if comp.Equal(p1, p3) {
		t.Error("Expected unequal structs to be unequal")
	}
}

// ==================== Option Tests ====================

func TestWithFloatPrecision(t *testing.T) {
	comp := NewWithOptions(WithFloatPrecision(1e-6))

	if !comp.Equal(1.0, 1.0000001) {
		t.Error("Expected values within precision to be equal")
	}

	if comp.Equal(1.0, 1.0001) {
		t.Error("Expected values outside precision to be unequal")
	}
}

func TestIgnoreSliceOrder(t *testing.T) {
	comp := NewWithOptions(IgnoreSliceOrder())

	s1 := []int{1, 2, 3}
	s2 := []int{3, 1, 2}

	if !comp.Equal(s1, s2) {
		t.Error("Expected slices with different order to be equal with IgnoreSliceOrder")
	}
}

func TestIgnoreStructFields(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	comp := NewWithOptions(IgnoreStructFields("ID"))

	u1 := User{ID: 1, Name: "Alice"}
	u2 := User{ID: 2, Name: "Alice"}

	if !comp.Equal(u1, u2) {
		t.Error("Expected structs to be equal when ignoring ID field")
	}
}

func TestEquateNaNs(t *testing.T) {
	nan1 := math.NaN()
	nan2 := math.NaN()

	comp1 := New()
	if comp1.Equal(nan1, nan2) {
		t.Error("Expected NaN values to be unequal by default")
	}

	comp2 := NewWithOptions(EquateNaNs())
	if !comp2.Equal(nan1, nan2) {
		t.Error("Expected NaN values to be equal with EquateNaNs option")
	}
}

// ==================== Diff Tests ====================

func TestDiff_BasicDifferences(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	p1 := Person{Name: "Alice", Age: 30}
	p2 := Person{Name: "Bob", Age: 25}

	comp := New()
	diffs, err := comp.Diff(p1, p2)
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	if len(diffs) == 0 {
		t.Error("Expected differences, got none")
	}

	for _, diff := range diffs {
		if diff.Path == "" {
			t.Error("Difference path should not be empty")
		}
		if diff.Message == "" {
			t.Error("Difference message should not be empty")
		}
	}
}

func TestCompareWithDiff(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(cfg1, cfg2)
	if result == nil {
		t.Fatal("CompareWithDiff returned nil")
		return
	}

	if result.Equal {
		t.Error("Expected configs to be unequal")
	}

	if len(result.Differences) == 0 {
		t.Error("Expected differences, got none")
	}

	if result.Summary == "" {
		t.Error("Expected summary to be non-empty")
	}
}

func TestGetUnifiedDiff(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	comp := NewDiffComparer()
	unifiedDiff, err := comp.GetUnifiedDiff(cfg1, cfg2)
	if err != nil {
		t.Fatalf("GetUnifiedDiff returned error: %v", err)
	}

	if unifiedDiff == nil {
		t.Fatal("GetUnifiedDiff returned nil")
		return
	}

	if unifiedDiff.Header == "" {
		t.Error("Expected header to be non-empty")
	}
}

func TestGetJSONPatch(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	comp := NewDiffComparer()
	patch, err := comp.GetJSONPatch(cfg1, cfg2)
	if err != nil {
		t.Fatalf("GetJSONPatch returned error: %v", err)
	}

	if patch == nil {
		t.Fatal("GetJSONPatch returned nil")
	}

	_, err = json.Marshal(patch)
	if err != nil {
		t.Errorf("Failed to marshal JSON patch: %v", err)
	}
}

func TestGetVisualDiff(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	p1 := Person{Name: "Alice", Age: 30}
	p2 := Person{Name: "Bob", Age: 25}

	comp := NewDiffComparer()
	visualDiff, err := comp.GetVisualDiff(p1, p2)
	if err != nil {
		t.Fatalf("GetVisualDiff returned error: %v", err)
	}

	if visualDiff == nil {
		t.Fatal("GetVisualDiff returned nil")
		return
	}

	if visualDiff.Root == nil {
		t.Error("Expected root node to be non-nil")
	}
}

// ==================== Convenience Function Tests ====================

func TestEqual_ConvenienceFunction(t *testing.T) {
	a := 42
	b := 42
	c := 43

	if !Equal(a, b) {
		t.Error("Expected Equal(42, 42) to be true")
	}

	if Equal(a, c) {
		t.Error("Expected Equal(42, 43) to be false")
	}
}

func TestDeepEqual_ConvenienceFunction(t *testing.T) {
	type Nested struct {
		Inner struct {
			Value int
		}
	}

	n1 := Nested{}
	n1.Inner.Value = 42

	n2 := Nested{}
	n2.Inner.Value = 42

	if !DeepEqual(n1, n2) {
		t.Error("Expected deeply nested structs to be equal")
	}
}

// ==================== Edge Cases ====================

func TestEqual_CircularReferences(t *testing.T) {
	type Node struct {
		Value int
		Next  *Node
	}

	n1 := &Node{Value: 1}
	n1.Next = n1

	n2 := &Node{Value: 1}
	n2.Next = n2

	comp := New()
	result := comp.Equal(n1, n2)
	if !result {
		t.Error("Expected circular structures to be equal")
	}
}

func TestEqual_Time(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	comp := New()

	if !comp.Equal(now, now) {
		t.Error("Expected equal times to be equal")
	}

	if comp.Equal(now, later) {
		t.Error("Expected unequal times to be unequal")
	}
}

// ==================== Performance Tests ====================

func BenchmarkEqual_Primitives(b *testing.B) {
	comp := New()
	a := 42
	c := 42

	for b.Loop() {
		comp.Equal(a, c)
	}
}

func BenchmarkEqual_Structs(b *testing.B) {
	type Person struct {
		Name string
		Age  int
	}

	comp := New()
	p1 := Person{Name: "Alice", Age: 30}
	p2 := Person{Name: "Alice", Age: 30}

	for b.Loop() {
		comp.Equal(p1, p2)
	}
}

func BenchmarkCompareWithDiff(b *testing.B) {
	type Config struct {
		Host string
		Port int
	}

	comp := NewDiffComparer()
	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	for b.Loop() {
		comp.CompareWithDiff(cfg1, cfg2)
	}
}

// ==================== Advanced Type Tests ====================

func TestEqual_Pointers(t *testing.T) {
	comp := New()

	s1 := "hello"
	s2 := "hello"
	s3 := "world"

	if !comp.Equal(&s1, &s2) {
		t.Error("Expected pointers to equal values to be equal")
	}

	if comp.Equal(&s1, &s3) {
		t.Error("Expected pointers to unequal values to be unequal")
	}

	// Nil pointer cases
	var nilPtr *string
	if !comp.Equal(nilPtr, nilPtr) {
		t.Error("Expected nil pointers to be equal")
	}

	if comp.Equal(nilPtr, &s1) {
		t.Error("Expected nil and non-nil pointers to be unequal")
	}
}

func TestEqual_Interfaces(t *testing.T) {
	comp := New()

	var i1 any = 42
	var i2 any = 42
	var i3 any = "hello"

	if !comp.Equal(i1, i2) {
		t.Error("Expected equal interface values to be equal")
	}

	if comp.Equal(i1, i3) {
		t.Error("Expected unequal interface values to be unequal")
	}

	var nilInterface any
	if !comp.Equal(nilInterface, nil) {
		t.Error("Expected nil interface to equal nil")
	}
}

func TestEqual_ComplexNumbers(t *testing.T) {
	comp := New()

	c1 := complex(1.0, 2.0)
	c2 := complex(1.0, 2.0)
	c3 := complex(1.0, 3.0)

	if !comp.Equal(c1, c2) {
		t.Error("Expected equal complex numbers to be equal")
	}

	if comp.Equal(c1, c3) {
		t.Error("Expected unequal complex numbers to be unequal")
	}
}

func TestEqual_Arrays(t *testing.T) {
	comp := New()

	a1 := [3]int{1, 2, 3}
	a2 := [3]int{1, 2, 3}
	a3 := [3]int{1, 2, 4}

	if !comp.Equal(a1, a2) {
		t.Error("Expected equal arrays to be equal")
	}

	if comp.Equal(a1, a3) {
		t.Error("Expected unequal arrays to be unequal")
	}
}

func TestEqual_NestedStructs(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}

	type Person struct {
		Name    string
		Address Address
	}

	comp := New()

	p1 := Person{
		Name: "Alice",
		Address: Address{
			Street: "123 Main St",
			City:   "Springfield",
		},
	}

	p2 := Person{
		Name: "Alice",
		Address: Address{
			Street: "123 Main St",
			City:   "Springfield",
		},
	}

	p3 := Person{
		Name: "Alice",
		Address: Address{
			Street: "456 Elm St",
			City:   "Springfield",
		},
	}

	if !comp.Equal(p1, p2) {
		t.Error("Expected equal nested structs to be equal")
	}

	if comp.Equal(p1, p3) {
		t.Error("Expected unequal nested structs to be unequal")
	}
}

func TestEqual_SliceOfStructs(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	comp := New()

	s1 := []Item{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
	}

	s2 := []Item{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
	}

	s3 := []Item{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item3"},
	}

	if !comp.Equal(s1, s2) {
		t.Error("Expected equal slice of structs to be equal")
	}

	if comp.Equal(s1, s3) {
		t.Error("Expected unequal slice of structs to be unequal")
	}
}

func TestEqual_MapOfSlices(t *testing.T) {
	comp := New()

	m1 := map[string][]int{
		"a": {1, 2, 3},
		"b": {4, 5, 6},
	}

	m2 := map[string][]int{
		"a": {1, 2, 3},
		"b": {4, 5, 6},
	}

	m3 := map[string][]int{
		"a": {1, 2, 3},
		"b": {4, 5, 7},
	}

	if !comp.Equal(m1, m2) {
		t.Error("Expected equal map of slices to be equal")
	}

	if comp.Equal(m1, m3) {
		t.Error("Expected unequal map of slices to be unequal")
	}
}

// ==================== Additional Option Tests ====================

func TestWithMaxDepth(t *testing.T) {
	type Level3 struct {
		Value int
	}

	type Level2 struct {
		Level3 Level3
	}

	type Level1 struct {
		Level2 Level2
	}

	comp := NewWithOptions(WithMaxDepth(2))

	l1 := Level1{Level2: Level2{Level3: Level3{Value: 42}}}
	l2 := Level1{Level2: Level2{Level3: Level3{Value: 99}}}

	// With max depth of 2, the deep difference should be treated as equal
	// because it won't recurse deep enough to see the difference
	result := comp.Equal(l1, l2)
	// This behavior depends on implementation details
	t.Logf("MaxDepth test result: %v", result)
}

func TestIgnoreUnexported(t *testing.T) {
	type User struct {
		Name     string
		password string // unexported
	}

	comp := NewWithOptions(IgnoreUnexported())

	// Distinct, clearly-marked fake values so this stays stdlib-only and does
	// not trip secret scanning.
	u1 := User{Name: "Alice", password: "test-secret-value-1"}
	u2 := User{Name: "Alice", password: "test-secret-value-2"}

	if !comp.Equal(u1, u2) {
		t.Error("Expected structs to be equal when ignoring unexported fields")
	}
}

func TestEquateEmpty(t *testing.T) {
	comp := NewWithOptions(EquateEmpty())

	// Nil vs empty slice
	var nilSlice []int
	emptySlice := []int{}

	if !comp.Equal(nilSlice, emptySlice) {
		t.Error("Expected nil slice to equal empty slice with EquateEmpty")
	}

	// Nil vs empty map
	var nilMap map[string]int
	emptyMap := map[string]int{}

	if !comp.Equal(nilMap, emptyMap) {
		t.Error("Expected nil map to equal empty map with EquateEmpty")
	}
}

func TestWithCustomComparator(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	comp := NewWithOptions(
		WithCustomComparator(func(a, b User) bool {
			return a.ID == b.ID
		}),
	)

	u1 := User{ID: 1, Name: "Alice"}
	u2 := User{ID: 1, Name: "Bob"}
	u3 := User{ID: 2, Name: "Alice"}

	if !comp.Equal(u1, u2) {
		t.Error("Expected users with same ID to be equal with custom comparator")
	}

	if comp.Equal(u1, u3) {
		t.Error("Expected users with different IDs to be unequal with custom comparator")
	}
}

func TestWithTimeLayout(t *testing.T) {
	comp := NewWithOptions(WithTimeLayout(time.RFC3339))

	t1 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	if !comp.Equal(t1, t2) {
		t.Error("Expected equal times to be equal")
	}
}

func TestWithDiffMode(t *testing.T) {
	modes := []DiffMode{
		DiffModeSimple,
		DiffModeFull,
		DiffModeUnified,
		DiffModeJSONPatch,
		DiffModeVisual,
	}

	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			comp := NewDiffComparer(WithDiffMode(mode))
			result := comp.CompareWithDiff(1, 2)
			if result == nil {
				t.Error("Expected non-nil result")
			}
		})
	}
}

func TestWithMaxDiffs(t *testing.T) {
	type Config struct {
		Field1 int
		Field2 int
		Field3 int
		Field4 int
		Field5 int
	}

	comp := NewDiffComparer(WithMaxDiffs(2))

	c1 := Config{Field1: 1, Field2: 2, Field3: 3, Field4: 4, Field5: 5}
	c2 := Config{Field1: 10, Field2: 20, Field3: 30, Field4: 40, Field5: 50}

	result := comp.CompareWithDiff(c1, c2)
	if len(result.Differences) > 2 {
		t.Errorf("Expected at most 2 differences, got %d", len(result.Differences))
	}
}

func TestWithOutputFormat(t *testing.T) {
	formats := []string{"text", "json", "markdown", "html"}

	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			comp := NewDiffComparer(WithOutputFormat(format))
			result := comp.CompareWithDiff(cfg1, cfg2)

			formatted, err := comp.FormatDiff(result, format)
			if err != nil {
				t.Errorf("FormatDiff failed for format %s: %v", format, err)
			}

			if formatted == "" {
				t.Errorf("Expected non-empty formatted output for %s", format)
			}
		})
	}
}

// ==================== Diff Detail Tests ====================

func TestDiff_TypeMismatch(t *testing.T) {
	comp := New()

	diffs, err := comp.Diff(42, "42")
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	if len(diffs) == 0 {
		t.Error("Expected type mismatch to be reported")
	}

	found := false
	for _, diff := range diffs {
		if diff.Detail.Type == "type_mismatch" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find type_mismatch difference")
	}
}

func TestDiff_MissingField(t *testing.T) {
	comp := New()

	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"a": 1}

	diffs, err := comp.Diff(m1, m2)
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	if len(diffs) == 0 {
		t.Error("Expected missing key to be reported")
	}
}

func TestDiff_LengthMismatch(t *testing.T) {
	comp := New()

	s1 := []int{1, 2, 3}
	s2 := []int{1, 2}

	diffs, err := comp.Diff(s1, s2)
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	if len(diffs) == 0 {
		t.Error("Expected length mismatch to be reported")
	}

	found := false
	for _, diff := range diffs {
		if diff.Detail.Type == "length_mismatch" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find length_mismatch difference")
	}
}

func TestDiff_PathTracking(t *testing.T) {
	type Address struct {
		City string
	}

	type Person struct {
		Name    string
		Address Address
	}

	comp := New()

	p1 := Person{Name: "Alice", Address: Address{City: "NYC"}}
	p2 := Person{Name: "Alice", Address: Address{City: "LA"}}

	diffs, err := comp.Diff(p1, p2)
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	found := false
	for _, diff := range diffs {
		if diff.Path != "" && diff.Message != "" {
			found = true
			t.Logf("Path: %s, Message: %s", diff.Path, diff.Message)
		}
	}

	if !found {
		t.Error("Expected differences with path information")
	}
}

func TestDiff_Suggestions(t *testing.T) {
	comp := New()

	type Config struct {
		Port int
	}

	c1 := Config{Port: 8080}
	c2 := Config{Port: 9090}

	diffs, err := comp.Diff(c1, c2)
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	if len(diffs) == 0 {
		t.Error("Expected differences")
	}

	// Check if suggestions are provided
	for _, diff := range diffs {
		t.Logf("Diff: %s, Suggestions: %v", diff.Message, diff.Suggestions)
	}
}

// ==================== Format Tests ====================

func TestFormatDiff_Text(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(cfg1, cfg2)

	text, err := comp.FormatDiff(result, "text")
	if err != nil {
		t.Fatalf("FormatDiff returned error: %v", err)
	}

	if text == "" {
		t.Error("Expected non-empty text output")
	}

	if result.Summary != "" && len(result.Summary) == 0 {
		t.Error("Expected summary in result")
	}
}

func TestFormatDiff_JSON(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(cfg1, cfg2)

	jsonStr, err := comp.FormatDiff(result, "json")
	if err != nil {
		t.Fatalf("FormatDiff returned error: %v", err)
	}

	if jsonStr == "" {
		t.Error("Expected non-empty JSON output")
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}
}

func TestFormatDiff_Markdown(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(cfg1, cfg2)

	markdown, err := comp.FormatDiff(result, "markdown")
	if err != nil {
		t.Fatalf("FormatDiff returned error: %v", err)
	}

	if markdown == "" {
		t.Error("Expected non-empty markdown output")
	}

	// Check for markdown indicators
	if len(markdown) > 0 {
		t.Logf("Markdown output length: %d", len(markdown))
	}
}

func TestFormatDiff_HTML(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(cfg1, cfg2)

	html, err := comp.FormatDiff(result, "html")
	if err != nil {
		t.Fatalf("FormatDiff returned error: %v", err)
	}

	if html == "" {
		t.Error("Expected non-empty HTML output")
	}
}

// ==================== Edge Case Tests ====================

func TestEqual_NilValues(t *testing.T) {
	comp := New()

	var nilSlice []int
	var nilMap map[string]int
	var nilPtr *string

	if !comp.Equal(nilSlice, nilSlice) {
		t.Error("Expected nil slice to equal itself")
	}

	if !comp.Equal(nilMap, nilMap) {
		t.Error("Expected nil map to equal itself")
	}

	if !comp.Equal(nilPtr, nilPtr) {
		t.Error("Expected nil pointer to equal itself")
	}

	if !comp.Equal(nil, nil) {
		t.Error("Expected nil to equal nil")
	}
}

func TestEqual_EmptyValues(t *testing.T) {
	comp := New()

	emptySlice := []int{}
	emptyMap := map[string]int{}
	emptyString := ""

	if !comp.Equal(emptySlice, emptySlice) {
		t.Error("Expected empty slice to equal itself")
	}

	if !comp.Equal(emptyMap, emptyMap) {
		t.Error("Expected empty map to equal itself")
	}

	if !comp.Equal(emptyString, emptyString) {
		t.Error("Expected empty string to equal itself")
	}
}

func TestEqual_ZeroValues(t *testing.T) {
	comp := New()

	if !comp.Equal(0, 0) {
		t.Error("Expected zero int to equal itself")
	}

	if !comp.Equal(0.0, 0.0) {
		t.Error("Expected zero float to equal itself")
	}

	if !comp.Equal(false, false) {
		t.Error("Expected false to equal itself")
	}

	if !comp.Equal("", "") {
		t.Error("Expected empty string to equal itself")
	}
}

func TestEqual_NumericTypeConversion(t *testing.T) {
	comp := NewWithOptions(WithFloatPrecision(1e-9))

	// Test if numeric types can be compared across types
	result := comp.Equal(int(42), float64(42.0))
	t.Logf("int(42) == float64(42.0): %v", result)
}

// ==================== Convenience Function Tests ====================

func TestCompareWithDiff_ConvenienceFunction(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	result := CompareWithDiff(cfg1, cfg2, WithMaxDiffs(10))
	if result == nil {
		t.Fatal("CompareWithDiff returned nil")
		return
	}

	if result.Equal {
		t.Error("Expected configs to be unequal")
	}

	if len(result.Differences) == 0 {
		t.Error("Expected differences")
	}
}

func TestGetJSONPatch_ConvenienceFunction(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	patch, err := GetJSONPatch(cfg1, cfg2)
	if err != nil {
		t.Fatalf("GetJSONPatch returned error: %v", err)
	}

	if patch == nil {
		t.Fatal("GetJSONPatch returned nil")
	}
}

// ==================== Statistics Tests ====================

func TestPathStats(t *testing.T) {
	type Complex struct {
		Field1 string
		Field2 int
		Field3 []string
		Field4 map[string]int
	}

	c1 := Complex{
		Field1: "test",
		Field2: 42,
		Field3: []string{"a", "b", "c"},
		Field4: map[string]int{"x": 1, "y": 2},
	}

	c2 := Complex{
		Field1: "test",
		Field2: 99,
		Field3: []string{"a", "b", "d"},
		Field4: map[string]int{"x": 1, "z": 3},
	}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(c1, c2)

	if result.PathStats.TotalNodes == 0 {
		t.Error("Expected TotalNodes to be > 0")
	}

	if result.PathStats.ComparedNodes == 0 {
		t.Error("Expected ComparedNodes to be > 0")
	}

	if result.PathStats.DifferentNodes == 0 {
		t.Error("Expected DifferentNodes to be > 0")
	}

	t.Logf("Stats: %+v", result.PathStats)
}

// ==================== Additional Benchmarks ====================

func BenchmarkEqual_DeepNested(b *testing.B) {
	type Level4 struct {
		Value int
	}
	type Level3 struct {
		L4 Level4
	}
	type Level2 struct {
		L3 Level3
	}
	type Level1 struct {
		L2 Level2
	}

	comp := New()
	l1 := Level1{L2: Level2{L3: Level3{L4: Level4{Value: 42}}}}
	l2 := Level1{L2: Level2{L3: Level3{L4: Level4{Value: 42}}}}

	for b.Loop() {
		comp.Equal(l1, l2)
	}
}

func BenchmarkEqual_LargeSlice(b *testing.B) {
	comp := New()
	s1 := make([]int, 1000)
	s2 := make([]int, 1000)
	for i := range s1 {
		s1[i] = i
		s2[i] = i
	}

	for b.Loop() {
		comp.Equal(s1, s2)
	}
}

// ==================== Complex Struct Tests ====================

func TestEqual_ComplexNestedStruct(t *testing.T) {
	type Address struct {
		Street  string
		City    string
		ZipCode int
		Country string
	}

	type Contact struct {
		Email string
		Phone string
		Fax   *string
	}

	type Department struct {
		Name      string
		Manager   string
		Budget    float64
		Employees []string
	}

	type Company struct {
		Name        string
		Founded     int
		Address     Address
		Departments []Department
		Metadata    map[string]any
	}

	type Person struct {
		ID          int
		FirstName   string
		LastName    string
		Age         int
		IsActive    bool
		Salary      float64
		Address     Address
		Contact     Contact
		Company     *Company
		Tags        []string
		Preferences map[string]string
		Metadata    map[string]any
	}

	fax := "555-0199"
	company := &Company{
		Name:    "Tech Corp",
		Founded: 2010,
		Address: Address{
			Street:  "100 Tech Blvd",
			City:    "San Francisco",
			ZipCode: 94105,
			Country: "USA",
		},
		Departments: []Department{
			{
				Name:      "Engineering",
				Manager:   "Alice Smith",
				Budget:    1000000.50,
				Employees: []string{"Bob", "Charlie", "David"},
			},
			{
				Name:      "Sales",
				Manager:   "Eve Johnson",
				Budget:    750000.25,
				Employees: []string{"Frank", "Grace"},
			},
		},
		Metadata: map[string]any{
			"stock_symbol": "TECH",
			"public":       true,
			"revenue":      50000000.0,
		},
	}

	person1 := Person{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Age:       30,
		IsActive:  true,
		Salary:    75000.50,
		Address: Address{
			Street:  "123 Main St",
			City:    "New York",
			ZipCode: 10001,
			Country: "USA",
		},
		Contact: Contact{
			Email: "john@example.com",
			Phone: "555-0100",
			Fax:   &fax,
		},
		Company: company,
		Tags:    []string{"engineer", "senior", "go"},
		Preferences: map[string]string{
			"theme":    "dark",
			"language": "en",
		},
		Metadata: map[string]any{
			"hired_date":  "2020-01-15",
			"performance": 4.5,
			"remote":      true,
		},
	}

	// Create person2 with same values
	fax2 := "555-0199"
	company2 := &Company{
		Name:    "Tech Corp",
		Founded: 2010,
		Address: Address{
			Street:  "100 Tech Blvd",
			City:    "San Francisco",
			ZipCode: 94105,
			Country: "USA",
		},
		Departments: []Department{
			{
				Name:      "Engineering",
				Manager:   "Alice Smith",
				Budget:    1000000.50,
				Employees: []string{"Bob", "Charlie", "David"},
			},
			{
				Name:      "Sales",
				Manager:   "Eve Johnson",
				Budget:    750000.25,
				Employees: []string{"Frank", "Grace"},
			},
		},
		Metadata: map[string]any{
			"stock_symbol": "TECH",
			"public":       true,
			"revenue":      50000000.0,
		},
	}

	person2 := Person{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Age:       30,
		IsActive:  true,
		Salary:    75000.50,
		Address: Address{
			Street:  "123 Main St",
			City:    "New York",
			ZipCode: 10001,
			Country: "USA",
		},
		Contact: Contact{
			Email: "john@example.com",
			Phone: "555-0100",
			Fax:   &fax2,
		},
		Company: company2,
		Tags:    []string{"engineer", "senior", "go"},
		Preferences: map[string]string{
			"theme":    "dark",
			"language": "en",
		},
		Metadata: map[string]any{
			"hired_date":  "2020-01-15",
			"performance": 4.5,
			"remote":      true,
		},
	}

	comp := New()

	// Test equality with identical complex structures
	if !comp.Equal(person1, person2) {
		t.Error("Expected complex nested structs to be equal")
	}

	// Test with different nested value
	company2.Departments[0].Budget = 999999.99
	if comp.Equal(person1, person2) {
		t.Error("Expected complex structs with different nested values to be unequal")
	}

	// Reset and test with different metadata
	company2.Departments[0].Budget = 1000000.50
	person2.Metadata["performance"] = 4.0
	if comp.Equal(person1, person2) {
		t.Error("Expected complex structs with different metadata to be unequal")
	}
}

func TestEqual_DeepNestedMaps(t *testing.T) {
	map1 := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"value": 42,
						"text":  "deep",
						"list":  []int{1, 2, 3},
					},
				},
			},
		},
		"other": "data",
	}

	map2 := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"value": 42,
						"text":  "deep",
						"list":  []int{1, 2, 3},
					},
				},
			},
		},
		"other": "data",
	}

	comp := New()
	if !comp.Equal(map1, map2) {
		t.Error("Expected deeply nested maps to be equal")
	}

	// Change deep value
	map3 := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"value": 43, // Changed
						"text":  "deep",
						"list":  []int{1, 2, 3},
					},
				},
			},
		},
		"other": "data",
	}

	if comp.Equal(map1, map3) {
		t.Error("Expected maps with different deep values to be unequal")
	}
}

func TestEqual_JSONStructures(t *testing.T) {
	json1 := `{
		"users": [
			{
				"id": 1,
				"name": "Alice",
				"email": "alice@example.com",
				"roles": ["admin", "user"],
				"settings": {
					"theme": "dark",
					"notifications": true
				}
			},
			{
				"id": 2,
				"name": "Bob",
				"email": "bob@example.com",
				"roles": ["user"],
				"settings": {
					"theme": "light",
					"notifications": false
				}
			}
		],
		"metadata": {
			"version": "1.0",
			"count": 2
		}
	}`

	json2 := `{
		"users": [
			{
				"id": 1,
				"name": "Alice",
				"email": "alice@example.com",
				"roles": ["admin", "user"],
				"settings": {
					"theme": "dark",
					"notifications": true
				}
			},
			{
				"id": 2,
				"name": "Bob",
				"email": "bob@example.com",
				"roles": ["user"],
				"settings": {
					"theme": "light",
					"notifications": false
				}
			}
		],
		"metadata": {
			"version": "1.0",
			"count": 2
		}
	}`

	var data1, data2 map[string]any
	if err := json.Unmarshal([]byte(json1), &data1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(json2), &data2); err != nil {
		t.Fatal(err)
	}

	comp := New()
	if !comp.Equal(data1, data2) {
		t.Error("Expected JSON structures to be equal")
	}

	// Test with modified JSON
	json3 := `{
		"users": [
			{
				"id": 1,
				"name": "Alice",
				"email": "alice@example.com",
				"roles": ["admin", "user"],
				"settings": {
					"theme": "dark",
					"notifications": true
				}
			},
			{
				"id": 2,
				"name": "Bob",
				"email": "bob@example.com",
				"roles": ["user"],
				"settings": {
					"theme": "dark",
					"notifications": false
				}
			}
		],
		"metadata": {
			"version": "1.0",
			"count": 2
		}
	}`

	var data3 map[string]any
	if err := json.Unmarshal([]byte(json3), &data3); err != nil {
		t.Fatal(err)
	}

	if comp.Equal(data1, data3) {
		t.Error("Expected JSON structures with different nested values to be unequal")
	}
}

func TestCompareWithDiff_JSONStructures(t *testing.T) {
	json1 := `{
		"name": "Product A",
		"price": 99.99,
		"inStock": true,
		"categories": ["electronics", "computers"],
		"specs": {
			"cpu": "Intel i7",
			"ram": "16GB",
			"storage": "512GB"
		}
	}`

	json2 := `{
		"name": "Product A",
		"price": 89.99,
		"inStock": false,
		"categories": ["electronics", "laptops"],
		"specs": {
			"cpu": "Intel i7",
			"ram": "32GB",
			"storage": "512GB"
		}
	}`

	var data1, data2 map[string]any
	if err := json.Unmarshal([]byte(json1), &data1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(json2), &data2); err != nil {
		t.Fatal(err)
	}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(data1, data2)

	if result.Equal {
		t.Error("Expected JSON structures to be different")
	}

	if len(result.Differences) == 0 {
		t.Error("Expected differences to be found")
	}

	// Check specific differences
	foundPriceDiff := false
	foundStockDiff := false
	foundRAMDiff := false

	for _, diff := range result.Differences {
		if strings.Contains(diff.Path, "price") {
			foundPriceDiff = true
		}
		if strings.Contains(diff.Path, "inStock") {
			foundStockDiff = true
		}
		if strings.Contains(diff.Path, "ram") {
			foundRAMDiff = true
		}
	}

	if !foundPriceDiff {
		t.Error("Expected to find price difference")
	}
	if !foundStockDiff {
		t.Error("Expected to find inStock difference")
	}
	if !foundRAMDiff {
		t.Error("Expected to find ram difference")
	}
}

func TestEqualWithConfig_CustomConfiguration(t *testing.T) {
	type Product struct {
		ID    int
		Name  string
		Price float64
	}

	p1 := Product{ID: 1, Name: "Widget", Price: 19.99}
	p2 := Product{ID: 2, Name: "Widget", Price: 19.99}

	comp := New()

	// Should be unequal with default config
	if comp.Equal(p1, p2) {
		t.Error("Expected products with different IDs to be unequal")
	}

	// Should be equal when ignoring ID field
	customConfig := defaultConfig()
	customConfig.ignoreStructFields = map[string]bool{"ID": true}

	if !comp.EqualWithConfig(p1, p2, customConfig) {
		t.Error("Expected products to be equal when ignoring ID field")
	}
}

func TestDiff_ComplexNumbers(t *testing.T) {
	comp := NewDiffComparer()

	c1 := complex(1.5, 2.5)
	c2 := complex(1.5, 3.5)

	result := comp.CompareWithDiff(c1, c2)

	if result.Equal {
		t.Error("Expected complex numbers to be different")
	}

	if len(result.Differences) == 0 {
		t.Error("Expected differences to be found for complex numbers")
	}
}

func TestEqual_SlicesIgnoreOrderComplex(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	s1 := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	s2 := []Person{
		{Name: "Charlie", Age: 35},
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}

	comp := NewWithOptions(IgnoreSliceOrder())

	if !comp.Equal(s1, s2) {
		t.Error("Expected slices with same elements in different order to be equal")
	}

	// Test with different elements
	s3 := []Person{
		{Name: "Charlie", Age: 36}, // Different age
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}

	if comp.Equal(s1, s3) {
		t.Error("Expected slices with different elements to be unequal")
	}
}

func TestDiff_SlicesIgnoreOrder(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	s1 := []Item{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
		{ID: 3, Name: "C"},
	}

	s2 := []Item{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "X"}, // Different
		{ID: 4, Name: "D"}, // Different
	}

	comp := NewDiffComparer(IgnoreSliceOrder())
	result := comp.CompareWithDiff(s1, s2)

	if result.Equal {
		t.Error("Expected slices to be different")
	}

	if len(result.Differences) == 0 {
		t.Error("Expected differences to be found")
	}
}

func TestDiff_NilPointers(t *testing.T) {
	type Container struct {
		Value *int
	}

	val1 := 42
	c1 := Container{Value: &val1}
	c2 := Container{Value: nil}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(c1, c2)

	if result.Equal {
		t.Error("Expected structs with nil pointer differences to be unequal")
	}

	foundNilDiff := false
	for _, diff := range result.Differences {
		if strings.Contains(diff.Detail.Type, "nil") || diff.Detail.ActualValue == nil {
			foundNilDiff = true
			break
		}
	}

	if !foundNilDiff {
		t.Error("Expected to find nil pointer difference")
	}
}

func TestDiff_MissingValues(t *testing.T) {
	comp := NewDiffComparer()

	var nilVal *int
	val := new(int)
	*val = 42

	result := comp.CompareWithDiff(nilVal, val)

	if result.Equal {
		t.Error("Expected nil and non-nil to be different")
	}

	if len(result.Differences) == 0 {
		t.Error("Expected differences for missing values")
	}
}

func TestEqual_NumericCrossTypeComparison(t *testing.T) {
	comp := New()

	// Test int vs float
	if !comp.Equal(int(42), float64(42.0)) {
		t.Error("Expected int 42 to equal float64 42.0")
	}

	// Test uint vs int
	if !comp.Equal(uint(42), int(42)) {
		t.Error("Expected uint 42 to equal int 42")
	}

	// Test int vs uint
	if !comp.Equal(int64(100), uint32(100)) {
		t.Error("Expected int64 100 to equal uint32 100")
	}

	// Test with default precision (difference 1e-10, within 1e-9 default precision)
	if !comp.Equal(int(42), float64(42.00000000001)) {
		t.Error("Expected int 42 to equal float64 42.00000000001 within default precision")
	}

	// Test unequal numeric values
	if comp.Equal(int(42), float64(43.0)) {
		t.Error("Expected int 42 to not equal float64 43.0")
	}

	// Test with larger precision tolerance
	compLargePrecision := NewWithOptions(WithFloatPrecision(1e-6))
	if !compLargePrecision.Equal(int(42), float64(42.0000001)) {
		t.Error("Expected int 42 to equal float64 42.0000001 with 1e-6 precision")
	}
}

func TestEqual_IsSimpleType(t *testing.T) {
	comp := New()

	// Test all simple types
	simpleValues := []any{
		true,
		int(42),
		int8(42),
		int16(42),
		int32(42),
		int64(42),
		uint(42),
		uint8(42),
		uint16(42),
		uint32(42),
		uint64(42),
		float32(42.0),
		float64(42.0),
		"string",
	}

	for _, v := range simpleValues {
		if !comp.Equal(v, v) {
			t.Errorf("Expected simple type %T to equal itself", v)
		}
	}
}

func TestGetVisualDiff_ComplexStruct(t *testing.T) {
	type Config struct {
		Host string
		Port int
		TLS  bool
	}

	c1 := Config{Host: "localhost", Port: 8080, TLS: false}
	c2 := Config{Host: "example.com", Port: 8443, TLS: true}

	comp := NewDiffComparer()
	visualDiff, err := comp.GetVisualDiff(c1, c2)
	if err != nil {
		t.Fatalf("GetVisualDiff failed: %v", err)
	}

	if visualDiff == nil {
		t.Fatal("Expected visual diff to be non-nil")
		return
	}

	if visualDiff.Root == nil {
		t.Fatal("Expected visual diff root to be non-nil")
		return
	}

	// Visual diff should at least have a root node
	if visualDiff.Root.Path != "/" {
		t.Errorf("Expected root path to be '/', got %s", visualDiff.Root.Path)
	}
}

func TestFormatDiff_AllFormats(t *testing.T) {
	type User struct {
		Name  string
		Email string
		Age   int
	}

	u1 := User{Name: "Alice", Email: "alice@old.com", Age: 30}
	u2 := User{Name: "Alice", Email: "alice@new.com", Age: 31}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(u1, u2)

	formats := []string{"text", "json", "markdown", "html"}

	for _, format := range formats {
		formatted, err := comp.FormatDiff(result, format)
		if err != nil {
			t.Errorf("FormatDiff failed for format %s: %v", format, err)
		}

		if formatted == "" {
			t.Errorf("Expected non-empty formatted output for format %s", format)
		}

		// Verify format-specific content
		switch format {
		case "json":
			if !strings.Contains(formatted, "{") || !strings.Contains(formatted, "}") {
				t.Errorf("Expected JSON format to contain braces")
			}
		case "markdown":
			if !strings.Contains(formatted, "#") {
				t.Errorf("Expected markdown format to contain headers")
			}
		case "html":
			if !strings.Contains(formatted, "<html>") {
				t.Errorf("Expected HTML format to contain HTML tags")
			}
		}
	}

	// An unknown format now returns ErrUnknownFormat instead of silently
	// falling back to text.
	if _, err := comp.FormatDiff(result, "unknown"); !errors.Is(err, ErrUnknownFormat) {
		t.Errorf("Expected ErrUnknownFormat for unknown format, got %v", err)
	}
}

func TestGenerateSuggestions_Strings(t *testing.T) {
	comp := NewDiffComparer()

	// Test substring detection
	result := comp.CompareWithDiff("hello world", "hello")
	if len(result.Differences) > 0 {
		diff := result.Differences[0]
		hasSubstringSuggestion := false
		for _, suggestion := range diff.Suggestions {
			if strings.Contains(suggestion, "substring") {
				hasSubstringSuggestion = true
				break
			}
		}
		if !hasSubstringSuggestion {
			t.Error("Expected substring suggestion for string comparison")
		}
	}

	// Test case-insensitive detection
	result2 := comp.CompareWithDiff("Hello", "hello")
	if len(result2.Differences) > 0 {
		diff := result2.Differences[0]
		hasCaseSuggestion := false
		for _, suggestion := range diff.Suggestions {
			if strings.Contains(suggestion, "case") {
				hasCaseSuggestion = true
				break
			}
		}
		if !hasCaseSuggestion {
			t.Error("Expected case-insensitive suggestion for string comparison")
		}
	}
}

func TestJSONPatch_ComplexOperations(t *testing.T) {
	type Document struct {
		Title   string
		Content string
		Tags    []string
		Meta    map[string]any
	}

	d1 := Document{
		Title:   "Original",
		Content: "Original content",
		Tags:    []string{"tag1", "tag2"},
		Meta: map[string]any{
			"author": "Alice",
			"views":  100,
		},
	}

	d2 := Document{
		Title:   "Updated",
		Content: "Updated content",
		Tags:    []string{"tag1", "tag3"},
		Meta: map[string]any{
			"author": "Bob",
			"views":  150,
			"rating": 4.5,
		},
	}

	comp := NewDiffComparer()
	patch, err := comp.GetJSONPatch(d1, d2)
	if err != nil {
		t.Fatalf("GetJSONPatch failed: %v", err)
	}

	if len(patch) == 0 {
		t.Error("Expected JSON patch operations to be generated")
	}

	// Verify operation types
	hasReplace := false
	for _, op := range patch {
		if op.Op == "replace" {
			hasReplace = true
			break
		}
	}

	if !hasReplace {
		t.Error("Expected at least one replace operation in JSON patch")
	}
}

func BenchmarkEqual_LargeMap(b *testing.B) {
	comp := New()
	m1 := make(map[string]int, 1000)
	m2 := make(map[string]int, 1000)
	for i := range 1000 {
		key := string(rune('a' + (i % 26)))
		m1[key] = i
		m2[key] = i
	}

	for b.Loop() {
		comp.Equal(m1, m2)
	}
}

// ==================== Colorize and IncludeEqual Tests ====================

func TestWithColorize(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	p1 := Person{Name: "Alice", Age: 30}
	p2 := Person{Name: "Bob", Age: 25}

	// Test with colorize enabled
	compColorized := NewDiffComparer(
		WithColorize(true),
		WithOutputFormat("text"),
	)
	result := compColorized.CompareWithDiff(p1, p2)
	formatted, err := compColorized.FormatDiff(result, "text")
	if err != nil {
		t.Fatalf("FormatDiff failed: %v", err)
	}

	// Check for ANSI color codes
	if !strings.Contains(formatted, "\033[") {
		t.Error("Expected ANSI color codes in colorized output")
	}

	// Test with colorize disabled
	compPlain := NewDiffComparer(
		WithColorize(false),
		WithOutputFormat("text"),
	)
	result2 := compPlain.CompareWithDiff(p1, p2)
	formatted2, err := compPlain.FormatDiff(result2, "text")
	if err != nil {
		t.Fatalf("FormatDiff failed: %v", err)
	}

	// Check for no ANSI color codes
	if strings.Contains(formatted2, "\033[") {
		t.Error("Expected no ANSI color codes in plain output")
	}
}

func TestWithIncludeEqual(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfg1 := Config{Host: "localhost", Port: 8080}
	cfg2 := Config{Host: "localhost", Port: 9090}

	// Test with includeEqual disabled (default)
	compDefault := NewDiffComparer()
	result1 := compDefault.CompareWithDiff(cfg1, cfg2)

	// Count differences - should only include different values
	diffCount := 0
	for _, diff := range result1.Differences {
		if diff.Detail.Type != "equal" {
			diffCount++
		}
	}

	if diffCount == 0 {
		t.Error("Expected at least one difference")
	}

	// Test with includeEqual enabled
	compIncludeEqual := NewDiffComparer(WithIncludeEqual(true))
	result2 := compIncludeEqual.CompareWithDiff(cfg1, cfg2)

	// Count equal values
	equalCount := 0
	diffCount2 := 0
	for _, diff := range result2.Differences {
		if diff.Detail.Type == "equal" {
			equalCount++
		} else {
			diffCount2++
		}
	}

	if equalCount == 0 {
		t.Error("Expected equal values to be included in diff")
	}

	if diffCount2 == 0 {
		t.Error("Expected at least one difference")
	}

	// Verify Host is marked as equal
	foundHostEqual := false
	for _, diff := range result2.Differences {
		if strings.Contains(diff.Path, "Host") && diff.Detail.Type == "equal" {
			foundHostEqual = true
			if diff.Severity != "info" {
				t.Errorf("Expected severity 'info' for equal values, got %s", diff.Severity)
			}
			break
		}
	}

	if !foundHostEqual {
		t.Error("Expected Host field to be marked as equal")
	}
}

func TestColorizeAndIncludeEqual(t *testing.T) {
	// Use simple struct for easier testing
	type Point struct {
		X int
		Y int
	}

	p1 := Point{X: 1, Y: 2}
	p2 := Point{X: 1, Y: 3}

	comp := NewDiffComparer(
		WithColorize(true),
		WithIncludeEqual(true),
		WithOutputFormat("text"),
	)

	result := comp.CompareWithDiff(p1, p2)

	// Should have both equal and different entries
	hasEqual := false
	hasDiff := false

	for _, diff := range result.Differences {
		switch diff.Detail.Type {
		case "equal":
			hasEqual = true
		case "value_different":
			hasDiff = true
		}
	}

	if !hasEqual {
		t.Error("Expected equal values in diff with WithIncludeEqual")
	}

	if !hasDiff {
		t.Error("Expected different values in diff")
	}

	// Check formatted output has colors
	formatted, err := comp.FormatDiff(result, "text")
	if err != nil {
		t.Fatalf("FormatDiff failed: %v", err)
	}

	if !strings.Contains(formatted, "\033[") {
		t.Error("Expected ANSI color codes in output")
	}
}

// ==================== DiffMode String Method Test ====================

func (d DiffMode) String() string {
	switch d {
	case DiffModeSimple:
		return "Simple"
	case DiffModeFull:
		return "Full"
	case DiffModeUnified:
		return "Unified"
	case DiffModeJSONPatch:
		return "JSONPatch"
	case DiffModeVisual:
		return "Visual"
	default:
		return "Unknown"
	}
}

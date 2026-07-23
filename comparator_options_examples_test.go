package comparator

import (
	"fmt"
	"math"
	"time"
)

// ExampleEqual demonstrates the package-level Equal convenience function for a
// quick deep comparison without constructing a Comparator.
func ExampleEqual() {
	a := map[string]int{"cpu": 2, "mem": 8}
	b := map[string]int{"cpu": 2, "mem": 8}

	fmt.Println(Equal(a, b))
	fmt.Println(Equal(a, map[string]int{"cpu": 4, "mem": 8}))
	// Output:
	// true
	// false
}

// ExampleDeepEqual demonstrates DeepEqual, which behaves like Equal but keeps a
// name that is familiar to users of reflect.DeepEqual.
func ExampleDeepEqual() {
	type point struct{ X, Y int }

	fmt.Println(DeepEqual(point{1, 2}, point{1, 2}))
	fmt.Println(DeepEqual(point{1, 2}, point{1, 3}))
	// Output:
	// true
	// false
}

// ExampleEqual_options shows how options can be passed to the package-level
// helpers to relax the comparison.
func ExampleEqual_options() {
	left := []int{3, 1, 2}
	right := []int{1, 2, 3}

	fmt.Println(Equal(left, right))
	fmt.Println(Equal(left, right, IgnoreSliceOrder()))
	// Output:
	// false
	// true
}

// ExampleWithFloatPrecision shows how to treat floating-point values that are
// close enough as equal.
func ExampleWithFloatPrecision() {
	comp := NewWithOptions(WithFloatPrecision(1e-3))

	fmt.Println(comp.Equal(0.1+0.2, 0.3))
	// Output:
	// true
}

// ExampleIgnoreStructFields skips volatile fields such as IDs and timestamps so
// that two otherwise-identical records compare as equal.
func ExampleIgnoreStructFields() {
	type user struct {
		ID        int
		Name      string
		UpdatedAt string
	}

	comp := NewWithOptions(IgnoreStructFields("ID", "UpdatedAt"))

	u1 := user{ID: 1, Name: "Alice", UpdatedAt: "2026-01-01"}
	u2 := user{ID: 2, Name: "Alice", UpdatedAt: "2026-07-23"}

	fmt.Println(comp.Equal(u1, u2))
	// Output:
	// true
}

// ExampleIgnoreUnexported ignores unexported struct fields during comparison.
func ExampleIgnoreUnexported() {
	type credential struct {
		Name  string
		token string // unexported
	}

	comp := NewWithOptions(IgnoreUnexported())

	c1 := credential{Name: "svc", token: "aaa"}
	c2 := credential{Name: "svc", token: "bbb"}

	fmt.Println(comp.Equal(c1, c2))
	// Output:
	// true
}

// ExampleEquateEmpty treats nil and empty containers as equal.
func ExampleEquateEmpty() {
	comp := NewWithOptions(EquateEmpty())

	var nilSlice []string
	emptySlice := []string{}

	fmt.Println(comp.Equal(nilSlice, emptySlice))
	// Output:
	// true
}

// ExampleEquateNaNs treats NaN values as equal, which the IEEE-754 standard does
// not do by default.
func ExampleEquateNaNs() {
	comp := NewWithOptions(EquateNaNs())

	nan := math.NaN()
	fmt.Println(comp.Equal(nan, nan))
	// Output:
	// true
}

// ExampleWithCustomComparator registers domain-specific comparison logic for a
// concrete type.
func ExampleWithCustomComparator() {
	type account struct {
		ID   int
		Name string
	}

	// Two accounts are considered equal when they share the same ID.
	comp := NewWithOptions(WithCustomComparator(func(a, b account) bool {
		return a.ID == b.ID
	}))

	a1 := account{ID: 7, Name: "prod"}
	a2 := account{ID: 7, Name: "production"}

	fmt.Println(comp.Equal(a1, a2))
	// Output:
	// true
}

// ExampleWithTimeLayout controls how time.Time values are rendered in diff
// output, while equality still uses the native time comparison.
func ExampleWithTimeLayout() {
	comp := NewWithOptions(WithTimeLayout(time.RFC3339))

	t1 := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)

	fmt.Println(comp.Equal(t1, t2))
	// Output:
	// true
}

// ExampleGetJSONPatch generates an RFC 6902 JSON Patch describing how to turn
// the first document into the second.
func ExampleGetJSONPatch() {
	type profile struct {
		Name string `json:"name"`
	}

	patch, err := GetJSONPatch(profile{Name: "Jon"}, profile{Name: "John"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(len(patch))
	fmt.Println(patch[0].Op)
	fmt.Println(patch[0].Path)
	// Output:
	// 1
	// replace
	// /Name
}

// ExampleDiffComparer_GetUnifiedDiff produces a Unix diff-style report.
func ExampleDiffComparer_GetUnifiedDiff() {
	type release struct {
		Version string
	}

	comp := NewDiffComparer()
	diff, err := comp.GetUnifiedDiff(release{Version: "1.0.0"}, release{Version: "1.1.0"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(len(diff.Chunks) > 0)
	// Output:
	// true
}

// ExampleDiffComparer_GetVisualDiff builds a hierarchical tree of the
// differences between two values.
func ExampleDiffComparer_GetVisualDiff() {
	type server struct {
		Host string
		Port int
	}

	comp := NewDiffComparer()
	visual, err := comp.GetVisualDiff(
		server{Host: "localhost", Port: 8080},
		server{Host: "localhost", Port: 9090},
	)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(visual.Root != nil)
	// Output:
	// true
}

// ExampleDiffComparer_FormatDiff renders a diff result as Markdown.
func ExampleDiffComparer_FormatDiff() {
	type plan struct {
		Tier string
	}

	comp := NewDiffComparer(WithOutputFormat("markdown"))
	result := comp.CompareWithDiff(plan{Tier: "free"}, plan{Tier: "pro"})

	formatted, err := comp.FormatDiff(result, "markdown")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(len(formatted) > 0)
	// Output:
	// true
}

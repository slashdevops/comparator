package comparator

import (
	"context"
	"fmt"
)

// ExampleEqualT demonstrates the type-safe generic equality helper.
func ExampleEqualT() {
	type point struct{ X, Y int }

	fmt.Println(EqualT(point{1, 2}, point{1, 2}))
	fmt.Println(EqualT(point{1, 2}, point{3, 4}))
	// Output:
	// true
	// false
}

// ExampleWithEquateEmpty shows disabling the default nil/empty equivalence.
func ExampleWithEquateEmpty() {
	var nilSlice []int
	emptySlice := []int{}

	fmt.Println(NewWithOptions(WithEquateEmpty(true)).Equal(nilSlice, emptySlice))
	fmt.Println(NewWithOptions(WithEquateEmpty(false)).Equal(nilSlice, emptySlice))
	// Output:
	// true
	// false
}

// ExampleWithFieldNaming shows JSON-tag-aware JSON Patch pointers.
func ExampleWithFieldNaming() {
	type profile struct {
		FullName string `json:"full_name"`
	}

	patch, _ := GetJSONPatchT(
		profile{FullName: "Jon"},
		profile{FullName: "John"},
		WithFieldNaming(JSONTagNaming),
	)

	fmt.Println(patch[0].Op, patch[0].Path)
	// Output:
	// replace /full_name
}

// ExampleApplyJSONPatch applies a patch to a JSON-shaped document.
func ExampleApplyJSONPatch() {
	doc := map[string]any{"name": "Jon"}
	patch := []JSONPatchOperation{
		{Op: "replace", Path: "/name", Value: "John"},
		{Op: "add", Path: "/email", Value: "john@example.com"},
	}

	got, err := ApplyJSONPatch(doc, patch)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	m := got.(map[string]any)
	fmt.Println(m["name"], m["email"])
	// Output:
	// John john@example.com
}

// ExampleWithIgnoreMapKeys ignores volatile map keys.
func ExampleWithIgnoreMapKeys() {
	a := map[string]int{"value": 1, "updatedAt": 100}
	b := map[string]int{"value": 1, "updatedAt": 999}

	fmt.Println(NewWithOptions(WithIgnoreMapKeys("updatedAt")).Equal(a, b))
	// Output:
	// true
}

// ExampleWithReporter streams differences via a callback.
func ExampleWithReporter() {
	type cfg struct{ A, B int }

	count := 0
	comp := NewDiffComparer(WithReporter(func(Difference) { count++ }))
	comp.CompareWithDiff(cfg{1, 2}, cfg{1, 3})

	fmt.Println(count > 0)
	// Output:
	// true
}

// ExampleEqualCtx shows a context-aware comparison that is canceled up front.
func ExampleEqualCtx() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := EqualCtx(ctx, []int{1, 2, 3}, []int{1, 2, 3})
	fmt.Println(err != nil)
	// Output:
	// true
}

// ExampleRegisterFormatter adds a custom output format.
func ExampleRegisterFormatter() {
	RegisterFormatter("summary-only", func(r *DiffResult) (string, error) {
		if r.Equal {
			return "equal", nil
		}
		return "different", nil
	})
	defer UnregisterFormatter("summary-only")

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(1, 2)
	out, _ := comp.FormatDiff(result, "summary-only")

	fmt.Println(out)
	// Output:
	// different
}

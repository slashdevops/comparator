package comparator

import "fmt"

func ExampleNew() {
	comp := New()
	fmt.Println(comp.Equal(42, 42))
	fmt.Println(comp.Equal("left", "right"))
	// Output:
	// true
	// false
}

func ExampleIgnoreSliceOrder() {
	comp := NewWithOptions(IgnoreSliceOrder())
	left := []string{"api", "worker", "db"}
	right := []string{"db", "api", "worker"}

	fmt.Println(comp.Equal(left, right))
	// Output:
	// true
}

func ExampleNewDiffComparer() {
	type serviceConfig struct {
		Host string
		Port int
	}

	left := serviceConfig{Host: "localhost", Port: 8080}
	right := serviceConfig{Host: "localhost", Port: 9090}

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(left, right)

	fmt.Println(result.Equal)
	fmt.Println(len(result.Differences) > 0)
	// Output:
	// false
	// true
}

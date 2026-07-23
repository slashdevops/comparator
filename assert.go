package comparator

import (
	"fmt"
	"strings"
)

// TestingT is the subset of *testing.T (and *testing.B) used by the assertion
// helpers. Accepting an interface avoids importing the testing package into
// production builds and lets callers supply their own compatible type.
type TestingT interface {
	Helper()
	Errorf(format string, args ...any)
}

// tbFatal is optionally implemented by *testing.T/*testing.B via Fatalf. When
// available, the Require* helpers use it to stop the test immediately.
type tbFatal interface {
	Fatalf(format string, args ...any)
}

// AssertEqual fails the test (via t.Errorf) when want and got are not equal
// under the given options, printing a readable diff of the differences. It
// returns true when the values are equal. Use it as a drop-in equality check in
// tests:
//
//	comparator.AssertEqual(t, wantUser, gotUser, comparator.IgnoreStructFields("ID"))
func AssertEqual(t TestingT, want, got any, opts ...Option) bool {
	t.Helper()
	if ok, msg := diffMessage(want, got, opts...); !ok {
		t.Errorf("%s", msg)
		return false
	}
	return true
}

// AssertNotEqual fails the test when want and got are equal under the given
// options. It returns true when the values differ.
func AssertNotEqual(t TestingT, want, got any, opts ...Option) bool {
	t.Helper()
	if EqualT(want, got, opts...) {
		t.Errorf("expected values to differ, but they are equal")
		return false
	}
	return true
}

// RequireEqual behaves like AssertEqual but stops the test immediately (via
// Fatalf when t supports it) instead of merely marking it failed.
func RequireEqual(t TestingT, want, got any, opts ...Option) {
	t.Helper()
	if ok, msg := diffMessage(want, got, opts...); !ok {
		fail(t, msg)
	}
}

// diffMessage reports whether want and got are equal and, when not, a
// human-readable multi-line message describing the differences.
func diffMessage(want, got any, opts ...Option) (bool, string) {
	comp := NewDiffComparer(opts...)
	result := comp.CompareWithDiff(want, got)
	if result.Equal {
		return true, ""
	}

	var b strings.Builder
	b.WriteString("values are not equal:\n")
	b.WriteString(result.Summary)
	for _, d := range result.Differences {
		if d.Detail.Type == "equal" {
			continue
		}
		fmt.Fprintf(&b, "\n  %s: %s", d.Path, d.Message)
	}
	return false, b.String()
}

func fail(t TestingT, msg string) {
	if ft, ok := t.(tbFatal); ok {
		ft.Fatalf("%s", msg)
		return
	}
	t.Errorf("%s", msg)
}

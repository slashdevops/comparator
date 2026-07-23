package comparator

import (
	"context"
	"errors"
	"testing"
)

// ---- #1 Equatable / Comparable ---------------------------------------------

type equatablePoint struct {
	X, Y int
	Tag  string // deliberately not part of equality
}

func (p equatablePoint) Equals(other equatablePoint) bool {
	return p.X == other.X && p.Y == other.Y
}

type comparableVersion struct {
	Major int
}

func (v comparableVersion) CompareTo(other comparableVersion) int {
	return v.Major - other.Major
}

func TestUserEquality_Equatable(t *testing.T) {
	a := equatablePoint{X: 1, Y: 2, Tag: "left"}
	b := equatablePoint{X: 1, Y: 2, Tag: "right"}

	if !New().Equal(a, b) {
		t.Error("expected Equatable.Equals to make the points equal despite different Tag")
	}

	c := equatablePoint{X: 9, Y: 9, Tag: "left"}
	if New().Equal(a, c) {
		t.Error("expected Equatable.Equals to report differing coordinates as unequal")
	}
}

func TestUserEquality_Comparable(t *testing.T) {
	if !New().Equal(comparableVersion{Major: 3}, comparableVersion{Major: 3}) {
		t.Error("expected CompareTo == 0 to mean equal")
	}
	if New().Equal(comparableVersion{Major: 3}, comparableVersion{Major: 4}) {
		t.Error("expected CompareTo != 0 to mean not equal")
	}
}

func TestUserEquality_CustomComparatorTakesPrecedence(t *testing.T) {
	// A custom comparator must win over the type's own Equals method.
	comp := NewWithOptions(WithCustomComparator(func(a, b equatablePoint) bool {
		return a.Tag == b.Tag
	}))

	a := equatablePoint{X: 1, Y: 2, Tag: "same"}
	b := equatablePoint{X: 5, Y: 6, Tag: "same"}
	if !comp.Equal(a, b) {
		t.Error("expected custom comparator (by Tag) to take precedence over Equals")
	}
}

// ---- #2 Unified diff context lines -----------------------------------------

func TestUnifiedDiff_TypesAndContext(t *testing.T) {
	a := "line1\nline2\nline3\nline4\nline5"
	b := "line1\nline2\nCHANGED\nline4\nline5"

	ud, err := NewDiffComparer().GetUnifiedDiff(a, b)
	if err != nil {
		t.Fatalf("GetUnifiedDiff: %v", err)
	}
	if len(ud.Chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	var hasContext, hasAdd, hasRemove bool
	for _, ch := range ud.Chunks {
		for _, c := range ch.Changes {
			switch c.Type {
			case "context":
				hasContext = true
			case "add":
				hasAdd = true
			case "remove":
				hasRemove = true
			}
		}
	}
	if !hasContext {
		t.Error("expected context lines around the change")
	}
	if !hasAdd || !hasRemove {
		t.Error("expected both add and remove changes")
	}
}

func TestUnifiedDiff_NoChangesNoChunks(t *testing.T) {
	ud, err := NewDiffComparer().GetUnifiedDiff("same", "same")
	if err != nil {
		t.Fatalf("GetUnifiedDiff: %v", err)
	}
	if len(ud.Chunks) != 0 {
		t.Errorf("expected no chunks for identical input, got %d", len(ud.Chunks))
	}
}

// ---- #3 WithEquateEmpty toggle ---------------------------------------------

func TestWithEquateEmpty_Disable(t *testing.T) {
	var nilSlice []int
	emptySlice := []int{}

	if !NewWithOptions(WithEquateEmpty(true)).Equal(nilSlice, emptySlice) {
		t.Error("expected nil and empty to be equal when enabled")
	}
	if NewWithOptions(WithEquateEmpty(false)).Equal(nilSlice, emptySlice) {
		t.Error("expected nil and empty to differ when disabled")
	}
}

// ---- #4 Context cancellation -----------------------------------------------

func TestEqualCtx_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := EqualCtx(ctx, []int{1, 2, 3}, []int{1, 2, 3})
	if !errors.Is(err, ErrCanceled) {
		t.Errorf("expected ErrCanceled, got %v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected wrapped context.Canceled, got %v", err)
	}
}

func TestEqualCtx_Completes(t *testing.T) {
	ok, err := EqualCtx(context.Background(), 42, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected equal")
	}
}

// ---- #5 Generics -----------------------------------------------------------

func TestGenerics(t *testing.T) {
	if !EqualT(1, 1) {
		t.Error("EqualT should report equal ints")
	}
	if !DeepEqualT([]string{"a"}, []string{"a"}) {
		t.Error("DeepEqualT should report equal slices")
	}
	diffs, err := DiffT("a", "b")
	if err != nil || len(diffs) == 0 {
		t.Errorf("DiffT should report differences, got %d diffs err=%v", len(diffs), err)
	}
	if CompareWithDiffT(1, 2).Equal {
		t.Error("CompareWithDiffT should report inequality")
	}
	patch, err := GetJSONPatchT(struct{ Name string }{"a"}, struct{ Name string }{"b"})
	if err != nil || len(patch) == 0 {
		t.Errorf("GetJSONPatchT should produce a patch, got %d err=%v", len(patch), err)
	}
}

// ---- #6 JSON-tag field naming ----------------------------------------------

func TestFieldNaming_JSONTag(t *testing.T) {
	type profile struct {
		FullName string `json:"full_name"`
	}

	patch, err := GetJSONPatchT(
		profile{FullName: "Jon"},
		profile{FullName: "John"},
		WithFieldNaming(JSONTagNaming),
	)
	if err != nil {
		t.Fatalf("GetJSONPatch: %v", err)
	}
	if len(patch) != 1 || patch[0].Path != "/full_name" {
		t.Errorf("expected pointer /full_name, got %+v", patch)
	}
}

// ---- #7 ApplyJSONPatch -----------------------------------------------------

func TestApplyJSONPatch_Operations(t *testing.T) {
	doc := map[string]any{
		"name":  "Jon",
		"tags":  []any{"a", "b"},
		"stale": true,
	}
	patch := []JSONPatchOperation{
		{Op: "replace", Path: "/name", Value: "John"},
		{Op: "add", Path: "/email", Value: "john@example.com"},
		{Op: "remove", Path: "/stale"},
		{Op: "add", Path: "/tags/-", Value: "c"},
		{Op: "test", Path: "/name", Value: "John"},
	}

	got, err := ApplyJSONPatch(doc, patch)
	if err != nil {
		t.Fatalf("ApplyJSONPatch: %v", err)
	}

	m := got.(map[string]any)
	if m["name"] != "John" {
		t.Errorf("replace failed: %v", m["name"])
	}
	if m["email"] != "john@example.com" {
		t.Errorf("add failed: %v", m["email"])
	}
	if _, ok := m["stale"]; ok {
		t.Error("remove failed: stale still present")
	}
	if tags := m["tags"].([]any); len(tags) != 3 || tags[2] != "c" {
		t.Errorf("array append failed: %v", tags)
	}
}

func TestApplyJSONPatch_RoundTrip(t *testing.T) {
	type addr struct {
		City string `json:"city"`
		Zip  string `json:"zip"`
	}
	type user struct {
		Name    string   `json:"name"`
		Addr    addr     `json:"addr"`
		Aliases []string `json:"aliases"`
	}

	from := user{Name: "Jon", Addr: addr{City: "NYC", Zip: "10001"}, Aliases: []string{"j"}}
	to := user{Name: "John", Addr: addr{City: "NYC", Zip: "10002"}, Aliases: []string{"j", "johnny"}}

	patch, err := GetJSONPatchT(from, to, WithFieldNaming(JSONTagNaming))
	if err != nil {
		t.Fatalf("GetJSONPatch: %v", err)
	}

	normalizedFrom, _ := toJSONValue(from)
	got, err := ApplyJSONPatch(normalizedFrom, patch)
	if err != nil {
		t.Fatalf("ApplyJSONPatch: %v", err)
	}

	normalizedTo, _ := toJSONValue(to)
	if !Equal(got, normalizedTo) {
		t.Errorf("round trip mismatch:\n got=%v\nwant=%v", got, normalizedTo)
	}
}

func TestApplyJSONPatch_TestFailure(t *testing.T) {
	_, err := ApplyJSONPatch(map[string]any{"n": 1}, []JSONPatchOperation{
		{Op: "test", Path: "/n", Value: 2},
	})
	if !errors.Is(err, ErrInvalidPatch) {
		t.Errorf("expected ErrInvalidPatch on failed test op, got %v", err)
	}
}

// ---- #8 Reporter -----------------------------------------------------------

func TestWithReporter(t *testing.T) {
	type cfg struct {
		A, B, C int
	}

	var seen int
	comp := NewDiffComparer(WithReporter(func(Difference) { seen++ }))
	result := comp.CompareWithDiff(cfg{1, 2, 3}, cfg{1, 9, 8})

	if seen == 0 {
		t.Fatal("reporter was never called")
	}
	if seen != len(result.Differences) {
		t.Errorf("reporter calls (%d) != differences (%d)", seen, len(result.Differences))
	}
}

// ---- #9 Richer ignore rules ------------------------------------------------

func TestWithIgnoreMapKeys(t *testing.T) {
	a := map[string]int{"keep": 1, "ts": 100}
	b := map[string]int{"keep": 1, "ts": 999}

	if !NewWithOptions(WithIgnoreMapKeys("ts")).Equal(a, b) {
		t.Error("expected equality when ignoring the differing key")
	}
	if New().Equal(a, b) {
		t.Error("expected inequality without ignore")
	}
}

func TestWithIgnorePaths(t *testing.T) {
	type meta struct {
		UpdatedAt string
	}
	type doc struct {
		Name string
		Meta meta
	}

	a := doc{Name: "x", Meta: meta{UpdatedAt: "t1"}}
	b := doc{Name: "x", Meta: meta{UpdatedAt: "t2"}}

	if !NewWithOptions(WithIgnorePaths("Meta.UpdatedAt")).Equal(a, b) {
		t.Error("expected equality when ignoring Meta.UpdatedAt")
	}
}

func TestWithIgnorePathPatterns(t *testing.T) {
	type creds struct {
		User     string
		Password string
	}
	a := creds{User: "u", Password: "p1"}
	b := creds{User: "u", Password: "p2"}

	if !NewWithOptions(WithIgnorePathPatterns(`(^|\.)Password$`)).Equal(a, b) {
		t.Error("expected equality when ignoring Password by pattern")
	}
}

func TestStructTagIgnore(t *testing.T) {
	type record struct {
		Name  string
		Cache string `comparator:"-"`
	}
	a := record{Name: "x", Cache: "c1"}
	b := record{Name: "x", Cache: "c2"}

	if !New().Equal(a, b) {
		t.Error("expected the comparator:\"-\" tagged field to be ignored")
	}
}

// ---- #10 Pluggable formatters ----------------------------------------------

func TestRegisterFormatter(t *testing.T) {
	const name = "count-test-format"
	if !RegisterFormatter(name, func(r *DiffResult) (string, error) {
		return "diffs=" + itoa(len(r.Differences)), nil
	}) {
		t.Fatal("RegisterFormatter should succeed for a new name")
	}
	defer UnregisterFormatter(name)

	comp := NewDiffComparer()
	result := comp.CompareWithDiff(1, 2)
	out, err := comp.FormatDiff(result, name)
	if err != nil {
		t.Fatalf("FormatDiff custom: %v", err)
	}
	if out == "" {
		t.Error("expected custom formatter output")
	}

	if RegisterFormatter("json", func(*DiffResult) (string, error) { return "", nil }) {
		t.Error("RegisterFormatter must refuse to override a built-in format")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// ---- #11 String helpers ----------------------------------------------------

func TestUnifiedDiffString(t *testing.T) {
	ud, _ := NewDiffComparer().GetUnifiedDiff("a\nb", "a\nc")
	s := ud.String()
	if s == "" || !contains(s, "@@") {
		t.Errorf("expected a rendered unified diff, got %q", s)
	}
}

func TestVisualDiffString(t *testing.T) {
	type point struct{ X, Y int }
	vd, err := NewDiffComparer().GetVisualDiff(point{1, 2}, point{1, 3})
	if err != nil {
		t.Fatalf("GetVisualDiff: %v", err)
	}
	if vd.String() == "" {
		t.Error("expected a rendered visual diff")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---- #13 Assertion helpers -------------------------------------------------

// fakeT captures assertion output for testing the helpers.
type fakeT struct {
	failed bool
	msg    string
}

func (f *fakeT) Helper()                           {}
func (f *fakeT) Errorf(format string, args ...any) { f.failed = true; f.msg = format }
func (f *fakeT) Fatalf(format string, args ...any) { f.failed = true; f.msg = format }

func TestAssertEqual(t *testing.T) {
	ft := &fakeT{}
	if !AssertEqual(ft, 1, 1) {
		t.Error("AssertEqual should pass for equal values")
	}
	if ft.failed {
		t.Error("AssertEqual should not fail for equal values")
	}

	ft = &fakeT{}
	if AssertEqual(ft, 1, 2) {
		t.Error("AssertEqual should fail for unequal values")
	}
	if !ft.failed {
		t.Error("AssertEqual should mark the test failed for unequal values")
	}
}

func TestAssertNotEqual(t *testing.T) {
	ft := &fakeT{}
	if !AssertNotEqual(ft, 1, 2) {
		t.Error("AssertNotEqual should pass for differing values")
	}
	if ft.failed {
		t.Error("AssertNotEqual should not fail for differing values")
	}
}

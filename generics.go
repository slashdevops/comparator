package comparator

// This file provides a thin, type-safe layer over the any-based API. The generic
// helpers give compile-time type checking at call sites while delegating to the
// same comparison engine.

// EqualT reports whether two values of the same type are deeply equal. It is the
// type-safe counterpart of Equal: the compiler guarantees a and b share a type.
//
// Example:
//
//	if comparator.EqualT(user1, user2, comparator.IgnoreStructFields("ID")) {
//	    // ...
//	}
func EqualT[T any](a, b T, opts ...Option) bool {
	return NewWithOptions(opts...).Equal(a, b)
}

// DeepEqualT is an alias for EqualT, emphasizing the recursive comparison. It
// mirrors DeepEqual for callers who prefer that name.
func DeepEqualT[T any](a, b T, opts ...Option) bool {
	return EqualT(a, b, opts...)
}

// DiffT returns all differences between two values of the same type. It is the
// type-safe counterpart of a Comparator.Diff call.
func DiffT[T any](a, b T, opts ...Option) ([]Difference, error) {
	return NewWithOptions(opts...).Diff(a, b)
}

// CompareWithDiffT performs a comprehensive comparison between two values of the
// same type and returns a detailed result. It is the type-safe counterpart of
// CompareWithDiff.
func CompareWithDiffT[T any](a, b T, opts ...Option) *DiffResult {
	return NewDiffComparer(opts...).CompareWithDiff(a, b)
}

// GetJSONPatchT generates an RFC 6902 JSON Patch describing how to turn a into
// b, for values of the same type. It is the type-safe counterpart of
// GetJSONPatch.
func GetJSONPatchT[T any](a, b T, opts ...Option) ([]JSONPatchOperation, error) {
	return NewDiffComparer(opts...).GetJSONPatch(a, b)
}

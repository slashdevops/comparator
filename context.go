package comparator

import (
	"context"
	"fmt"
)

// ContextComparator is implemented by the comparators returned from New,
// NewWithOptions, and NewDiffComparer. Its methods accept a context.Context so
// long-running comparisons over large structures can be canceled or bounded by
// a deadline.
//
// A canceled context causes the comparison to stop early and return an error
// that wraps both ErrCanceled and the context's own error.
type ContextComparator interface {
	// EqualCtx behaves like Comparator.Equal but honors ctx cancellation.
	EqualCtx(ctx context.Context, a, b any) (bool, error)

	// DiffCtx behaves like Comparator.Diff but honors ctx cancellation.
	DiffCtx(ctx context.Context, a, b any) ([]Difference, error)

	// CompareWithDiffCtx behaves like DiffComparer.CompareWithDiff but honors
	// ctx cancellation.
	CompareWithDiffCtx(ctx context.Context, a, b any) (*DiffResult, error)
}

// ctxError wraps the context's error together with ErrCanceled so that callers
// can branch with errors.Is on either sentinel.
func ctxError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrCanceled, err)
	}
	return ErrCanceled
}

// EqualCtx performs a deep equality check that stops early if ctx is canceled.
func (c *defaultComparator) EqualCtx(ctx context.Context, a, b any) (bool, error) {
	c.ctx = ctx
	defer func() { c.ctx = nil }()

	result := c.Equal(a, b)
	if c.cancelled {
		return false, ctxError(ctx)
	}
	return result, nil
}

// DiffCtx collects differences and stops early if ctx is canceled.
func (c *defaultComparator) DiffCtx(ctx context.Context, a, b any) ([]Difference, error) {
	c.ctx = ctx
	defer func() { c.ctx = nil }()

	diffs, _ := c.Diff(a, b)
	if c.cancelled {
		return nil, ctxError(ctx)
	}
	return diffs, nil
}

// CompareWithDiffCtx performs a comprehensive comparison and stops early if ctx
// is canceled.
func (c *defaultComparator) CompareWithDiffCtx(ctx context.Context, a, b any) (*DiffResult, error) {
	c.ctx = ctx
	defer func() { c.ctx = nil }()

	result := c.CompareWithDiff(a, b)
	if c.cancelled {
		return nil, ctxError(ctx)
	}
	return result, nil
}

// EqualCtx is a package-level convenience for a one-off context-aware equality
// check.
func EqualCtx(ctx context.Context, a, b any, opts ...Option) (bool, error) {
	return NewWithOptions(opts...).(ContextComparator).EqualCtx(ctx, a, b)
}

// DiffCtx is a package-level convenience for a one-off context-aware diff.
func DiffCtx(ctx context.Context, a, b any, opts ...Option) ([]Difference, error) {
	return NewWithOptions(opts...).(ContextComparator).DiffCtx(ctx, a, b)
}

// CompareWithDiffCtx is a package-level convenience for a one-off context-aware
// comprehensive comparison.
func CompareWithDiffCtx(ctx context.Context, a, b any, opts ...Option) (*DiffResult, error) {
	return NewDiffComparer(opts...).(ContextComparator).CompareWithDiffCtx(ctx, a, b)
}

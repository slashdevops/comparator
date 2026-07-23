package comparator

import "errors"

// Sentinel errors returned by the package. Use errors.Is to branch on them.
var (
	// ErrUnknownFormat is returned by FormatDiff when the requested format is
	// not a built-in format ("text", "json", "markdown", "html") and no
	// formatter has been registered for it with RegisterFormatter.
	ErrUnknownFormat = errors.New("comparator: unknown output format")

	// ErrCanceled is returned by the context-aware comparison functions and
	// methods when the provided context is canceled or its deadline is exceeded
	// before the comparison completes. It wraps the context's own error, so
	// errors.Is(err, context.Canceled) and errors.Is(err, context.DeadlineExceeded)
	// also report true.
	ErrCanceled = errors.New("comparator: comparison canceled")

	// ErrInvalidPatch is returned by ApplyJSONPatch when a patch operation is
	// malformed, references an unreachable path, or specifies an unsupported
	// operation.
	ErrInvalidPatch = errors.New("comparator: invalid JSON patch")
)

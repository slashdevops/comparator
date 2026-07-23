package comparator

import (
	"strings"
	"sync"
)

// Formatter renders a DiffResult into a string representation. Register one with
// RegisterFormatter to teach FormatDiff a new output format (for example CSV,
// JUnit, or SARIF) without changing the package.
type Formatter func(result *DiffResult) (string, error)

var (
	formatterMu       sync.RWMutex
	customFormatters  = map[string]Formatter{}
	builtinFormatters = map[string]bool{
		"text": true, "json": true, "markdown": true, "html": true,
	}
)

// RegisterFormatter registers a custom formatter under the given name so that
// FormatDiff(result, name) and WithOutputFormat(name) can use it. The name is
// case-insensitive and must not collide with a built-in format ("text", "json",
// "markdown", "html"). Registering a nil formatter, an empty name, or a
// built-in name is a no-op that returns false.
//
// RegisterFormatter is safe for concurrent use. Registering the same name again
// replaces the previous formatter.
func RegisterFormatter(name string, fn Formatter) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || fn == nil || builtinFormatters[name] {
		return false
	}
	formatterMu.Lock()
	defer formatterMu.Unlock()
	customFormatters[name] = fn
	return true
}

// UnregisterFormatter removes a previously registered custom formatter. It
// returns true if a formatter was removed.
func UnregisterFormatter(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	formatterMu.Lock()
	defer formatterMu.Unlock()
	if _, ok := customFormatters[name]; !ok {
		return false
	}
	delete(customFormatters, name)
	return true
}

// lookupFormatter returns the custom formatter registered for a name, if any.
func lookupFormatter(name string) (Formatter, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	formatterMu.RLock()
	defer formatterMu.RUnlock()
	fn, ok := customFormatters[name]
	return fn, ok
}

package comparator

import "regexp"

// WithEquateEmpty controls whether nil and empty containers (slices and maps)
// are treated as equal. Unlike EquateEmpty, which can only enable the behavior,
// this option can also disable it (the default is enabled).
//
// Example:
//
//	// Treat nil and empty as distinct.
//	comp := comparator.NewWithOptions(comparator.WithEquateEmpty(false))
func WithEquateEmpty(enable bool) Option {
	return func(c *Config) {
		c.equateEmpty = enable
	}
}

// WithFieldNaming selects how struct field names appear in difference paths and
// JSON Patch pointers. The default is GoFieldNaming. Use JSONTagNaming to make
// paths and patches align with the JSON representation of the values.
//
// Example:
//
//	comp := comparator.NewDiffComparer(comparator.WithFieldNaming(comparator.JSONTagNaming))
//	// A field `Name string json:"name"` produces the pointer "/name".
func WithFieldNaming(naming FieldNaming) Option {
	return func(c *Config) {
		c.fieldNaming = naming
	}
}

// WithReporter registers a callback invoked for every difference as it is
// discovered during CompareWithDiff, Diff, and the Get*Diff methods. It is
// useful for streaming, logging, or integrating with test frameworks without
// waiting for the full result to be materialized.
//
// The reporter is called in traversal order and must not retain the Difference
// beyond the callback if the caller mutates shared state.
//
// Example:
//
//	comp := comparator.NewDiffComparer(comparator.WithReporter(func(d comparator.Difference) {
//	    log.Printf("%s: %s", d.Path, d.Message)
//	}))
func WithReporter(reporter func(Difference)) Option {
	return func(c *Config) {
		c.reporter = reporter
	}
}

// WithIgnoreMapKeys skips the named map keys anywhere they appear during
// comparison. Keys are matched by their default string representation, so
// non-string keys such as integers are matched by their formatted value.
//
// Example:
//
//	comp := comparator.NewWithOptions(comparator.WithIgnoreMapKeys("updatedAt", "etag"))
func WithIgnoreMapKeys(keys ...string) Option {
	return func(c *Config) {
		for _, k := range keys {
			c.ignoreMapKeys[k] = true
		}
	}
}

// WithIgnorePaths skips struct fields at the given canonical paths. Paths use
// dotted field names with bracketed index or key segments kept inline, matching
// the form reported in Difference.Path, e.g. "User.Address.Zip".
//
// Example:
//
//	comp := comparator.NewWithOptions(comparator.WithIgnorePaths("Meta.UpdatedAt", "Meta.Etag"))
func WithIgnorePaths(paths ...string) Option {
	return func(c *Config) {
		for _, p := range paths {
			c.ignorePaths[p] = true
		}
	}
}

// WithIgnorePathPatterns skips struct fields whose canonical path matches any of
// the given regular expressions. Invalid patterns are ignored so option
// construction never panics; use regexp.Compile beforehand if you need to
// validate patterns.
//
// Example:
//
//	// Ignore every field named "password" at any depth.
//	comp := comparator.NewWithOptions(comparator.WithIgnorePathPatterns(`(^|\.)[Pp]assword$`))
func WithIgnorePathPatterns(patterns ...string) Option {
	return func(c *Config) {
		for _, p := range patterns {
			if re, err := regexp.Compile(p); err == nil {
				c.ignorePathPatterns = append(c.ignorePathPatterns, re)
			}
		}
	}
}

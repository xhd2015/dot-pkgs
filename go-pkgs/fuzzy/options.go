package fuzzy

import "os"

// Option configures Match / MatchAll.
type Option func(*config)

type config struct {
	caseSensitive bool
	pathScheme    bool
	separators    map[rune]struct{}
}

func defaultConfig() config {
	return config{
		separators: defaultSeparators(),
	}
}

func applyOptions(opts []Option) config {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// WithCaseSensitive disables rune-wise case folding.
func WithCaseSensitive() Option {
	return func(c *config) {
		c.caseSensitive = true
	}
}

// WithPathScheme adds extra boundary weight on path separators
// ("/" and, on Windows, "\").
func WithPathScheme() Option {
	return func(c *config) {
		c.pathScheme = true
	}
}

// WithSeparators replaces the default word-separator set. Whitespace is
// still treated as a boundary regardless of this list.
func WithSeparators(seps []rune) Option {
	return func(c *config) {
		set := make(map[rune]struct{}, len(seps))
		for _, r := range seps {
			set[r] = struct{}{}
		}
		c.separators = set
	}
}

func defaultSeparators() map[rune]struct{} {
	return map[rune]struct{}{
		'-': {},
		'_': {},
		'/': {},
		'.': {},
	}
}

func isPathSep(r rune) bool {
	if r == '/' {
		return true
	}
	return os.PathSeparator != '/' && r == rune(os.PathSeparator)
}

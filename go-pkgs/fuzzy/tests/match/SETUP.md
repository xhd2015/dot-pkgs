# Scenario

**Feature**: fzf V1 subsequence match, token AND, and query split

```
# match pipeline
caller haystack + query -> Match | MatchAll | Tokens -> OK/score/spans or tokens

# Match (one token; spaces literal)
empty query -> OK score 0 unmatched haystack | subsequence -> OK + spans | else !OK

# MatchAll (AND tokens)
nil Tokens -> Tokens(Query) | empty tokens -> OK score 0 unmatched | all hit -> OK | else !OK

# spans
join Text == haystack; default fold; matched Text keeps haystack case
```

## Preconditions

- The `fuzzy` package is importable (`github.com/xhd2015/dot-pkgs/go-pkgs/fuzzy`).
- Greenfield: `Match` / `MatchAll` / `Tokens` / `Span` / `Option` do not exist
  yet. Root `Run` calls those APIs so the suite is compile-RED. Do not stub them.
- Planned API (for implementer; not defined here):
  - `func Match(haystack, query string, opts ...Option) Result`
  - `func MatchAll(haystack string, tokens []string, opts ...Option) Result`
  - `func Tokens(query string) []string`
  - `func WithCaseSensitive() Option`
  - `func WithPathScheme() Option`
  - `type Result struct { OK bool; Score int; Spans []Span }`
  - `type Span struct { Text string; Matched bool }`
- Parallel-safe: no `os.Chdir` / `t.Chdir` / `os.Setenv` / `t.Setenv`. The
  harness runs leaves under `t.Parallel()`.
- Pure functions of haystack, query/tokens, and options. No filesystem, cwd,
  or process env.

## Steps

1. Leaf `Setup` sets `req.Op` and the haystack / query / tokens / option flags.
2. Root `Run` builds `[]fuzzy.Option` from flags and calls `fuzzy.Match`,
   `fuzzy.MatchAll`, or `fuzzy.Tokens`.
3. Leaf `Assert` checks `OK`, score, span text/matched runs, or token split.

## Context

- `joinSpans` concatenates `Span.Text`. `matchedTexts` collects matched `Text`.
- `Match` does not split on spaces. `MatchAll` does (via `Tokens` when
  `req.Tokens` is nil).
- `WithPathScheme` is wired in `Run` but unused by this tree's leaves.

```go
import (
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/fuzzy"
)

func joinSpans(spans []fuzzy.Span) string {
	var b strings.Builder
	for _, s := range spans {
		b.WriteString(s.Text)
	}
	return b.String()
}

func matchedTexts(spans []fuzzy.Span) []string {
	var out []string
	for _, s := range spans {
		if s.Matched {
			out = append(out, s.Text)
		}
	}
	return out
}
```

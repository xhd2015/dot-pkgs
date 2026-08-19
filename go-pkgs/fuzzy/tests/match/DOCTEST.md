# Match — `fuzzy.Match` / `fuzzy.MatchAll` / `fuzzy.Tokens`

## Version
0.0.2

Unit-style doc tests for fzf-V1 subsequence matching in
`github.com/xhd2015/dot-pkgs/go-pkgs/fuzzy`. `Match` scores one literal token;
`MatchAll` ANDs whitespace-split tokens; `Tokens` is the splitter.

**Classic TDD:** `fuzzy` does not exist yet. Root `Run` calls `fuzzy.Match`,
`fuzzy.MatchAll`, and `fuzzy.Tokens` so the suite is compile-RED until the
implementer adds the API. Do not implement production code in this design pass.

## DSN (Domain Specific Notion)

### Participants

- **`Match`** — fzf V1 subsequence matcher for **one token**. Every query rune
  must appear in order in the haystack. Spaces in `Query` are **literal** (not
  token separators). Returns `OK`, `Score`, and `[]Span`.
- **`MatchAll`** — same subsequence engine, **AND** across tokens: every token
  must match. Tokens come from `Request.Tokens`, or from `Tokens(Query)` when
  that field is nil.
- **`Tokens`** — splits a query on whitespace and drops empty fields
  (`"  aid   user "` → `["aid","user"]`).
- **`Span`** — haystack slice `{Text, Matched}`. Adjacent same-`Matched` runes
  form one span. Concatenating every `Text` in order **equals the haystack**.
- **`Option`** — `WithCaseSensitive()` (default is case-insensitive);
  `WithPathScheme()` (path-aware ranking; not exercised as a leaf here).

### Behaviors

- **fzf V1 subsequence** — query characters must appear in order, not
  necessarily contiguously. No glob / regex / inversion.
- **Match is one token** — spaces are literal. `Match("aid-user", "aid user")`
  does not match (the space is a required character).
- **MatchAll ANDs tokens** — all tokens must match; missing any token → `!OK`.
- **spans join == haystack** — `joinSpans(spans)` is exactly the input
  haystack on a successful match (including empty-query / empty-token cases).
- **Default case-insensitive** — `"AID"` matches `"aid"` and matched `Text`
  keeps the haystack's original case. `WithCaseSensitive()` disables folding.
- **Empty query / empty tokens** — `OK`, score `0`, one unmatched span covering
  the whole haystack.
- **No match** — `OK` is false. Score / spans are unspecified.
- **Consecutive rank** — a contiguous hit outranks the same letters with a gap
  (`Match("ab","ab")` score > `Match("a-b","ab")` score).

### Inverse

No inverse. Matching is not a codec.

## Decision Tree

```
fuzzy/tests/match/                   [fzf V1 subsequence; Match vs MatchAll]
├── empty-query                      Match("foo","") OK score 0 [{foo,false}]
├── no-match                         Match("brainstorm","zzz") !OK
├── subsequence-spans                Match("brainstorm","bsm") OK; b/rain/s/tor/m
├── case-fold                        Match(...,"AID") OK; matched Text is "aid"
├── case-sensitive                   WithCaseSensitive Match(...,"AID") !OK
├── consecutive-rank                 Score("ab","ab") > Score("a-b","ab")
├── literal-space                    Match("aid-user","aid user") !OK
├── tokens                           Tokens("  aid   user ") → ["aid","user"]
├── match-all-and                    MatchAll AND "aid user" OK; join == haystack
├── match-all-miss                   MatchAll("followup","aid user") !OK
└── match-all-empty                  MatchAll("foo", empty) OK score 0 unmatched
```

### Parameter significance (high → low)

1. **Op** — `match` vs `match_all` vs `tokens` (which public function).
2. **Query emptiness / token count** — empty (score 0, unmatched span) vs
   one token vs AND of several tokens.
3. **Hit vs miss** — subsequence present or not (including literal space).
4. **Case** — default fold vs `WithCaseSensitive`.
5. **Adjacency** — consecutive vs gapped (ranking only).

## Test Index

| Leaf | Description |
|------|-------------|
| `empty-query` | `Match("foo","")` is OK, score 0, one unmatched span `"foo"` |
| `no-match` | `Match("brainstorm","zzz")` is not OK |
| `subsequence-spans` | `Match("brainstorm","bsm")` OK; spans `b` / `rain` / `s` / `tor` / `m` |
| `case-fold` | `Match("aid-user-do-human-verifications","AID")` OK; matched text is `"aid"` |
| `case-sensitive` | `WithCaseSensitive` `Match("aid-user","AID")` is not OK |
| `consecutive-rank` | `Score Match("ab","ab")` > `Score Match("a-b","ab")` |
| `literal-space` | `Match("aid-user","aid user")` is not OK (space is literal) |
| `tokens` | `Tokens("  aid   user ")` is `["aid","user"]` |
| `match-all-and` | `MatchAll` of `"aid user"` on the aid-user haystack is OK; join == haystack; `"aid"` and `"user"` matched |
| `match-all-miss` | `MatchAll("followup","aid user")` is not OK |
| `match-all-empty` | `MatchAll("foo", Tokens(""))` is OK, score 0, one unmatched span |

## How to Run

From the go-pkgs module root (package is local):

```sh
doctest vet ./fuzzy/tests/match
doctest test ./fuzzy/tests/match
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/fuzzy"
)

type Request struct {
	Op            string   // "match" | "match_all" | "tokens"
	Haystack      string
	Query         string   // for match; also used to derive tokens if Tokens field nil and Op=match_all
	Tokens        []string // for match_all; if nil and Op=match_all, call fuzzy.Tokens(Query)
	CaseSensitive bool
	PathScheme    bool
}

type Response struct {
	OK     bool
	Score  int
	Spans  []fuzzy.Span
	Tokens []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	var opts []fuzzy.Option
	if req.CaseSensitive {
		opts = append(opts, fuzzy.WithCaseSensitive())
	}
	if req.PathScheme {
		opts = append(opts, fuzzy.WithPathScheme())
	}
	switch req.Op {
	case "tokens":
		return &Response{Tokens: fuzzy.Tokens(req.Query)}, nil
	case "match_all":
		toks := req.Tokens
		if toks == nil {
			toks = fuzzy.Tokens(req.Query)
		}
		r := fuzzy.MatchAll(req.Haystack, toks, opts...)
		return &Response{OK: r.OK, Score: r.Score, Spans: r.Spans}, nil
	default: // match
		r := fuzzy.Match(req.Haystack, req.Query, opts...)
		return &Response{OK: r.OK, Score: r.Score, Spans: r.Spans}, nil
	}
}
```

# code-dup-stat

Detect similar Go functions across packages using n-gram or word-statistic similarity.

## Usage

```
code-dup-stat [--ngram N] [--threshold T] [--algorithm ngram|wordstat] [--dir DIR]
```

## Iterative development methodology

When improving the wordstat algorithm to reduce false positives:

1. **Run and spot false positives** — `code-dup-stat --algorithm=wordstat --dir <dir>`
2. **Read actual function source** to understand why they matched
3. **Add the pattern to testdata** — create `.go` files in `testdata/wordstat-no-dup/<pattern>/` capturing the failure case
4. **Fix the algorithm** — adjust weights, scoring formula, or filtering
5. **Run tests** — `go test ./dupstat/` and `go run ./cmd/code-dup-stat/tests/main.go`
6. **Repeat** until E2E output is clean

## Test data structure

```
cmd/code-dup-stat/testdata/
  wordstat-dup/        # pairs that SHOULD be detected
    pkg/handler_a.go
    pkg/handler_b.go
  wordstat-no-dup/     # pairs that should NOT be detected
    coincidental-vars/ # functions with coincidental variable names
      a.go
      b.go
    structural-vocab/  # functions sharing structural vocabulary
      a.go
      b.go
    wrapper-vs-large/  # small wrapper vs large function
      a.go
      b.go
    string-vocab-overlap/ # different logic, same string.* vocabulary
      a.go
      b.go
```

## Algorithm (wordstat)

1. Tokenize function bodies with Go scanner
2. Split identifiers into words: `ProcessUser` → `["process", "user"]`
3. Apply per-category weights: keywords/operators ×0.05, identifiers ×1.5, literals ×2.0
4. Filter single-letter words from multi-word splits (e.g., `freqA` → keep `"freq"`, drop `"a"`)
5. Compute TF-IDF across all functions: downweights words appearing in many functions
6. Score = `(Jaccard + Bidirectional-Containment) / 2`

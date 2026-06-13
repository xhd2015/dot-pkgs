# Manual Argument Parsing Detection Tests

Tests for the `manual-flag-parse` checker that detects manual flag parsing patterns
(`for`/`switch`, `for`/`if`) in Go source code and suggests using a proper flag library.

## How to Run

```sh
doctest test -v ./
```

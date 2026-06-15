# Error Unchecked or Ignored Detection Tests

Tests for the `err-unchecked-or-ignored` checker that detects error variables
that are checked (`if err != nil`) but not propagated in the return statement
(returning a fallback/default value instead of the error).

## How to Run

```sh
doctest test -v ./
```

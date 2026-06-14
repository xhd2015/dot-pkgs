# Type Assert Ignore OK Detection Tests

Tests for the `type-assert-ignore-ok` checker that detects type assertions
where the `ok` return value is discarded (assigned to `_`).

## How to Run

```sh
doctest test -v ./
```

# Scenario

**Feature**: cold Scan with `NoCache=true` skips all mirror writes

```
# write-disabled cold path
NoCache=true + CacheRoot set
  -> Scan full walk (discovery still runs)
  -> no entry.json under CacheRoot/mirror
```

## Preconditions

- `req.NoCache` is true.
- `CacheRoot` is still set (temp) so a mistaken write would be observable.

## Steps

1. Set `NoCache=true`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.NoCache = true
	return nil
}
```

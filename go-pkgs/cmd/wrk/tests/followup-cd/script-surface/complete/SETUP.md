# Scenario

**Feature**: bash-integration --complete lists flags including --no-cd

```
wrk --bash-integration --complete -- <words> <cword> -> flag candidates
```

## Steps

1. Set Mode to complete.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "complete"
	return nil
}
```

# Scenario

**Feature**: wrk --projects output format

```
wrk --projects -> one absolute path per line, lexicographically sorted
```

## Steps

- Descendants vary whether any projects have been recorded.

```go
func Setup(t *testing.T, req *Request) error {
	ensureProjectsHelpersUsed()
	return nil
}
```
# Scenario

**Feature**: print-script contains completion registration and WRK_HOME resolution

```
wrk --bash-integration -> script references complete -F _wrk, WRK_HOME, --complete
```

## Steps

1. Run `wrk --bash-integration` with default isolated env.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "print")
	requireNoPreseed(t, req)
	return nil
}
```
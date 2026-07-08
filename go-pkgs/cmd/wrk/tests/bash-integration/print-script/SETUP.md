# Scenario

**Feature**: wrk --bash-integration prints bash completion script

```
wrk --bash-integration -> stdout script with complete -F _wrk and --complete callback
```

## Steps

1. Set `req.Mode = "print"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "print"
	req.DryRun = false
	return nil
}
```
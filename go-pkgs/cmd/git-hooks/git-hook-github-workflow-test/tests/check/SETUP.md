## Preconditions

- The command is run in check mode.
- No `--fix` flag is present.

## Steps

1. Leave `req.Args` without `--fix`.
2. Allow each leaf to add origin filters or repository state.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "check mode"
	req.Args = append(req.Args, "--origin-domain=github.com")
	return nil
}
```

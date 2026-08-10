## Preconditions

- The command is run in fix mode.
- The `--fix` flag is present.

## Steps

1. Add `--fix` to the command arguments.
2. Allow each leaf to add origin filters or repository state.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "fix mode"
	req.Args = append(req.Args, "--fix")
	return nil
}
```

## Preconditions

- The user asks for help.

## Steps

1. Run the command with `--help`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "help"
	req.Args = []string{"--help"}
	return nil
}
```

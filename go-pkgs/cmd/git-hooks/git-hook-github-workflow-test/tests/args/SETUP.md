## Preconditions

- Argument parsing behavior is under test.

## Steps

1. Let each leaf select the exact CLI arguments.
2. Run the command and capture output and exit code.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "args"
	req.Args = nil
	return nil
}
```

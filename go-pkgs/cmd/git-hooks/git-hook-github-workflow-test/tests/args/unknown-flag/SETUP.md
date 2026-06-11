## Preconditions

- The user provides an unsupported flag.

## Steps

1. Run the command with `--unknown`.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "unknown flag"
	req.Args = []string{"--unknown"}
	return nil
}
```

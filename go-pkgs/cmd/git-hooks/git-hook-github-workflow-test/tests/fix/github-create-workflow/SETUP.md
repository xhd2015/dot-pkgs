## Preconditions

- The repository origin is `github.com`.
- `.github/workflows/test.yml` does not exist.
- `go.mod` declares Go version `1.22`.

## Steps

1. Keep the default GitHub origin and `go.mod`.
2. Run the command with `--fix`.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "fix github create workflow"
	return nil
}
```

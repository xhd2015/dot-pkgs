## Preconditions

- The repository origin is `github.com`.
- `.github/workflows/test.yml` does not exist.

## Steps

1. Keep the default GitHub origin.
2. Run the command in check mode.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "check github missing workflow"
	return nil
}
```

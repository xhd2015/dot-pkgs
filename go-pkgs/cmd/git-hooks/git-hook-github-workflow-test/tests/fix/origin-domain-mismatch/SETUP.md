## Preconditions

- The repository origin host is `github.com`.
- The `--origin-domain` flag is set to a different host.

## Steps

1. Run the command with `--fix --origin-domain=git.example.com`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "fix origin domain mismatch"
	req.Args = []string{"--fix", "--origin-domain=git.example.com"}
	return nil
}
```

## Preconditions

- The repository origin host is `github.com`.
- The `--origin-domain` flag is set to a different host.

## Steps

1. Replace the inherited `--origin-domain=github.com` argument with `--origin-domain=git.example.com`.
2. Run the command in check mode.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "check origin domain mismatch"
	req.Args = []string{"--origin-domain=git.example.com"}
	return nil
}
```

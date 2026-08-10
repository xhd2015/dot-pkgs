## Preconditions

- The repository origin host is not `github.com`.
- `.github/workflows/test.yml` does not exist.

## Steps

1. Set the remote origin to `git@git.example.com:owner/repo.git`.
2. Run the command with `--fix`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "fix non github origin"
	return setOrigin(req, "git@git.example.com:owner/repo.git")
}
```

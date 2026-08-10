## Preconditions

- The repository origin host is not `github.com`.
- `.github/workflows/test.yml` does not exist.

## Steps

1. Set the remote origin to `git@git.example.com:owner/repo.git`.
2. Run the command in check mode with no origin filter so the hook would otherwise evaluate the repository.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "check non github origin"
	req.Args = nil
	return setOrigin(req, "git@git.example.com:owner/repo.git")
}
```

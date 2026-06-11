## Preconditions

- The repository origin is `github.com`.
- `.github/workflows/test.yml` already exists.

## Steps

1. Create an existing workflow file before running the command.
2. Run the command in check mode.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "check github existing workflow"
	return writeFile(req.RepoDir, ".github/workflows/test.yml", "name: existing\n")
}
```

## Preconditions

- The repository origin is `github.com`.
- `.github/workflows/test.yml` already exists.

## Steps

1. Create an existing workflow file.
2. Run the command with `--fix`.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "fix github existing workflow"
	return writeFile(req.RepoDir, ".github/workflows/test.yml", "name: existing\n")
}
```

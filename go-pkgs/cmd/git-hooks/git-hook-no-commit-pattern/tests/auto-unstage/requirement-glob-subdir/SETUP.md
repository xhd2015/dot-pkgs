# Scenario

**Feature**: `--auto-unstage` with `REQUIREMENT-*.md` unstages requirement docs in subdirectories

```
# stage go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md
# run with --auto-unstage "REQUIREMENT-*.md"
# -> match -> print path -> unstage -> exit 0
git add go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md
  -> hook --auto-unstage "REQUIREMENT-*.md"
  -> go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md matches
  -> print path -> restore --staged -> exit 0
```

## Preconditions

- Repository has an initial commit.
- `go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md` is staged.

## Steps

1. Create an initial commit.
2. Create and stage `go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md`.
3. Run the hook with `--auto-unstage REQUIREMENT-*.md`.

```go
func Setup(t *testing.T, req *Request) error {
	if err := initGitRepoWithCommit(req.RepoDir); err != nil {
		return err
	}
	if err := writeAndStage(req.RepoDir, "go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md", "# design\n"); err != nil {
		return err
	}
	req.Args = []string{"--auto-unstage", "REQUIREMENT-*.md"}
	return nil
}
```
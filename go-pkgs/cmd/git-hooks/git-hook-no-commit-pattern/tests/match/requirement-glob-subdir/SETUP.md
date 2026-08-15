# Scenario

**Feature**: `REQUIREMENT-*.md` matches staged requirement docs in subdirectories

```
# stage go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md, pattern "REQUIREMENT-*.md"
# -> basename matches pattern regardless of directory prefix -> print path -> exit 1
stage go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md
  -> "REQUIREMENT-*.md"
  -> go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md matches
  -> print "go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md"
  -> exit 1
```

## Preconditions

- Files are staged in the repository.

## Steps

1. Create and stage `go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md`.
2. Run the hook with pattern `REQUIREMENT-*.md`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"REQUIREMENT-*.md"}
	if err := writeAndStage(req.RepoDir, "go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md", "# design\n"); err != nil {
		return err
	}
	return nil
}
```
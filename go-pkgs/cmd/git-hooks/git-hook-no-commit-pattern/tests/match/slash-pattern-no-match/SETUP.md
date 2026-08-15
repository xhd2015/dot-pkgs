# Scenario

**Feature**: patterns containing `/` match only the full staged path

```
# stage other/foo.md, pattern "go-pkgs/*.md" -> no match -> exit 0
stage other/foo.md -> "go-pkgs/*.md" -> no match -> exit 0
```

## Preconditions

- Files are staged in the repository.

## Steps

1. Create and stage `other/foo.md`.
2. Run the hook with pattern `go-pkgs/*.md`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"go-pkgs/*.md"}
	if err := writeAndStage(req.RepoDir, "other/foo.md", "# foo\n"); err != nil {
		return err
	}
	return nil
}
```
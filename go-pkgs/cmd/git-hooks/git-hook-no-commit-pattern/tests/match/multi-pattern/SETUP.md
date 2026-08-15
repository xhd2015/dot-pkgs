# Scenario

**Feature**: multiple patterns provided, staged file matches one of them

```
# stage README.md, patterns "REQUIREMENT-* *.md" -> README.md matches *.md -> print -> exit 1
stage README.md -> patterns "REQUIREMENT-*" "*.md" -> *.md matches README.md -> print -> exit 1
```

## Preconditions

- Files are staged in the repository.

## Steps

1. Create and stage `README.md`.
2. Run the hook with two patterns: `REQUIREMENT-*` and `*.md`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"REQUIREMENT-*", "*.md"}
	if err := writeAndStage(req.RepoDir, "README.md", "# Readme\n"); err != nil {
		return err
	}
	return nil
}
```

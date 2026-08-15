# Scenario

**Feature**: `--auto-unstage` combined with `--origin-domain` mismatch still skips

```
# hook --auto-unstage --origin-domain other.com "*.go" -> origin is github.com -> skip -> exit 0, nothing unstaged
--auto-unstage + --origin-domain mismatch -> domain gate skip -> no file checking -> exit 0, no output
```

## Preconditions

- Repository origin is `git@github.com:owner/repo.git` (set by root SETUP).
- A staged file matches the pattern but the domain gate skips.

## Steps

1. Stage `main.go` which matches `*.go`.
2. Run the hook with `--auto-unstage --origin-domain other.com *.go`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	req.Args = []string{"--auto-unstage", "--origin-domain", "other.com", "*.go"}
	return nil
}
```

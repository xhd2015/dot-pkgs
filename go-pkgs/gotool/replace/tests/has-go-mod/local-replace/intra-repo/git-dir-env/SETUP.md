# Scenario

**Feature**: inherited Git hook environment does not affect intra-repo replace classification

```
# GIT_DIR from a hook process + nested go.mod replace => ../ -> still same repo
repo/.git exported as GIT_DIR -> go-pkgs/cmd/go.mod replace old => ../ -> target go-pkgs is intra-repo
```

## Preconditions

- The test repository has a nested module at `go-pkgs/cmd`.
- That nested module has `replace github.com/xhd2015/dot-pkgs/go-pkgs => ../`.
- `../` resolves to the `go-pkgs` directory inside the same repository.
- `GIT_DIR` is set like Git does when running hooks.

## Steps

1. Create `go-pkgs/go.mod` and `go-pkgs/cmd/go.mod`.
2. Call `replace.CheckLocalReplaces(repo)`.
   Do not `t.Setenv("GIT_DIR")`: doctest leaves always call `t.Parallel()`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := writeGoMod(req.RootDir, "go-pkgs/go.mod", "module github.com/xhd2015/dot-pkgs/go-pkgs\n\ngo 1.22\n"); err != nil {
		return err
	}
	if err := writeGoMod(req.RootDir, "go-pkgs/cmd/go.mod", "module github.com/xhd2015/dot-pkgs/go-pkgs/cmd\n\ngo 1.22\n\nrequire github.com/xhd2015/dot-pkgs/go-pkgs v0.0.0\n\nreplace github.com/xhd2015/dot-pkgs/go-pkgs => ../\n"); err != nil {
		return err
	}
	return nil
}

```

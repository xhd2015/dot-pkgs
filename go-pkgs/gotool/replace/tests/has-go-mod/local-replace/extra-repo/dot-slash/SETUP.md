# Scenario

**Feature**: `./` replace target outside git repo (target directory does not exist)

```
# replace old => ./nonexistent -> path still under scan root even if missing
go.mod -> replace old => ./nonexistent -> target missing, still under tree -> IsIntraRepo = true
```

## Preconditions

- A root go.mod exists with `replace example.com/old => ./nonexistent`.
- The `./nonexistent` directory does NOT exist.

## Steps

1. Write `go.mod` with a `./` local replace to a nonexistent directory.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	content := "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => ./nonexistent\n"
	return writeGoMod(req.RootDir, "go.mod", content)
}
```
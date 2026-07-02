# Scenario

**Feature**: `./` replace target outside git repo (target directory does not exist)

```
# replace old => ./nonexistent -> ./nonexistent does not exist -> os.Stat fails -> extra-repo
go.mod -> replace old => ./nonexistent -> target does not exist -> IsIntraRepo = false
```

## Preconditions

- A root go.mod exists with `replace example.com/old => ./nonexistent`.
- The `./nonexistent` directory does NOT exist.

## Steps

1. Write `go.mod` with a `./` local replace to a nonexistent directory.

```go
func Setup(t *testing.T, req *Request) error {
	content := "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => ./nonexistent\n"
	return writeGoMod(req.RootDir, "go.mod", content)
}
```
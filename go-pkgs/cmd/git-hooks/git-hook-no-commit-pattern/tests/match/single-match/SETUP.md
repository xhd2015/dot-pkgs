# Scenario

**Feature**: one staged file matches the pattern

```
# stage main.go, pattern "*.go" -> match main.go -> print -> exit 1
stage main.go -> "*.go" -> main.go matches -> print "main.go" -> exit 1
```

## Preconditions

- Files are staged in the repository.

## Steps

1. Create and stage `main.go`.
2. Run the hook with pattern `*.go`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"*.go"}
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	return nil
}
```

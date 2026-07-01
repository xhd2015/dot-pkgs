# Scenario

**Feature**: multiple staged files match the pattern

```
# stage main.go test.go, pattern "*.go" -> match both -> print each -> exit 1
stage main.go test.go -> "*.go" -> main.go matches, test.go matches -> print both -> exit 1
```

## Preconditions

- Multiple files are staged in the repository.

## Steps

1. Create and stage `main.go` and `test.go`.
2. Run the hook with pattern `*.go`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"*.go"}
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	if err := writeAndStage(req.RepoDir, "test.go", "package test\n"); err != nil {
		return err
	}
	return nil
}
```

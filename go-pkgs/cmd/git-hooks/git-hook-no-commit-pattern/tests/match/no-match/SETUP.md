# Scenario

**Feature**: staged files exist but none match the pattern

```
# stage .go files, pattern *.md -> no match -> exit 0
write main.go test.go -> stage -> pattern "*.md" -> no match -> exit 0
```

## Preconditions

- A root go.mod exists with no replace directives.

## Steps

1. Create and stage `main.go` and `test.go`.
2. Run the hook with pattern `*.md`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"*.md"}
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	if err := writeAndStage(req.RepoDir, "test.go", "package test\n"); err != nil {
		return err
	}
	return nil
}
```

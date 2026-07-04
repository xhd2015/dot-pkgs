# Scenario

**Feature**: slashless `.vscode` matches any path segment, including middle segments

```
# stage vendor/.vscode/extensions.json, pattern ".vscode"
# -> middle segment ".vscode" matches -> print path -> exit 1
stage vendor/.vscode/extensions.json -> ".vscode" -> match -> exit 1
```

## Preconditions

- Files are staged in the repository.

## Steps

1. Create and stage `vendor/.vscode/extensions.json`.
2. Run the hook with pattern `.vscode`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{".vscode"}
	if err := writeAndStage(req.RepoDir, "vendor/.vscode/extensions.json", "{}\n"); err != nil {
		return err
	}
	return nil
}
```
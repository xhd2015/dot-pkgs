# Scenario

**Feature**: gotool Replace adds filesystem replace directive

```
# consumer requires dep; Replace(depDir) -> go mod edit -replace mod=absDir
consumer (require dep) + dep module dir -> replace.Replace -> go.mod has replace
```

## Steps

1. Create dep git module `example.com/dep`.
2. Create consumer module requiring `example.com/dep`.
3. Set operation to `replace`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	workspace := newWorkspace(t)
	depDir := initDepModuleRepo(t, workspace, depModulePath)
	consumer := initConsumerModule(t, workspace, true)

	req.Operation = "replace"
	req.ConsumerDir = consumer
	req.TargetDir = depDir
	return nil
}
```
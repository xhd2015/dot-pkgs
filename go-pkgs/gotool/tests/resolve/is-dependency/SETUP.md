# Scenario

**Feature**: resolve detects module listed in consumer require

```
# consumer go.mod requires dep -> ResolveLocalModules -> IsDependency true
```

## Steps

1. Create dep module repo.
2. Create consumer with `-require example.com/dep`.
3. Set operation to `resolve`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	workspace := newWorkspace(t)
	depDir := initDepModuleRepo(t, workspace, depModulePath)
	consumer := initConsumerModule(t, workspace, true)

	req.Operation = "resolve"
	req.ConsumerDir = consumer
	req.LocalModDir = depDir
	return nil
}
```
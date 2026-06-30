# Scenario

**Feature**: resolve reports false when module is not in consumer go.mod

```
# consumer without dep require -> ResolveLocalModules -> IsDependency false
```

## Steps

1. Create dep module repo.
2. Create consumer module **without** requiring dep.
3. Set operation to `resolve`.

```go
func Setup(t *testing.T, req *Request) error {
	workspace := newWorkspace(t)
	depDir := initDepModuleRepo(t, workspace, depModulePath)
	consumer := initConsumerModule(t, workspace, false)

	req.Operation = "resolve"
	req.ConsumerDir = consumer
	req.LocalModDir = depDir
	return nil
}
```
# Scenario

**Feature**: Pin errors when DepDir has no version tag and Version is empty

```
# untagged DepDir + Version="" -> Pin error (no tag); go.mod unchanged
```

## Steps

1. Create untagged fixture module repo (no `vN.N.N` tags).
2. Create consumer with require + replace.
3. Pin with empty Version (must resolve latest tag → fail).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	workspace := newWorkspace(t)
	fixtureDir := initUntaggedFixtureRepo(t, workspace)
	consumer := initConsumerWithReplace(t, workspace, fixtureDir)

	req.Operation = "pin"
	req.ConsumerDir = consumer
	req.TargetDir = fixtureDir
	req.Version = ""
	req.DryRun = false
	return nil
}
```

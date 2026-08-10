# Scenario

**Feature**: gotool Update drops replace and pins require to latest git tag

```
# consumer with replace -> Update(fixtureDir) -> require v1.0.0, no replace
```

## Steps

1. Create tagged fixture module repo (`v1.0.0` tag, post-tag commit on HEAD).
2. Create consumer with replace pointing at fixture dir.
3. Set operation to `update`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	workspace := newWorkspace(t)
	fixtureDir := initTaggedFixtureRepo(t, workspace)
	consumer := initConsumerWithReplace(t, workspace, fixtureDir)

	req.Operation = "update"
	req.ConsumerDir = consumer
	req.TargetDir = fixtureDir
	return nil
}
```
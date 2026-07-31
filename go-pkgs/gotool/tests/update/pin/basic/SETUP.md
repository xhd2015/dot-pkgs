# Scenario

**Feature**: Pin drops replace and pins require to latest git tag

```
# consumer require@v0.0.1 + replace -> Pin(ConsumerDir, DepDir) -> require v1.0.0, no replace
```

## Steps

1. Create tagged fixture module repo (`v1.0.0` tag, post-tag commit on HEAD).
2. Create consumer with require `v0.0.1` and replace pointing at fixture dir.
3. Set operation to `pin` (empty Version = latest tag; DryRun false).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	workspace := newWorkspace(t)
	fixtureDir := initTaggedFixtureRepo(t, workspace)
	consumer := initConsumerWithReplace(t, workspace, fixtureDir)

	req.Operation = "pin"
	req.ConsumerDir = consumer
	req.TargetDir = fixtureDir
	req.Version = ""
	req.DryRun = false
	return nil
}
```

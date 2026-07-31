# Scenario

**Feature**: Pin DryRun plans pin without writing consumer go.mod

```
# would pin to v1.0.0; DryRun=true -> PinResult filled; require still v0.0.1; replace remains
```

## Steps

1. Create tagged fixture (`v1.0.0`) and consumer with require `v0.0.1` + replace.
2. Set `DryRun=true`, empty Version (latest tag plan).

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
	req.DryRun = true
	return nil
}
```

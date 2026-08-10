# Scenario

**Feature**: Pin with explicit Version forces require version (may skip tag lookup)

```
# consumer require@v0.0.1 + replace -> Pin(Version=v0.0.5) -> require v0.0.5, no replace
# DepDir has tag v1.0.0 only — forced Version must not require that tag to match
```

## Steps

1. Create tagged fixture (`v1.0.0`) so module path/tags exist if product still looks them up.
2. Create consumer with require + replace.
3. Pin with `Version=v0.0.5` (valid semver form; not the latest tag).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	workspace := newWorkspace(t)
	fixtureDir := initTaggedFixtureRepo(t, workspace)
	consumer := initConsumerWithReplace(t, workspace, fixtureDir)

	req.Operation = "pin"
	req.ConsumerDir = consumer
	req.TargetDir = fixtureDir
	req.Version = "v0.0.5"
	req.DryRun = false
	return nil
}
```

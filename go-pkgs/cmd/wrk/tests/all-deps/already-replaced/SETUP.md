# Scenario

**Feature**: wrk --all-deps tolerates and skips dependencies already replaced in go.mod

```
# consumer requires dep1+dep2; dep1 pre-replaced => ./external/preexisting; scan-root has both -> dep1 skipped, dep2 linked
consumer (dep1 pre-replaced => ./external/preexisting, requires dep1+dep2) + scan-root (mydep1, mydep2) -> wrk --all-deps -> dep2 external wt + dep1 replace unchanged + wrk 1 deps
```

## Steps

1. Create a scan-root holding `mydep1` (module `example.com/dep1`) and `mydep2` (module `example.com/dep2`).
2. Create a consumer git repo requiring both deps, with a pre-existing `replace example.com/dep1 => ./external/preexisting`.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)
	initAllDepsRepo(t, filepath.Join(scanRoot, "mydep1"), "example.com/dep1", "dep1")
	initAllDepsRepo(t, filepath.Join(scanRoot, "mydep2"), "example.com/dep2", "dep2")

	extraGoMod := "replace example.com/dep1 => ./external/preexisting"
	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, extraGoMod)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}
```

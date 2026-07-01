# Scenario

**Feature**: wrk --all-deps --dry-run prints the plan for every matched dep but leaves the filesystem and go.mod untouched

```
# consumer requires dep1+dep2; scan-root has mydep1 + mydep2 -> dry-run prints would: lines, no side effects
consumer (requires dep1, dep2) + scan-root (mydep1=example.com/dep1, mydep2=example.com/dep2) -> wrk --all-deps --dry-run -> would: wrk example.com/dep1 at ./external/mydep1-main-2026-06-30 / would: wrk example.com/dep2 at ./external/mydep2-main-2026-06-30 / would: wrk 2 deps
```

## Steps

1. Create a scan-root temp dir holding two dep repos: `mydep1` (module `example.com/dep1`) and `mydep2` (module `example.com/dep2`).
2. Create a consumer git repo requiring both `example.com/dep1` and `example.com/dep2`.
3. Run `wrk --all-deps --dry-run --scan-root <scanRoot>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)
	initAllDepsRepo(t, filepath.Join(scanRoot, "mydep1"), "example.com/dep1", "dep1")
	initAllDepsRepo(t, filepath.Join(scanRoot, "mydep2"), "example.com/dep2", "dep2")

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--dry-run", "--scan-root", scanRoot}
	return nil
}
```

# Scenario

**Feature**: wrk --all-deps skips required dependencies that have no local repo in the scan root

```
# consumer requires dep1+dep2; scan-root has only mydep1 -> dep1 linked, dep2 not replaced
consumer (requires dep1, dep2) + scan-root (mydep1=example.com/dep1) -> wrk --all-deps -> dep1 external wt + 1 replace + wrk 1 deps
```

## Steps

1. Create a scan-root holding only `mydep1` (module `example.com/dep1`).
2. Create a consumer git repo requiring both `example.com/dep1` and `example.com/dep2`.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)
	initAllDepsRepo(t, filepath.Join(scanRoot, "mydep1"), "example.com/dep1", "dep1")

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}
```

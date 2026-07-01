# Scenario

**Feature**: wrk --all-deps skips the consumer's own module when the consumer lives inside the scan root

```
# consumer inside scan-root alongside mydep1; consumer requires dep1 -> dep1 linked, consumer's own module skipped
scan-root (consumer=example.com/consumer, mydep1=example.com/dep1) -> wrk --all-deps -> dep1 external wt + 1 replace + wrk 1 deps (no self-replace)
```

## Steps

1. Create a scan-root holding `mydep1` (module `example.com/dep1`).
2. Create the consumer git repo (module `example.com/consumer`) **inside** the scan-root, requiring `example.com/dep1`.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)
	initAllDepsRepo(t, filepath.Join(scanRoot, "mydep1"), "example.com/dep1", "dep1")

	// consumer lives INSIDE the scan-root so the scan also discovers it; its own
	// module path (example.com/consumer) must be skipped.
	consumer := initAllDepsConsumer(t, scanRoot, []string{"example.com/dep1"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}
```

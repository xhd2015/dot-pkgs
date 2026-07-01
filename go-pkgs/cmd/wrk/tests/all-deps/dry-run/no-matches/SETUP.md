# Scenario

**Feature**: wrk --all-deps --dry-run with no matching local repos prints would: wrk 0 deps and makes no changes

```
# consumer requires dep1+dep2; scan-root empty -> dry-run prints would: wrk 0 deps, no side effects
consumer (requires dep1, dep2) + empty scan-root -> wrk --all-deps --dry-run -> would: wrk 0 deps (exit 0)
```

## Steps

1. Create an empty scan-root dir.
2. Create a consumer git repo requiring `example.com/dep1` and `example.com/dep2`.
3. Run `wrk --all-deps --dry-run --scan-root <scanRoot>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--dry-run", "--scan-root", scanRoot}
	return nil
}
```

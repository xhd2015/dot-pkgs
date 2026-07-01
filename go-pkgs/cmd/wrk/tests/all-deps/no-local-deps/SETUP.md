# Scenario

**Feature**: wrk --all-deps with no matching local repos prints wrk 0 deps and makes no changes

```
# consumer requires dep1+dep2; scan-root empty -> wrk --all-deps -> wrk 0 deps, no replaces, no external/
consumer (requires dep1, dep2) + empty scan-root -> wrk --all-deps -> wrk 0 deps (exit 0)
```

## Steps

1. Create an empty scan-root dir.
2. Create a consumer git repo requiring `example.com/dep1` and `example.com/dep2`.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}
```

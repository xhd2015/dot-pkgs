# Scenario

**Feature**: wrk --all-deps links every required dependency that has a local repo in the scan root

```
# consumer requires dep1+dep2; scan-root has mydep1 + mydep2 -> both linked, replaced, tidied once
consumer (requires dep1, dep2) + scan-root (mydep1=example.com/dep1, mydep2=example.com/dep2) -> wrk --all-deps -> 2 external wts + 2 replaces + /external once
```

## Steps

1. Create a scan-root temp dir holding two dep repos: `mydep1` (module `example.com/dep1`) and `mydep2` (module `example.com/dep2`).
2. Create a consumer git repo requiring both `example.com/dep1` and `example.com/dep2`.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the consumer.

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
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}
```

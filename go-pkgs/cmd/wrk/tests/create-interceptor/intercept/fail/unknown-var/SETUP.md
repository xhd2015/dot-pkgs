# Scenario

**Feature**: unknown template var fails hard before exec

```
argv refs ${no_such} -> non-zero; clear error; fake not run; no worktree
```

## Steps

1. Override config with `argv: ["kool", "${no_such}"]`.
2. Fake remains installed so a mistaken exec would log.
3. Run bare create.

```go
func Setup(t *testing.T, req *Request) error {
	writeInterceptorConfig(t, req.WrkHome, true, []string{fakeInterceptorName, "${no_such}"}, nil)
	return nil
}
```

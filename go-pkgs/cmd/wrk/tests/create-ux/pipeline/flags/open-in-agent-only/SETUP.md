# Scenario

**Feature**: `--open-in-agent` without terminal runs agent-run in current process

```
wrk -t 'ship feature' --open-in-agent
  -> create; agent-run cwd=wt; no space; no iterm
```

## Steps

1. Run with task + `--open-in-agent`.

```go
func Setup(t *testing.T, req *Request) error {
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent"}
	return nil
}
```

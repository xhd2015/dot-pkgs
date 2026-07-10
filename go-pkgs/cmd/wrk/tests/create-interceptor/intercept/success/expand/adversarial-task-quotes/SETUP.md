# Scenario

**Feature**: `${intent_prompt|shell_safe}` encodes double quotes in task text

```
wrk -t 'fix "quoted" task'
  -> intent_prompt raw: /intent-route fix "quoted" task
  -> shell_safe one word; quotes do not break --send payload
```

## Steps

1. Recipe config from expand grouping.
2. Set task description containing double quotes.
3. Run create with `-t`.

```go
func Setup(t *testing.T, req *Request) error {
	req.TaskDesc = `fix "quoted" task`
	req.TaskFlag = "-t"
	return nil
}
```

# Scenario

**Feature**: JSON string-array vars expand then join with newline for use sites

```
vars.send = [ "line-one-…", "line-two-…" ]
argv … --send ${send}
  -> fake --send arg contains "line-one…\nline-two…"
```

## Steps

1. Recipe config already defines `send` as a two-element array.
2. Use a simple task without quotes so the multiline join is the focus.
3. Run create with `-t`.

```go
func Setup(t *testing.T, req *Request) error {
	req.TaskDesc = "hello world"
	req.TaskFlag = "-t"
	return nil
}
```

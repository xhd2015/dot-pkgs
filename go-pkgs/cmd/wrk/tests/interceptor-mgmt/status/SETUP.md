# Scenario

**Feature**: wrk --interceptor --status reports absent|disabled|enabled

```
wrk --interceptor --status
  -> state: absent|disabled|enabled
  -> path: <abs-config-path>
  -> argv0: <first argv or ->
```

## Preconditions

- Config presence and `enabled` determine `state`.

## Steps

- Leaves seed or omit interceptor block; run `--status`.

## Context

- Exactly three stdout lines, each ending with `\n`.
- `argv0` is `-` when absent or argv empty.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--status")
	return nil
}
```

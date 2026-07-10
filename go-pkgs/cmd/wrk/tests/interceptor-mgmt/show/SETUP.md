# Scenario

**Feature**: wrk --interceptor --show prints create.interceptor JSON

```
wrk --interceptor --show
  -> present: pretty JSON of create.interceptor + \n; exit 0
  -> absent: non-zero; useful stderr; empty or non-JSON stdout
```

## Steps

- Leaves seed or omit interceptor; run `--show`.

## Context

- Pretty-printed JSON (indent) of the interceptor object only, not the whole config file.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--show")
	return nil
}
```

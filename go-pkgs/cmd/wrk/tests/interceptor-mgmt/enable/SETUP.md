# Scenario

**Feature**: wrk --interceptor --enable sets enabled=true

```
wrk --interceptor --enable
  -> no block: non-zero (hint --init)
  -> present: set enabled true; empty stdout
```

## Steps

- Leaves seed or omit interceptor; run `--enable`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--enable")
	return nil
}
```

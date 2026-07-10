# Scenario

**Feature**: wrk --interceptor --disable sets enabled=false

```
wrk --interceptor --disable
  -> present enabled true: set false; empty stdout
```

## Steps

- Leaf seeds enabled interceptor; run `--disable`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--disable")
	return nil
}
```

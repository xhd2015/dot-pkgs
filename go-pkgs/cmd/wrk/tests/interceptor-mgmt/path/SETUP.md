# Scenario

**Feature**: wrk --interceptor --path prints absolute config.json path

```
wrk --interceptor --path -> stdout {WRK_HOME}/config.json abs path + \n
```

## Preconditions

- Parent provides neutral cwd and helpers.

## Steps

- Leaf runs `--path` whether or not the file exists.

## Context

- Exit 0 even when config.json is missing.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--path")
	return nil
}
```

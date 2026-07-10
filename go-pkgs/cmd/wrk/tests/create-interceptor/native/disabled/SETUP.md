# Scenario

**Feature**: interceptor present but `enabled: false`

```
config.json create.interceptor.enabled=false + fake kool on PATH
  -> native create; kool not exec'd
```

## Steps

- Write config with `enabled: false` and a non-empty argv pointing at fake `kool`.
- Install fake so a mistaken intercept would leave a log.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```

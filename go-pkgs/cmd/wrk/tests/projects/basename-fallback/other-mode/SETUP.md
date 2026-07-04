# Scenario

**Feature**: non-create modes skip basename fallback

```
# saved project exists but mode is not create -> no lookup
wrk <basename> --list -> normal missing-dir error
```

## Steps

- Descendants record a saved project and invoke a non-create mode with basename `<dir>`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureBasenameFallbackHelpersUsed()
	return nil
}
```
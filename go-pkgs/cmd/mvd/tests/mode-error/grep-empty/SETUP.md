# Scenario

**Feature**: mvd --grep with empty pattern requires a non-empty filter

```
# present-but-empty --grep (absent ≠ present empty)
mvd --grep '' -> non-zero; error requires non-empty pattern
```

## Steps

1. Run `mvd --grep` with an empty string value (flag present, value empty).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--grep", ""}
	return nil
}
```

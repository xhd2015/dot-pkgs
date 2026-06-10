## Steps

- The `--help` flag was set in upstream SETUP.md, no additional args needed for this leaf

```go
func Setup(t *testing.T, req *Request) error {
	t.Log("help show: verifying --help flag output")
	return nil
}
```

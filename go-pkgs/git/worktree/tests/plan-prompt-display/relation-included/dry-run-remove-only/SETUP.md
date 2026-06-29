# Scenario

**Feature**: included branch dry-run — remove commands only

```go
func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	return nil
}
```
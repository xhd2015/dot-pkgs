# Scenario

**Feature**: Contents API lookup

```
Contents(sessionID, cfg) -> ContentsResult or error
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "contents"
	return nil
}
```

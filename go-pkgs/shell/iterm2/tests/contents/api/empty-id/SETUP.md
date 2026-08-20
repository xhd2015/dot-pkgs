# Scenario

**Feature**: empty session id errors before osascript

```
Contents("") -> error
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "contents"
	req.SessionID = ""
	req.HomeApp = true
	return nil
}
```

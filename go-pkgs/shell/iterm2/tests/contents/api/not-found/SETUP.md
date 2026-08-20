# Scenario

**Feature**: no install has the session

```
Contents -> session not found
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "contents"
	req.SessionID = defaultContentsSessionID
	req.HomeApp = true
	req.NotFoundApps = []string{"/Users/me/Applications/iTerm.app"}
	return nil
}
```

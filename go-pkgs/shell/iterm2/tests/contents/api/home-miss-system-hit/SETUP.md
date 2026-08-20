# Scenario

**Feature**: home not found then system hit

```
Contents: home tell not found -> system tell body, app=/Applications/iTerm.app
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "contents"
	req.SessionID = defaultContentsSessionID
	req.Body = "system pane"
	req.HomeApp = true
	req.SysApp = true
	req.NotFoundApps = []string{"/Users/me/Applications/iTerm.app"}
	return nil
}
```

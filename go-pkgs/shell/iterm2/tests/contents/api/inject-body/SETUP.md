# Scenario

**Feature**: injected Exec body is returned; home app wins first

```
Contents(uuid) home running -> body, app=~/Applications/iTerm.app
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "contents"
	req.SessionID = defaultContentsSessionID
	req.Body = "❯ hello pane"
	req.HomeApp = true
	req.SysApp = true
	return nil
}
```

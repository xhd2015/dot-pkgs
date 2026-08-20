# Scenario

**Feature**: not-running home is skipped (does not launch)

```
home exists but not running, system running -> only system Exec
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "contents"
	req.SessionID = defaultContentsSessionID
	req.Body = "sys"
	req.HomeApp = true
	req.SysApp = true
	req.Running = []string{"/Applications/iTerm.app"}
	return nil
}
```

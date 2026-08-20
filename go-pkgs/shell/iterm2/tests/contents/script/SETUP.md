# Scenario

**Feature**: BuildContentsScript emits a no-focus contents dump

```
BuildContentsScript(uuid, appPath) -> tell + contents of aSession
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-script"
	if req.SessionID == "" {
		req.SessionID = defaultContentsSessionID
	}
	return nil
}
```

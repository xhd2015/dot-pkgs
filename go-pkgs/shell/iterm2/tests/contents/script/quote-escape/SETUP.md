# Scenario

**Feature**: UUID quotes are escaped in the contents script

```
BuildContentsScript(`aa"bb`) -> escaped \" in AppleScript
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-script"
	req.SessionID = `aa"bb`
	return nil
}
```

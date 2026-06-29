# Scenario

**Feature**: ErrNotInstalled when injectable Installed is false

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = t.TempDir()
	req.UseInstalledOverride = true
	req.InstalledOK = false
	return nil
}
```
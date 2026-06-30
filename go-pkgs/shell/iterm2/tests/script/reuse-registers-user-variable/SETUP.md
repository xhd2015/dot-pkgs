# Scenario

**Feature**: reuse miss branch registers `user.koolTargetDir` for back-to-back reuse

```
caller dir -> BuildReuseCurrentSessionScript -> miss branch: cd + set user.koolTargetDir
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-register")
	req.Mode = "reuse"
	return nil
}
```
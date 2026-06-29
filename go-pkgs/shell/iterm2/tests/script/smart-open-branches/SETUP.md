# Scenario

**Feature**: smart-open branches in generated script

```
BuildScript -> scan paths -> create tab OR create window
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-doctest-proj")
	return nil
}
```
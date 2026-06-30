# Scenario

**Feature**: reuse match branch selects matchingWindow before tab/session focus

```
BuildReuseCurrentSessionScript -> match: select matchingWindow -> select tab/session
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-match-window")
	req.Mode = "reuse"
	return nil
}
```
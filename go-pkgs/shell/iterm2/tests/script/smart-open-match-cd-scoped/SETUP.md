# Scenario

**Feature**: smart-open match branch scopes cd to matchingWindow's new tab

```
BuildScript -> match: create tab in matchingWindow -> cd in that tab (not frontmost)
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-smart-match-cd")
	return nil
}
```
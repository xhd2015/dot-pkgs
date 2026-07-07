# Scenario

**Feature**: summary includes codex lines when present; omits grok when absent

```
FormatSummaryLines -> codex sessions/skills when HasCodex; no grok lines when HasGrok false
```

## Steps

1. Set `Mode = format-summary`.
2. Build summary with `HasCodex=true`, `HasGrok=false`.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/analyse"
)

func Setup(t *testing.T, req *Request) error {
	req.Mode = "format-summary"
	req.Summary = analyse.ScanSummary{
		Home:          "/tmp/home",
		EntryCount:    1,
		DirCount:      1,
		TotalHuman:    "1 KB",
		HasCodex:      true,
		CodexSessions: 2,
		CodexSkills:   1,
		HasGrok:       false,
	}
	return nil
}
```
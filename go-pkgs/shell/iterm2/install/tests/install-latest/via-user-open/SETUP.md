# Scenario

**Feature**: `InstallLatest` with optional user-driven handoff
(`InstallViaUserOpen` + injectable `Open` / `ClearQuarantineFn`)

```
InstallViaUserOpen=false → place + register + verify; Open/Clear never called
InstallViaUserOpen=true  → extract staged → open staged app; no place/clear/register
```

## Steps

1. Inherit `Operation=install-latest` from parent.
2. Use the same fake-HTTP zip fixture as `pipeline-fake-http`.
3. Leaves set `InstallViaUserOpen` / `OpenShouldFail`.
4. Root Run always injects recording `Open` + `ClearQuarantineFn` (no real
   `open` / `xattr`; no env mutation).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Parent already sets Operation=install-latest + SkipScriptable.
	req.HTTPMode = "install-latest"
	req.FinalZipName = "iTerm2-3_6_11.zip"
	req.TargetApp = "" // place path only used when InstallViaUserOpen=false
	req.SkipScriptable = true
	return nil
}
```

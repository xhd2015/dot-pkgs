# Scenario

**Live e2e**: real Vision OCR on the clipboard sample PNG.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	caseDir := d.DOCTEST_CASE
	if !filepath.IsAbs(caseDir) {
		caseDir = filepath.Join(d.DOCTEST_ROOT, caseDir)
	}
	req.ImagePath = filepath.Join(caseDir, "testdata", "clipboard-sample.png")
	return nil
}
```

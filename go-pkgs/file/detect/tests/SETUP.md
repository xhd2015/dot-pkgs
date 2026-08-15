# Scenario

**Feature**: `detect.DetectFileType` classifies files as text or binary from a path

```
# direct detect — no HTTP/CLI
caller path -> DetectFileType -> (desc, isBinary, err)
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/file/detect` is importable.
- Leaves set `req.Path` to either leaf `testdata/` fixtures or temp files written in Setup.
- Do not special-case path suffixes (e.g. `*.snapshot.txt`); assertions only check
  `DetectFileType` return values.

## Context

- Bug repro: UTF-8 TTY snapshots were misclassified as binary when the 512-byte sniff
  window split a multi-byte rune and/or the first byte was ≥ 0x80.
- Root `Run` always calls `detect.DetectFileType(req.Path)`.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func writeTempFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}
```

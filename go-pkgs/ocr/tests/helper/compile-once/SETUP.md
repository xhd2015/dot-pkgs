# Scenario

**Feature**: fake swiftc compiles once; second RecognizeFile reuses cache

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.CacheDir = t.TempDir()
	swiftc, _ := writeFakeSwiftc(t)
	req.Swiftc = swiftc
	return nil
}
```

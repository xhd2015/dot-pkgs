# Scenario

Short UTF-8 Chinese text is under SafeMax → CheckWriteText.OK.

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "check"
	req.CheckInput = "你好世界 — UTF-8 中文测试"
	return nil
}
```

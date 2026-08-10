# Scenario

Open-inject-shaped multi-line string (<<'EOF', __seq_, 中文) long enough to SoftExceed.

## Steps

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "check"
	var b strings.Builder
	b.WriteString("SeaTalk local-bot session open\n")
	b.WriteString("Reply: use <<'EOF' for multiple lines\n")
	b.WriteString("First message from master:\n")
	mid := "u7oirumAclDLrBQB-RCY350DvXNAhNTU8hLcwlSxv3iceGxuZ7eY6jtI"
	b.WriteString("[image] /tmp/media/by-id/" + mid + "__seq_1.png\n")
	b.WriteString("[image] /tmp/media/by-id/" + mid + "__seq_2.png\n")
	b.WriteString("我看下 @Mad Max\n")
	for b.Len() <= applescript.WriteTextSoftMaxBytes {
		b.WriteString(strings.Repeat("字", 40) + "\n")
	}
	req.CheckInput = b.String()
	return nil
}
```

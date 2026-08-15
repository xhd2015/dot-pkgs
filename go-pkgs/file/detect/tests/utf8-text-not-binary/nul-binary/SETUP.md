# Scenario

**Feature**: NUL byte in content forces binary classification

```
"hello\x00world" -> DetectFileType -> isBinary=true
```

## Steps

1. Write a temp file containing a mid-stream NUL.
2. Set `req.Path` to that file.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Path = writeTempFile(t, "with-nul.bin", []byte("hello\x00world"))
	return nil
}
```

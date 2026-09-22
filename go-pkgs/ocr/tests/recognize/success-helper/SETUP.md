# Scenario

**Feature**: RecognizeFile with injected HelperPath returns helper stdout

```
RecognizeFile(img, HelperPath=fake) -> Text="hello vision", Engine=vision
```

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	img := filepath.Join(t.TempDir(), "in.png")
	if err := os.WriteFile(img, []byte("png"), 0644); err != nil {
		t.Fatal(err)
	}
	req.ImagePath = img
	req.HelperPath = writeFakeHelper(t, "hello vision\n")
	return nil
}
```

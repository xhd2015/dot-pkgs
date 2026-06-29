## Expected

- Prompt uses `master` in question and `# master: fast forward`.
- Must not contain `merge into main?` or `# main: fast forward`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(resp.Output, "merge into main?") || strings.Contains(resp.Output, "# main: fast forward") {
		t.Fatalf("output must use master label, got:\n%s", resp.Output)
	}
	assert.Output(t, resp.Output, `
<contains>
branch feature is ahead, merge into master?
  # master: fast forward
</contains>`)
}
```
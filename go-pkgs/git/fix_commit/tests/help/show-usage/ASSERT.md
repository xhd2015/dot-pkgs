## Expected

- Exit 0. Stderr empty.
- Stdout is the locked usage (all flags, including `-h` as the help spelling)
  and ends with a newline.

## Expected Output

```
Usage: kool git fix-commit <sha> [OPTIONS]

  -m, --message <msg>     replace the full commit message
  --name <name>           replace the author name
  --email <email>         replace the author email
  --strip-co-author       remove Co-authored-by lines from the message;
                          errors if none are present
  --remote <name>         remote for tag delete/push (default: origin)
  --push                  also force-with-lease push updated branches
                          whose upstream still points at the old sha
  --dry-run               print the plan; do not rewrite refs or remotes
  -C, --dir <dir>         repository directory (default: current directory)
  -h, --help              show this help
```

## Errors

- Harness `err` is nil.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	requireHarnessOK(t, err)
	requireExit(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("stderr=%q want empty", resp.Stderr)
	}
	assertOutput(t, resp.Stdout, `Usage: kool git fix-commit <sha> [OPTIONS]

  -m, --message <msg>     replace the full commit message
  --name <name>           replace the author name
  --email <email>         replace the author email
  --strip-co-author       remove Co-authored-by lines from the message;
                          errors if none are present
  --remote <name>         remote for tag delete/push (default: origin)
  --push                  also force-with-lease push updated branches
                          whose upstream still points at the old sha
  --dry-run               print the plan; do not rewrite refs or remotes
  -C, --dir <dir>         repository directory (default: current directory)
  -h, --help              show this help
`)
}
```

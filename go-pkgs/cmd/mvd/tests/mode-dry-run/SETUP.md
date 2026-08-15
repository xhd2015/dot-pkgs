## Steps
- All tests in this mode exercise `mvd --dry-run` combined with various subcommands.
- `--dry-run` is a boolean flag that, when set, skips all actual modifications but still runs path resolution and validation.
- Modification commands print `dry-run: would <action>` and return nil (exit 0).
- Read-only commands (`--list`, `--which`, `--picker-list`, `--grep`, `--print`) are unaffected and produce normal output.
- Validation errors still surface with non-zero exit codes, regardless of `--dry-run`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    t.Logf("mode: dry-run")
    return nil
}
```

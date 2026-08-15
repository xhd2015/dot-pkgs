# Scenario

**Feature**: mvd --grep with empty pattern requires a non-empty filter

```
# present-but-empty --grep (absent ≠ present empty)
mvd --grep '' -> non-zero; error requires non-empty pattern
```

## Steps

1. Run `mvd --grep` with an empty string value (flag present, value empty).

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--grep", ""}
	return nil
}
```

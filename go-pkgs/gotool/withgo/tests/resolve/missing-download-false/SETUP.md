# Scenario

**Feature**: missing dest with Download false is an error; Install unused

```
# dest $InstallDir/go1.19.13 does not exist
missing dest + Download:false -> ResolveGoroot -> error; Install not called
```

## Steps

1. Leave dest missing. Set `Download` false.
2. Set a Prompt that must not appear (install will not run).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Download = false
	req.Prompt = "should-not-print\n"
	return nil
}
```

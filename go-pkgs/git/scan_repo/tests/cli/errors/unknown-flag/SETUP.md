# Scenario

**Feature**: unrecognized CLI flag

```
RunCLI --unknown -> flag parse error on stderr
```

## Steps

1. Pass an unknown flag without `--root`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--unknown"}
	return nil
}
```
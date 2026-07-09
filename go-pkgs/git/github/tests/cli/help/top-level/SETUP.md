# Scenario

**Feature**: top-level `--help` prints github command usage

```
# no subcommand
RunCLI --help -> top-level usage -> stdout
```

## Steps

1. Set `req.Args` to `["--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```

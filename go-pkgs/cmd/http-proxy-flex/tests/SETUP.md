## Preconditions

- The `http-proxy` Go module exists at `../` relative to the test root
- `go build` can compile the binary

## Steps

1. Build the `http-proxy` binary to a temp location
2. Execute the binary with the arguments provided in `req.Args`
3. Capture combined stdout+stderr output and exit code

## Context

- The binary is a forward HTTP proxy that listens on a local port
- It supports `--listen-port`, `--upstream-proxy`, `--fallback-direct`, and `--help`

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Request struct {
	Args []string
}

type Response struct {
	Output   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	srcDir := filepath.Join(DOCTEST_ROOT, "..")
	binPath := filepath.Join(os.TempDir(), "http-proxy-test")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = srcDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed:\n%s", string(buildOut))
	}

	cmd := exec.Command(binPath, req.Args...)
	output, _ := cmd.CombinedOutput()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return &Response{
		Output:   string(output),
		ExitCode: exitCode,
	}, nil
}
```

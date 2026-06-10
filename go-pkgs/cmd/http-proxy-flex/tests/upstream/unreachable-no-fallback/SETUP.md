## Steps

- Run `http-proxy` with `--upstream-proxy http://127.0.0.1:19999` (nothing listening there) and NO `--fallback-direct`
- The process should exit with an error

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Run(t *testing.T, req *Request) (*Response, error) {
	srcDir := filepath.Join(DOCTEST_ROOT, "..")
	binPath := filepath.Join(os.TempDir(), "http-proxy-test")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = srcDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed:\n%s", string(buildOut))
	}

	cmd := exec.Command(binPath,
		"--upstream-proxy", "http://127.0.0.1:19999",
	)
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

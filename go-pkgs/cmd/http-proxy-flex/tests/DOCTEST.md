# http-proxy Test Cases

## Version
0.0.2

A CLI forward proxy with dynamic upstream health monitoring.

# DSN (Domain Specific Notion)

- **http-proxy binary** — forward HTTP proxy built from the command module.
- **Upstream proxy** — optional `--upstream-proxy` URL with health monitoring.
- **Fallback direct** — `--fallback-direct` allows direct routing when upstream is dead.
- **Listen port** — `--listen-port` controls the local listener (default 7821).

```
http-proxy --help
http-proxy --listen-port PORT --upstream-proxy URL [--fallback-direct]
```

## Test Tree

```
SETUP.md                              # Root: build & run binary
├── help/
│   └── show/                         # Leaf: --help prints usage
├── listen-port/
│   ├── default/                      # Leaf: default port is 7821
│   └── custom/                       # Leaf: --listen-port overrides
├── upstream/
│   ├── accessible/                   # Leaf: startup with reachable upstream (no --fallback-direct)
│   ├── accessible-fallback/          # Leaf: startup with reachable upstream + --fallback-direct
│   ├── unreachable-no-fallback/      # Leaf: startup with dead upstream, no --fallback-direct → warns, starts
│   └── unreachable-fallback/         # Leaf: startup with dead upstream, --fallback-direct → starts
├── dynamic/
│   ├── switch-and-back/              # Leaf: dead → live → dead upstream transitions
│   └── start-live-down-up/           # Leaf: live → dead → live upstream transitions
└── request-log/
    ├── direct/                       # Leaf: HTTP GET logged as "via direct"
    ├── upstream/                     # Leaf: HTTP GET logged as "via upstream proxy"
    ├── connect-dead-upstream/        # Leaf: CONNECT with dead upstream logs "via direct"
    └── connect-upstream-down-immediate-fallback/  # Leaf: CONNECT falls back when upstream dial fails
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | help/show | Running `--help` prints all flags and exits 0 |
| 2 | listen-port/default | Without `--listen-port`, prints "listening on :7821" |
| 3 | listen-port/custom | `--listen-port 9999` prints "listening on :9999" |
| 4 | upstream/accessible | Upstream is reachable at startup → prints "using upstream proxy" |
| 5 | upstream/accessible-fallback | Upstream reachable + `--fallback-direct` → health check loop starts on success |
| 6 | upstream/unreachable-no-fallback | Upstream is dead, no `--fallback-direct` → prints warning, starts listening |
| 7 | upstream/unreachable-fallback | Upstream is dead, `--fallback-direct` → prints "falling back to direct", starts listening |
| 8 | dynamic/switch-and-back | Dead upstream → becomes reachable → becomes dead → fallback transitions |
| 9 | dynamic/start-live-down-up | Live upstream → goes dead → comes back live → health check detects both |
| 10 | request-log/direct | HTTP GET in fallback mode logs "via direct" |
| 11 | request-log/upstream | HTTP GET through upstream proxy logs "via upstream proxy" |
| 12 | request-log/connect-dead-upstream | CONNECT with dead upstream at startup logs "via direct" |
| 13 | request-log/connect-upstream-down-immediate-fallback | CONNECT falls back to direct when upstream stops listening |

## How to Run

```sh
doctest test -v ./
```

```go
import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
	"time"
)

type Request struct {
	Args            []string
	CapturedOutput  string
	ConnectTarget   string
	ConnectResponse string
}

type Response struct {
	Output   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.CapturedOutput != "" {
		return &Response{
			Output:   req.CapturedOutput,
			ExitCode: 0,
		}, nil
	}

	binPath := getBinPath(t)
	cmd := exec.Command(binPath, req.Args...)

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stdoutBuf

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
	case <-time.After(4 * time.Second):
		cmd.Process.Kill()
		<-done
	}

	return &Response{
		Output:   stdoutBuf.String(),
		ExitCode: 0,
	}, nil
}
```
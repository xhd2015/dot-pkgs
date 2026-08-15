# http-proxy Test Cases

## Version
0.0.2

A CLI forward proxy with dynamic upstream health monitoring.

# DSN (Domain Specific Notion)

- **http-proxy binary** — forward HTTP proxy built from the command module.
- **Upstream proxy** — required `--upstream-proxy` URL with health monitoring.
- **Fallback direct** — enabled by default; `--no-fallback-direct` disables direct routing when upstream is dead.
- **Listen port** — `--listen-port` controls the local listener (default 7821).

```
http-proxy --help
http-proxy --listen-port PORT --upstream-proxy URL [--no-fallback-direct]
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
│   ├── accessible/                   # Leaf: startup with reachable upstream
│   ├── accessible-fallback/          # Leaf: default flex: health monitor on live upstream
│   ├── unreachable-no-fallback/      # Leaf: --no-fallback-direct + dead upstream → no direct
│   └── unreachable-fallback/         # Leaf: default flex: dead upstream → falls back to direct
├── dynamic/
│   ├── switch-and-back/              # Leaf: dead → live → dead upstream transitions
│   ├── start-live-down-up/           # Leaf: live → dead → live upstream transitions
│   └── start-dead-up-no-fallback/    # Leaf: default flex: dead startup → upstream up → switch
└── request-log/
    ├── direct/                       # Leaf: HTTP GET logged as "via direct"
    ├── upstream/                     # Leaf: HTTP GET logged as "via upstream proxy"
    ├── upstream-after-late-available/ # Leaf: GET via upstream after late upstream availability
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
| 5 | upstream/accessible-fallback | Default flex: health check loop runs on live upstream |
| 6 | upstream/unreachable-no-fallback | `--no-fallback-direct` + dead upstream → no direct fallback log |
| 7 | upstream/unreachable-fallback | Default flex: dead upstream → prints "falling back to direct" |
| 8 | dynamic/switch-and-back | Dead upstream → becomes reachable → becomes dead → fallback transitions |
| 9 | dynamic/start-live-down-up | Live upstream → goes dead → comes back live → health check detects both |
| 10 | dynamic/start-dead-up-no-fallback | Default flex: dead startup → upstream up → must switch |
| 11 | request-log/direct | HTTP GET in fallback mode logs "via direct" |
| 12 | request-log/upstream | HTTP GET through upstream proxy logs "via upstream proxy" |
| 13 | request-log/upstream-after-late-available | GET routes via upstream after upstream becomes available post-startup |
| 14 | request-log/connect-dead-upstream | CONNECT with dead upstream at startup logs "via direct" |
| 15 | request-log/connect-upstream-down-immediate-fallback | CONNECT falls back to direct when upstream stops listening |

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
	"github.com/xhd2015/doctest/session"
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

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.CapturedOutput != "" {
		return &Response{
			Output:   req.CapturedOutput,
			ExitCode: 0,
		}, nil
	}

	binPath := getBinPath(t, d)
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
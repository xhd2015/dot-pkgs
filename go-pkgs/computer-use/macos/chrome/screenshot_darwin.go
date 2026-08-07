//go:build darwin

package chrome

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// snapStep captures the interactive display after a named step when ScreenshotDir is set.
// Files: <ScreenshotDir>/NN-<step>.png  (also logs path + front tab URL on stdout).
func snapStep(opts LoadUnpackedOpts, res *LoadUnpackedResult, step string, n *int) {
	if strings.TrimSpace(opts.ScreenshotDir) == "" {
		return
	}
	dir := opts.ScreenshotDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		stepf(opts.Stderr, "warning: screenshot mkdir: %v", err)
		return
	}
	*n++
	// sanitize step for filename
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		if r == ' ' || r == '/' {
			return '-'
		}
		return -1
	}, step)
	if safe == "" {
		safe = "step"
	}
	path := filepath.Join(dir, fmt.Sprintf("%02d-%s.png", *n, safe))
	// -x: no sound; capture main display
	cmd := exec.Command("screencapture", "-x", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		stepf(opts.Stderr, "warning: screenshot %s failed: %v %s", step, err, strings.TrimSpace(string(out)))
		return
	}
	url := frontTabURL(nilCtx(), opts.AppName)
	stepf(opts.Stdout, "  shot      %s  %s  (front=%s)", step, path, url)
	if res != nil {
		res.Screenshots = append(res.Screenshots, path)
	}
	// tiny settle so files flush before next UI action
	time.Sleep(80 * time.Millisecond)
}

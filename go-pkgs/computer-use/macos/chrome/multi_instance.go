package chrome

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

// ChromeMainProcesses lists running Google Chrome main-process command lines
// (best-effort; macOS via ps). Used to detect temp user-data-dir instances.
func ChromeMainProcesses() []string {
	if runtime.GOOS != "darwin" {
		return nil
	}
	out, err := exec.Command("ps", "-ww", "-eo", "args=").Output()
	if err != nil {
		return nil
	}
	var lines []string
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		s := strings.TrimSpace(string(line))
		if s == "" {
			continue
		}
		// Main binary: .../MacOS/Google Chrome [args] — exclude Helpers.
		if !strings.Contains(s, "/MacOS/Google Chrome") {
			continue
		}
		if strings.Contains(s, "Helper") {
			continue
		}
		lines = append(lines, s)
	}
	return lines
}

// MultiInstanceHint returns a non-empty warning when multiple Chrome mains or a
// temp user-data-dir profile is present.
func MultiInstanceHint() string {
	procs := ChromeMainProcesses()
	if len(procs) == 0 {
		return ""
	}
	temp := 0
	for _, p := range procs {
		if strings.Contains(p, "user-data-dir=") &&
			(strings.Contains(p, "/tmp/") || strings.Contains(p, "chrome-ext-test")) {
			temp++
		}
	}
	if len(procs) > 1 || temp > 0 {
		return "multiple Google Chrome processes detected; AppleScript may target a temp profile — quit temp Chromes (e.g. --user-data-dir=/tmp/…) before Load unpacked"
	}
	return ""
}

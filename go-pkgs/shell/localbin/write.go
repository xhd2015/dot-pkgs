package localbin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/os/exectry"
)

// WriteScript writes body as an executable file at binDir/name.
// Creates binDir if needed. Always overwrites. Ensures a trailing newline.
func WriteScript(binDir, name, body string) (path string, err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("localbin: empty script name")
	}
	if strings.Contains(name, string(os.PathSeparator)) || name == "." || name == ".." {
		return "", fmt.Errorf("localbin: invalid script name %q", name)
	}
	binDir = strings.TrimSpace(binDir)
	if binDir == "" {
		return "", fmt.Errorf("localbin: empty bin dir")
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", binDir, err)
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	path = filepath.Join(binDir, name)
	if err := exectry.WriteExecutable(path, []byte(body)); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return path, nil
}

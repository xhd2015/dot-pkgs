package withgo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// ModuleGoLine reads go.mod in modDir and returns the go directive as
// major.minor with a "go" prefix (go 1.19.13 → go1.19). Missing file or
// missing go line is an error.
func ModuleGoLine(modDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		return "", err
	}
	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", err
	}
	if f.Go == nil || f.Go.Version == "" {
		return "", fmt.Errorf("go.mod: missing go directive")
	}
	ver := f.Go.Version
	parts := strings.Split(ver, ".")
	if len(parts) >= 2 {
		ver = parts[0] + "." + parts[1]
	}
	if strings.HasPrefix(ver, "go") {
		return ver, nil
	}
	return "go" + ver, nil
}

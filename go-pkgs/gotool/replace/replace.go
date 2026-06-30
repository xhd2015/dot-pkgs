package replace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
)

// Replace adds go mod edit -replace "$mod=$absDir" in the current working directory.
// dir must be an existing Go module directory.
func Replace(dir string) (absDir string, modulePath string, err error) {
	if dir == "" {
		return "", "", fmt.Errorf("requires dir")
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", "", fmt.Errorf("no such dir: %s", dir)
	}

	absDir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	modInfo, err := resolve.GetModuleInfo(dir)
	if err != nil {
		return "", "", err
	}

	if modInfo.Module.Path == "" {
		return "", "", fmt.Errorf("not a go module: %s", dir)
	}

	err = commands.GoModEditReplace(modInfo.Module.Path, absDir, nil)
	if err != nil {
		return "", "", err
	}

	return absDir, modInfo.Module.Path, nil
}

// ReplaceIn adds a replace directive in consumerDir's go.mod.
func ReplaceIn(consumerDir, dir string) (absDir string, modulePath string, err error) {
	if consumerDir == "" {
		return "", "", fmt.Errorf("requires consumer dir")
	}
	if dir == "" {
		return "", "", fmt.Errorf("requires dir")
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", "", fmt.Errorf("no such dir: %s", dir)
	}

	absDir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	modInfo, err := resolve.GetModuleInfo(dir)
	if err != nil {
		return "", "", err
	}

	if modInfo.Module.Path == "" {
		return "", "", fmt.Errorf("not a go module: %s", dir)
	}

	opts := &commands.GoModEditOptions{Dir: consumerDir, Stderr: false, Stdout: false}
	if err := commands.GoModEditReplace(modInfo.Module.Path, absDir, opts); err != nil {
		return "", "", err
	}

	return absDir, modInfo.Module.Path, nil
}
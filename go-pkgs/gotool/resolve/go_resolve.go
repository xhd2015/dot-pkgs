package resolve

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
)

// GoResolveVersion resolves version to a normalized version representation.
func GoResolveVersion(dir string, modPath string, version string) (string, error) {
	if modPath == "" {
		return "", fmt.Errorf("requires mod path")
	}
	if dir == "" {
		return "", fmt.Errorf("not supported")
	}
	pkg, err := findFirstPkg(dir, modPath, true)
	if err != nil {
		return "", err
	}
	if pkg == "" {
		return "", fmt.Errorf("no packages found in %s", dir)
	}

	tmpDir, err := os.MkdirTemp("", "go-resolve")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	goMod := filepath.Join(tmpDir, "go.mod")
	os.WriteFile(goMod, []byte(fmt.Sprintf(`module resolve
go 1.18
require %s %s`, modPath, version)), 0644)

	mainFile := filepath.Join(tmpDir, "main.go")
	os.WriteFile(mainFile, []byte(fmt.Sprintf(`package main
import 	_ %q
`, pkg)), 0644)

	opts := &commands.GoModEditOptions{Dir: tmpDir}
	err = commands.GoModTidy(opts)
	if err != nil {
		return "", err
	}
	modInfo, err := GetModuleInfo(tmpDir)
	if err != nil {
		return "", err
	}
	for _, require := range modInfo.Require {
		if require.Path == modPath {
			return require.Version, nil
		}
	}
	return "", fmt.Errorf("module %s not found", modPath)
}

// ResolveModPathFromPossibleDir resolves dir and module path from a directory or module path.
func ResolveModPathFromPossibleDir(dirOrModPath string) (dir string, modPath string, err error) {
	if dirOrModPath == "" {
		return "", "", fmt.Errorf("requires dir")
	}
	statDir, statErr := os.Stat(dirOrModPath)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return "", "", statErr
		}
	}
	if statDir == nil {
		return "", dirOrModPath, nil
	}
	if !statDir.IsDir() {
		return "", "", fmt.Errorf("not a go module: %s", dirOrModPath)
	}
	modInfo, err := GetModuleInfo(dirOrModPath)
	if err != nil {
		return "", "", err
	}
	if modInfo.Module.Path == "" {
		return "", "", fmt.Errorf("not a go module: %s", dirOrModPath)
	}
	return dirOrModPath, modInfo.Module.Path, nil
}

func findFirstPkg(dir string, modPath string, root bool) (pkg string, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	if !root {
		for _, file := range files {
			if file.Name() == "go.mod" {
				return "", nil
			}
		}
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".go") {
			return modPath, nil
		}
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		name := file.Name()
		subPath := filepath.Join(dir, name)
		pkg, err := findFirstPkg(subPath, modPath+"/"+name, false)
		if err != nil {
			return "", fmt.Errorf("%s: %w", subPath, err)
		}
		if pkg != "" {
			return pkg, nil
		}
	}
	return "", nil
}
package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
)

var semverPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+`)

// Update drops replace for the target module and sets require to the latest git tag.
func Update(dir string) error {
	if dir == "" {
		return fmt.Errorf("requires dir")
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("no such dir: %s", dir)
	}

	opts := &commands.GoModEditOptions{Dir: dir, Stderr: true}
	mod, err := commands.GoModEditJSON(opts)
	if err != nil {
		return fmt.Errorf("failed to get module info: %w", err)
	}
	if mod.Module.Path == "" {
		return fmt.Errorf("not a go module: %s", dir)
	}

	versionPrefix, err := CalculateVersionPrefix(dir, mod.Module.Path)
	if err != nil {
		return fmt.Errorf("failed to calculate version prefix for %s: %w", mod.Module.Path, err)
	}
	latestTag, err := GetLatestVersionTag(dir, versionPrefix)
	if err != nil {
		return fmt.Errorf("failed to get latest version tag for %s: %w", mod.Module.Path, err)
	}
	version := StripVersionPrefix(versionPrefix, latestTag)
	if !isValidVersionTag(version) {
		return fmt.Errorf("latest version tag %s resolved to invalid version %s", latestTag, version)
	}

	if err := commands.GoModDropReplace(mod.Module.Path, nil); err != nil {
		return err
	}
	if err := commands.GoModEditRequire(mod.Module.Path, version, nil); err != nil {
		return err
	}

	msgCmd := exec.Command("git", "log", "-1", "--format=%s", latestTag)
	msgCmd.Dir = dir
	msgCmd.Stderr = os.Stderr
	msgOutput, err := msgCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit message: %v", err)
	}
	msg := strings.TrimSpace(string(msgOutput))
	fmt.Printf("commit message: %s\n", msg)
	return nil
}

func isValidVersionTag(version string) bool {
	return version != "" && semverPattern.MatchString(version)
}

// CalculateVersionPrefix returns the git tag prefix for a module directory.
func CalculateVersionPrefix(targetDir, modulePath string) (string, error) {
	gitRoot, subPathList, err := getSubPath(targetDir)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(filepath.Join(gitRoot, "go.mod")); os.IsNotExist(err) {
		return addVersionPrefix(strings.Join(subPathList, "/")), nil
	}

	rootModPath, err := resolve.GetRootModulePath(targetDir)
	if err != nil {
		return "", err
	}

	subModulePath, ok := cutSubmoduleSuffix(rootModPath, modulePath)
	if !ok {
		return "", fmt.Errorf("module path %s is not a submodule of root module path %s", modulePath, rootModPath)
	}

	return addVersionPrefix(subModulePath), nil
}

func cutSubmoduleSuffix(parentModulePath, childModulePath string) (string, bool) {
	if !strings.HasPrefix(childModulePath, parentModulePath) {
		return "", false
	}
	if len(childModulePath) == len(parentModulePath) {
		return "", true
	}
	if childModulePath[len(parentModulePath)] != '/' {
		return "", false
	}
	return childModulePath[len(parentModulePath)+1:], true
}
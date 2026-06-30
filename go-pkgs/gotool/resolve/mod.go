package resolve

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ModuleInfo is the JSON shape returned by "go mod edit -json".
type ModuleInfo struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
	Replace []struct {
		Old struct {
			Path string `json:"Path"`
		} `json:"Old"`
		New struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"New"`
	} `json:"Replace"`
}

// GetModuleInfo runs "go mod edit -json" in dir.
func GetModuleInfo(dir string) (*ModuleInfo, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("resolve go mod: %s %w", dir, err)
	}

	var modInfo ModuleInfo
	if err := json.Unmarshal(output, &modInfo); err != nil {
		return nil, fmt.Errorf("failed to parse module info: %v", err)
	}

	return &modInfo, nil
}

// GetRootModulePath returns the module path from go.mod at the git repository root.
func GetRootModulePath(targetDir string) (string, error) {
	gitRoot, err := showTopLevel(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to get git root for %s: %w", targetDir, err)
	}

	rootModInfo, err := GetModuleInfo(gitRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get root module info for %s: %w", gitRoot, err)
	}

	return rootModInfo.Module.Path, nil
}

// LocalModuleInfo represents information about a resolved local module.
type LocalModuleInfo struct {
	LocalPath      string
	ModuleInfo     *ModuleInfo
	IsDependency   bool
	IsReplaced     bool
	CurrentVersion string
}

// ResolveLocalModules resolves local module directories against currentDir's go.mod.
func ResolveLocalModules(currentDir string, localModDirs []string) (*ModuleInfo, []*LocalModuleInfo, error) {
	currentModInfo, err := GetModuleInfo(currentDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current module info: %w", err)
	}

	var resolvedModules []*LocalModuleInfo
	for _, localModDir := range localModDirs {
		resolved, err := resolveLocalModule(localModDir, currentModInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve local module %s: %w", localModDir, err)
		}
		if resolved != nil {
			resolvedModules = append(resolvedModules, resolved)
		}
	}

	return currentModInfo, resolvedModules, nil
}

// IsDependency reports whether localModDir's module path appears in consumerDir's go.mod.
func IsDependency(consumerDir, localModDir string) (bool, string, error) {
	_, resolved, err := ResolveLocalModules(consumerDir, []string{localModDir})
	if err != nil {
		return false, "", err
	}
	if len(resolved) == 0 {
		return false, "", fmt.Errorf("failed to resolve local module %s", localModDir)
	}
	return resolved[0].IsDependency, resolved[0].ModuleInfo.Module.Path, nil
}

// HasLocalFilesystemReplace reports whether modInfo contains a filesystem replace directive.
func HasLocalFilesystemReplace(modInfo *ModuleInfo) bool {
	if modInfo == nil {
		return false
	}
	for _, repl := range modInfo.Replace {
		p := repl.New.Path
		if p == "" || repl.New.Version != "" {
			continue
		}
		if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") || filepath.IsAbs(p) {
			return true
		}
	}
	return false
}

func resolveLocalModule(localModDir string, currentModInfo *ModuleInfo) (*LocalModuleInfo, error) {
	absPath, err := filepath.Abs(localModDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", localModDir, err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("local module directory does not exist: %s: %w", absPath, err)
	}

	localModInfo, err := GetModuleInfo(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get module info for %s: %w", absPath, err)
	}

	if localModInfo.Module.Path == "" {
		return nil, fmt.Errorf("not a go module: %s", absPath)
	}

	modulePath := localModInfo.Module.Path

	var isDependency bool
	var isReplaced bool
	var currentVersion string

	for _, req := range currentModInfo.Require {
		if req.Path == modulePath {
			isDependency = true
			currentVersion = req.Version
			break
		}
	}

	for _, repl := range currentModInfo.Replace {
		if repl.Old.Path == modulePath {
			isReplaced = true
			if !isDependency {
				isDependency = true
			}
			break
		}
	}

	return &LocalModuleInfo{
		LocalPath:      absPath,
		ModuleInfo:     localModInfo,
		IsDependency:   isDependency,
		IsReplaced:     isReplaced,
		CurrentVersion: currentVersion,
	}, nil
}

func showTopLevel(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
package npm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type lockfileMarker struct {
	name    string
	manager Manager
}

var lockfileMarkers = []lockfileMarker{
	{name: "pnpm-lock.yaml", manager: ManagerPnpm},
	{name: "bun.lockb", manager: ManagerBun},
	{name: "bun.lock", manager: ManagerBun},
	{name: "package-lock.json", manager: ManagerNpm},
	{name: "yarn.lock", manager: ManagerYarn},
}

// DetectFromDir inspects projectDir for package-manager indicators.
func DetectFromDir(projectDir string) Trace {
	return detect(filepath.Clean(projectDir), "")
}

// DetectFromNodeModules inspects the parent of nodeModulesAbsPath.
func DetectFromNodeModules(nodeModulesAbsPath string) Trace {
	clean := filepath.Clean(nodeModulesAbsPath)
	return detect(filepath.Dir(clean), clean)
}

// ManagerFromNodeModules returns the detected manager name for node_modules paths.
func ManagerFromNodeModules(nodeModulesAbsPath string) string {
	return string(DetectFromNodeModules(nodeModulesAbsPath).Manager)
}

// HasPackageJSON reports whether package.json exists beside node_modules.
func HasPackageJSON(nodeModulesAbsPath string) bool {
	projectRoot := filepath.Dir(filepath.Clean(nodeModulesAbsPath))
	return fileExists(filepath.Join(projectRoot, "package.json"))
}

func detect(projectRoot, nodeModulesAbsPath string) Trace {
	trace := Trace{
		ProjectRoot:        projectRoot,
		NodeModulesAbsPath: nodeModulesAbsPath,
		Steps:              make([]string, 0, 16),
	}

	if nodeModulesAbsPath != "" {
		traceStep(&trace.Steps, "node_modules = %s", nodeModulesAbsPath)
	}
	traceStep(&trace.Steps, "projectRoot = %s", projectRoot)

	pkgPath := filepath.Join(projectRoot, "package.json")
	trace.HasPackageJSON = fileExists(pkgPath)

	for _, marker := range lockfileMarkers {
		path := filepath.Join(projectRoot, marker.name)
		if fileExists(path) {
			trace.Signals = append(trace.Signals, Signal{
				Manager: marker.manager,
				Source:  marker.name,
			})
			traceStep(&trace.Steps, "found %s", marker.name)
		}
	}

	pnpmStore := filepath.Join(projectRoot, "node_modules", ".pnpm")
	if nodeModulesAbsPath != "" {
		pnpmStore = filepath.Join(nodeModulesAbsPath, ".pnpm")
	}
	if dirExists(pnpmStore) {
		trace.Signals = append(trace.Signals, Signal{
			Manager: ManagerPnpm,
			Source:  "node_modules/.pnpm",
		})
		traceStep(&trace.Steps, "found node_modules/.pnpm at %s", pnpmStore)
	}

	if trace.HasPackageJSON {
		if field := parsePackageManagerField(pkgPath); field != "" {
			trace.Signals = append(trace.Signals, Signal{
				Manager: field,
				Source:  "package.json packageManager",
			})
			traceStep(&trace.Steps, "package.json packageManager -> %s", field)
		}
	}

	trace.Manager = pickManager(trace.Signals, trace.HasPackageJSON)
	if len(trace.Signals) == 0 {
		if trace.HasPackageJSON {
			traceStep(&trace.Steps, "no indicators; default npm")
		} else {
			traceStep(&trace.Steps, "no indicators; unknown")
		}
	} else {
		traceStep(&trace.Steps, "resolved manager = %s", trace.Manager)
	}
	return trace
}

func pickManager(signals []Signal, hasPackageJSON bool) Manager {
	if len(signals) == 0 {
		if hasPackageJSON {
			return ManagerNpm
		}
		return ManagerUnknown
	}

	seen := make(map[Manager]bool, len(signals))
	for _, signal := range signals {
		seen[signal.Manager] = true
	}

	for _, manager := range detectionPriority {
		if seen[manager] {
			return manager
		}
	}
	return ManagerUnknown
}

func parsePackageManagerField(pkgPath string) Manager {
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return ""
	}
	var pkg struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}
	tool, _, ok := strings.Cut(pkg.PackageManager, "@")
	if !ok || tool == "" {
		return ""
	}
	manager := Manager(tool)
	if knownManager(manager) {
		return manager
	}
	return ""
}

func traceStep(steps *[]string, format string, args ...interface{}) {
	*steps = append(*steps, fmt.Sprintf(format, args...))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
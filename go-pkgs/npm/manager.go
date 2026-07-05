package npm

import "os/exec"

// Manager identifies a JavaScript package manager.
type Manager string

const (
	ManagerPnpm    Manager = "pnpm"
	ManagerBun     Manager = "bun"
	ManagerNpm     Manager = "npm"
	ManagerYarn    Manager = "yarn"
	ManagerUnknown Manager = "unknown"
)

// detectionPriority orders managers when multiple indicators are present.
var detectionPriority = []Manager{
	ManagerPnpm,
	ManagerBun,
	ManagerNpm,
	ManagerYarn,
}

// Signal records one package-manager indicator found in a project.
type Signal struct {
	Manager Manager
	Source  string
}

// Trace records how a manager was resolved.
type Trace struct {
	ProjectRoot        string
	NodeModulesAbsPath string
	Manager            Manager
	HasPackageJSON     bool
	Signals            []Signal
	Steps              []string
}

// Available reports whether manager's CLI is present in PATH.
func Available(manager Manager) bool {
	if manager == ManagerUnknown || manager == "" {
		return false
	}
	_, err := exec.LookPath(string(manager))
	return err == nil
}

func knownManager(manager Manager) bool {
	switch manager {
	case ManagerPnpm, ManagerBun, ManagerNpm, ManagerYarn:
		return true
	default:
		return false
	}
}
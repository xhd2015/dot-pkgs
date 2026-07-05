package npm

import (
	"errors"
	"fmt"
)

// Resolve picks a package manager for projectDir.
// pref may be "auto", "", or an explicit manager name (pnpm, bun, npm, yarn).
func Resolve(projectDir string, pref string) (Manager, error) {
	switch pref {
	case "", "auto":
		return resolveAuto(projectDir)
	default:
		manager := Manager(pref)
		if !knownManager(manager) {
			return "", fmt.Errorf("unknown package manager %q: expected auto, pnpm, bun, npm, or yarn", pref)
		}
		if !Available(manager) {
			return "", fmt.Errorf("cannot find %s in PATH", pref)
		}
		return manager, nil
	}
}

func resolveAuto(projectDir string) (Manager, error) {
	trace := DetectFromDir(projectDir)
	candidates := candidateManagers(trace)

	for _, manager := range candidates {
		if Available(manager) {
			return manager, nil
		}
	}

	for _, manager := range detectionPriority {
		if Available(manager) {
			return manager, nil
		}
	}

	return "", errors.New("cannot find pnpm, bun, or npm in PATH")
}

func candidateManagers(trace Trace) []Manager {
	seen := make(map[Manager]bool)
	var candidates []Manager

	add := func(manager Manager) {
		if manager == ManagerUnknown || manager == "" || seen[manager] {
			return
		}
		seen[manager] = true
		candidates = append(candidates, manager)
	}

	add(trace.Manager)
	for _, signal := range trace.Signals {
		add(signal.Manager)
	}
	for _, manager := range detectionPriority {
		add(manager)
	}
	return candidates
}
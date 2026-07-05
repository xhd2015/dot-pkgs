package npm

// InstallOptions configures dependency installation.
type InstallOptions struct {
	FrozenLockfile bool
}

// InstallArgs returns CLI arguments for installing dependencies.
func InstallArgs(manager Manager, opts InstallOptions) []string {
	switch manager {
	case ManagerPnpm:
		if opts.FrozenLockfile {
			return []string{"install", "--frozen-lockfile"}
		}
		return []string{"install"}
	case ManagerBun:
		if opts.FrozenLockfile {
			return []string{"install", "--frozen-lockfile"}
		}
		return []string{"install"}
	case ManagerNpm:
		if opts.FrozenLockfile {
			return []string{"ci"}
		}
		return []string{"install", "--no-package-lock"}
	case ManagerYarn:
		if opts.FrozenLockfile {
			return []string{"install", "--frozen-lockfile"}
		}
		return []string{"install"}
	default:
		panic("unsupported package manager: " + manager)
	}
}

// InstallCommand returns the executable name and install arguments.
func InstallCommand(manager Manager, opts InstallOptions) (string, []string) {
	return string(manager), InstallArgs(manager, opts)
}
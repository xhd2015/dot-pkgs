package singboxtun

import (
	"fmt"

	"github.com/xhd2015/dot-pkgs/go-pkgs/sudosetup"
)

func defaultEnsureSudoSetup(singBoxPath string, noSetup bool, cacheDirName, sudoersName string) error {
	if noSetup {
		return nil
	}
	mgr := &sudosetup.Manager{
		Config: sudosetup.Config{
			CacheDirName: resolveCacheDirName(cacheDirName),
			SudoersName:  resolveSudoersName(sudoersName),
		},
		Rule: sudosetup.Rule{
			Command:     singBoxPath,
			ArgsPattern: "run -c *",
		},
	}
	if installed, _ := mgr.IsInstalled(); installed {
		return nil
	}
	if err := mgr.EnsureInstalled(); err != nil {
		return fmt.Errorf("sudo NOPASSWD setup: %w", err)
	}
	return nil
}

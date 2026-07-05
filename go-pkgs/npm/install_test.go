package npm

import (
	"reflect"
	"testing"
)

func TestInstallArgs(t *testing.T) {
	tests := []struct {
		manager Manager
		opts    InstallOptions
		want    []string
	}{
		{ManagerPnpm, InstallOptions{}, []string{"install"}},
		{ManagerPnpm, InstallOptions{FrozenLockfile: true}, []string{"install", "--frozen-lockfile"}},
		{ManagerBun, InstallOptions{}, []string{"install"}},
		{ManagerBun, InstallOptions{FrozenLockfile: true}, []string{"install", "--frozen-lockfile"}},
		{ManagerNpm, InstallOptions{}, []string{"install", "--no-package-lock"}},
		{ManagerNpm, InstallOptions{FrozenLockfile: true}, []string{"ci"}},
		{ManagerYarn, InstallOptions{FrozenLockfile: true}, []string{"install", "--frozen-lockfile"}},
	}

	for _, tt := range tests {
		got := InstallArgs(tt.manager, tt.opts)
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("InstallArgs(%q, %+v) = %v, want %v", tt.manager, tt.opts, got, tt.want)
		}
	}
}

func TestInstallCommand(t *testing.T) {
	name, args := InstallCommand(ManagerPnpm, InstallOptions{})
	if name != "pnpm" || len(args) != 1 || args[0] != "install" {
		t.Fatalf("InstallCommand() = (%q, %v)", name, args)
	}
}
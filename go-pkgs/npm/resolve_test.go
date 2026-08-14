package npm

import (
	"testing"
)

func requireAvailable(t *testing.T, manager Manager) {
	t.Helper()
	if !Available(manager) {
		t.Skipf("%s not available on PATH", manager)
	}
}

func TestResolveExplicitPnpm(t *testing.T) {
	requireAvailable(t, ManagerPnpm)
	dir := setupProject(t, nil)
	manager, err := Resolve(dir, "pnpm")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if manager != ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", manager)
	}
}

func TestResolveAutoFromPackageManagerField(t *testing.T) {
	requireAvailable(t, ManagerPnpm)
	dir := setupProject(t, map[string]string{
		"package.json": `{"name":"demo","packageManager":"pnpm@11.10.0"}`,
	})
	manager, err := Resolve(dir, "auto")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if manager != ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", manager)
	}
}

func TestResolveRejectsUnknownManager(t *testing.T) {
	dir := setupProject(t, nil)
	if _, err := Resolve(dir, "yarnberry"); err == nil {
		t.Fatal("expected error for unknown manager")
	}
}
package npm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFromDirPnpmLock(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json":    `{"name":"demo"}`,
		"pnpm-lock.yaml":  "lockfileVersion: '9.0'\n",
	})
	trace := DetectFromDir(dir)
	if trace.Manager != ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", trace.Manager)
	}
}

func TestDetectFromDirMixedLockfilesPreferPnpm(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json":       `{"name":"demo","packageManager":"npm@10.0.0"}`,
		"pnpm-lock.yaml":     "lockfileVersion: '9.0'\n",
		"bun.lock":           "{}",
		"package-lock.json":  `{"lockfileVersion":3}`,
	})
	trace := DetectFromDir(dir)
	if trace.Manager != ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", trace.Manager)
	}
	if len(trace.Signals) < 4 {
		t.Fatalf("expected multiple signals, got %d", len(trace.Signals))
	}
}

func TestDetectFromDirBunBeforeNpm(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json":      `{"name":"demo"}`,
		"bun.lock":          "{}",
		"package-lock.json": `{"lockfileVersion":3}`,
	})
	trace := DetectFromDir(dir)
	if trace.Manager != ManagerBun {
		t.Fatalf("manager = %q, want bun", trace.Manager)
	}
}

func TestDetectFromDirPackageManagerField(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json": `{"name":"demo","packageManager":"pnpm@11.10.0"}`,
	})
	trace := DetectFromDir(dir)
	if trace.Manager != ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", trace.Manager)
	}
}

func TestDetectFromDirDefaultNpm(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json": `{"name":"demo"}`,
	})
	trace := DetectFromDir(dir)
	if trace.Manager != ManagerNpm {
		t.Fatalf("manager = %q, want npm", trace.Manager)
	}
}

func TestDetectFromDirUnknown(t *testing.T) {
	dir := setupProject(t, nil)
	trace := DetectFromDir(dir)
	if trace.Manager != ManagerUnknown {
		t.Fatalf("manager = %q, want unknown", trace.Manager)
	}
}

func TestDetectFromNodeModulesPnpmStore(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json": `{"name":"demo"}`,
	})
	pnpmDir := filepath.Join(dir, "node_modules", ".pnpm", "pkg@1.0.0")
	if err := os.MkdirAll(pnpmDir, 0755); err != nil {
		t.Fatal(err)
	}

	trace := DetectFromNodeModules(filepath.Join(dir, "node_modules"))
	if trace.Manager != ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", trace.Manager)
	}
	if !trace.HasPackageJSON {
		t.Fatal("expected hasPackageJSON = true")
	}
}

func TestHasPackageJSON(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json": `{"name":"demo"}`,
	})
	if !HasPackageJSON(filepath.Join(dir, "node_modules")) {
		t.Fatal("expected package.json detected")
	}
}

func setupProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "npm-detect-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
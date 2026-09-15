package installplan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscover_CmdOnly(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/app")
	writeMain(t, filepath.Join(mod, "cmd", "tool"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Items) != 1 {
		t.Fatalf("items=%v", plan.Items)
	}
	it := plan.Items[0]
	if it.BinName != "tool" || it.RelPath != "./cmd/tool" || it.Method != MethodGoInstall {
		t.Fatalf("item=%+v", it)
	}
}

func TestDiscover_NamedScriptBeatsCmd(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/app")
	writeMain(t, filepath.Join(mod, "cmd", "tool"))
	writeMain(t, filepath.Join(mod, "script", "tool", "install"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Items) != 1 {
		t.Fatalf("items=%v", plan.Items)
	}
	it := plan.Items[0]
	if it.RelPath != "./script/tool/install" || it.Method != MethodGoRunInstall {
		t.Fatalf("item=%+v", it)
	}
	if !hasDiag(plan.Diagnostics, DiagKindPreferScript, "tool") {
		t.Fatalf("missing prefer-script: %v", plan.Diagnostics)
	}
}

func TestDiscover_ModuleName_NamedOverBareOverCmd(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/wrk")
	writeMain(t, filepath.Join(mod, "cmd", "wrk"))
	writeMain(t, filepath.Join(mod, "script", "install"))
	writeMain(t, filepath.Join(mod, "script", "wrk", "install"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	it := mustItem(t, plan, "wrk")
	if it.RelPath != "./script/wrk/install" {
		t.Fatalf("rel=%s", it.RelPath)
	}
	d := mustDiag(t, plan.Diagnostics, DiagKindPreferScript, "wrk")
	if d.Paths[0] != "./script/wrk/install" {
		t.Fatalf("winner=%v", d.Paths)
	}
	if !containsAll(d.Paths, "./script/install", "./cmd/wrk") {
		t.Fatalf("skipped=%v", d.Paths)
	}
}

func TestDiscover_ModuleName_BareOverCmd(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/wrk")
	writeMain(t, filepath.Join(mod, "cmd", "wrk"))
	writeMain(t, filepath.Join(mod, "script", "install"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	it := mustItem(t, plan, "wrk")
	if it.RelPath != "./script/install" {
		t.Fatalf("rel=%s", it.RelPath)
	}
	d := mustDiag(t, plan.Diagnostics, DiagKindPreferScript, "wrk")
	if d.Paths[0] != "./script/install" || !containsAll(d.Paths, "./cmd/wrk") {
		t.Fatalf("paths=%v", d.Paths)
	}
}

func TestDiscover_OtherCmdIgnoresBareScriptInstall(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/wrk")
	writeMain(t, filepath.Join(mod, "script", "install"))
	writeMain(t, filepath.Join(mod, "cmd", "debug"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	debug := mustItem(t, plan, "debug")
	if debug.RelPath != "./cmd/debug" || debug.Method != MethodGoInstall {
		t.Fatalf("debug=%+v", debug)
	}
	wrk := mustItem(t, plan, "wrk")
	if wrk.RelPath != "./script/install" {
		t.Fatalf("wrk=%+v", wrk)
	}
}

func TestDiscover_NestedExtraWarnsWhenNamedExists(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/app")
	writeMain(t, filepath.Join(mod, "script", "foo", "install"))
	writeMain(t, filepath.Join(mod, "script", "x", "foo", "install"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	it := mustItem(t, plan, "foo")
	if it.RelPath != "./script/foo/install" {
		t.Fatalf("rel=%s", it.RelPath)
	}
	d := mustDiag(t, plan.Diagnostics, DiagKindNestedScript, "foo")
	if !containsAll(d.Paths, "./script/foo/install", "./script/x/foo/install") {
		t.Fatalf("nested paths=%v", d.Paths)
	}
}

func TestDiscover_UniqueNestedFallback(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/app")
	writeMain(t, filepath.Join(mod, "script", "x", "foo", "install"))
	writeMain(t, filepath.Join(mod, "cmd", "foo"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	it := mustItem(t, plan, "foo")
	if it.RelPath != "./script/x/foo/install" {
		t.Fatalf("rel=%s", it.RelPath)
	}
	if !hasDiag(plan.Diagnostics, DiagKindPreferScript, "foo") {
		t.Fatalf("want prefer-script over cmd: %v", plan.Diagnostics)
	}
}

func TestDiscover_AmbiguousCmdFallsBackToNamedScript(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/app")
	writeMain(t, filepath.Join(mod, "cmd", "foo"))
	writeMain(t, filepath.Join(mod, "cmd", "nested", "foo"))
	writeMain(t, filepath.Join(mod, "script", "foo", "install"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	it := mustItem(t, plan, "foo")
	if it.RelPath != "./script/foo/install" {
		t.Fatalf("rel=%s", it.RelPath)
	}
	if !hasDiag(plan.Diagnostics, DiagKindAmbiguousCmd, "foo") {
		t.Fatalf("want ambiguous-cmd: %v", plan.Diagnostics)
	}
}

func TestDiscover_AmbiguousNestedOmitsScriptUsesCmd(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/app")
	writeMain(t, filepath.Join(mod, "script", "a", "foo", "install"))
	writeMain(t, filepath.Join(mod, "script", "b", "foo", "install"))
	writeMain(t, filepath.Join(mod, "cmd", "foo"))

	plan, err := Discover(mod)
	if err != nil {
		t.Fatal(err)
	}
	it := mustItem(t, plan, "foo")
	if it.RelPath != "./cmd/foo" {
		t.Fatalf("rel=%s", it.RelPath)
	}
	if !hasDiag(plan.Diagnostics, DiagKindAmbiguousScript, "foo") {
		t.Fatalf("want ambiguous-script: %v", plan.Diagnostics)
	}
}

func TestLookup_Missing(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/app")
	writeMain(t, filepath.Join(mod, "cmd", "only"))
	plan, err := DiscoverMulti([]string{mod})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lookup(plan, []string{"missing"})
	if err == nil || !strings.Contains(err.Error(), `no install candidate for "missing"`) {
		t.Fatalf("err=%v", err)
	}
}

func TestLookup_CrossModuleCollision(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	writeGoMod(t, a, "example.com/a")
	writeGoMod(t, b, "example.com/b")
	writeMain(t, filepath.Join(a, "cmd", "same"))
	writeMain(t, filepath.Join(b, "cmd", "same"))
	plan, err := DiscoverMulti([]string{a, b})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lookup(plan, []string{"same"})
	if err == nil || !strings.Contains(err.Error(), "multiple modules") {
		t.Fatalf("err=%v", err)
	}
}

func TestLookup_SelectsAndKeepsDiags(t *testing.T) {
	mod := t.TempDir()
	writeGoMod(t, mod, "example.com/app")
	writeMain(t, filepath.Join(mod, "cmd", "alpha"))
	writeMain(t, filepath.Join(mod, "cmd", "beta"))
	writeMain(t, filepath.Join(mod, "script", "alpha", "install"))
	full, err := DiscoverMulti([]string{mod})
	if err != nil {
		t.Fatal(err)
	}
	named, err := Lookup(full, []string{"beta", "alpha", "beta"})
	if err != nil {
		t.Fatal(err)
	}
	if len(named.Modules) != 1 || len(named.Modules[0].Items) != 2 {
		t.Fatalf("named=%+v", named)
	}
	if named.Modules[0].Items[0].BinName != "beta" || named.Modules[0].Items[1].BinName != "alpha" {
		t.Fatalf("order=%v", named.Modules[0].Items)
	}
	if !hasDiag(named.Modules[0].Diagnostics, DiagKindPreferScript, "alpha") {
		t.Fatalf("diags=%v", named.Modules[0].Diagnostics)
	}
}

func writeGoMod(t *testing.T, dir, modulePath string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "module " + modulePath + "\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeMain(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustItem(t *testing.T, plan *ModulePlan, name string) Item {
	t.Helper()
	for _, it := range plan.Items {
		if it.BinName == name {
			return it
		}
	}
	t.Fatalf("no item %q in %v", name, plan.Items)
	return Item{}
}

func hasDiag(diags []Diagnostic, kind, bin string) bool {
	for _, d := range diags {
		if d.Kind == kind && d.BinName == bin {
			return true
		}
	}
	return false
}

func mustDiag(t *testing.T, diags []Diagnostic, kind, bin string) Diagnostic {
	t.Helper()
	for _, d := range diags {
		if d.Kind == kind && d.BinName == bin {
			return d
		}
	}
	t.Fatalf("missing diag kind=%s bin=%s in %v", kind, bin, diags)
	return Diagnostic{}
}

func containsAll(got []string, want ...string) bool {
	set := map[string]struct{}{}
	for _, g := range got {
		set[g] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; !ok {
			return false
		}
	}
	return true
}

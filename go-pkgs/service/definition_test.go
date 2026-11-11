package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeDefinitionDefaultsNameToID(t *testing.T) {
	def, err := NormalizeDefinition(Definition{ID: "  llm-proxy  ", Command: "  llm-proxy serve  "})
	if err != nil {
		t.Fatalf("NormalizeDefinition() error = %v", err)
	}
	if def.ID != "llm-proxy" || def.Name != "llm-proxy" || def.Command != "llm-proxy serve" {
		t.Fatalf("unexpected normalization: %+v", def)
	}
	if def.WorkingDir == "" || !filepath.IsAbs(def.WorkingDir) {
		t.Fatalf("WorkingDir = %q, want absolute", def.WorkingDir)
	}
}

func TestNormalizeDefinitionKeepsExplicitWorkingDir(t *testing.T) {
	dir := t.TempDir()
	def, err := NormalizeDefinition(Definition{ID: "a", Command: "true", WorkingDir: dir})
	if err != nil {
		t.Fatalf("NormalizeDefinition() error = %v", err)
	}
	if def.WorkingDir != dir {
		t.Fatalf("WorkingDir = %q, want %q", def.WorkingDir, dir)
	}
}

func TestNormalizeDefinitionRejectsVariableExtraPath(t *testing.T) {
	_, err := NormalizeDefinition(Definition{ID: "a", Command: "true", ExtraPath: "$HOME/bin"})
	if err == nil {
		t.Fatal("NormalizeDefinition() error = nil, want variable rejection")
	}
	if !strings.Contains(err.Error(), "extra path") {
		t.Fatalf("error = %v, want extra path context", err)
	}
}

func TestValidateDefinition(t *testing.T) {
	cases := []struct {
		name string
		def  Definition
		want string
	}{
		{"missing id", Definition{Command: "true"}, "service id is required"},
		{"missing name", Definition{ID: "a"}, "service name is required"},
		{"missing command", Definition{ID: "a", Name: "A"}, "service command is required"},
		{"bad env key", Definition{ID: "a", Name: "A", Command: "true", ExtraEnv: map[string]string{"1BAD": "x"}}, "invalid environment variable name: 1BAD"},
		{"relative extra path", Definition{ID: "a", Name: "A", Command: "true", ExtraPath: "bin"}, "extra path"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDefinition(tc.def)
			if err == nil {
				t.Fatalf("ValidateDefinition() error = nil, want %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestDefinitionIsEnabledUnlessExplicitlyDisabled(t *testing.T) {
	if !(Definition{}).IsEnabled() {
		t.Fatal("absent Enabled must mean enabled")
	}
	if !(Definition{Enabled: Bool(true)}).IsEnabled() {
		t.Fatal("Enabled=true must be enabled")
	}
	if (Definition{Enabled: Bool(false)}).IsEnabled() {
		t.Fatal("Enabled=false must be disabled")
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"/usr/local/bin":          "/usr/local/bin",
		"/usr/local/bin:/usr/bin": "'/usr/local/bin:/usr/bin'",
		"":                        "",
		"/tmp/a b":                `'/tmp/a b'`,
		"it's":                    `'it'"'"'s'`,
	}
	for in, want := range cases {
		if got := ShellQuote(in); got != want {
			t.Fatalf("ShellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEnvValueAndSetEnvValue(t *testing.T) {
	env := []string{"PATH=/bin", "FOO=one\ntwo"}
	if got := EnvValue(env, "FOO"); got != "one\ntwo" {
		t.Fatalf("EnvValue(FOO) = %q", got)
	}
	if got := EnvValue(env, "MISSING"); got != "" {
		t.Fatalf("EnvValue(MISSING) = %q, want empty", got)
	}
	updated := SetEnvValue(env, "FOO", "three")
	if got := EnvValue(updated, "FOO"); got != "three" {
		t.Fatalf("EnvValue(FOO) after set = %q", got)
	}
	if len(updated) != len(env) {
		t.Fatalf("SetEnvValue replaced an entry instead of updating: %v", updated)
	}
	appended := SetEnvValue(env, "BAR", "x")
	if EnvValue(appended, "BAR") != "x" || len(appended) != len(env)+1 {
		t.Fatalf("SetEnvValue did not append: %v", appended)
	}
}

func TestNormalizeEnvTrimsAndDropsBlankKeys(t *testing.T) {
	got := NormalizeEnv(map[string]string{" B ": "2", "A": "1", "": "y", "   ": "z"})
	if len(got) != 2 || got["A"] != "1" || got["B"] != "2" {
		t.Fatalf("NormalizeEnv() = %v", got)
	}
	if NormalizeEnv(nil) != nil {
		t.Fatal("NormalizeEnv(nil) must stay nil")
	}
}

func TestIsValidEnvKey(t *testing.T) {
	for _, key := range []string{"A", "_x", "A1_b"} {
		if !IsValidEnvKey(key) {
			t.Fatalf("IsValidEnvKey(%q) = false", key)
		}
	}
	for _, key := range []string{"", "1A", "A-B", "A B", "A=B"} {
		if IsValidEnvKey(key) {
			t.Fatalf("IsValidEnvKey(%q) = true", key)
		}
	}
}

func TestStringMapEqual(t *testing.T) {
	if !StringMapEqual(nil, map[string]string{}) {
		t.Fatal("nil and empty maps must be equal")
	}
	if StringMapEqual(map[string]string{"a": "1"}, map[string]string{"a": "2"}) {
		t.Fatal("different values must not be equal")
	}
	if !StringMapEqual(map[string]string{"a": "1", "b": "2"}, map[string]string{"b": "2", "a": "1"}) {
		t.Fatal("map order must not matter")
	}
}

func TestLogPath(t *testing.T) {
	if got, want := LogPath("/tmp/logs", "llm-proxy"), filepath.Join("/tmp/logs", "llm-proxy.log"); got != want {
		t.Fatalf("LogPath() = %q, want %q", got, want)
	}
}

func TestSortDefinitionsAndIDs(t *testing.T) {
	defs := []Definition{{ID: "c"}, {ID: "a"}, {ID: "b"}}
	SortDefinitions(defs)
	if ids := DefinitionIDs(defs); strings.Join(ids, ",") != "a,b,c" {
		t.Fatalf("DefinitionIDs() = %v", ids)
	}
}

func TestEnsureWorkingDirCreatesNested(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b")
	if err := EnsureWorkingDir(dir); err != nil {
		t.Fatalf("EnsureWorkingDir() error = %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("working dir not created: %v", err)
	}
	if err := EnsureWorkingDir(""); err != nil {
		t.Fatalf("EnsureWorkingDir(\"\") error = %v", err)
	}
}

func TestCloneStringMapCopiesAndKeepsNil(t *testing.T) {
	if got := CloneStringMap(nil); got != nil {
		t.Fatalf("CloneStringMap(nil) = %v", got)
	}
	src := map[string]string{"A": "1"}
	dst := CloneStringMap(src)
	if len(dst) != 1 || dst["A"] != "1" {
		t.Fatalf("CloneStringMap() = %v", dst)
	}
	dst["A"] = "2"
	if src["A"] != "1" {
		t.Fatalf("CloneStringMap() aliases the source")
	}
}

func TestNormalizeWorkingDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}
	if got := NormalizeWorkingDir(""); got != home {
		t.Fatalf("NormalizeWorkingDir(\"\") = %q, want the home directory", got)
	}
	got := NormalizeWorkingDir("relative/dir")
	if !filepath.IsAbs(got) || !strings.HasSuffix(got, "relative/dir") {
		t.Fatalf("NormalizeWorkingDir(relative) = %q, want an absolute path", got)
	}
	abs := filepath.Join(t.TempDir(), "nested")
	if got := NormalizeWorkingDir(abs); got != abs {
		t.Fatalf("NormalizeWorkingDir(abs) = %q, want %q", got, abs)
	}
}

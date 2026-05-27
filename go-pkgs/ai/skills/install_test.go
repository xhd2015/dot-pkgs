package skills

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleInstallCodex(t *testing.T) {
	tmpDir := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(prevWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tempdir: %v", err)
	}

	err = HandleInstall(InstallOptions{
		SkillDirName: "example-skill",
		SkillContent: "# test skill\n",
	}, []string{"--codex"})
	if err != nil {
		t.Fatalf("HandleInstall(--codex): %v", err)
	}

	skillFile := filepath.Join(tmpDir, ".codex", "skills", "example-skill", "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# test skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}
}

func TestHandleInstallOpencode(t *testing.T) {
	tmpDir := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(prevWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tempdir: %v", err)
	}

	err = HandleInstall(InstallOptions{
		SkillDirName: "example-skill",
		SkillContent: "# test skill\n",
	}, []string{"--opencode"})
	if err != nil {
		t.Fatalf("HandleInstall(--opencode): %v", err)
	}

	skillFile := filepath.Join(tmpDir, ".opencode", "skills", "example-skill", "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# test skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}
}

func TestHandleInstallGeneralAgents(t *testing.T) {
	tmpDir := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(prevWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tempdir: %v", err)
	}

	err = HandleInstall(InstallOptions{
		SkillDirName: "example-skill",
		SkillContent: "# test skill\n",
	}, []string{"--general-agents"})
	if err != nil {
		t.Fatalf("HandleInstall(--general-agents): %v", err)
	}

	skillFile := filepath.Join(tmpDir, ".agents", "skills", "example-skill", "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# test skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}
}

func TestHandleInstallDefaultsToGeneralAgents(t *testing.T) {
	tmpDir := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(prevWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tempdir: %v", err)
	}

	err = HandleInstall(InstallOptions{
		SkillDirName: "example-skill",
		SkillContent: "# test skill\n",
	}, nil)
	if err != nil {
		t.Fatalf("HandleInstall(default): %v", err)
	}

	skillFile := filepath.Join(tmpDir, ".agents", "skills", "example-skill", "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# test skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}
}

func TestHandleInstallMultipleFlags(t *testing.T) {
	tmpDir := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(prevWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir tempdir: %v", err)
	}

	err = HandleInstall(InstallOptions{
		SkillDirName: "example-skill",
		SkillContent: "# multi skill\n",
	}, []string{"--codex", "--opencode", "--general-agents"})
	if err != nil {
		t.Fatalf("HandleInstall(--codex --opencode --general-agents): %v", err)
	}

	codexFile := filepath.Join(tmpDir, ".codex", "skills", "example-skill", "SKILL.md")
	opencodeFile := filepath.Join(tmpDir, ".opencode", "skills", "example-skill", "SKILL.md")
	generalAgentsFile := filepath.Join(tmpDir, ".agents", "skills", "example-skill", "SKILL.md")
	for _, f := range []string{codexFile, opencodeFile, generalAgentsFile} {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if string(content) != "# multi skill\n" {
			t.Fatalf("unexpected content in %s: %q", f, string(content))
		}
	}
}

func TestHandleInstallOverwritesExistingDirectoryByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "example-skill")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "SKILL.md"), []byte("# old skill\n"), 0644); err != nil {
		t.Fatalf("write old skill: %v", err)
	}
	staleFile := filepath.Join(targetDir, "stale.txt")
	if err := os.WriteFile(staleFile, []byte("stale"), 0644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	stdout, err := captureStdout(t, func() error {
		return HandleInstall(InstallOptions{
			SkillDirName: "example-skill",
			SkillContent: "# new skill\n",
		}, []string{targetDir})
	})
	if err != nil {
		t.Fatalf("HandleInstall: %v\nstdout:\n%s", err, stdout)
	}
	if strings.Contains(stdout, "md5") {
		t.Fatalf("stdout should not include md5:\n%s", stdout)
	}

	content, err := os.ReadFile(filepath.Join(targetDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# new skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Fatalf("stale file still exists or stat failed with unexpected error: %v", err)
	}
}

func TestHandleInstallSkipsUnchangedSkillFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "example-skill")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	skillFile := filepath.Join(targetDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# same skill\n"), 0644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}
	staleFile := filepath.Join(targetDir, "stale.txt")
	if err := os.WriteFile(staleFile, []byte("stale"), 0644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	stdout, err := captureStdout(t, func() error {
		return HandleInstall(InstallOptions{
			SkillDirName: "example-skill",
			SkillContent: "# same skill\n",
		}, []string{targetDir})
	})
	if err != nil {
		t.Fatalf("HandleInstall: %v\nstdout:\n%s", err, stdout)
	}
	if !strings.Contains(stdout, "Skill is up to date: "+targetDir) {
		t.Fatalf("stdout missing up-to-date log:\n%s", stdout)
	}
	if strings.Contains(stdout, "md5") {
		t.Fatalf("stdout should not include md5 for up-to-date install:\n%s", stdout)
	}
	if _, err := os.Stat(staleFile); err != nil {
		t.Fatalf("stale file should be kept when skill is unchanged: %v", err)
	}
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# same skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}
}

func TestHandleInstallNoOverrideKeepsExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "example-skill")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	skillFile := filepath.Join(targetDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# old skill\n"), 0644); err != nil {
		t.Fatalf("write old skill: %v", err)
	}

	withNonInteractiveStdin(t, func() {
		err := HandleInstall(InstallOptions{
			SkillDirName: "example-skill",
			SkillContent: "# new skill\n",
		}, []string{"--no-override", targetDir})
		if err != nil {
			t.Fatalf("HandleInstall(--no-override): %v", err)
		}
	})

	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# old skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}
}

func TestHandleInstallDryRunDoesNotOverwriteExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "example-skill")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	skillFile := filepath.Join(targetDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# old skill\n"), 0644); err != nil {
		t.Fatalf("write old skill: %v", err)
	}

	err := HandleInstall(InstallOptions{
		SkillDirName: "example-skill",
		SkillContent: "# new skill\n",
	}, []string{"--dry-run", targetDir})
	if err != nil {
		t.Fatalf("HandleInstall(--dry-run): %v", err)
	}

	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# old skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}
}

func TestHandleInstallWritesExtraFiles(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "example-skill")

	err := HandleInstall(InstallOptions{
		SkillDirName: "example-skill",
		SkillContent: "# skill\n",
		ExtraFiles: []InstallFile{
			{Path: "topics/flags.md", Content: []byte("# flags\n")},
		},
	}, []string{targetDir})
	if err != nil {
		t.Fatalf("HandleInstall: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(targetDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if string(content) != "# skill\n" {
		t.Fatalf("unexpected skill content: %q", string(content))
	}

	content, err = os.ReadFile(filepath.Join(targetDir, "topics", "flags.md"))
	if err != nil {
		t.Fatalf("read extra file: %v", err)
	}
	if string(content) != "# flags\n" {
		t.Fatalf("unexpected extra file content: %q", string(content))
	}
}

func withNonInteractiveStdin(t *testing.T, fn func()) {
	t.Helper()
	stdinPath := filepath.Join(t.TempDir(), "stdin")
	if err := os.WriteFile(stdinPath, nil, 0644); err != nil {
		t.Fatalf("write stdin fixture: %v", err)
	}
	stdin, err := os.Open(stdinPath)
	if err != nil {
		t.Fatalf("open stdin fixture: %v", err)
	}
	defer stdin.Close()

	prevStdin := os.Stdin
	os.Stdin = stdin
	defer func() {
		os.Stdin = prevStdin
	}()

	fn()
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	stdoutCh := make(chan []byte, 1)
	readErrCh := make(chan error, 1)
	go func() {
		data, readErr := io.ReadAll(reader)
		stdoutCh <- data
		readErrCh <- readErr
	}()

	os.Stdout = writer
	runErr := fn()
	os.Stdout = oldStdout
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	data := <-stdoutCh
	if err := <-readErrCh; err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return string(data), runErr
}

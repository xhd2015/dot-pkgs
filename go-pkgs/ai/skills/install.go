package skills

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

type InstallOptions struct {
	SkillDirName string
	// CursorDirName is kept for compatibility with older callers.
	CursorDirName string
	SkillContent  string
	ExtraFiles    []InstallFile
	Usage         string
}

type InstallFile struct {
	Path    string
	Content []byte
}

func HandleInstall(opts InstallOptions, args []string) error {
	usage := strings.TrimSpace(opts.Usage)
	if usage == "" {
		usage = "install"
	}
	skillDirName := opts.SkillDirName
	if skillDirName == "" {
		skillDirName = opts.CursorDirName
	}
	var dryRun bool
	var cursor bool
	var codex bool
	var opencode bool
	var generalAgents bool
	var noOverride bool
	var force bool
	args, err := flags.Bool("--dry-run", &dryRun).
		Bool("--cursor", &cursor).
		Bool("--codex", &codex).
		Bool("--opencode", &opencode).
		Bool("--general-agents", &generalAgents).
		Bool("--no-override", &noOverride).
		Bool("--force", &force).
		Help("-h,--help", fmt.Sprintf(`
Usage: %s [OPTIONS] [<dir>]

Install skill SKILL.md to a directory.
When no <dir> or target flag is provided, install to .agents/skills/%s.

Options:
  --cursor     Install to .cursor/skills/%s (no dir argument needed)
  --codex      Install to .codex/skills/%s (no dir argument needed)
  --opencode   Install to .opencode/skills/%s (no dir argument needed)
  --general-agents
               Install to .agents/skills/%s (no dir argument needed)
  --no-override
               Do not automatically overwrite an existing non-empty directory
  --dry-run    Show what would be created without actually creating anything

Multiple --cursor/--codex/--opencode/--general-agents flags can be combined to install to several locations at once.
`, usage, skillDirName, skillDirName, skillDirName, skillDirName, skillDirName)).Parse(args)
	if err != nil {
		return err
	}
	if force {
		noOverride = false
	}

	var dirs []string
	if cursor {
		dirs = append(dirs, filepath.Join(".cursor", "skills", skillDirName))
	}
	if codex {
		dirs = append(dirs, filepath.Join(".codex", "skills", skillDirName))
	}
	if opencode {
		dirs = append(dirs, filepath.Join(".opencode", "skills", skillDirName))
	}
	if generalAgents {
		dirs = append(dirs, filepath.Join(".agents", "skills", skillDirName))
	}
	if len(dirs) == 0 {
		if len(args) > 0 {
			dirs = append(dirs, args[0])
		} else {
			dirs = append(dirs, filepath.Join(".agents", "skills", skillDirName))
		}
	}

	for _, dir := range dirs {
		if err := installTo(dir, opts.SkillContent, opts.ExtraFiles, dryRun, noOverride); err != nil {
			return err
		}
	}
	return nil
}

func installTo(dir string, skillContent string, extraFiles []InstallFile, dryRun bool, noOverride bool) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	files, err := resolveInstallFiles(absDir, extraFiles)
	if err != nil {
		return err
	}

	entries, readErr := os.ReadDir(absDir)
	if readErr != nil && !os.IsNotExist(readErr) {
		return fmt.Errorf("read directory %s: %w", absDir, readErr)
	}

	exists := readErr == nil
	nonEmpty := exists && len(entries) > 0
	needsConfirmation := nonEmpty && noOverride
	willOverwrite := exists && !noOverride

	skillFile := filepath.Join(absDir, "SKILL.md")
	newContent := []byte(skillContent)
	_, same, compareErr := sameFileMD5(skillFile, newContent)
	if compareErr != nil {
		return compareErr
	}
	if same {
		for _, f := range files {
			_, fileSame, compareErr := sameFileMD5(f.dest, f.content)
			if compareErr != nil {
				return compareErr
			}
			if !fileSame {
				same = false
				break
			}
		}
	}

	if dryRun {
		if same {
			fmt.Printf("[dry-run] Skill is up to date: %s\n", absDir)
			return nil
		} else if willOverwrite {
			fmt.Printf("[dry-run] Would overwrite directory: %s\n", absDir)
		} else if needsConfirmation {
			fmt.Printf("[dry-run] Would require confirmation before overwriting directory: %s\n", absDir)
		} else if !exists {
			fmt.Printf("[dry-run] Would create directory: %s\n", absDir)
		} else {
			fmt.Printf("[dry-run] Would use existing directory: %s\n", absDir)
		}
		fmt.Printf("[dry-run] Would create file: %s\n", skillFile)
		for _, f := range files {
			fmt.Printf("[dry-run] Would create file: %s\n", f.dest)
		}
		return nil
	}

	if same {
		fmt.Printf("Skill is up to date: %s\n", absDir)
		return nil
	}

	if needsConfirmation {
		if !confirmOverwrite(absDir) {
			fmt.Println("Aborted.")
			return nil
		}
		willOverwrite = true
	}

	if willOverwrite {
		if err := os.RemoveAll(absDir); err != nil {
			return fmt.Errorf("remove directory %s: %w", absDir, err)
		}
		exists = false
	}

	if !exists {
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", absDir, err)
		}
	}

	if err := os.WriteFile(skillFile, newContent, 0644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f.dest), 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", filepath.Dir(f.dest), err)
		}
		if err := os.WriteFile(f.dest, f.content, 0644); err != nil {
			return fmt.Errorf("write %s: %w", f.dest, err)
		}
	}

	fmt.Printf("Installed skill to: %s\n", absDir)
	fmt.Printf("  - %s\n", skillFile)
	for _, f := range files {
		fmt.Printf("  - %s\n", f.dest)
	}
	return nil
}

type resolvedInstallFile struct {
	dest    string
	content []byte
}

func resolveInstallFiles(absDir string, extraFiles []InstallFile) ([]resolvedInstallFile, error) {
	files := make([]resolvedInstallFile, 0, len(extraFiles))
	for _, f := range extraFiles {
		rel := filepath.Clean(filepath.FromSlash(f.Path))
		if rel == "." || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return nil, fmt.Errorf("invalid install file path: %s", f.Path)
		}
		if rel == "SKILL.md" {
			return nil, fmt.Errorf("extra install file cannot replace SKILL.md")
		}
		files = append(files, resolvedInstallFile{
			dest:    filepath.Join(absDir, rel),
			content: f.Content,
		})
	}
	return files, nil
}

func sameFileMD5(path string, content []byte) (string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("open existing SKILL.md %s: %w", path, err)
	}
	defer f.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", false, fmt.Errorf("read existing SKILL.md %s: %w", path, err)
	}
	currentMD5 := fmt.Sprintf("%x", hash.Sum(nil))
	return currentMD5, currentMD5 == md5Hex(content), nil
}

func md5Hex(content []byte) string {
	return fmt.Sprintf("%x", md5.Sum(content))
}

func confirmOverwrite(dir string) bool {
	f, _ := os.Stdin.Stat()
	if f == nil || (f.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	fmt.Printf("Directory %s is not empty. Overwrite? [y/N] ", dir)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

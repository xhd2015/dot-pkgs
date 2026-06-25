package main

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
)

const help = `
Usage: git-hook-github-workflow-test [OPTIONS]

Ensure GitHub repositories have a Go test workflow.

Options:
  --fix                           create .github/workflows/test.yml when missing
  --origin-domain DOMAIN          only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN  skip when remote origin host matches DOMAIN
  -h, --help                      show help message
`

const workflowPath = ".github/workflows/test.yml"

//go:embed test.yml
var workflowTemplate string

type config struct {
	domainFilter githook.DomainFilter
	fix          bool
	showHelp     bool
}

type goModule struct {
	Dir       string
	GoVersion string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "git-hook-github-workflow-test: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithOutput(args, os.Stdout)
}

func runWithOutput(args []string, out io.Writer) error {
	cfg, err := parseArgs(args)
	if err != nil {
		return err
	}
	if cfg.showHelp {
		fmt.Fprint(out, strings.TrimPrefix(help, "\n"))
		return nil
	}

	shouldRun, err := cfg.domainFilter.ShouldRun()
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	host, err := originHost()
	if err != nil {
		return err
	}
	if !strings.EqualFold(host, "github.com") {
		if cfg.fix {
			return fmt.Errorf("--fix requires remote origin host to be github.com, got %q", host)
		}
		return nil
	}

	root, err := repoRoot()
	if err != nil {
		return err
	}

	modules, err := discoverGoModules(root)
	if err != nil {
		return err
	}
	expectedWorkflow, err := workflowContent(modules)
	if err != nil {
		return err
	}

	workflowFile := filepath.Join(root, workflowPath)
	exists, err := fileExists(workflowFile)
	if err != nil {
		return err
	}
	if exists {
		current, err := os.ReadFile(workflowFile)
		if err != nil {
			return fmt.Errorf("read %s: %w", workflowPath, err)
		}
		if string(current) == expectedWorkflow {
			if cfg.fix {
				fmt.Fprintf(out, "%s already matches recommended workflow; nothing changed.\n", workflowPath)
			}
			return nil
		}
		if !cfg.fix {
			fmt.Fprintf(out, "warning: %s differs from recommended workflow. Run git-hook-github-workflow-test --fix to update it.\n", workflowPath)
			return nil
		}
		if err := os.WriteFile(workflowFile, []byte(expectedWorkflow), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", workflowPath, err)
		}
		fmt.Fprintf(out, "updated %s\n", workflowPath)
		return nil
	}

	if !exists {
		if cfg.fix {
			if err := os.MkdirAll(filepath.Join(root, ".github", "workflows"), 0o755); err != nil {
				return fmt.Errorf("create .github/workflows: %w", err)
			}
			if err := os.WriteFile(workflowFile, []byte(expectedWorkflow), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", workflowPath, err)
			}
			fmt.Fprintf(out, "created %s\n", workflowPath)
			return nil
		}
		fmt.Fprintf(out, "warning: %s is missing. Run git-hook-github-workflow-test --fix to create it.\n", workflowPath)
		return nil
	}
	return nil
}

func parseArgs(args []string) (config, error) {
	var cfg config
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if matched, next, err := githook.ParseDomainFlag(args, i, &cfg.domainFilter); matched {
			if err != nil {
				return cfg, err
			}
			i = next
			continue
		}
		switch {
		case arg == "-h" || arg == "--help":
			cfg.showHelp = true
			return cfg, nil
		case arg == "--fix":
			cfg.fix = true
		case strings.HasPrefix(arg, "-"):
			return cfg, fmt.Errorf("unknown flag: %s", arg)
		default:
			return cfg, fmt.Errorf("unexpected arg: %s", arg)
		}
	}
	if err := cfg.domainFilter.Normalize(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func originHost() (string, error) {
	remote, ok, err := githook.GitOptionalOutput("config", "--get", "remote.origin.url")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return githook.OriginHost(strings.TrimSpace(remote)), nil
}

func repoRoot() (string, error) {
	root, err := githook.GitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return filepath.Clean(strings.TrimSpace(root)), nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat %s: %w", path, err)
}

func discoverGoModules(root string) ([]goModule, error) {
	if files, ok, err := gitVisibleFiles(root); err != nil {
		return nil, err
	} else if ok {
		return goModulesFromFiles(root, files)
	}

	return walkGoModules(root)
}

func gitVisibleFiles(root string) ([]string, bool, error) {
	cmd := exec.Command("git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 128 {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("git ls-files failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if len(output) == 0 {
		return nil, true, nil
	}
	rawFiles := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
	files := make([]string, 0, len(rawFiles))
	for _, file := range rawFiles {
		if file != "" {
			files = append(files, file)
		}
	}
	return files, true, nil
}

func goModulesFromFiles(root string, files []string) ([]goModule, error) {
	var modules []goModule
	for _, file := range files {
		if filepath.Base(file) != "go.mod" {
			continue
		}
		if hasTestdataPath(file) {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(file))
		version, err := goModVersion(path)
		if err != nil {
			return nil, err
		}
		dir := filepath.Dir(file)
		if dir == "." {
			dir = ""
		}
		modules = append(modules, goModule{
			Dir:       filepath.ToSlash(dir),
			GoVersion: version,
		})
	}
	return modules, nil
}

func walkGoModules(root string) ([]goModule, error) {
	var modules []goModule
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".github", "vendor", "node_modules", "testdata":
				if path != root {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if d.Name() != "go.mod" {
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if hasTestdataPath(relPath) {
			return nil
		}
		version, err := goModVersion(path)
		if err != nil {
			return err
		}
		dir := filepath.Dir(path)
		relDir, err := filepath.Rel(root, dir)
		if err != nil {
			return err
		}
		if relDir == "." {
			dir = ""
		} else {
			dir = relDir
		}
		modules = append(modules, goModule{
			Dir:       filepath.ToSlash(dir),
			GoVersion: version,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover go modules: %w", err)
	}
	return modules, nil
}

func hasTestdataPath(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == "testdata" {
			return true
		}
	}
	return false
}

func goModVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[0] == "go" {
			return fields[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan %s: %w", path, err)
	}
	return "", fmt.Errorf("%s missing go directive", path)
}

func workflowContent(modules []goModule) (string, error) {
	goVersion := "latest"
	for _, module := range modules {
		if module.GoVersion == "" {
			continue
		}
		if goVersion == "latest" || compareGoVersion(module.GoVersion, goVersion) > 0 {
			goVersion = module.GoVersion
		}
	}
	content := strings.ReplaceAll(workflowTemplate, "__GO_VERSION__", goVersion)
	content = strings.ReplaceAll(content, "__GO_TEST_STEPS__", goTestSteps(modules))
	return content, nil
}

func goTestSteps(modules []goModule) string {
	if len(modules) == 0 {
		return "      - name: Go test\n        run: echo \"no go.mod files found; skipping go test\"\n"
	}
	var b strings.Builder
	for _, module := range modules {
		name := "Go test"
		cmd := "go test -v ./..."
		if module.Dir != "" {
			name = "Go test " + module.Dir
			cmd = "go -C " + module.Dir + " test -v ./..."
		}
		fmt.Fprintf(&b, "      - name: %s\n        run: %s\n", name, cmd)
	}
	return b.String()
}

func compareGoVersion(a string, b string) int {
	ap := parseVersionParts(a)
	bp := parseVersionParts(b)
	for i := 0; i < len(ap) || i < len(bp); i++ {
		var av, bv int
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		if av > bv {
			return 1
		}
		if av < bv {
			return -1
		}
	}
	return 0
}

func parseVersionParts(version string) []int {
	fields := strings.FieldsFunc(version, func(r rune) bool {
		return r == '.' || r == '-'
	})
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil {
			break
		}
		parts = append(parts, n)
	}
	return parts
}

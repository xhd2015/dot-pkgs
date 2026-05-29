package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
)

const help = `
Usage: git-hook-no-commit-sub-module [OPTIONS]

Reject staged files that belong to a separate git repository (submodule)
to prevent accidentally committing submodule contents.

Options:
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  -h, --help                        show help message
`

var errSubModuleFound = errors.New("submodule files found in staged changes")

type config struct {
	domainFilter githook.DomainFilter
	showHelp     bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errSubModuleFound) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git-hook-no-commit-sub-module: %v\n", err)
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

	files, err := stagedFiles()
	if err != nil {
		return err
	}

	subModules := detectSubModules(files)
	if len(subModules) > 0 {
		fmt.Fprintln(out, "Staged files belong to submodules:")
		for _, sm := range subModules {
			fmt.Fprintf(out, "  %s/\n", sm)
		}
		fmt.Fprintln(out, "\nUse git rm --cached <file> to unstage, or use git submodule instead.")
		return errSubModuleFound
	}
	return nil
}

func detectSubModules(files []string) []string {
	seen := make(map[string]bool)
	var subModules []string
	for _, f := range files {
		dir := filepath.Dir(f)
		for {
			if dir == "." {
				break
			}
			gitPath := filepath.Join(dir, ".git")
			if info, err := os.Stat(gitPath); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
				if !seen[dir] {
					seen[dir] = true
					subModules = append(subModules, dir)
				}
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return subModules
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

func stagedFiles() ([]string, error) {
	output, err := githook.GitOutput("diff", "--cached", "--name-only", "--diff-filter=ACMRT", "--")
	if err != nil {
		return nil, err
	}
	var files []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			files = append(files, name)
		}
	}
	return files, scanner.Err()
}

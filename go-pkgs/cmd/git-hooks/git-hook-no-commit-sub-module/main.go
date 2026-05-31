package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/submodule"
)

const help = `
Usage: git-hook-no-commit-sub-module [OPTIONS]

Reject staged files that belong to a separate git repository (submodule)
to prevent accidentally committing submodule contents.

Options:
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  --auto-unstage                    automatically unstage submodule files instead of
                                    failing (use for hooks that run early)
  -h, --help                        show help message
`

var errSubModuleFound = errors.New("submodule files found in staged changes")

type config struct {
	domainFilter githook.DomainFilter
	showHelp     bool
	autoUnstage  bool
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

	subModules := submodule.DetectSubModules(files)
	if len(subModules) > 0 {
		fmt.Fprintln(out, "Staged files belong to submodules:")
		for _, sm := range subModules {
			fmt.Fprintf(out, "  %s/\n", sm)
		}
		if cfg.autoUnstage {
			var paths []string
			for _, f := range files {
				for _, smDir := range subModules {
					if strings.HasPrefix(f, smDir+"/") || f == smDir {
						paths = append(paths, f)
						break
					}
				}
			}
			if err := githook.RestoreStaged(paths...); err != nil {
				return err
			}
			return nil
		}
		fmt.Fprintln(out, "\nUse git restore --staged <file> to unstage, or use git submodule instead.")
		return errSubModuleFound
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
		case arg == "--auto-unstage":
			cfg.autoUnstage = true
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

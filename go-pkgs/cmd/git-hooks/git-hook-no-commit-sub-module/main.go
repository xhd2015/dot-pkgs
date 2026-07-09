package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/submodule"
	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/gitops/gitwrite"
	lessflags "github.com/xhd2015/less-flags"
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
  -h,--help                         show help message
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
	cfg, err := parseArgs(args, out)
	if err != nil {
		return err
	}
	if cfg.showHelp {
		return nil
	}

	shouldRun, err := cfg.domainFilter.ShouldRun()
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	files, err := git.GetStagedFiles(".")
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
			if err := gitwrite.RestoreStaged(".", paths...); err != nil {
				return err
			}
			return nil
		}
		fmt.Fprintln(out, "\nUse git restore --staged <file> to unstage, or use git submodule instead.")
		return errSubModuleFound
	}
	return nil
}

func parseArgs(args []string, out io.Writer) (config, error) {
	var cfg config
	var originDomain *string
	var excludeOriginDomain *string
	var autoUnstage *bool

	remaining, err := lessflags.
		String("--origin-domain", &originDomain).
		String("--exclude-origin-domain", &excludeOriginDomain).
		Bool("--auto-unstage", &autoUnstage).
		HelpFunc("-h,--help", func() {
			fmt.Fprint(out, strings.TrimPrefix(help, "\n"))
		}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if errors.Is(err, lessflags.ErrHelp) {
			cfg.showHelp = true
			return cfg, nil
		}
		return cfg, mapUnknownFlagErr(err)
	}
	if len(remaining) > 0 {
		return cfg, fmt.Errorf("unexpected arg: %s", remaining[0])
	}

	if originDomain != nil {
		cfg.domainFilter.OriginDomain = *originDomain
	}
	if excludeOriginDomain != nil {
		cfg.domainFilter.ExcludeOriginDomain = *excludeOriginDomain
	}
	if autoUnstage != nil {
		cfg.autoUnstage = *autoUnstage
	}
	if err := cfg.domainFilter.Normalize(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func mapUnknownFlagErr(err error) error {
	const prefix = "unrecognized flag: "
	if msg := err.Error(); strings.HasPrefix(msg, prefix) {
		return fmt.Errorf("unknown flag: %s", strings.TrimPrefix(msg, prefix))
	}
	return err
}

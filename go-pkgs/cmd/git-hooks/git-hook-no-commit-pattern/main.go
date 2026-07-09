package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/gitops/gitwrite"
	lessflags "github.com/xhd2015/less-flags"
)

const help = `
Usage: git-hook-no-commit-pattern [OPTIONS] PATTERN...

Reject staged files whose paths match any glob pattern.

Options:
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  --auto-unstage                    automatically unstage matched files instead of
                                    failing (use for hooks that run early)
  -h,--help                         show help message
`

var errPatternsMatched = errors.New("patterns matched")

type config struct {
	domainFilter githook.DomainFilter
	showHelp     bool
	autoUnstage  bool
	patterns     []string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errPatternsMatched) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git-hook-no-commit-pattern: %v\n", err)
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

	var matched []string
	for _, file := range files {
		ok, err := matchesAny(file, cfg.patterns)
		if err != nil {
			return err
		}
		if ok {
			fmt.Fprintln(out, file)
			matched = append(matched, file)
		}
	}
	if len(matched) > 0 {
		if cfg.autoUnstage {
			if err := gitwrite.RestoreStaged(".", matched...); err != nil {
				return err
			}
			return nil
		}
		return errPatternsMatched
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

	if originDomain != nil {
		cfg.domainFilter.OriginDomain = *originDomain
	}
	if excludeOriginDomain != nil {
		cfg.domainFilter.ExcludeOriginDomain = *excludeOriginDomain
	}
	if autoUnstage != nil {
		cfg.autoUnstage = *autoUnstage
	}
	cfg.patterns = remaining

	if err := cfg.domainFilter.Normalize(); err != nil {
		return cfg, err
	}
	if len(cfg.patterns) == 0 && !cfg.showHelp {
		return cfg, fmt.Errorf("at least one pattern is required")
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

func matchesAny(name string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		matched, err := matchesPattern(pattern, name)
		if err != nil {
			return false, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func matchesPattern(pattern, name string) (bool, error) {
	if strings.Contains(pattern, "/") {
		return path.Match(pattern, name)
	}
	for _, segment := range strings.Split(name, "/") {
		matched, err := path.Match(pattern, segment)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

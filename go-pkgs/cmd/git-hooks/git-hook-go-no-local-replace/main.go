package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	lessflags "github.com/xhd2015/less-flags"
)

const help = `
Usage: git-hook-go-no-local-replace [OPTIONS]

Reject local path replace directives in go.mod files.

Options:
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  --strict                          block all local replaces (including intra-repo)
  -h,--help                         show help message
`

var errLocalReplaceFound = errors.New("local replace found")

type config struct {
	domainFilter githook.DomainFilter
	showHelp     bool
	strict       bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errLocalReplaceFound) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git-hook-go-no-local-replace: %v\n", err)
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

	cwd, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	issues, err := replace.CheckLocalReplaces(cwd)
	if err != nil {
		return err
	}

	scanTop, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}

	found := false
	for _, issue := range issues {
		// Lenient default: skip intra-repo replaces.
		if issue.IsIntraRepo && !cfg.strict {
			continue
		}
		fmt.Fprintln(out, replace.FormatIssueLine(scanTop, issue))
		found = true
	}
	if found {
		return errLocalReplaceFound
	}
	return nil
}

func parseArgs(args []string, out io.Writer) (config, error) {
	var cfg config
	var originDomain *string
	var excludeOriginDomain *string
	var strict *bool

	remaining, err := lessflags.
		String("--origin-domain", &originDomain).
		String("--exclude-origin-domain", &excludeOriginDomain).
		Bool("--strict", &strict).
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
	if strict != nil {
		cfg.strict = *strict
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

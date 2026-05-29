package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: git-hook-wip-protect [OPTIONS]

Reject WIP (work-in-progress) commits to prevent accidentally committing
or pushing unfinished work.

Options:
  --phase PHASE                     hook phase: pre-commit or push
                                    (default: $GIT_HOOK_PHASE if set, otherwise pre-commit)
  --is-amend                        is running pre-commit's amend mode
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  -h, --help                        show help message
`

var errWipProtected = errors.New("wip commit detected")

type config struct {
	domainFilter githook.DomainFilter
	phase        string
	amend        bool
	showHelp     bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errWipProtected) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git-hook-wip-protect: %v\n", err)
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

	isWip, msg, err := checkHeadIsWip()
	if err != nil {
		return err
	}
	if !isWip {
		return nil
	}

	switch cfg.phase {
	case "push":
		fmt.Fprintf(out, "HEAD commit is a WIP: %s\n", msg)
		fmt.Fprintln(out, "Cannot push WIP commits. Please rewrite the commit message first.")
		return errWipProtected
	default:
		if cfg.amend {
			return nil
		}
		fmt.Fprintf(out, "HEAD commit is a WIP: %s\n", msg)
		fmt.Fprintln(out, "Rewrite the commit message with git commit --amend.")
		return errWipProtected
	}
}

func parseArgs(args []string, out io.Writer) (config, error) {
	var cfg config
	cfg.phase = phaseFromEnv()

	var phase *string
	var isAmendFlag *bool
	var originDomain *string
	var excludeOriginDomain *string

	_, err := flags.
		String("--phase", &phase).
		Bool("--is-amend", &isAmendFlag).
		String("--origin-domain", &originDomain).
		String("--exclude-origin-domain", &excludeOriginDomain).
		HelpFunc("-h,--help", func() {
			fmt.Fprint(out, strings.TrimPrefix(help, "\n"))
		}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if errors.Is(err, flags.ErrHelp) {
			cfg.showHelp = true
			return cfg, nil
		}
		return cfg, fmt.Errorf("parse flags: %w", err)
	}

	if phase != nil {
		cfg.phase = *phase
	}

	var effectiveIsAmend bool
	if isAmendFlag != nil {
		effectiveIsAmend = *isAmendFlag
	} else {
		effectiveIsAmend = os.Getenv("GIT_HOOK_AMEND") == "1"
	}
	cfg.amend = effectiveIsAmend

	if originDomain != nil {
		cfg.domainFilter.OriginDomain = *originDomain
	}
	if excludeOriginDomain != nil {
		cfg.domainFilter.ExcludeOriginDomain = *excludeOriginDomain
	}

	if err := cfg.domainFilter.Normalize(); err != nil {
		return cfg, err
	}

	switch cfg.phase {
	case "pre-commit", "push":
	default:
		return cfg, fmt.Errorf("unknown phase: %s (expected pre-commit or push)", cfg.phase)
	}

	return cfg, nil
}

func phaseFromEnv() string {
	if phase := os.Getenv("GIT_HOOK_PHASE"); phase != "" {
		return phase
	}
	return "pre-commit"
}

func checkHeadIsWip() (bool, string, error) {
	msg, ok, err := githook.GitOptionalOutput("log", "-1", "--format=%B")
	if err != nil {
		return false, "", nil
	}
	if !ok {
		return false, "", nil
	}
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return false, "", nil
	}
	lower := strings.ToLower(msg)
	if lower == "wip" || strings.HasPrefix(lower, "wip:") {
		return true, msg, nil
	}
	return false, "", nil
}

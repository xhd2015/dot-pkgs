package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	scan "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

const help = `
Usage: git-hook-go-no-local-replace [OPTIONS]

Reject local path replace directives in go.mod files.

Options:
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  -h, --help                        show help message
`

var errLocalReplaceFound = errors.New("local replace found")

type config struct {
	domainFilter githook.DomainFilter
	showHelp     bool
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

	modules, err := scan.Scan(".", scan.Options{})
	if err != nil {
		return err
	}

	found := false
	for _, module := range modules {
		for _, replace := range module.Replaces {
			if replace.NewVersion == "" {
				fmt.Fprintln(out, replace.NewPath)
				found = true
			}
		}
	}
	if found {
		return errLocalReplaceFound
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

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	"github.com/xhd2015/dot-pkgs/go-pkgs/file/detect"
)

const help = `
Usage: git-hook-no-commit-binary [OPTIONS]

Reject staged binary files to prevent accidentally committing them.

Options:
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  -h, --help                        show help message
`

var errBinaryFilesFound = errors.New("binary files found")

type config struct {
	domainFilter githook.DomainFilter
	showHelp     bool
}

type binaryFile struct {
	path string
	desc string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errBinaryFilesFound) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git-hook-no-commit-binary: %v\n", err)
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

	var binaries []binaryFile
	for _, f := range files {
		desc, isBin, err := detect.DetectFileType(f)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("check file %s: %w", f, err)
		}
		if isBin {
			binaries = append(binaries, binaryFile{path: f, desc: desc})
		}
	}

	if len(binaries) > 0 {
		fmt.Fprintln(out, "Binary files detected in staged changes:")
		for _, bf := range binaries {
			fmt.Fprintf(out, "  %s (%s)\n", bf.path, bf.desc)
		}
		fmt.Fprintln(out, "\nUse git rm --cached <file> to unstage, or add to .gitignore if needed.")
		return errBinaryFilesFound
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

package scan_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

var errHelp = lessflags.ErrHelp

const cliHelp = `Usage: scan-repos [OPTIONS]

Discover git repositories under filesystem roots.

Options:
  --root PATH              Root directory to scan (required, repeatable)
  --max-depth N            Maximum directory depth relative to each root (0 = unlimited)
  --ignore-dir PATH        Full directory path to skip after normalization (repeatable)
  --ignore-dir-basename NAME Directory basename to skip anywhere in the tree (repeatable)
  -v, --verbose            Warn on stderr when skipping unreadable directories
  --list-remotes           List git remotes and append origin info to lines output
  --list-worktrees         Enrich main repos with git worktree metadata
  --json                   Output JSON array instead of tab-separated lines
  -h, --help               Show this help message
`

func (r RepoType) String() string {
	return string(r)
}

func RunCLI(args []string) error {
	var roots []string
	var ignoreDirs []string
	var ignoreDirBasenames []string
	var maxDepth int
	var verbose bool
	var listRemotes bool
	var listWorktrees bool
	var jsonOut bool

	remain, err := lessflags.StringSlice("--root", &roots).
		StringSlice("--ignore-dir", &ignoreDirs).
		StringSlice("--ignore-dir-basename", &ignoreDirBasenames).
		Int("--max-depth", &maxDepth).
		Bool("-v,--verbose", &verbose).
		Bool("--list-remotes", &listRemotes).
		Bool("--list-worktrees", &listWorktrees).
		Bool("--json", &jsonOut).
		Help("-h,--help", cliHelp).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == errHelp {
			return nil
		}
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	if len(remain) > 0 {
		err := fmt.Errorf("unrecognized arguments: %s", strings.Join(remain, " "))
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if len(roots) == 0 {
		err := fmt.Errorf("at least one --root is required")
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	normalizedIgnoreDirs := make([]string, 0, len(ignoreDirs))
	for _, dir := range ignoreDirs {
		norm, normErr := normalizeIgnoreDir(dir)
		if normErr != nil {
			err := fmt.Errorf("%s: %w", dir, normErr)
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		normalizedIgnoreDirs = append(normalizedIgnoreDirs, norm)
	}

	repos, err := Scan(context.Background(), Options{
		Roots:              roots,
		MaxDepth:           maxDepth,
		IgnoreDirs:         normalizedIgnoreDirs,
		IgnoreDirBasenames: ignoreDirBasenames,
		Verbose:            verbose,
		ListRemotes:        listRemotes,
		ListWorktrees:      listWorktrees,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if jsonOut {
		if repos == nil {
			repos = []Repo{}
		}
		data, err := json.Marshal(repos)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		fmt.Print(string(data))
		return nil
	}

	for _, repo := range repos {
		line := repo.Path + "\t" + repo.RepoType.String()
		if listRemotes {
			if origin := findOriginRemote(repo.Remotes); origin != nil {
				line += fmt.Sprintf("\torigin:%s/%s@%s", origin.Owner, origin.Repo, origin.Host)
			}
		}
		fmt.Println(line)
	}
	return nil
}

func findOriginRemote(remotes []Remote) *Remote {
	for i := range remotes {
		if remotes[i].Name == "origin" {
			return &remotes[i]
		}
	}
	return nil
}
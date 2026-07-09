package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

const cliHelp = `
kool github enhances GitHub CLI workflows.

Usage: kool github <cmd> [OPTIONS]

Available commands:
  repo                             repository commands
  help                             show help message

Examples:
  kool github repo list            list repositories
  kool github repo list --help     show list command help
`

const repoHelp = `
kool github repo - repository commands

Usage: kool github repo <cmd> [OPTIONS]

Available commands:
  list                             list repositories
  help                             show help message

Examples:
  kool github repo list            list repositories
  kool github repo list --help     show list command options
`

const listHelp = `
kool github repo list - List GitHub repositories

Usage: kool github repo list [OPTIONS]

Options:
  --search-description <keyword>  search repository name, description, and readme
  --search-code <keyword>         search repository code
  --owner <user>                  limit to owner (repeatable)
  --limit <n>                     maximum results per owner (default: 30)
  --json                          output JSON array of RepoResult
  -h,--help                       show help message

Examples:
  kool github repo list
  kool github repo list --owner alice --owner bob
  kool github repo list --search-description widget --json
`

// RunCLI is the entry point for the kool github command-line interface.
func RunCLI(args []string) error {
	err := runCLI(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}

func runCLI(args []string) error {
	if len(args) == 0 {
		fmt.Println(strings.TrimPrefix(cliHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Println(strings.TrimPrefix(cliHelp, "\n"))
		return nil
	case "repo":
		return runRepoCLI(args[1:])
	default:
		return fmt.Errorf("unrecognized github command: %s", args[0])
	}
}

func runRepoCLI(args []string) error {
	if len(args) == 0 {
		fmt.Println(strings.TrimPrefix(repoHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Println(strings.TrimPrefix(repoHelp, "\n"))
		return nil
	case "list":
		return runRepoList(args[1:])
	default:
		return fmt.Errorf("unrecognized repo command: %s", args[0])
	}
}

func runRepoList(args []string) error {
	var (
		searchDescription string
		searchCode        string
		owners            []string
		limit             int
		asJSON            bool
	)

	args, err := lessflags.
		String("--search-description", &searchDescription).
		String("--search-code", &searchCode).
		StringSlice("--owner", &owners).
		Int("--limit", &limit).
		Bool("--json", &asJSON).
		Help("-h,--help", listHelp).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			return nil
		}
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}

	results, err := ListRepos(context.Background(), ListReposOptions{
		Owners:            owners,
		SearchDescription: searchDescription,
		SearchCode:        searchCode,
		Limit:             limit,
	})
	if err != nil {
		return err
	}

	if asJSON {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("encode JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	for _, result := range results {
		reasons := make([]string, len(result.MatchedBy))
		for i, reason := range result.MatchedBy {
			reasons[i] = string(reason)
		}
		fmt.Fprintf(os.Stdout, "%s\t%s\n", result.FullName, strings.Join(reasons, ","))
	}
	return nil
}
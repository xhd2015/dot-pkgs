package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Repo is a normalized GitHub repository record.
type Repo struct {
	Owner       string
	Name        string
	FullName    string
	URL         string
	Description string
	IsFork      bool
	IsArchived  bool
}

// Options configures ListOwned.
type Options struct {
	Owners          []string
	Limit           int
	IncludeArchived bool
	IncludeForks    bool
	GhBin           string
}

// DefaultOptions returns Options with package defaults for the given owners.
func DefaultOptions(owners []string) Options {
	return Options{
		Owners:          owners,
		Limit:           100,
		IncludeArchived: false,
		IncludeForks:    true,
		GhBin:           ghBinFromEnv(),
	}
}

// ListOwned lists repositories owned by the configured GitHub users via gh.
func ListOwned(ctx context.Context, opts Options) ([]Repo, error) {
	if err := validateOwners(opts.Owners); err != nil {
		return nil, err
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 100
	}

	ghBin, err := resolveGhBin(opts.GhBin)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	repos := make([]Repo, 0)

	for _, owner := range opts.Owners {
		raw, runErr := runGh(ctx, ghBin, owner, limit, opts.IncludeArchived, opts.IncludeForks)
		if runErr != nil {
			return nil, runErr
		}

		parsed, parseErr := parseGhRepos(raw)
		if parseErr != nil {
			return nil, parseErr
		}

		for _, repo := range parsed {
			if _, ok := seen[repo.FullName]; ok {
				continue
			}
			seen[repo.FullName] = struct{}{}
			repos = append(repos, repo)
		}
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].FullName < repos[j].FullName
	})

	return repos, nil
}

func validateOwners(owners []string) error {
	if len(owners) == 0 {
		return fmt.Errorf("at least one owner is required")
	}
	for _, owner := range owners {
		if strings.TrimSpace(owner) == "" {
			return fmt.Errorf("invalid owner: empty owner string")
		}
	}
	return nil
}
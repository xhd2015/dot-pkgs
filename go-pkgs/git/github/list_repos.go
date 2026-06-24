package github

import (
	"context"
	"sort"
)

// MatchReason tags why a repository appeared in ListRepos results.
type MatchReason string

const (
	MatchOwned       MatchReason = "owned"
	MatchDescription MatchReason = "description"
	MatchCode        MatchReason = "code"
)

// RepoResult is a repository with match provenance.
type RepoResult struct {
	Repo
	MatchedBy []MatchReason `json:"matched_by"`
}

// ListReposOptions configures ListRepos.
type ListReposOptions struct {
	Owners            []string
	SearchDescription string
	SearchCode        string
	Limit             int
	GhBin             string
}

// ListRepos lists repositories via gh with optional search and matched_by provenance.
func ListRepos(ctx context.Context, opts ListReposOptions) ([]RepoResult, error) {
	ghBin, err := resolveGhBin(opts.GhBin)
	if err != nil {
		return nil, err
	}

	login, err := EnsureAuthenticated(ctx, ghBin)
	if err != nil {
		return nil, err
	}

	owners := opts.Owners
	if len(owners) > 0 {
		if err := validateOwners(owners); err != nil {
			return nil, err
		}
	} else {
		owners = []string{login}
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 30
	}

	searchDescription := opts.SearchDescription != ""
	searchCode := opts.SearchCode != ""

	if !searchDescription && !searchCode {
		return listReposOwned(ctx, ghBin, owners, limit)
	}

	return listReposSearch(ctx, ghBin, owners, opts.SearchDescription, opts.SearchCode, limit, searchDescription, searchCode)
}

func listReposOwned(ctx context.Context, ghBin string, owners []string, limit int) ([]RepoResult, error) {
	repos, err := ListOwned(ctx, Options{
		Owners:          owners,
		Limit:           limit,
		IncludeArchived: false,
		IncludeForks:    true,
		GhBin:           ghBin,
	})
	if err != nil {
		return nil, err
	}

	results := make([]RepoResult, 0, len(repos))
	for _, repo := range repos {
		results = append(results, RepoResult{
			Repo:      repo,
			MatchedBy: []MatchReason{MatchOwned},
		})
	}
	return results, nil
}

func listReposSearch(
	ctx context.Context,
	ghBin string,
	owners []string,
	searchDescription, searchCode string,
	limit int,
	runDescription, runCode bool,
) ([]RepoResult, error) {
	byFullName := make(map[string]*RepoResult)

	if runDescription {
		for _, owner := range owners {
			raw, runErr := runSearchRepos(ctx, ghBin, owner, searchDescription, limit)
			if runErr != nil {
				return nil, runErr
			}
			repos, parseErr := parseSearchRepos(raw)
			if parseErr != nil {
				return nil, parseErr
			}
			for _, repo := range repos {
				mergeRepoResult(byFullName, repo, MatchDescription)
			}
		}
	}

	if runCode {
		for _, owner := range owners {
			raw, runErr := runSearchCode(ctx, ghBin, owner, searchCode, limit)
			if runErr != nil {
				return nil, runErr
			}
			repos, parseErr := parseSearchCode(raw)
			if parseErr != nil {
				return nil, parseErr
			}
			for _, repo := range repos {
				mergeRepoResult(byFullName, repo, MatchCode)
			}
		}
	}

	results := make([]RepoResult, 0, len(byFullName))
	for _, result := range byFullName {
		results = append(results, *result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].FullName < results[j].FullName
	})

	return results, nil
}

func mergeRepoResult(byFullName map[string]*RepoResult, repo Repo, reason MatchReason) {
	existing, ok := byFullName[repo.FullName]
	if !ok {
		byFullName[repo.FullName] = &RepoResult{
			Repo:      repo,
			MatchedBy: []MatchReason{reason},
		}
		return
	}
	existing.MatchedBy = appendMatchReason(existing.MatchedBy, reason)
}

func appendMatchReason(reasons []MatchReason, reason MatchReason) []MatchReason {
	for _, existing := range reasons {
		if existing == reason {
			return reasons
		}
	}
	return append(reasons, reason)
}
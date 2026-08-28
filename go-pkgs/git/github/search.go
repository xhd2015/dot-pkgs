package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type ghSearchRepoWire struct {
	Name        string      `json:"name"`
	FullName    string      `json:"fullName"`
	URL         string      `json:"url"`
	Description string      `json:"description"`
	IsFork      bool        `json:"isFork"`
	IsArchived  bool        `json:"isArchived"`
	Owner       ghRepoOwner `json:"owner"`
}

type ghSearchCodeWire struct {
	Repository ghRepoWire `json:"repository"`
}

func runSearchRepos(ctx context.Context, ghBin, owner, keyword string, limit int) ([]byte, error) {
	args := []string{
		"search", "repos", keyword,
		"--owner", owner,
		"--json", "name,fullName,url,description,owner,isFork,isArchived",
		"--limit", fmt.Sprintf("%d", limit),
	}
	return runGhCommand(ctx, ghBin, "search repos", owner, args)
}

func runSearchCode(ctx context.Context, ghBin, owner, keyword string, limit int) ([]byte, error) {
	args := []string{
		"search", "code", keyword,
		"--owner", owner,
		"--json", "repository",
		"--limit", fmt.Sprintf("%d", limit),
	}
	return runGhCommand(ctx, ghBin, "search code", owner, args)
}

func runGhCommand(ctx context.Context, ghBin, subcommand, owner string, args []string) ([]byte, error) {
	stdout, stderr, err := outputWithETXTBSYRetry(ctx, ghBin, args...)
	if err != nil {
		return nil, wrapGhSubcommandError(subcommand, owner, err, stderr)
	}
	return stdout, nil
}

func wrapGhSubcommandError(subcommand, owner string, err error, stderr string) error {
	stderr = strings.TrimSpace(stderr)

	if exitErr, ok := err.(*exec.ExitError); ok {
		code := exitErr.ExitCode()
		lower := strings.ToLower(stderr)
		if code == 4 && strings.Contains(lower, "auth") {
			if stderr != "" {
				return fmt.Errorf("%s: run `gh auth login` to authenticate", stderr)
			}
			return fmt.Errorf("gh auth login required")
		}
		if stderr != "" {
			return fmt.Errorf("gh %s %s: %s", subcommand, owner, stderr)
		}
		return fmt.Errorf("gh %s %s: %w", subcommand, owner, err)
	}

	if pathErr, ok := err.(*exec.Error); ok {
		if pathErr.Err == exec.ErrNotFound {
			return fmt.Errorf("gh not found")
		}
		if strings.Contains(pathErr.Error(), "no such file") || strings.Contains(pathErr.Error(), "not found") {
			return fmt.Errorf("gh not found: %w", err)
		}
	}

	if stderr != "" {
		return fmt.Errorf("gh %s %s: %s: %w", subcommand, owner, stderr, err)
	}
	return fmt.Errorf("gh %s %s: %w", subcommand, owner, err)
}

func parseSearchRepos(data []byte) ([]Repo, error) {
	var wire []ghSearchRepoWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("decode gh search repos JSON: %w", err)
	}

	repos := make([]Repo, 0, len(wire))
	for _, item := range wire {
		owner := item.Owner.Login
		fullName := item.FullName
		if fullName == "" {
			fullName = buildFullName(owner, item.Name)
		}
		repos = append(repos, Repo{
			Owner:       owner,
			Name:        item.Name,
			FullName:    fullName,
			URL:         NormalizeRepoURL(owner, item.Name, item.URL),
			Description: item.Description,
			IsFork:      item.IsFork,
			IsArchived:  item.IsArchived,
		})
	}
	return repos, nil
}

func parseSearchCode(data []byte) ([]Repo, error) {
	var wire []ghSearchCodeWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("decode gh search code JSON: %w", err)
	}

	seen := make(map[string]struct{})
	repos := make([]Repo, 0, len(wire))
	for _, item := range wire {
		repo, err := repoFromGhWire(item.Repository)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[repo.FullName]; ok {
			continue
		}
		seen[repo.FullName] = struct{}{}
		repos = append(repos, repo)
	}
	return repos, nil
}

func repoFromGhWire(item ghRepoWire) (Repo, error) {
	owner := item.Owner.Login
	name := item.Name
	fullName := item.FullName
	if fullName == "" {
		fullName = item.NameWithOwner
	}
	if owner == "" && fullName != "" {
		parts := strings.SplitN(fullName, "/", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			owner = parts[0]
			if name == "" {
				name = parts[1]
			}
		}
	}
	if owner == "" {
		return Repo{}, fmt.Errorf("decode gh search code JSON: missing repository owner")
	}
	if fullName == "" {
		fullName = buildFullName(owner, name)
	}
	if name == "" {
		if i := strings.Index(fullName, "/"); i >= 0 {
			name = fullName[i+1:]
		}
	}
	return Repo{
		Owner:       owner,
		Name:        name,
		FullName:    fullName,
		URL:         NormalizeRepoURL(owner, name, item.URL),
		Description: item.Description,
		IsFork:      item.IsFork,
		IsArchived:  item.IsArchived,
	}, nil
}
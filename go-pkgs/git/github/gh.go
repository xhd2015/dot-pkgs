package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ghRepoOwner struct {
	Login string `json:"login"`
}

type ghRepoWire struct {
	Name        string      `json:"name"`
	URL         string      `json:"url"`
	Description string      `json:"description"`
	IsFork      bool        `json:"isFork"`
	IsArchived  bool        `json:"isArchived"`
	Owner       ghRepoOwner `json:"owner"`
}

func ghBinFromEnv() string {
	if v := os.Getenv("GH_BIN"); v != "" {
		return v
	}
	return "gh"
}

func resolveGhBin(override string) (string, error) {
	bin := override
	if bin == "" {
		bin = ghBinFromEnv()
	}
	if bin != "gh" {
		if _, err := os.Stat(bin); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("gh not found: %s", bin)
			}
			return "", fmt.Errorf("gh not found: %w", err)
		}
	}
	return bin, nil
}

func runGh(ctx context.Context, ghBin, owner string, limit int, includeArchived, includeForks bool) ([]byte, error) {
	args := []string{
		"repo", "list", owner,
		"--json", "name,url,description,isFork,isArchived,owner",
		"--limit", fmt.Sprintf("%d", limit),
	}
	if !includeArchived {
		args = append(args, "--no-archived")
	}
	if !includeForks {
		args = append(args, "--source")
	}

	cmd := exec.CommandContext(ctx, ghBin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.Output()
	if err != nil {
		return nil, wrapGhError(owner, err, stderr.String())
	}
	return stdout, nil
}

func wrapGhError(owner string, err error, stderr string) error {
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
			return fmt.Errorf("gh repo list %s: %s", owner, stderr)
		}
		return fmt.Errorf("gh repo list %s: %w", owner, err)
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
		return fmt.Errorf("gh repo list %s: %s: %w", owner, stderr, err)
	}
	return fmt.Errorf("gh repo list %s: %w", owner, err)
}

func parseGhRepos(data []byte) ([]Repo, error) {
	var wire []ghRepoWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("decode gh repo list JSON: %w", err)
	}

	repos := make([]Repo, 0, len(wire))
	for _, item := range wire {
		owner := item.Owner.Login
		repos = append(repos, Repo{
			Owner:       owner,
			Name:        item.Name,
			FullName:    buildFullName(owner, item.Name),
			URL:         NormalizeRepoURL(owner, item.Name, item.URL),
			Description: item.Description,
			IsFork:      item.IsFork,
			IsArchived:  item.IsArchived,
		})
	}
	return repos, nil
}

func buildFullName(owner, name string) string {
	return owner + "/" + name
}

// NormalizeRepoURL returns the canonical https://github.com/owner/name URL.
func NormalizeRepoURL(owner, name, raw string) string {
	_ = raw
	return fmt.Sprintf("https://github.com/%s/%s", owner, name)
}
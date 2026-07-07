package reposnapshot

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/checkout"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

type Node struct {
	Path      string
	Checkout  checkout.Meta
	Worktrees []Node
	Error     string
}

type RootErrorEntry struct {
	Path  string
	Error string
}

type Snapshot struct {
	Nodes      []Node
	RootErrors []RootErrorEntry
}

func Build(result scan_repo.Result, rel func(abs string) string) Snapshot {
	mains := make(map[string]scan_repo.Repo)
	worktreeRepos := make(map[string]scan_repo.Repo)

	for _, repo := range result.Repos {
		switch repo.RepoType {
		case scan_repo.RepoTypeMain:
			mains[repo.Path] = repo
		case scan_repo.RepoTypeWorktree:
			worktreeRepos[repo.Path] = repo
		}
	}

	mainPaths := make([]string, 0, len(mains))
	for path := range mains {
		mainPaths = append(mainPaths, path)
	}
	sort.Strings(mainPaths)

	ctx := context.Background()
	opts := checkout.Options{}

	nodes := make([]Node, 0, len(mainPaths)+len(result.RootErrors))
	for _, mainPath := range mainPaths {
		mainRepo := mains[mainPath]
		node := nodeFromCheckout(mainPath, checkout.Enrich(ctx, mainPath, opts), rel)
		node.Error = mergeErrors(node.Checkout.Error, mainRepo.Error)
		node.Checkout.Error = ""

		wtPaths := collectWorktreePaths(mainRepo, worktreeRepos)
		for _, wtPath := range wtPaths {
			wtNode := nodeFromCheckout(wtPath, checkout.Enrich(ctx, wtPath, opts), rel)
			if wtRepo, ok := worktreeRepos[wtPath]; ok {
				wtNode.Error = mergeErrors(wtNode.Checkout.Error, wtRepo.Error)
			} else {
				wtNode.Error = wtNode.Checkout.Error
			}
			wtNode.Checkout.Error = ""
			node.Worktrees = append(node.Worktrees, wtNode)
		}
		sort.Slice(node.Worktrees, func(i, j int) bool {
			return node.Worktrees[i].Path < node.Worktrees[j].Path
		})
		nodes = append(nodes, node)
	}

	rootErrors := make([]RootErrorEntry, 0, len(result.RootErrors))
	for _, re := range result.RootErrors {
		entry := RootErrorEntry{
			Path:  rel(re.Root),
			Error: re.Error,
		}
		rootErrors = append(rootErrors, entry)
		nodes = append(nodes, Node{
			Path:  entry.Path,
			Error: "scan failed: " + re.Error,
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Path < nodes[j].Path
	})

	return Snapshot{
		Nodes:      nodes,
		RootErrors: rootErrors,
	}
}

func nodeFromCheckout(absPath string, meta checkout.Meta, rel func(abs string) string) Node {
	return Node{
		Path:     rel(absPath),
		Checkout: meta,
	}
}

func collectWorktreePaths(mainRepo scan_repo.Repo, worktreeRepos map[string]scan_repo.Repo) []string {
	seen := make(map[string]struct{})
	var paths []string

	mainGitDir := filepath.Clean(mainRepo.GitDir)
	for _, wt := range mainRepo.Worktrees {
		if wt.IsMain {
			continue
		}
		path := filepath.Clean(wt.Path)
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	for path, repo := range worktreeRepos {
		if filepath.Clean(repo.GitDir) != mainGitDir {
			continue
		}
		path = filepath.Clean(path)
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func mergeErrors(parts ...string) string {
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, "; ")
}
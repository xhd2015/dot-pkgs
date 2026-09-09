package scan_repo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseOriginIdentity(t *testing.T) {
	cases := []struct {
		raw       string
		wantHost  string
		wantOwner string
		wantRepo  string
	}{
		{"https://github.com/WiseWiseWiser/mobile-coding-connector.git", "github.com", "WiseWiseWiser", "mobile-coding-connector"},
		{"ssh://git@github.com/WiseWiseWiser/mobile-coding-connector.git", "github.com", "WiseWiseWiser", "mobile-coding-connector"},
		{"git@github.com:WiseWiseWiser/mobile-coding-connector.git", "github.com", "WiseWiseWiser", "mobile-coding-connector"},
		{"WiseWiseWiser/mobile-coding-connector", "github.com", "WiseWiseWiser", "mobile-coding-connector"},
	}
	for _, tc := range cases {
		got, err := ParseOriginIdentity(tc.raw)
		if err != nil {
			t.Fatalf("ParseOriginIdentity(%q): %v", tc.raw, err)
		}
		if got.Host != tc.wantHost || got.Owner != tc.wantOwner || got.Repo != tc.wantRepo {
			t.Fatalf("ParseOriginIdentity(%q) = %+v, want host=%s owner=%s repo=%s",
				tc.raw, got, tc.wantHost, tc.wantOwner, tc.wantRepo)
		}
	}
}

func TestOriginRankPreference(t *testing.T) {
	cwd := "/Users/me/proj"
	under := rankedCandidate{Repo: Repo{Path: "/Users/me/proj/external/ai-critic", RepoType: RepoTypeWorktree}, reason: ReasonUnderCWD, priority: 0}
	parent := rankedCandidate{Repo: Repo{Path: "/Users/me", RepoType: RepoTypeMain}, reason: ReasonContainsCWD, priority: 1}
	main := rankedCandidate{Repo: Repo{Path: "/opt/ai-critic", RepoType: RepoTypeMain}, reason: ReasonMain, priority: 2}

	p, r := originRank(under.Path, under.RepoType, cwd)
	if p != 0 || r != ReasonUnderCWD {
		t.Fatalf("under-cwd rank = %d %s", p, r)
	}
	p, r = originRank(parent.Path, parent.RepoType, cwd)
	if p != 1 || r != ReasonContainsCWD {
		t.Fatalf("contains-cwd rank = %d %s", p, r)
	}
	p, r = originRank(main.Path, main.RepoType, cwd)
	if p != 2 || r != ReasonMain {
		t.Fatalf("main rank = %d %s", p, r)
	}

	chosen := pickOriginCandidate([]Repo{main.Repo, parent.Repo, under.Repo}, cwd)
	if chosen.Path != under.Path || chosen.reason != ReasonUnderCWD {
		t.Fatalf("pick = %s (%s), want under-cwd", chosen.Path, chosen.reason)
	}

	chosen = pickOriginCandidate([]Repo{main.Repo, parent.Repo}, cwd)
	if chosen.Path != parent.Path || chosen.reason != ReasonContainsCWD {
		t.Fatalf("pick = %s (%s), want contains-cwd", chosen.Path, chosen.reason)
	}

	chosen = pickOriginCandidate([]Repo{main.Repo}, cwd)
	if chosen.Path != main.Path || chosen.reason != ReasonMain {
		t.Fatalf("pick = %s (%s), want main", chosen.Path, chosen.reason)
	}
}

func TestResolveByOriginPrefersUnderCWD(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	cacheRoot := t.TempDir()
	cwd := filepath.Join(root, "proj")
	mainDir := filepath.Join(root, "main-checkout")
	underDir := filepath.Join(cwd, "external", "ai-critic")
	otherMain := filepath.Join(root, "other-main")

	for _, dir := range []string{cwd, mainDir, underDir, otherMain} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	initWithOrigin := func(dir, url string) {
		t.Helper()
		run := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
			}
		}
		run("init")
		run("config", "user.email", "test@example.com")
		run("config", "user.name", "Test")
		run("remote", "add", "origin", url)
		if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0644); err != nil {
			t.Fatal(err)
		}
		run("-c", "core.hooksPath=/dev/null", "add", "README")
		run("-c", "core.hooksPath=/dev/null", "commit", "-m", "init")
	}

	origin := "git@github.com:WiseWiseWiser/mobile-coding-connector.git"
	initWithOrigin(mainDir, origin)
	initWithOrigin(underDir, origin)
	initWithOrigin(otherMain, "git@github.com:other/other.git")

	got, err := ResolveByOrigin(context.Background(), Options{
		Roots:     []string{root},
		CacheRoot: cacheRoot,
		NoCache:   false,
	}, origin, cwd)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(underDir)
	if got.Path != want {
		t.Fatalf("Path = %q, want %q", got.Path, want)
	}
	if got.Reason != ReasonUnderCWD {
		t.Fatalf("Reason = %q, want %s", got.Reason, ReasonUnderCWD)
	}
}

func TestResolveByOriginUsesCacheOnSecondScan(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	cacheRoot := t.TempDir()
	repoDir := filepath.Join(root, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	for _, args := range [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"remote", "add", "origin", "https://github.com/WiseWiseWiser/mobile-coding-connector.git"},
	} {
		c := exec.Command("git", args...)
		c.Dir = repoDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(repoDir, "README"), []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"-c", "core.hooksPath=/dev/null", "add", "README"},
		{"-c", "core.hooksPath=/dev/null", "commit", "-m", "init"},
	} {
		c := exec.Command("git", args...)
		c.Dir = repoDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	opts := Options{Roots: []string{root}, CacheRoot: cacheRoot}
	if _, err := ResolveByOrigin(context.Background(), opts, "WiseWiseWiser/mobile-coding-connector", ""); err != nil {
		t.Fatalf("first resolve: %v", err)
	}
	// Second call should succeed via warm cache path (same CacheRoot).
	got, err := ResolveByOrigin(context.Background(), opts, "https://github.com/WiseWiseWiser/mobile-coding-connector.git", "")
	if err != nil {
		t.Fatalf("second resolve: %v", err)
	}
	if got.Path != filepath.Clean(repoDir) {
		t.Fatalf("Path = %q, want %q", got.Path, repoDir)
	}
	if got.Reason != ReasonMain {
		t.Fatalf("Reason = %q, want %s", got.Reason, ReasonMain)
	}
}

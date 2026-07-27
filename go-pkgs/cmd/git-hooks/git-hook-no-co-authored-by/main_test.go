package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	var out bytes.Buffer
	err := runWithIO([]string{"-h"}, strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Usage: git-hook-no-co-authored-by",
		"--origin-domain",
		"pre-push",
		"Co-authored-by",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q; got:\n%s", want, got)
		}
	}
}

func TestEmptyStdinPasses(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	doCommit(t, repo, "feat: initial")

	var out bytes.Buffer
	if err := runWithIO(nil, strings.NewReader(""), &out); err != nil {
		t.Fatalf("empty stdin should pass: %v\n%s", err, out.String())
	}
}

func TestCleanPushPasses(t *testing.T) {
	remote, clone := initRemoteAndClone(t)
	_ = remote
	t.Chdir(clone)

	base := revParse(t, clone, "HEAD")
	doCommit(t, clone, "feat: clean")
	local := revParse(t, clone, "HEAD")

	stdin := pushLine("refs/heads/master", local, "refs/heads/master", base)
	var out bytes.Buffer
	if err := runWithIO(nil, strings.NewReader(stdin), &out); err != nil {
		t.Fatalf("clean push should pass: %v\n%s", err, out.String())
	}
}

func TestOlderUnpushedWithTrailerRejected(t *testing.T) {
	// remote tip clean; two local commits: first has Co-authored-by, second (HEAD) is clean.
	remote, clone := initRemoteAndClone(t)
	_ = remote
	t.Chdir(clone)

	base := revParse(t, clone, "HEAD")
	doCommit(t, clone, "feat: with bot\n\nBody.\n\nCo-authored-by: CommandCodeBot <noreply@commandcode.ai>")
	badSHA := revParse(t, clone, "HEAD")
	doCommit(t, clone, "feat: clean tip")
	local := revParse(t, clone, "HEAD")

	stdin := pushLine("refs/heads/master", local, "refs/heads/master", base)
	var out bytes.Buffer
	err := runWithIO(nil, strings.NewReader(stdin), &out)
	if !errors.Is(err, errCoAuthoredByFound) {
		t.Fatalf("expected reject, got %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Error: cannot push commits with Co-authored-by trailer:") {
		t.Fatalf("missing error header:\n%s", out.String())
	}
	if !strings.Contains(out.String(), shortSHA(badSHA)) {
		t.Fatalf("should list older bad SHA %s:\n%s", shortSHA(badSHA), out.String())
	}
	if !strings.Contains(out.String(), "feat: with bot") {
		t.Fatalf("should list bad subject:\n%s", out.String())
	}
}

func TestAlreadyPushedTrailerExcluded(t *testing.T) {
	// Commit with trailer is already on remote; new tip is clean → pass.
	_, clone := initRemoteAndClone(t)
	t.Chdir(clone)

	doCommit(t, clone, "feat: already remote\n\nCo-authored-by: Bot <bot@example.com>")
	mustRun(t, clone, "git", "push", "origin", "HEAD")
	pushed := revParse(t, clone, "HEAD")

	doCommit(t, clone, "feat: clean after trailer")
	local := revParse(t, clone, "HEAD")

	stdin := pushLine("refs/heads/master", local, "refs/heads/master", pushed)
	var out bytes.Buffer
	if err := runWithIO(nil, strings.NewReader(stdin), &out); err != nil {
		t.Fatalf("already-pushed trailer must be excluded: %v\n%s", err, out.String())
	}
}

func TestNewBranchLocalOnlyTrailerRejected(t *testing.T) {
	_, clone := initRemoteAndClone(t)
	t.Chdir(clone)

	// Keep origin/master without the trailer; create a new local branch with trailer.
	mustRun(t, clone, "git", "checkout", "-b", "feature")
	doCommit(t, clone, "feat: feature\n\nCo-authored-by: Bot <bot@example.com>")
	local := revParse(t, clone, "HEAD")
	z40 := strings.Repeat("0", 40)

	stdin := pushLine("refs/heads/feature", local, "refs/heads/feature", z40)
	var out bytes.Buffer
	err := runWithIO(nil, strings.NewReader(stdin), &out)
	if !errors.Is(err, errCoAuthoredByFound) {
		t.Fatalf("new branch local trailer should reject: %v\n%s", err, out.String())
	}
}

func TestNewBranchHistoryAlreadyOnRemotesPasses(t *testing.T) {
	_, clone := initRemoteAndClone(t)
	t.Chdir(clone)

	// Trailer commit pushed on master; new branch from that tip with no new commits
	// (or only clean commits) should not re-flag already-pushed history.
	doCommit(t, clone, "feat: pushed trailer\n\nCo-authored-by: Bot <bot@example.com>")
	mustRun(t, clone, "git", "push", "origin", "HEAD")
	mustRun(t, clone, "git", "checkout", "-b", "feature2")
	// No new commits: push of same tip as new branch name — rev-list local --not --remotes is empty.
	local := revParse(t, clone, "HEAD")
	z40 := strings.Repeat("0", 40)

	stdin := pushLine("refs/heads/feature2", local, "refs/heads/feature2", z40)
	var out bytes.Buffer
	if err := runWithIO(nil, strings.NewReader(stdin), &out); err != nil {
		t.Fatalf("new branch with only already-remote history should pass: %v\n%s", err, out.String())
	}
}

func TestDeleteBranchPasses(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	doCommit(t, repo, "feat: x")
	remote := revParse(t, repo, "HEAD")
	z40 := strings.Repeat("0", 40)

	stdin := pushLine("refs/heads/master", z40, "refs/heads/master", remote)
	var out bytes.Buffer
	if err := runWithIO(nil, strings.NewReader(stdin), &out); err != nil {
		t.Fatalf("delete should pass: %v\n%s", err, out.String())
	}
}

func TestMixedCaseTrailerRejected(t *testing.T) {
	_, clone := initRemoteAndClone(t)
	t.Chdir(clone)
	base := revParse(t, clone, "HEAD")
	doCommit(t, clone, "feat: mixed\n\nCo-Authored-By: Bot <bot@example.com>")
	local := revParse(t, clone, "HEAD")

	stdin := pushLine("refs/heads/master", local, "refs/heads/master", base)
	var out bytes.Buffer
	err := runWithIO(nil, strings.NewReader(stdin), &out)
	if !errors.Is(err, errCoAuthoredByFound) {
		t.Fatalf("mixed case should reject: %v\n%s", err, out.String())
	}
}

func TestProseWithoutColonPasses(t *testing.T) {
	_, clone := initRemoteAndClone(t)
	t.Chdir(clone)
	base := revParse(t, clone, "HEAD")
	doCommit(t, clone, "docs: mention co-authored by Alice in prose")
	local := revParse(t, clone, "HEAD")

	stdin := pushLine("refs/heads/master", local, "refs/heads/master", base)
	var out bytes.Buffer
	if err := runWithIO(nil, strings.NewReader(stdin), &out); err != nil {
		t.Fatalf("prose without colon should pass: %v\n%s", err, out.String())
	}
}

func TestDomainGateSkips(t *testing.T) {
	_, clone := initRemoteAndClone(t)
	t.Chdir(clone)
	// Point origin at a host we can filter.
	mustRun(t, clone, "git", "remote", "set-url", "origin", "git@git.xxx.com:team/repo.git")
	base := revParse(t, clone, "HEAD")
	doCommit(t, clone, "feat: bad\n\nCo-authored-by: Bot <bot@example.com>")
	local := revParse(t, clone, "HEAD")

	stdin := pushLine("refs/heads/master", local, "refs/heads/master", base)
	var out bytes.Buffer
	if err := runWithIO([]string{"--origin-domain", "other.example.com"}, strings.NewReader(stdin), &out); err != nil {
		t.Fatalf("mismatched domain should skip: %v\n%s", err, out.String())
	}

	err := runWithIO([]string{"--origin-domain", "git.xxx.com"}, strings.NewReader(stdin), &out)
	if !errors.Is(err, errCoAuthoredByFound) {
		t.Fatalf("matching domain should reject: %v\n%s", err, out.String())
	}
}

func TestMessageHasCoAuthoredBy(t *testing.T) {
	if !messageHasCoAuthoredBy("x\n\nCo-authored-by: A <a@b.c>\n") {
		t.Fatal("expected match")
	}
	if messageHasCoAuthoredBy("co-authored by alice") {
		t.Fatal("prose without colon must not match")
	}
	if !messageHasCoAuthoredBy("CO-AUTHORED-BY: x") {
		t.Fatal("case insensitive")
	}
}

func TestInvalidStdinLine(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	var out bytes.Buffer
	err := runWithIO(nil, strings.NewReader("only three fields here\n"), &out)
	if err == nil {
		t.Fatal("expected invalid stdin error")
	}
}

// --- helpers ---

func pushLine(localRef, localSHA, remoteRef, remoteSHA string) string {
	return fmt.Sprintf("%s %s %s %s\n", localRef, localSHA, remoteRef, remoteSHA)
}

func initRemoteAndClone(t *testing.T) (remote, clone string) {
	t.Helper()
	remote = t.TempDir()
	mustRun(t, remote, "git", "init", "--bare")

	// Seed bare remote via a temp seed repo.
	seed := t.TempDir()
	mustRun(t, seed, "git", "init")
	mustRun(t, seed, "git", "config", "user.email", "test@example.com")
	mustRun(t, seed, "git", "config", "user.name", "Test User")
	// Default branch master for stable ref names in tests.
	mustRun(t, seed, "git", "checkout", "-b", "master")
	writeFile(t, filepath.Join(seed, "README"), "seed\n")
	mustRun(t, seed, "git", "add", "README")
	mustRun(t, seed, "git", "commit", "-m", "initial")
	mustRun(t, seed, "git", "remote", "add", "origin", remote)
	mustRun(t, seed, "git", "push", "-u", "origin", "master")

	clone = t.TempDir()
	mustRun(t, "", "git", "clone", remote, clone)
	mustRun(t, clone, "git", "config", "user.email", "test@example.com")
	mustRun(t, clone, "git", "config", "user.name", "Test User")
	return remote, clone
}

var commitSeq int

func doCommit(t *testing.T, repo string, msg string) {
	t.Helper()
	// Unique path so successive commits always succeed (works before first commit too).
	commitSeq++
	name := fmt.Sprintf("f-%d.txt", commitSeq)
	writeFile(t, filepath.Join(repo, name), msg+"\n")
	mustRun(t, repo, "git", "add", name)
	mustRun(t, repo, "git", "commit", "-m", msg)
}

func revParse(t *testing.T, repo, rev string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repo, "rev-parse", rev).CombinedOutput()
	if err != nil {
		// optional missing ref
		return ""
	}
	return strings.TrimSpace(string(out))
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	mustRun(t, repo, "git", "init")
	mustRun(t, repo, "git", "config", "user.email", "test@example.com")
	mustRun(t, repo, "git", "config", "user.name", "Test User")
	mustRun(t, repo, "git", "checkout", "-b", "master")
	return repo
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}


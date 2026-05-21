package main

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestRunAcceptsMatchingAuthor(t *testing.T) {
	repo := initGitRepo(t, "Xxx User", "xxx@xx.xx")
	t.Chdir(repo)

	var out bytes.Buffer
	if err := runWithOutput([]string{"--name", "xxx user", "--email", "contains:@xx.xx"}, &out); err != nil {
		t.Fatalf("expected matching author to pass, got %v\n%s", err, out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output on success, got:\n%s", out.String())
	}
}

func TestRunFailsMismatchedAuthor(t *testing.T) {
	repo := initGitRepo(t, "Xxx User", "xxx@xx.xx")
	t.Chdir(repo)

	var out bytes.Buffer
	err := runWithOutput([]string{"--name", "other user", "--email", "ends-with:@example.com"}, &out)
	if !errors.Is(err, errAuthorCheckFailed) {
		t.Fatalf("expected author-check failure, got %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `name "Xxx User" does not satisfy equals "other user"`) {
		t.Fatalf("missing name failure:\n%s", got)
	}
	if !strings.Contains(got, `email "xxx@xx.xx" does not satisfy ends with "@example.com"`) {
		t.Fatalf("missing email failure:\n%s", got)
	}
}

func TestRunSupportsNegatedAuthorConditions(t *testing.T) {
	repo := initGitRepo(t, "Xxx User", "xxx@xx.xx")
	t.Chdir(repo)

	var out bytes.Buffer
	if err := runWithOutput([]string{"--not-name", "contains:bot", "--email", "!ends-with:@gmail.com"}, &out); err != nil {
		t.Fatalf("expected negated checks to pass, got %v\n%s", err, out.String())
	}

	err := runWithOutput([]string{"--not-email", "contains:xx.xx"}, &out)
	if !errors.Is(err, errAuthorCheckFailed) {
		t.Fatalf("expected negated email check to fail, got %v", err)
	}
}

func TestOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t, "Xxx User", "xxx@xx.xx")
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")

	var out bytes.Buffer
	if err := runWithOutput([]string{"--origin-domain", "other.example.com", "--name", "wrong"}, &out); err != nil {
		t.Fatalf("expected mismatched origin domain to skip, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain does not match, got:\n%s", out.String())
	}

	err := runWithOutput([]string{"--origin-domain", "git.xxx.com", "--name", "wrong"}, &out)
	if !errors.Is(err, errAuthorCheckFailed) {
		t.Fatalf("expected matching origin domain to scan, got %v", err)
	}
}

func TestExcludeOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t, "Xxx User", "xxx@xx.xx")
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "https://git.xxx.com/team/repo.git")

	var out bytes.Buffer
	if err := runWithOutput([]string{"--exclude-origin-domain", "git.xxx.com", "--name", "wrong"}, &out); err != nil {
		t.Fatalf("expected excluded origin domain to skip, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain is excluded, got:\n%s", out.String())
	}

	err := runWithOutput([]string{"--exclude-origin-domain", "other.example.com", "--name", "wrong"}, &out)
	if !errors.Is(err, errAuthorCheckFailed) {
		t.Fatalf("expected non-excluded origin domain to scan, got %v", err)
	}
}

func TestConditionOperations(t *testing.T) {
	tests := []struct {
		raw    string
		actual string
		want   bool
	}{
		{raw: "xxx user", actual: "Xxx User", want: true},
		{raw: "contains:USER", actual: "Xxx User", want: true},
		{raw: "starts-with:xxx", actual: "Xxx User", want: true},
		{raw: "ends-with:USER", actual: "Xxx User", want: true},
		{raw: "!contains:bot", actual: "Xxx User", want: true},
		{raw: "not:ends-with:gmail.com", actual: "xxx@xx.xx", want: true},
	}
	for _, tt := range tests {
		cond, err := parseCondition(tt.raw)
		if err != nil {
			t.Fatalf("parseCondition(%q): %v", tt.raw, err)
		}
		if got := cond.matches(tt.actual); got != tt.want {
			t.Fatalf("condition %q on %q = %v, want %v", tt.raw, tt.actual, got, tt.want)
		}
	}
}

func initGitRepo(t *testing.T, name string, email string) string {
	t.Helper()
	repo := t.TempDir()
	mustRun(t, repo, "git", "init")
	mustRun(t, repo, "git", "config", "user.name", name)
	mustRun(t, repo, "git", "config", "user.email", email)
	return repo
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}

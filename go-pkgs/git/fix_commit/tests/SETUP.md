# Scenario

**Feature**: `fix_commit.RunCLI` rewrites one commit and retargets exact-tip refs

```
# in-process CLI; repo dir via -C; writers injected
Caller -> RunCLI(args, stdout, stderr) -> Git -C <dir> -> new commit + moved refs

# fatals return Error: …; harness appends them to stderr
RunCLI err -> Run appends err.Error() -> resp.Stderr
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/fix_commit` is the SUT and
  does **not** exist yet (classic TDD, compile-RED).
- Real `git` on PATH; helpers skip when missing. Never hit the network.
- Each leaf uses its own `t.TempDir()` repo. No `os.Chdir` / `t.Chdir` /
  `os.Setenv` / `t.Setenv`. Git identity is injected via `cmd.Env`.
- Process cwd is undetermined; every git-talking leaf passes `-C <abs dir>`.

## Steps

1. Root helpers build isolated git fixtures (`master`, known author/committer
   dates, optional bare remotes).
2. Leaf `Setup` sets `req.Args` (and amends topology / message as needed).
3. Root `Run` calls `fix_commit.RunCLI(req.Args, stdout, stderr)`.

## Context

- Default branch is `master`.
- Fixture author `Alice <alice@example.com>` at `2020-01-02T03:04:05+0000`.
- Fixture committer `Committer <committer@example.com>` at
  `2020-06-07T08:09:10+0000`. Distinct so asserts can prove which side changed.
- Rewrite fixtures have an `init` parent plus a target commit so `HEAD^` exists.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

const (
	fixtureAuthorName     = "Alice"
	fixtureAuthorEmail    = "alice@example.com"
	fixtureAuthorDate     = "2020-01-02T03:04:05+0000"
	fixtureCommitterName  = "Committer"
	fixtureCommitterEmail = "committer@example.com"
	fixtureCommitterDate  = "2020-06-07T08:09:10+0000"
	fixtureUserName       = "Test User"
	fixtureUserEmail      = "test@example.com"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	skipIfNoGit(t)
	return nil
}

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
}

func gitEnv() []string {
	return append(os.Environ(),
		"GIT_AUTHOR_NAME="+fixtureAuthorName,
		"GIT_AUTHOR_EMAIL="+fixtureAuthorEmail,
		"GIT_AUTHOR_DATE="+fixtureAuthorDate,
		"GIT_COMMITTER_NAME="+fixtureCommitterName,
		"GIT_COMMITTER_EMAIL="+fixtureCommitterEmail,
		"GIT_COMMITTER_DATE="+fixtureCommitterDate,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_TERMINAL_PROMPT=0",
		"GCM_INTERACTIVE=never",
	)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	skipIfNoGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "master")
	runGit(t, dir, "config", "user.name", fixtureUserName)
	runGit(t, dir, "config", "user.email", fixtureUserEmail)
	runGit(t, dir, "config", "commit.gpgsign", "false")
	runGit(t, dir, "config", "tag.gpgsign", "false")
	runGit(t, dir, "config", "init.defaultBranch", "master")
	return dir
}

func commitFile(t *testing.T, dir, relPath, content, msg string) string {
	t.Helper()
	writeFile(t, filepath.Join(dir, relPath), content)
	runGit(t, dir, "add", relPath)
	runGit(t, dir, "commit", "-m", msg)
	return runGitOutput(t, dir, "rev-parse", "HEAD")
}

func setHEADMessage(t *testing.T, dir, msg string) {
	t.Helper()
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	path := filepath.Join(dir, ".git", "FIX_COMMIT_MSG")
	writeFile(t, path, msg)
	runGit(t, dir, "commit", "--amend", "-F", path)
}

func fillCommitMeta(t *testing.T, req *Request) {
	t.Helper()
	req.OldSHA = runGitOutput(t, req.Dir, "rev-parse", "HEAD")
	req.TreeSHA = runGitOutput(t, req.Dir, "rev-parse", "HEAD^{tree}")
	req.ParentSHA = runGitOutput(t, req.Dir, "rev-parse", "HEAD^")
	req.AuthorName = fixtureAuthorName
	req.AuthorEmail = fixtureAuthorEmail
	req.CommitterName = fixtureCommitterName
	req.CommitterEmail = fixtureCommitterEmail
	req.Message = runGitOutput(t, req.Dir, "log", "-1", "--format=%B")
	flags := []string{}
	if len(req.Args) >= 3 && req.Args[0] == "-C" {
		flags = append(flags, req.Args[3:]...)
	}
	req.Args = append([]string{"-C", req.Dir, req.OldSHA}, flags...)
}

func addBareRemote(t *testing.T, repo, name string) string {
	t.Helper()
	parent := t.TempDir()
	bare := filepath.Join(parent, name+".git")
	runGit(t, parent, "init", "--bare", name+".git")
	runGit(t, repo, "remote", "add", name, bare)
	return bare
}

func peeled(t *testing.T, dir, rev string) string {
	t.Helper()
	return runGitOutput(t, dir, "rev-parse", rev+"^{commit}")
}

func commitField(t *testing.T, dir, sha, format string) string {
	t.Helper()
	return runGitOutput(t, dir, "show", "-s", "--format="+format, sha)
}

func parseRewrote(t *testing.T, stdout string) (oldSHA, newSHA string) {
	t.Helper()
	line := stdout
	if i := strings.IndexByte(stdout, '\n'); i >= 0 {
		line = stdout[:i]
	}
	const prefix = "rewrote "
	if !strings.HasPrefix(line, prefix) {
		t.Fatalf("stdout first line %q, want rewrote old -> new", line)
	}
	parts := strings.Split(strings.TrimPrefix(line, prefix), " -> ")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		t.Fatalf("stdout first line %q, want rewrote old -> new", line)
	}
	return parts[0], parts[1]
}

func outputTemplate(want string) string {
	if !strings.HasSuffix(want, "\n") {
		want += "\n"
	}
	var b strings.Builder
	b.WriteString("---\nversion: 3\n---\n")
	for _, line := range strings.Split(strings.TrimSuffix(want, "\n"), "\n") {
		b.WriteString(regexp.QuoteMeta(line))
		b.WriteByte('\n')
	}
	return b.String()
}

func assertOutput(t *testing.T, actual, want string) {
	t.Helper()
	assert.Output(t, actual, outputTemplate(want))
}

func requireHarnessOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func requireExit(t *testing.T, resp *Response, code int) {
	t.Helper()
	if resp.ExitCode != code {
		t.Fatalf("ExitCode=%d want %d\nstdout=%q\nstderr=%q", resp.ExitCode, code, resp.Stdout, resp.Stderr)
	}
}

type successReport struct {
	Old, New         string
	AuthorName       string
	AuthorEmail      string
	MessageFirstLine string
	Stripped         bool
	Branches         []string
	Tags             []string
	Pushed           []string
}

func formatSuccess(r successReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "rewrote %s -> %s\n", r.Old, r.New)
	fmt.Fprintf(&b, "  author:  %s <%s>\n", r.AuthorName, r.AuthorEmail)
	fmt.Fprintf(&b, "  message: %s\n", r.MessageFirstLine)
	if r.Stripped {
		b.WriteString("  stripped Co-authored-by\n")
	}
	if len(r.Branches) > 0 {
		b.WriteString("  branches:\n")
		for _, br := range r.Branches {
			fmt.Fprintf(&b, "    %s\n", br)
		}
	}
	if len(r.Tags) > 0 {
		b.WriteString("  tags:\n")
		for _, line := range r.Tags {
			fmt.Fprintf(&b, "    %s\n", line)
		}
	}
	if len(r.Pushed) > 0 {
		b.WriteString("  pushed:\n")
		for _, br := range r.Pushed {
			fmt.Fprintf(&b, "    %s\n", br)
		}
	}
	return b.String()
}

func assertUnchangedSHA(t *testing.T, req *Request) {
	t.Helper()
	got := runGitOutput(t, req.Dir, "rev-parse", "HEAD")
	if got != req.OldSHA {
		t.Fatalf("HEAD moved: got %s want %s", got, req.OldSHA)
	}
}

func assertNewCommit(t *testing.T, req *Request, newSHA, wantName, wantEmail, wantSubject string) {
	t.Helper()
	if newSHA == req.OldSHA {
		t.Fatal("new SHA equals old SHA")
	}
	if got := commitField(t, req.Dir, newSHA, "%T"); got != req.TreeSHA {
		t.Fatalf("tree %s want %s", got, req.TreeSHA)
	}
	if got := commitField(t, req.Dir, newSHA, "%P"); got != req.ParentSHA {
		t.Fatalf("parents %s want %s", got, req.ParentSHA)
	}
	if got := commitField(t, req.Dir, newSHA, "%an"); got != wantName {
		t.Fatalf("author name %q want %q", got, wantName)
	}
	if got := commitField(t, req.Dir, newSHA, "%ae"); got != wantEmail {
		t.Fatalf("author email %q want %q", got, wantEmail)
	}
	oldAT := commitField(t, req.Dir, req.OldSHA, "%at")
	if got := commitField(t, req.Dir, newSHA, "%at"); got != oldAT {
		t.Fatalf("author date unix %s want %s", got, oldAT)
	}
	if got := commitField(t, req.Dir, newSHA, "%cn"); got != req.CommitterName {
		t.Fatalf("committer name %q want %q", got, req.CommitterName)
	}
	if got := commitField(t, req.Dir, newSHA, "%ce"); got != req.CommitterEmail {
		t.Fatalf("committer email %q want %q", got, req.CommitterEmail)
	}
	oldCT := commitField(t, req.Dir, req.OldSHA, "%ct")
	if got := commitField(t, req.Dir, newSHA, "%ct"); got != oldCT {
		t.Fatalf("committer date unix %s want %s", got, oldCT)
	}
	if got := commitField(t, req.Dir, newSHA, "%s"); got != wantSubject {
		t.Fatalf("subject %q want %q", got, wantSubject)
	}
}

func assertFullMessage(t *testing.T, dir, sha, want string) {
	t.Helper()
	got := runGitOutput(t, dir, "log", "-1", "--format=%B", sha)
	want = strings.TrimSpace(want)
	if got != want {
		t.Fatalf("full message %q want %q", got, want)
	}
}
```

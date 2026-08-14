package fix_commit

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
	lessflags "github.com/xhd2015/less-flags"
)

const usage = `Usage: kool git fix-commit <sha> [OPTIONS]

  -m, --message <msg>     replace the full commit message
  --name <name>           replace the author name
  --email <email>         replace the author email
  --strip-co-author       remove Co-authored-by lines from the message;
                          errors if none are present
  --remote <name>         remote for tag delete/push (default: origin)
  --push                  also force-with-lease push updated branches
                          whose upstream still points at the old sha
  --dry-run               print the plan; do not rewrite refs or remotes
  -C, --dir <dir>         repository directory (default: current directory)
  -h, --help              show this help
`

func RunCLI(args []string, stdout, stderr io.Writer) error {
	var message string
	var name string
	var email string
	var strip bool
	var remote string
	var doPush bool
	var dryRun bool
	var dir string

	remain, err := lessflags.String("-m,--message", &message).
		String("--name", &name).
		String("--email", &email).
		Bool("--strip-co-author", &strip).
		String("--remote", &remote).
		Bool("--push", &doPush).
		Bool("--dry-run", &dryRun).
		String("-C,--dir", &dir).
		HelpFunc("-h,--help", func() {
			fmt.Fprint(stdout, usage)
		}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			return nil
		}
		msg := err.Error()
		if strings.HasPrefix(msg, "unrecognized flag: ") {
			return fmt.Errorf("Error: unknown flag: %s", strings.TrimPrefix(msg, "unrecognized flag: "))
		}
		return fmt.Errorf("Error: %s", msg)
	}

	if len(remain) == 0 {
		return fmt.Errorf("Error: fix-commit requires <sha>")
	}
	shaToken := remain[0]
	hasMessage := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-m" || a == "--message" || strings.HasPrefix(a, "-m=") || strings.HasPrefix(a, "--message=") {
			hasMessage = true
			break
		}
	}
	if !hasMessage && name == "" && email == "" && !strip {
		return fmt.Errorf("Error: at least one of -m, --email, --name, --strip-co-author is required")
	}
	if strip && hasMessage {
		return fmt.Errorf("Error: --strip-co-author and -m cannot be used together")
	}

	if dir == "" {
		dir = "."
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("Error: not a git repository: %s", dir)
	}
	ctx := context.Background()
	if _, err := cmd.RunCombined(ctx, absDir, "rev-parse", "--git-dir"); err != nil {
		return fmt.Errorf("Error: not a git repository: %s", absDir)
	}

	oldSHA, err := cmd.Run(ctx, absDir, "rev-parse", "--verify", shaToken+"^{commit}")
	if err != nil {
		return fmt.Errorf("Error: unknown revision: %s", shaToken)
	}

	meta, err := loadCommit(ctx, absDir, oldSHA)
	if err != nil {
		return err
	}

	newName := meta.authorName
	newEmail := meta.authorEmail
	if name != "" {
		newName = name
	}
	if email != "" {
		newEmail = email
	}

	newMsg := meta.message
	stripped := false
	if hasMessage {
		newMsg = message
		if !strings.HasSuffix(newMsg, "\n") {
			newMsg += "\n"
		}
	} else if strip {
		var err error
		newMsg, err = stripCoAuthor(meta.message)
		if err != nil {
			return err
		}
		stripped = true
	}

	if strings.TrimSpace(newMsg) == strings.TrimSpace(meta.message) && newName == meta.authorName && newEmail == meta.authorEmail {
		return fmt.Errorf("Error: nothing to change")
	}

	branches, err := listExactTipBranches(ctx, absDir, oldSHA)
	if err != nil {
		return err
	}
	detached, err := isDetachedAt(ctx, absDir, oldSHA)
	if err != nil {
		return err
	}
	hasDescendants, err := commitHasDescendants(ctx, absDir, oldSHA)
	if err != nil {
		return err
	}
	if hasDescendants {
		fmt.Fprintf(stderr, "warning: commit has descendants; those commits still parent %s\n", oldSHA)
	}

	if remote == "" {
		remote = "origin"
	}
	tags, err := listTagsAt(ctx, absDir, oldSHA)
	if err != nil {
		return err
	}
	remoteExists := false
	if len(tags) > 0 {
		remoteExists, err = remotePresent(ctx, absDir, remote)
		if err != nil {
			return err
		}
		if !remoteExists {
			fmt.Fprintf(stderr, "warning: remote %s not found; skip remote tag operations\n", remote)
		}
	}
	remoteTagSHA := map[string]string{}
	if len(tags) > 0 && remoteExists {
		remoteTagSHA, err = lsRemoteTags(ctx, absDir, remote)
		if err != nil {
			return err
		}
	}

	plans := make([]tagPlan, 0, len(tags))
	for _, tagName := range tags {
		p := tagPlan{name: tagName}
		p.objectType, err = cmd.Run(ctx, absDir, "cat-file", "-t", tagName)
		if err != nil {
			return err
		}
		if p.objectType == "tag" {
			p.message, err = cmd.RunCombined(ctx, absDir, "tag", "-l", "--format=%(contents)", tagName)
			if err != nil {
				return err
			}
			p.taggerName, err = cmd.RunCombined(ctx, absDir, "tag", "-l", "--format=%(taggername)", tagName)
			if err != nil {
				return err
			}
			p.taggerEmail, err = cmd.RunCombined(ctx, absDir, "tag", "-l", "--format=%(taggeremail)", tagName)
			if err != nil {
				return err
			}
			p.taggerUnix, err = cmd.RunCombined(ctx, absDir, "tag", "-l", "--format=%(taggerdate:unix)", tagName)
			if err != nil {
				return err
			}
		}
		if remoteExists {
			if peeled, ok := remoteTagSHA[tagName]; ok {
				if peeled == oldSHA {
					p.remoteDelete = true
					p.remotePush = true
				} else {
					fmt.Fprintf(stderr, "warning: remote %s tag %s points at a different commit; skip remote delete\n", remote, tagName)
				}
			} else {
				p.remoteMissing = true
			}
		}
		plans = append(plans, p)
	}

	var pushBranches []string
	if doPush {
		for _, br := range branches {
			upSHA, ok, err := branchUpstreamSHA(ctx, absDir, br)
			if err != nil {
				return err
			}
			if ok && upSHA == oldSHA {
				pushBranches = append(pushBranches, br)
			}
		}
	}

	firstLine := messageFirstLine(newMsg)
	if dryRun {
		fmt.Fprintf(stdout, "[dry-run] would rewrite %s\n", oldSHA)
		fmt.Fprintf(stdout, "[dry-run]   author:  %s <%s>\n", newName, newEmail)
		fmt.Fprintf(stdout, "[dry-run]   message: %s\n", firstLine)
		if stripped {
			fmt.Fprint(stdout, "[dry-run]   stripped Co-authored-by\n")
		}
		for _, br := range branches {
			fmt.Fprintf(stdout, "[dry-run]   move branch %s\n", br)
		}
		for _, p := range plans {
			fmt.Fprintf(stdout, "[dry-run]   delete local tag %s\n", p.name)
			if p.remoteDelete {
				fmt.Fprintf(stdout, "[dry-run]   git push %s --delete refs/tags/%s\n", remote, p.name)
			}
			if p.remoteMissing {
				fmt.Fprintf(stdout, "[dry-run]   notice: remote %s has no tag %s\n", remote, p.name)
			}
			fmt.Fprintf(stdout, "[dry-run]   tag %s at new commit\n", p.name)
			if p.remotePush {
				fmt.Fprintf(stdout, "[dry-run]   git push %s %s\n", remote, p.name)
			}
		}
		return nil
	}

	newSHA, err := commitTree(ctx, absDir, meta, newName, newEmail, newMsg)
	if err != nil {
		return err
	}

	for _, br := range branches {
		if _, err := cmd.RunCombined(ctx, absDir, "update-ref", "refs/heads/"+br, newSHA); err != nil {
			return err
		}
	}
	if detached {
		if _, err := cmd.RunCombined(ctx, absDir, "update-ref", "HEAD", newSHA); err != nil {
			return err
		}
	}

	var tagLines []string
	for _, p := range plans {
		if _, err := cmd.RunCombined(ctx, absDir, "tag", "-d", p.name); err != nil {
			return err
		}
		if p.remoteDelete {
			if _, err := cmd.RunCombined(ctx, absDir, "push", remote, "--delete", "refs/tags/"+p.name); err != nil {
				return err
			}
		}
		if err := recreateTag(ctx, absDir, newSHA, p); err != nil {
			return err
		}
		if p.remotePush {
			if _, err := cmd.RunCombined(ctx, absDir, "push", remote, p.name); err != nil {
				return err
			}
			tagLines = append(tagLines, fmt.Sprintf("%s  delete local+%s, retag, push", p.name, remote))
		} else {
			tagLines = append(tagLines, fmt.Sprintf("%s  delete local, retag", p.name))
		}
	}

	var pushed []string
	for _, br := range pushBranches {
		remoteName, mergeRef, err := branchUpstream(ctx, absDir, br)
		if err != nil {
			return err
		}
		dst := strings.TrimPrefix(mergeRef, "refs/heads/")
		if _, err := cmd.RunCombined(ctx, absDir, "push", "--force-with-lease", remoteName, br+":"+dst); err != nil {
			return err
		}
		pushed = append(pushed, br)
	}

	fmt.Fprintf(stdout, "rewrote %s -> %s\n", oldSHA, newSHA)
	fmt.Fprintf(stdout, "  author:  %s <%s>\n", newName, newEmail)
	fmt.Fprintf(stdout, "  message: %s\n", firstLine)
	if stripped {
		fmt.Fprint(stdout, "  stripped Co-authored-by\n")
	}
	if len(branches) > 0 {
		fmt.Fprint(stdout, "  branches:\n")
		for _, br := range branches {
			fmt.Fprintf(stdout, "    %s\n", br)
		}
	}
	if len(tagLines) > 0 {
		fmt.Fprint(stdout, "  tags:\n")
		for _, line := range tagLines {
			fmt.Fprintf(stdout, "    %s\n", line)
		}
	}
	if len(pushed) > 0 {
		fmt.Fprint(stdout, "  pushed:\n")
		for _, br := range pushed {
			fmt.Fprintf(stdout, "    %s\n", br)
		}
	}
	for _, p := range plans {
		if p.remoteMissing {
			fmt.Fprintf(stdout, "notice: remote %s has no tag %s\n", remote, p.name)
		}
	}
	return nil
}

type tagPlan struct {
	name         string
	objectType   string
	message      string
	taggerName   string
	taggerEmail  string
	taggerUnix   string
	remoteDelete  bool
	remotePush    bool
	remoteMissing bool
}

type commitMeta struct {
	tree           string
	parents        []string
	authorName     string
	authorEmail    string
	authorUnix     string
	committerName  string
	committerEmail string
	committerUnix  string
	message        string
}

func loadCommit(ctx context.Context, dir, sha string) (*commitMeta, error) {
	const fmtSpec = "%an%x1f%ae%x1f%at%x1f%cn%x1f%ce%x1f%ct%x1f%T%x1f%P"
	line, err := cmd.Run(ctx, dir, "show", "-s", "--format="+fmtSpec, sha)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(line, "\x1f")
	if len(parts) != 8 {
		return nil, fmt.Errorf("Error: unable to parse commit %s", sha)
	}
	msg, err := cmd.RunCombined(ctx, dir, "log", "-1", "--format=%B", sha)
	if err != nil {
		return nil, err
	}
	return &commitMeta{
		authorName:     parts[0],
		authorEmail:    parts[1],
		authorUnix:     parts[2],
		committerName:  parts[3],
		committerEmail: parts[4],
		committerUnix:  parts[5],
		tree:           parts[6],
		parents:        strings.Fields(parts[7]),
		message:        msg,
	}, nil
}

func stripCoAuthor(msg string) (string, error) {
	lines := strings.Split(msg, "\n")
	kept := make([]string, 0, len(lines))
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed), "co-authored-by:") {
			found = true
			continue
		}
		kept = append(kept, line)
	}
	if !found {
		return "", fmt.Errorf("Error: commit message has no Co-authored-by line")
	}
	leftover := strings.TrimSpace(strings.Join(kept, "\n"))
	if leftover == "" {
		return "", fmt.Errorf("Error:")
	}
	return leftover + "\n", nil
}

func listExactTipBranches(ctx context.Context, dir, sha string) ([]string, error) {
	out, err := cmd.RunCombined(ctx, dir, "for-each-ref", "--format=%(refname:short)", "--points-at="+sha, "refs/heads")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	names := strings.Split(out, "\n")
	sort.Strings(names)
	return names, nil
}

func isDetachedAt(ctx context.Context, dir, sha string) (bool, error) {
	_, ok, err := cmd.RunOptional(ctx, dir, "symbolic-ref", "-q", "HEAD")
	if err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	head, err := cmd.Run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		return false, err
	}
	return head == sha, nil
}

func commitHasDescendants(ctx context.Context, dir, sha string) (bool, error) {
	out, err := cmd.RunCombined(ctx, dir, "rev-list", "--all", "--parents")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		for _, p := range fields[1:] {
			if p == sha {
				return true, nil
			}
		}
	}
	return false, nil
}

func listTagsAt(ctx context.Context, dir, sha string) ([]string, error) {
	out, err := cmd.RunCombined(ctx, dir, "tag", "--points-at", sha)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	names := strings.Split(out, "\n")
	sort.Strings(names)
	return names, nil
}

func remotePresent(ctx context.Context, dir, name string) (bool, error) {
	out, err := cmd.RunCombined(ctx, dir, "remote")
	if err != nil {
		return false, err
	}
	for _, r := range strings.Fields(out) {
		if r == name {
			return true, nil
		}
	}
	return false, nil
}

func lsRemoteTags(ctx context.Context, dir, remote string) (map[string]string, error) {
	out, err := cmd.RunCombined(ctx, dir, "ls-remote", "--tags", remote)
	if err != nil {
		return nil, err
	}
	tags := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		sha, ref := fields[0], fields[1]
		if !strings.HasPrefix(ref, "refs/tags/") {
			continue
		}
		name := strings.TrimPrefix(ref, "refs/tags/")
		if strings.HasSuffix(name, "^{}") {
			tags[strings.TrimSuffix(name, "^{}")] = sha
			continue
		}
		if _, ok := tags[name]; !ok {
			tags[name] = sha
		}
	}
	return tags, nil
}

func branchUpstreamSHA(ctx context.Context, dir, branch string) (string, bool, error) {
	out, ok, err := cmd.RunOptional(ctx, dir, "rev-parse", branch+"@{upstream}")
	if err != nil {
		// missing upstream can also surface as a non-empty error
		msg := err.Error()
		if strings.Contains(msg, "no upstream") || strings.Contains(msg, "does not point to a branch") {
			return "", false, nil
		}
		return "", false, err
	}
	if !ok || out == "" {
		return "", false, nil
	}
	return out, true, nil
}

func branchUpstream(ctx context.Context, dir, branch string) (remote, merge string, err error) {
	remote, err = cmd.Run(ctx, dir, "config", "--get", "branch."+branch+".remote")
	if err != nil {
		return "", "", err
	}
	merge, err = cmd.Run(ctx, dir, "config", "--get", "branch."+branch+".merge")
	if err != nil {
		return "", "", err
	}
	return remote, merge, nil
}

func commitTree(ctx context.Context, dir string, meta *commitMeta, authorName, authorEmail, message string) (string, error) {
	args := []string{"-c", "commit.gpgsign=false", "commit-tree", meta.tree}
	for _, p := range meta.parents {
		args = append(args, "-p", p)
	}
	args = append(args, "-m", strings.TrimSuffix(message, "\n"))
	env := []string{
		"GIT_AUTHOR_NAME=" + authorName,
		"GIT_AUTHOR_EMAIL=" + authorEmail,
		"GIT_AUTHOR_DATE=" + meta.authorUnix + " +0000",
		"GIT_COMMITTER_NAME=" + meta.committerName,
		"GIT_COMMITTER_EMAIL=" + meta.committerEmail,
		"GIT_COMMITTER_DATE=" + meta.committerUnix + " +0000",
	}
	return cmd.RunEnv(ctx, dir, env, args...)
}

func recreateTag(ctx context.Context, dir, newSHA string, p tagPlan) error {
	if p.objectType != "tag" {
		_, err := cmd.RunCombined(ctx, dir, "tag", p.name, newSHA)
		return err
	}
	email := strings.TrimSpace(p.taggerEmail)
	email = strings.TrimPrefix(email, "<")
	email = strings.TrimSuffix(email, ">")
	env := []string{
		"GIT_COMMITTER_NAME=" + p.taggerName,
		"GIT_COMMITTER_EMAIL=" + email,
		"GIT_COMMITTER_DATE=" + p.taggerUnix + " +0000",
	}
	_, err := cmd.RunEnv(ctx, dir, env, "-c", "tag.gpgsign=false", "tag", "-a", p.name, newSHA, "-m", strings.TrimSpace(p.message))
	return err
}

func messageFirstLine(msg string) string {
	line := msg
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		line = msg[:i]
	}
	return strings.TrimRight(line, "\r")
}

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	lessflags "github.com/xhd2015/less-flags"
)

const help = `
Usage: git-hook-no-co-authored-by [OPTIONS]

Reject push if any about-to-push commit message contains a Co-authored-by trailer.

Reads git pre-push stdin (one line per ref update):
  <local_ref> <local_sha> <remote_ref> <remote_sha>

Commit sets checked (already-pushed commits are excluded):
  - update:  git rev-list <remote_sha>..<local_sha>
  - new branch (remote_sha all-zero): git rev-list <local_sha> --not --remotes
  - delete (local_sha all-zero): skip

Matching is case-insensitive and requires a colon (Co-authored-by:).
Prose without a colon is allowed.

Options:
  --origin-domain DOMAIN            only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN    skip when remote origin host matches DOMAIN
  -h, --help                        show help message
`

var errCoAuthoredByFound = errors.New("co-authored-by trailer found")

type config struct {
	domainFilter githook.DomainFilter
	showHelp     bool
}

type pushUpdate struct {
	localRef  string
	localSHA  string
	remoteRef string
	remoteSHA string
}

type commitHit struct {
	sha     string
	subject string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errCoAuthoredByFound) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git-hook-no-co-authored-by: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithIO(args, os.Stdin, os.Stdout)
}

func runWithIO(args []string, stdin io.Reader, out io.Writer) error {
	cfg, err := parseArgs(args, out)
	if err != nil {
		return err
	}
	if cfg.showHelp {
		return nil
	}

	shouldRun, err := cfg.domainFilter.ShouldRun()
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	updates, err := parsePushUpdates(stdin)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		// Empty stdin (e.g. bare git-hooks pre-push run): nothing to check.
		return nil
	}

	seen := make(map[string]struct{})
	var hits []commitHit
	for _, u := range updates {
		shas, err := commitsToCheck(u)
		if err != nil {
			return err
		}
		for _, sha := range shas {
			if _, ok := seen[sha]; ok {
				continue
			}
			seen[sha] = struct{}{}
			msg, err := commitMessage(sha)
			if err != nil {
				return err
			}
			if messageHasCoAuthoredBy(msg) {
				hits = append(hits, commitHit{
					sha:     shortSHA(sha),
					subject: firstLine(msg),
				})
			}
		}
	}
	if len(hits) == 0 {
		return nil
	}

	fmt.Fprintln(out, "Error: cannot push commits with Co-authored-by trailer:")
	for _, h := range hits {
		fmt.Fprintf(out, "  %s  %s\n", h.sha, h.subject)
	}
	fmt.Fprintln(out, "Remove the trailer (e.g. git rebase / amend) or drop those commits before pushing.")
	return errCoAuthoredByFound
}

func parseArgs(args []string, out io.Writer) (config, error) {
	var cfg config
	var originDomain *string
	var excludeOriginDomain *string

	_, err := lessflags.
		String("--origin-domain", &originDomain).
		String("--exclude-origin-domain", &excludeOriginDomain).
		HelpFunc("-h,--help", func() {
			fmt.Fprint(out, strings.TrimPrefix(help, "\n"))
		}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if errors.Is(err, lessflags.ErrHelp) {
			cfg.showHelp = true
			return cfg, nil
		}
		return cfg, fmt.Errorf("parse flags: %w", err)
	}

	if originDomain != nil {
		cfg.domainFilter.OriginDomain = *originDomain
	}
	if excludeOriginDomain != nil {
		cfg.domainFilter.ExcludeOriginDomain = *excludeOriginDomain
	}
	if err := cfg.domainFilter.Normalize(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func parsePushUpdates(r io.Reader) ([]pushUpdate, error) {
	var updates []pushUpdate
	sc := bufio.NewScanner(r)
	// Commit SHAs + refs fit well under default buffer; raise for safety.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return nil, fmt.Errorf("pre-push stdin line %d: want 4 fields, got %d: %q", lineNo, len(fields), line)
		}
		updates = append(updates, pushUpdate{
			localRef:  fields[0],
			localSHA:  fields[1],
			remoteRef: fields[2],
			remoteSHA: fields[3],
		})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read pre-push stdin: %w", err)
	}
	return updates, nil
}

func commitsToCheck(u pushUpdate) ([]string, error) {
	if isZeroSHA(u.localSHA) {
		// Branch delete: nothing to push.
		return nil, nil
	}
	var args []string
	if isZeroSHA(u.remoteSHA) {
		// New remote branch: only commits not already on any remote-tracking ref.
		args = []string{"rev-list", u.localSHA, "--not", "--remotes"}
	} else {
		args = []string{"rev-list", u.remoteSHA + ".." + u.localSHA}
	}
	out, ok, err := githook.GitOptionalOutput(args...)
	if err != nil {
		return nil, err
	}
	if !ok || strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var shas []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		sha := strings.TrimSpace(line)
		if sha != "" {
			shas = append(shas, sha)
		}
	}
	return shas, nil
}

func commitMessage(sha string) (string, error) {
	msg, err := githook.GitOutput("log", "-1", "--format=%B", sha)
	if err != nil {
		return "", err
	}
	return msg, nil
}

// messageHasCoAuthoredBy reports trailer-style Co-authored-by: (colon required).
func messageHasCoAuthoredBy(msg string) bool {
	return strings.Contains(strings.ToLower(msg), "co-authored-by:")
}

func isZeroSHA(sha string) bool {
	if sha == "" {
		return true
	}
	for _, c := range sha {
		if c != '0' {
			return false
		}
	}
	return true
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

func firstLine(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return ""
	}
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		return strings.TrimSpace(msg[:i])
	}
	return msg
}

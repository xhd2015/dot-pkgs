package checkout

import (
	"context"
	"fmt"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

const defaultShortSHALength = 7

type Meta struct {
	Branch    string
	CommitSHA string
	CommitMsg string
	Status    string
	OriginURL string
	Error     string
}

type Options struct {
	ShortSHALength     int
	StatusStyle        status.FormatStyle
	PorcelainUntracked bool // default true when unset
}

func Enrich(ctx context.Context, repoPath string, opts Options) (meta Meta) {
	defer func() {
		enrichOrigin(ctx, repoPath, &meta)
	}()

	shortLen := opts.ShortSHALength
	if shortLen <= 0 {
		shortLen = defaultShortSHALength
	}

	if _, err := cmd.Run(ctx, repoPath, "rev-parse", "--verify", "HEAD"); err != nil {
		if isUnbornHEADError(err) {
			meta.Error = "no commits (HEAD unborn)"
			return meta
		}
		meta.Error = normalizeEnrichError(err)
		return meta
	}

	branch, err := cmd.Run(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		meta.Error = normalizeEnrichError(err)
		return meta
	}

	sha, err := cmd.Run(ctx, repoPath, "rev-parse", fmt.Sprintf("--short=%d", shortLen), "HEAD")
	if err != nil {
		meta.Error = normalizeEnrichError(err)
		return meta
	}

	meta.Branch = branch
	meta.CommitSHA = sha

	msg, err := cmd.Run(ctx, repoPath, "log", "-1", "--format=%s")
	if err != nil {
		meta.Error = normalizeEnrichError(err)
		return meta
	}
	meta.CommitMsg = sanitizeCommitMsg(msg)

	statusArgs := []string{"status", "--porcelain"}
	if opts.StatusStyle == status.StyleWrk && !opts.PorcelainUntracked {
		statusArgs = append(statusArgs, "--untracked-files=no")
	}
	porcelain, err := cmd.Run(ctx, repoPath, statusArgs...)
	if err != nil {
		meta.Error = mergeErrors(meta.Error, normalizeEnrichError(err))
		return meta
	}
	switch opts.StatusStyle {
	case status.StyleWrk:
		meta.Status = status.FormatWrk(status.ParsePorcelainWrk(porcelain))
	default:
		meta.Status = status.Format(status.ParsePorcelain(porcelain), status.FormatBackup)
	}
	return meta
}

func enrichOrigin(ctx context.Context, repoPath string, meta *Meta) {
	url, ok, err := cmd.RunOptional(ctx, repoPath, "config", "--get", "remote.origin.url")
	if err != nil || !ok {
		return
	}
	meta.OriginURL = url
}

func mergeErrors(existing, msg string) string {
	if msg == "" {
		return existing
	}
	if existing == "" {
		return msg
	}
	return existing + "; " + msg
}

func normalizeEnrichError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	msg = strings.TrimSpace(msg)
	if idx := strings.Index(msg, "\n"); idx >= 0 {
		msg = strings.TrimSpace(msg[:idx])
	}
	return msg
}

func isUnbornHEADError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown revision") ||
		strings.Contains(msg, "bad revision") ||
		strings.Contains(msg, "needed a single revision") ||
		strings.Contains(msg, "does not have any commits yet") ||
		strings.Contains(msg, "head unborn") ||
		strings.Contains(msg, "no such ref") ||
		(strings.Contains(msg, "rev-parse") && strings.Contains(msg, "head") && strings.Contains(msg, "exit status 128"))
}

func sanitizeCommitMsg(msg string) string {
	msg = strings.ReplaceAll(msg, "\r\n", " ")
	msg = strings.ReplaceAll(msg, "\n", " ")
	msg = strings.ReplaceAll(msg, "\r", " ")
	return strings.TrimSpace(msg)
}
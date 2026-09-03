package rg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// SearchOpts configures Search / SearchStream.
type SearchOpts struct {
	// Bin is the rg executable path (required).
	Bin string
	// Roots are directories/files to search (required; typically hub root).
	Roots []string
	// Pattern is a literal, case-insensitive needle (-i -F).
	Pattern string
	// Globs are passed as repeated -g (e.g. "*.md", "!__meta_knowledge__/**").
	Globs []string
	// MaxCount limits matches per file (-m); 0 = rg default (unlimited).
	MaxCount int
	// Run, when non-nil, replaces exec of rg for buffered Search (tests).
	Run func(ctx context.Context, bin string, args []string) (stdout []byte, exitCode int, err error)
	// RunStream, when non-nil, replaces exec for SearchStream (tests).
	// Reader is NDJSON lines; wait returns process exit after the reader is drained.
	RunStream func(ctx context.Context, bin string, args []string) (r io.ReadCloser, wait func() (exitCode int, err error), err error)
}

// Match is one content hit.
type Match struct {
	Path    string
	LineNum int
	Line    string
}

// Search runs rg -i -F --json over Roots and returns content matches.
// Exit code 1 (no matches) is success with empty results.
func Search(ctx context.Context, opts SearchOpts) ([]Match, error) {
	var out []Match
	err := SearchStream(ctx, opts, func(m Match) error {
		out = append(out, m)
		return nil
	})
	return out, err
}

// SearchStream runs rg and invokes emit for each match as soon as its JSON
// line is parsed. Exit code 1 (no matches) is success with zero emits.
// If emit returns a non-nil error, streaming stops and that error is returned.
// Stdout is always closed before Wait so an early emit stop cannot deadlock
// against a producer blocked on a full pipe.
func SearchStream(ctx context.Context, opts SearchOpts, emit func(Match) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if emit == nil {
		return fmt.Errorf("rg search: nil emit")
	}
	bin := strings.TrimSpace(opts.Bin)
	if bin == "" {
		return fmt.Errorf("rg search: empty bin")
	}
	pattern := opts.Pattern
	if pattern == "" {
		return fmt.Errorf("rg search: empty pattern")
	}
	if len(opts.Roots) == 0 {
		return fmt.Errorf("rg search: empty roots")
	}

	args := buildSearchArgs(opts)

	if opts.RunStream == nil && opts.Run != nil {
		// Buffered injectable: parse after full stdout (tests).
		stdout, exitCode, err := opts.Run(ctx, bin, args)
		if err != nil && exitCode != 1 {
			return fmt.Errorf("rg search: %w", err)
		}
		if exitCode == 2 {
			return fmt.Errorf("rg search: exit 2")
		}
		return parseJSONMatchesEmit(stdout, emit)
	}

	runStream := opts.RunStream
	if runStream == nil {
		runStream = defaultRunSearchStream
	}
	r, wait, err := runStream(ctx, bin, args)
	if err != nil {
		return fmt.Errorf("rg search: start: %w", err)
	}
	defer r.Close()

	scanErr := parseJSONMatchesReader(r, emit)
	// Close before Wait: early emit stop otherwise leaves the producer blocked
	// on a full stdout pipe while we wait for it to exit (kb find deadlock).
	_ = r.Close()
	exitCode, waitErr := wait()
	if scanErr != nil {
		return scanErr
	}
	if waitErr != nil && exitCode != 1 {
		return fmt.Errorf("rg search: %w", waitErr)
	}
	if exitCode == 2 {
		return fmt.Errorf("rg search: exit 2")
	}
	return nil
}

func buildSearchArgs(opts SearchOpts) []string {
	args := []string{"--json", "-i", "-F", "--color", "never"}
	if opts.MaxCount > 0 {
		args = append(args, "-m", fmt.Sprintf("%d", opts.MaxCount))
	}
	for _, g := range opts.Globs {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		args = append(args, "-g", g)
	}
	args = append(args, "--", opts.Pattern)
	args = append(args, opts.Roots...)
	return args
}

func defaultRunSearchStream(ctx context.Context, bin string, args []string) (io.ReadCloser, func() (int, error), error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	wait := func() (int, error) {
		err := cmd.Wait()
		if err == nil {
			return 0, nil
		}
		if ee, ok := err.(*exec.ExitError); ok {
			code := ee.ExitCode()
			if code == 1 {
				return 1, nil
			}
			if stderrBuf.Len() > 0 {
				return code, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderrBuf.String()))
			}
			return code, err
		}
		return -1, err
	}
	return stdout, wait, nil
}

// rg --json match message (subset).
type jsonMsg struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber int `json:"line_number"`
	} `json:"data"`
}

func parseJSONMatches(stdout []byte) ([]Match, error) {
	var out []Match
	err := parseJSONMatchesEmit(stdout, func(m Match) error {
		out = append(out, m)
		return nil
	})
	return out, err
}

func parseJSONMatchesEmit(stdout []byte, emit func(Match) error) error {
	return parseJSONMatchesReader(bytes.NewReader(stdout), emit)
}

func parseJSONMatchesReader(r io.Reader, emit func(Match) error) error {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var msg jsonMsg
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		if msg.Type != "match" {
			continue
		}
		text := strings.TrimRight(msg.Data.Lines.Text, "\r\n")
		if err := emit(Match{
			Path:    msg.Data.Path.Text,
			LineNum: msg.Data.LineNumber,
			Line:    text,
		}); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("rg search: scan json: %w", err)
	}
	return nil
}

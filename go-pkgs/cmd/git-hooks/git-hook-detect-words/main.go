package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
	lessflags "github.com/xhd2015/less-flags"
)

const help = `
Usage: git-hook-detect-words [OPTIONS] <keyword>...

Detect forbidden keywords in staged added or updated lines.

Options:
  --origin-domain DOMAIN           only run when remote origin host matches DOMAIN
  --exclude-origin-domain DOMAIN   skip when remote origin host matches DOMAIN
  -h,--help                        show help message
`

const (
	matchColor = "\x1b[01;31m"
	resetColor = "\x1b[m"
)

var errForbiddenWordsFound = errors.New("forbidden keywords found")
var hunkHeaderRE = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

type config struct {
	domainFilter githook.DomainFilter
	keywords     []string
	showHelp     bool
}

type finding struct {
	file string
	line int
	text string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errForbiddenWordsFound) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git-hook-detect-words: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithOutput(args, os.Stdout)
}

func runWithOutput(args []string, out io.Writer) error {
	cfg, err := parseArgs(args, out)
	if err != nil {
		return err
	}
	if cfg.showHelp {
		return nil
	}
	if len(cfg.keywords) == 0 {
		return fmt.Errorf("requires at least one keyword")
	}
	shouldRun, err := cfg.domainFilter.ShouldRun()
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	re, err := keywordRegexp(cfg.keywords)
	if err != nil {
		return err
	}
	diff, err := githook.GitOutput("diff", "--cached", "--no-ext-diff", "--no-color", "--unified=0", "--diff-filter=ACMRT", "--")
	if err != nil {
		return err
	}

	findings := findKeywordMatches(diff, re)
	for _, f := range findings {
		fmt.Fprintf(out, "%s:%d:%s\n", f.file, f.line, highlightMatches(f.text, re))
	}
	if len(findings) > 0 {
		return errForbiddenWordsFound
	}
	return nil
}

func parseArgs(args []string, out io.Writer) (config, error) {
	var cfg config
	var originDomain *string
	var excludeOriginDomain *string

	remaining, err := lessflags.
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
		return cfg, mapUnknownFlagErr(err)
	}

	if originDomain != nil {
		cfg.domainFilter.OriginDomain = *originDomain
	}
	if excludeOriginDomain != nil {
		cfg.domainFilter.ExcludeOriginDomain = *excludeOriginDomain
	}
	cfg.keywords = remaining

	if err := cfg.domainFilter.Normalize(); err != nil {
		return cfg, err
	}
	for _, keyword := range cfg.keywords {
		if keyword == "" {
			return cfg, fmt.Errorf("keyword must not be empty")
		}
	}
	return cfg, nil
}

func mapUnknownFlagErr(err error) error {
	const prefix = "unrecognized flag: "
	if msg := err.Error(); strings.HasPrefix(msg, prefix) {
		return fmt.Errorf("unknown flag: %s", strings.TrimPrefix(msg, prefix))
	}
	return err
}

func keywordRegexp(keywords []string) (*regexp.Regexp, error) {
	seen := make(map[string]bool, len(keywords))
	normalized := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		key := strings.ToLower(keyword)
		if seen[key] {
			continue
		}
		seen[key] = true
		normalized = append(normalized, keyword)
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		return len(normalized[i]) > len(normalized[j])
	})
	escaped := make([]string, 0, len(normalized))
	for _, keyword := range normalized {
		escaped = append(escaped, regexp.QuoteMeta(keyword))
	}
	return regexp.Compile(`(?i:` + strings.Join(escaped, "|") + `)`)
}

func findKeywordMatches(diff string, re *regexp.Regexp) []finding {
	var findings []finding
	var file string
	var newLine int
	var inHunk bool

	scanner := bufio.NewScanner(strings.NewReader(diff))
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "diff --git ") {
			file = ""
			inHunk = false
			continue
		}
		if matches := hunkHeaderRE.FindStringSubmatch(line); matches != nil {
			n, err := strconv.Atoi(matches[1])
			if err != nil {
				inHunk = false
				continue
			}
			newLine = n
			inHunk = true
			continue
		}
		if inHunk {
			if line == `\ No newline at end of file` {
				continue
			}
			if strings.HasPrefix(line, "+") {
				text := strings.TrimPrefix(line, "+")
				if file != "" && re.MatchString(text) {
					findings = append(findings, finding{
						file: file,
						line: newLine,
						text: text,
					})
				}
				newLine++
				continue
			}
			if strings.HasPrefix(line, " ") {
				newLine++
				continue
			}
			if strings.HasPrefix(line, "-") {
				continue
			}
			inHunk = false
		}
		if strings.HasPrefix(line, "+++ ") {
			file = diffPath(strings.TrimPrefix(line, "+++ "))
		}
	}
	return findings
}

func diffPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "/dev/null" {
		return ""
	}
	if strings.HasPrefix(path, `"`) {
		if unquoted, err := strconv.Unquote(path); err == nil {
			path = unquoted
		}
	}
	return strings.TrimPrefix(path, "b/")
}

func highlightMatches(line string, re *regexp.Regexp) string {
	ranges := re.FindAllStringIndex(line, -1)
	if len(ranges) == 0 {
		return line
	}
	var b strings.Builder
	last := 0
	for _, r := range ranges {
		if r[0] < last {
			continue
		}
		b.WriteString(line[last:r[0]])
		b.WriteString(matchColor)
		b.WriteString(line[r[0]:r[1]])
		b.WriteString(resetColor)
		last = r[1]
	}
	b.WriteString(line[last:])
	return b.String()
}

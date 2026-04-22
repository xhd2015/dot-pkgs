package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

//go:embed SKILL.md
var skillTemplate string

//go:embed topics
var topicsFS embed.FS

const topicsDir = "topics"

const help = `
Usage: go-best-practice-skill <command> [ARGS]
       go-best-practice-skill <topic>[/<sub-topic>[/...]]

Commands:
  install [<dir>]    Install SKILL.md + topics to a directory (or use --cursor)
  topics             List all available top-level topics
  <topic-path>       Print the detailed content for a topic or sub-topic

Topics are organized hierarchically. Address a nested topic with a
slash-separated path, e.g. "flags-parsing/types".

Options:
  -h, --help    Show this help message
`

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 {
		fmt.Print(help)
		fmt.Println()
		return printTopicIndex()
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Print(help)
		fmt.Println()
		return printTopicIndex()
	case "install", "create-skill":
		return handleInstall(args[1:])
	case "topics", "list":
		return printTopicIndex()
	}

	content, ok, err := readTopic(args[0])
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unknown command or topic: %s (run `go-best-practice-skill topics` to list available topics)", args[0])
	}
	fmt.Print(content)
	if !strings.HasSuffix(content, "\n") {
		fmt.Println()
	}
	return nil
}

// listTopics returns all topic paths found under topicsDir, relative
// to topicsDir and slash-separated. E.g. "flags-parsing",
// "flags-parsing/types". Each returned path corresponds to a
// reachable <path>.md file in the embedded FS.
func listTopics() ([]string, error) {
	var topics []string
	err := fs.WalkDir(topicsFS, topicsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".md") {
			return nil
		}
		rel, err := filepath.Rel(topicsDir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		rel = strings.TrimSuffix(rel, ".md")
		topics = append(topics, rel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk embedded topics: %w", err)
	}
	sort.Strings(topics)
	return topics, nil
}

func printTopicIndex() error {
	topics, err := listTopics()
	if err != nil {
		return err
	}
	fmt.Println("Available topics:")
	for _, t := range topics {
		depth := strings.Count(t, "/")
		indent := strings.Repeat("  ", depth)
		label := t
		if idx := strings.LastIndex(t, "/"); idx >= 0 {
			label = t[idx+1:]
		}
		fmt.Printf("  %s- %s\n", indent, label)
	}
	return nil
}

// readTopic resolves a slash-separated topic path like
// "flags-quickstart/types" against the embedded topics tree. A path
// segment is matched either as `<segment>.md` (leaf) or as
// `<segment>/` (directory) where the next segment is searched.
func readTopic(topicPath string) (string, bool, error) {
	topicPath = strings.Trim(topicPath, "/")
	if topicPath == "" {
		return "", false, nil
	}

	segments := strings.Split(topicPath, "/")
	if err := validateSegments(segments); err != nil {
		return "", false, err
	}

	embedPath := path.Join(append([]string{topicsDir}, segments...)...) + ".md"
	data, err := topicsFS.ReadFile(embedPath)
	if err == nil {
		return string(data), true, nil
	}
	if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("read topic %s: %w", topicPath, err)
	}
	return "", false, nil
}

func validateSegments(segments []string) error {
	for _, s := range segments {
		if s == "" || s == "." || s == ".." {
			return fmt.Errorf("invalid topic path segment: %q", s)
		}
	}
	return nil
}

func handleInstall(args []string) error {
	var dryRun, cursor, force bool
	args, err := flags.Bool("--dry-run", &dryRun).
		Bool("--cursor", &cursor).
		Bool("--force", &force).
		Help("-h,--help", `
Usage: install [OPTIONS] [<dir>]

Install SKILL.md + topics/ to a directory.

Options:
  --cursor     Install to .cursor/skills/go-best-practice (no dir argument needed)
  --force      Overwrite existing non-empty directory without prompting
  --dry-run    Show what would be created without actually creating anything
`).Parse(args)
	if err != nil {
		return err
	}

	var dir string
	if cursor {
		dir = filepath.Join(".cursor", "skills", "go-best-practice")
	} else if len(args) > 0 {
		dir = args[0]
	} else {
		return fmt.Errorf("install requires a directory path argument or --cursor flag")
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil && !os.IsNotExist(readErr) {
		return fmt.Errorf("read directory %s: %w", dir, readErr)
	}

	if readErr == nil && len(entries) > 0 && !force {
		if !confirmOverwrite(dir) {
			fmt.Println("Aborted.")
			return nil
		}
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove directory %s: %w", dir, err)
		}
		readErr = os.ErrNotExist
	}

	skillFile := filepath.Join(dir, "SKILL.md")

	files, err := collectTopicFiles(dir)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("[dry-run] Would create directory: %s\n", dir)
		fmt.Printf("[dry-run] Would create file: %s\n", skillFile)
		for _, f := range files {
			fmt.Printf("[dry-run] Would create file: %s\n", f.dest)
		}
		return nil
	}

	if readErr != nil {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(skillFile, []byte(skillTemplate), 0644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}

	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f.dest), 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", filepath.Dir(f.dest), err)
		}
		if err := os.WriteFile(f.dest, f.data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", f.dest, err)
		}
	}

	fmt.Printf("Installed skill to: %s\n", dir)
	fmt.Printf("  - %s\n", skillFile)
	for _, f := range files {
		fmt.Printf("  - %s\n", f.dest)
	}
	return nil
}

type installFile struct {
	dest string
	data []byte
}

func collectTopicFiles(dir string) ([]installFile, error) {
	var files []installFile
	err := fs.WalkDir(topicsFS, topicsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := topicsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}
		rel, err := filepath.Rel(topicsDir, path)
		if err != nil {
			return fmt.Errorf("rel path %s: %w", path, err)
		}
		files = append(files, installFile{
			dest: filepath.Join(dir, topicsDir, rel),
			data: data,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func confirmOverwrite(dir string) bool {
	f, _ := os.Stdin.Stat()
	if f == nil || (f.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	fmt.Printf("Directory %s is not empty. Overwrite? [y/N] ", dir)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

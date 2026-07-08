package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli/storage"
)

const (
	wrkMarkerBegin = "# === wrk integration begin ==="
	wrkMarkerEnd   = "# === wrk integration end ==="
)

const bashIntegrationScript = `#!/usr/bin/env bash
# wrk bash tab completion (installed at ${WRK_HOME:-$HOME/.wrk}/integration/bash.sh)

_wrk() {
  local candidates
  candidates=$(wrk --bash-integration --complete -- "${COMP_WORDS[@]}" "$COMP_CWORD")
  COMPREPLY=()
  if [[ -n "$candidates" ]]; then
    while IFS= read -r line; do
      [[ -n "$line" ]] && COMPREPLY+=("$line")
    done <<< "$candidates"
  fi
}

complete -F _wrk wrk
`

const wrkMarkerBlock = `# === wrk integration begin ===
_wrk_home="${WRK_HOME:-$HOME/.wrk}"
[[ -f "$_wrk_home/integration/bash.sh" ]] && source "$_wrk_home/integration/bash.sh"
# === wrk integration end ===
`

var wrkCompletionFlags = []string{
	"-h", "--help",
	"-l", "--list",
	"--done",
	"--merge-back",
	"--status",
	"--repos",
	"--projects",
	"--fetch",
	"-v", "--verbose",
	"--color",
	"--add",
	"--rm",
	"--where",
	"--dep",
	"--all-deps",
	"--dry-run",
	"-t", "--task",
	"--set-task",
	"-y", "--yes",
	"--confirm-from-stdin",
	"--no-in-module-replace",
	"--bash-integration",
	"--install",
	"--uninstall",
	"--complete",
}

var bashIntegrationDisallowedFlags = []string{
	"--done", "--merge-back", "-l", "--list", "--repos", "--projects",
	"--fetch", "-v", "--verbose", "--color", "--add", "--rm", "--where",
	"--dep", "--all-deps", "-t", "--task", "--set-task", "-y", "--yes",
	"--confirm-from-stdin", "--no-in-module-replace",
}

// CompletionRequest carries bash COMP_WORDS and COMP_CWORD for tab completion.
type CompletionRequest struct {
	Words []string
	CWord int
}

// ExitCodeError signals a non-zero exit without stderr output.
type ExitCodeError struct {
	Code int
}

func (e ExitCodeError) Error() string { return "" }

func runBashIntegration(args []string) error {
	if err := checkBashIntegrationMutualExclusion(args); err != nil {
		return err
	}

	action, dryRun, completeReq, err := parseBashIntegrationArgs(args)
	if err != nil {
		return err
	}

	switch action {
	case "":
		fmt.Print(bashIntegrationScript)
		return nil
	case "--install":
		if dryRun {
			return installBashIntegrationDryRun()
		}
		return installBashIntegration()
	case "--uninstall":
		if dryRun {
			return uninstallBashIntegrationDryRun()
		}
		return uninstallBashIntegration()
	case "--status":
		if dryRun {
			return fmt.Errorf("wrk: unknown integration action %q", "--dry-run")
		}
		if code := statusBashIntegration(); code != 0 {
			return ExitCodeError{Code: code}
		}
		return nil
	case "--complete":
		return runBashComplete(completeReq)
	default:
		return fmt.Errorf("wrk: unknown integration action %q", action)
	}
}

func checkBashIntegrationMutualExclusion(args []string) error {
	allowed := map[string]bool{
		"--bash-integration": true,
		"--install":          true,
		"--uninstall":        true,
		"--status":           true,
		"--dry-run":          true,
		"--complete":         true,
		"-h":                 true,
		"--help":             true,
		"--":                 true,
	}

	afterCompleteSep := false
	for _, arg := range args {
		if afterCompleteSep {
			continue
		}
		if arg == "--" {
			afterCompleteSep = true
			continue
		}
		if allowed[arg] {
			continue
		}
		for _, d := range bashIntegrationDisallowedFlags {
			if arg == d {
				return fmt.Errorf("wrk: --bash-integration is mutually exclusive with other modes")
			}
		}
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("wrk: --bash-integration is mutually exclusive with other modes")
		}
		return fmt.Errorf("wrk: --bash-integration is mutually exclusive with other modes")
	}
	return nil
}

func parseBashIntegrationArgs(args []string) (action string, dryRun bool, completeReq *CompletionRequest, err error) {
	afterBashIntegration := false
	afterCompleteSep := false
	var completeWords []string

	for _, arg := range args {
		if !afterBashIntegration {
			if arg == "--bash-integration" {
				afterBashIntegration = true
			}
			continue
		}
		if afterCompleteSep {
			completeWords = append(completeWords, arg)
			continue
		}
		switch arg {
		case "--install", "--uninstall", "--status", "--complete":
			if action != "" {
				return "", false, nil, fmt.Errorf("wrk: unknown integration action %q", arg)
			}
			action = arg
		case "--dry-run":
			dryRun = true
		case "--":
			if action != "--complete" {
				return "", false, nil, fmt.Errorf("wrk: unknown integration action %q", arg)
			}
			afterCompleteSep = true
		case "-h", "--help":
			// Hidden from main help; no dedicated integration help in tests.
		default:
			return "", false, nil, fmt.Errorf("wrk: unknown integration action %q", arg)
		}
	}

	if action == "--complete" {
		if len(completeWords) < 1 {
			return "", false, nil, fmt.Errorf("wrk: --complete requires words and cword after --")
		}
		cwordStr := completeWords[len(completeWords)-1]
		cword, convErr := strconv.Atoi(cwordStr)
		if convErr != nil {
			return "", false, nil, fmt.Errorf("wrk: invalid completion cword %q", cwordStr)
		}
		words := completeWords[:len(completeWords)-1]
		completeReq = &CompletionRequest{Words: words, CWord: cword}
	}
	return action, dryRun, completeReq, nil
}

func bashIntegrationPaths() (home, wrkHome, scriptPath, bashProfilePath, bashrcPath string, err error) {
	home, err = os.UserHomeDir()
	if err != nil {
		return "", "", "", "", "", err
	}
	wrkHome, err = resolveWrkHome()
	if err != nil {
		return "", "", "", "", "", err
	}
	scriptPath = filepath.Join(wrkHome, "integration", "bash.sh")
	bashProfilePath = filepath.Join(home, ".bash_profile")
	bashrcPath = filepath.Join(home, ".bashrc")
	return home, wrkHome, scriptPath, bashProfilePath, bashrcPath, nil
}

func scriptPresent(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func markerPresent(profilePath string) bool {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), wrkMarkerBegin)
}

func fullyInstalled(home string) bool {
	return markerPresent(filepath.Join(home, ".bash_profile")) &&
		markerPresent(filepath.Join(home, ".bashrc"))
}

func fullyUninstalled(home string) bool {
	return !markerPresent(filepath.Join(home, ".bash_profile")) &&
		!markerPresent(filepath.Join(home, ".bashrc"))
}

func installBashIntegrationDryRun() error {
	_, wrkHome, scriptPath, bashProfilePath, bashrcPath, err := bashIntegrationPaths()
	if err != nil {
		return err
	}
	home, _ := os.UserHomeDir()

	if fullyInstalled(home) {
		fmt.Println("wrk bash integration: already installed")
		if scriptPresent(scriptPath) {
			fmt.Printf("script: %s (exists)\n", scriptPath)
		} else {
			fmt.Printf("script: %s (absent)\n", scriptPath)
		}
		if markerPresent(bashProfilePath) {
			fmt.Printf("bash_profile: %s (marker present)\n", bashProfilePath)
		} else {
			fmt.Printf("bash_profile: %s (marker absent)\n", bashProfilePath)
		}
		if markerPresent(bashrcPath) {
			fmt.Printf("bashrc: %s (marker present)\n", bashrcPath)
		} else {
			fmt.Printf("bashrc: %s (marker absent)\n", bashrcPath)
		}
		fmt.Println("no changes needed")
		fmt.Println()
		return nil
	}

	fmt.Println("dry-run: would write integration/bash.sh")
	fmt.Println("dry-run: would append marker block to ~/.bash_profile")
	fmt.Println("dry-run: would append marker block to ~/.bashrc")
	fmt.Println()
	fmt.Print(wrkMarkerBlock)
	fmt.Println()
	_ = wrkHome
	return nil
}

func uninstallBashIntegrationDryRun() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	bashProfilePath := filepath.Join(home, ".bash_profile")
	bashrcPath := filepath.Join(home, ".bashrc")

	if fullyUninstalled(home) {
		fmt.Println("wrk bash integration: already uninstalled")
		fmt.Printf("bash_profile: %s (marker absent)\n", bashProfilePath)
		fmt.Printf("bashrc: %s (marker absent)\n", bashrcPath)
		fmt.Println("no changes needed")
		fmt.Println()
		return nil
	}

	fmt.Println("dry-run: would remove marker block from ~/.bash_profile")
	fmt.Println("dry-run: would remove marker block from ~/.bashrc")
	fmt.Println()
	fmt.Print(wrkMarkerBlock)
	fmt.Println()
	return nil
}

func statusBashIntegration() (exitCode int) {
	_, wrkHome, scriptPath, bashProfilePath, bashrcPath, err := bashIntegrationPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	scriptExists := scriptPresent(scriptPath)
	profileMarker := markerPresent(bashProfilePath)
	bashrcMarker := markerPresent(bashrcPath)

	var state string
	exitCode = 1
	switch {
	case scriptExists && profileMarker && bashrcMarker:
		state = "installed"
		exitCode = 0
	case !scriptExists && !profileMarker && !bashrcMarker:
		state = "not installed"
	default:
		state = "partial"
	}

	fmt.Printf("bash integration: %s\n", state)
	if scriptExists {
		fmt.Printf("script: %s (present)\n", scriptPath)
	} else {
		fmt.Printf("script: %s (absent)\n", scriptPath)
	}
	if profileMarker {
		fmt.Printf("bash_profile: %s (marker present)\n", bashProfilePath)
	} else {
		fmt.Printf("bash_profile: %s (marker absent)\n", bashProfilePath)
	}
	if bashrcMarker {
		fmt.Printf("bashrc: %s (marker present)\n", bashrcPath)
	} else {
		fmt.Printf("bashrc: %s (marker absent)\n", bashrcPath)
	}
	fmt.Println()
	_ = wrkHome
	return exitCode
}

func installBashIntegration() error {
	_, wrkHome, scriptPath, _, _, err := bashIntegrationPaths()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	integrationDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(integrationDir, 0o755); err != nil {
		return err
	}
	if !scriptPresent(scriptPath) {
		if err := os.WriteFile(scriptPath, []byte(bashIntegrationScript), 0o644); err != nil {
			return err
		}
	}

	for _, profilePath := range []string{
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".bashrc"),
	} {
		if err := appendMarkerToProfile(profilePath); err != nil {
			return err
		}
	}
	_ = wrkHome
	return nil
}

func appendMarkerToProfile(profilePath string) error {
	profile, err := os.ReadFile(profilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(profile)
	if strings.Contains(content, wrkMarkerBegin) {
		return nil
	}

	var builder strings.Builder
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		builder.WriteString(content)
		builder.WriteString("\n")
	} else {
		builder.WriteString(content)
	}
	builder.WriteString(wrkMarkerBlock)
	return os.WriteFile(profilePath, []byte(builder.String()), 0o644)
}

func uninstallBashIntegration() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	for _, profilePath := range []string{
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".bashrc"),
	} {
		if err := stripMarkerFromProfile(profilePath); err != nil {
			return err
		}
	}
	return nil
}

func stripMarkerFromProfile(profilePath string) error {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	var out []string
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == wrkMarkerBegin {
			inBlock = true
			continue
		}
		if trimmed == wrkMarkerEnd {
			inBlock = false
			continue
		}
		if inBlock {
			continue
		}
		out = append(out, line)
	}

	newContent := strings.Join(out, "\n")
	if len(data) > 0 && strings.HasSuffix(string(data), "\n") && newContent != "" {
		newContent += "\n"
	}
	return os.WriteFile(profilePath, []byte(newContent), 0o644)
}

func runBashComplete(req *CompletionRequest) error {
	if req == nil {
		return fmt.Errorf("wrk: --complete requires words and cword after --")
	}
	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	candidates := Complete(wrkHome, *req)
	if len(candidates) == 0 {
		return nil
	}
	for _, c := range candidates {
		fmt.Println(c)
	}
	fmt.Println()
	return nil
}

// Complete returns bash tab-completion candidates for the given request.
func Complete(wrkHome string, req CompletionRequest) []string {
	if req.CWord < 0 || req.CWord >= len(req.Words) {
		return nil
	}
	cur := req.Words[req.CWord]

	kind, prefix := completionContext(req.Words, req.CWord)
	switch kind {
	case "flags":
		return filterFlags(prefix)
	case "basenames":
		candidates, err := listBasenameCandidates(wrkHome, prefix)
		if err != nil {
			return nil
		}
		return candidates
	default:
		_ = cur
		return nil
	}
}

func completionContext(words []string, cword int) (kind, prefix string) {
	if cword < 0 || cword >= len(words) {
		return "none", ""
	}
	cur := words[cword]

	if strings.HasPrefix(cur, "-") {
		return "flags", cur
	}

	if cword > 0 {
		switch words[cword-1] {
		case "--dep", "--where", "--add", "--rm", "-l", "--list", "--status":
			return "basenames", cur
		case "-t", "--task", "--set-task":
			return "none", ""
		}
	}

	if cword == 1 && !strings.HasPrefix(cur, "-") {
		return "basenames", cur
	}

	if cword == 2 && len(words) > 1 && !strings.HasPrefix(words[1], "-") {
		return "basenames", cur
	}

	return "none", ""
}

func filterFlags(prefix string) []string {
	var out []string
	for _, flag := range wrkCompletionFlags {
		if strings.HasPrefix(flag, prefix) {
			out = append(out, flag)
		}
	}
	return out
}

func listBasenameCandidates(wrkHome, prefix string) ([]string, error) {
	paths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var basenames []string
	for _, p := range paths {
		base := filepath.Base(p)
		if _, ok := seen[base]; ok {
			continue
		}
		seen[base] = struct{}{}
		if prefix == "" || strings.HasPrefix(base, prefix) {
			basenames = append(basenames, base)
		}
	}
	sort.Strings(basenames)
	return basenames, nil
}
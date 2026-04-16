package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

//go:embed SKILL.md
var skillTemplate string

const help = `
Usage: playwright-debug-skill <command> [ARGS]

Commands:
  run <js_script>    Run a Playwright script (default if no command given)
  install [<dir>]    Install skill SKILL.md to a directory (or use --cursor)

The run command wraps your script with browser setup. You get:
  - browser  (Chromium instance)
  - page     (new page in browser)
  - chromium (Playwright chromium object)

Example:
  playwright-debug-skill 'await page.goto("https://example.com"); console.log(await page.title());'

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
		return nil
	}

	for _, a := range args {
		if a == "-h" || a == "--help" {
			fmt.Print(help)
			return nil
		}
	}

	switch args[0] {
	case "install", "create-skill":
		return handleInstall(args[1:])
	case "run":
		if len(args) < 2 {
			return fmt.Errorf("run requires a JavaScript script argument")
		}
		return handleRun(args[1])
	default:
		return handleRun(strings.Join(args, " "))
	}
}

func handleInstall(args []string) error {
	var dryRun, cursor, force bool
	args, err := flags.Bool("--dry-run", &dryRun).
		Bool("--cursor", &cursor).
		Bool("--force", &force).
		Help("-h,--help", `
Usage: install [OPTIONS] [<dir>]

Install skill SKILL.md to a directory.

Options:
  --cursor     Install to .cursor/skills/playwright-debug (no dir argument needed)
  --force      Overwrite existing non-empty directory without prompting
  --dry-run    Show what would be created without actually creating anything
`).Parse(args)
	if err != nil {
		return err
	}

	var dir string
	if cursor {
		dir = filepath.Join(".cursor", "skills", "playwright-debug")
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

	if dryRun {
		fmt.Printf("[dry-run] Would create directory: %s\n", dir)
		fmt.Printf("[dry-run] Would create file: %s\n", skillFile)
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

	fmt.Printf("Installed skill to: %s\n", dir)
	fmt.Printf("  - %s\n", skillFile)
	return nil
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

func cacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".playwright-debug", "node_package")
}

func ensurePlaywright() (string, error) {
	dir := cacheDir()

	packageJSON := filepath.Join(dir, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Println("Initializing playwright cache directory...")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("create cache dir: %w", err)
		}
		cmd := exec.Command("npm", "init", "-y")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("npm init: %w", err)
		}
	}

	nodeModules := filepath.Join(dir, "node_modules", "playwright")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		fmt.Println("Installing playwright...")
		cmd := exec.Command("npm", "install", "playwright")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("npm install playwright: %w", err)
		}

		fmt.Println("Installing Chromium browser...")
		npx := "npx"
		if runtime.GOOS == "windows" {
			npx = "npx.cmd"
		}
		cmd = exec.Command(npx, "playwright", "install", "chromium")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("playwright install chromium: %w", err)
		}
	}

	return dir, nil
}

func handleRun(script string) error {
	dir, err := ensurePlaywright()
	if err != nil {
		return err
	}

	wrapper := fmt.Sprintf(`
const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  try {
    %s
  } finally {
    await browser.close();
  }
})();
`, script)

	cmd := exec.Command("node", "-e", wrapper)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("playwright script failed: %w", err)
	}
	return nil
}

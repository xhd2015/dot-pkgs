package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/ai/skills"
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
	return skills.HandleInstall(skills.InstallOptions{
		SkillDirName: "playwright-debug",
		SkillContent: skillTemplate,
		Usage:        "install",
	}, args)
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

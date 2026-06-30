package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// GoMod represents the structure returned by "go mod edit -json".
type GoMod struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
	Replace []struct {
		Old struct {
			Path string `json:"Path"`
		} `json:"Old"`
		New struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"New"`
	} `json:"Replace"`
}

// GoModEditOptions represents options for go mod edit commands.
type GoModEditOptions struct {
	Dir    string // Working directory for the command
	Stderr bool   // Whether to show stderr output
	Stdout bool   // Whether to show stdout output
}

// DefaultGoModEditOptions returns default options for go mod edit commands.
func DefaultGoModEditOptions() *GoModEditOptions {
	return &GoModEditOptions{
		Dir:    ".",
		Stderr: true,
		Stdout: true,
	}
}

func (opts *GoModEditOptions) dir() string {
	if opts == nil || opts.Dir == "" {
		return "."
	}
	return opts.Dir
}

// GoModEditJSON executes "go mod edit -json" and returns the parsed module information.
func GoModEditJSON(opts *GoModEditOptions) (*GoMod, error) {
	if opts == nil {
		opts = DefaultGoModEditOptions()
	}

	modCmd := exec.Command("go", "mod", "edit", "-json")
	modCmd.Dir = opts.dir()
	if opts.Stderr {
		modCmd.Stderr = os.Stderr
	}

	output, err := modCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute go mod edit -json: %w", err)
	}

	var result GoMod
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal go mod edit -json output: %w", err)
	}

	return &result, nil
}

// GoModEditReplace executes "go mod edit -replace" to add a replacement directive.
func GoModEditReplace(modulePath, replacement string, opts *GoModEditOptions) error {
	if opts == nil {
		opts = DefaultGoModEditOptions()
	}

	replaceArg := fmt.Sprintf("%s=%s", modulePath, replacement)
	editCmd := exec.Command("go", "mod", "edit", "-replace", replaceArg)
	editCmd.Dir = opts.dir()

	if opts.Stderr {
		editCmd.Stderr = os.Stderr
	}
	if opts.Stdout {
		editCmd.Stdout = os.Stdout
	}

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute go mod edit -replace %s: %w", replaceArg, err)
	}

	return nil
}

// GoModDropReplace executes "go mod edit -dropreplace" to remove a replacement directive.
func GoModDropReplace(modulePath string, opts *GoModEditOptions) error {
	if opts == nil {
		opts = DefaultGoModEditOptions()
	}

	dropCmd := exec.Command("go", "mod", "edit", "-dropreplace", modulePath)
	dropCmd.Dir = opts.dir()
	dropCmd.Stderr = os.Stderr

	if err := dropCmd.Run(); err != nil {
		return fmt.Errorf("failed to drop replacement for %s: %w", modulePath, err)
	}
	return nil
}

// GoModEditRequire executes "go mod edit -require" to update module version.
func GoModEditRequire(modulePath, version string, opts *GoModEditOptions) error {
	if opts == nil {
		opts = DefaultGoModEditOptions()
	}

	requireArg := fmt.Sprintf("-require=%s@%s", modulePath, version)
	requireCmd := exec.Command("go", "mod", "edit", requireArg)
	requireCmd.Dir = opts.dir()
	requireCmd.Stderr = os.Stderr

	if err := requireCmd.Run(); err != nil {
		return fmt.Errorf("failed to update module version for %s@%s: %w", modulePath, version, err)
	}
	return nil
}

// GoModTidy executes "go mod tidy" in the specified directory.
func GoModTidy(opts *GoModEditOptions) error {
	if opts == nil {
		opts = DefaultGoModEditOptions()
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = opts.dir()

	if opts.Stderr {
		tidyCmd.Stderr = os.Stderr
	}
	if opts.Stdout {
		tidyCmd.Stdout = os.Stdout
	}

	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute go mod tidy: %w", err)
	}

	return nil
}
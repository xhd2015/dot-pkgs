package update

import (
	"fmt"
	"os"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
)

// PinOptions configures Pin.
type PinOptions struct {
	ConsumerDir string // go.mod directory to edit
	DepDir      string // dependency module directory (for path + tags)
	Version     string // optional go require version e.g. "v0.0.97"; empty = latest tag
	DryRun      bool
}

// PinResult is the planned or applied pin outcome.
type PinResult struct {
	ModulePath string
	Tag        string // full git tag when resolved from tags; may be empty if Version forced without tag lookup
	Version    string // go require version
}

// Pin drops replace for the dep module in ConsumerDir and sets require to the
// resolved version. When Version is empty, the latest matching git tag under
// DepDir is used. DryRun fills PinResult without writing consumer go.mod.
// Edits use commands.GoModEditOptions{Dir: ConsumerDir} — no process Chdir.
func Pin(opts PinOptions) (PinResult, error) {
	if opts.DepDir == "" {
		return PinResult{}, fmt.Errorf("requires DepDir")
	}
	if opts.ConsumerDir == "" {
		return PinResult{}, fmt.Errorf("requires ConsumerDir")
	}
	if _, err := os.Stat(opts.DepDir); os.IsNotExist(err) {
		return PinResult{}, fmt.Errorf("no such dir: %s", opts.DepDir)
	}

	depOpts := &commands.GoModEditOptions{Dir: opts.DepDir, Stderr: true}
	mod, err := commands.GoModEditJSON(depOpts)
	if err != nil {
		return PinResult{}, fmt.Errorf("failed to get module info: %w", err)
	}
	if mod.Module.Path == "" {
		return PinResult{}, fmt.Errorf("not a go module: %s", opts.DepDir)
	}

	result := PinResult{ModulePath: mod.Module.Path}

	if opts.Version != "" {
		if !isValidVersionTag(opts.Version) {
			return PinResult{}, fmt.Errorf("invalid version %s (want vN.N.N)", opts.Version)
		}
		result.Version = opts.Version
		// Tag left empty when Version is forced without tag lookup.
	} else {
		versionPrefix, err := CalculateVersionPrefix(opts.DepDir, mod.Module.Path)
		if err != nil {
			return PinResult{}, fmt.Errorf("failed to calculate version prefix for %s: %w", mod.Module.Path, err)
		}
		latestTag, err := GetLatestVersionTag(opts.DepDir, versionPrefix)
		if err != nil {
			return PinResult{}, fmt.Errorf("failed to get latest version tag for %s: %w", mod.Module.Path, err)
		}
		version := StripVersionPrefix(versionPrefix, latestTag)
		if !isValidVersionTag(version) {
			return PinResult{}, fmt.Errorf("latest version tag %s resolved to invalid version %s", latestTag, version)
		}
		result.Tag = latestTag
		result.Version = version
	}

	if opts.DryRun {
		return result, nil
	}

	consumerOpts := &commands.GoModEditOptions{Dir: opts.ConsumerDir}
	if err := commands.GoModDropReplace(mod.Module.Path, consumerOpts); err != nil {
		return PinResult{}, err
	}
	if err := commands.GoModEditRequire(mod.Module.Path, result.Version, consumerOpts); err != nil {
		return PinResult{}, err
	}
	return result, nil
}

package air

import (
	"fmt"
	"strconv"
	"strings"
)

// BuildArgs turns Options into air CLI flags (without the "air" binary name).
func BuildArgs(opts Options) ([]string, error) {
	if strings.TrimSpace(opts.BuildCmd) == "" {
		return nil, fmt.Errorf("air: BuildCmd is required")
	}
	if strings.TrimSpace(opts.Entrypoint) == "" {
		return nil, fmt.Errorf("air: Entrypoint is required")
	}

	delayMs := opts.DelayMs
	if delayMs <= 0 {
		delayMs = 500
	}
	includeExt := opts.IncludeExt
	if len(includeExt) == 0 {
		includeExt = []string{"go"}
	}
	excludeRegex := opts.ExcludeRegex
	if excludeRegex == "" {
		excludeRegex = `_test\.go`
	}
	excludeUnchanged := true
	if opts.ExcludeUnchanged != nil {
		excludeUnchanged = *opts.ExcludeUnchanged
	}

	args := []string{
		"--build.cmd", opts.BuildCmd,
		"--build.entrypoint", opts.Entrypoint,
		"--build.delay", strconv.Itoa(delayMs),
		"--build.include_ext", strings.Join(includeExt, ","),
		"--build.exclude_regex", excludeRegex,
		"--build.exclude_unchanged", strconv.FormatBool(excludeUnchanged),
	}
	if len(opts.IncludeDir) > 0 {
		args = append(args, "--build.include_dir", strings.Join(opts.IncludeDir, ","))
	}
	if len(opts.ExcludeDir) > 0 {
		args = append(args, "--build.exclude_dir", strings.Join(opts.ExcludeDir, ","))
	}
	for _, a := range opts.ArgsBin {
		args = append(args, "--build.args_bin", a)
	}
	return args, nil
}

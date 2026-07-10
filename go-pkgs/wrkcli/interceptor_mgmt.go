package wrkcli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli/storage"
)

// Neutral --init stub argv (disabled until the user configures a real interceptor).
var interceptorInitStubArgv = []string{"echo", "wrk-interceptor-not-configured"}

var interceptorMgmtActions = map[string]bool{
	"--status":  true,
	"--show":    true,
	"--path":    true,
	"--enable":  true,
	"--disable": true,
	"--init":    true,
	"--check":   true,
	"--dry-run": true,
}

// Flags that belong to other wrk modes and must not appear with --interceptor.
var interceptorDisallowedFlags = []string{
	"--done", "--merge-back", "-l", "--list", "--repos", "--projects",
	"--fetch", "-v", "--verbose", "--color", "--add", "--rm", "--where",
	"--cd", "--dep", "--all-deps", "--set-task", "-y", "--yes",
	"--confirm-from-stdin", "--no-in-module-replace", "--no-cd", "--force-cd",
	"--no-interceptor", "--bash-integration", "--install", "--uninstall",
	"--complete",
}

type interceptorMgmtOpts struct {
	action     string
	force      bool
	help       bool
	createArgs []string // remaining args after --dry-run [--]
}

// runInterceptorMgmt handles wrk --interceptor <action> [...].
func runInterceptorMgmt(origWd string, args []string, ctx *invocationContext) error {
	opts, err := parseInterceptorMgmtArgs(args)
	if err != nil {
		return err
	}
	if opts.help {
		fmt.Print(interceptorMgmtUsage())
		return nil
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	ctx.wrkHome = wrkHome
	ctx.workDir = origWd
	ctx.command = "interceptor"
	ctx.eventArgs = interceptorEventArgs(opts)
	if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
		return err
	}

	switch opts.action {
	case "--path":
		return interceptorMgmtPath(wrkHome)
	case "--status":
		return interceptorMgmtStatus(wrkHome)
	case "--show":
		return interceptorMgmtShow(wrkHome)
	case "--enable":
		return interceptorMgmtEnable(wrkHome)
	case "--disable":
		return interceptorMgmtDisable(wrkHome)
	case "--init":
		return interceptorMgmtInit(wrkHome, opts.force)
	case "--check":
		return interceptorMgmtCheck(wrkHome, origWd)
	case "--dry-run":
		return interceptorMgmtDryRun(wrkHome, origWd, opts.createArgs)
	default:
		return fmt.Errorf("wrk: --interceptor requires an action (--status, --show, --path, --enable, --disable, --init, --check, --dry-run)")
	}
}

func interceptorEventArgs(opts interceptorMgmtOpts) []string {
	out := []string{opts.action}
	if opts.force {
		out = append(out, "--force")
	}
	if len(opts.createArgs) > 0 {
		out = append(out, opts.createArgs...)
	}
	return out
}

func parseInterceptorMgmtArgs(args []string) (interceptorMgmtOpts, error) {
	var opts interceptorMgmtOpts
	seenInterceptor := false
	inCreateArgs := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if inCreateArgs {
			opts.createArgs = append(opts.createArgs, arg)
			continue
		}

		if arg == "--interceptor" {
			if seenInterceptor {
				return opts, fmt.Errorf("wrk: duplicate --interceptor")
			}
			seenInterceptor = true
			continue
		}

		// Flags before --interceptor: treat as mutual exclusion with other modes.
		if !seenInterceptor {
			if strings.HasPrefix(arg, "-") {
				return opts, fmt.Errorf("wrk: --interceptor is mutually exclusive with other modes")
			}
			return opts, fmt.Errorf("wrk: --interceptor is mutually exclusive with other modes")
		}

		if interceptorMgmtActions[arg] {
			if opts.action != "" {
				return opts, fmt.Errorf("wrk: --interceptor requires exactly one action (got %s and %s)", opts.action, arg)
			}
			opts.action = arg
			if arg == "--dry-run" {
				// Remaining args (after optional --) are create-args for expansion.
				// Consume the rest of the loop via inCreateArgs after optional --.
				if i+1 < len(args) && args[i+1] == "--" {
					i++ // skip --
				}
				inCreateArgs = true
			}
			continue
		}

		switch arg {
		case "--force":
			opts.force = true
		case "-h", "--help":
			opts.help = true
		case "--":
			return opts, fmt.Errorf("wrk: unexpected -- outside --dry-run")
		default:
			if isInterceptorDisallowedFlag(arg) {
				return opts, fmt.Errorf("wrk: --interceptor is mutually exclusive with other modes")
			}
			if strings.HasPrefix(arg, "-") {
				return opts, fmt.Errorf("wrk: --interceptor is mutually exclusive with other modes")
			}
			return opts, fmt.Errorf("wrk: --interceptor is mutually exclusive with other modes")
		}
	}

	if !seenInterceptor {
		return opts, fmt.Errorf("wrk: --interceptor required")
	}

	if opts.help {
		return opts, nil
	}

	if opts.action == "" {
		return opts, fmt.Errorf("wrk: --interceptor requires an action (--status, --show, --path, --enable, --disable, --init, --check, --dry-run)")
	}

	if opts.force && opts.action != "--init" {
		return opts, fmt.Errorf("wrk: --force is only valid with --interceptor --init")
	}

	return opts, nil
}

func isInterceptorDisallowedFlag(arg string) bool {
	// Strip =value forms if any.
	name := arg
	if i := strings.IndexByte(arg, '='); i >= 0 {
		name = arg[:i]
	}
	for _, d := range interceptorDisallowedFlags {
		if name == d {
			return true
		}
	}
	// --status as wrk mode is an interceptor action when after --interceptor;
	// --list etc. already listed. Also catch --task/-t outside dry-run create args
	// (they only appear inside create args after --dry-run).
	return false
}

func interceptorConfigPath(wrkHome string) string {
	return filepath.Join(wrkHome, "config.json")
}

func interceptorMgmtPath(wrkHome string) error {
	fmt.Println(interceptorConfigPath(wrkHome))
	return nil
}

func interceptorMgmtStatus(wrkHome string) error {
	path := interceptorConfigPath(wrkHome)
	ic, err := loadCreateInterceptor(wrkHome)
	if err != nil {
		return err
	}
	state := "absent"
	argv0 := "-"
	if ic != nil {
		if ic.Enabled {
			state = "enabled"
		} else {
			state = "disabled"
		}
		if len(ic.Argv) > 0 {
			argv0 = ic.Argv[0]
		}
	}
	fmt.Printf("state: %s\npath: %s\nargv0: %s\n", state, path, argv0)
	return nil
}

func interceptorMgmtShow(wrkHome string) error {
	ic, err := loadCreateInterceptor(wrkHome)
	if err != nil {
		return err
	}
	if ic == nil {
		return fmt.Errorf("wrk: no create interceptor configured")
	}
	data, err := json.MarshalIndent(ic, "", "  ")
	if err != nil {
		return fmt.Errorf("wrk: marshal interceptor: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func interceptorMgmtEnable(wrkHome string) error {
	root, err := loadConfigMap(wrkHome)
	if err != nil {
		return err
	}
	ic, err := interceptorFromConfigMap(root)
	if err != nil {
		return err
	}
	if ic == nil {
		return fmt.Errorf("wrk: no create interceptor configured; run wrk --interceptor --init first")
	}
	ic.Enabled = true
	if err := setInterceptorInConfigMap(root, ic); err != nil {
		return err
	}
	return saveConfigMap(wrkHome, root)
}

func interceptorMgmtDisable(wrkHome string) error {
	root, err := loadConfigMap(wrkHome)
	if err != nil {
		return err
	}
	if root == nil {
		// No config / no block: no-op success.
		return nil
	}
	ic, err := interceptorFromConfigMap(root)
	if err != nil {
		return err
	}
	if ic == nil {
		return nil
	}
	ic.Enabled = false
	if err := setInterceptorInConfigMap(root, ic); err != nil {
		return err
	}
	return saveConfigMap(wrkHome, root)
}

func interceptorMgmtInit(wrkHome string, force bool) error {
	root, err := loadConfigMap(wrkHome)
	if err != nil {
		return err
	}
	if root == nil {
		root = map[string]interface{}{}
	}

	existing, err := interceptorFromConfigMap(root)
	if err != nil {
		return err
	}
	if existing != nil && !force {
		return fmt.Errorf("wrk: create interceptor already exists (use --force to overwrite)")
	}

	// Ensure version if missing.
	if _, ok := root["version"]; !ok {
		root["version"] = 1
	}

	stub := &CreateInterceptor{
		Enabled: false,
		Argv:    append([]string(nil), interceptorInitStubArgv...),
	}
	if err := setInterceptorInConfigMap(root, stub); err != nil {
		return err
	}
	return saveConfigMap(wrkHome, root)
}

func interceptorMgmtCheck(wrkHome, cwd string) error {
	ic, err := loadCreateInterceptor(wrkHome)
	if err != nil {
		return err
	}
	if ic == nil {
		return nil
	}
	if ic.Enabled && len(ic.Argv) == 0 {
		return fmt.Errorf("wrk: create interceptor enabled but argv is empty")
	}
	if len(ic.Argv) == 0 {
		return nil
	}
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		absCwd = cwd
	}
	_, err = expandCreateInterceptor(ic, createInterceptorInput{
		wrkHome: wrkHome,
		workDir: absCwd,
		origWd:  absCwd,
	})
	return err
}

func interceptorMgmtDryRun(wrkHome, cwd string, createArgs []string) error {
	ic, err := loadCreateInterceptor(wrkHome)
	if err != nil {
		return err
	}
	if ic == nil {
		return fmt.Errorf("wrk: no create interceptor configured")
	}
	if !ic.Enabled {
		return fmt.Errorf("wrk: create interceptor is not enabled")
	}
	if len(ic.Argv) == 0 {
		return fmt.Errorf("wrk: create interceptor enabled but argv is empty")
	}

	task, remaining, err := peelTaskFromCreateArgs(createArgs)
	if err != nil {
		return err
	}

	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		absCwd = cwd
	}

	expanded, err := expandCreateInterceptor(ic, createInterceptorInput{
		wrkHome: wrkHome,
		workDir: absCwd,
		origWd:  absCwd,
		task:    task,
		args:    remaining,
	})
	if err != nil {
		return err
	}
	for _, a := range expanded {
		fmt.Println(a)
	}
	return nil
}

// peelTaskFromCreateArgs extracts -t/--task value and returns remaining args.
func peelTaskFromCreateArgs(args []string) (task string, remaining []string, err error) {
	remaining = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-t" || a == "--task":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("wrk: %s requires a value", a)
			}
			task = args[i+1]
			i++
		case strings.HasPrefix(a, "--task="):
			task = strings.TrimPrefix(a, "--task=")
		case strings.HasPrefix(a, "-t="):
			task = strings.TrimPrefix(a, "-t=")
		default:
			remaining = append(remaining, a)
		}
	}
	return task, remaining, nil
}

func interceptorMgmtUsage() string {
	return `wrk --interceptor — manage create.interceptor in $WRK_HOME/config.json

Usage:
  wrk --interceptor --status
  wrk --interceptor --show
  wrk --interceptor --path
  wrk --interceptor --enable
  wrk --interceptor --disable
  wrk --interceptor --init [--force]
  wrk --interceptor --check
  wrk --interceptor --dry-run [--] [create-args...]

Actions:
  --status     print state (absent|disabled|enabled), path, argv0
  --show       pretty-print create.interceptor JSON
  --path       print absolute path to config.json
  --enable     set enabled=true (requires existing interceptor; see --init)
  --disable    set enabled=false
  --init       write disabled neutral stub (refuse if exists unless --force)
  --check      validate interceptor when present
  --dry-run    expand argv with create-args; print one arg per line (no exec)
`
}

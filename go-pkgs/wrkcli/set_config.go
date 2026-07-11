package wrkcli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli/storage"
)

// set-config is mutually exclusive with these mode / create-flow flags.
var setConfigDisallowedFlags = []string{
	"--done", "--merge-back", "-l", "--list", "--status", "--repos", "--projects",
	"--fetch", "--add", "--rm", "--where", "--cd", "--main",
	"--dep", "--all-deps", "--dry-run",
	"-t", "--task", "--set-task",
	"--exec",
	"--bash-integration", "--interceptor",
	"--no-interceptor",
}

type setConfigOpts struct {
	show   bool
	create bool

	newWindow     bool
	noNewWindow   bool
	newTerminal   bool
	reuseTerminal bool
	smartTerminal bool
	noNewTerminal bool
	openInAgent   bool
	noOpenInAgent bool

	// other disallowed tokens / positionals for mutual exclusion messages
	conflict string
}

// runSetConfig handles wrk --set-config [...].
func runSetConfig(origWd string, args []string, ctx *invocationContext) error {
	opts, err := parseSetConfigArgs(args)
	if err != nil {
		return err
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	ctx.wrkHome = wrkHome
	ctx.workDir = origWd
	ctx.command = "set-config"
	ctx.eventArgs = extractEventArgs(args, nil)
	if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
		return err
	}
	if err := ctx.autoRecord(); err != nil {
		return err
	}

	if opts.conflict != "" {
		return fmt.Errorf("wrk: --set-config is mutually exclusive with other modes")
	}

	if opts.show {
		if opts.create || opts.anyCreateFlag() {
			return fmt.Errorf("wrk: --set-config --show is mutually exclusive with --create flags")
		}
		return setConfigShow(wrkHome)
	}

	if !opts.create {
		return fmt.Errorf("wrk: --set-config requires --create or --show")
	}

	// Validate create UX flag conflicts under set-config.
	f := createUXFlags{
		newWindow:     opts.newWindow,
		noNewWindow:   opts.noNewWindow,
		newTerminal:   opts.newTerminal,
		reuseTerminal: opts.reuseTerminal,
		smartTerminal: opts.smartTerminal,
		noNewTerminal: opts.noNewTerminal,
		openInAgent:   opts.openInAgent,
		noOpenInAgent: opts.noOpenInAgent,
	}
	if err := f.validate(); err != nil {
		return err
	}
	if !f.any() {
		return fmt.Errorf("wrk: --set-config --create requires at least one create UX flag")
	}

	return setConfigWriteCreate(wrkHome, f)
}

func (o setConfigOpts) anyCreateFlag() bool {
	return o.newWindow || o.noNewWindow ||
		o.newTerminal || o.reuseTerminal || o.smartTerminal || o.noNewTerminal ||
		o.openInAgent || o.noOpenInAgent
}

func parseSetConfigArgs(args []string) (setConfigOpts, error) {
	var opts setConfigOpts
	sawSetConfig := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--set-config":
			if sawSetConfig {
				return opts, fmt.Errorf("wrk: duplicate --set-config")
			}
			sawSetConfig = true
		case "--show":
			opts.show = true
		case "--create":
			opts.create = true
		case "--new-window":
			opts.newWindow = true
		case "--no-new-window":
			opts.noNewWindow = true
		case "--new-terminal":
			opts.newTerminal = true
		case "--reuse-terminal":
			opts.reuseTerminal = true
		case "--smart-terminal":
			opts.smartTerminal = true
		case "--no-new-terminal":
			opts.noNewTerminal = true
		case "--open-in-agent":
			opts.openInAgent = true
		case "--no-open-in-agent":
			opts.noOpenInAgent = true
		case "--":
			// anything after is positional / conflict
			if i+1 < len(args) {
				opts.conflict = args[i+1]
			} else {
				opts.conflict = "--"
			}
			return opts, nil
		default:
			if isSetConfigDisallowed(arg) {
				opts.conflict = arg
				return opts, nil
			}
			// String flags with values that belong to other modes.
			if arg == "--add" || arg == "--rm" || arg == "--where" || arg == "--dep" ||
				arg == "--task" || arg == "-t" || arg == "--set-task" {
				opts.conflict = arg
				return opts, nil
			}
			if strings.HasPrefix(arg, "-") {
				return opts, fmt.Errorf("wrk: unrecognized flag: %s", arg)
			}
			// Positional (e.g. create dir) is mutually exclusive.
			opts.conflict = arg
			return opts, nil
		}
	}
	if !sawSetConfig {
		return opts, fmt.Errorf("wrk: --set-config required")
	}
	return opts, nil
}

func isSetConfigDisallowed(arg string) bool {
	for _, d := range setConfigDisallowedFlags {
		if arg == d {
			return true
		}
	}
	return false
}

func setConfigShow(wrkHome string) error {
	path := filepath.Join(wrkHome, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Empty object is valid JSON for show when no config yet.
			fmt.Println("{}")
			return nil
		}
		return fmt.Errorf("wrk: read config.json: %w", err)
	}
	// Pretty-print for stable readability; accept already-pretty input.
	var any interface{}
	if err := json.Unmarshal(data, &any); err != nil {
		return fmt.Errorf("wrk: parse config.json: %w", err)
	}
	out, err := json.MarshalIndent(any, "", "  ")
	if err != nil {
		return fmt.Errorf("wrk: marshal config.json: %w", err)
	}
	fmt.Println(string(out))
	return nil
}

func setConfigWriteCreate(wrkHome string, f createUXFlags) error {
	root, err := loadConfigMap(wrkHome)
	if err != nil {
		return err
	}
	if root == nil {
		root = map[string]interface{}{}
	}
	if _, ok := root["version"]; !ok {
		root["version"] = 1
	}

	createMap, err := ensureCreateMap(root)
	if err != nil {
		return err
	}

	// Window
	if f.noNewWindow {
		delete(createMap, "window")
	}
	if f.newWindow {
		createMap["window"] = map[string]interface{}{"mode": "new"}
		// Implication: also persist terminal.mode=new unless a terminal flag sets otherwise.
		if !f.newTerminal && !f.reuseTerminal && !f.smartTerminal && !f.noNewTerminal {
			createMap["terminal"] = map[string]interface{}{"mode": "new"}
		}
	}

	// Terminal
	if f.noNewTerminal {
		delete(createMap, "terminal")
	}
	if f.newTerminal {
		createMap["terminal"] = map[string]interface{}{"mode": "new"}
	}
	if f.reuseTerminal {
		createMap["terminal"] = map[string]interface{}{"mode": "reuse"}
	}
	if f.smartTerminal {
		createMap["terminal"] = map[string]interface{}{"mode": "smart"}
	}

	// Agent
	if f.noOpenInAgent {
		agent := existingAgentMap(createMap)
		agent["enabled"] = false
		// Keep runner/template/args if present; ensure enabled is explicit.
		createMap["agent"] = agent
	}
	if f.openInAgent {
		agent := existingAgentMap(createMap)
		agent["enabled"] = true
		if _, ok := agent["runner"]; !ok || agent["runner"] == "" {
			agent["runner"] = defaultAgentRunner
		}
		if _, ok := agent["prompt_template"]; !ok || agent["prompt_template"] == "" {
			agent["prompt_template"] = defaultAgentPromptTemplate
		}
		if _, ok := agent["args"]; !ok {
			agent["args"] = defaultAgentArgs()
		}
		createMap["agent"] = agent
	}

	return saveConfigMap(wrkHome, root)
}

func existingAgentMap(createMap map[string]interface{}) map[string]interface{} {
	if v, ok := createMap["agent"]; ok && v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			// Shallow copy so we don't mutate unexpected shared maps.
			out := make(map[string]interface{}, len(m)+4)
			for k, val := range m {
				out[k] = val
			}
			return out
		}
	}
	return map[string]interface{}{}
}

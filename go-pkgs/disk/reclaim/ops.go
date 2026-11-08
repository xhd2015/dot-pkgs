package reclaim

import "strings"

// OpKind classifies one cleanup action. The kind drives both the display
// string and the badge in the ai-critic report, so the class list can never
// describe something the step does not do.
type OpKind string

const (
	// OpRemoveDir deletes a directory tree: rendered as `rm -rf '<path>'`.
	OpRemoveDir OpKind = "rm-rf"
	// OpRemoveFiles deletes files matching a pattern: rendered as `rm <pat>`.
	OpRemoveFiles OpKind = "rm"
	// OpExec delegates to an official tool command: `go clean -cache`.
	OpExec OpKind = "exec"
	// OpNone touches nothing (inventory-only rows such as ShipIt).
	OpNone OpKind = "none"
)

// Op is one action a cleanup step performs, described declaratively.
//
// Apply is still the executor; Ops is the audit/display projection of the same
// intent. TestStepDisplayMatchesApply keeps the two from drifting: for every
// step, the tmp-home fixture asserts that exactly the displayed paths vanish.
type Op struct {
	Kind  OpKind   `json:"kind"`
	Paths []string `json:"paths,omitempty"`
	// Argv is the official command for OpExec, one token per element.
	Argv []string `json:"argv,omitempty"`
	// Sudo marks an action that needs root; the UI shows a sudo badge and
	// never runs it unless the user selected it explicitly.
	Sudo bool `json:"sudo,omitempty"`
	// Text overrides the rendered command for steps whose shell line is more
	// specific than the canonical rendering (e.g. chflags + rm).
	Text string `json:"text,omitempty"`
	// Skip qualifies the display when the action filters targets, e.g.
	// "（跳过正在运行的）".
	Skip string `json:"skip,omitempty"`
	// Hidden marks an action that runs as part of the step but is not the
	// headline the user sees first.
	Hidden bool `json:"hidden,omitempty"`
}

// Display renders the op as the shell line a user could run by hand.
func (o Op) Display() string {
	var s string
	switch {
	case o.Text != "":
		s = o.Text
	case o.Kind == OpRemoveDir && len(o.Paths) > 0:
		s = "rm -rf " + shQuote(o.Paths[0])
	case o.Kind == OpRemoveFiles && len(o.Paths) > 0:
		s = "rm " + o.Paths[0]
	case o.Kind == OpExec && len(o.Argv) > 0:
		s = strings.Join(o.Argv, " ")
		if o.Sudo {
			s = "sudo " + s
		}
	}
	return s + o.Skip
}

// headlineOps returns the display command and target paths for a step: the
// first op that is not Hidden.
func headlineOps(ops []Op) (string, []string) {
	for _, op := range ops {
		if op.Hidden {
			continue
		}
		return op.Display(), op.Paths
	}
	return "", nil
}

// stepOps is the display/audit projection of every cleanup step. It mirrors
// stepMetas (which keeps the legacy summary/notes) and TestStepDisplayMatchesApply
// pins both to the tmp-home fixture so a display can never promise more than
// Apply removes.
func stepOps(opts Options) map[string][]Op {
	h := opts.Home
	return map[string][]Op{
		"sandbox-gocache":      {dirOp(j(h, ".sandbox", "gocache"))},
		"codex-logs":           {filesOp(j(h, ".codex", "logs_2.sqlite*"))},
		"home-logs":            {{Kind: OpRemoveFiles, Paths: []string{j(h, "ai-critic-server.log"), j(h, "ai-critic-server-keep-alive.log")}, Text: "rm " + j(h, "ai-critic-server*.log")}},
		"brew-cleanup":         {execOp([]string{"brew", "cleanup", "-s", "--prune=all"}, j(h, "Library", "Caches", "Homebrew")), {Kind: OpRemoveDir, Paths: []string{j(h, "Library", "Caches", "Homebrew")}, Hidden: true}},
		"npm-cache":            {execOp([]string{"npm", "cache", "clean", "--force"}, j(h, ".npm", "_cacache"))},
		"remote-agent-logs":    {filesOp(j(h, "Library", "Caches", "remote-agent", "*.log"))},
		"podman-applehv-cache": {dirOp(j(h, ".local", "share", "containers", "podman", "machine", "applehv", "cache"))},
		"iterm-bak-apps":       {{Kind: OpRemoveDir, Paths: []string{j(h, "Applications", "iTerm*.app.bak*"), "/Applications/iTerm*.app.bak*"}, Text: "rm -rf <bak .app>", Skip: "（跳过正在运行的）"}},
		"go-build-cache":       {execOp([]string{"go", "clean", "-cache"}, j(h, "Library", "Caches", "go-build"))},
		"pnpm-store":           {execOp([]string{"pnpm", "store", "prune"}, j(h, "Library", "pnpm", "store"))},
		"npm-npx":              {dirOp(j(h, ".npm", "_npx"))},
		"bun-cache":            {dirOp(j(h, ".bun", "install", "cache"))},
		"codex-releases":       {{Kind: OpRemoveDir, Paths: []string{j(h, ".codex", "packages", "standalone", "releases")}, Text: "rm -rf <旧 release 目录>"}},
		"xcache":               {dirOp(j(h, ".cache"))},
		"ms-playwright":        {dirOp(msPlaywrightPath(h))},
		"opencode-snapshot":    {dirOp(openCodeSnapshotPath(h))},
		"ios-sim-runtimes":     {execOp([]string{"xcrun", "simctl", "runtime", "delete", "all"}, "xcrun simctl runtime list")},
		"ios-sim-mobileasset":  {mobileAssetOp(opts)},
		"dirs-cleaner":         {{Kind: OpExec, Paths: []string{dirsCleanerRoot(opts)}, Sudo: true, Text: DirsCleanerPurgeCommand(opts)}},
		"shipit":               {{Kind: OpNone, Paths: []string{j(h, "Library", "Caches", "*.ShipIt")}, Text: "（无；在对应应用里检查更新即可释放）"}},
	}
}

func dirOp(path string) Op { return Op{Kind: OpRemoveDir, Paths: []string{path}} }

func filesOp(pattern string) Op { return Op{Kind: OpRemoveFiles, Paths: []string{pattern}} }

func execOp(argv []string, paths ...string) Op {
	return Op{Kind: OpExec, Argv: argv, Paths: paths}
}

// mobileAssetOp describes the SIP-blocked MobileAsset delete: chflags first,
// then rm, both as root.
func mobileAssetOp(opts Options) Op {
	root, err := resolveIOSSimAssetRoot(opts)
	if err != nil || root == "" {
		root = defaultIOSSimAssetRoot
	}
	return Op{
		Kind:  OpRemoveDir,
		Paths: []string{root},
		Sudo:  true,
		Text:  "sudo chflags -R norestricted … && sudo rm -rf " + shQuote(root),
	}
}

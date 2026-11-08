package reclaim

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Plan groups.
const (
	GroupDefault    = "default"
	GroupAggressive = "aggressive"
	GroupSystem     = "system"
	GroupInventory  = "inventory"
)

// Plan is one measured view of every step — shared by `--dry-run --json` and
// the web UI so both audit the same numbers the CLI acts on.
type Plan struct {
	ScannedAt   string            `json:"scannedAt"`
	ScanMillis  int64             `json:"scanMillis"`
	Home        string            `json:"home"`
	Volume      *VolumeInfo       `json:"volume,omitempty"`
	Steps       []PlanStep        `json:"steps"`
	DirsCleaner DirsCleanerStatus `json:"dirsCleaner"`
}

// PlanStep is one selectable (or inventory-only) row.
type PlanStep struct {
	ID      string   `json:"id"`
	Group   string   `json:"group"`
	Summary string   `json:"summary"`
	Paths   []string `json:"paths"`
	Command string   `json:"command"`
	// Display is the typed action list Command is rendered from. It carries
	// the badge (rm -rf / official tool / needs root) the report UI shows.
	Display []Op   `json:"display,omitempty"`
	Flag    string `json:"flag,omitempty"`
	// SizeBytes is nil when the size needs admin to measure.
	SizeBytes  *int64   `json:"sizeBytes"`
	Selected   bool     `json:"selected"` // would run under the given options
	NeedsAdmin bool     `json:"needsAdmin,omitempty"`
	ReadOnly   bool     `json:"readOnly,omitempty"` // inventory only, never deleted
	Details    []string `json:"details,omitempty"`
	Notes      []string `json:"notes,omitempty"`
}

// VolumeInfo is the data volume holding Home.
type VolumeInfo struct {
	Path       string `json:"path"`
	TotalBytes int64  `json:"totalBytes"`
	FreeBytes  int64  `json:"freeBytes"`
}

type stepMeta struct {
	summary string
	paths   func(home string) []string
	command func(home string) string
	notes   []string
}

func j(home string, parts ...string) string {
	return filepath.Join(append([]string{home}, parts...)...)
}

func one(f func(home string) string) func(string) []string {
	return func(home string) []string { return []string{f(home)} }
}

func rmrf(f func(home string) string) func(string) string {
	return func(home string) string { return "rm -rf " + shQuote(f(home)) }
}

var stepMetas = map[string]stepMeta{
	"sandbox-gocache": {
		summary: "沙箱 Go 构建缓存",
		paths:   func(h string) []string { return []string{j(h, ".sandbox", "gocache")} },
		command: func(h string) string { return "rm -rf " + shQuote(j(h, ".sandbox", "gocache")) },
	},
	"codex-logs": {
		summary: "Codex 日志数据库",
		paths:   func(h string) []string { return []string{j(h, ".codex", "logs_2.sqlite*")} },
		command: func(h string) string { return "rm " + j(h, ".codex", "logs_2.sqlite*") },
		notes:   []string{"只删 logs_2.sqlite*；.codex 下其他文件保留"},
	},
	"home-logs": {
		summary: "ai-critic 服务日志",
		paths: func(h string) []string {
			return []string{j(h, "ai-critic-server.log"), j(h, "ai-critic-server-keep-alive.log")}
		},
		command: func(h string) string { return "rm " + j(h, "ai-critic-server*.log") },
	},
	"brew-cleanup": {
		summary: "Homebrew 下载缓存",
		paths:   func(h string) []string { return []string{j(h, "Library", "Caches", "Homebrew")} },
		command: func(h string) string { return "brew cleanup -s --prune=all" },
	},
	"npm-cache": {
		summary: "npm 缓存",
		paths:   func(h string) []string { return []string{j(h, ".npm", "_cacache")} },
		command: func(h string) string { return "npm cache clean --force" },
	},
	"remote-agent-logs": {
		summary: "remote-agent 日志",
		paths:   func(h string) []string { return []string{j(h, "Library", "Caches", "remote-agent", "*.log")} },
		command: func(h string) string { return "rm " + j(h, "Library", "Caches", "remote-agent", "*.log") },
	},
	"podman-applehv-cache": {
		summary: "Podman 虚拟机镜像下载缓存",
		paths: func(h string) []string {
			return []string{j(h, ".local", "share", "containers", "podman", "machine", "applehv", "cache")}
		},
		command: func(h string) string {
			return "rm -rf " + shQuote(j(h, ".local", "share", "containers", "podman", "machine", "applehv", "cache"))
		},
		notes: []string{"同级的 *.raw 虚拟机磁盘不受影响；更多见 clear-tmp podman how-to"},
	},
	stepITermBakApps: {
		summary: "iTerm 备份 .app",
		paths: func(h string) []string {
			return []string{j(h, "Applications", "iTerm*.app.bak*"), "/Applications/iTerm*.app.bak*"}
		},
		command: func(h string) string { return "rm -rf <bak .app>（跳过正在运行的）" },
	},
	"go-build-cache": {
		summary: "Go 编译缓存（可重建；go mod 缓存永不清理）",
		paths:   one(goBuildCachePath),
		command: func(h string) string { return "go clean -cache" },
	},
	"pnpm-store": {
		summary: "pnpm store 中未被引用的包",
		paths:   one(pnpmStorePath),
		command: func(h string) string { return "pnpm store prune" },
	},
	"npm-npx":   {summary: "npx 一次性包缓存", paths: one(npmNpxPath), command: rmrf(npmNpxPath)},
	"bun-cache": {summary: "bun install 缓存", paths: one(bunCachePath), command: rmrf(bunCachePath)},
	"codex-releases": {
		summary: "旧版本 codex 运行时（保留最新）",
		paths:   one(codexReleasesDir),
		command: func(h string) string { return "rm -rf <旧 release 目录>" },
	},
	"xcache": {
		summary: "整个 ~/.cache（仅重新下载成本）",
		paths:   one(xcachePath),
		command: rmrf(xcachePath),
	},
	stepMSPlaywright:     {summary: "Playwright 浏览器", paths: one(msPlaywrightPath), command: rmrf(msPlaywrightPath)},
	stepOpenCodeSnapshot: {summary: "OpenCode 快照", paths: one(openCodeSnapshotPath), command: rmrf(openCodeSnapshotPath)},
}

func stepFlag(id string, tier Tier) string {
	switch id {
	case stepMSPlaywright:
		return "--ms-playwright"
	case stepOpenCodeSnapshot:
		return "--opencode-snapshot"
	case stepIOSSimRuntimes:
		return "--ios-sim"
	case stepIOSSimMobileAsst:
		return "--ios-sim-assets"
	case stepDirsCleaner:
		return "--dirs-cleaner"
	}
	if tier == TierAggressive {
		return "--include " + id
	}
	return ""
}

// Scan measures every step without changing anything. Selection fields in opts
// (Only / Include / Include* / Aggressive) only drive PlanStep.Selected.
func Scan(opts Options) (*Plan, error) {
	return scanPlan(opts, nil)
}

// StepInfo is the public registry metadata: what a host UI needs to list every
// cleanup class without measuring any of them.
type StepInfo struct {
	ID         string `json:"id"`
	Group      string `json:"group"`
	Summary    string `json:"summary"`
	NeedsAdmin bool   `json:"needsAdmin,omitempty"`
	ReadOnly   bool   `json:"readOnly,omitempty"`
	// Actions is the display projection of the step (same ops a dry-run JSON
	// reports), so a UI can render `rm -rf ...` / `go clean -cache` without
	// running a scan.
	Actions []Op     `json:"actions"`
	Paths   []string `json:"paths,omitempty"`
}

// StepInfos returns every cleanup class in display order: default, aggressive,
// system (admin), then the read-only inventory rows. No filesystem work beyond
// resolving paths.
func StepInfos(home string) []StepInfo {
	opts := Options{Home: home}
	ops := stepOps(opts)
	infos := make([]StepInfo, 0, 24)
	add := func(info StepInfo) {
		info.Actions = ops[info.ID]
		if info.Paths == nil {
			info.Paths = opPaths(info.Actions)
		}
		infos = append(infos, info)
	}
	for _, def := range allSteps() {
		group := GroupDefault
		if def.Tier == TierAggressive {
			group = GroupAggressive
		}
		add(StepInfo{ID: def.ID, Group: group, Summary: stepMetas[def.ID].summary})
	}
	add(StepInfo{ID: stepIOSSimRuntimes, Group: GroupSystem, Summary: "iOS 模拟器运行时（全部）", NeedsAdmin: true})
	add(StepInfo{ID: stepIOSSimMobileAsst, Group: GroupSystem, Summary: "iOS 模拟器 MobileAsset 缓存（SIP 常拦截）", NeedsAdmin: true})
	add(StepInfo{ID: stepDirsCleaner, Group: GroupSystem, Summary: "系统临时目录回收站（开机搬入的 /tmp 与 $TMPDIR）", NeedsAdmin: true})
	add(StepInfo{ID: "shipit", Group: GroupInventory, Summary: "应用自动更新缓存（*.ShipIt）— 只列出，永不删除", ReadOnly: true})
	add(StepInfo{ID: "go-mod-cache", Group: GroupInventory, Summary: "Go 模块缓存 — 只列出，永不删除（清了要重新下载全部模块）", ReadOnly: true})
	for i := range infos {
		if !infos[i].NeedsAdmin && infos[i].Group != GroupInventory {
			for _, op := range infos[i].Actions {
				if op.Sudo {
					infos[i].NeedsAdmin = true
				}
			}
		}
	}
	return infos
}

// opPaths collects the concrete paths of a step's ops (deduped, in order).
func opPaths(ops []Op) []string {
	var out []string
	seen := map[string]bool{}
	for _, op := range ops {
		for _, p := range op.Paths {
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// ScanSteps measures only the requested step IDs. The system tail (xcrun
// runtimes, MobileAsset, dirs_cleaner) and ShipIt are measured only when their
// IDs are requested, so a host UI can size the classes it shows without
// shelling out to xcrun or reading the unified log.
//
// Step order follows the registry, so ScanSteps(ids) is a subset of Scan().
func ScanSteps(opts Options, ids []string) (*Plan, error) {
	ids = splitIDList(ids)
	if len(ids) == 0 {
		return nil, errors.New("ScanSteps: no step ids")
	}
	known := scanStepIDs()
	for _, id := range ids {
		if !containsString(known, id) {
			return nil, fmt.Errorf("ScanSteps: unknown step %q (known: %s)", id, strings.Join(known, ", "))
		}
	}
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	return scanPlan(opts, want)
}

// scanPlan is Scan with an optional step filter. want == nil means all steps.
func scanPlan(opts Options, want map[string]bool) (*Plan, error) {
	if opts.Home == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		opts.Home = home
	}
	opts, err := applySelection(opts)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	plan := &Plan{Home: opts.Home, Steps: []PlanStep{}}
	ops := stepOps(opts)

	for _, def := range allSteps() {
		if want != nil && !want[def.ID] {
			continue
		}
		meta := stepMetas[def.ID]
		size := def.Measure(opts)
		group := GroupDefault
		if def.Tier == TierAggressive {
			group = GroupAggressive
		}
		ps := PlanStep{
			ID:        def.ID,
			Group:     group,
			Summary:   meta.summary,
			Flag:      stepFlag(def.ID, def.Tier),
			SizeBytes: &size,
			Selected:  stepEnabled(opts, def),
			Notes:     meta.notes,
		}
		ps.Display = ops[def.ID]
		if cmd, paths := headlineOps(ps.Display); cmd != "" {
			ps.Command = cmd
			if len(paths) > 0 {
				ps.Paths = paths
			}
		} else {
			if meta.paths != nil {
				ps.Paths = meta.paths(opts.Home)
			}
			if meta.command != nil {
				ps.Command = meta.command(opts.Home)
			}
		}
		switch def.ID {
		case stepITermBakApps:
			for _, e := range listITermBakApps(opts).Entries {
				d := e.Path + "  " + humanSize(e.Size)
				if e.Running {
					d += "  （运行中，跳过）"
				}
				ps.Details = append(ps.Details, d)
			}
		case "codex-releases":
			for _, p := range staleCodexReleases(opts.Home) {
				ps.Details = append(ps.Details, p)
			}
		case "codex-logs":
			matches, _ := filepath.Glob(filepath.Join(opts.Home, ".codex", "logs_2.sqlite*"))
			ps.Details = append(ps.Details, matches...)
		}
		plan.Steps = append(plan.Steps, ps)
	}

	if want == nil || want[stepIOSSimRuntimes] {
		// iOS Simulator runtimes.
		inv := measureIOSSim(opts)
		iosStep := PlanStep{
			ID: stepIOSSimRuntimes, Group: GroupSystem, Summary: "iOS 模拟器运行时（全部）",
			Display:  ops[stepIOSSimRuntimes],
			Flag:     stepFlag(stepIOSSimRuntimes, TierDefault),
			Selected: opts.IncludeIOSSim,
		}
		iosStep.Command, iosStep.Paths = headlineOps(iosStep.Display)
		iosSize := inv.TotalSize
		iosStep.SizeBytes = &iosSize
		switch {
		case inv.ToolMissing:
			iosStep.Notes = append(iosStep.Notes, "未找到 xcrun")
		case inv.Err != nil:
			iosStep.Notes = append(iosStep.Notes, "simctl 失败: "+inv.Err.Error())
		}
		for _, r := range inv.Runtimes {
			iosStep.Details = append(iosStep.Details, r.Name+" ("+r.Build+")  "+humanSize(r.SizeBytes))
		}
		plan.Steps = append(plan.Steps, iosStep)
	}

	if want == nil || want[stepIOSSimMobileAsst] {
		// MobileAsset (sudo; SIP usually blocks).
		assetRoot, assetSize, assetErr := measureIOSSimAssets(opts)
		assetStep := PlanStep{
			ID: stepIOSSimMobileAsst, Group: GroupSystem, Summary: "iOS 模拟器 MobileAsset 缓存（SIP 常拦截）",
			Display:    ops[stepIOSSimMobileAsst],
			Flag:       stepFlag(stepIOSSimMobileAsst, TierDefault),
			SizeBytes:  &assetSize,
			Selected:   opts.IncludeIOSSimAssets,
			NeedsAdmin: true,
			Notes:      []string{"SIP 拦截时：恢复模式 → csrutil disable → rm → csrutil enable"},
		}
		assetStep.Command, assetStep.Paths = headlineOps(assetStep.Display)
		if assetRoot != "" {
			assetStep.Paths = []string{assetRoot}
		}
		if assetErr != nil {
			assetStep.Notes = append(assetStep.Notes, assetErr.Error())
		}
		plan.Steps = append(plan.Steps, assetStep)
	}

	if want == nil || want[stepDirsCleaner] {
		// dirs_cleaner.
		dc := DirsCleanerBasic(opts)
		plan.DirsCleaner = dc
		dcStep := PlanStep{
			ID: stepDirsCleaner, Group: GroupSystem, Summary: "系统临时目录回收站（开机搬入的 /tmp 与 $TMPDIR）",
			Display:    ops[stepDirsCleaner],
			Flag:       stepFlag(stepDirsCleaner, TierDefault),
			SizeBytes:  dc.SizeBytes,
			Selected:   opts.IncludeDirsCleaner,
			NeedsAdmin: true,
			Notes:      []string{"保留目录本身，只删内容；诊断见 dirs_cleaner 标签页 / clear-tmp dirs-cleaner status"},
		}
		dcStep.Command, dcStep.Paths = headlineOps(dcStep.Display)
		dcStep.Paths = []string{dc.Root}
		plan.Steps = append(plan.Steps, dcStep)
	}

	if want == nil || want["shipit"] {
		// ShipIt: inventory only.
		ship := measureShipIt(opts.Home)
		shipStep := PlanStep{
			ID: "shipit", Group: GroupInventory, Summary: "应用自动更新缓存（*.ShipIt）— 只列出，永不删除",
			Display:   ops["shipit"],
			SizeBytes: &ship.Total,
			ReadOnly:  true,
		}
		shipStep.Command, shipStep.Paths = headlineOps(shipStep.Display)
		for _, e := range ship.Entries {
			shipStep.Details = append(shipStep.Details, e.Name+"  "+humanSize(e.Size)+"  → "+e.Hint)
		}
		plan.Steps = append(plan.Steps, shipStep)
	}

	if want == nil || want["go-mod-cache"] {
		// Go module cache: inventory only. Removing it forces a network
		// re-fetch of every module, which costs far more than the space.
		size := measureGoModCache(opts)
		modStep := PlanStep{
			ID: "go-mod-cache", Group: GroupInventory, Summary: "Go 模块缓存 — 只列出，永不删除（清了要重新下载全部模块）",
			Display:   ops["go-mod-cache"],
			SizeBytes: &size,
			ReadOnly:  true,
			Notes:     []string{"清了要重新下载全部模块；clear-tmp 不清理它"},
		}
		modStep.Command, modStep.Paths = headlineOps(modStep.Display)
		plan.Steps = append(plan.Steps, modStep)
	}

	if v, ok := volumeInfo(opts.Home); ok {
		plan.Volume = &v
	}
	plan.ScannedAt = formatTime(start)
	plan.ScanMillis = time.Since(start).Milliseconds()
	return plan, nil
}

func volumeInfo(path string) (VolumeInfo, bool) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return VolumeInfo{}, false
	}
	bs := int64(st.Bsize)
	return VolumeInfo{
		Path:       path,
		TotalBytes: int64(st.Blocks) * bs,
		FreeBytes:  int64(st.Bavail) * bs,
	}, true
}

func writeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

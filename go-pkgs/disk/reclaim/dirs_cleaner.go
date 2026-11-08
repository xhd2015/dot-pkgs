package reclaim

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/disk/elevate"
)

// macOS moves /tmp and every user's $TMPDIR into dirs_cleaner at boot
// (/usr/libexec/dirs_cleaner, RENAME_SWAP) and CacheDelete's deleted_helper
// purges it later. A path longer than PATH_MAX makes that purge fail with
// "File name too long", so every boot's temp tree piles up there.
const (
	defaultDirsCleanerRoot = "/private/var/dirs_cleaner"
	pathMax                = 1024
	longPathWarn           = 900
)

// DirsCleanerStatus is the read-only state of /private/var/dirs_cleaner.
type DirsCleanerStatus struct {
	Root     string `json:"root"`
	Exists   bool   `json:"exists"`
	Readable bool   `json:"readable"`
	// Batches is the number of swapped-in temp trees (nlink-2 when unreadable).
	Batches int    `json:"batches"`
	ModTime string `json:"modTime,omitempty"`
	// SizeBytes is nil when the directory needs admin to measure.
	SizeBytes *int64               `json:"sizeBytes,omitempty"`
	Purge     *PurgeHealth         `json:"purge,omitempty"`
	TempDirs  []TempDirProbe       `json:"tempDirs,omitempty"`
	Analysis  *DirsCleanerAnalysis `json:"analysis,omitempty"`
}

// PurgeHealth summarises deleted_helper purge attempts from the unified log.
type PurgeHealth struct {
	Window       string     `json:"window"`
	Attempts     int        `json:"attempts"`
	Failures     int        `json:"failures"`
	FirstFailure string     `json:"firstFailure,omitempty"`
	LastAttempt  string     `json:"lastAttempt,omitempty"`
	LastOK       bool       `json:"lastOk"`
	LastError    string     `json:"lastError,omitempty"`
	Stuck        bool       `json:"stuck"` // last attempt failed with File name too long
	Events       []LogEvent `json:"events"`
	Err          string     `json:"error,omitempty"`
}

// LogEvent is one relevant unified-log line (chronological).
type LogEvent struct {
	Time    string `json:"time"`
	Kind    string `json:"kind"` // attempt | boot | sandbox
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// TempDirProbe measures a temp root that the next boot moves into dirs_cleaner.
type TempDirProbe struct {
	Label       string `json:"label"`
	Path        string `json:"path"`
	SizeBytes   int64  `json:"sizeBytes"`
	LongestLen  int    `json:"longestLen"`
	LongestPath string `json:"longestPath,omitempty"`
	OverWarn    bool   `json:"overWarn"`
}

// DirsCleanerAnalysis is the admin (read-only) per-batch breakdown.
type DirsCleanerAnalysis struct {
	AnalysedAt      string             `json:"analysedAt"`
	TotalBytes      int64              `json:"totalBytes"`
	RestrictedFiles int                `json:"restrictedFiles"`
	Batches         []DirsCleanerBatch `json:"batches"`
	Offenders       []PathOffender     `json:"offenders"`
}

// DirsCleanerBatch is one top-level entry under dirs_cleaner.
type DirsCleanerBatch struct {
	Name       string `json:"name"`
	ModTime    string `json:"modTime,omitempty"`
	SizeBytes  int64  `json:"sizeBytes"`
	MaxPathLen int    `json:"maxPathLen"`
	Restricted int    `json:"restricted"`
}

// PathOffender is a path long enough to stall the system purge.
type PathOffender struct {
	Batch   string `json:"batch"`
	Path    string `json:"path"` // relative to the dirs_cleaner root
	Len     int    `json:"len"`  // absolute length
	Pattern string `json:"pattern,omitempty"`
	Repeats int    `json:"repeats,omitempty"`
}

// PurgeOutcome is the result of the privileged purge.
type PurgeOutcome struct {
	Before     int64    `json:"before"`
	After      int64    `json:"after"`
	Left       int      `json:"left"`
	Restricted []string `json:"restricted,omitempty"`
}

func dirsCleanerRoot(opts Options) string {
	if opts.DirsCleanerRoot != "" {
		return opts.DirsCleanerRoot
	}
	return defaultDirsCleanerRoot
}

// DirsCleanerBasic reads metadata only (no log query, no temp walk).
func DirsCleanerBasic(opts Options) DirsCleanerStatus {
	root := dirsCleanerRoot(opts)
	st := DirsCleanerStatus{Root: root}
	info, err := os.Stat(root)
	if err != nil {
		return st
	}
	st.Exists = true
	st.ModTime = formatTime(info.ModTime())
	if entries, err := os.ReadDir(root); err == nil {
		st.Readable = true
		st.Batches = len(entries)
		size := pathSize(root)
		st.SizeBytes = &size
		return st
	}
	if sys, ok := info.Sys().(*syscall.Stat_t); ok && int(sys.Nlink) >= 2 {
		st.Batches = int(sys.Nlink) - 2
	}
	return st
}

// DirsCleanerFullStatus adds purge health (log show) and temp-dir probes.
// since: "" or "boot" → since last boot; otherwise a log(1) --last value (1h, 24h, 7d).
func DirsCleanerFullStatus(opts Options, since string) DirsCleanerStatus {
	st := DirsCleanerBasic(opts)
	done := make(chan PurgeHealth, 1)
	go func() { done <- queryPurgeHealth(opts, since) }()
	st.TempDirs = probeTempDirs(opts)
	ph := <-done
	st.Purge = &ph
	return st
}

// Only deleted_helper: adding kernel/dirs_cleaner (boot-task + sandbox
// context lines) makes log(1) evaluate every kernel message and turned a
// ~3s query into minutes under load.
const purgeLogPredicate = `process == "deleted_helper" AND eventMessage CONTAINS "dirs_cleaner"`

func logShowArgs(since string) []string {
	window := since
	if window == "" {
		window = "boot"
	}
	return []string{"show", "--last", window, "--style", "ndjson", "--predicate", purgeLogPredicate}
}

func queryPurgeHealth(opts Options, since string) PurgeHealth {
	args := logShowArgs(since)
	window := since
	if window == "" {
		window = "boot"
	}
	var out []byte
	var err error
	if opts.LogShow != nil {
		out, err = opts.LogShow(args)
	} else {
		out, err = exec.Command("log", args...).Output()
	}
	if err != nil && len(out) == 0 {
		return PurgeHealth{Window: window, Err: fmt.Sprintf("log show failed: %v", err)}
	}
	ph := parsePurgeLog(out)
	ph.Window = window
	return ph
}

type logLine struct {
	Timestamp string `json:"timestamp"`
	PID       int    `json:"processID"`
	Message   string `json:"eventMessage"`
}

// parsePurgeLog classifies `log show --style ndjson` lines.
//
// An attempt is "purge_orphans … clearing: …/dirs_cleaner/"; a following
// "nuke_dir: removefile error for …/dirs_cleaner/ : <err>" from the same pid
// marks it failed. Boot-task and sandbox-deny lines are kept for context.
func parsePurgeLog(data []byte) PurgeHealth {
	ph := PurgeHealth{Events: []LogEvent{}}
	openAttempt := map[int]int{} // pid → index into Events
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var l logLine
		if json.Unmarshal(line, &l) != nil || l.Message == "" {
			continue
		}
		ts := parseLogTime(l.Timestamp)
		msg := l.Message
		switch {
		case strings.Contains(msg, "purge_orphans") && strings.Contains(msg, "dirs_cleaner"):
			ph.Events = append(ph.Events, LogEvent{Time: ts, Kind: "attempt", Message: strings.TrimSpace(trimBefore(msg, "purge_orphans"))})
			openAttempt[l.PID] = len(ph.Events) - 1
		case strings.HasPrefix(msg, "nuke_dir: removefile error") && strings.Contains(msg, "dirs_cleaner"):
			errText := msg
			if i := strings.LastIndex(msg, " : "); i >= 0 {
				errText = strings.TrimSpace(msg[i+3:])
			}
			if idx, ok := openAttempt[l.PID]; ok && ph.Events[idx].Error == "" {
				ph.Events[idx].Error = errText
				delete(openAttempt, l.PID)
			} else {
				ph.Events = append(ph.Events, LogEvent{Time: ts, Kind: "attempt", Message: "removefile", Error: errText})
			}
		case strings.Contains(msg, "[Event: dirs_cleaner") && strings.Contains(msg, "boot-task") || strings.Contains(msg, "[Event: dirs_cleaner") && strings.Contains(msg, "boot task"):
			ph.Events = append(ph.Events, LogEvent{Time: ts, Kind: "boot", Message: strings.TrimSpace(trimBefore(msg, "[Event: "))})
		case strings.Contains(msg, "RENAME_SWAP") && strings.Contains(msg, "denied"):
			ph.Events = append(ph.Events, LogEvent{Time: ts, Kind: "sandbox", Message: strings.TrimSpace(msg)})
		}
	}
	for _, e := range ph.Events {
		if e.Kind != "attempt" {
			continue
		}
		ph.Attempts++
		ph.LastAttempt = e.Time
		ph.LastOK = e.Error == ""
		ph.LastError = e.Error
		if e.Error != "" {
			ph.Failures++
			if ph.FirstFailure == "" {
				ph.FirstFailure = e.Time
			}
		}
	}
	ph.Stuck = ph.Attempts > 0 && !ph.LastOK && strings.Contains(ph.LastError, "File name too long")
	return ph
}

func trimBefore(s, marker string) string {
	if i := strings.Index(s, marker); i >= 0 {
		return s[i:]
	}
	return s
}

// parseLogTime converts "2026-10-29 16:54:50.203203+0800" to RFC3339 local.
func parseLogTime(s string) string {
	for _, layout := range []string{"2006-01-02 15:04:05.000000-0700", "2006-01-02 15:04:05-0700"} {
		if t, err := time.Parse(layout, s); err == nil {
			return formatTime(t)
		}
	}
	return s
}

func formatTime(t time.Time) string {
	return t.Local().Format(time.RFC3339)
}

type tempDirSpec struct{ label, path string }

func tempDirSpecs(opts Options) []tempDirSpec {
	if len(opts.TempDirs) > 0 {
		specs := make([]tempDirSpec, 0, len(opts.TempDirs))
		for _, p := range opts.TempDirs {
			specs = append(specs, tempDirSpec{label: p, path: p})
		}
		return specs
	}
	tmp := os.TempDir()
	if real, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = real
	}
	return []tempDirSpec{
		{label: "$TMPDIR", path: tmp},
		{label: "/private/tmp", path: "/private/tmp"},
		{label: "/private/var/tmp", path: "/private/var/tmp"},
	}
}

// probeTempDirs walks each temp root: size + longest absolute path.
func probeTempDirs(opts Options) []TempDirProbe {
	var out []TempDirProbe
	for _, spec := range tempDirSpecs(opts) {
		p := TempDirProbe{Label: spec.label, Path: spec.path}
		_ = filepath.WalkDir(spec.path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if d != nil && d.IsDir() && path != spec.path {
					return filepath.SkipDir
				}
				return nil
			}
			if len(path) > p.LongestLen {
				p.LongestLen = len(path)
				p.LongestPath = path
			}
			if !d.IsDir() {
				if info, ierr := d.Info(); ierr == nil && info.Mode().IsRegular() {
					p.SizeBytes += info.Size()
				}
			}
			return nil
		})
		p.OverWarn = p.LongestLen > longPathWarn
		out = append(out, p)
	}
	return out
}

// measureDirsCleaner returns the size when readable (root / fixtures).
func measureDirsCleaner(opts Options) (int64, bool) {
	root := dirsCleanerRoot(opts)
	if _, err := os.ReadDir(root); err != nil {
		if os.IsNotExist(err) {
			return 0, true
		}
		return 0, false
	}
	return pathSize(root), true
}

const dirsCleanerAnalyseScript = `root=__ROOT__
cd "$root" 2>/dev/null || exit 0
for e in * .[!.]*; do
  [ -e "$e" ] || continue
  kb=$(/usr/bin/du -sk "$e" 2>/dev/null | /usr/bin/cut -f1)
  mt=$(/usr/bin/stat -f %m "$e" 2>/dev/null)
  rc=$(/usr/bin/find "$e" -flags +restricted 2>/dev/null | /usr/bin/wc -l | /usr/bin/tr -d ' ')
  printf 'B\t%s\t%s\t%s\t%s\n' "$e" "${mt:-0}" "${kb:-0}" "${rc:-0}"
  /usr/bin/find "$e" 2>/dev/null | /usr/bin/awk -v b="$e" -v base=${#root} -v lim=__LIMIT__ 'BEGIN{m=0;n=0} { l=length($0)+base+1; if (l>m) m=l; if (l>lim && n<5) { n++; printf "L\t%s\t%d\t%s\n", b, l, $0 } } END { printf "M\t%s\t%d\n", b, m }'
done
`

const dirsCleanerPurgeScript = `root=__ROOT__
if [ ! -d "$root" ]; then printf 'BEFORE\t0\nAFTER\t0\nLEFT\t0\n'; exit 0; fi
b=$(/usr/bin/du -sk "$root" 2>/dev/null | /usr/bin/cut -f1)
/usr/bin/find "$root" -mindepth 1 -maxdepth 1 -exec /bin/rm -rf {} + 2>/dev/null
a=$(/usr/bin/du -sk "$root" 2>/dev/null | /usr/bin/cut -f1)
printf 'BEFORE\t%s\nAFTER\t%s\n' "${b:-0}" "${a:-0}"
/usr/bin/find "$root" -mindepth 1 -flags +restricted 2>/dev/null | /usr/bin/head -20 | /usr/bin/awk '{print "R\t" $0}'
/usr/bin/find "$root" -mindepth 1 -maxdepth 1 2>/dev/null | /usr/bin/wc -l | /usr/bin/awk '{print "LEFT\t" $1}'
exit 0
`

// DirsCleanerAnalyseScript renders the read-only admin script (shown for audit).
func DirsCleanerAnalyseScript(opts Options) string {
	s := strings.ReplaceAll(dirsCleanerAnalyseScript, "__ROOT__", shQuote(dirsCleanerRoot(opts)))
	return strings.ReplaceAll(s, "__LIMIT__", strconv.Itoa(longPathWarn))
}

// DirsCleanerPurgeScript renders the purge script (shown for audit).
func DirsCleanerPurgeScript(opts Options) string {
	return strings.ReplaceAll(dirsCleanerPurgeScript, "__ROOT__", shQuote(dirsCleanerRoot(opts)))
}

// DirsCleanerPurgeCommand is the one-line equivalent printed in reports.
func DirsCleanerPurgeCommand(opts Options) string {
	return "sudo find " + shQuote(dirsCleanerRoot(opts)) + " -mindepth 1 -maxdepth 1 -exec rm -rf {} +"
}

// AnalyseDirsCleaner runs the read-only per-batch analysis (elevating if needed).
func AnalyseDirsCleaner(opts Options) (*DirsCleanerAnalysis, error) {
	out, err := runDirsCleanerScript(opts, "dirs-cleaner analyse", DirsCleanerAnalyseScript(opts), false)
	if err != nil {
		return nil, err
	}
	a := parseAnalyseOutput(out)
	a.AnalysedAt = formatTime(time.Now())
	return a, nil
}

func parseAnalyseOutput(out []byte) *DirsCleanerAnalysis {
	a := &DirsCleanerAnalysis{Batches: []DirsCleanerBatch{}, Offenders: []PathOffender{}}
	idx := map[string]int{}
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.SplitN(line, "\t", 5)
		switch {
		case len(f) == 5 && f[0] == "B":
			mt, _ := strconv.ParseInt(f[2], 10, 64)
			kb, _ := strconv.ParseInt(f[3], 10, 64)
			rc, _ := strconv.Atoi(f[4])
			b := DirsCleanerBatch{Name: f[1], SizeBytes: kb * 1024, Restricted: rc}
			if mt > 0 {
				b.ModTime = formatTime(time.Unix(mt, 0))
			}
			idx[b.Name] = len(a.Batches)
			a.Batches = append(a.Batches, b)
			a.TotalBytes += b.SizeBytes
			a.RestrictedFiles += rc
		case len(f) >= 3 && f[0] == "M":
			n, _ := strconv.Atoi(f[2])
			if i, ok := idx[f[1]]; ok {
				a.Batches[i].MaxPathLen = n
			}
		case len(f) >= 4 && f[0] == "L":
			n, _ := strconv.Atoi(f[2])
			path := strings.Join(f[3:], "\t")
			pattern, repeats := detectRepeat(path)
			a.Offenders = append(a.Offenders, PathOffender{Batch: f[1], Path: path, Len: n, Pattern: pattern, Repeats: repeats})
		}
	}
	sort.SliceStable(a.Batches, func(i, j int) bool { return a.Batches[i].SizeBytes > a.Batches[j].SizeBytes })
	sort.SliceStable(a.Offenders, func(i, j int) bool { return a.Offenders[i].Len > a.Offenders[j].Len })
	return a
}

// PurgeDirsCleaner removes every batch and keeps the directory.
func PurgeDirsCleaner(opts Options) (*PurgeOutcome, error) {
	root := dirsCleanerRoot(opts)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return &PurgeOutcome{}, nil
	}
	out, err := runDirsCleanerScript(opts, "dirs-cleaner purge", DirsCleanerPurgeScript(opts), true)
	if err != nil {
		return nil, err
	}
	res := &PurgeOutcome{}
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.SplitN(line, "\t", 2)
		if len(f) != 2 {
			continue
		}
		switch f[0] {
		case "BEFORE":
			kb, _ := strconv.ParseInt(strings.TrimSpace(f[1]), 10, 64)
			res.Before = kb * 1024
		case "AFTER":
			kb, _ := strconv.ParseInt(strings.TrimSpace(f[1]), 10, 64)
			res.After = kb * 1024
		case "LEFT":
			res.Left, _ = strconv.Atoi(strings.TrimSpace(f[1]))
		case "R":
			res.Restricted = append(res.Restricted, f[1])
		}
	}
	return res, nil
}

// runDirsCleanerScript runs a fixed dirs_cleaner script, elevating only when
// the root is not accessible to this process. Non-default roots never elevate.
func runDirsCleanerScript(opts Options, purpose, script string, write bool) ([]byte, error) {
	root := dirsCleanerRoot(opts)
	accessible := false
	if write {
		accessible = canWriteAssetRoot(root)
	} else {
		_, err := os.ReadDir(root)
		accessible = err == nil || os.IsNotExist(err)
	}
	if geteuid(opts) == 0 || accessible {
		return runShellScript(opts, script)
	}
	if opts.Elevate != nil {
		return opts.Elevate(purpose, script)
	}
	if opts.ElevateMode == ElevateOSAScript {
		if root != defaultDirsCleanerRoot {
			return nil, fmt.Errorf("refusing to elevate for non-default dirs_cleaner root %q", root)
		}
		return osascriptElevate(opts, purpose, script)
	}
	if !isInteractiveTTY(opts) {
		return nil, fmt.Errorf("%s requires admin privileges; re-run in a terminal or under sudo", purpose)
	}
	if root != defaultDirsCleanerRoot {
		return nil, fmt.Errorf("refusing to elevate for non-default dirs_cleaner root %q", root)
	}
	warnf(opts.Stderr, "%s requires admin privileges; prompting for sudo password", purpose)
	return sudoScript(opts, purpose, script)
}

func runShellScript(opts Options, script string) ([]byte, error) {
	return elevate.RunShell(script)
}

func sudoScript(opts Options, purpose, script string) ([]byte, error) {
	return elevate.Run(elevate.Config{
		Purpose: purpose,
		Prompt:  opts.ElevatePrompt,
		Script:  script,
		Mode:    elevate.ModeSudo,
		Stderr:  opts.Stderr,
	})
}

// osascriptElevate shows the macOS administrator dialog. The script travels
// base64-encoded so no AppleScript quoting is involved; the password never
// passes through clear-tmp.
func osascriptElevate(opts Options, purpose, script string) ([]byte, error) {
	return elevate.Run(elevate.Config{
		Purpose: purpose,
		Prompt:  opts.ElevatePrompt,
		Script:  script,
		Mode:    elevate.ModeOsaScript,
		Stderr:  opts.Stderr,
	})
}

// detectRepeat finds the earliest path component that repeats and returns the
// cycle from its first occurrence up to the second, e.g.
// a/b/x/c/x/c/x/c/d → ("x/c/", 3). No repeat → ("", 0).
func detectRepeat(path string) (string, int) {
	comps := strings.Split(strings.Trim(path, "/"), "/")
	count := map[string]int{}
	first := map[string]int{}
	for i, c := range comps {
		if c == "" || c == "." {
			continue
		}
		count[c]++
		if _, ok := first[c]; !ok {
			first[c] = i
		}
	}
	best := ""
	for c, n := range count {
		if n < 2 {
			continue
		}
		if best == "" || n > count[best] || (n == count[best] && first[c] < first[best]) {
			best = c
		}
	}
	if best == "" {
		return "", 0
	}
	start := first[best]
	end := len(comps)
	for j := start + 1; j < len(comps); j++ {
		if comps[j] == best {
			end = j
			break
		}
	}
	seg := comps[start:end]
	if len(seg) > 8 {
		seg = append(append([]string(nil), seg[:7]...), "…")
	}
	return strings.Join(seg, "/") + "/", count[best]
}

func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// truncMiddle shortens s to n runes keeping both ends.
func truncMiddle(s string, n int) string {
	r := []rune(s)
	if len(r) <= n || n < 5 {
		return s
	}
	half := (n - 1) / 2
	return string(r[:half]) + "…" + string(r[len(r)-(n-1-half):])
}

// printDirsCleanerSection is the dry-run inventory block (metadata only).
func printDirsCleanerSection(w io.Writer, use bool, st DirsCleanerStatus, included bool) {
	fmt.Fprintln(w, gray(use, "dirs_cleaner ("+st.Root+"):"))
	if !st.Exists {
		fmt.Fprintln(w, "  (not present)")
	} else {
		size := "size needs admin"
		if st.SizeBytes != nil {
			size = fmt.Sprintf("%s (%d B)", humanSize(*st.SizeBytes), *st.SizeBytes)
		}
		status := "not purged by default"
		if included {
			status = "would purge / included"
		}
		fmt.Fprintf(w, "  %d batches, last written %s, %s  [%s]\n", st.Batches, st.ModTime, size, status)
	}
	fmt.Fprintln(w, "note: purge needs --dirs-cleaner (sudo); health check: clear-tmp dirs-cleaner status")
}

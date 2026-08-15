# Cloudflare managed-tunnels registry (owned coverage)

Coverage-backfill doctests for stable on-disk layout under
`{ConfigDir}/managed-tunnels/<tunnelNameSafe>/` plus pure multi-host ingress
config build from `TunnelState`. **No** cloudflared process, flock, DNS, or
Attach/Detach lifecycle.

Package under test:

`github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`

Run from this **go-pkgs module root** (no `replace` needed).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Participants**

- **ConfigDir** is the cloudflared home root (default `~/.cloudflared` via
  existing `DefaultConfigDir()`). Tests inject a temp dir — never product
  names like `spl-seatalk` in library path segments.
- **ManagedTunnelsRoot** joins ConfigDir with the fixed segment
  `managed-tunnels`.
- **TunnelNameSafe** maps a human tunnel name to a filesystem-safe directory
  segment: normal `[a-zA-Z0-9_-]` names lowercased; path-unsafe characters
  (`/`, `\`, etc.) replaced with `-`.
- **ManagedTunnelDir** resolves
  `{ConfigDir}/managed-tunnels/<TunnelNameSafe(name)>`; empty tunnel name is
  an error.
- **TunnelState** is the `state.json` document: tunnel metadata plus
  `Hosts map[hostname]*HostEntry` (service URL, optional owner PID / attach
  time).
- **LoadTunnelState / SaveTunnelState** read and write `state.json` under a
  tunnel directory. Missing file loads as empty state with non-nil `Hosts`
  and no error (attach-friendly). Save creates parent dirs.
- **BuildConfigFromState** is pure: copies tunnel id / credentials file from
  state, emits one ingress rule per host, **sorted by hostname** for
  determinism, then a catch-all `{Service: "http_status:404"}` with empty
  hostname last.
- **WriteConfig** already exists; optional after build. This tree focuses on
  path helpers, state I/O, and build purity — no process start.

**Behaviors**

```text
# paths
configDir
  -> ManagedTunnelsRoot(configDir) = Join(configDir, "managed-tunnels")
tunnelName
  -> TunnelNameSafe(tunnelName)  # lower + replace path-unsafe with '-'
configDir, tunnelName
  -> ManagedTunnelDir(...) = Join(root, safeName) | error if empty name

# state.json under tunnel dir
dir, TunnelState
  -> SaveTunnelState(dir, st) writes state.json (mkdir)
  -> LoadTunnelState(dir) reads back; missing file -> empty Hosts, nil err

# pure ingress build
TunnelState{Hosts}
  -> BuildConfigFromState(st)
  -> Config.Ingress = sorted host rules + http_status:404 catch-all last
```

**On-disk layout (normative)**

```text
{ConfigDir}/managed-tunnels/<tunnelNameSafe>/
  state.json
  config.yml   # optional; may be written by helpers after BuildConfigFromState
```

## Decision Tree

```text
tests/cloudflare-managed-tunnels/
├── DOCTEST.md
├── SETUP.md
├── paths/                               # pure path / name helpers
│   ├── root/
│   │   └── joins-managed-tunnels/       # ManagedTunnelsRoot joins segment
│   ├── name-safe/
│   │   ├── lowercases-alnum/            # My-Tunnel -> my-tunnel
│   │   └── replaces-path-unsafe/        # slash (and kin) not in result
│   └── tunnel-dir/
│       ├── team-shared/                 # full dir under managed-tunnels
│       └── empty-name-errors/           # empty tunnel name -> error
├── state/                               # state.json save / load
│   ├── save-load-roundtrip/             # 1 host survives save+load
│   └── load-missing-empty/              # missing file -> empty Hosts, nil
└── build-config/                        # BuildConfigFromState pure build
    ├── single-host-plus-404/            # 1 host rule + 404 catch-all
    ├── multi-host-both-and-404-last/    # 2 hosts present; 404 last only
    └── host-order-sorted/               # sorted hostnames; deterministic
```

### Parameter significance (high → low)

1. **Surface** — paths vs state I/O vs pure config build (largest API split)
2. **Path helper** — root join vs name-safe vs full tunnel dir
3. **State outcome** — successful round-trip vs missing-file empty load
4. **Build shape** — single host vs multi-host presence/404 vs sort order
5. **Fixture values** — concrete names, hostnames, services (leaf constants)

## Test Index

| # | Leaf | Mode | Assert focus |
|---|------|------|--------------|
| 1 | `paths/root/joins-managed-tunnels` | `managed_root` | `ManagedTunnelsRoot("/tmp/cf")` → `/tmp/cf/managed-tunnels` |
| 2 | `paths/name-safe/lowercases-alnum` | `name_safe` | `TunnelNameSafe("My-Tunnel")` → `my-tunnel` |
| 3 | `paths/name-safe/replaces-path-unsafe` | `name_safe` | slash input → result contains no `/` or `\` |
| 4 | `paths/tunnel-dir/team-shared` | `tunnel_dir` | dir ends with `managed-tunnels/team-shared` (or safe form) |
| 5 | `paths/tunnel-dir/empty-name-errors` | `tunnel_dir` | empty name → non-nil error |
| 6 | `state/save-load-roundtrip` | `save_load` | 1 host state equal after Save+Load |
| 7 | `state/load-missing-empty` | `load_missing` | missing state.json → Hosts non-nil empty map, err nil |
| 8 | `build-config/single-host-plus-404` | `build_config` | ingress len 2: host rule then `http_status:404` |
| 9 | `build-config/multi-host-both-and-404-last` | `build_config` | both hostnames present; last rule empty host + 404 |
| 10 | `build-config/host-order-sorted` | `build_config` | hostnames alphabetical; two builds identical order |

## How to Run

```sh
cd <go-pkgs-module-root>   # this module; package is local

doctest vet ./tests/cloudflare-managed-tunnels
doctest test ./tests/cloudflare-managed-tunnels
doctest test -v ./tests/cloudflare-managed-tunnels
```

**Coverage backfill / GREEN expected:** implementation is present in
`cloudflare/managed.go` (`ManagedTunnelsRoot`, `TunnelNameSafe`,
`ManagedTunnelDir`, `TunnelState` / `HostEntry`, `LoadTunnelState`,
`SaveTunnelState`, `BuildConfigFromState`). Leaves exercise existing correct
behavior — no classic-TDD RED intent.

### API surface under test

```go
func ManagedTunnelsRoot(configDir string) string
func TunnelNameSafe(tunnelName string) string
func ManagedTunnelDir(configDir, tunnelName string) (string, error) // empty name → error

type TunnelState struct {
	TunnelName      string                `json:"tunnel_name"`
	TunnelID        string                `json:"tunnel_id,omitempty"`
	CredentialsFile string                `json:"credentials_file,omitempty"`
	ConnectorPID    int                   `json:"connector_pid,omitempty"`
	Hosts           map[string]*HostEntry `json:"hosts"` // key = hostname
}

type HostEntry struct {
	Service    string `json:"service"`
	OwnerPID   int    `json:"owner_pid,omitempty"`
	AttachedAt string `json:"attached_at,omitempty"` // RFC3339 optional
}

func LoadTunnelState(dir string) (*TunnelState, error)  // missing → empty Hosts, nil
func SaveTunnelState(dir string, st *TunnelState) error // mkdir + state.json

func BuildConfigFromState(st *TunnelState) (*Config, error)
// Ingress: sorted hostnames → rules, then catch-all http_status:404 (empty Hostname).
// Tunnel and CredentialsFile copied from state.
```

### Out of scope

- flock / concurrent attach
- Attach / Detach lifecycle
- cloudflared process start/restart
- DNS route helpers
- spl CLI / local-bot wiring

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

// Request selects a surface and supplies fixtures.
type Request struct {
	// Mode:
	//   "managed_root" | "name_safe" | "tunnel_dir" |
	//   "save_load" | "load_missing" | "build_config"
	Mode string

	// ConfigDir for ManagedTunnelsRoot / ManagedTunnelDir (paths leaves).
	ConfigDir string
	// TunnelName for TunnelNameSafe / ManagedTunnelDir.
	TunnelName string

	// StateDir is the per-tunnel directory for save/load leaves.
	// Empty in Setup → Run uses t.TempDir() subdir for save_load, or a
	// non-existent state path under temp for load_missing.
	StateDir string

	// StateIn is the TunnelState written on save_load or built for build_config.
	StateIn *cloudflare.TunnelState

	// ExpectError: leaf expects Run/API to return a non-nil error (empty name).
	ExpectError bool

	DecisionPath []string
}

// Response holds path strings, loaded state, or built config.
type Response struct {
	Path string // ManagedTunnelsRoot or ManagedTunnelDir or SafeName

	State  *cloudflare.TunnelState
	Config *cloudflare.Config
	// Config2 is a second BuildConfigFromState result (host-order leaf only).
	Config2 *cloudflare.Config

	// DirExists is true when StateDir existed before load (diagnostic).
	DirExists bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	switch strings.TrimSpace(req.Mode) {
	case "managed_root":
		p := cloudflare.ManagedTunnelsRoot(req.ConfigDir)
		return &Response{Path: p}, nil

	case "name_safe":
		s := cloudflare.TunnelNameSafe(req.TunnelName)
		return &Response{Path: s}, nil

	case "tunnel_dir":
		dir, err := cloudflare.ManagedTunnelDir(req.ConfigDir, req.TunnelName)
		if err != nil {
			return &Response{Path: dir}, err
		}
		return &Response{Path: dir}, nil

	case "save_load":
		dir := req.StateDir
		if strings.TrimSpace(dir) == "" {
			dir = filepath.Join(t.TempDir(), "managed-tunnels", "fixture-tunnel")
		}
		if req.StateIn == nil {
			return nil, fmt.Errorf("save_load requires StateIn")
		}
		if err := cloudflare.SaveTunnelState(dir, req.StateIn); err != nil {
			return nil, err
		}
		got, err := cloudflare.LoadTunnelState(dir)
		if err != nil {
			return &Response{Path: dir}, err
		}
		return &Response{Path: dir, State: got}, nil

	case "load_missing":
		dir := req.StateDir
		if strings.TrimSpace(dir) == "" {
			// Empty dir that does not contain state.json.
			dir = filepath.Join(t.TempDir(), "managed-tunnels", "no-such-state")
		}
		_, statErr := os.Stat(filepath.Join(dir, "state.json"))
		dirExists := statErr == nil
		got, err := cloudflare.LoadTunnelState(dir)
		return &Response{Path: dir, State: got, DirExists: dirExists}, err

	case "build_config":
		if req.StateIn == nil {
			return nil, fmt.Errorf("build_config requires StateIn")
		}
		cfg, err := cloudflare.BuildConfigFromState(req.StateIn)
		if err != nil {
			return nil, err
		}
		cfg2, err2 := cloudflare.BuildConfigFromState(req.StateIn)
		if err2 != nil {
			return &Response{Config: cfg}, err2
		}
		return &Response{Config: cfg, Config2: cfg2}, nil

	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func assertPathEq(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func assertNoPathSep(t *testing.T, s string) {
	t.Helper()
	if strings.ContainsAny(s, `/\`) {
		t.Fatalf("TunnelNameSafe result %q must not contain path separators", s)
	}
	if s == "" {
		t.Fatal("TunnelNameSafe result empty")
	}
}

func assertHostsEqual(t *testing.T, got, want map[string]*cloudflare.HostEntry) {
	t.Helper()
	if got == nil {
		t.Fatal("Hosts map is nil")
	}
	if len(got) != len(want) {
		t.Fatalf("Hosts len = %d, want %d", len(got), len(want))
	}
	for k, we := range want {
		ge, ok := got[k]
		if !ok {
			t.Fatalf("Hosts missing key %q", k)
		}
		if ge == nil {
			t.Fatalf("Hosts[%q] is nil", k)
		}
		if we == nil {
			continue
		}
		if ge.Service != we.Service {
			t.Fatalf("Hosts[%q].Service = %q, want %q", k, ge.Service, we.Service)
		}
		if ge.OwnerPID != we.OwnerPID {
			t.Fatalf("Hosts[%q].OwnerPID = %d, want %d", k, ge.OwnerPID, we.OwnerPID)
		}
	}
}

func ingressHostnames(cfg *cloudflare.Config) []string {
	if cfg == nil {
		return nil
	}
	var out []string
	for _, r := range cfg.Ingress {
		out = append(out, r.Hostname)
	}
	return out
}

func hostRulesOnly(cfg *cloudflare.Config) []cloudflare.IngressRule {
	if cfg == nil {
		return nil
	}
	var out []cloudflare.IngressRule
	for _, r := range cfg.Ingress {
		if r.Hostname != "" {
			out = append(out, r)
		}
	}
	return out
}

func lastIngress(cfg *cloudflare.Config) (cloudflare.IngressRule, bool) {
	if cfg == nil || len(cfg.Ingress) == 0 {
		return cloudflare.IngressRule{}, false
	}
	return cfg.Ingress[len(cfg.Ingress)-1], true
}

func sortedKeys(m map[string]*cloudflare.HostEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mustState(tunnelName, tunnelID, cred string, hosts map[string]*cloudflare.HostEntry) *cloudflare.TunnelState {
	if hosts == nil {
		hosts = map[string]*cloudflare.HostEntry{}
	}
	return &cloudflare.TunnelState{
		TunnelName:      tunnelName,
		TunnelID:        tunnelID,
		CredentialsFile: cred,
		Hosts:           hosts,
	}
}
```

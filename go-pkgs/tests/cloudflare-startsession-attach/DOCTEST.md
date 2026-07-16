# Cloudflare StartSession → Attach (owned coverage)

Coverage-backfill doctests for `StartSession` as a thin wrapper around managed
multi-host **`Attach`** so shared tunnel names are multi-host-safe; preserve
`PublicBaseURL` and managed `Stop` (Detach) semantics. **No** real network or
real `cloudflared` binary.

Package under test:

`github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`

Run from this **go-pkgs module root** (no `replace` needed).

Depends on managed helpers and **Attach + Detach (GREEN)**.
**`StartSession` wraps `Attach`** with `SessionOptions.ConfigDir` — GREEN expected.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Participants**

- **StartSession** is the product-facing session entrypoint. It must call
  `Attach(AttachOptions{...})` (or equivalent) so multi-host registry merge
  and managed Stop work through the same API callers already use.
- **SessionOptions.ConfigDir** is the cloudflared home root for managed paths.
  Empty → `DefaultConfigDir()`. Tests inject `t.TempDir()`.
- **Attach** is the multi-host managed entrypoint.
- **Session.PublicBaseURL** remains `https://<Domain>` (no path).
- **Session.Stop** for sessions from StartSession-via-Attach must be managed:
  Detach one host; siblings stay.
- **ConfigDir / Managed tunnel directory** layout:
  `{ConfigDir}/managed-tunnels/<TunnelNameSafe>/{lock,state.json,config.yml}`.
- **CommandRunner** is the same injectable fake as attach/detach trees.
  Soft-succeeds `tunnel list|create|info|route dns|run` and materializes dummy
  credentials under sandboxed `HOME`. Each `Exec(…, "run", …)` counts as
  connector start/restart.
- **TunnelState / Hosts** is the multi-host registry. StartSession(A) then
  StartSession(B) same TunnelName+ConfigDir → Hosts has A and B.

**Behaviors**

```text
# first session (managed ConfigDir)
StartSession(Domain=A, ConfigDir=tmp, Runner=fake)
  -> Session.PublicBaseURL = https://A
  -> managed state Hosts[A] set
  -> config.yml under ManagedTunnelDir(ConfigDir, TunnelName)

# multi-host same tunnel
StartSession(A); StartSession(B) same TunnelName + ConfigDir
  -> LoadTunnelState Hosts has A and B

# stop one host
StartSession(A); StartSession(B); Stop(A)
  -> Hosts only B (managed Detach path)

# fake runner
StartSession(..., Runner=fake)
  -> no panic; RunCount ≥ 1 (tunnel … run Exec)
```

## Decision Tree

```text
tests/cloudflare-startsession-attach/
├── DOCTEST.md
├── SETUP.md
├── first-session/                       # single StartSession + managed ConfigDir
│   └── public-base-url/                 # PublicBaseURL + Hosts[A] under managed dir
├── multi-host/                          # two StartSession same tunnel/config
│   ├── hosts-merge/                     # Hosts A and B after A then B
│   └── stop-leaves-sibling/             # Stop(A) leaves only B
└── runner/                              # injectable CommandRunner path
    └── exec-run/                        # RunCount ≥ 1; no panic; managed config
```

### Parameter significance (high → low)

1. **Outcome class** — first session contract vs multi-host merge vs partial stop vs runner
2. **Host cardinality / sequence** — one StartSession vs two; optional Stop after dual
3. **Observable focus** — PublicBaseURL+managed registry vs Hosts map vs Stop residual vs run count
4. **Fixture values** — concrete hostnames, ports, tunnel names (leaf constants)

## Test Index

| # | Leaf | Assert focus |
|---|------|--------------|
| 1 | `first-session/public-base-url` | StartSession(A, ConfigDir): PublicBaseURL `https://A`; managed Hosts[A]; ConfigPath under ManagedTunnelDir |
| 2 | `multi-host/hosts-merge` | StartSession A then B same TunnelName+ConfigDir → Hosts has A and B |
| 3 | `multi-host/stop-leaves-sibling` | After A+B, Stop(A) → Hosts only B |
| 4 | `runner/exec-run` | Fake Runner: no panic; RunCount ≥ 1; managed config path used when present |

## How to Run

```sh
cd <go-pkgs-module-root>   # this module; package is local

doctest vet ./tests/cloudflare-startsession-attach
doctest test ./tests/cloudflare-startsession-attach
doctest test -v ./tests/cloudflare-startsession-attach
```

**Coverage backfill / GREEN expected:** `SessionOptions.ConfigDir` and
`StartSession` → `Attach` present in `cloudflare/types.go` / `session.go`.

### API surface under test

```go
type SessionOptions struct {
    Domain     string
    LocalURL   string
    TunnelName string
    WorkDir    string // legacy; may be unused when ConfigDir drives managed path
    ConfigDir  string // empty → DefaultConfigDir for managed path when using Attach
    Log        io.Writer
    Runner     CommandRunner
    DNSDeleter DNSDeleter
}

// StartSession(opts) (*Session, error)
// Implementation: call Attach(AttachOptions{
//   Domain, LocalURL, TunnelName, ConfigDir, Log, Runner, DNSDeleter from opts,
// })
// Preserve PublicBaseURL == "https://"+Domain and managed Session.Stop → Detach.
```

### Out of scope

- spl CLI help / wiring
- docs polish
- Real OS process StartProcess path (tests inject Runner only)
- Changing Attach/Detach contracts beyond what StartSession needs to forward

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
	"gopkg.in/yaml.v3"
)

// sessionStep is one StartSession call sharing ConfigDir / TunnelName / Runner.
type sessionStep struct {
	Domain   string
	LocalURL string
}

// stopStep releases one hostname after StartSession setup via Session.Stop.
type stopStep struct {
	Domain string
}

// Request selects StartSession inputs (single call or multi-step sequence).
type Request struct {
	// Domain / LocalURL used when Sequence is empty (single StartSession).
	Domain   string
	LocalURL string

	// TunnelName empty → production default (DefaultTunnelName).
	TunnelName string

	// ConfigDir is the managed cloudflared root. Empty in leaf → root Setup
	// fills t.TempDir(). Always non-empty before Run for lifecycle leaves.
	// SessionOptions.ConfigDir drives managed Attach path.
	ConfigDir string

	// Sequence, when non-empty, runs multiple StartSession calls sharing
	// ConfigDir, TunnelName, and the same fake runner (multi-host leaves).
	Sequence []sessionStep

	// StopSequence runs Session.Stop in order after StartSession sequence.
	StopSequence []stopStep

	// ExpectError: leaf expects StartSession/Stop to return a non-nil error.
	ExpectError bool

	// HomeDir is a sandboxed HOME so DefaultConfigDir() resolves under temp.
	// Root Setup sets this and t.Setenv("HOME", HomeDir).
	HomeDir string

	DecisionPath []string
}

// Response captures StartSession results and fake-runner observations.
type Response struct {
	// Session is the last successful StartSession session (nil on error).
	Session *cloudflare.Session

	// Sessions holds every successful session in a multi-step Sequence.
	Sessions []*cloudflare.Session

	// SessionByDomain maps Domain → last StartSession session for that host.
	SessionByDomain map[string]*cloudflare.Session

	// State is LoadTunnelState after the last successful action.
	State *cloudflare.TunnelState

	// Config is the parsed managed config.yml after the last success.
	Config *cloudflare.Config

	// ConfigPath is filepath.Join(ManagedDir, "config.yml") when present.
	ConfigPath string

	// ManagedDir is ManagedTunnelDir(ConfigDir, resolved tunnel name).
	ManagedDir string

	// Runner is the fake used for this Run (for Inspect helpers).
	Runner *fakeCloudflared

	// RunCount is how many Exec calls included a "run" subcommand token.
	RunCount int

	// RouteDNSCount is how many Exec calls looked like tunnel route dns.
	RouteDNSCount int

	// LastStopErr is the error from the last Stop step (nil if none).
	LastStopErr error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	steps := req.Sequence
	if len(steps) == 0 {
		steps = []sessionStep{{Domain: req.Domain, LocalURL: req.LocalURL}}
	}

	fake := newFakeCloudflared(t, req.HomeDir)
	resp := &Response{
		Runner:          fake,
		SessionByDomain: map[string]*cloudflare.Session{},
	}

	var lastErr error
	for i, step := range steps {
		// SessionOptions.ConfigDir drives managed Attach path.
		opts := cloudflare.SessionOptions{
			Domain:     step.Domain,
			LocalURL:   step.LocalURL,
			TunnelName: req.TunnelName,
			ConfigDir:  req.ConfigDir,
			Runner:     fake,
			Log:        os.Stderr,
		}
		sess, err := cloudflare.StartSession(opts)
		lastErr = err
		if err != nil {
			resp.RunCount = fake.runCount()
			resp.RouteDNSCount = fake.routeDNSCount()
			return resp, err
		}
		resp.Session = sess
		resp.Sessions = append(resp.Sessions, sess)
		resp.SessionByDomain[step.Domain] = sess
		t.Logf("startsession step %d ok domain=%q tunnel=%q config=%q public=%q",
			i, sess.Domain, sess.TunnelName, sess.ConfigPath, sess.PublicBaseURL())
	}

	// Resolve managed dir for post-conditions (use last session tunnel name).
	name := req.TunnelName
	if strings.TrimSpace(name) == "" {
		name = cloudflare.DefaultTunnelName
	}
	if resp.Session != nil && strings.TrimSpace(resp.Session.TunnelName) != "" {
		name = resp.Session.TunnelName
	}
	dir, derr := cloudflare.ManagedTunnelDir(req.ConfigDir, name)
	if derr != nil {
		resp.RunCount = fake.runCount()
		resp.RouteDNSCount = fake.routeDNSCount()
		return resp, fmt.Errorf("ManagedTunnelDir: %w", derr)
	}
	resp.ManagedDir = dir
	resp.ConfigPath = filepath.Join(dir, "config.yml")

	// --- Stop sequence (managed Stop → Detach) ---
	for i, step := range req.StopSequence {
		sess := resp.SessionByDomain[step.Domain]
		if sess == nil {
			lastErr = fmt.Errorf("no StartSession session for domain %q", step.Domain)
			resp.LastStopErr = lastErr
			resp.RunCount = fake.runCount()
			resp.RouteDNSCount = fake.routeDNSCount()
			return resp, lastErr
		}
		err := sess.Stop()
		resp.LastStopErr = err
		lastErr = err
		t.Logf("stop step %d domain=%q Session.Stop err=%v", i, step.Domain, err)
		if err != nil && !req.ExpectError {
			_ = loadStateAndConfig(resp, dir)
			resp.RunCount = fake.runCount()
			resp.RouteDNSCount = fake.routeDNSCount()
			return resp, err
		}
	}

	if err := loadStateAndConfig(resp, dir); err != nil {
		// LoadTunnelState treats missing file as empty Hosts + nil err; only
		// hard errors propagate.
		resp.RunCount = fake.runCount()
		resp.RouteDNSCount = fake.routeDNSCount()
		return resp, err
	}

	resp.RunCount = fake.runCount()
	resp.RouteDNSCount = fake.routeDNSCount()
	return resp, lastErr
}

func loadStateAndConfig(resp *Response, dir string) error {
	st, lerr := cloudflare.LoadTunnelState(dir)
	if lerr != nil {
		return lerr
	}
	resp.State = st
	if data, rerr := os.ReadFile(resp.ConfigPath); rerr == nil {
		var cfg cloudflare.Config
		if yerr := yaml.Unmarshal(data, &cfg); yerr == nil {
			resp.Config = &cfg
		}
	}
	return nil
}

// fakeCloudflared soft-succeeds cloudflared CLI shapes used by
// ensureTunnelForSession / RouteDNS / Attach / StartSession connector start.
// Same soft contracts as tests/cloudflare-tunnel-attach and
// tests/cloudflare-tunnel-detach.
type fakeCloudflared struct {
	mu sync.Mutex

	homeCloudflared string
	tunnelID        string
	tunnelName      string
	credFile        string
	created         bool

	calls []fakeCall
}

type fakeCall struct {
	Name string
	Args []string
}

const fixtureTunnelID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

func newFakeCloudflared(t *testing.T, homeDir string) *fakeCloudflared {
	t.Helper()
	cfDir := filepath.Join(homeDir, ".cloudflared")
	if err := os.MkdirAll(cfDir, 0o755); err != nil {
		t.Fatalf("mkdir cloudflared home: %v", err)
	}
	return &fakeCloudflared{
		homeCloudflared: cfDir,
		tunnelID:        fixtureTunnelID,
	}
}

func (f *fakeCloudflared) Exec(name string, args ...string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	cp := append([]string(nil), args...)
	f.calls = append(f.calls, fakeCall{Name: name, Args: cp})

	joined := strings.Join(args, " ")

	if len(args) == 0 {
		return []byte("cloudflared version fake"), nil
	}

	// tunnel list --output json
	if containsArg(args, "list") && containsArg(args, "--output") {
		if f.created && f.tunnelName != "" {
			b, _ := json.Marshal([]map[string]string{
				{"id": f.tunnelID, "name": f.tunnelName},
			})
			return b, nil
		}
		return []byte(`[]`), nil
	}

	// tunnel create <name>
	if containsArg(args, "create") {
		tname := args[len(args)-1]
		if tname == "create" || tname == "tunnel" {
			tname = cloudflare.DefaultTunnelName
		}
		f.tunnelName = tname
		if err := f.writeCredsLocked(); err != nil {
			return nil, err
		}
		f.created = true
		msg := fmt.Sprintf(
			"Created tunnel %s with id %s\ncredentials written to %s\n",
			f.tunnelName, f.tunnelID, f.credFile,
		)
		return []byte(msg), nil
	}

	// tunnel info <name|id>
	if containsArg(args, "info") {
		if !f.created {
			_ = f.writeCredsLocked()
			f.created = true
		}
		msg := fmt.Sprintf("NAME: %s\nID: %s\n", f.tunnelName, f.tunnelID)
		return []byte(msg), nil
	}

	// tunnel route dns
	if containsArg(args, "route") && containsArg(args, "dns") {
		return []byte("Added CNAME " + joined), nil
	}

	// tunnel … run …
	if containsArg(args, "run") {
		return []byte("fake connector running\n"), nil
	}

	return []byte("ok"), nil
}

func (f *fakeCloudflared) writeCredsLocked() error {
	if f.credFile == "" {
		f.credFile = filepath.Join(f.homeCloudflared, f.tunnelID+".json")
	}
	body := fmt.Sprintf(
		`{"AccountTag":"test","TunnelID":%q,"TunnelName":%q,"TunnelSecret":"dGVzdA=="}`+"\n",
		f.tunnelID, f.tunnelName,
	)
	if err := os.MkdirAll(filepath.Dir(f.credFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(f.credFile, []byte(body), 0o644)
}

func (f *fakeCloudflared) runCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if containsArg(c.Args, "run") {
			n++
		}
	}
	return n
}

func (f *fakeCloudflared) routeDNSCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if containsArg(c.Args, "route") && containsArg(c.Args, "dns") {
			n++
		}
	}
	return n
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

func lastIngress(cfg *cloudflare.Config) (cloudflare.IngressRule, bool) {
	if cfg == nil || len(cfg.Ingress) == 0 {
		return cloudflare.IngressRule{}, false
	}
	return cfg.Ingress[len(cfg.Ingress)-1], true
}

func hostRuleServices(cfg *cloudflare.Config) map[string]string {
	out := map[string]string{}
	if cfg == nil {
		return out
	}
	for _, r := range cfg.Ingress {
		if r.Hostname != "" {
			out[r.Hostname] = r.Service
		}
	}
	return out
}
```

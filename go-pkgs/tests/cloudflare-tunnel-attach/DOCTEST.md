# Cloudflare managed tunnel Attach (owned coverage)

Coverage-backfill doctests for `cloudflare.Attach`: merges a hostname into a
named managed-tunnel registry under flock, routes DNS, writes multi-host
`config.yml`, and ensures a single connector via injectable `CommandRunner`
(`tunnel … run`). **No** real network or real `cloudflared` binary.

Package under test:

`github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`

Run from this **go-pkgs module root** (no `replace` needed).

Depends on managed helpers: `ManagedTunnelDir`, `LoadTunnelState`,
`SaveTunnelState`, `BuildConfigFromState`, `WriteConfig`. **`Attach` is
present** — GREEN expected.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Participants**

- **Attach** is the multi-host managed entrypoint: given Domain (+ optional
  LocalURL, TunnelName, ConfigDir, Runner), it updates on-disk registry state
  and keeps one healthy connector for that tunnel name.
- **ConfigDir** is the cloudflared home root (tests inject `t.TempDir()`; never
  product names in library path segments). Managed layout:
  `{ConfigDir}/managed-tunnels/<TunnelNameSafe>/`.
- **Managed tunnel directory** holds `lock`, `state.json`, and `config.yml`.
- **TunnelState / Hosts** map public hostnames to local service URLs (and owner
  metadata). Attach merges Hosts; does not wipe sibling hosts.
- **CommandRunner** is an injectable fake in tests. Soft-succeeds
  `tunnel list|create|info|route dns|run` and materializes dummy credential
  JSON so `ensureTunnelForSession` can resolve tunnel ID + credentials without
  a real binary. `HOME` is redirected to a temp dir so
  `DefaultConfigDir()` (`~/.cloudflared`) stays sandboxed.
- **Connector** in fake mode is not an OS process: each
  `Exec("cloudflared", "tunnel", …, "run", name)` counts as a connector
  start/restart. Ingress add/update → another `run` Exec.
- **Session** is the return value for the attached domain (Domain, TunnelName,
  ConfigPath under managed dir, public `https://` URL). Detach/Stop covered in
  the detach tree.

**Behaviors**

```text
# validation
Domain empty
  -> Attach error; no requirement that state.json exists

# resolve defaults
TunnelName empty -> DefaultTunnelName
LocalURL empty   -> http://127.0.0.1:6321
ConfigDir empty  -> DefaultConfigDir()  # tests always inject ConfigDir

# critical section (per managed tunnel dir)
mkdir ManagedTunnelDir
flock exclusive on dir/lock
ensureTunnelForSession(runner, name) -> tunnelID, credFile
RouteDNS(runner, name, domain)
LoadTunnelState; Hosts[domain] = {Service: localURL, OwnerPID: ...}
state.TunnelName / TunnelID / CredentialsFile set
BuildConfigFromState -> WriteConfig(dir/config.yml)
# connector (fake Runner != nil):
#   rewrite config; Exec tunnel run when first attach or ingress changed
SaveTunnelState
unlock
return Session for domain

# multi-attach same TunnelName + ConfigDir
attach A then B
  -> Hosts has A and B; config both hostnames + 404 last; run count increases

# update same host
attach A:url1 then A:url2
  -> Hosts[A].Service = url2; run count increases on ingress change
```

## Decision Tree

```text
tests/cloudflare-tunnel-attach/
├── DOCTEST.md
├── SETUP.md
├── validation/                          # input rejection before registry work
│   └── domain-empty-errors/             # Domain "" → error
├── defaults/                            # empty optional fields resolve correctly
│   ├── empty-tunnel-name/               # TunnelName "" → DefaultTunnelName
│   └── empty-local-url/                 # LocalURL "" → :6321 service in state
└── lifecycle/                           # successful Attach with fake runner
    ├── first-attach/                    # empty registry → single host
    │   ├── registry-and-config/         # 1 host; config host+404; under managed dir
    │   ├── dns-route/                   # Exec includes route dns + hostname
    │   └── connector-run/               # tunnel run Exec count ≥ 1
    ├── second-host/                     # second distinct hostname same tunnel
    │   └── merge-and-restart/           # Hosts A+B; 404 last; run count ↑
    └── update-url/                      # same hostname, new LocalURL
        └── service-change-restart/      # Service rewritten; run count ↑
```

### Parameter significance (high → low)

1. **Outcome class** — validation error vs defaults vs lifecycle success
2. **Lifecycle mutation** — first attach vs second host merge vs same-host URL update
3. **Observable focus (first attach)** — registry/config shape vs DNS args vs run count
4. **Fixture values** — concrete hostnames, ports, tunnel names (leaf constants)

## Test Index

| # | Leaf | Assert focus |
|---|------|--------------|
| 1 | `validation/domain-empty-errors` | empty Domain → non-nil error; no forced state file |
| 2 | `defaults/empty-tunnel-name` | Session/state TunnelName is `DefaultTunnelName` |
| 3 | `defaults/empty-local-url` | Hosts[domain].Service is `http://127.0.0.1:6321` |
| 4 | `lifecycle/first-attach/registry-and-config` | 1 host in state; config 2 ingress rules (host+404); `config.yml` under ManagedTunnelDir |
| 5 | `lifecycle/first-attach/dns-route` | fake calls include `route` + `dns` + domain |
| 6 | `lifecycle/first-attach/connector-run` | at least one `tunnel … run` Exec |
| 7 | `lifecycle/second-host/merge-and-restart` | Hosts A and B; both hostnames in config; 404 last; RunCount > after first |
| 8 | `lifecycle/update-url/service-change-restart` | Hosts[A].Service is new URL; RunCount increases on change |

## How to Run

```sh
cd <go-pkgs-module-root>   # this module; package is local

doctest vet ./tests/cloudflare-tunnel-attach
doctest test ./tests/cloudflare-tunnel-attach
doctest test -v ./tests/cloudflare-tunnel-attach
```

**Coverage backfill / GREEN expected:** `Attach` / `AttachOptions` present in
`cloudflare/attach.go`. Leaves exercise existing correct behavior.

### API surface under test

```go
type AttachOptions struct {
    Domain     string // required public hostname
    LocalURL   string // e.g. http://127.0.0.1:6321; default http://127.0.0.1:6321
    TunnelName string // empty → DefaultTunnelName
    ConfigDir  string // empty → DefaultConfigDir(); tests use t.TempDir()
    Log        io.Writer
    Runner     CommandRunner // nil = real cloudflared; non-nil = fake (tests)
    // OwnerPID optional; default os.Getpid()
}

// Attach merges hostname into managed registry and ensures connector.
func Attach(opts AttachOptions) (*Session, error)
```

### Out of scope

- Detach / Stop correctness (separate tree)
- StartSession rewrite coverage (separate tree)
- spl CLI wiring
- DNS delete
- Real OS process StartProcess path (tests inject Runner only)

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

// attachStep is one Attach call in a shared ConfigDir/Runner sequence.
type attachStep struct {
	Domain   string
	LocalURL string
}

// Request selects Attach inputs (single call or multi-step sequence).
type Request struct {
	// Domain / LocalURL used when Sequence is empty (single Attach).
	Domain   string
	LocalURL string

	// TunnelName empty → production default (DefaultTunnelName).
	TunnelName string

	// ConfigDir is the managed cloudflared root. Empty in leaf → root Setup
	// fills t.TempDir(). Always non-empty before Run for lifecycle leaves.
	ConfigDir string

	// Sequence, when non-empty, runs multiple Attach calls sharing ConfigDir,
	// TunnelName, and the same fake runner (second-host / update-url leaves).
	Sequence []attachStep

	// ExpectError: leaf expects Attach to return a non-nil error.
	ExpectError bool

	// HomeDir is a sandboxed HOME so DefaultConfigDir() resolves under temp.
	// Root Setup sets this and t.Setenv("HOME", HomeDir).
	HomeDir string

	DecisionPath []string
}

// Response captures Attach results and fake-runner observations.
type Response struct {
	// Session is the last successful Attach session (nil on error).
	Session *cloudflare.Session

	// Sessions holds every successful session in a multi-step Sequence.
	Sessions []*cloudflare.Session

	// State is LoadTunnelState after the last successful Attach.
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
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	steps := req.Sequence
	if len(steps) == 0 {
		steps = []attachStep{{Domain: req.Domain, LocalURL: req.LocalURL}}
	}

	fake := newFakeCloudflared(t, req.HomeDir)
	resp := &Response{Runner: fake}

	var lastErr error
	for i, step := range steps {
		opts := cloudflare.AttachOptions{
			Domain:     step.Domain,
			LocalURL:   step.LocalURL,
			TunnelName: req.TunnelName,
			ConfigDir:  req.ConfigDir,
			Runner:     fake,
			Log:        os.Stderr,
		}
		sess, err := cloudflare.Attach(opts)
		lastErr = err
		if err != nil {
			// On validation / early failure, still surface runner stats.
			resp.RunCount = fake.runCount()
			resp.RouteDNSCount = fake.routeDNSCount()
			return resp, err
		}
		resp.Session = sess
		resp.Sessions = append(resp.Sessions, sess)
		t.Logf("attach step %d ok domain=%q tunnel=%q config=%q", i, sess.Domain, sess.TunnelName, sess.ConfigPath)
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

	st, lerr := cloudflare.LoadTunnelState(dir)
	if lerr != nil {
		resp.RunCount = fake.runCount()
		resp.RouteDNSCount = fake.routeDNSCount()
		return resp, lerr
	}
	resp.State = st

	if data, rerr := os.ReadFile(resp.ConfigPath); rerr == nil {
		var cfg cloudflare.Config
		if yerr := yaml.Unmarshal(data, &cfg); yerr == nil {
			resp.Config = &cfg
		}
	}

	resp.RunCount = fake.runCount()
	resp.RouteDNSCount = fake.routeDNSCount()
	return resp, lastErr
}

// fakeCloudflared soft-succeeds cloudflared CLI shapes used by
// ensureTunnelForSession / RouteDNS / Attach connector start.
//
// On tunnel create it writes a dummy credentials JSON under
// $HOME/.cloudflared/<uuid>.json (Home sandboxed by root Setup) so
// FindTunnelIDAndCreds and create-output parsing both succeed.
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

	// Status / bare look-path style.
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
			// Still emit a UUID so parseUUIDFromInfo can recover if list lied.
			_ = f.writeCredsLocked()
			f.created = true
		}
		msg := fmt.Sprintf("NAME: %s\nID: %s\n", f.tunnelName, f.tunnelID)
		return []byte(msg), nil
	}

	// tunnel route dns --overwrite-dns <tunnel> <hostname>
	if containsArg(args, "route") && containsArg(args, "dns") {
		return []byte("Added CNAME " + joined), nil
	}

	// tunnel … run …
	if containsArg(args, "run") {
		return []byte("fake connector running\n"), nil
	}

	// Soft-succeed unknown cloudflared shapes so Attach is not blocked by extras.
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

func (f *fakeCloudflared) hasRouteDNSFor(hostname string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.calls {
		if containsArg(c.Args, "route") && containsArg(c.Args, "dns") && containsArg(c.Args, hostname) {
			return true
		}
	}
	return false
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

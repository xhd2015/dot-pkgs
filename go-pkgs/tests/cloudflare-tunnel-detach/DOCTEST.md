# Cloudflare managed tunnel Detach / Session.Stop (owned coverage)

Coverage-backfill doctests for releasing one hostname from a managed named-tunnel
registry without killing sibling hosts; stop the connector only when the last
host is gone; best-effort DNS delete. **No** real network or real `cloudflared`
binary.

Package under test:

`github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`

Run from this **go-pkgs module root** (no `replace` needed).

Depends on managed helpers and **Attach (GREEN)**: `ManagedTunnelDir`,
`LoadTunnelState`, `BuildConfigFromState` / config parse, `Attach` + fake
`CommandRunner`. **`Detach` and managed `Session.Stop` are present** — GREEN
expected.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Participants**

- **Detach** is the multi-host managed release entrypoint: given Domain (+
  TunnelName, ConfigDir, optional Runner / DNSDeleter), it removes one host
  from the on-disk registry under flock and rewrites config / connector.
- **Session.Stop** for sessions created via **Attach** should call Detach
  (or equivalent managed path) using Domain + tunnel identity + runner +
  dnsDeleter carried on the Session.
- **ConfigDir** is the cloudflared home root (tests inject `t.TempDir()`).
  Managed layout: `{ConfigDir}/managed-tunnels/<TunnelNameSafe>/`.
- **Managed tunnel directory** holds `lock`, `state.json`, and `config.yml`.
- **TunnelState / Hosts** map public hostnames to local services. Detach
  deletes `Hosts[domain]` only; siblings stay.
- **CommandRunner** is the same injectable fake as the attach tree. Soft-succeeds
  `tunnel list|create|info|route dns|run` and materializes dummy credentials
  under sandboxed `HOME`. Each `Exec(…, "run", …)` counts as connector
  start/restart.
- **Connector** logical-up after partial detach: `state.ConnectorPID > 0`
  and/or an additional `run` Exec with remaining hosts. Last host detach sets
  `ConnectorPID == 0` and does not leave a logical connector alive.
- **DNSDeleter** is optional best-effort cleanup. Delete failures are logged;
  Detach / Stop still return nil for DNS-only failures.

**Behaviors**

```text
# partial detach (siblings remain)
Attach A, Attach B; Stop(A) / Detach(A)
  -> Hosts only B; config only B + 404 last
  -> connector still up (ConnectorPID > 0 OR run re-invoked)

# last host
Attach A, Attach B; Stop(A); Stop(B)
  -> Hosts empty; ConnectorPID == 0; no host rules (404 only OK)

Attach A; Stop(A)
  -> Hosts empty; ConnectorPID == 0

# DNS best-effort
Detach/Stop with DNSDeleter error
  -> host removed from registry; API error is nil (DNS failure ignored)

# missing host
Detach domain not in Hosts
  -> no-op success (preferred); registry unchanged for other hosts
```

## Decision Tree

```text
tests/cloudflare-tunnel-detach/
├── DOCTEST.md
├── SETUP.md
├── partial-detach/                      # Stop one host; siblings stay up
│   └── siblings-remain/                 # A+B, Stop(A) → only B; connector up
├── last-host/                           # final host tears down connector
│   ├── single-host-stops/               # Attach A, Stop(A) → empty; PID 0
│   └── second-stop-clears/              # A+B, Stop A then B → empty; PID 0
├── dns/                                 # best-effort DeleteDNS
│   └── delete-fails-ok/                 # DNS error; Stop/Detach still nil
└── missing-host/                        # domain not in registry
    └── detach-noop/                     # Detach missing → success no-op
```

### Parameter significance (high → low)

1. **Outcome class** — partial detach vs last-host teardown vs DNS soft-fail vs missing no-op
2. **Host cardinality before release** — two hosts (sibling path) vs one host
3. **Stop sequence depth** — single Stop vs Stop then Stop (last host after partial)
4. **API path** — Session.Stop (managed attach) vs package `Detach` (missing / DNS inject)
5. **Fixture values** — concrete hostnames, ports, tunnel names (leaf constants)

## Test Index

| # | Leaf | Assert focus |
|---|------|--------------|
| 1 | `partial-detach/siblings-remain` | After A+B and Stop(A): Hosts only B; config B+404; connector still up |
| 2 | `last-host/single-host-stops` | Attach A, Stop(A): Hosts empty; ConnectorPID 0 |
| 3 | `last-host/second-stop-clears` | A+B then Stop A then B: Hosts empty; ConnectorPID 0; no host rules |
| 4 | `dns/delete-fails-ok` | Failing DNSDeleter; Detach/Stop returns nil; host removed; delete attempted |
| 5 | `missing-host/detach-noop` | Detach unknown domain succeeds; existing hosts unchanged |

## How to Run

```sh
cd <go-pkgs-module-root>   # this module; package is local

doctest vet ./tests/cloudflare-tunnel-detach
doctest test ./tests/cloudflare-tunnel-detach
doctest test -v ./tests/cloudflare-tunnel-detach
```

**Coverage backfill / GREEN expected:** `Detach` / `DetachOptions` and managed
`Session.Stop` present in `cloudflare/detach.go` / `session.go`.

### API surface under test

```go
type DetachOptions struct {
    Domain     string
    TunnelName string
    ConfigDir  string
    Log        io.Writer
    Runner     CommandRunner
    DNSDeleter DNSDeleter // optional; best-effort
}

// Detach removes one hostname from the managed tunnel registry.
func Detach(opts DetachOptions) error

// Session.Stop for managed attach sessions:
// - flock, remove Hosts[Domain]
// - rewrite config.yml from remaining hosts
// - if hosts remain: restart connector (runner Exec run) with remaining hosts
// - if hosts empty: stop real process if any; set ConnectorPID=0
// - best-effort DeleteDNS(domain); DNS errors logged; Stop returns nil for DNS failures
```

### Out of scope

- StartSession rewrite coverage (separate tree)
- spl CLI wiring
- Cloudflare account prune / real DNS API
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

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
	"gopkg.in/yaml.v3"
)

// attachStep is one Attach call sharing ConfigDir / TunnelName / Runner.
type attachStep struct {
	Domain   string
	LocalURL string
}

// stopStep releases one hostname after Attach setup.
type stopStep struct {
	Domain string
	// ViaDetach forces package Detach instead of Session.Stop.
	// Used for missing-host and DNS inject leaves.
	ViaDetach bool
}

// Request selects Attach setup then Stop/Detach actions.
type Request struct {
	// TunnelName empty → production default (DefaultTunnelName).
	TunnelName string

	// ConfigDir is the managed cloudflared root. Empty in leaf → root Setup
	// fills t.TempDir().
	ConfigDir string

	// AttachSequence runs Attach calls before any Stop/Detach (may be empty
	// for pure Detach missing-host against empty registry).
	AttachSequence []attachStep

	// StopSequence runs Session.Stop or Detach in order after Attach.
	StopSequence []stopStep

	// FailDNS makes the injectable DNSDeleter return an error on DeleteHostname.
	FailDNS bool

	// ExpectError: leaf expects final Stop/Detach (or Attach) to return error.
	ExpectError bool

	// HomeDir is a sandboxed HOME so DefaultConfigDir() resolves under temp.
	// Root Setup sets this and t.Setenv("HOME", HomeDir).
	HomeDir string

	DecisionPath []string
}

// Response captures post-condition state after Attach + Stop/Detach.
type Response struct {
	// Sessions holds successful Attach sessions keyed by order.
	Sessions []*cloudflare.Session

	// SessionByDomain maps Domain → last Attach session for that host.
	SessionByDomain map[string]*cloudflare.Session

	// State is LoadTunnelState after the last Stop/Detach (or after Attach
	// if no stop steps).
	State *cloudflare.TunnelState

	// Config is the parsed managed config.yml after the last action.
	Config *cloudflare.Config

	// ConfigPath is filepath.Join(ManagedDir, "config.yml") when present.
	ConfigPath string

	// ManagedDir is ManagedTunnelDir(ConfigDir, resolved tunnel name).
	ManagedDir string

	// Runner is the fake used for this Run.
	Runner *fakeCloudflared

	// RunCount is how many Exec calls included a "run" subcommand token
	// (across Attach + Stop/Detach restarts).
	RunCount int

	// RunCountAfterAttach is run count immediately after Attach sequence,
	// before Stop/Detach (for partial-detach restart checks).
	RunCountAfterAttach int

	// RouteDNSCount is how many Exec calls looked like tunnel route dns.
	RouteDNSCount int

	// DNSDeleteCount is how many DeleteHostname calls the fake deleter saw.
	DNSDeleteCount int

	// LastStopErr is the error from the last Stop/Detach step (nil if none).
	LastStopErr error

	// DNS is the fake DNS deleter when constructed.
	DNS *fakeDNSDeleter
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	fake := newFakeCloudflared(t, req.HomeDir)
	dns := &fakeDNSDeleter{fail: req.FailDNS}
	resp := &Response{
		Runner:          fake,
		DNS:             dns,
		SessionByDomain: map[string]*cloudflare.Session{},
	}

	name := req.TunnelName
	if strings.TrimSpace(name) == "" {
		name = cloudflare.DefaultTunnelName
	}

	// --- Attach setup ---
	var lastErr error
	for i, step := range req.AttachSequence {
		opts := cloudflare.AttachOptions{
			Domain:     step.Domain,
			LocalURL:   step.LocalURL,
			TunnelName: req.TunnelName,
			ConfigDir:  req.ConfigDir,
			Runner:     fake,
			Log:        os.Stderr,
		}
		// Prefer wiring DNSDeleter onto Attach→Session when available so
		// Session.Stop can best-effort delete DNS. DNS leaf uses Detach with
		// DNSDeleter directly.
		sess, err := cloudflare.Attach(opts)
		lastErr = err
		if err != nil {
			resp.RunCount = fake.runCount()
			resp.RouteDNSCount = fake.routeDNSCount()
			resp.DNSDeleteCount = dns.deleteCount()
			return resp, err
		}
		resp.Sessions = append(resp.Sessions, sess)
		resp.SessionByDomain[step.Domain] = sess
		t.Logf("attach step %d ok domain=%q tunnel=%q config=%q", i, sess.Domain, sess.TunnelName, sess.ConfigPath)
		if strings.TrimSpace(sess.TunnelName) != "" {
			name = sess.TunnelName
		}
	}

	resp.RunCountAfterAttach = fake.runCount()

	// Resolve managed dir for post-conditions.
	dir, derr := cloudflare.ManagedTunnelDir(req.ConfigDir, name)
	if derr != nil {
		resp.RunCount = fake.runCount()
		resp.RouteDNSCount = fake.routeDNSCount()
		resp.DNSDeleteCount = dns.deleteCount()
		return resp, fmt.Errorf("ManagedTunnelDir: %w", derr)
	}
	resp.ManagedDir = dir
	resp.ConfigPath = filepath.Join(dir, "config.yml")

	// --- Stop / Detach sequence ---
	for i, step := range req.StopSequence {
		var err error
		if step.ViaDetach {
			err = cloudflare.Detach(cloudflare.DetachOptions{
				Domain:     step.Domain,
				TunnelName: name,
				ConfigDir:  req.ConfigDir,
				Runner:     fake,
				DNSDeleter: dns,
				Log:        os.Stderr,
			})
			t.Logf("detach step %d domain=%q err=%v", i, step.Domain, err)
		} else {
			sess := resp.SessionByDomain[step.Domain]
			if sess == nil {
				// No Attach session for domain: fall back to Detach so leaves
				// can still express "release this host" without a session.
				err = cloudflare.Detach(cloudflare.DetachOptions{
					Domain:     step.Domain,
					TunnelName: name,
					ConfigDir:  req.ConfigDir,
					Runner:     fake,
					DNSDeleter: dns,
					Log:        os.Stderr,
				})
				t.Logf("stop step %d domain=%q (no session → Detach) err=%v", i, step.Domain, err)
			} else {
				err = sess.Stop()
				t.Logf("stop step %d domain=%q Session.Stop err=%v", i, step.Domain, err)
			}
		}
		resp.LastStopErr = err
		lastErr = err
		if err != nil && !req.ExpectError {
			// Still snapshot state for debugging, then return.
			_ = loadStateAndConfig(resp, dir)
			resp.RunCount = fake.runCount()
			resp.RouteDNSCount = fake.routeDNSCount()
			resp.DNSDeleteCount = dns.deleteCount()
			return resp, err
		}
	}

	if err := loadStateAndConfig(resp, dir); err != nil {
		resp.RunCount = fake.runCount()
		resp.RouteDNSCount = fake.routeDNSCount()
		resp.DNSDeleteCount = dns.deleteCount()
		return resp, err
	}

	resp.RunCount = fake.runCount()
	resp.RouteDNSCount = fake.routeDNSCount()
	resp.DNSDeleteCount = dns.deleteCount()
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

// fakeDNSDeleter records DeleteHostname calls; optionally fails.
type fakeDNSDeleter struct {
	mu      sync.Mutex
	fail    bool
	calls   []string
	lastErr error
}

func (d *fakeDNSDeleter) DeleteHostname(hostname string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = append(d.calls, hostname)
	if d.fail {
		d.lastErr = fmt.Errorf("simulated DNS delete failure for %s", hostname)
		return d.lastErr
	}
	return nil
}

func (d *fakeDNSDeleter) deleteCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.calls)
}

func (d *fakeDNSDeleter) hasDeleteFor(hostname string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, h := range d.calls {
		if h == hostname {
			return true
		}
	}
	return false
}

// fakeCloudflared soft-succeeds cloudflared CLI shapes used by
// ensureTunnelForSession / RouteDNS / Attach / Detach connector start.
// Same soft contracts as tests/cloudflare-tunnel-attach.
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

// connectorStillUp reports logical connector presence after partial detach.
// Accept either positive ConnectorPID or an additional run after Attach.
func connectorStillUp(resp *Response) bool {
	if resp == nil {
		return false
	}
	if resp.State != nil && resp.State.ConnectorPID > 0 {
		return true
	}
	// Partial detach may restart via runner without rewriting PID the same way.
	if resp.RunCount > resp.RunCountAfterAttach {
		return true
	}
	return false
}
```

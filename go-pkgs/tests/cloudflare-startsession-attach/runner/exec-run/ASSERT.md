## Expected

- StartSession succeeds (`err == nil`) — no panic.
- `RunCount ≥ 1` (at least one Exec whose args include the token `run`).
- Managed registry has Hosts[`a.example.com`] (StartSession used Attach path).
- Soft: if any run call has `--config`, the next arg should be the managed
  `ConfigPath` under ManagedDir.

## Side Effects

- Fake connector only; no OS process required in runner mode.

## Errors

- None.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("StartSession error: %v", err)
	}
	if resp == nil || resp.Session == nil {
		t.Fatal("nil response or Session")
	}
	if resp.RunCount < 1 {
		t.Fatalf("RunCount = %d, want ≥ 1 (StartSession must Exec tunnel run via Runner)", resp.RunCount)
	}

	const host = "a.example.com"
	if resp.State == nil {
		t.Fatal("nil State (managed state.json missing — StartSession must Attach for multi-host safety)")
	}
	if resp.State.Hosts[host] == nil {
		t.Fatalf("Hosts missing %q; Hosts=%v (StartSession must merge into managed registry)", host, resp.State.Hosts)
	}

	// Soft: if --config appears on a run call, it should reference managed config.
	if resp.Runner != nil && resp.ManagedDir != "" {
		wantConfig := resp.ConfigPath
		if wantConfig == "" {
			wantConfig = filepath.Join(resp.ManagedDir, "config.yml")
		}
		foundConfig := false
		var configPath string
		resp.Runner.mu.Lock()
		for _, c := range resp.Runner.calls {
			if !containsArg(c.Args, "run") {
				continue
			}
			for i, a := range c.Args {
				if a == "--config" && i+1 < len(c.Args) {
					foundConfig = true
					configPath = c.Args[i+1]
				}
			}
		}
		resp.Runner.mu.Unlock()
		if foundConfig {
			if filepath.Clean(configPath) != filepath.Clean(wantConfig) &&
				filepath.Dir(filepath.Clean(configPath)) != filepath.Clean(resp.ManagedDir) {
				t.Fatalf("run --config = %q, want under ManagedDir %q (or %q)", configPath, resp.ManagedDir, wantConfig)
			}
		} else {
			t.Logf("note: no run call carried --config (implementer may use alternate start shape); RunCount=%d", resp.RunCount)
		}
	}
}
```

## Expected

- Attach succeeds.
- `RunCount ≥ 1` (at least one Exec whose args include the token `run`).
- Prefer that a run call also includes `--config` pointing at the managed
  `config.yml` when Session/ConfigPath is known (soft check: if any run call
  has `--config`, the next arg should be ConfigPath).

## Side Effects

- Fake connector only; no process group / PID required in runner mode.

## Errors

- None.

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Attach error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.RunCount < 1 {
		t.Fatalf("RunCount = %d, want ≥ 1 (first attach must start connector)", resp.RunCount)
	}
	// Soft: if --config appears on a run call, it should reference managed config.
	if resp.Runner != nil && resp.ConfigPath != "" {
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
		if foundConfig && configPath != resp.ConfigPath {
			t.Fatalf("run --config = %q, want %q", configPath, resp.ConfigPath)
		}
		if !foundConfig {
			t.Logf("note: no run call carried --config (implementer may use alternate start shape); RunCount=%d", resp.RunCount)
		}
	}
}
```

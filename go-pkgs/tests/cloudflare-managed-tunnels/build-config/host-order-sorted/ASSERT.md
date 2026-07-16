## Expected

- Build succeeds for both Config and Config2.
- Host rules (non-empty hostnames) appear in alphabetical order:
  `a.example.com`, `m.example.com`, `z.example.com`.
- Final rule is still the 404 catch-all.
- Config and Config2 have identical hostname sequences (determinism).

## Side Effects

- None (pure build; Run calls BuildConfigFromState twice).

## Errors

- None.

## Exit Code

- 0

```go
import (
	"reflect"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("BuildConfigFromState error: %v", err)
	}
	if resp == nil || resp.Config == nil {
		t.Fatal("nil response or Config")
	}
	if resp.Config2 == nil {
		t.Fatal("nil Config2; Run must build twice for determinism")
	}

	wantHosts := []string{"a.example.com", "m.example.com", "z.example.com"}
	rules := hostRulesOnly(resp.Config)
	if len(rules) != len(wantHosts) {
		t.Fatalf("host rules len = %d, want %d", len(rules), len(wantHosts))
	}
	gotHosts := make([]string, len(rules))
	for i, r := range rules {
		gotHosts[i] = r.Hostname
	}
	if !reflect.DeepEqual(gotHosts, wantHosts) {
		t.Fatalf("host order = %v, want sorted %v", gotHosts, wantHosts)
	}

	// Full ingress hostnames: sorted hosts then empty catch-all.
	full := ingressHostnames(resp.Config)
	wantFull := append(append([]string{}, wantHosts...), "")
	if !reflect.DeepEqual(full, wantFull) {
		t.Fatalf("ingress hostnames = %v, want %v", full, wantFull)
	}

	full2 := ingressHostnames(resp.Config2)
	if !reflect.DeepEqual(full, full2) {
		t.Fatalf("non-deterministic order: first=%v second=%v", full, full2)
	}

	last, ok := lastIngress(resp.Config)
	if !ok || last.Service != "http_status:404" || last.Hostname != "" {
		t.Fatalf("last rule = %+v, want empty host http_status:404", last)
	}
}
```

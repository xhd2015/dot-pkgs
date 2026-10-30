package portmap

import (
	"context"
	"fmt"
	"testing"
)

type fakeDNS struct {
	names    []string
	conflict bool
}

func (d *fakeDNS) VerifyDomain(context.Context, string) error { return nil }
func (d *fakeDNS) EnsureTunnelCNAME(_ context.Context, host, id string) error {
	if d.conflict {
		return fmt.Errorf("hostname conflict")
	}
	d.names = append(d.names, host)
	return nil
}
func TestManagedPublisherKeepsSiblingConnector(t *testing.T) {
	dns := &fakeDNS{}
	var attached, detached []string
	p := &ManagedPublisher{TunnelName: "eluc-ports-test", DNS: dns,
		ensureTunnelFn: func(context.Context) (string, error) { return "tunnel-id", nil },
		attachFn:       func(_ context.Context, h, o string) error { attached = append(attached, h); return nil },
		detachFn:       func(_ context.Context, h string) error { detached = append(detached, h); return nil }}
	ctx := context.Background()
	for _, h := range []string{"dev-port-1.example.com", "dev-port-2.example.com"} {
		if err := p.Publish(ctx, h, "http://127.0.0.1:9999"); err != nil {
			t.Fatal(err)
		}
	}
	if len(attached) != 1 || attached[0] != "*.example.com" || len(dns.names) != 2 {
		t.Fatalf("attach=%v dns=%v", attached, dns.names)
	}
	if err := p.Unpublish(ctx, "dev-port-1.example.com"); err != nil {
		t.Fatal(err)
	}
	if len(detached) != 0 {
		t.Fatal("removing sibling restarted connector")
	}
	if err := p.Unpublish(ctx, "dev-port-2.example.com"); err != nil {
		t.Fatal(err)
	}
	if len(detached) != 1 || detached[0] != "*.example.com" {
		t.Fatal(detached)
	}
}
func TestManagedPublisherRefusesConflictBeforeIngress(t *testing.T) {
	p := &ManagedPublisher{TunnelName: "eluc-ports-test", DNS: &fakeDNS{conflict: true}, ensureTunnelFn: func(context.Context) (string, error) { return "id", nil }, attachFn: func(context.Context, string, string) error { t.Fatal("attached despite conflict"); return nil }}
	if p.Publish(context.Background(), "p.example.com", "http://127.0.0.1:1") == nil {
		t.Fatal("expected conflict")
	}
}

func TestManagedPublisherStepsReportEachPublishStage(t *testing.T) {
	var steps []string
	p := &ManagedPublisher{TunnelName: "eluc-ports-test", DNS: &fakeDNS{},
		ensureTunnelFn: func(context.Context) (string, error) { return "tunnel-id", nil },
		attachFn:       func(context.Context, string, string) error { return nil },
		Step:           func(format string, args ...any) { steps = append(steps, fmt.Sprintf(format, args...)) }}
	ctx := context.Background()
	for _, h := range []string{"dev-port-1.example.com", "dev-port-2.example.com"} {
		if err := p.Publish(ctx, h, "http://127.0.0.1:9999"); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{
		"verifying Cloudflare domain example.com…",
		"ensuring dedicated tunnel eluc-ports-test…",
		"publishing DNS dev-port-1.example.com…",
		"starting tunnel connector…",
		// The sibling port reuses the live connector: no second start line.
		"verifying Cloudflare domain example.com…",
		"ensuring dedicated tunnel eluc-ports-test…",
		"publishing DNS dev-port-2.example.com…",
	}
	if len(steps) != len(want) {
		t.Fatalf("steps=%v", steps)
	}
	for i := range want {
		if steps[i] != want[i] {
			t.Fatalf("step %d: got %q, want %q (all: %v)", i, steps[i], want[i], steps)
		}
	}
}

func TestManagedPublisherStepIsOptional(t *testing.T) {
	p := &ManagedPublisher{TunnelName: "eluc-ports-test", DNS: &fakeDNS{},
		ensureTunnelFn: func(context.Context) (string, error) { return "id", nil },
		attachFn:       func(context.Context, string, string) error { return nil }}
	if err := p.Publish(context.Background(), "dev-port-1.example.com", "http://127.0.0.1:9999"); err != nil {
		t.Fatal(err)
	}
}

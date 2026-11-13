package portmap

import (
	"context"
	"testing"
)

func TestReconcileSharedIngressWithoutRemovingSiblings(t *testing.T) {
	var attaches, detaches int
	p := &ManagedPublisher{TunnelName: "test", DNS: &fakeDNS{},
		ensureTunnelFn: func(context.Context) (string, error) { return "id", nil },
		attachFn: func(_ context.Context, host, origin string) error {
			if host != "*.example.com" || origin != "http://127.0.0.1:1234" {
				t.Fatalf("%s %s", host, origin)
			}
			attaches++
			return nil
		},
		detachFn: func(context.Context, string) error { detaches++; return nil },
	}
	ctx := context.Background()
	for _, h := range []string{"one.example.com", "two.example.com"} {
		if err := p.Publish(ctx, h, "http://127.0.0.1:1234"); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.Reconcile(ctx, "http://127.0.0.1:1234"); err != nil {
		t.Fatal(err)
	}
	if attaches != 2 || detaches != 0 || len(p.hosts["*.example.com"]) != 2 {
		t.Fatalf("attaches=%d detaches=%d hosts=%v", attaches, detaches, p.hosts)
	}
}

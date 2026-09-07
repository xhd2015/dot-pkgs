package singboxtun

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchExcludeRefreshRestartsOnChange(t *testing.T) {
	old := lookupHostIPv4
	var flip atomic.Bool
	lookupHostIPv4 = func(host string) []net.IP {
		if flip.Load() {
			return []net.IP{net.ParseIP("9.9.9.9")}
		}
		return []net.IP{net.ParseIP("8.8.4.4")}
	}
	defer func() { lookupHostIPv4 = old }()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "run.json")
	if err := os.WriteFile(cfgPath, []byte(`{"ok":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	hosts := []string{"edge.example"}
	bundle := &tunRunBundle{
		configPath:   cfgPath,
		excludeHosts: hosts,
		excludeCIDRs: resolveExcludeHostCIDRs(hosts),
		buildOpts: &BuildConfigOptions{
			LocalSocksPort:    11080,
			ProxyHost:         "hub.example.com",
			HttpOnly:          true,
			DNSHijack:         true,
			SkipBindInterface: true,
			InitialUseProxy:   true,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	restartCh := make(chan struct{}, 1)
	go watchExcludeRefresh(ctx, bundle, 20*time.Millisecond, restartCh)

	select {
	case <-time.After(200 * time.Millisecond):
		// no restart yet
	case <-restartCh:
		t.Fatal("unexpected restart before IP change")
	}

	flip.Store(true)
	select {
	case <-restartCh:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("expected restart after exclude /32 change")
	}
}

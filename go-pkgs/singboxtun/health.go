package singboxtun

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

const (
	upstreamHealthInterval = time.Second
	upstreamSocksTimeout   = 2 * time.Second
)

// upstreamReachable is overridden in tests.
var upstreamReachable = defaultUpstreamReachable

func defaultUpstreamReachable(port int) bool {
	// Real SOCKS5 CONNECT: a listening -D port with a dead reverse splice still
	// accepts TCP but fails CONNECT (hotspot / Wi‑Fi flap). Tests override
	// upstreamReachable when they only need a TCP stub.
	return probeSocks5Connect(fmt.Sprintf("127.0.0.1:%d", port), upstreamSocksTimeout)
}

// ProbeUpstreamProxy reports whether the local SOCKS upstream can complete CONNECT.
func ProbeUpstreamProxy(port int) bool {
	return upstreamReachable(port)
}

func probeSocks5Connect(socksAddr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", socksAddr, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		return false
	}
	var greet [2]byte
	if _, err := io.ReadFull(conn, greet[:]); err != nil || greet[0] != 0x05 || greet[1] != 0x00 {
		return false
	}
	req := []byte{0x05, 0x01, 0x00, 0x01, 1, 1, 1, 1, 0, 0}
	binary.BigEndian.PutUint16(req[8:], 443)
	if _, err := conn.Write(req); err != nil {
		return false
	}
	var resp [10]byte
	if _, err := io.ReadFull(conn, resp[:]); err != nil {
		return false
	}
	return resp[0] == 0x05 && resp[1] == 0x00
}

// switchWebOutbound is overridden in tests.
var switchWebOutbound = defaultSwitchWebOutbound

func defaultSwitchWebOutbound(outbound string) error {
	body, err := json.Marshal(map[string]string{"name": outbound})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("http://%s/proxies/%s", clashAPIListen, webSelectorTag), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("clash api PUT /proxies/%s: HTTP %d", webSelectorTag, resp.StatusCode)
	}
	return nil
}

// WebHealthOptions configures http-only selector health behavior.
type WebHealthOptions struct {
	// NoDirectFallback keeps the selector on proxy when SOCKS is down (no flip
	// to direct). Use on DNS-polluted hotspots where direct is worse than fail.
	NoDirectFallback bool
}

// StartWebOutboundHealthMonitor toggles the web selector between proxy and direct.
func StartWebOutboundHealthMonitor(socksPort int) func() {
	return StartWebOutboundHealthMonitorOpts(socksPort, WebHealthOptions{})
}

// StartWebOutboundHealthMonitorOpts is StartWebOutboundHealthMonitor with options.
func StartWebOutboundHealthMonitorOpts(socksPort int, opts WebHealthOptions) func() {
	ctx, cancel := context.WithCancel(context.Background())
	var usingProxy atomic.Bool
	reachable := ProbeUpstreamProxy(socksPort)
	usingProxy.Store(reachable)

	if opts.NoDirectFallback {
		_ = switchWebOutbound(proxyOutboundTag)
		if !reachable {
			fmt.Fprintf(os.Stderr, "warning: SOCKS upstream down; keeping web→proxy (no direct fallback)\n")
		}
	} else {
		initial := directOutboundTag
		if reachable {
			initial = proxyOutboundTag
		}
		_ = switchWebOutbound(initial)
	}

	var lastDownWarn time.Time
	go func() {
		ticker := time.NewTicker(upstreamHealthInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reachable := ProbeUpstreamProxy(socksPort)
				if opts.NoDirectFallback {
					if !reachable && time.Since(lastDownWarn) > 10*time.Second {
						fmt.Fprintf(os.Stderr, "warning: SOCKS upstream down; web stays on proxy (waiting for recycle)\n")
						lastDownWarn = time.Now()
					}
					if reachable {
						_ = switchWebOutbound(proxyOutboundTag)
					}
					continue
				}
				prev := usingProxy.Load()
				if reachable == prev {
					continue
				}
				usingProxy.Store(reachable)
				next := directOutboundTag
				if reachable {
					next = proxyOutboundTag
				}
				if err := switchWebOutbound(next); err != nil {
					fmt.Fprintf(os.Stderr, "warning: switch web outbound to %s: %v\n", next, err)
				} else if !reachable {
					fmt.Fprintf(os.Stderr, "warning: SOCKS upstream down; web→direct fallback\n")
				}
			}
		}
	}()
	return cancel
}
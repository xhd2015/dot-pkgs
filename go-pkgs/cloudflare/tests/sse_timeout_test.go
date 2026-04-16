// Test: Cloudflare/cloudflared idle timeout on SSE connections
//
// Evidence (2026-04-16):
//   An SSE server sends one event then goes silent (no heartbeat, no data).
//   Two clients connect: one directly to localhost, one via Cloudflare tunnel.
//
//   Results:
//     localhost_direct:  survived 5m30s (hit our max wait), connection was still alive
//     via_cloudflare:    died at 2m5s with "stream error: stream ID 1; INTERNAL_ERROR; received from peer"
//
//   cloudflared logs at disconnect time:
//     ERR error="context canceled" connIndex=0 event=1 ingressRule=0 originService=http://localhost:18767
//     ERR failed to serve incoming request error="Failed to proxy HTTP: context canceled"
//
// Conclusion:
//   Cloudflare/cloudflared enforces an ~2 minute idle timeout on HTTP/2 streams.
//   When no data flows for ~2min, the edge sends an INTERNAL_ERROR stream reset.
//   Fix: send SSE heartbeats (": heartbeat\n\n") every 15s to keep the stream active.
package tests

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

const (
	tunnelName = "test-http2-tunnel"
	localPort  = 18767
)

// DOMAIN_SUFFIX=your-domain.com SSE_TIMEOUT_TEST=1 go test -v -run TestSSETimeoutComparison -timeout 0 ./
func TestSSETimeoutComparison(t *testing.T) {
	if os.Getenv("SSE_TIMEOUT_TEST") == "" {
		t.Skip("set SSE_TIMEOUT_TEST=1 to run")
	}
	domainSuffix := os.Getenv("DOMAIN_SUFFIX")
	if domainSuffix == "" {
		t.Fatal("DOMAIN_SUFFIX env is required (e.g. DOMAIN_SUFFIX=your-domain.com)")
	}
	tunnelDomain := "test-http2." + domainSuffix

	maxWait := 5*time.Minute + 30*time.Second

	// Start a local SSE server on localPort
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "no flusher", 500)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		fmt.Fprintf(w, "event: meta\ndata: {\"status\":\"started\"}\n\n")
		flusher.Flush()

		<-r.Context().Done()
		t.Logf("[server] client disconnected")
	})

	srv := &http.Server{Addr: fmt.Sprintf(":%d", localPort), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	defer srv.Close()
	time.Sleep(500 * time.Millisecond)

	// Ensure cloudflared tunnel is set up and running
	credFile, err := ensureTunnel(t)
	if err != nil {
		t.Fatalf("ensureTunnel: %v", err)
	}
	if err := ensureRoute(t, tunnelDomain); err != nil {
		t.Fatalf("ensureRoute: %v", err)
	}
	stopTunnel, err := startTunnelRun(t, credFile, tunnelDomain)
	if err != nil {
		t.Fatalf("startTunnelRun: %v", err)
	}
	defer stopTunnel()

	// Wait for tunnel to be ready (can take 30-60s for connections to register)
	t.Log("Waiting for tunnel to be ready...")
	waitForEndpoint(t, fmt.Sprintf("https://%s/sse", tunnelDomain), 90*time.Second)

	localURL := fmt.Sprintf("http://localhost:%d/sse", localPort)
	cloudflareURL := fmt.Sprintf("https://%s/sse", tunnelDomain)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		t.Run("localhost_direct", func(t *testing.T) {
			testSSEGet(t, localURL, maxWait, false)
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		t.Run("via_cloudflare", func(t *testing.T) {
			testSSEGet(t, cloudflareURL, maxWait, true)
		})
	}()

	wg.Wait()
}

// ensureTunnel checks if the tunnel exists; if not, creates it.
// Returns the credentials file path.
func ensureTunnel(t *testing.T) (string, error) {
	t.Helper()

	logCmd(t, "cloudflared", "tunnel", "list", "--output", "json")
	tunnelID, err := findTunnelID(tunnelName)
	if err != nil {
		return "", fmt.Errorf("tunnel list: %w", err)
	}

	if tunnelID == "" {
		logCmd(t, "cloudflared", "tunnel", "create", tunnelName)
		cmd := exec.Command("cloudflared", "tunnel", "create", tunnelName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("tunnel create: %w", err)
		}
		tunnelID, err = findTunnelID(tunnelName)
		if err != nil || tunnelID == "" {
			return "", fmt.Errorf("tunnel created but ID not found")
		}
	}

	t.Logf("Tunnel %q exists (ID: %s)", tunnelName, tunnelID)

	home, _ := os.UserHomeDir()
	credFile := filepath.Join(home, ".cloudflared", tunnelID+".json")
	if _, err := os.Stat(credFile); err != nil {
		return "", fmt.Errorf("credentials file not found: %s", credFile)
	}
	t.Logf("Credentials file: %s", credFile)
	return credFile, nil
}

type tunnelEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func logCmd(t *testing.T, name string, args ...string) {
	t.Helper()
	t.Logf("$ %s %s", name, strings.Join(args, " "))
}

func findTunnelID(name string) (string, error) {
	out, err := exec.Command("cloudflared", "tunnel", "list", "--output", "json").Output()
	if err != nil {
		return "", err
	}
	var tunnels []tunnelEntry
	if err := json.Unmarshal(out, &tunnels); err != nil {
		return "", fmt.Errorf("parse tunnel list: %w", err)
	}
	for _, t := range tunnels {
		if t.Name == name {
			return t.ID, nil
		}
	}
	return "", nil
}

// ensureRoute checks if the DNS route exists; if not, creates it.
func ensureRoute(t *testing.T, tunnelDomain string) error {
	t.Helper()

	logCmd(t, "dig", "+short", "CNAME", tunnelDomain)
	out, err := exec.Command("dig", "+short", "CNAME", tunnelDomain).Output()
	if err == nil && strings.Contains(string(out), "cfargotunnel.com") {
		t.Logf("DNS route for %s already exists", tunnelDomain)
		return nil
	}

	logCmd(t, "cloudflared", "tunnel", "route", "dns", tunnelName, tunnelDomain)
	cmd := exec.Command("cloudflared", "tunnel", "route", "dns", tunnelName, tunnelDomain)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tunnel route dns: %w", err)
	}
	return nil
}

// startTunnelRun writes a temp config and starts the tunnel in background.
// Returns a stop function.
func startTunnelRun(t *testing.T, credFile string, tunnelDomain string) (func(), error) {
	t.Helper()

	configContent := fmt.Sprintf(`tunnel: %s
credentials-file: %s

ingress:
  - hostname: %s
    service: http://localhost:%d
  - service: http_status:404
`, tunnelName, credFile, tunnelDomain, localPort)

	configPath := filepath.Join(os.TempDir(), "config-test-http2-tunnel.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}
	t.Logf("Tunnel config: %s", configPath)

	logCmd(t, "cloudflared", "tunnel", "--config", configPath, "run")
	cmd := exec.Command("cloudflared", "tunnel", "--config", configPath, "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("tunnel run: %w", err)
	}
	t.Logf("Started cloudflared tunnel (PID %d)", cmd.Process.Pid)

	return func() {
		t.Logf("Stopping cloudflared tunnel (PID %d)...", cmd.Process.Pid)
		_ = cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	}, nil
}

// waitForEndpoint polls an HTTPS URL until it gets a 200 response or timeout.
func waitForEndpoint(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		resp, err := http.DefaultClient.Do(req)
		cancel()
		if err != nil {
			t.Logf("Waiting for tunnel... (error: %v)", err)
			time.Sleep(3 * time.Second)
			continue
		}
		status := resp.StatusCode
		resp.Body.Close()
		if status == 200 {
			t.Logf("Tunnel is ready (status %d)", status)
			return
		}
		t.Logf("Waiting for tunnel... (status %d)", status)
		time.Sleep(3 * time.Second)
	}
	t.Logf("Warning: tunnel may not be ready after %v", timeout)
}

func testSSEGet(t *testing.T, url string, maxWait time.Duration, useHTTP2 bool) {
	ctx, cancel := context.WithTimeout(context.Background(), maxWait)
	defer cancel()

	transport := &http.Transport{}
	if useHTTP2 {
		transport.TLSClientConfig = &tls.Config{}
		_ = http2.ConfigureTransport(transport)
	}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	start := time.Now()
	t.Logf("Connecting to %s ...", url)

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Connection FAILED after %v: %v", time.Since(start), err)
		return
	}
	defer resp.Body.Close()

	t.Logf("Connected, status=%d, proto=%s", resp.StatusCode, resp.Proto)

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		t.Logf("Non-200 body: %s", string(bodyBytes))
		return
	}

	readSSEBody(t, url, resp.Body, start)
}

func readSSEBody(t *testing.T, url string, body io.Reader, start time.Time) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		elapsed := time.Since(start).Round(time.Second)
		lineCount++
		if lineCount <= 20 || strings.HasPrefix(line, "event:") {
			t.Logf("[%v] %s", elapsed, line)
		}
	}

	elapsed := time.Since(start)
	if err := scanner.Err(); err != nil {
		t.Logf("Stream ERROR after %v: %v", elapsed, err)
	} else {
		t.Logf("Stream ended cleanly after %v (%d lines)", elapsed, lineCount)
	}
}

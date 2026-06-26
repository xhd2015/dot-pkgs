package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/xhd2015/less-flags"
)

const help = `Usage: http-proxy [OPTIONS]

Start a forward HTTP proxy with optional upstream proxy and fallback.

Options:
  --listen-port PORT         Port to listen on (default: 7821)
  --upstream-proxy URL       Upstream proxy URL (e.g. http://localhost:1087)
  --fallback-direct          Fall back to direct network access if upstream is unreachable
  -h, --help                 Show this help message
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "http-proxy: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var listenPort int
	var upstreamProxy string
	var fallbackDirect bool

	args, err := lessflags.Int("--listen-port", &listenPort).
		String("--upstream-proxy", &upstreamProxy).
		Bool("--fallback-direct", &fallbackDirect).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	_ = args

	if upstreamProxy == "" {
		return fmt.Errorf("--upstream-proxy is required")
	}

	if listenPort <= 0 {
		listenPort = 7821
	}

	proxyHost, proxyPort, err := net.SplitHostPort(extractHost(upstreamProxy))
	if err != nil {
		return fmt.Errorf("invalid --upstream-proxy URL %q: parse host:port: %w", upstreamProxy, err)
	}
	proxyAddr := net.JoinHostPort(proxyHost, proxyPort)
	if proxyPort == "" {
		proxyAddr = net.JoinHostPort(proxyHost, "80")
	}

	handler := NewProxyHandler(fallbackDirect)

	// Probe upstream synchronously before accepting requests so the first
	// proxied request uses the correct transport (avoids race with async setup).
	if tcpDial(proxyAddr, 100*time.Millisecond) {
		log.Printf("using upstream proxy %s", upstreamProxy)
		handler.SetTransport(newProxyTransport(upstreamProxy), proxyAddr)
	} else {
		log.Printf("upstream proxy unreachable, falling back to direct")
	}

	if fallbackDirect {
		go healthCheckLoop(handler, proxyAddr, upstreamProxy)
	}

	addr := fmt.Sprintf(":%d", listenPort)
	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, handler)
}

func healthCheckLoop(handler *ProxyHandler, proxyAddr string, upstreamURL string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastState := handler.usingProxy

	for range ticker.C {
		reachable := tcpDial(proxyAddr, 100*time.Millisecond)

		if reachable && !lastState {
			log.Printf("upstream proxy available, switching")
			handler.SetTransport(newProxyTransport(upstreamURL), proxyAddr)
			lastState = true
		} else if !reachable && lastState {
			log.Printf("upstream proxy unreachable, falling back to direct")
			handler.SetTransport(newDirectTransport(), "")
			lastState = false
		}
	}
}

func extractHost(rawURL string) string {
	if len(rawURL) == 0 {
		return ""
	}
	s := rawURL
	if len(s) > 7 && (s[:7] == "http://" || s[:7] == "HTTP://") {
		s = s[7:]
	} else if len(s) > 8 && (s[:8] == "https://" || s[:8] == "HTTPS://") {
		s = s[8:]
	}
	return s
}

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const connectDialTimeout = 2 * time.Second

type ProxyHandler struct {
	transport atomic.Pointer[http.Transport]

	mu             sync.RWMutex
	proxyAddr      string // "host:port" of upstream proxy, empty = direct
	usingProxy     bool
	fallbackDirect bool
}

func NewProxyHandler(fallbackDirect bool) *ProxyHandler {
	h := &ProxyHandler{fallbackDirect: fallbackDirect}
	h.transport.Store(newDirectTransport())
	return h
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		h.handleConnect(w, r)
		return
	}
	h.handleHTTP(w, r)
}

func (h *ProxyHandler) SetTransport(t *http.Transport, proxyAddr string) {
	h.transport.Store(t)
	h.mu.Lock()
	h.proxyAddr = proxyAddr
	h.usingProxy = proxyAddr != ""
	h.mu.Unlock()
}

func (h *ProxyHandler) getProxyAddr() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.proxyAddr
}

func (h *ProxyHandler) handleConnect(w http.ResponseWriter, r *http.Request) {
	destConn, via, err := h.dialConnectDest(r)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	log.Printf("CONNECT %s via %s", r.Host, via)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		destConn.Close()
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		destConn.Close()
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	go func() {
		io.Copy(destConn, clientConn)
		destConn.Close()
	}()
	io.Copy(clientConn, destConn)
	clientConn.Close()
}

func (h *ProxyHandler) dialConnectDest(r *http.Request) (net.Conn, string, error) {
	h.mu.RLock()
	proxyAddr := h.proxyAddr
	fallbackDirect := h.fallbackDirect
	h.mu.RUnlock()

	if proxyAddr != "" {
		conn, err := h.dialViaUpstreamProxy(proxyAddr, r.Host)
		if err == nil {
			return conn, "upstream proxy", nil
		}
		if !fallbackDirect {
			log.Printf("connect to upstream proxy %s: %v", proxyAddr, err)
			return nil, "upstream proxy", err
		}
		h.SetTransport(newDirectTransport(), "")
	}

	conn, err := net.DialTimeout("tcp", r.Host, connectDialTimeout)
	if err != nil {
		return nil, "direct", err
	}
	return conn, "direct", nil
}

func (h *ProxyHandler) dialViaUpstreamProxy(proxyAddr, targetHost string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", proxyAddr, connectDialTimeout)
	if err != nil {
		return nil, err
	}

	connectReq := "CONNECT " + targetHost + " HTTP/1.1\r\nHost: " + targetHost + "\r\n\r\n"
	if _, err := io.WriteString(conn, connectReq); err != nil {
		conn.Close()
		return nil, err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n < 12 {
		conn.Close()
		return nil, err
	}
	respStr := string(buf[:n])
	if len(respStr) < 12 || respStr[9:12] != "200" {
		conn.Close()
		return nil, fmt.Errorf("upstream proxy CONNECT failed: %s", respStr)
	}
	return conn, nil
}

func (h *ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	via := "direct"
	if h.usingProxy {
		via = "upstream proxy"
	}
	h.mu.RUnlock()
	log.Printf("%s %s via %s", r.Method, r.URL.String(), via)

	transport := h.transport.Load()
	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

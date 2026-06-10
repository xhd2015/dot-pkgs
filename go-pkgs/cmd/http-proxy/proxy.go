package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type ProxyHandler struct {
	transport atomic.Pointer[http.Transport]

	mu         sync.RWMutex
	proxyAddr  string // "host:port" of upstream proxy, empty = direct
	usingProxy bool
}

func NewProxyHandler() *ProxyHandler {
	h := &ProxyHandler{}
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
	h.mu.RLock()
	proxyAddr := h.proxyAddr
	h.mu.RUnlock()

	via := "direct"
	if proxyAddr != "" {
		via = "upstream proxy"
	}
	log.Printf("CONNECT %s via %s", r.Host, via)

	var destConn net.Conn
	var err error

	if proxyAddr != "" {
		destConn, err = net.DialTimeout("tcp", proxyAddr, 10*time.Second)
		if err != nil {
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			log.Printf("connect to upstream proxy %s: %v", proxyAddr, err)
			return
		}

		connectReq := "CONNECT " + r.Host + " HTTP/1.1\r\nHost: " + r.Host + "\r\n\r\n"
		if _, err := io.WriteString(destConn, connectReq); err != nil {
			destConn.Close()
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}

		buf := make([]byte, 4096)
		n, err := destConn.Read(buf)
		if err != nil || n < 12 {
			destConn.Close()
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
		respStr := string(buf[:n])
		if len(respStr) < 12 || respStr[9:12] != "200" {
			destConn.Close()
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
	} else {
		destConn, err = net.DialTimeout("tcp", r.Host, 10*time.Second)
		if err != nil {
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
	}

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

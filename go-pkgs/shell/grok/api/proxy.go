package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	// DefaultBaseURL is the Grok CLI chat proxy host (no path).
	DefaultBaseURL = "https://cli-chat-proxy.grok.com"
	// DefaultPort is the loopback listen port (Codex 8891, Command Code 8892).
	DefaultPort = 8893
)

// HandlerOpts configures NewHandler.
type HandlerOpts struct {
	// Home is the Grok config dir (auth.json, models_cache.json).
	Home string
	// BaseURL is the upstream host (default DefaultBaseURL).
	BaseURL string
	// Endpoint is the loopback catalog base_url, e.g. http://localhost:8893/v1.
	Endpoint string
	// HTTP is used as the reverse-proxy transport when set.
	HTTP *http.Client
	// Logf receives brief request lines. Never log tokens.
	Logf func(string, ...any)
	// LoadAuth injectable; nil → LoadAuth.
	LoadAuth func(path string) (Auth, error)
	// Ensure injectable; nil → EnsureAccessToken.
	Ensure func(ctx context.Context, opts EnsureOpts) (Auth, error)
}

type proxyHandler struct {
	opts     HandlerOpts
	authPath string
	target   *url.URL
	proxy    *httputil.ReverseProxy
}

func (h *proxyHandler) logf(format string, args ...any) {
	if h.opts.Logf != nil {
		h.opts.Logf(format, args...)
	}
}

func (h *proxyHandler) loadAuth() (Auth, error) {
	load := h.opts.LoadAuth
	if load == nil {
		load = LoadAuth
	}
	return load(h.authPath)
}

func (h *proxyHandler) ensure(ctx context.Context, auth Auth, force bool) (Auth, error) {
	ensure := h.opts.Ensure
	if ensure == nil {
		ensure = EnsureAccessToken
	}
	return ensure(ctx, EnsureOpts{
		Auth:         auth,
		AuthPath:     h.authPath,
		ForceRefresh: force,
		HTTPClient:   h.opts.HTTP,
	})
}

// NewHandler fail-fast loads credentials and returns a reverse proxy to
// cli-chat-proxy plus local GET /v1/models and /v1/models-v2.
func NewHandler(opts HandlerOpts) (http.Handler, error) {
	home := strings.TrimSpace(opts.Home)
	if home == "" {
		var err error
		home, err = DefaultHome()
		if err != nil {
			return nil, err
		}
		opts.Home = home
	}
	authPath, err := AuthPath(home)
	if err != nil {
		return nil, err
	}
	load := opts.LoadAuth
	if load == nil {
		load = LoadAuth
	}
	if _, err := load(authPath); err != nil {
		return nil, err
	}

	base := strings.TrimSpace(opts.BaseURL)
	if base == "" {
		base = DefaultBaseURL
	}
	target, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("grok api: invalid base url: %w", err)
	}

	h := &proxyHandler{opts: opts, authPath: authPath, target: target}
	transport := h.roundTripper()
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.Host = target.Host
		},
		Transport:     transport,
		FlushInterval: -1,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			h.logf("upstream error: %v", err)
			http.Error(w, "grok proxy: "+err.Error(), http.StatusBadGateway)
		},
	}
	h.proxy = proxy

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models-v2", h.serveModelsV2)
	mux.HandleFunc("/v1/models", h.serveModels)
	mux.Handle("/", proxy)
	return mux, nil
}

func (h *proxyHandler) roundTripper() http.RoundTripper {
	base := http.DefaultTransport
	if h.opts.HTTP != nil && h.opts.HTTP.Transport != nil {
		base = h.opts.HTTP.Transport
	}
	return &sessionTransport{handler: h, base: base}
}

type sessionTransport struct {
	handler *proxyHandler
	base    http.RoundTripper
}

func (t *sessionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := readAllBody(req)
	if err != nil {
		return nil, err
	}

	model := ""
	if req.Method == http.MethodPost && isJSONContentType(req.Header.Get("Content-Type")) {
		rewrittenModel, rewritten, rerr := RewriteRequestJSON(body)
		if rerr == nil {
			model = rewrittenModel
			if rewritten != nil {
				body = rewritten
			}
		}
	}
	setBody(req, body)

	auth, err := t.handler.loadAuth()
	if err != nil {
		return nil, err
	}
	auth, err = t.handler.ensure(req.Context(), auth, false)
	if err != nil {
		return nil, err
	}
	ApplySessionHeaders(req.Header, auth, model, t.handler.clientVersion())
	t.handler.logf("Request: %s %s (model=%s)", req.Method, req.URL.Path, model)

	rt := t.base
	if rt == nil {
		rt = http.DefaultTransport
	}
	resp, err := rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		return resp, nil
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	auth, err = t.handler.ensure(req.Context(), auth, true)
	if err != nil {
		return nil, err
	}
	retry, err := http.NewRequestWithContext(req.Context(), req.Method, req.URL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	retry.Header = req.Header.Clone()
	retry.Host = req.Host
	retry.ContentLength = int64(len(body))
	ApplySessionHeaders(retry.Header, auth, model, t.handler.clientVersion())
	t.handler.logf("Retry after %d with refreshed token", resp.StatusCode)
	return rt.RoundTrip(retry)
}

func (h *proxyHandler) serveModelsV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cache, err := h.readCache()
	if err != nil {
		http.Error(w, "Grok model catalog unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	if cache.ETag != "" {
		w.Header().Set("ETag", cache.ETag)
	}
	writeJSON(w, ModelsV2(cache, h.opts.Endpoint))
}

func (h *proxyHandler) serveModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cache, err := h.readCache()
	if err != nil {
		http.Error(w, "Grok model catalog unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, OpenAIModels(cache))
}

func (h *proxyHandler) clientVersion() string {
	cache, err := h.readCache()
	if err == nil && strings.TrimSpace(cache.GrokVersion) != "" {
		return strings.TrimSpace(cache.GrokVersion)
	}
	return DefaultClientVersion
}

func (h *proxyHandler) readCache() (ModelsCache, error) {
	path, err := ModelsCachePath(h.opts.Home)
	if err != nil {
		return ModelsCache{}, err
	}
	return ReadModelsCache(path)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "encode: "+err.Error(), http.StatusInternalServerError)
	}
}

func readAllBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func setBody(req *http.Request, body []byte) {
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
}

func isJSONContentType(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "application/json" || strings.HasPrefix(v, "application/json;")
}

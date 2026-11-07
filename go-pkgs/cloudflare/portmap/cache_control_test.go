package portmap

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func TestProxyResponseCacheControl(t *testing.T) {
	const immutable = "private, max-age=31536000, immutable"
	cases := []struct {
		name, method, path, contentType string
		status                          int
		cache                           []string
		want                            string
	}{
		{"private asset", "GET", "/assets/index-Abcd1234.js?v=stable", "application/javascript", 200, []string{immutable}, immutable},
		{"head", "HEAD", "/assets/index-Abcd1234.js", "application/javascript", 200, []string{immutable}, immutable},
		{"partial content", "GET", "/assets/index-Abcd1234.js", "application/javascript", 206, []string{immutable}, immutable},
		{"not modified", "GET", "/assets/index-Abcd1234.js", "application/javascript", 304, []string{immutable}, immutable},
		{"public stays browser only", "GET", "/assets/index-Abcd1234.js", "application/javascript", 200, []string{"public, max-age=600, immutable"}, "private, max-age=600, immutable"},
		{"shared ttl removed", "GET", "/assets/index-Abcd1234.js", "application/javascript", 200, []string{"public, max-age=60, s-maxage=3600"}, "private, max-age=60"},
		{"implicit shared caching prevented", "GET", "/assets/index-Abcd1234.js", "application/javascript", 200, []string{"max-age=600, immutable"}, "private, max-age=600, immutable"},
		{"revalidation retained", "GET", "/assets/index.js", "application/javascript", 200, []string{"private, no-cache"}, "private, no-cache"},
		{"no cache made private", "GET", "/assets/index.js", "application/javascript", 200, []string{"no-cache"}, "private, no-cache"},
		{"no store retained", "GET", "/api/private", "application/json", 200, []string{"no-store"}, "no-store"},
		{"private no store retained", "GET", "/api/private", "application/json", 200, []string{"private, no-store"}, "private, no-store"},
		{"origin explicitly revalidates html", "GET", "/app", "text/html", 200, []string{"private, no-cache"}, "private, no-cache"},
		{"origin explicitly caches api privately", "GET", "/api/image", "image/png", 200, []string{"private, max-age=60"}, "private, max-age=60"},
		{"unclassified html", "GET", "/app", "text/html", 200, nil, "no-store"},
		{"unclassified api", "GET", "/api/state", "application/json", 200, nil, "no-store"},
		{"hash alone does not grant ttl", "GET", "/assets/index-Abcd1234.js", "application/javascript", 200, nil, "no-store"},
		{"empty policy", "GET", "/assets/index.js", "application/javascript", 200, []string{" , "}, "no-store"},
		{"multiple lines", "GET", "/assets/index.js", "application/javascript", 200, []string{"public", "max-age=60, immutable", "s-maxage=300"}, "private, max-age=60, immutable"},
		{"case insensitive", "GET", "/assets/index.js", "application/javascript", 200, []string{"PuBliC, Max-Age=60, S-MaXaGe=300, IMMUTABLE"}, "private, Max-Age=60, IMMUTABLE"},
		{"qualified private strengthened", "GET", "/assets/index.js", "application/javascript", 200, []string{`private="Set-Cookie, Authorization", max-age=60`}, "private, max-age=60"},
		{"quoted extension", "GET", "/assets/index.js", "application/javascript", 200, []string{`public, max-age=60, example="a,b"`}, `private, max-age=60, example="a,b"`},
		{"escaped quoted extension", "GET", "/assets/index.js", "application/javascript", 200, []string{`public, example="a\"b,c", max-age=60`}, `private, example="a\"b,c", max-age=60`},
		{"malformed quote fails closed", "GET", "/assets/index.js", "application/javascript", 200, []string{`public, example="unterminated`}, "no-store"},
		{"malformed directive name", "GET", "/assets/index.js", "application/javascript", 200, []string{"private, bad name=60"}, "no-store"},
		{"malformed token value", "GET", "/assets/index.js", "application/javascript", 200, []string{"private, example=bad value"}, "no-store"},
		{"trailing quoted junk", "GET", "/assets/index.js", "application/javascript", 200, []string{`private, example="value"junk`}, "no-store"},
		{"embedded quotes", "GET", "/assets/index.js", "application/javascript", 200, []string{`private, example="a" "b"`}, "no-store"},
		{"invalid suffix after no store", "GET", "/assets/index.js", "application/javascript", 200, []string{"private, no-store, bad name=60"}, "no-store"},
		{"empty quoted extension", "GET", "/assets/index.js", "application/javascript", 200, []string{`private, example=""`}, `private, example=""`},
		{"missing chunk", "GET", "/assets/missing-Abcd1234.js", "application/javascript", 404, []string{immutable}, "no-store"},
		{"server error", "GET", "/assets/index.js", "application/javascript", 500, []string{immutable}, "no-store"},
		{"auth error", "GET", "/api/state", "application/json", 401, []string{immutable}, "no-store"},
		{"redirect", "GET", "/app", "text/html", 302, []string{immutable}, "no-store"},
		{"post", "POST", "/api/state", "application/json", 201, []string{immutable}, "no-store"},
		{"put", "PUT", "/api/state", "application/json", 200, []string{immutable}, "no-store"},
		{"delete", "DELETE", "/api/state", "application/json", 204, []string{immutable}, "no-store"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RequestURI() != tc.path || r.Method != tc.method {
					t.Errorf("request rewritten: %s %s", r.Method, r.URL.RequestURI())
				}
				if r.Header.Get("If-None-Match") != `"asset-version"` || r.Header.Get("If-Modified-Since") != "Thu, 02 Jan 2025 03:04:05 GMT" || r.Header.Get("Cache-Control") != "max-age=0" {
					t.Error("request cache validators or revalidation policy were changed")
				}
				for _, value := range tc.cache {
					w.Header().Add("Cache-Control", value)
				}
				w.Header().Set("Content-Type", tc.contentType)
				w.Header().Set("ETag", `"asset-version"`)
				w.Header().Set("Last-Modified", "Thu, 02 Jan 2025 03:04:05 GMT")
				w.Header().Set("Vary", "Accept-Encoding")
				w.Header().Set("CDN-Cache-Control", "public, s-maxage=3600")
				w.Header().Set("Cloudflare-CDN-Cache-Control", "public, max-age=3600")
				w.Header().Set("Surrogate-Control", "max-age=3600")
				if tc.status == 206 {
					w.Header().Set("Content-Range", "bytes 0-2/10")
				}
				w.WriteHeader(tc.status)
				if tc.method != "HEAD" && tc.status != 304 && tc.status != 204 {
					_, _ = fmt.Fprint(w, "abc")
				}
			}))
			defer upstream.Close()
			target, _ := url.Parse(upstream.URL)
			port, _ := strconv.Atoi(target.Port())
			manager, _, _ := testManager(t)
			if _, err := manager.Ensure(context.Background(), "cache.example.com", port); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(tc.method, "https://cache.example.com"+tc.path, nil)
			req.Header.Set("If-None-Match", `"asset-version"`)
			req.Header.Set("If-Modified-Since", "Thu, 02 Jan 2025 03:04:05 GMT")
			req.Header.Set("Cache-Control", "max-age=0")
			rec := httptest.NewRecorder()
			manager.ServeHTTP(rec, req)
			if rec.Code != tc.status || strings.Join(rec.Header().Values("Cache-Control"), ", ") != tc.want {
				t.Fatalf("status=%d cache=%q, want status=%d cache=%q", rec.Code, rec.Header().Values("Cache-Control"), tc.status, tc.want)
			}
			for key, want := range map[string]string{"ETag": `"asset-version"`, "Last-Modified": "Thu, 02 Jan 2025 03:04:05 GMT", "Vary": "Accept-Encoding"} {
				if got := rec.Header().Get(key); got != want {
					t.Errorf("%s=%q, want %q", key, got, want)
				}
			}
			for _, key := range []string{"CDN-Cache-Control", "Cloudflare-CDN-Cache-Control", "Surrogate-Control"} {
				if rec.Header().Get(key) != "" {
					t.Errorf("shared-cache override escaped relay: %s=%q", key, rec.Header().Get(key))
				}
			}
			if tc.status == 206 && rec.Header().Get("Content-Range") != "bytes 0-2/10" {
				t.Error("partial content metadata was changed")
			}
			if tc.method == "HEAD" || tc.status == 304 || tc.status == 204 {
				if rec.Body.Len() != 0 {
					t.Errorf("unexpected response body: %q", rec.Body.String())
				}
			} else if rec.Body.String() != "abc" {
				t.Errorf("response body changed: %q", rec.Body.String())
			}
		})
	}
}

func TestProxyPrivateCachingDoesNotBypassWrapper(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Cache-Control", "public, max-age=600, s-maxage=3600, immutable")
		_, _ = fmt.Fprint(w, "protected asset")
	}))
	defer upstream.Close()
	target, _ := url.Parse(upstream.URL)
	port, _ := strconv.Atoi(target.Port())
	manager, err := New(Options{Publisher: &fakePublisher{}, Wrap: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Test-Auth") != "allowed" {
				w.Header().Set("Cache-Control", "no-store")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	if _, err := manager.Ensure(context.Background(), "cache.example.com", port); err != nil {
		t.Fatal(err)
	}
	for _, authed := range []bool{false, true, false} {
		req := httptest.NewRequest("GET", "https://cache.example.com/assets/app-Abcd1234.js", nil)
		if authed {
			req.Header.Set("X-Test-Auth", "allowed")
		}
		rec := httptest.NewRecorder()
		manager.server.Handler.ServeHTTP(rec, req)
		if authed {
			if rec.Code != 200 || rec.Header().Get("Cache-Control") != "private, max-age=600, immutable" || rec.Body.String() != "protected asset" {
				t.Fatalf("authenticated response: status=%d cache=%q body=%q", rec.Code, rec.Header().Get("Cache-Control"), rec.Body.String())
			}
		} else if rec.Code != 401 || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("unauthenticated response: status=%d cache=%q", rec.Code, rec.Header().Get("Cache-Control"))
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("unauthenticated request reached upstream: calls=%d", calls.Load())
	}
}

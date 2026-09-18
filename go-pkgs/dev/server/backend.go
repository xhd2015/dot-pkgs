package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

const readyPath = "/__dev/ready"
const runEnv = "DOT_PKGS_DEV_RUN_ID"

func identityHandler(id string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == readyPath {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, id)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func serve(ctx context.Context, cfg Config, opts options, id string) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.port))
	if err != nil {
		return fmt.Errorf("backend bind: %w", err)
	}
	defer ln.Close()
	b := Backend{Root: cfg.Root, Port: opts.port, URL: fmt.Sprintf("http://127.0.0.1:%d", opts.port)}
	if opts.vitePort != 0 {
		b.FrontendURL = fmt.Sprintf("http://127.0.0.1:%d", opts.vitePort)
	}
	handler, err := cfg.BackendHandler(ctx, b)
	if err != nil {
		return err
	}
	cleanup := func() {}
	if cfg.OnListen != nil {
		fn, err := cfg.OnListen(b)
		if err != nil {
			return err
		}
		if fn != nil {
			var once sync.Once
			cleanup = func() { once.Do(fn) }
			defer cleanup()
		}
	}
	srv := &http.Server{ReadHeaderTimeout: 10 * time.Second}
	srv.Handler = identityHandler(id, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__dev/stop" {
			if r.Method != "POST" || r.Header.Get("X-Dev-Run-ID") != id {
				http.Error(w, "invalid dev invocation", http.StatusForbidden)
				return
			}
			// Acknowledge only after application cleanup. Air may kill a backend
			// immediately on supervisor exit even with send_interrupt enabled.
			cleanup()
			w.WriteHeader(http.StatusNoContent)
			go func() {
				stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_ = srv.Shutdown(stopCtx)
			}()
			return
		}
		handler.ServeHTTP(w, r)
	}))
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = srv.Close()
		case <-done:
		}
	}()
	err = srv.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

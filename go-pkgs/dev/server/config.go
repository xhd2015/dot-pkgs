// Package server provides a shared Go/Air and Vite development supervisor.
// It currently targets macOS and Linux (Unix process groups).
package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Name                      string
	Root                      string
	BuildPackage              string
	BackendPort, FrontendPort int
	BrowserPath               string
	WatchDirs                 []string
	Frontend                  *Vite
	BackendHandler            func(context.Context, Backend) (http.Handler, error)
	// OnListen registers application discovery after binding and initialization.
	// Its cleanup runs before the child exits, including on Air rebuilds.
	OnListen       func(Backend) (func(), error)
	StartupTimeout time.Duration
	Stdout, Stderr io.Writer
}

type Backend struct {
	Root, URL, FrontendURL string
	Port                   int
}

type Vite struct {
	Dir        string
	ConfigFile string   // relative to Dir; e.g. vite.config.ts
	Install    []string // e.g. pnpm install --frozen-lockfile
	Command    []string // e.g. pnpm exec vite (flags are appended)
}

// FindRoot walks upwards looking for go.mod plus a project-specific marker.
func FindRoot(marker string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir, nil
			}
		}
		next := filepath.Dir(dir)
		if next == dir {
			return "", fmt.Errorf("run from a repository containing go.mod and %s", marker)
		}
		dir = next
	}
}

package rg

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/github/release"
	"github.com/xhd2015/dot-pkgs/go-pkgs/os/exectry"
)

// InstallOpts configures InstallLatest.
type InstallOpts struct {
	Home       string
	DestDir    string // empty → {Home}/.local/bin
	CacheDir   string // empty → MkdirTemp
	GOOS       string // empty → runtime.GOOS
	GOARCH     string // empty → runtime.GOARCH
	HTTPClient *http.Client
	// FetchLatestTag overrides release.FetchLatestReleaseTag when non-nil.
	FetchLatestTag func(ctx context.Context) (string, error)
	// DownloadURL overrides constructed GitHub asset URL when non-empty (tests).
	DownloadURL string
}

// InstallResult is the outcome of InstallLatest.
type InstallResult struct {
	Tag     string
	Version string
	URL     string
	BinPath string
}

// InstallLatest downloads the official precompiled rg for this platform into DestDir.
// Unsupported GOOS/GOARCH returns a hard error (no brew/cargo fallback).
func InstallLatest(ctx context.Context, opts InstallOpts) (InstallResult, error) {
	var result InstallResult
	if ctx == nil {
		ctx = context.Background()
	}

	goos := strings.TrimSpace(opts.GOOS)
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := strings.TrimSpace(opts.GOARCH)
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	triple, ok := targetTriple(goos, goarch)
	if !ok {
		return result, fmt.Errorf("rg: no precompiled binary for %s/%s", goos, goarch)
	}

	fetch := opts.FetchLatestTag
	if fetch == nil {
		fetch = func(ctx context.Context) (string, error) {
			return release.FetchLatestReleaseTag(ctx, githubOwner, githubRepo)
		}
	}
	tag, err := fetch(ctx)
	if err != nil {
		return result, fmt.Errorf("rg: fetch latest release: %w", err)
	}
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	if tag == "" {
		return result, fmt.Errorf("rg: empty latest release tag")
	}
	result.Tag = tag
	result.Version = tag

	url := strings.TrimSpace(opts.DownloadURL)
	if url == "" {
		url = DownloadURL(tag, triple, goos)
	}
	result.URL = url

	home := strings.TrimSpace(opts.Home)
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return result, fmt.Errorf("rg: home dir: %w", err)
		}
		home = h
	}
	destDir := strings.TrimSpace(opts.DestDir)
	if destDir == "" {
		destDir = filepath.Join(home, ".local", "bin")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return result, fmt.Errorf("rg: mkdir dest: %w", err)
	}

	cacheDir := strings.TrimSpace(opts.CacheDir)
	cleanupCache := false
	if cacheDir == "" {
		tmp, err := os.MkdirTemp("", "rg-install-*")
		if err != nil {
			return result, fmt.Errorf("rg: cache dir: %w", err)
		}
		cacheDir = tmp
		cleanupCache = true
	}
	if cleanupCache {
		defer os.RemoveAll(cacheDir)
	}

	archivePath := filepath.Join(cacheDir, AssetName(tag, triple, goos))
	if err := downloadFile(ctx, url, archivePath, opts.HTTPClient); err != nil {
		return result, err
	}

	extractDir := filepath.Join(cacheDir, "extract")
	binSrc, err := extractRgBinary(archivePath, extractDir, goos)
	if err != nil {
		return result, err
	}

	destName := binaryName(goos)
	destPath := filepath.Join(destDir, destName)
	body, err := os.ReadFile(binSrc)
	if err != nil {
		return result, fmt.Errorf("rg: read extracted binary: %w", err)
	}
	if err := exectry.WriteExecutable(destPath, body); err != nil {
		return result, fmt.Errorf("rg: install binary: %w", err)
	}
	result.BinPath = destPath
	return result, nil
}

func downloadFile(ctx context.Context, url, destPath string, client *http.Client) error {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("rg: download: build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("rg: download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rg: download: HTTP %d from %s", resp.StatusCode, url)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("rg: download mkdir: %w", err)
	}
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("rg: download create: %w", err)
	}
	defer f.Close()
	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("rg: download write: %w", err)
	}
	if n == 0 {
		_ = os.Remove(destPath)
		return fmt.Errorf("rg: download: empty body from %s", url)
	}
	return nil
}

func extractRgBinary(archivePath, workDir, goos string) (binPath string, err error) {
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", fmt.Errorf("rg: extract mkdir: %w", err)
	}
	want := binaryName(goos)
	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		return extractNamedFromZip(archivePath, workDir, want)
	}
	return extractNamedFromTarGz(archivePath, workDir, want)
}

func extractNamedFromTarGz(archivePath, workDir, wantBase string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("rg: open archive: %w", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("rg: gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("rg: tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base != wantBase {
			continue
		}
		dest := filepath.Join(workDir, base)
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return "", fmt.Errorf("rg: write extracted: %w", err)
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return "", fmt.Errorf("rg: copy extracted: %w", err)
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return dest, nil
	}
	return "", fmt.Errorf("rg: archive missing %s", wantBase)
}

func extractNamedFromZip(archivePath, workDir, wantBase string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("rg: open zip: %w", err)
	}
	defer r.Close()
	for _, zf := range r.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(zf.Name) != wantBase {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return "", fmt.Errorf("rg: open zip entry: %w", err)
		}
		dest := filepath.Join(workDir, wantBase)
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			rc.Close()
			return "", fmt.Errorf("rg: write zip entry: %w", err)
		}
		_, copyErr := io.Copy(out, rc)
		rc.Close()
		closeErr := out.Close()
		if copyErr != nil {
			return "", fmt.Errorf("rg: copy zip entry: %w", copyErr)
		}
		if closeErr != nil {
			return "", closeErr
		}
		return dest, nil
	}
	return "", fmt.Errorf("rg: zip missing %s", wantBase)
}

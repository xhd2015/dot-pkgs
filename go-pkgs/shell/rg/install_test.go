package rg

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallLatestUnsupportedPlatform(t *testing.T) {
	t.Parallel()
	_, err := InstallLatest(context.Background(), InstallOpts{
		GOOS:   "plan9",
		GOARCH: "amd64",
		Home:   t.TempDir(),
	})
	if err == nil || err.Error() == "" {
		t.Fatal("expected error")
	}
	if want := "no precompiled binary for plan9/amd64"; !strings.Contains(err.Error(), want) {
		t.Fatalf("err=%v", err)
	}
}

func TestInstallLatestFromTarGz(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte("#!/bin/sh\necho ripgrep 15.2.0\n")
	hdr := &tar.Header{
		Name: "ripgrep-15.2.0-aarch64-apple-darwin/rg",
		Mode: 0o755,
		Size: int64(len(body)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	archive := buf.Bytes()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(archive)
	}))
	defer srv.Close()

	home := t.TempDir()
	dest := filepath.Join(home, ".local", "bin")
	res, err := InstallLatest(context.Background(), InstallOpts{
		Home:        home,
		DestDir:     dest,
		GOOS:        "darwin",
		GOARCH:      "arm64",
		HTTPClient:  srv.Client(),
		DownloadURL: srv.URL + "/ripgrep.tgz",
		FetchLatestTag: func(ctx context.Context) (string, error) {
			return "15.2.0", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.BinPath != filepath.Join(dest, "rg") {
		t.Fatalf("bin=%q", res.BinPath)
	}
	got, err := os.ReadFile(res.BinPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("installed body mismatch")
	}
}


package rg

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	githubOwner = "BurntSushi"
	githubRepo  = "ripgrep"
)

// targetTriple maps GOOS/GOARCH to a ripgrep release target triple.
// Unsupported pairs return ("", false).
func targetTriple(goos, goarch string) (triple string, ok bool) {
	goos = strings.TrimSpace(goos)
	goarch = strings.TrimSpace(goarch)
	switch goos + "/" + goarch {
	case "darwin/arm64":
		return "aarch64-apple-darwin", true
	case "darwin/amd64":
		return "x86_64-apple-darwin", true
	case "linux/amd64":
		return "x86_64-unknown-linux-musl", true
	case "linux/arm64":
		return "aarch64-unknown-linux-musl", true
	case "windows/amd64":
		return "x86_64-pc-windows-msvc", true
	case "windows/arm64":
		return "aarch64-pc-windows-msvc", true
	default:
		return "", false
	}
}

// HostTargetTriple returns the release triple for the running platform.
func HostTargetTriple() (string, error) {
	t, ok := targetTriple(runtime.GOOS, runtime.GOARCH)
	if !ok {
		return "", fmt.Errorf("rg: no precompiled binary for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return t, nil
}

// archiveExt is .zip on Windows, .tar.gz elsewhere.
func archiveExt(goos string) string {
	if goos == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

// binaryName inside the archive.
func binaryName(goos string) string {
	if goos == "windows" {
		return "rg.exe"
	}
	return "rg"
}

// AssetName builds the GitHub release asset basename for tag+triple.
// tag is the release tag without a leading "v" (ripgrep uses bare semver tags).
func AssetName(tag, triple, goos string) string {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	return fmt.Sprintf("ripgrep-%s-%s%s", tag, triple, archiveExt(goos))
}

// DownloadURL builds the browser download URL for a release asset.
func DownloadURL(tag, triple, goos string) string {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	return fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/%s/%s",
		githubOwner, githubRepo, tag, AssetName(tag, triple, goos),
	)
}

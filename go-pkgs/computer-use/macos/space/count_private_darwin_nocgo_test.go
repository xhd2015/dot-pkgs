//go:build darwin && !cgo

package space

import (
	"strings"
	"testing"
)

func TestCountDesktopsPrivateAPIRequiresCgo(t *testing.T) {
	_, err := countDesktopsPrivateAPI()
	if err == nil {
		t.Fatal("expected error without cgo")
	}
	if !strings.Contains(err.Error(), "cgo") {
		t.Fatalf("error %q: want substring cgo", err)
	}
}

func TestLiveManagedDisplaySpacesRequiresCgo(t *testing.T) {
	_, err := liveManagedDisplaySpacesImpl()
	if err == nil {
		t.Fatal("expected error without cgo")
	}
	if !strings.Contains(err.Error(), "cgo") {
		t.Fatalf("error %q: want substring cgo", err)
	}
}

func TestLiveCopySpacesForWindowsRequiresCgo(t *testing.T) {
	_, err := liveCopySpacesForWindowsImpl(0, nil)
	if err == nil {
		t.Fatal("expected error without cgo")
	}
	if !strings.Contains(err.Error(), "cgo") {
		t.Fatalf("error %q: want substring cgo", err)
	}
}

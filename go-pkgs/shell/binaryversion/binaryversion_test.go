package binaryversion

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/os/exectry"
)

func TestNewestUsesHighestVersionAndStableTieBreak(t *testing.T) {
	probe := func(_ context.Context, binary string) (string, error) {
		switch binary {
		case "/first":
			return "tool 1.2.0", nil
		case "/newest":
			return "v1.10.0", nil
		case "/same":
			return "1.10.0", nil
		default:
			return "", errors.New("unusable")
		}
	}
	got, err := Newest(context.Background(), []Candidate{
		{Path: "/first", Via: "preferred"},
		{Path: "/bad"},
		{Path: "/newest", Via: "path"},
		{Path: "/same", Via: "fallback"},
	}, probe)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != "/newest" || got.Version != "1.10.0" || got.Via != "path" {
		t.Fatalf("Newest = %+v", got)
	}
}

func TestFindSkipsDuplicatesAndUnparseableVersions(t *testing.T) {
	calls := 0
	probe := func(_ context.Context, binary string) (string, error) {
		calls++
		if binary == "/bad" {
			return "development", nil
		}
		return "1.2.3", nil
	}
	found := Find(context.Background(), []Candidate{{Path: "/good"}, {Path: "/good"}, {Path: "/bad"}}, probe)
	if calls != 2 || len(found) != 1 || found[0].Path != "/good" {
		t.Fatalf("calls=%d found=%+v", calls, found)
	}
}

func TestCompareSemver(t *testing.T) {
	got, err := CompareSemver("v0.0.9", "0.0.10")
	if err != nil || got >= 0 {
		t.Fatalf("CompareSemver = %d, %v", got, err)
	}
}

func TestCommandSupportsCLISpecificVersionArguments(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "tool")
	if err := exectry.WriteExecutable(bin, []byte("#!/bin/sh\n[ \"$1\" = version ] || exit 2\nprintf 'tool v2.3.4\\n'\n")); err != nil {
		t.Fatal(err)
	}
	got, err := Command("version")(context.Background(), bin)
	if err != nil || got != "2.3.4" {
		t.Fatalf("Command(version) = %q, %v", got, err)
	}
}

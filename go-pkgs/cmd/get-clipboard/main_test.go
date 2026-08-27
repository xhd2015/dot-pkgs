package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestPrintAndMaybeOpen(t *testing.T) {
	orig := openCmd
	t.Cleanup(func() { openCmd = orig })

	t.Run("no_open", func(t *testing.T) {
		var called []string
		openCmd = func(path string) error {
			called = append(called, path)
			return nil
		}
		var buf bytes.Buffer
		if err := printAndMaybeOpen(&buf, "/tmp/demo.png", false); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "/tmp/demo.png\n" {
			t.Errorf("stdout = %q, want path line", got)
		}
		if len(called) != 0 {
			t.Errorf("openCmd called with %v, want none", called)
		}
	})

	t.Run("with_open", func(t *testing.T) {
		var called []string
		openCmd = func(path string) error {
			called = append(called, path)
			return nil
		}
		var buf bytes.Buffer
		if err := printAndMaybeOpen(&buf, "/tmp/shot.png", true); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "/tmp/shot.png\n" {
			t.Errorf("stdout = %q, want path line", got)
		}
		if len(called) != 1 || called[0] != "/tmp/shot.png" {
			t.Errorf("openCmd called with %v, want [/tmp/shot.png]", called)
		}
	})

	t.Run("open_fails", func(t *testing.T) {
		openCmd = func(path string) error {
			return errors.New("boom")
		}
		var buf bytes.Buffer
		err := printAndMaybeOpen(&buf, "/tmp/fail.png", true)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "open /tmp/fail.png") {
			t.Errorf("error = %v, want open path prefix", err)
		}
		if got := buf.String(); got != "/tmp/fail.png\n" {
			t.Errorf("stdout = %q, want path printed before open error", got)
		}
	})
}

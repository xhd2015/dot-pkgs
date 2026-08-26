package lineprefix

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
)

func TestNew_singleFullLine(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "[x] ")
	if _, err := w.Write([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "[x] hello\n" {
		t.Fatalf("got %q", got)
	}
}

func TestNew_splitWrites(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "P:")
	if _, err := w.Write([]byte("hel")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("lo\n")); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "P:hello\n" {
		t.Fatalf("got %q", got)
	}
}

func TestNew_twoLinesOneWrite(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, ">")
	if _, err := w.Write([]byte("a\nb\n")); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != ">a\n>b\n" {
		t.Fatalf("got %q", got)
	}
}

func TestNew_noTrailingNewline(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "[t] ")
	if _, err := w.Write([]byte("partial")); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "[t] partial" {
		t.Fatalf("got %q", got)
	}
}

func TestNew_emptyPrefix(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "")
	if _, err := w.Write([]byte("a\nb\n")); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "a\nb\n" {
		t.Fatalf("got %q", got)
	}
}

func TestNew_nilUnderlying(t *testing.T) {
	w := New(nil, "[x] ")
	n, err := w.Write([]byte("hi\n"))
	if err != nil || n != 3 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if w.Unwrap() != io.Discard {
		t.Fatalf("unwrap=%v", w.Unwrap())
	}
}

func TestBracket(t *testing.T) {
	var buf bytes.Buffer
	w := Bracket(&buf, "codex")
	if _, err := w.Write([]byte("err\n")); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "[codex] err\n" {
		t.Fatalf("got %q", got)
	}
}

func TestBracket_emptyTag(t *testing.T) {
	var buf bytes.Buffer
	w := Bracket(&buf, "")
	if _, err := w.Write([]byte("x\n")); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "x\n" {
		t.Fatalf("got %q", got)
	}
}

func TestNested(t *testing.T) {
	var buf bytes.Buffer
	// Closer-to-source wraps the outer job decorator (Marcus tee), matching
	// Codex → [codex] → process stderr → [knowledge-sink] → hub.
	hub := Bracket(&buf, "knowledge-sink")
	agent := Bracket(hub, "codex")
	if _, err := agent.Write([]byte("boom\n")); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if got != "[knowledge-sink] [codex] boom\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWriteReturnLen(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "PREFIX:")
	in := []byte("one\ntwo\n")
	n, err := w.Write(in)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(in) {
		t.Fatalf("n=%d want %d", n, len(in))
	}
}

func TestConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "P:")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = w.Write([]byte("line\n"))
		}()
	}
	wg.Wait()
	got := buf.String()
	if strings.Count(got, "P:line\n") != 50 {
		t.Fatalf("got count=%d body=%q", strings.Count(got, "P:line\n"), got)
	}
	if strings.Contains(got, "PP:") || strings.Contains(got, "lineP:") {
		t.Fatalf("torn prefix: %q", got)
	}
}

func TestCRLF(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "X:")
	if _, err := w.Write([]byte("a\r\nb\r\n")); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "X:a\r\nX:b\r\n" {
		t.Fatalf("got %q", got)
	}
}

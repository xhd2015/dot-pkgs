package qemu

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOutputLinesEmitsBeforeProcessExits(t *testing.T) {
	got := make(chan string, 1)
	done := make(chan struct{})
	go func() {
		_, _ = (LocalHost{}).OutputLines(context.Background(), "bash", []string{"-c", "echo fast; sleep 2; echo slow"}, func(line string) {
			if line == "fast" {
				select {
				case got <- line:
				default:
				}
			}
		})
		close(done)
	}()
	select {
	case line := <-got:
		if line != "fast" {
			t.Fatalf("line = %q", line)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("first line was not delivered before the process finished")
	}
	select {
	case <-done:
		t.Fatal("process exited before the sleep, so the line was not streamed")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestOutputLinesTimeoutKeepsEarlierLines(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	var lines []string
	_, err := (LocalHost{}).OutputLines(ctx, "bash", []string{"-c", "echo fast; sleep 30; echo slow"}, func(line string) {
		lines = append(lines, line)
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}
	if len(lines) != 1 || lines[0] != "fast" {
		t.Fatalf("lines = %v", lines)
	}
}

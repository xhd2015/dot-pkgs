package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

// OpenLog opens path for appending, creating the directory and the file when
// needed. Managed output is append-only: a restart must not erase the crash
// that caused it.
func OpenLog(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return f, nil
}

// AppendLog writes one formatted line to path, creating it when needed. A
// trailing newline is added unless format already ends with one, so callers and
// child-process output never merge into a single line.
func AppendLog(path, format string, args ...any) error {
	f, err := OpenLog(path)
	if err != nil {
		return err
	}
	defer f.Close()
	message := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	_, err = f.WriteString(message)
	return err
}

// TailLog returns up to lines of path, oldest first.
//
// A missing file is not an error: a service that never started has no log yet,
// and the GUI shows an empty pane rather than a failure.
func TailLog(path string, lines int) ([]string, error) {
	if lines <= 0 {
		return nil, nil
	}
	reversed := make([]string, 0, lines)
	err := logs.ScanLinesFromTail(path, logs.ScanLinesFromTailOptions{}, func(line string) (bool, error) {
		reversed = append(reversed, line)
		return len(reversed) >= lines, nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]string, len(reversed))
	for i, line := range reversed {
		out[len(reversed)-1-i] = line
	}
	return out, nil
}

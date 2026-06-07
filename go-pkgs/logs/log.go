package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	go_filepath "path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileOp describes the kind of file system operation that triggered an event.
type FileOp int

const (
	FileCreated FileOp = iota
	FileModified
)

// FileEvent describes a file system event delivered to WatchFileEvents callbacks.
type FileEvent struct {
	Path string
	Op   FileOp
}

// WatchFileEventsOptions controls debounce behaviour for WatchFileEvents.
type WatchFileEventsOptions struct {
	// DisableDebounce disables event coalescing.
	// When true every fsnotify event fires the callback immediately.
	DisableDebounce bool

	// Debounce is the quiet period to wait before firing the callback.
	// Defaults to 20 ms. Ignored when DisableDebounce is true.
	Debounce time.Duration
}

// WatchContentOptions controls behaviour for Watch.
type WatchContentOptions struct {
	// DisableDebounce disables event coalescing.
	// When true every fsnotify event fires the callback immediately.
	DisableDebounce bool

	// Debounce is the quiet period to wait before firing the callback.
	// Defaults to 20 ms. Ignored when DisableDebounce is true.
	Debounce time.Duration
}

// WatchLineOptions controls behaviour for WatchLine.
type WatchLineOptions struct {
	// DisableDebounce disables event coalescing.
	// When true every fsnotify event fires the callback immediately.
	DisableDebounce bool

	// Debounce is the quiet period to wait before firing the callback.
	// Defaults to 20 ms. Ignored when DisableDebounce is true.
	Debounce time.Duration
}

func contentDebounce(opts WatchContentOptions) time.Duration {
	if opts.DisableDebounce {
		return 0
	}
	if opts.Debounce > 0 {
		return opts.Debounce
	}
	return 20 * time.Millisecond
}

func lineDebounce(opts WatchLineOptions) time.Duration {
	if opts.DisableDebounce {
		return 0
	}
	if opts.Debounce > 0 {
		return opts.Debounce
	}
	return 20 * time.Millisecond
}

type fileEventKind int

const (
	fileKindCreated fileEventKind = iota
	fileKindModified
)

// watchFileEvents is the lowest-level core. It sets up the fsnotify watcher,
// registers the directory and file, and dispatches events through onEvent.
// onEvent receives fileKindCreated for creation events and fileKindModified
// for writes. A single os.WriteFile on a new file produces one Create event
// (even when the OS combines Create|Write in the raw mask).
func watchFileEvents(ctx context.Context, filepath string, onEvent func(kind fileEventKind) error) error {
	if realPath, err := go_filepath.EvalSymlinks(filepath); err == nil {
		filepath = realPath
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Add the file's directory to watch for file creation
	dir := path.Dir(filepath)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", dir, err)
	}

	// If file exists, add it to the watcher and initialize lastPos to current size
	if fileExists(filepath) {
		if err := watcher.Add(filepath); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", filepath, err)
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// If the file was created, start watching it
			if event.Op&fsnotify.Create != 0 && event.Name == filepath {
				if err := watcher.Add(filepath); err != nil {
					return fmt.Errorf("failed to watch newly created file %s: %w", filepath, err)
				}
			}

			// If the file was modified or created, notify

			if event.Name == filepath && (event.Op&fsnotify.Create != 0 || event.Op&fsnotify.Write != 0) {
				kind := fileKindModified
				if event.Op&fsnotify.Create != 0 {
					kind = fileKindCreated
				}
				if err := onEvent(kind); err != nil {
					return err
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("watcher error: %w", err)

		case <-ctx.Done():
			return nil
		}
	}
}

// watchFile wraps watchFileEvents with lastPos tracking and debounce so
// callers can do incremental reads (only new bytes since the last event).
// When debounce > 0, rapid writes are coalesced; the handler fires after
// the quiet period.
func watchFile(ctx context.Context, filepath string, debounce time.Duration, onWrite func(lastPos *int64) error) error {
	// Keep track of last read position
	lastPos := int64(0)
	if fi, err := os.Stat(filepath); err == nil {
		lastPos = fi.Size()
	}

	var mu sync.Mutex
	var timer *time.Timer
	var hasPending bool

	return watchFileEvents(ctx, filepath, func(kind fileEventKind) error {
		_ = kind

		if debounce <= 0 {
			return onWrite(&lastPos)
		}

		mu.Lock()
		hasPending = true
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounce, func() {
			mu.Lock()
			if !hasPending {
				mu.Unlock()
				return
			}
			hasPending = false
			mu.Unlock()
			if err := onWrite(&lastPos); err != nil {
				fmt.Fprintf(os.Stderr, "watchFile onWrite error: %v\n", err)
			}
		})
		mu.Unlock()
		return nil
	})
}

// Watch monitors a file for changes and calls the provided callback function
// for each new content chunk detected.
//
// Parameters:
//   - ctx: Context for cancellation
//   - filepath: Path to the file to monitor
//   - callback: Function called for each new content chunk
//
// The callback function should return an error if processing should stop.
// The function continues until the context is cancelled, an error occurs,
// or the callback returns an error.
//
// If the file doesn't exist initially, it will wait for it to be created.
func Watch(ctx context.Context, filepath string, opts WatchContentOptions, callback func(content []byte) error) error {
	return watchFile(ctx, filepath, contentDebounce(opts), func(lastPos *int64) error {
		newPos, err := readNewContent(filepath, *lastPos, callback)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filepath, err)
			return nil
		}
		*lastPos = newPos
		return nil
	})
}

// WatchLine monitors a file for changes and calls the provided callback
// for each complete line detected. Unlike Watch, which delivers raw byte
// chunks, WatchLine buffers content and splits on newlines so the callback
// always receives one full line at a time.
//
// The callback receives the line content without the trailing newline
// character. If the line ends with \r\n, the \r is also stripped.
func WatchLine(ctx context.Context, filepath string, opts WatchLineOptions, callback func(line string) error) error {
	var leftover []byte
	return watchFile(ctx, filepath, lineDebounce(opts), func(lastPos *int64) error {
		newPos, err := readNewContent(filepath, *lastPos, func(content []byte) error {
			data := append(leftover, content...)
			leftover = nil
			for {
				idx := bytes.IndexByte(data, '\n')
				if idx < 0 {
					leftover = data
					break
				}
				line := string(data[:idx])
				line = strings.TrimSuffix(line, "\r")
				if err := callback(line); err != nil {
					return err
				}
				data = data[idx+1:]
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filepath, err)
			return nil
		}
		*lastPos = newPos
		return nil
	})
}

// WatchFileEvents monitors a file for creation and modification events.
// Unlike Watch / WatchLine it does not read file content; it delivers
// FileEvent notifications whenever the file is created or written to.
//
// On platforms that emit multiple events for a single write (macOS),
// options.Debounce coalesces them into one callback after the quiet
// period. The default debounce is 20 ms. Set DisableDebounce to true
// for immediate delivery.
func WatchFileEvents(ctx context.Context, filepath string, opts WatchFileEventsOptions, callback func(event FileEvent) error) error {
	debounce := opts.Debounce
	if !opts.DisableDebounce && debounce <= 0 {
		debounce = 20 * time.Millisecond
	}

	var mu sync.Mutex
	var timer *time.Timer
	var pendingOp FileOp
	var hasPending bool

	return watchFileEvents(ctx, filepath, func(kind fileEventKind) error {
		op := FileModified
		if kind == fileKindCreated {
			op = FileCreated
		}

		if opts.DisableDebounce || debounce <= 0 {
			return callback(FileEvent{Path: filepath, Op: op})
		}

		mu.Lock()
		if !hasPending || op == FileCreated {
			pendingOp = op
		}
		hasPending = true
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounce, func() {
			mu.Lock()
			if !hasPending {
				mu.Unlock()
				return
			}
			op := pendingOp
			hasPending = false
			mu.Unlock()
			if err := callback(FileEvent{Path: filepath, Op: op}); err != nil {
				fmt.Fprintf(os.Stderr, "WatchFileEvents callback error: %v\n", err)
			}
		})
		mu.Unlock()
		return nil
	})
}

// Pipe monitors a file for changes and outputs new content to the provided
// writer, prefixing each chunk with the specified prefix.
//
// Parameters:
//   - ctx: Context for cancellation
//   - filepath: Path to the file to monitor
//   - prefix: Prefix to prepend to each output chunk
//   - output: Writer where output will be written (typically os.Stdout)
//
// The function continues until the context is cancelled or an error occurs.
// If the file doesn't exist initially, it will wait for it to be created.
func Pipe(ctx context.Context, filepath string, prefix string, output io.Writer) error {
	return Watch(ctx, filepath, WatchContentOptions{}, func(line []byte) error {
		if prefix != "" {
			_, err := fmt.Fprint(output, prefix)
			if err != nil {
				return err
			}
		}
		_, err := output.Write(line)
		if err != nil {
			return err
		}
		if prefix != "" && !bytes.HasSuffix(line, []byte("\n")) {
			_, err = fmt.Fprintln(output)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// readNewContent reads content from a file starting at lastPos position.
// It returns the new file position and any error encountered.
func readNewContent(filepath string, lastPos int64, callback func(content []byte) error) (int64, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return lastPos, err
	}
	defer file.Close()

	// Get current file size
	fileInfo, err := file.Stat()
	if err != nil {
		return lastPos, err
	}

	// Handle file truncation - reset position to beginning
	if fileInfo.Size() < lastPos {
		lastPos = 0
	}

	// If no new content, return early
	if fileInfo.Size() == lastPos {
		return lastPos, nil
	}

	// Seek to last read position
	_, err = file.Seek(lastPos, io.SeekStart)
	if err != nil {
		return lastPos, err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return lastPos, err
	}

	if err := callback(content); err != nil {
		return lastPos, err
	}

	return lastPos + int64(len(content)), nil
}

// fileExists checks if a file exists at the given path
func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

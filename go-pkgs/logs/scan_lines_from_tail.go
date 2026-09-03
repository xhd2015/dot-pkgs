package logs

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const defaultScanLinesFromTailChunk = 4096

// ScanLinesFromTailOptions controls ScanLinesFromTail.
type ScanLinesFromTailOptions struct {
	// ChunkSize is the first reverse-read window in bytes.
	// 0 means auto (4096). When a window has no complete line, the window
	// doubles until a newline is found or the start of the file is reached.
	ChunkSize int
}

// ScanLinesFromTail reads path from the end and calls fn for each complete
// line, newest to oldest. It is a one-shot reverse scan (not tail -f; use
// WatchLine to follow appends).
//
// Mid-line bytes at the leading edge of a window are never passed to fn; they
// are carried into the next older chunk. A final line without a trailing
// newline is still emitted. fn may stop early by returning stop=true.
func ScanLinesFromTail(path string, opts ScanLinesFromTailOptions, fn func(line string) (stop bool, err error)) error {
	if fn == nil {
		return fmt.Errorf("logs.ScanLinesFromTail: nil callback")
	}
	chunk := opts.ChunkSize
	if chunk <= 0 {
		chunk = defaultScanLinesFromTailChunk
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	size := fi.Size()
	if size == 0 {
		return nil
	}

	end := size
	var carry []byte
	for end > 0 {
		win := int64(chunk)
		if win > end {
			win = end
		}
		start := end - win
		buf := make([]byte, win)
		if _, err := f.ReadAt(buf, start); err != nil && err != io.EOF {
			return err
		}
		data := append(buf, carry...)

		nl := bytes.IndexByte(data, '\n')
		if nl < 0 {
			if start == 0 {
				if len(data) > 0 {
					stop, err := fn(string(data))
					if err != nil || stop {
						return err
					}
				}
				return nil
			}
			carry = data
			end = start
			if chunk < int(size) {
				next := chunk * 2
				if next < chunk {
					next = int(size)
				}
				chunk = next
			}
			continue
		}

		prefix := data[:nl]
		rest := data[nl+1:]

		stop, err := emitLinesNewestFirst(rest, fn)
		if err != nil || stop {
			return err
		}

		if start == 0 {
			if len(prefix) > 0 {
				stop, err := fn(string(prefix))
				if err != nil || stop {
					return err
				}
			}
			return nil
		}

		carry = append([]byte(nil), prefix...)
		end = start
		// After a successful window with newlines, reset chunk to the
		// configured/auto size for subsequent older windows.
		if opts.ChunkSize > 0 {
			chunk = opts.ChunkSize
		} else {
			chunk = defaultScanLinesFromTailChunk
		}
	}

	if len(carry) > 0 {
		stop, err := fn(string(carry))
		if err != nil || stop {
			return err
		}
	}
	return nil
}

func emitLinesNewestFirst(rest []byte, fn func(line string) (stop bool, err error)) (bool, error) {
	if len(rest) == 0 {
		return false, nil
	}
	// Split rest on \n; emit from last segment to first.
	// Trailing \n yields a final empty segment — skip empties.
	parts := bytes.Split(rest, []byte{'\n'})
	for i := len(parts) - 1; i >= 0; i-- {
		if len(parts[i]) == 0 {
			continue
		}
		stop, err := fn(string(parts[i]))
		if err != nil || stop {
			return stop, err
		}
	}
	return false, nil
}

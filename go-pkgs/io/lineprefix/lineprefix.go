// Package lineprefix decorates an io.Writer so each line is prefixed.
//
// Use at call sites when handing a writer to a child (e.g. cmd.Stderr),
// not by guessing origin from line content after the fact.
package lineprefix

import (
	"io"
	"sync"
)

// Writer inserts Prefix at the start of every line written to the underlying writer.
// Partial lines spanning Write calls are buffered. Safe for concurrent Write.
type Writer struct {
	mu      sync.Mutex
	w       io.Writer
	prefix  []byte
	atStart bool // next byte begins a line
}

// New returns a line-prefixing decorator around w.
// The prefix string is used exactly as given (include brackets and trailing space yourself).
// nil w uses io.Discard. Empty prefix still wraps (passthrough of line bytes).
func New(w io.Writer, prefix string) *Writer {
	if w == nil {
		w = io.Discard
	}
	return &Writer{
		w:       w,
		prefix:  []byte(prefix),
		atStart: true,
	}
}

// Bracket is New(w, "["+tag+"] "). Empty tag behaves like New(w, "").
func Bracket(w io.Writer, tag string) *Writer {
	if tag == "" {
		return New(w, "")
	}
	return New(w, "["+tag+"] ")
}

// Write implements io.Writer.
// On success it returns len(p) even when more bytes (the prefixes) were written underneath.
func (p *Writer) Write(b []byte) (int, error) {
	if p == nil {
		return 0, io.ErrClosedPipe
	}
	if len(b) == 0 {
		return 0, nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	consumed := 0
	for consumed < len(b) {
		if p.atStart {
			if len(p.prefix) > 0 {
				if _, err := p.w.Write(p.prefix); err != nil {
					return consumed, err
				}
			}
			p.atStart = false
		}
		i := indexByte(b[consumed:], '\n')
		if i < 0 {
			// No newline: append remainder to buf and underlying in one write of remainder.
			chunk := b[consumed:]
			if _, err := p.w.Write(chunk); err != nil {
				return consumed, err
			}
			consumed = len(b)
			break
		}
		// Include the newline in this line.
		end := consumed + i + 1
		chunk := b[consumed:end]
		if _, err := p.w.Write(chunk); err != nil {
			return consumed, err
		}
		consumed = end
		p.atStart = true
	}
	return len(b), nil
}

// Flush writes nothing extra for the current design: incomplete line bytes were
// already forwarded to the underlying writer (only the next line's prefix is deferred).
// Kept for API stability and so callers can defer Flush after a child exits.
func (p *Writer) Flush() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return nil
}

// Close calls Flush, then Close on the underlying writer if it implements io.Closer.
func (p *Writer) Close() error {
	if p == nil {
		return nil
	}
	if err := p.Flush(); err != nil {
		return err
	}
	p.mu.Lock()
	w := p.w
	p.mu.Unlock()
	if c, ok := w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Unwrap returns the underlying writer.
func (p *Writer) Unwrap() io.Writer {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.w
}

func indexByte(b []byte, c byte) int {
	for i, v := range b {
		if v == c {
			return i
		}
	}
	return -1
}

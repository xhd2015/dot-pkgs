package govet

import (
	"fmt"
	"strings"
)

type FileLenChecker struct {
	MaxLines int
}

func (c *FileLenChecker) Name() string {
	return "file-length"
}

func (c *FileLenChecker) CheckFile(filename string, src []byte) []Violation {
	if c.MaxLines <= 0 {
		return nil
	}

	lines := strings.Count(string(src), "\n")
	if len(src) > 0 && src[len(src)-1] != '\n' {
		lines++
	}

	if lines > c.MaxLines {
		return []Violation{{
			File:    filename,
			Line:    0,
			Col:     0,
			Message: fmt.Sprintf("file has %d lines, exceeds maximum of %d", lines, c.MaxLines),
			Checker: c.Name(),
		}}
	}
	return nil
}

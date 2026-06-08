package filelen

import (
	"fmt"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Checker struct {
	MaxLines int
}

func (c *Checker) Name() string {
	return "file-length"
}

func (c *Checker) CheckFile(filename string, src []byte) []types.Violation {
	if c.MaxLines <= 0 {
		return nil
	}

	lines := strings.Count(string(src), "\n")
	if len(src) > 0 && src[len(src)-1] != '\n' {
		lines++
	}

	if lines > c.MaxLines {
		return []types.Violation{{
			File:    filename,
			Line:    0,
			Col:     0,
			Message: fmt.Sprintf("file has %d lines, exceeds maximum of %d", lines, c.MaxLines),
			Checker: c.Name(),
		}}
	}
	return nil
}

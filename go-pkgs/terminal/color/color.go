package color

import (
	"errors"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Mode selects when ANSI color is applied.
type Mode int

const (
	Auto Mode = iota
	Always
	Never
)

// Resolve reports whether color should be enabled.
// Always is true and Never is false regardless of TTY or NO_COLOR.
// Auto is false when noColorEnv is non-empty, otherwise it follows stdoutIsTTY.
func Resolve(mode Mode, stdoutIsTTY bool, noColorEnv string) bool {
	switch mode {
	case Always:
		return true
	case Never:
		return false
	default:
		if noColorEnv != "" {
			return false
		}
		return stdoutIsTTY
	}
}

// ModeFromFlags maps --color / --no-color to a Mode.
// Both flags set is an error.
func ModeFromFlags(color, noColor bool) (Mode, error) {
	if color && noColor {
		return Auto, errors.New("--color and --no-color cannot be specified together")
	}
	if color {
		return Always, nil
	}
	if noColor {
		return Never, nil
	}
	return Auto, nil
}

// Code is one SGR sequence for Paint (and the convenience methods).
type Code string

const (
	Reset  Code = "\x1b[0m"
	Bold   Code = "\x1b[1m"
	Dim    Code = "\x1b[2m"
	Strike Code = "\x1b[9m"
	Red    Code = "\x1b[31m"
	Green  Code = "\x1b[32m"
	Yellow Code = "\x1b[33m"
	Blue   Code = "\x1b[34m"
	Gray   Code = "\x1b[90m"
)

// Style wraps text in SGR sequences when Enabled is true.
type Style struct {
	Enabled bool
}

// Paint prefixes text with codes then Reset. !Enabled or empty text is unchanged.
func (s Style) Paint(text string, codes ...Code) string {
	if !s.Enabled || text == "" {
		return text
	}
	var b strings.Builder
	for _, c := range codes {
		b.WriteString(string(c))
	}
	b.WriteString(text)
	b.WriteString(string(Reset))
	return b.String()
}

func (s Style) Green(text string) string  { return s.Paint(text, Green) }
func (s Style) Red(text string) string    { return s.Paint(text, Red) }
func (s Style) Yellow(text string) string { return s.Paint(text, Yellow) }
func (s Style) Blue(text string) string   { return s.Paint(text, Blue) }
func (s Style) Gray(text string) string   { return s.Paint(text, Gray) }
func (s Style) Bold(text string) string   { return s.Paint(text, Bold) }
func (s Style) Dim(text string) string    { return s.Paint(text, Dim) }
func (s Style) Strike(text string) string { return s.Paint(text, Strike) }

// WriterIsTTY reports whether w is a terminal. Writers without Fd() are false.
func WriterIsTTY(w io.Writer) bool {
	type fdWriter interface {
		Fd() uintptr
	}
	fw, ok := w.(fdWriter)
	if !ok {
		return false
	}
	return term.IsTerminal(int(fw.Fd()))
}

// Enabled reports whether color is enabled for os.Stdout and the process NO_COLOR.
func Enabled(mode Mode) bool {
	return Resolve(mode, WriterIsTTY(os.Stdout), os.Getenv("NO_COLOR"))
}

// EnabledFor reports whether color is enabled for w and the process NO_COLOR.
func EnabledFor(mode Mode, w io.Writer) bool {
	return Resolve(mode, WriterIsTTY(w), os.Getenv("NO_COLOR"))
}

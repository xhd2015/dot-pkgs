package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/getclipboard"
	"github.com/xhd2015/less-flags"
)

const help = `
Usage: get-clipboard [OPTIONS]

Read system clipboard content:
  - Plain text is printed to stdout
  - Image, SVG, HTML, and other binary formats are saved to a file

Options:
  -o, --output FILE   output file path (by default a timestamped name with auto extension)
  -n, --name NAME     write under /tmp/NAME.<ext>; if exists, use NAME-1.<ext>, NAME-2.<ext>, ...
      --open          after saving a file, run: open <path>
  -h, --help          show this help message
`

// openCmd runs macOS open for a path. Overridable in tests.
var openCmd = func(path string) error {
	return exec.Command("open", path).Run()
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "get-clipboard: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithOutput(args, os.Stdout)
}

func runWithOutput(args []string, out io.Writer) error {
	var output string
	var name string
	var doOpen bool
	_, err := lessflags.String("-o,--output", &output).
		String("-n,--name", &name).
		Bool("--open", &doOpen).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if output != "" && name != "" {
		return fmt.Errorf("--name and --output are mutually exclusive")
	}
	if name != "" {
		name, err = getclipboard.ValidateName(name)
		if err != nil {
			return err
		}
	}

	c, err := getclipboard.Read(getclipboard.System())
	if err != nil {
		return err
	}

	switch c.Kind {
	case getclipboard.KindEmpty:
		return fmt.Errorf("clipboard contains unsupported content")
	case getclipboard.KindUnsupported:
		if len(c.Available) > 0 {
			return fmt.Errorf("clipboard contains unsupported content (available: %s)", strings.Join(c.Available, ", "))
		}
		return fmt.Errorf("clipboard contains unsupported content")
	case getclipboard.KindText:
		// --open is ignored for plain text (no file to open) — CLI prints text.
		fmt.Fprint(out, string(c.Data))
		return nil
	default:
		res, err := c.Dump(getclipboard.DumpOptions{Output: output, Name: name, Dir: "/tmp"})
		if err != nil {
			return err
		}
		return printAndMaybeOpen(out, res.Path, doOpen)
	}
}

// printAndMaybeOpen prints the saved path, then optionally runs open <path>.
func printAndMaybeOpen(out io.Writer, path string, doOpen bool) error {
	fmt.Fprintln(out, path)
	if !doOpen {
		return nil
	}
	if err := openCmd(path); err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	return nil
}

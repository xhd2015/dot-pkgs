package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/xhd2015/less-flags"
	"golang.design/x/clipboard"
)

const help = `
Usage: get-clipboard [OPTIONS]

Read system clipboard content:
  - Plain text is printed to stdout
  - Images are saved to a file (auto-generated name, or via -o/--output)
  - Other content types produce an error

Options:
  -o, --output FILE   output file path for image (by default a timestamped name)
  -h, --help          show this help message
`

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
	_, err := lessflags.String("-o,--output", &output).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	err = clipboard.Init()
	if err != nil {
		return fmt.Errorf("init clipboard: %w", err)
	}

	imgData := clipboard.Read(clipboard.FmtImage)
	if len(imgData) > 0 {
		ext := detectImageFormat(imgData)
		if ext == "" {
			return fmt.Errorf("cannot determine image format (magic bytes unrecognized)")
		}
		filename := output
		if filename == "" {
			filename = generateFilename(imgData, ext)
		}
		if err := os.WriteFile(filename, imgData, 0644); err != nil {
			return fmt.Errorf("write image: %w", err)
		}
		fmt.Fprintln(out, filename)
		return nil
	}

	textData := clipboard.Read(clipboard.FmtText)
	if len(textData) > 0 {
		fmt.Fprint(out, string(textData))
		return nil
	}

	return fmt.Errorf("clipboard contains unsupported content")
}

func detectImageFormat(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	if data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' {
		return "png"
	}
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpg"
	}
	if len(data) >= 6 &&
		data[0] == 'G' && data[1] == 'I' && data[2] == 'F' && data[3] == '8' &&
		(data[4] == '7' || data[4] == '9') && data[5] == 'a' {
		return "gif"
	}
	if data[0] == 'B' && data[1] == 'M' {
		return "bmp"
	}
	if data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A && data[3] == 0x00 {
		return "tiff"
	}
	if data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00 && data[3] == 0x2A {
		return "tiff"
	}
	return ""
}

func generateFilename(data []byte, ext string) string {
	ts := time.Now().Format("2006-01-02-15-04-05")
	h := md5.Sum(data)
	hash := fmt.Sprintf("%x", h)[:8]
	return fmt.Sprintf("%s-clipboard-%s.%s", ts, hash, ext)
}

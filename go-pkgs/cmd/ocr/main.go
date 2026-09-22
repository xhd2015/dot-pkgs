package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/xhd2015/dot-pkgs/go-pkgs/ocr"
	"github.com/xhd2015/less-flags"
)

const help = `
Usage: ocr [OPTIONS] <image>

Recognize text in an image using macOS Apple Vision (OCR).

The Vision Swift helper is embedded and compiled on first use into the
user cache (typically ~/Library/Caches/dot-pkgs/ocr/).

Options:
  -h, --help   show this help message
`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "ocr: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	return runWith(args, out, ocr.Options{})
}

func runWith(args []string, out io.Writer, opts ocr.Options) error {
	remaining, err := lessflags.Help("-h,--help", help).Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		return fmt.Errorf("missing image path")
	}
	if len(remaining) > 1 {
		return fmt.Errorf("unexpected argument: %s", remaining[1])
	}

	res, err := ocr.RecognizeFile(context.Background(), remaining[0], opts)
	if err != nil {
		return err
	}
	if res.Text != "" {
		fmt.Fprintln(out, res.Text)
	}
	return nil
}

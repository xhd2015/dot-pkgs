package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/dupstat"
	"github.com/xhd2015/less-flags"
)

const help = `
Usage: code-dup-stat [OPTIONS] [DIR]

Detect similar functions across packages to spot potential code duplication.

Options:
  --ngram K          n-gram size (default: 5)
  --threshold T      similarity threshold 0.0-1.0 (default: 0.5)
  --algorithm ALGO   similarity algorithm: ngram (default) or wordstat
  --dir DIR          root directory to scan (default: go module root or current dir)
  -h, --help         show this help message
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "code-dup-stat: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithOutput(args, os.Stdout)
}

func runWithOutput(args []string, out io.Writer) error {
	var ngramFlag int
	var thresholdStr string
	var dirFlag string
	var algoFlag string

	_, err := lessflags.Int("--ngram", &ngramFlag).
		String("--threshold", &thresholdStr).
		String("--algorithm", &algoFlag).
		String("--dir", &dirFlag).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	k := ngramFlag
	if k <= 0 {
		k = 5
	}
	threshold := 0.5
	if thresholdStr != "" {
		threshold, err = strconv.ParseFloat(thresholdStr, 64)
		if err != nil {
			return fmt.Errorf("invalid threshold: %w", err)
		}
		if threshold < 0 || threshold > 1.0 {
			return fmt.Errorf("threshold must be between 0 and 1")
		}
	}

	dir := dirFlag
	if dir == "" {
		dir = "."
	}

	algorithm := algoFlag
	if algorithm == "" {
		algorithm = dupstat.AlgoNgram
	}
	if algorithm != dupstat.AlgoNgram && algorithm != dupstat.AlgoWordstat {
		return fmt.Errorf("unknown algorithm: %s (valid: ngram, wordstat)", algorithm)
	}

	groups, err := dupstat.Analyze(dir, k, threshold, algorithm)
	if err != nil {
		return err
	}

	printResults(out, groups, algorithm)
	return nil
}

func printResults(out io.Writer, groups []dupstat.Group, algorithm string) {
	if len(groups) == 0 {
		fmt.Fprintln(out, "No similar function pairs found.")
		return
	}

	for i, g := range groups {
		for j, p := range g.Pairs {
			if j == 0 {
				kindLabel := "CROSS-PACKAGE"
				if g.Kind == "same-package" {
					kindLabel = "SAME-PACKAGE"
				}
				fmt.Fprintf(out, "Group %d [%s]\n", i+1, kindLabel)
				fmt.Fprintln(out, strings.Repeat("-", 50))
			}

			aFile := shortPath(p.FuncA.File, p.FuncA.PkgPath)
			bFile := shortPath(p.FuncB.File, p.FuncB.PkgPath)

			aSig := formatFuncSig(p.FuncA)
			bSig := formatFuncSig(p.FuncB)

			fmt.Fprintf(out, "  %s:%d  %s\n", aFile, p.FuncA.Line, aSig)
			fmt.Fprintf(out, "  %s:%d  %s\n", bFile, p.FuncB.Line, bSig)

			if algorithm == dupstat.AlgoWordstat {
				maxScore := maxFloat64(p.WordJaccard, p.WordContainment, p.WordOverlap)
				fmt.Fprintf(out, "  similarity: %.2f  wordstat(j=%.2f c=%.2f o=%.2f)\n",
					maxScore,
					p.WordJaccard, p.WordContainment, p.WordOverlap,
				)
			} else {
				maxScore := maxFloat64(
					p.RawJaccard, p.RawContainment,
					p.NormJaccard, p.NormContainment,
					p.MixedJaccard, p.MixedContainment,
				)
				fmt.Fprintf(out, "  similarity: %.2f  raw(j=%.2f c=%.2f) norm(j=%.2f c=%.2f) mixed(j=%.2f c=%.2f)\n",
					maxScore,
					p.RawJaccard, p.RawContainment,
					p.NormJaccard, p.NormContainment,
					p.MixedJaccard, p.MixedContainment,
				)
			}
			if j < len(g.Pairs)-1 {
				fmt.Fprintln(out)
			}
		}
		if i < len(groups)-1 {
			fmt.Fprintln(out)
		}
	}
}

func maxFloat64(vals ...float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func shortPath(absPath string, pkgPath string) string {
	parts := strings.Split(pkgPath, "/")
	if len(parts) > 0 && parts[0] != "" {
		return filepath.Join(parts...)
	}
	return filepath.Base(absPath)
}

func formatFuncSig(fn *dupstat.Function) string {
	if fn.Receiver != "" {
		return fmt.Sprintf("func (%s) %s()", fn.Receiver, fn.Name)
	}
	return fmt.Sprintf("func %s()", fn.Name)
}

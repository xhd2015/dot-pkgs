package govet

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/filelen"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/pattern"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/stdflag"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Violation = types.Violation
type Config = types.Config
type FileChecker = types.FileChecker
type ASTChecker = types.ASTChecker

var ResolvePatterns = pattern.ResolvePatterns

func Run(cfg Config, dirs []string) ([]Violation, error) {
	var fileCheckers []FileChecker
	var astCheckers []ASTChecker

	if cfg.FileMaxLines > 0 {
		fileCheckers = append(fileCheckers, &filelen.Checker{MaxLines: cfg.FileMaxLines})
	}
	astCheckers = append(astCheckers, &stdflag.Checker{})

	excludeSet := make(map[string]bool, len(cfg.Excludes))
	for _, e := range cfg.Excludes {
		excludeSet[e] = true
	}

	var allViolations []Violation

	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				base := filepath.Base(path)
				if base == "vendor" || base == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			for _, fc := range fileCheckers {
				if excludeSet[fc.Name()] {
					continue
				}
				violations := fc.CheckFile(path, src)
				allViolations = append(allViolations, violations...)
			}

			if len(astCheckers) > 0 {
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, path, src, parser.ImportsOnly)
				if err != nil {
					return nil
				}
				for _, ac := range astCheckers {
					if excludeSet[ac.Name()] {
						continue
					}
					violations := ac.CheckAST(fset, f)
					allViolations = append(allViolations, violations...)
				}
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", dir, err)
		}
	}

	return allViolations, nil
}

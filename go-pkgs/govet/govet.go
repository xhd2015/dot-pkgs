package govet

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/builtingovet"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/cleanfunc"
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

func Run(cfg Config) ([]Violation, error) {
	var fileCheckers []FileChecker
	var astCheckers []ASTChecker

	if cfg.FileMaxLines > 0 {
		fileCheckers = append(fileCheckers, &filelen.Checker{MaxLines: cfg.FileMaxLines})
	}
	astCheckers = append(astCheckers, &stdflag.Checker{}, &cleanfunc.Checker{})

	builtinVetChecker := &builtingovet.Checker{}

	excludeSet := make(map[string]bool, len(cfg.Excludes))
	for _, e := range cfg.Excludes {
		excludeSet[e] = true
	}

	paths := cfg.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var allViolations []Violation

	for _, p := range paths {
		err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
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

			violations, err := checkFile(path, fileCheckers, astCheckers, excludeSet)
			if err != nil {
				return err
			}
			allViolations = append(allViolations, violations...)

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", p, err)
		}
	}

	runBuiltinVet(paths, builtinVetChecker, excludeSet, &allViolations)

	return allViolations, nil
}

func checkFile(path string, fileCheckers []FileChecker, astCheckers []ASTChecker, excludeSet map[string]bool) ([]Violation, error) {
	var allViolations []Violation

	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
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
		f, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return allViolations, nil
		}
		for _, ac := range astCheckers {
			if excludeSet[ac.Name()] {
				continue
			}
			violations := ac.CheckAST(fset, f)
			allViolations = append(allViolations, violations...)
		}
	}

	return allViolations, nil
}

func runBuiltinVet(paths []string, checker *builtingovet.Checker, excludeSet map[string]bool, allViolations *[]Violation) {
	if excludeSet[checker.Name()] {
		return
	}
	dirs := uniqueDirs(paths)
	for _, dir := range dirs {
		violations, err := checker.Check(dir)
		if err != nil {
			continue
		}
		*allViolations = append(*allViolations, violations...)
	}
}

func uniqueDirs(paths []string) []string {
	dirSet := make(map[string]bool)
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			dirSet[p] = true
		} else {
			dirSet[filepath.Dir(p)] = true
		}
	}
	var dirs []string
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	var result []string
	for _, dir := range dirs {
		isChild := false
		for _, other := range dirs {
			if dir == other {
				continue
			}
			rel, err := filepath.Rel(other, dir)
			if err != nil {
				continue
			}
			if rel != "." && !strings.HasPrefix(rel, "..") {
				isChild = true
				break
			}
		}
		if !isChild {
			result = append(result, dir)
		}
	}
	return result
}

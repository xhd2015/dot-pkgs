package govet

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/builtingovet"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/cleanfunc"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/filelen"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/stdflag"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Violation = types.Violation
type Config = types.Config
type FileChecker = types.FileChecker
type ASTChecker = types.ASTChecker

func Run(cfg Config) ([]Violation, error) {
	var fileCheckers []FileChecker
	var astCheckers []ASTChecker

	if cfg.FileMaxLines > 0 {
		fileCheckers = append(fileCheckers, &filelen.Checker{MaxLines: cfg.FileMaxLines})
	}
	astCheckers = append(astCheckers, &stdflag.Checker{}, &stdflag.ManualFlagChecker{}, &cleanfunc.Checker{})

	builtinVetChecker := &builtingovet.Checker{}

	excludeCheckerSet := make(map[string]bool, len(cfg.ExcludeCheckers))
	for _, e := range cfg.ExcludeCheckers {
		excludeCheckerSet[e] = true
	}

	var allViolations []Violation

	for _, path := range cfg.Files {
		violations, err := checkFile(path, fileCheckers, astCheckers, excludeCheckerSet)
		if err != nil {
			return nil, fmt.Errorf("check %s: %w", path, err)
		}
		allViolations = append(allViolations, violations...)
	}

	runBuiltinVet(cfg.Files, builtinVetChecker, excludeCheckerSet, &allViolations)

	return allViolations, nil
}

func checkFile(path string, fileCheckers []FileChecker, astCheckers []ASTChecker, excludeCheckerSet map[string]bool) ([]Violation, error) {
	var allViolations []Violation

	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	for _, fc := range fileCheckers {
		if excludeCheckerSet[fc.Name()] {
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
			if excludeCheckerSet[ac.Name()] {
				continue
			}
			violations := ac.CheckAST(fset, f)
			allViolations = append(allViolations, violations...)
		}
	}

	return allViolations, nil
}

func runBuiltinVet(files []string, checker *builtingovet.Checker, excludeCheckerSet map[string]bool, allViolations *[]Violation) {
	if excludeCheckerSet[checker.Name()] {
		return
	}
	violations, err := checker.Check(files)
	if err != nil {
		return
	}
	*allViolations = append(*allViolations, violations...)
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/dupstat"
)

const threshold = 0.3
const ngramK = 5

type testCase struct {
	name      string
	dir       string
	algorithm string
	checks    []func(groups []dupstat.Group) error
}

func main() {
	failed := false

	for _, tc := range allTests() {
		fmt.Printf("  %s ... ", tc.name)
		algo := tc.algorithm
		if algo == "" {
			algo = dupstat.AlgoNgram
		}
		groups, err := dupstat.Analyze(tc.dir, ngramK, threshold, algo)
		if err != nil {
			fmt.Printf("FAIL\n    error: %v\n", err)
			failed = true
			continue
		}
		if err := runChecks(groups, tc.checks); err != nil {
			fmt.Printf("FAIL\n    %v\n", err)
			failed = true
		} else {
			fmt.Println("ok")
		}
	}

	if failed {
		os.Exit(1)
	}
	fmt.Println("\nAll tests passed.")
}

func runChecks(groups []dupstat.Group, checks []func([]dupstat.Group) error) error {
	for _, check := range checks {
		if err := check(groups); err != nil {
			return err
		}
	}
	return nil
}

func findPairByNames(groups []dupstat.Group, nameA, nameB string) (*dupstat.FuncPair, *dupstat.Group, bool) {
	for i := range groups {
		for j := range groups[i].Pairs {
			p := &groups[i].Pairs[j]
			if (p.FuncA.Name == nameA && p.FuncB.Name == nameB) ||
				(p.FuncA.Name == nameB && p.FuncB.Name == nameA) {
				return p, &groups[i], true
			}
		}
	}
	return nil, nil, false
}

func allTests() []testCase {
	baseDir := filepath.Join("cmd", "code-dup-stat", "testdata")

	return []testCase{
		{
			name: "cross-pkg-dup",
			dir:  filepath.Join(baseDir, "cross-pkg-dup"),
			checks: []func(groups []dupstat.Group) error{
				func(groups []dupstat.Group) error {
					p, g, ok := findPairByNames(groups, "ValidateEmail", "CheckEmail")
					if !ok {
						return fmt.Errorf("expected ValidateEmail/CheckEmail pair")
					}
					if g.Kind != "cross-package" {
						return fmt.Errorf("expected cross-package group, got %s", g.Kind)
					}
					if p.NormJaccard < 0.9 {
						return fmt.Errorf("expected high normalized jaccard, got %.2f", p.NormJaccard)
					}
					return nil
				},
				func(groups []dupstat.Group) error {
					p, g, ok := findPairByNames(groups, "HashPassword", "EncryptPassword")
					if !ok {
						return fmt.Errorf("expected HashPassword/EncryptPassword pair")
					}
					if g.Kind != "cross-package" {
						return fmt.Errorf("expected cross-package group, got %s", g.Kind)
					}
					if p.NormJaccard < 0.9 {
						return fmt.Errorf("expected high normalized jaccard, got %.2f", p.NormJaccard)
					}
					return nil
				},
			},
		},
		{
			name: "same-pkg-dup",
			dir:  filepath.Join(baseDir, "same-pkg-dup"),
			checks: []func(groups []dupstat.Group) error{
				func(groups []dupstat.Group) error {
					_, g, ok := findPairByNames(groups, "ReadConfig", "WriteConfig")
					if !ok {
						return fmt.Errorf("expected ReadConfig/WriteConfig pair")
					}
					if g.Kind != "same-package" {
						return fmt.Errorf("expected same-package group, got %s", g.Kind)
					}
					return nil
				},
				func(groups []dupstat.Group) error {
					for _, g := range groups {
						if g.Kind == "cross-package" {
							return fmt.Errorf("unexpected cross-package group")
						}
					}
					return nil
				},
			},
		},
		{
			name: "structural-dup",
			dir:  filepath.Join(baseDir, "structural-dup"),
			checks: []func(groups []dupstat.Group) error{
				func(groups []dupstat.Group) error {
					p, _, ok := findPairByNames(groups, "ProcessOrder", "ProcessPayment")
					if !ok {
						return fmt.Errorf("expected ProcessOrder/ProcessPayment pair")
					}
					if p.NormContainment < 0.5 {
						return fmt.Errorf("expected norm containment >= 0.5, got %.2f", p.NormContainment)
					}
					if p.RawJaccard >= p.NormJaccard && p.RawJaccard > 0.6 {
						return fmt.Errorf("expected raw score lower than normalized for structural dup")
					}
					return nil
				},
			},
		},
		{
			name: "no-dup",
			dir:  filepath.Join(baseDir, "no-dup"),
			checks: []func(groups []dupstat.Group) error{
				func(groups []dupstat.Group) error {
					if len(groups) > 0 {
						var pairs []string
						for _, g := range groups {
							for _, p := range g.Pairs {
								pairs = append(pairs, fmt.Sprintf("%s/%s", p.FuncA.Name, p.FuncB.Name))
							}
						}
						return fmt.Errorf("expected no groups, got: %s", strings.Join(pairs, ", "))
					}
					return nil
				},
			},
		},
		{
			name: "subset-dup",
			dir:  filepath.Join(baseDir, "subset-dup"),
			checks: []func(groups []dupstat.Group) error{
				func(groups []dupstat.Group) error {
					p, _, ok := findPairByNames(groups, "DoQuick", "DoFull")
					if !ok {
						return fmt.Errorf("expected DoQuick/DoFull pair")
					}
					if p.RawContainment <= p.RawJaccard {
						return fmt.Errorf("expected containment > jaccard for subset dup, got c=%.2f j=%.2f", p.RawContainment, p.RawJaccard)
					}
					return nil
				},
			},
		},
		{
			name:      "wordstat-dup",
			dir:       filepath.Join(baseDir, "wordstat-dup"),
			algorithm: dupstat.AlgoWordstat,
			checks: []func(groups []dupstat.Group) error{
				func(groups []dupstat.Group) error {
					p, _, ok := findPairByNames(groups, "ProcessUser", "HandleRequest")
					if !ok {
						return fmt.Errorf("expected ProcessUser/HandleRequest pair")
					}
					if p.WordJaccard < 0.5 {
						return fmt.Errorf("expected wordstat jaccard >= 0.5, got %.2f", p.WordJaccard)
					}
					if p.WordContainment < 0.5 {
						return fmt.Errorf("expected wordstat containment >= 0.5, got %.2f", p.WordContainment)
					}
					return nil
				},
			},
		},
	}
}

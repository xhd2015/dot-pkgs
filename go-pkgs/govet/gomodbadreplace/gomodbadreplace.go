package gomodbadreplace

import (
	"fmt"
	"os"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
	"golang.org/x/mod/modfile"
)

type Checker struct{}

func (c *Checker) Name() string {
	return "gomod-bad-replace"
}

func (c *Checker) Check(goModPath string) ([]types.Violation, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", goModPath, err)
	}

	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", goModPath, err)
	}

	var violations []types.Violation
	for _, rep := range f.Replace {
		if rep.New.Version == "" {
			violations = append(violations, types.Violation{
				File:    goModPath,
				Line:    0,
				Col:     0,
				Message: fmt.Sprintf("local replace directive: %s => %s (replace to a specific version or another module instead)", rep.Old.Path, rep.New.Path),
				Checker: c.Name(),
			})
		}
	}
	return violations, nil
}

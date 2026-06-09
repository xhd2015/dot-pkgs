package builtingovet

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Checker struct{}

func (c *Checker) Name() string {
	return "builtin-go-vet"
}

var vetLineRe = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(.*)`)

func (c *Checker) Check(files []string) ([]types.Violation, error) {
	if len(files) == 0 {
		return nil, nil
	}
	args := append([]string{"vet"}, files...)
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("go vet: %w", err)
		}
	}

	return parseVetOutput(output, c.Name()), nil
}

func parseVetOutput(output []byte, checkerName string) []types.Violation {
	var violations []types.Violation
	for _, line := range bytes.Split(output, []byte("\n")) {
		matches := vetLineRe.FindSubmatch(line)
		if matches == nil {
			continue
		}
		file := string(matches[1])
		lineNo, err1 := strconv.Atoi(string(matches[2]))
		colNo, err2 := strconv.Atoi(string(matches[3]))
		if err1 != nil || err2 != nil {
			continue
		}
		msg := string(matches[4])
		violations = append(violations, types.Violation{
			File:    file,
			Line:    lineNo,
			Col:     colNo,
			Message: msg,
			Checker: checkerName,
		})
	}
	return violations
}

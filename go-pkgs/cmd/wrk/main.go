package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	args := os.Args[1:]
	if dir, remaining, ok := extractDir(args); ok {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("resolve dir: %w", err)
		}
		if _, err := os.Stat(absDir); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("wrk: %s does not exist", absDir)
			}
			return fmt.Errorf("stat dir: %w", err)
		}
		cwd = absDir
		args = remaining
	}

	return wrkcli.Run(cwd, args)
}

func extractDir(args []string) (dir string, remaining []string, ok bool) {
	var flags []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--dep" {
			flags = append(flags, arg)
			if i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			continue
		}
		return arg, append(flags, args[i+1:]...), true
	}
	return "", flags, false
}
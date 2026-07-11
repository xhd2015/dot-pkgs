package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli"
)

func main() {
	if err := run(); err != nil {
		var exitErr wrkcli.ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	return wrkcli.Run(os.Args[1:])
}

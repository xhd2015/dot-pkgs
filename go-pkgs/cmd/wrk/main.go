package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	return wrkcli.Run(os.Args[1:])
}

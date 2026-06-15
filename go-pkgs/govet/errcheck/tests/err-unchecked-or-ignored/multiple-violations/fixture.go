package p

import (
	"os"
	"strconv"
)

func f() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".agent-hub"
	}

	n, err := strconv.Atoi("42")
	if err != nil {
		return "0"
	}
	_ = n
	return home
}

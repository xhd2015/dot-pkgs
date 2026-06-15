package p

import (
	"fmt"
	"os"
)

func f() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home, nil
}

package p

import (
	"fmt"
	"os"
)

func f() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("new error")
	}
	return home, nil
}

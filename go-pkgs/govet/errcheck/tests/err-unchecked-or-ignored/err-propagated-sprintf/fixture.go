package p

import (
	"fmt"
	"os"
)

func f() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf("home err: %v", err), nil
	}
	return home, nil
}

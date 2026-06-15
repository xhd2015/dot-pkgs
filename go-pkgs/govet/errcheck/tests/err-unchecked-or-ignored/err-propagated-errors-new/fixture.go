package p

import (
	"errors"
	"os"
)

func f() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New(err.Error())
	}
	return home, nil
}

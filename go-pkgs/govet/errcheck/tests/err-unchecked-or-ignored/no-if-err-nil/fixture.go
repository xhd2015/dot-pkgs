package p

import "os"

func f() (string, error) {
	home, err := os.UserHomeDir()
	return home, err
}

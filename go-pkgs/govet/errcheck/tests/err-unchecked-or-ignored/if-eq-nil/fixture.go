package p

import "os"

func f() string {
	home, err := os.UserHomeDir()
	if err == nil {
		return home
	}
	return ""
}

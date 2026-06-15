package p

import "os"

var defaultHome = ".agent-hub"

func f() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultHome
	}
	return home
}

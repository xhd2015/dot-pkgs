package p

import "os"

func f(x int) string {
	home, err := os.UserHomeDir()
	if err != nil {
		if x > 0 {
			return "positive"
		}
		return "fallback"
	}
	return home
}

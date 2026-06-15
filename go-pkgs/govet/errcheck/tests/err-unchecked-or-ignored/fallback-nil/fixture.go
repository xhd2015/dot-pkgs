package p

import "os"

func f() interface{} {
	_, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return "ok"
}

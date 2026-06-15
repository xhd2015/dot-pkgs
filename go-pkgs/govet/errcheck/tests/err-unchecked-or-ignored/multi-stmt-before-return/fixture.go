package p

import (
	"log"
	"os"
)

func f() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("failed to get home")
		return ".agent-hub"
	}
	return home
}

//go:build !darwin || !cgo

package chrome

import "fmt"

func axIsTrusted() bool { return false }

func axClickNamed(appName, name string) error {
	return fmt.Errorf("chrome: AX click requires darwin+cgo (app=%s name=%s)", appName, name)
}

func axExistsNamed(appName, name string) bool { return false }

func axCountNamed(appName, name string) int { return -1 }

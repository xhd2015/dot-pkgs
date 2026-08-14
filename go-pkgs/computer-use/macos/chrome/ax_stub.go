//go:build !darwin || !cgo

package chrome

import "fmt"

func axIsTrusted() bool { return false }

func axClickNamed(appName, name string) error {
	return fmt.Errorf("chrome: AX click requires darwin+cgo (app=%s name=%s)", appName, name)
}

func axExistsNamed(appName, name string) bool { return false }

func axCountNamed(appName, name string) int { return -1 }

func axCollectNamedCards(appName, verifyName string) []axExtCard {
	_, _ = appName, verifyName
	return nil
}

func axQuartzClick(x, y float64) { _, _ = x, y }

func axClickTopRightRemove(appName string) bool { _ = appName; return false }

// axExtCard is defined here when AX CGO is unavailable so remove_*.go can compile.
type axExtCard struct {
	Version string
	CX, CY  float64
}

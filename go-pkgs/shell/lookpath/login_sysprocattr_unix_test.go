//go:build unix

package lookpath

import "testing"

func TestLoginSysProcAttr_Setsid(t *testing.T) {
	attr := loginSysProcAttr()
	if attr == nil {
		t.Fatal("loginSysProcAttr() = nil")
	}
	if !attr.Setsid {
		t.Fatal("Setsid = false, want true")
	}
}

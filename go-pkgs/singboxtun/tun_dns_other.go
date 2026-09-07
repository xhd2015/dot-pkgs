//go:build !darwin

package singboxtun

func restoreStuckTunDNS() error {
	return nil
}

func configurePlatformTunDNS() (restore func(), err error) {
	return func() {}, nil
}
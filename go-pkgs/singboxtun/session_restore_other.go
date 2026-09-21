//go:build !darwin

package singboxtun

func saveTunSessionSnapshot(service string, previousDNS []string, dnsTouched bool, previousProxy serviceProxyState, proxyTouched bool) error {
	return nil
}

func clearTunSessionSnapshot() {}

func RestoreTunSessionSideEffects() error {
	return nil
}

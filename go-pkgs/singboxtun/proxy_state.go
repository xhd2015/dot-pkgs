package singboxtun

// Shared proxy snapshot types used by foreground restore on every GOOS.
// Darwin fills them from networksetup; other platforms use zero values.

type proxyEndpoint struct {
	enabled bool
	server  string
	port    int
}

type serviceProxyState struct {
	web    proxyEndpoint
	secure proxyEndpoint
	socks  proxyEndpoint
}

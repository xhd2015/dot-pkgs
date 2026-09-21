package singboxtun

func buildTunConfigOptions(localSocksPort int, skipBind bool) *BuildConfigOptions {
	opts := &BuildConfigOptions{
		LocalSocksPort:    localSocksPort,
		SkipBindInterface: skipBind,
	}
	if !skipBind {
		if iface := defaultOutboundBindInterface(); iface != "" {
			opts.BindInterface = iface
		}
	}
	return opts
}

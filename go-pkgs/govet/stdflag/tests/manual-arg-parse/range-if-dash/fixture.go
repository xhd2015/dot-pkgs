package p

func run(args []string) {
	for _, arg := range args {
		if arg == "--debug" {
			_ = arg
		}
	}
}

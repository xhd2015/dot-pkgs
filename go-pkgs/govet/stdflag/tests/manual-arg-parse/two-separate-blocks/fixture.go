package p

func run(args []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			_ = args[i+1]
		}
	}
	for _, arg := range args {
		if arg == "--debug" {
			_ = arg
		}
	}
}

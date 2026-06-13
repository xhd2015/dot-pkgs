package p

func run(args []string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--verbose" {
			_ = args[i+1]
		}
	}
}

package p

func run(args []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "daemon":
			_ = args[i+1]
		case "notify":
			_ = args[i+1]
		}
	}
}

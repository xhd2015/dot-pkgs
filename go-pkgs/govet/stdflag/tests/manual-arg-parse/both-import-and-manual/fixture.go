package p

import "flag"

func run(args []string) {
	for _, arg := range args {
		switch arg {
		case "--verbose":
			_ = arg
		}
	}
	_ = flag.String
}

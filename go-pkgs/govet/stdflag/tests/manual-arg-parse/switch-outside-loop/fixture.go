package p

func main() {
	args := []string{"--json", "data"}
	switch args[0] {
	case "--json":
		_ = args[1]
	}
}

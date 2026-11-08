package reclaim

// ANSI color helpers. Only applied when UseColor is true.
const (
	ansiReset  = "\033[0m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
	ansiGray   = "\033[90m"
)

func colorize(use bool, code, s string) string {
	if !use || s == "" {
		return s
	}
	return code + s + ansiReset
}

func green(use bool, s string) string  { return colorize(use, ansiGreen, s) }
func yellow(use bool, s string) string { return colorize(use, ansiYellow, s) }
func red(use bool, s string) string    { return colorize(use, ansiRed, s) }
func gray(use bool, s string) string   { return colorize(use, ansiGray, s) }

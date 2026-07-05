package wrkcli

const (
	ansiRed    = "\x1b[31m"
	ansiOrange = "\x1b[33m"
	ansiReset  = "\x1b[0m"
)

func colorize(s, code string) string {
	return code + s + ansiReset
}
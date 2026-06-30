package ptywrap

import (
	"bytes"
	"regexp"
)

const maxScrollback = 256 * 1024

var terminalQueryRe = regexp.MustCompile(
	`\x1b\]\d+;[^\x07\x1b]*\?\x07` +
		`|\x1b\]\d+;[^\x07\x1b]*\?\x1b\\` +
		`|\x1b\[[>=]?c` +
		`|\x1b\[[56]n` +
		`|\x1b\[\?\d+\$p`,
)

func stripTerminalQueries(data []byte) []byte {
	return terminalQueryRe.ReplaceAll(data, nil)
}

func isAlternateScreenActive(data []byte) bool {
	enterSeq := []byte("\x1b[?1049h")
	exitSeq := []byte("\x1b[?1049l")
	lastEnter := bytes.LastIndex(data, enterSeq)
	if lastEnter == -1 {
		return false
	}
	lastExit := bytes.LastIndex(data, exitSeq)
	return lastExit < lastEnter
}

func stripAlternateScreenPairs(data []byte) []byte {
	enterSeq := []byte("\x1b[?1049h")
	exitSeq := []byte("\x1b[?1049l")
	for {
		enterIdx := bytes.Index(data, enterSeq)
		if enterIdx == -1 {
			break
		}
		exitIdx := bytes.Index(data[enterIdx:], exitSeq)
		if exitIdx == -1 {
			break
		}
		endIdx := enterIdx + exitIdx + len(exitSeq)
		data = append(data[:enterIdx], data[endIdx:]...)
	}
	return data
}
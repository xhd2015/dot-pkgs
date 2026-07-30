package proc

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

// ParsePSOutput parses `ps -ax -o pid=,ppid=,command=` style output into rows.
// Leading pid/ppid integers; remainder is Cmd. Empty lines and lines lacking
// two leading integers are skipped. Never panics.
func ParsePSOutput(out []byte) []Proc {
	var procs []Proc
	sc := bufio.NewScanner(bytes.NewReader(out))
	// Long command lines.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// pid and ppid are leading integers; remainder is command.
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		cmd := ""
		if len(fields) > 2 {
			// fields-based join loses multiple spaces; acceptable for classify.
			cmd = strings.Join(fields[2:], " ")
		}
		procs = append(procs, Proc{PID: pid, PPID: ppid, Cmd: cmd})
	}
	return procs
}

// ParseLsofFn parses `lsof -Fn` output into unique absolute open-file paths
// (first-seen order). Only `n…` name fields that are absolute paths other than
// `/` are kept; junk / non-`n` lines are ignored.
func ParseLsofFn(out []byte) []string {
	var paths []string
	seen := map[string]bool{}
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if len(line) < 2 || line[0] != 'n' {
			continue
		}
		path := line[1:]
		if path == "" || path == "/" {
			continue
		}
		if strings.HasPrefix(path, "/") && !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths
}

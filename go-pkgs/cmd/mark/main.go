package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/xhd2015/dot-pkgs/go-pkgs/libmark"
)

// used on terminal, like:
//
//	$ mark "I'm still waiting for result"
func main() {
	store := libmark.Default()
	pid := os.Getpid()
	if err := store.WriteLive(libmark.Record{
		PID:     pid,
		Content: strings.Join(os.Args[1:], " "),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "mark: %v\n", err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-ch

	if err := store.Archive(pid); err != nil {
		fmt.Fprintf(os.Stderr, "mark: %v\n", err)
		os.Exit(1)
	}
}

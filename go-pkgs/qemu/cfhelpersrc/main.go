// Command qemu-cf-helper runs guest cloudflared orchestration on the qemu host.
// Invoked as: qemu-cf-helper <action> [args]
// Config via flags (preferred) or env; stdout is key=value lines for Manager.ParseKV.
package main

import (
	"fmt"
	"os"
)

func main() {
	cfg := loadConfig(os.Args[1:])
	if cfg.err != nil {
		fmt.Fprintln(os.Stderr, cfg.err)
		os.Exit(2)
	}
	h := &helper{cfg: cfg.cfg}
	if err := h.run(cfg.action, cfg.args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

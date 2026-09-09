package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type config struct {
	Dir           string
	User          string
	SSHPort       int
	CFBin         string
	CFPid         string
	CFLog         string
	CFLoginLog    string
	CFURLFile     string
	Cert          string
	OldLog        string
	CFVersion     string
	DefaultOrigin string
	QemuPIDFile   string
}

type parsedArgs struct {
	cfg    config
	action string
	args   []string
	err    error
}

func defaults() config {
	return config{
		Dir:           envOr("QEMU_GUEST_DIR", "/root/qemu-guest"),
		User:          envOr("QEMU_GUEST_USER", "debian"),
		SSHPort:       envIntOr("QEMU_GUEST_SSH_PORT", 22221),
		CFBin:         envOr("QEMU_CF_BIN", "/usr/local/bin/cloudflared"),
		CFPid:         envOr("QEMU_CF_PID", "/var/run/work-qemu-cloudflared.pid"),
		CFLog:         envOr("QEMU_CF_LOG", "/var/log/work-qemu-cloudflared.log"),
		CFLoginLog:    "/var/log/work-qemu-cloudflared-login.log",
		CFURLFile:     "/root/.cloudflared/work-qemu-url",
		Cert:          envOr("QEMU_CF_CERT", "/root/.cloudflared/cert.pem"),
		OldLog:        "/var/log/cloudflared.log",
		CFVersion:     envOr("QEMU_CF_VERSION", "2025.11.1"),
		DefaultOrigin: envOr("QEMU_CF_ORIGIN_URL", "http://127.0.0.1:8080"),
		QemuPIDFile:   "", // filled after Dir known
	}
}

func loadConfig(argv []string) parsedArgs {
	cfg := defaults()
	i := 0
	for i < len(argv) {
		a := argv[i]
		if a == "--" {
			i++
			break
		}
		if !strings.HasPrefix(a, "--") {
			break
		}
		switch {
		case a == "--dir" && i+1 < len(argv):
			i++
			cfg.Dir = argv[i]
		case strings.HasPrefix(a, "--dir="):
			cfg.Dir = strings.TrimPrefix(a, "--dir=")
		case a == "--user" && i+1 < len(argv):
			i++
			cfg.User = argv[i]
		case strings.HasPrefix(a, "--user="):
			cfg.User = strings.TrimPrefix(a, "--user=")
		case a == "--port" && i+1 < len(argv):
			i++
			cfg.SSHPort = atoiDefault(argv[i], cfg.SSHPort)
		case strings.HasPrefix(a, "--port="):
			cfg.SSHPort = atoiDefault(strings.TrimPrefix(a, "--port="), cfg.SSHPort)
		case a == "--cf-bin" && i+1 < len(argv):
			i++
			cfg.CFBin = argv[i]
		case strings.HasPrefix(a, "--cf-bin="):
			cfg.CFBin = strings.TrimPrefix(a, "--cf-bin=")
		case a == "--cf-pid" && i+1 < len(argv):
			i++
			cfg.CFPid = argv[i]
		case strings.HasPrefix(a, "--cf-pid="):
			cfg.CFPid = strings.TrimPrefix(a, "--cf-pid=")
		case a == "--cf-log" && i+1 < len(argv):
			i++
			cfg.CFLog = argv[i]
		case strings.HasPrefix(a, "--cf-log="):
			cfg.CFLog = strings.TrimPrefix(a, "--cf-log=")
		case a == "--cert" && i+1 < len(argv):
			i++
			cfg.Cert = argv[i]
		case strings.HasPrefix(a, "--cert="):
			cfg.Cert = strings.TrimPrefix(a, "--cert=")
		case a == "--cf-ver" && i+1 < len(argv):
			i++
			cfg.CFVersion = argv[i]
		case strings.HasPrefix(a, "--cf-ver="):
			cfg.CFVersion = strings.TrimPrefix(a, "--cf-ver=")
		case a == "--default-origin" && i+1 < len(argv):
			i++
			cfg.DefaultOrigin = argv[i]
		case strings.HasPrefix(a, "--default-origin="):
			cfg.DefaultOrigin = strings.TrimPrefix(a, "--default-origin=")
		case a == "-h" || a == "--help":
			fmt.Fprint(os.Stdout, helpText)
			return parsedArgs{action: "help"}
		default:
			return parsedArgs{err: fmt.Errorf("unknown flag %s", a)}
		}
		i++
	}
	rest := argv[i:]
	if len(rest) == 0 {
		return parsedArgs{err: fmt.Errorf("action required")}
	}
	cfg.QemuPIDFile = cfg.Dir + "/qemu.pid"
	return parsedArgs{cfg: cfg, action: rest[0], args: rest[1:]}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envIntOr(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		return atoiDefault(v, def)
	}
	return def
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

const helpText = `Usage: qemu-cf-helper [flags] <action> [args]

Actions: status | start | stop | logs | purge | login | list |
         named-start | named-apply | named-drop-canonical

Flags:
  --dir PATH              qemu guest host dir (default /root/qemu-guest)
  --user NAME             guest ssh user (default debian)
  --port N                guest ssh port (default 22221)
  --cf-bin PATH           cloudflared in guest
  --cf-pid PATH           pid file in guest
  --cf-log PATH           log file in guest
  --cert PATH             cert.pem in guest
  --cf-ver VER            cloudflared download version
  --default-origin URL    default origin for start/status
`

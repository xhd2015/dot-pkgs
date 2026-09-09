package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"
)

func (h *helper) run(action string, args []string) error {
	if action == "help" {
		return nil
	}
	origin := h.cfg.DefaultOrigin
	if len(args) > 0 && (action == "status" || action == "start") {
		if args[0] != "" {
			origin = args[0]
		}
	}
	switch action {
	case "status":
		h.printStatus(origin)
		return nil
	case "start":
		return h.doStart(origin)
	case "stop":
		return h.doStop()
	case "logs":
		n := "80"
		if len(args) > 0 && args[0] != "" {
			n = args[0]
		}
		return h.doLogs(n)
	case "purge":
		deep := len(args) > 0 && args[0] == "--deep"
		return h.doPurge(deep)
	case "login":
		return h.doLogin()
	case "list":
		return h.doList()
	case "named-start":
		return h.doNamedStart(args)
	case "named-apply":
		return h.doNamedApply(args)
	case "named-drop-canonical":
		return h.doNamedDropCanonical(args)
	default:
		return fmt.Errorf("unknown action %s", action)
	}
}

func (h *helper) doStart(origin string) error {
	if err := h.requireGuest(origin); err != nil {
		return err
	}
	if err := h.ensureCF(); err != nil {
		return err
	}
	if pid := h.cfPID(); pid != "" {
		fmt.Println("skip=running")
		h.printStatus(origin)
		return nil
	}
	script := fmt.Sprintf(`set -e
mkdir -p /root/.cloudflared /var/log /var/run
: >> %s
nohup %s tunnel --no-autoupdate --url %q >%s 2>&1 &
echo $! > %s
`, h.cfg.CFLog, h.cfg.CFBin, origin, h.cfg.CFLog, h.cfg.CFPid)
	if _, err := h.guestSh(script); err != nil {
		return err
	}
	var url string
	for i := 0; i < 30; i++ {
		url = h.extractURL()
		if url != "" {
			break
		}
		if h.cfPID() == "" && i > 3 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if url != "" {
		_, _ = h.guestSh(fmt.Sprintf("echo %q > %s", url, h.cfg.CFURLFile))
	}
	h.printStatus(origin)
	if h.cfPID() == "" {
		return fmt.Errorf("cloudflared did not start")
	}
	return nil
}

func (h *helper) doStop() error {
	if !h.qemuAlive() || !h.guestOK() {
		fmt.Println("skip=already_stopped")
		return nil
	}
	pid := h.cfPID()
	if pid != "" {
		h.killCF()
		fmt.Printf("stopped pid=%s\n", pid)
	} else {
		fmt.Println("skip=already_stopped")
	}
	_, _ = h.guestCmd("rm", "-f", h.cfg.CFPid)
	return nil
}

func (h *helper) doLogs(n string) error {
	if err := h.requireGuest(h.cfg.DefaultOrigin); err != nil {
		return err
	}
	script := fmt.Sprintf(`if [ -f %s ]; then tail -n %s %s
elif [ -f %s ]; then tail -n %s %s
else echo no cloudflared log
fi`, h.cfg.CFLog, n, h.cfg.CFLog, h.cfg.OldLog, n, h.cfg.OldLog)
	out, err := h.guestSh(script)
	fmt.Print(out)
	return err
}

func (h *helper) doPurge(deep bool) error {
	if h.qemuAlive() && h.guestOK() {
		h.killCF()
		if lp := h.loginPID(); lp != "" {
			_, _ = h.guestCmd("kill", lp)
			_, _ = h.guestCmd("kill", "-9", lp)
		}
		_, _ = h.guestCmd("rm", "-f", h.cfg.CFPid, h.cfg.CFLog, h.cfg.CFLoginLog, h.cfg.CFURLFile, h.cfg.OldLog)
		if deep {
			_, _ = h.guestCmd("rm", "-f", h.cfg.Cert)
			fmt.Println("purged_cert=yes")
		}
	}
	fmt.Println("purged")
	return nil
}

func (h *helper) doLogin() error {
	if err := h.requireGuest(h.cfg.DefaultOrigin); err != nil {
		return err
	}
	if err := h.ensureCF(); err != nil {
		return err
	}
	if h.certOK() {
		fmt.Println("skip=already_logged_in")
		fmt.Println("cert=ok")
		h.printStatus(h.cfg.DefaultOrigin)
		return nil
	}
	if h.loginPID() == "" {
		script := fmt.Sprintf(`set -e
mkdir -p /root/.cloudflared /var/log
: >> %s
nohup %s tunnel login >%s 2>&1 &
`, h.cfg.CFLoginLog, h.cfg.CFBin, h.cfg.CFLoginLog)
		if _, err := h.guestSh(script); err != nil {
			return err
		}
	}
	var loginURL string
	for i := 0; i < 80; i++ {
		if loginURL == "" {
			loginURL = h.extractLoginURL()
			if loginURL != "" {
				fmt.Printf("login_url=%s\n", loginURL)
			}
		}
		if h.certOK() {
			fmt.Println("cert=ok")
			h.printStatus(h.cfg.DefaultOrigin)
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	if loginURL != "" {
		fmt.Printf("login_url=%s\n", loginURL)
	}
	fmt.Println("cert=fail")
	h.printStatus(h.cfg.DefaultOrigin)
	return fmt.Errorf("login timed out")
}

func (h *helper) doList() error {
	if err := h.requireGuest(h.cfg.DefaultOrigin); err != nil {
		return err
	}
	if !h.certOK() {
		fmt.Println("cert=missing")
		return fmt.Errorf("cert missing")
	}
	out, err := h.guestSh(fmt.Sprintf("export TUNNEL_ORIGIN_CERT=%s; %s tunnel list", h.cfg.Cert, h.cfg.CFBin))
	fmt.Print(out)
	return err
}

func (h *helper) doNamedStart(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("named-start requires tunnel name, origin, and at least one hostname")
	}
	name := args[0]
	origin := args[1]
	if origin == "" {
		origin = h.cfg.DefaultOrigin
	}
	hosts := args[2:]
	if name == "" || len(hosts) == 0 {
		return fmt.Errorf("named-start requires tunnel name and at least one hostname")
	}
	if err := h.requireGuest(origin); err != nil {
		return err
	}
	if err := h.ensureCF(); err != nil {
		return err
	}
	if !h.certOK() {
		fmt.Println("cert=missing")
		return fmt.Errorf("cert missing")
	}
	uuid, err := h.ensureTunnel(name)
	if err != nil {
		return err
	}
	fmt.Printf("uuid=%s\n", uuid)
	cfgPath := "/root/.cloudflared/" + name + ".yml"
	if err := h.writeIngress(cfgPath, uuid, hosts, func(string) string { return origin }); err != nil {
		return err
	}
	for _, host := range hosts {
		_, _ = h.guestSh(fmt.Sprintf("export TUNNEL_ORIGIN_CERT=%s; %s tunnel route dns -f %s %s", h.cfg.Cert, h.cfg.CFBin, name, host))
		fmt.Printf("routed=%s\n", host)
	}
	h.pkillCF()
	if err := h.startNamed(cfgPath, hosts[0]); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	h.printStatus(origin)
	if h.cfPID() == "" {
		return fmt.Errorf("named-start: cloudflared did not start")
	}
	fmt.Println("named=yes")
	fmt.Printf("tunnel=%s\n", name)
	return nil
}

func (h *helper) doNamedApply(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("named-apply requires tunnel name and at least one host=origin pair")
	}
	name := args[0]
	pairs := args[1:]
	if name == "" {
		return fmt.Errorf("named-apply requires tunnel name")
	}
	routes := map[string]string{}
	var hosts []string
	for _, pair := range pairs {
		host, origin, ok := strings.Cut(pair, "=")
		if !ok || host == "" || origin == "" {
			return fmt.Errorf("invalid pair %s (want host=origin)", pair)
		}
		hosts = append(hosts, host)
		routes[host] = origin
	}
	if err := h.requireGuest(h.cfg.DefaultOrigin); err != nil {
		return err
	}
	if err := h.ensureCF(); err != nil {
		return err
	}
	if !h.certOK() {
		fmt.Println("cert=missing")
		return fmt.Errorf("cert missing")
	}
	uuid, err := h.ensureTunnel(name)
	if err != nil {
		return err
	}
	fmt.Printf("uuid=%s\n", uuid)
	cfgPath := "/root/.cloudflared/" + name + ".yml"
	for _, host := range hosts {
		_, _ = h.guestSh(fmt.Sprintf("export TUNNEL_ORIGIN_CERT=%s; %s tunnel route dns -f %s %s", h.cfg.Cert, h.cfg.CFBin, name, host))
		fmt.Printf("routed=%s\n", host)
		fmt.Printf("origin_%s=%s\n", host, routes[host])
	}
	if err := h.writeIngress(cfgPath, uuid, hosts, func(host string) string { return routes[host] }); err != nil {
		return err
	}
	h.pkillCF()
	if err := h.startNamed(cfgPath, hosts[0]); err != nil {
		return err
	}
	var pid string
	for i := 0; i < 15; i++ {
		pid = h.cfPID()
		if pid != "" {
			break
		}
		time.Sleep(time.Second)
	}
	h.printStatus(h.cfg.DefaultOrigin)
	if pid == "" {
		out, _ := h.guestSh(fmt.Sprintf("tail -n 40 %s", h.cfg.CFLog))
		fmt.Fprint(os.Stderr, out)
		return fmt.Errorf("named-apply: cloudflared did not start")
	}
	fmt.Println("named=yes")
	fmt.Printf("tunnel=%s\n", name)
	fmt.Println("applied=yes")
	return nil
}

func (h *helper) doNamedDropCanonical(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("named-drop-canonical requires name origin drop keep")
	}
	name, origin, drop, keep := args[0], args[1], args[2], args[3]
	if name == "" || keep == "" {
		return fmt.Errorf("named-drop-canonical requires name and keep hostname")
	}
	if err := h.requireGuest(origin); err != nil {
		return err
	}
	uuid := h.findTunnelUUID(name)
	if uuid == "" {
		return fmt.Errorf("tunnel %s not found", name)
	}
	cfgPath := "/root/.cloudflared/" + name + ".yml"
	if err := h.writeIngress(cfgPath, uuid, []string{keep}, func(string) string { return origin }); err != nil {
		return err
	}
	fmt.Printf("dropped=%s\n", drop)
	fmt.Printf("kept=%s\n", keep)
	h.pkillCF()
	if err := h.startNamed(cfgPath, keep); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	h.printStatus(origin)
	fmt.Println("named=yes")
	return nil
}

func (h *helper) ensureTunnel(name string) (string, error) {
	uuid := h.findTunnelUUID(name)
	if uuid != "" {
		return uuid, nil
	}
	_, err := h.guestSh(fmt.Sprintf("export TUNNEL_ORIGIN_CERT=%s; %s tunnel create %s", h.cfg.Cert, h.cfg.CFBin, name))
	if err != nil {
		return "", fmt.Errorf("tunnel create: %w", err)
	}
	uuid = h.findTunnelUUID(name)
	if uuid == "" {
		return "", fmt.Errorf("tunnel %s created but uuid not found", name)
	}
	return uuid, nil
}

func (h *helper) writeIngress(cfgPath, uuid string, hosts []string, originFor func(string) string) error {
	var b strings.Builder
	b.WriteString("tunnel: " + uuid + "\n")
	b.WriteString("credentials-file: /root/.cloudflared/" + uuid + ".json\n")
	b.WriteString("protocol: http2\n")
	b.WriteString("ingress:\n")
	for _, host := range hosts {
		b.WriteString("  - hostname: " + host + "\n")
		b.WriteString("    service: " + originFor(host) + "\n")
	}
	b.WriteString("  - service: http_status:404\n")
	// Base64 avoids Go %q turning real newlines into literal \n inside bash.
	enc := base64.StdEncoding.EncodeToString([]byte(b.String()))
	script := fmt.Sprintf("echo %s | base64 -d > %s", enc, cfgPath)
	_, err := h.guestSh(script)
	return err
}

func (h *helper) startNamed(cfgPath, firstHost string) error {
	script := fmt.Sprintf(`set -e
mkdir -p /root/.cloudflared /var/log /var/run
: >> %s
nohup %s --no-autoupdate tunnel --config %s run >>%s 2>&1 &
echo $! > %s
echo https://%s > %s
`, h.cfg.CFLog, h.cfg.CFBin, cfgPath, h.cfg.CFLog, h.cfg.CFPid, firstHost, h.cfg.CFURLFile)
	_, err := h.guestSh(script)
	return err
}

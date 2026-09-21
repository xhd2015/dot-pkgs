package singboxtun

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const BrewInstallSingBoxCmd = "brew install sing-box"

const (
	defaultCacheDirName  = "singboxtun"
	defaultSudoersName   = "singboxtun"
	defaultDNSHijackHint = "Retry with --dns-hijack"
)

func PrintCommand(cmd string) {
	fmt.Printf("$ %s\n", cmd)
}

func SingBoxRunCommand(sudo bool, configPath string) string {
	if sudo {
		return fmt.Sprintf("sudo sing-box run -c %s", shellQuote(configPath))
	}
	return fmt.Sprintf("sing-box run -c %s", shellQuote(configPath))
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '/' && r != '_' && r != '-' && r != '.' && r != ':' {
			return fmt.Sprintf("%q", s)
		}
	}
	return s
}

type RunTunOptions struct {
	LocalSocksPort int
	ProxyHost      string
	ConfigFile     string
	Yes            bool
	NoInstall      bool
	NoSetupSudo    bool
	Detach         bool
	HttpOnly       bool
	DNSHijack      bool
	Policy         *DomainPolicy
	AlsoProxy      []AlsoProxyPattern
	CacheDirName   string // default "singboxtun"
	SudoersName    string // default "singboxtun"
	DNSHijackHint  string
	// NoBindInterface skips pinning direct outbound to the default-route NIC.
	// Needed for Wi‑Fi→hotspot: a frozen bind_interface (e.g. en0) blackholes
	// cloudflared + direct fallback after the uplink moves.
	NoBindInterface bool
	// NoDirectFallback keeps http-only web selector on proxy even when SOCKS
	// flaps. On polluted hotspots, direct re-resolves to wrong IPs (e.g. google
	// → 104.244/31.13) and times out; failing on proxy is better than that.
	NoDirectFallback bool
	// ExtraExcludeHosts are resolved to /32 route_exclude entries (in addition to
	// ProxyHost + built-in argotunnel edge names).
	ExtraExcludeHosts []string
	// ExcludeRefreshInterval re-resolves exclude hosts periodically. Zero means
	// DefaultExcludeRefreshInterval (1h). Negative disables refresh.
	ExcludeRefreshInterval time.Duration
}

// RunHttpOnlyOptions is deprecated; use RunTunOptions with HttpOnly set.
type RunHttpOnlyOptions = RunTunOptions

type BuildConfigOptions struct {
	BindInterface     string
	SkipBindInterface bool // if true, never auto-fill BindInterface from default route
	LocalSocksPort    int  // required > 0; SOCKS outbound to local proxy
	ProxyHost         string
	ExtraExcludeHosts []string // optional; merged with ProxyHost + argotunnel defaults
	HttpOnly          bool
	Policy            *DomainPolicy
	AlsoProxy         []AlsoProxyPattern // http-only TCP exceptions via proxy
	DNSHijack         bool               // http-only: optional; full VPN always hijacks DNS
	InitialUseProxy   bool               // http-only selector default when upstream is up
}

type TestHooks struct {
	LookPath        func(name string) (string, error)
	IsTTY           func() bool
	Confirm         func(prompt string) bool
	BrewInstall     func() error
	Geteuid         func() int
	RunSingBox      func(ctx context.Context, sudo bool, configPath string) error
	StartDetached   func(configPath, logPath string, useSudo bool) (pid int, err error)
	EnsureSudoSetup func(singBoxPath string, noSetup bool, cacheDirName, sudoersName string) error
	UserCacheDir    func() (string, error)
}

var currentHooks = TestHooks{
	LookPath: defaultLookPath,
	IsTTY:    defaultIsTTY,
	Confirm: func(prompt string) bool {
		fmt.Print(prompt)
		var s string
		fmt.Scanln(&s)
		return s == "y" || s == "Y" || s == "yes"
	},
	BrewInstall: func() error {
		cmd := exec.Command("brew", "install", "sing-box")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	},
	Geteuid:    os.Geteuid,
	RunSingBox: runSingBoxForeground,
	StartDetached: func(configPath, logPath string, useSudo bool) (int, error) {
		PrintCommand(SingBoxRunCommand(useSudo, configPath))
		args := []string{"sing-box", "run", "-c", configPath}
		if useSudo {
			args = append([]string{"sudo"}, args...)
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = singBoxProcessEnv()
		logFile, err := os.Create(logPath)
		if err != nil {
			return 0, fmt.Errorf("create log file: %w", err)
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		if err := cmd.Start(); err != nil {
			logFile.Close()
			return 0, err
		}
		return cmd.Process.Pid, nil
	},
	UserCacheDir: func() (string, error) {
		return os.UserCacheDir()
	},
	EnsureSudoSetup: defaultEnsureSudoSetup,
}

func defaultLookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func defaultIsTTY() bool {
	fi, _ := os.Stdout.Stat()
	return fi != nil && fi.Mode()&os.ModeCharDevice != 0
}

func InstallTestHooks(h TestHooks) func() {
	old := currentHooks
	currentHooks = h
	return func() {
		currentHooks = old
	}
}

func resolveCacheDirName(name string) string {
	if name == "" {
		return defaultCacheDirName
	}
	return name
}

func resolveSudoersName(name string) string {
	if name == "" {
		return defaultSudoersName
	}
	return name
}

func resolveDNSHijackHint(hint string) string {
	if hint == "" {
		return defaultDNSHijackHint
	}
	return hint
}

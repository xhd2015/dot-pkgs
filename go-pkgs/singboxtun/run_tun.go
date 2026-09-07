package singboxtun

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type tunRunBundle struct {
	configPath    string
	cleanupConfig func()
	stopHealth    func()
	buildOpts     *BuildConfigOptions
	excludeHosts  []string
	excludeCIDRs  []string // last resolved dynamic /32s (sorted)
}

// RunTun starts sing-box TUN (full VPN or --http-only mode) over a local SOCKS proxy.
func RunTun(opts RunTunOptions) error {
	cacheDirName := resolveCacheDirName(opts.CacheDirName)
	sudoersName := resolveSudoersName(opts.SudoersName)
	dnsHint := resolveDNSHijackHint(opts.DNSHijackHint)

	prevCache := activeCacheDirName
	activeCacheDirName = cacheDirName
	defer func() { activeCacheDirName = prevCache }()

	if opts.HttpOnly && !opts.DNSHijack {
		SetSkipPlatformTunDNS(true)
		defer SetSkipPlatformTunDNS(false)
	}

	if err := restoreStuckTunDNS(); err != nil {
		return err
	}

	if opts.ConfigFile == "" && opts.HttpOnly {
		MaybeWarnDNSPollution(opts.DNSHijack, dnsHint)
	}

	bundle, err := prepareTunRun(opts)
	if err != nil {
		return err
	}

	singBoxPath, singBoxErr := currentHooks.LookPath("sing-box")
	if singBoxErr != nil {
		if opts.NoInstall {
			return fmt.Errorf("sing-box not installed (--no-install set)")
		}
		if !currentHooks.IsTTY() {
			return fmt.Errorf("sing-box not installed; install it with: %s", BrewInstallSingBoxCmd)
		}
		fmt.Println("sing-box is not installed.")
		PrintCommand(BrewInstallSingBoxCmd)
		shouldInstall := opts.Yes
		if !shouldInstall {
			confirmed := currentHooks.Confirm("Install via Homebrew? [y/N] ")
			if !confirmed {
				return fmt.Errorf("sing-box install declined")
			}
		}
		if err := currentHooks.BrewInstall(); err != nil {
			return fmt.Errorf("brew install sing-box failed: %w", err)
		}
		singBoxPath, singBoxErr = currentHooks.LookPath("sing-box")
	}

	euid := currentHooks.Geteuid()
	needSudo := euid != 0

	if needSudo && !opts.NoSetupSudo && singBoxErr == nil {
		if err := currentHooks.EnsureSudoSetup(singBoxPath, opts.NoSetupSudo, cacheDirName, sudoersName); err != nil {
			return err
		}
	}

	if opts.Detach {
		if bundle.cleanupConfig != nil {
			defer bundle.cleanupConfig()
		}
		return runDetach(bundle.configPath, needSudo, opts.HttpOnly, cacheDirName)
	}

	if bundle.cleanupConfig != nil {
		defer bundle.cleanupConfig()
	}
	if bundle.stopHealth != nil {
		defer bundle.stopHealth()
	}

	if needSudo && !currentHooks.IsTTY() {
		return fmt.Errorf("sing-box needs root privileges; run with sudo or from a TTY")
	}

	if hasProxyEnv() {
		fmt.Println("Note: HTTP/SOCKS proxy env vars are cleared for sing-box (upstream SOCKS must be reached directly).")
	}
	if systemProxyEnabled() {
		fmt.Println("Note: macOS system HTTP/HTTPS/SOCKS proxy will be disabled while the TUN is up.")
	}

	if opts.ConfigFile == "" && opts.HttpOnly && opts.LocalSocksPort > 0 {
		stopHealth := StartWebOutboundHealthMonitorOpts(opts.LocalSocksPort, WebHealthOptions{
			NoDirectFallback: opts.NoDirectFallback,
		})
		bundle.stopHealth = stopHealth
		defer stopHealth()
	}

	return runSingBoxWithExcludeRefresh(opts, bundle, needSudo)
}

// RunHttpOnly is deprecated; use RunTun with HttpOnly set.
func RunHttpOnly(opts RunHttpOnlyOptions) error {
	opts.HttpOnly = true
	return RunTun(opts)
}

func prepareTunRun(opts RunTunOptions) (*tunRunBundle, error) {
	if opts.ConfigFile != "" {
		return &tunRunBundle{configPath: opts.ConfigFile}, nil
	}

	if opts.LocalSocksPort <= 0 {
		return nil, fmt.Errorf("LocalSocksPort required unless ConfigFile is set")
	}

	hosts := tunnelExcludeHosts(opts.ProxyHost, opts.ExtraExcludeHosts)
	dyn := resolveExcludeHostCIDRs(hosts)
	if opts.ProxyHost != "" && len(resolveHostIPv4CIDRs(opts.ProxyHost)) == 0 {
		fmt.Fprintf(os.Stderr, "warning: could not resolve %s for TUN route exclusions\n", opts.ProxyHost)
	}
	if len(dyn) == 0 {
		fmt.Fprintf(os.Stderr, "warning: no /32s resolved for tunnel exclude hosts %v\n", hosts)
	} else {
		fmt.Printf("TUN route_exclude: %d /32s from %d hosts\n", len(dyn), len(hosts))
	}

	buildOpts := buildTunConfigOptions(opts.LocalSocksPort, opts.NoBindInterface)
	buildOpts.ProxyHost = opts.ProxyHost
	buildOpts.ExtraExcludeHosts = opts.ExtraExcludeHosts
	buildOpts.HttpOnly = opts.HttpOnly
	buildOpts.Policy = opts.Policy
	buildOpts.AlsoProxy = opts.AlsoProxy
	buildOpts.DNSHijack = opts.DNSHijack
	if opts.HttpOnly {
		if opts.NoDirectFallback {
			buildOpts.InitialUseProxy = true
			if !ProbeUpstreamProxy(opts.LocalSocksPort) {
				fmt.Println("Upstream SOCKS unreachable; starting with web→proxy anyway (no direct fallback).")
			}
		} else {
			buildOpts.InitialUseProxy = ProbeUpstreamProxy(opts.LocalSocksPort)
			if !buildOpts.InitialUseProxy {
				fmt.Println("Upstream SOCKS unreachable; starting in direct-fallback mode.")
			}
		}
	}
	if len(opts.AlsoProxy) > 0 && !opts.DNSHijack {
		fmt.Fprintf(os.Stderr, "warning: --also-proxy without --dns-hijack may miss hosts that resolve to TUN-excluded private IPs (e.g. 10.0.0.0/8)\n")
	}

	if opts.HttpOnly {
		fmt.Println("Building sing-box HTTP-only TUN config...")
	} else {
		fmt.Println("Building sing-box TUN config...")
	}
	data, err := BuildTunConfig(buildOpts)
	if err != nil {
		return nil, err
	}

	prefix := "singbox-"
	if opts.HttpOnly {
		prefix = "singbox-http-only-"
	}
	tmpFile, err := os.CreateTemp("", prefix+"*.json")
	if err != nil {
		return nil, fmt.Errorf("create temp config: %w", err)
	}
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("write temp config: %w", err)
	}
	tmpFile.Close()

	configPath := tmpFile.Name()
	return &tunRunBundle{
		configPath:   configPath,
		buildOpts:    buildOpts,
		excludeHosts: hosts,
		excludeCIDRs: dyn,
		cleanupConfig: func() {
			_ = os.Remove(configPath)
		},
	}, nil
}

func runSingBoxWithExcludeRefresh(opts RunTunOptions, bundle *tunRunBundle, needSudo bool) error {
	parentCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	interval := opts.ExcludeRefreshInterval
	if interval == 0 {
		interval = DefaultExcludeRefreshInterval
	}

	restartCh := make(chan struct{}, 1)
	if interval > 0 && bundle.buildOpts != nil && opts.ConfigFile == "" {
		go watchExcludeRefresh(parentCtx, bundle, interval, restartCh)
	}

	for {
		runCtx, runCancel := context.WithCancel(parentCtx)
		errCh := make(chan error, 1)
		go func() {
			errCh <- currentHooks.RunSingBox(runCtx, needSudo, bundle.configPath)
		}()

		select {
		case <-parentCtx.Done():
			runCancel()
			<-errCh
			return nil
		case <-restartCh:
			fmt.Println("route_exclude /32 set changed; restarting sing-box...")
			runCancel()
			<-errCh
		case err := <-errCh:
			runCancel()
			// Restart may have cancelled the process; prefer planned refresh.
			select {
			case <-restartCh:
				fmt.Println("route_exclude /32 set changed; restarting sing-box...")
			default:
				if parentCtx.Err() != nil {
					return nil
				}
				return err
			}
		}
		select {
		case <-parentCtx.Done():
			return nil
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func watchExcludeRefresh(ctx context.Context, bundle *tunRunBundle, interval time.Duration, restartCh chan<- struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if bundle.buildOpts == nil {
				continue
			}
			next := resolveExcludeHostCIDRs(bundle.excludeHosts)
			if excludeCIDRsEqual(bundle.excludeCIDRs, next) {
				continue
			}
			fmt.Fprintf(os.Stderr, "warning: tunnel exclude /32s changed (%d → %d); rewriting config\n",
				len(bundle.excludeCIDRs), len(next))
			data, err := BuildTunConfig(bundle.buildOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: rebuild sing-box config: %v\n", err)
				continue
			}
			if err := os.WriteFile(bundle.configPath, data, 0o600); err != nil {
				fmt.Fprintf(os.Stderr, "warning: write sing-box config: %v\n", err)
				continue
			}
			bundle.excludeCIDRs = next
			select {
			case restartCh <- struct{}{}:
			default:
			}
		}
	}
}

func runDetach(configPath string, needSudo bool, httpOnly bool, cacheDirName string) error {
	cacheDir, err := currentHooks.UserCacheDir()
	if err != nil {
		return fmt.Errorf("cache dir: %w", err)
	}
	singBoxDir := filepath.Join(cacheDir, cacheDirName)
	if err := os.MkdirAll(singBoxDir, 0700); err != nil {
		return fmt.Errorf("create %s dir: %w", cacheDirName, err)
	}

	runConfigPath := filepath.Join(singBoxDir, "run.json")
	logPath := filepath.Join(singBoxDir, "sing-box.log")
	if httpOnly {
		runConfigPath = filepath.Join(singBoxDir, "http-only-run.json")
		logPath = filepath.Join(singBoxDir, "sing-box-http-only.log")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if err := os.WriteFile(runConfigPath, data, 0600); err != nil {
		return fmt.Errorf("write run config: %w", err)
	}

	pid, err := currentHooks.StartDetached(runConfigPath, logPath, needSudo)
	if err != nil {
		return fmt.Errorf("start detached: %w", err)
	}

	modeLabel := "VPN"
	if httpOnly {
		modeLabel = "HTTP-only"
	}
	fmt.Printf("sing-box %s started in background (PID: %d)\n", modeLabel, pid)
	fmt.Printf("Config: %s\n", runConfigPath)
	fmt.Printf("Log:    %s\n", logPath)
	if httpOnly {
		fmt.Println("Note: upstream health monitoring runs in foreground mode only; restart with --http-only if the SOCKS upstream flaps while detached.")
	}
	fmt.Println("Note: exclude-host /32 refresh (1h) runs in foreground mode only.")
	return nil
}

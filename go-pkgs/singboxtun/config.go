package singboxtun

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

const (
	remoteDNSAddr    = "8.8.8.8"
	bootstrapDNSAddr = "8.8.8.8"
	fakeIPRange      = "198.18.0.0/15"
	tunDNSAddress    = "172.19.0.2" // client DNS on 172.19.0.1/30 TUN subnet
)

var lanBypassCIDRs = []string{
	"127.0.0.0/8",
	"192.168.0.0/16",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"17.0.0.0/8",
}

// tunExcludedCIDRs bypasses the TUN at the OS level. Keep 172.16.0.0/12 out of
// this list so hotspot gateways like 172.20.10.1 still enter the TUN and DNS
// hijack can intercept port-53 queries instead of using polluted router DNS.
var tunExcludedCIDRs = []string{
	"192.168.0.0/16",
	"10.0.0.0/8",
}

// defaultTunnelExcludeHosts are resolved to /32s for route_exclude_address so
// cloudflared can reach Cloudflare tunnel edge without entering the TUN.
// Docs: region{1,2}.v2.argotunnel.com (and us-region*) on port 7844.
var defaultTunnelExcludeHosts = []string{
	"region1.v2.argotunnel.com",
	"region2.v2.argotunnel.com",
	"us-region1.v2.argotunnel.com",
	"us-region2.v2.argotunnel.com",
}

// Public DNS resolvers used by the bootstrap server must bypass the TUN at the
// OS level so the local SOCKS upstream can resolve hosts before the tunnel is up.
var bootstrapExcludeCIDRs = []string{
	"8.8.8.8/32",
	"1.1.1.1/32",
}

// DefaultExcludeRefreshInterval is how often we re-resolve tunnel exclude hosts.
const DefaultExcludeRefreshInterval = time.Hour

const hostLookupTimeout = 3 * time.Second

var bootstrapResolvers = []string{"1.1.1.1:53", "8.8.8.8:53"}

// lookupHostIPv4 is overridden in tests.
var lookupHostIPv4 = defaultLookupHostIPv4

func defaultLookupHostIPv4(host string) []net.IP {
	ctx, cancel := context.WithTimeout(context.Background(), hostLookupTimeout)
	defer cancel()

	ips, _ := lookupHostIPv4WithResolver(ctx, net.DefaultResolver, host)
	if len(ips) > 0 {
		return ips
	}
	for _, resolver := range bootstrapResolvers {
		ips, _ = lookupHostIPv4ViaUDP(ctx, resolver, host)
		if len(ips) > 0 {
			return ips
		}
	}
	return nil
}

func lookupHostIPv4WithResolver(ctx context.Context, resolver *net.Resolver, host string) ([]net.IP, error) {
	ips, err := resolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return nil, err
	}
	return ips, nil
}

func lookupHostIPv4ViaUDP(ctx context.Context, resolverAddr, host string) ([]net.IP, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "udp", resolverAddr)
		},
	}
	return lookupHostIPv4WithResolver(ctx, r, host)
}

func resolveHostIPv4CIDRs(host string) []string {
	ips := lookupHostIPv4(host)
	if len(ips) == 0 {
		return nil
	}
	var cidrs []string
	seen := make(map[string]struct{})
	for _, ip := range ips {
		ip4 := ip.To4()
		if ip4 == nil {
			continue
		}
		cidr := ip4.String() + "/32"
		if _, ok := seen[cidr]; ok {
			continue
		}
		seen[cidr] = struct{}{}
		cidrs = append(cidrs, cidr)
	}
	return cidrs
}

// tunnelExcludeHosts returns ProxyHost + built-in argotunnel edge names (+ extras).
func tunnelExcludeHosts(proxyHost string, extra []string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(h string) {
		h = strings.TrimSpace(h)
		if h == "" {
			return
		}
		if _, ok := seen[h]; ok {
			return
		}
		seen[h] = struct{}{}
		out = append(out, h)
	}
	add(proxyHost)
	for _, h := range defaultTunnelExcludeHosts {
		add(h)
	}
	for _, h := range extra {
		add(h)
	}
	return out
}

// resolveExcludeHostCIDRs resolves hosts via bootstrap DNS to unique sorted /32s.
func resolveExcludeHostCIDRs(hosts []string) []string {
	seen := make(map[string]struct{})
	var cidrs []string
	for _, h := range hosts {
		for _, c := range resolveHostIPv4CIDRs(h) {
			if _, ok := seen[c]; ok {
				continue
			}
			seen[c] = struct{}{}
			cidrs = append(cidrs, c)
		}
	}
	sort.Strings(cidrs)
	return cidrs
}

func tunRouteExcludeAddresses(proxyHost string, extraHosts []string) []string {
	hosts := tunnelExcludeHosts(proxyHost, extraHosts)
	dyn := resolveExcludeHostCIDRs(hosts)
	exclude := make([]string, 0, len(tunExcludedCIDRs)+len(bootstrapExcludeCIDRs)+len(dyn))
	exclude = append(exclude, tunExcludedCIDRs...)
	exclude = append(exclude, bootstrapExcludeCIDRs...)
	exclude = append(exclude, dyn...)
	return exclude
}

func excludeCIDRsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func buildTunRouteConfig(routeRules []map[string]any, localSocksPort int) map[string]any {
	resolver := "remote"
	if localSocksPort > 0 {
		// Required by sing-box 1.12+; user DNS still uses fakeip via DNS rules.
		resolver = "bootstrap"
	}
	return map[string]any{
		"rules":                   routeRules,
		"default_domain_resolver": resolver,
		"auto_detect_interface":   true,
		"final":                   "proxy",
	}
}

func remoteDNSDetour(localSocksPort int) string {
	if localSocksPort > 0 {
		return "proxy"
	}
	return "direct"
}

func buildTunDNSConfig(proxyHost string, localSocksPort int, skipBindInterface bool) map[string]any {
	useSocksDNS := localSocksPort > 0
	bootstrap := map[string]any{
		"type":        "udp",
		"tag":         "bootstrap",
		"server":      bootstrapDNSAddr,
		"server_port": 53,
	}
	// sing-box 1.13 rejects DNS detour→direct when the direct outbound has no
	// dialer options ("empty direct"). With SkipBindInterface we rely on OS
	// route_exclude for 8.8.8.8/32 instead of detouring through direct.
	if !skipBindInterface {
		bootstrap["detour"] = "direct"
	}
	servers := []map[string]any{
		{
			"type": "local",
			"tag":  "local",
		},
		bootstrap,
		{
			"type":        "fakeip",
			"tag":         "fakeip",
			"inet4_range": fakeIPRange,
		},
	}

	// PTR reject + FakeIP for ordinary A/AAAA. ProxyHost MUST be prepended
	// (bootstrap/direct) so the reverse-tunnel / CF hostname is never FakeIP'd —
	// otherwise hub probes and cloudflared hairpin into the TUN (198.18.0.0/15).
	rules := []map[string]any{
		{
			"query_type": []string{"PTR"},
			"domain_suffix": []string{
				".in-addr.arpa",
				".ip6.arpa",
			},
			"action": "reject",
		},
	}
	rules = append(rules, proxyHostBootstrapDNSRules(proxyHost)...)
	rules = append(rules,
		map[string]any{
			"query_type": []string{"A", "AAAA"},
			"action":     "route",
			"server":     "fakeip",
		},
	)
	if useSocksDNS {
		// SOCKS upstream resolves hostnames; fakeip answers instantly for the rest.
		rules = append(rules, map[string]any{
			"action": "route",
			"server": "fakeip",
		})
	} else {
		servers = append(servers, remoteDNSServer(localSocksPort))
		rules = append(rules, map[string]any{
			"action": "route",
			"server": "remote",
		})
	}
	return map[string]any{
		"servers":  servers,
		"rules":    rules,
		"strategy": "ipv4_only",
	}
}

// proxyHostBootstrapDNSRules forces real (non-FakeIP) answers for the transport
// hostname and common Cloudflare edge names used by the reverse tunnel.
func proxyHostBootstrapDNSRules(proxyHost string) []map[string]any {
	var domains []string
	if proxyHost != "" {
		domains = append(domains, proxyHost)
	}
	// Keep CF connector / edge resolution off FakeIP even when apps query them.
	domains = append(domains,
		"cloudflare.com",
		"cloudflare.net",
		"argotunnel.com",
	)
	seen := make(map[string]struct{}, len(domains))
	var rules []map[string]any
	for _, d := range domains {
		if d == "" {
			continue
		}
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		rule := map[string]any{
			"action": "route",
			"server": "bootstrap",
		}
		if strings.Contains(d, ".") && !strings.HasPrefix(d, "*.") && d != proxyHost {
			// Suffix match for well-known CF zones; exact match for ProxyHost.
			rule["domain_suffix"] = []string{d}
		} else {
			rule["domain"] = []string{d}
		}
		rules = append(rules, rule)
	}
	return rules
}

func remoteDNSServer(localSocksPort int) map[string]any {
	server := map[string]any{
		"type":        "udp",
		"tag":         "remote",
		"server":      remoteDNSAddr,
		"server_port": 53,
		"detour":      remoteDNSDetour(localSocksPort),
	}
	return server
}

// BuildTunConfig renders sing-box JSON for full VPN or http-only TUN over local SOCKS.
func BuildTunConfig(opts *BuildConfigOptions) ([]byte, error) {
	if opts == nil {
		return nil, fmt.Errorf("build options required")
	}
	if opts.LocalSocksPort <= 0 {
		return nil, fmt.Errorf("LocalSocksPort required")
	}
	if opts.HttpOnly {
		return buildSingBoxHttpOnlyTunConfig(opts)
	}
	return buildSingBoxFullTunConfig(opts)
}

func buildSingBoxHttpOnlyTunConfig(opts *BuildConfigOptions) ([]byte, error) {
	if opts == nil {
		opts = &BuildConfigOptions{HttpOnly: true}
	}
	localSocksPort := opts.LocalSocksPort
	if localSocksPort <= 0 {
		return nil, fmt.Errorf("http-only config requires local SOCKS port")
	}

	defaultOutbound := directOutboundTag
	if opts.InitialUseProxy {
		defaultOutbound = proxyOutboundTag
	}

	routeRules := []map[string]any{{"action": "sniff"}}
	if opts.DNSHijack {
		routeRules = appendHttpOnlyDNSRouteRules(routeRules)
	}
	routeRules = appendBuiltinBypassRules(routeRules, opts.ProxyHost)
	routeRules = appendPolicyRouteRules(routeRules, opts.Policy, webSelectorTag, true)
	// After HTTP catch-all: non-HTTP ports (e.g. :22, :6606) still match here.
	routeRules = appendAlsoProxyRouteRules(routeRules, opts.AlsoProxy, webSelectorTag)

	if !opts.SkipBindInterface {
		if bindIface := defaultOutboundBindInterface(); bindIface != "" && opts.BindInterface == "" {
			opts.BindInterface = bindIface
		}
	}
	proxyOutbound, directOutbound := buildSocksProxyOutbounds(localSocksPort, opts.BindInterface)

	dnsCfg := buildHttpOnlyDNSConfigLocal()
	strictRoute := false
	if opts.DNSHijack {
		dnsCfg = buildTunDNSConfig(opts.ProxyHost, localSocksPort, opts.SkipBindInterface)
		strictRoute = true
	}

	routeCfg := map[string]any{
		"rules":                 routeRules,
		"auto_detect_interface": true,
		"final":                 directOutboundTag,
	}
	if opts.DNSHijack {
		routeCfg["default_domain_resolver"] = "bootstrap"
	}

	cfg := map[string]any{
		"log": map[string]any{
			"level":  "info",
			"output": singBoxLogPath(),
		},
		"dns": dnsCfg,
		"experimental": map[string]any{
			"clash_api": map[string]any{
				"external_controller": clashAPIListen,
			},
		},
		"inbounds": []map[string]any{{
			"type": "tun", "tag": "tun-in",
			"address": []string{"172.19.0.1/30"}, "mtu": 1280,
			"auto_route": true, "strict_route": strictRoute, "stack": "system",
			"route_exclude_address": tunRouteExcludeAddresses(opts.ProxyHost, opts.ExtraExcludeHosts),
		}},
		"outbounds": []map[string]any{
			{
				"type": "selector", "tag": webSelectorTag,
				"outbounds": []string{proxyOutboundTag, directOutboundTag},
				"default":   defaultOutbound,
			},
			proxyOutbound,
			directOutbound,
		},
		"route": routeCfg,
	}
	return marshalSingBoxConfig(cfg)
}

func buildSingBoxFullTunConfig(opts *BuildConfigOptions) ([]byte, error) {
	if opts == nil {
		return nil, fmt.Errorf("build options required")
	}
	localSocksPort := opts.LocalSocksPort
	if localSocksPort <= 0 {
		return nil, fmt.Errorf("full TUN config requires local SOCKS port")
	}

	bindIface := opts.BindInterface
	if bindIface == "" && !opts.SkipBindInterface {
		bindIface = defaultOutboundBindInterface()
	}
	policy := opts.Policy
	proxyHost := opts.ProxyHost

	routeRules := []map[string]any{{"action": "sniff"}}
	routeRules = append(routeRules, map[string]any{
		"type": "logical", "mode": "or",
		"rules":  []map[string]any{{"protocol": "dns"}, {"port": 53}},
		"action": "hijack-dns",
	})
	routeRules = appendBuiltinBypassRules(routeRules, proxyHost)
	routeRules = appendPolicyRouteRules(routeRules, policy, proxyOutboundTag, false)

	finalOutbound := proxyOutboundTag
	if policy != nil && policy.Mode == PolicyWhitelist {
		finalOutbound = directOutboundTag
	}

	proxyOutbound, directOutbound := buildSocksProxyOutbounds(localSocksPort, bindIface)

	routeCfg := buildTunRouteConfig(routeRules, localSocksPort)
	routeCfg["final"] = finalOutbound

	cfg := map[string]any{
		"log": map[string]any{"level": "info", "output": singBoxLogPath()},
		"dns": buildTunDNSConfig(proxyHost, localSocksPort, opts.SkipBindInterface),
		"inbounds": []map[string]any{{
			"type": "tun", "tag": "tun-in",
			"address": []string{"172.19.0.1/30"}, "mtu": 1280,
			"auto_route": true, "strict_route": true, "stack": "system",
			"route_exclude_address": tunRouteExcludeAddresses(proxyHost, opts.ExtraExcludeHosts),
		}},
		"outbounds": []map[string]any{proxyOutbound, directOutbound},
		"route":     routeCfg,
	}
	return marshalSingBoxConfig(cfg)
}

func buildSocksProxyOutbounds(localSocksPort int, bindIface string) (map[string]any, map[string]any) {
	proxyOutbound := map[string]any{
		"type": "socks", "tag": proxyOutboundTag,
		"server": "127.0.0.1", "server_port": localSocksPort, "version": "5",
		"udp_over_tcp": map[string]any{"enabled": true, "version": 2},
	}
	directOutbound := map[string]any{"type": "direct", "tag": directOutboundTag}
	if bindIface != "" {
		directOutbound["bind_interface"] = bindIface
	}
	return proxyOutbound, directOutbound
}

func marshalSingBoxConfig(cfg map[string]any) ([]byte, error) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal sing-box config: %w", err)
	}
	return data, nil
}

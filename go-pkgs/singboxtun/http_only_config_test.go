package singboxtun

import (
	"encoding/json"
	"testing"
)

func TestBuildSingBoxHttpOnlyTunConfigDefaults(t *testing.T) {
	data, err := BuildSingBoxHttpOnlyTunConfig(&HttpOnlyConfigOptions{
		LocalSocksPort:  11080,
		ProxyHost:       "proxy.example.com",
		InitialUseProxy: true,
	})
	if err != nil {
		t.Fatalf("BuildSingBoxHttpOnlyTunConfig: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	route, _ := cfg["route"].(map[string]any)
	if route["final"] != directOutboundTag {
		t.Fatalf("final = %v, want %s", route["final"], directOutboundTag)
	}
	if _, ok := route["default_domain_resolver"]; ok {
		t.Fatalf("default_domain_resolver should be omitted without --dns-hijack")
	}
	rules := routeRulesFromCfg(cfg)
	if rulesContainHijackDNS(rules) {
		t.Fatalf("hijack-dns should be off by default: %v", rules)
	}
	dns, _ := cfg["dns"].(map[string]any)
	servers, _ := dns["servers"].([]any)
	if len(servers) != 1 {
		t.Fatalf("dns servers = %v, want local only", servers)
	}
}

func TestBuildSingBoxHttpOnlyTunConfigDNSHijack(t *testing.T) {
	data, err := BuildSingBoxHttpOnlyTunConfig(&HttpOnlyConfigOptions{
		LocalSocksPort:  11080,
		ProxyHost:       "proxy.example.com",
		InitialUseProxy: true,
		DNSHijack:       true,
	})
	if err != nil {
		t.Fatalf("BuildSingBoxHttpOnlyTunConfig: %v", err)
	}
	var cfg map[string]any
	_ = json.Unmarshal(data, &cfg)
	rules := routeRulesFromCfg(cfg)
	if !rulesContainHijackDNS(rules) {
		t.Fatalf("missing hijack-dns rule: %v", rules)
	}
	route, _ := cfg["route"].(map[string]any)
	if route["default_domain_resolver"] != "bootstrap" {
		t.Fatalf("default_domain_resolver = %v, want bootstrap", route["default_domain_resolver"])
	}
	dns, _ := cfg["dns"].(map[string]any)
	servers, _ := dns["servers"].([]any)
	foundFakeIP := false
	for _, s := range servers {
		server, _ := s.(map[string]any)
		if server["type"] == "fakeip" {
			foundFakeIP = true
		}
	}
	if !foundFakeIP {
		t.Fatalf("dns servers = %v, want fakeip for --dns-hijack", servers)
	}
	dnsRules, _ := dns["rules"].([]any)
	if !dnsRulesBootstrapProxyHost(dnsRules, "proxy.example.com") {
		t.Fatalf("dns rules missing bootstrap for ProxyHost before fakeip: %v", dnsRules)
	}
	tun := findTunInbound(cfg)
	if tun["strict_route"] != true {
		t.Fatalf("strict_route = %v, want true with --dns-hijack", tun["strict_route"])
	}
	excludes, _ := tun["route_exclude_address"].([]any)
	// Static CF CIDRs removed; excludes are LAN + bootstrap DNS + resolved /32s.
	if excludeContainsCIDR(excludes, "198.41.128.0/17") {
		t.Fatalf("static CF CIDR should not appear: %v", excludes)
	}
	if !excludeContainsCIDR(excludes, "8.8.8.8/32") {
		t.Fatalf("route_exclude missing bootstrap DNS 8.8.8.8/32: %v", excludes)
	}
}

func excludeContainsCIDR(excludes []any, want string) bool {
	for _, e := range excludes {
		if e == want {
			return true
		}
	}
	return false
}

func dnsRulesBootstrapProxyHost(rules []any, host string) bool {
	for _, r := range rules {
		rule, _ := r.(map[string]any)
		if rule["server"] != "bootstrap" {
			continue
		}
		if domains, ok := rule["domain"].([]any); ok {
			for _, d := range domains {
				if d == host {
					return true
				}
			}
		}
		if domains, ok := rule["domain"].([]string); ok {
			for _, d := range domains {
				if d == host {
					return true
				}
			}
		}
	}
	return false
}

func TestBuildSingBoxFullVPNBlacklistExclude(t *testing.T) {
	policy, err := ParseDomainPolicy(PolicyInput{Exclude: []string{"github.com"}})
	if err != nil {
		t.Fatalf("ParseDomainPolicy: %v", err)
	}
	data, err := BuildTunConfig(&BuildConfigOptions{
		LocalSocksPort: 11080,
		ProxyHost:      "proxy.example.com",
		Policy:         policy,
	})
	if err != nil {
		t.Fatalf("BuildTunConfig: %v", err)
	}
	var cfg map[string]any
	_ = json.Unmarshal(data, &cfg)
	route, _ := cfg["route"].(map[string]any)
	if route["final"] != proxyOutboundTag {
		t.Fatalf("final = %v, want %s", route["final"], proxyOutboundTag)
	}
	rules := routeRulesFromCfg(cfg)
	if rulesContainCatchAll(rules) {
		t.Fatalf("full VPN should not include HTTP catch-all: %v", rules)
	}
}

func TestBuildSingBoxHttpOnlyAlsoProxyRules(t *testing.T) {
	also, err := ParseAlsoProxyPatterns([]string{
		"git.example.com:22",
		"*.db.internal:6606",
	})
	if err != nil {
		t.Fatalf("ParseAlsoProxyPatterns: %v", err)
	}
	data, err := BuildSingBoxHttpOnlyTunConfig(&HttpOnlyConfigOptions{
		LocalSocksPort: 11080,
		ProxyHost:      "proxy.example.com",
		AlsoProxy:      also,
	})
	if err != nil {
		t.Fatalf("BuildSingBoxHttpOnlyTunConfig: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rules := routeRulesFromCfg(cfg)
	if !rulesContainCatchAll(rules) {
		t.Fatalf("http-only catch-all should remain: %v", rules)
	}
	foundSSH := false
	foundMySQL := false
	for _, r := range rules {
		if r["outbound"] != webSelectorTag {
			continue
		}
		if dom, _ := r["domain"].([]any); len(dom) == 1 && dom[0] == "git.example.com" {
			if port, ok := asInt(r["port"]); ok && port == 22 {
				foundSSH = true
			}
		}
		if suf, _ := r["domain_suffix"].([]any); len(suf) == 1 && suf[0] == ".db.internal" {
			if port, ok := asInt(r["port"]); ok && port == 6606 {
				foundMySQL = true
			}
		}
	}
	if !foundSSH {
		t.Fatalf("missing git.example.com:22 also-proxy rule: %v", rules)
	}
	if !foundMySQL {
		t.Fatalf("missing *.db.internal:6606 also-proxy rule: %v", rules)
	}
}

func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}

func TestBuildSingBoxHttpOnlyWhitelistOmitsCatchAll(t *testing.T) {
	policy, err := ParseDomainPolicy(PolicyInput{Include: []string{"*.corp.com"}})
	if err != nil {
		t.Fatalf("ParseDomainPolicy: %v", err)
	}
	data, err := BuildSingBoxHttpOnlyTunConfig(&HttpOnlyConfigOptions{
		LocalSocksPort: 11080,
		ProxyHost:      "proxy.example.com",
		Policy:         policy,
	})
	if err != nil {
		t.Fatalf("BuildSingBoxHttpOnlyTunConfig: %v", err)
	}
	var cfg map[string]any
	_ = json.Unmarshal(data, &cfg)
	rules := routeRulesFromCfg(cfg)
	if rulesContainCatchAll(rules) {
		t.Fatalf("whitelist should not include catch-all: %v", rules)
	}
}

func findTunInbound(cfg map[string]any) map[string]any {
	inbounds, _ := cfg["inbounds"].([]any)
	for _, in := range inbounds {
		m, _ := in.(map[string]any)
		if m != nil && m["type"] == "tun" {
			return m
		}
	}
	return nil
}

func routeRulesFromCfg(cfg map[string]any) []map[string]any {
	route, _ := cfg["route"].(map[string]any)
	rules, _ := route["rules"].([]any)
	var out []map[string]any
	for _, r := range rules {
		if m, ok := r.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func rulesContainHijackDNS(rules []map[string]any) bool {
	for _, r := range rules {
		if r["action"] == "hijack-dns" {
			return true
		}
	}
	return false
}

func rulesContainCatchAll(rules []map[string]any) bool {
	for _, r := range rules {
		if r["type"] == "logical" && r["outbound"] == webSelectorTag {
			return true
		}
	}
	return false
}

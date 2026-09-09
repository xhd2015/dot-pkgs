package singboxtun

import (
	"encoding/json"
	"net"
	"testing"
)

func TestBuildTunDNSConfigProxyHostBeforeFakeIP(t *testing.T) {
	cfg := buildTunDNSConfig("workdev-to-local.example.com", 11080, true)
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	rules, _ := parsed["rules"].([]any)
	var sawBootstrap, sawFakeIP bool
	bootstrapIdx, fakeIdx := -1, -1
	for i, r := range rules {
		rule, _ := r.(map[string]any)
		switch rule["server"] {
		case "bootstrap":
			if domains, ok := rule["domain"].([]any); ok {
				for _, d := range domains {
					if d == "workdev-to-local.example.com" {
						sawBootstrap = true
						bootstrapIdx = i
					}
				}
			}
		case "fakeip":
			if !sawFakeIP {
				sawFakeIP = true
				fakeIdx = i
			}
		}
	}
	if !sawBootstrap {
		t.Fatalf("missing ProxyHost bootstrap rule: %v", rules)
	}
	if !sawFakeIP {
		t.Fatalf("missing fakeip rule: %v", rules)
	}
	if bootstrapIdx > fakeIdx {
		t.Fatalf("ProxyHost bootstrap must precede fakeip (bootstrap=%d fakeip=%d)", bootstrapIdx, fakeIdx)
	}
}

func TestProxyHostBootstrapDNSRulesIncludesCFSuffixes(t *testing.T) {
	rules := proxyHostBootstrapDNSRules("hub.example.com")
	var hasExact, hasCF bool
	for _, rule := range rules {
		if domains, ok := rule["domain"].([]string); ok {
			for _, d := range domains {
				if d == "hub.example.com" {
					hasExact = true
				}
			}
		}
		if suffixes, ok := rule["domain_suffix"].([]string); ok {
			for _, s := range suffixes {
				if s == "cloudflare.com" {
					hasCF = true
				}
			}
		}
	}
	if !hasExact || !hasCF {
		t.Fatalf("rules=%v want exact hub + cloudflare.com suffix", rules)
	}
}

func TestTunRouteExcludeResolvesHostsToSlash32(t *testing.T) {
	old := lookupHostIPv4
	lookupHostIPv4 = func(host string) []net.IP {
		switch host {
		case "hub.example.com":
			return []net.IP{net.ParseIP("172.67.1.1")}
		case "region1.v2.argotunnel.com":
			return []net.IP{net.ParseIP("198.41.192.7"), net.ParseIP("198.41.192.67")}
		case "region2.v2.argotunnel.com":
			return []net.IP{net.ParseIP("198.41.200.13")}
		default:
			return nil
		}
	}
	defer func() { lookupHostIPv4 = old }()

	ex := tunRouteExcludeAddresses("hub.example.com", nil)
	want := map[string]bool{
		"172.67.1.1/32":    false,
		"198.41.192.7/32":  false,
		"198.41.192.67/32": false,
		"198.41.200.13/32": false,
		"8.8.8.8/32":       false,
		"10.0.0.0/8":       false,
	}
	for _, cidr := range ex {
		if _, ok := want[cidr]; ok {
			want[cidr] = true
		}
	}
	for cidr, ok := range want {
		if !ok {
			t.Fatalf("missing %s in %v", cidr, ex)
		}
	}
	for _, cidr := range ex {
		if cidr == "198.41.128.0/17" || cidr == "104.16.0.0/13" {
			t.Fatalf("static CF CIDR %s should not appear: %v", cidr, ex)
		}
	}
}

func TestTunnelExcludeHostsMergesProxyAndDefaults(t *testing.T) {
	hosts := tunnelExcludeHosts("hub.example.com", []string{"extra.example"})
	joined := map[string]bool{}
	for _, h := range hosts {
		joined[h] = true
	}
	for _, want := range []string{"hub.example.com", "region1.v2.argotunnel.com", "extra.example"} {
		if !joined[want] {
			t.Fatalf("missing %s in %v", want, hosts)
		}
	}
}

func TestExcludeCIDRsEqual(t *testing.T) {
	a := []string{"1.1.1.1/32", "8.8.8.8/32"}
	b := []string{"1.1.1.1/32", "8.8.8.8/32"}
	c := []string{"1.1.1.1/32"}
	if !excludeCIDRsEqual(a, b) {
		t.Fatal("expected equal")
	}
	if excludeCIDRsEqual(a, c) {
		t.Fatal("expected not equal")
	}
}

func TestSkipBindInterfaceOmitsBindAndBootstrapDetour(t *testing.T) {
	old := lookupHostIPv4
	lookupHostIPv4 = func(host string) []net.IP {
		return []net.IP{net.ParseIP("1.2.3.4")}
	}
	defer func() { lookupHostIPv4 = old }()

	data, err := BuildTunConfig(&BuildConfigOptions{
		LocalSocksPort:    11080,
		ProxyHost:         "proxy.example.com",
		HttpOnly:          true,
		DNSHijack:         true,
		SkipBindInterface: true,
		InitialUseProxy:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	for _, o := range cfg["outbounds"].([]any) {
		ob := o.(map[string]any)
		if ob["tag"] == "direct" {
			if _, has := ob["bind_interface"]; has {
				t.Fatalf("direct should not pin bind_interface when SkipBindInterface: %v", ob)
			}
		}
	}
	dns, _ := cfg["dns"].(map[string]any)
	for _, s := range dns["servers"].([]any) {
		server, _ := s.(map[string]any)
		if server["tag"] == "bootstrap" {
			if _, has := server["detour"]; has {
				t.Fatalf("bootstrap must not detour→direct when SkipBindInterface: %v", server)
			}
		}
	}
}

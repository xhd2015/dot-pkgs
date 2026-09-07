package singboxtun

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"
)

func TestDefaultLookupHostIPv4RespectsTimeout(t *testing.T) {
	start := time.Now()
	ips := defaultLookupHostIPv4("definitely-does-not-exist-xyz.invalid")
	elapsed := time.Since(start)
	if elapsed > hostLookupTimeout+time.Second {
		t.Fatalf("lookup took %s, want <= %s", elapsed, hostLookupTimeout)
	}
	if len(ips) != 0 {
		t.Fatalf("ips = %#v, want nil", ips)
	}
}

func TestResolveHostIPv4CIDRsFromIPs(t *testing.T) {
	old := lookupHostIPv4
	defer func() { lookupHostIPv4 = old }()

	lookupHostIPv4 = func(host string) []net.IP {
		return []net.IP{
			net.ParseIP("1.2.3.4"),
			net.ParseIP("1.2.3.4"),
			net.ParseIP("2001:db8::1"),
		}
	}

	cidrs := resolveHostIPv4CIDRs("proxy.example.com")
	if len(cidrs) != 1 || cidrs[0] != "1.2.3.4/32" {
		t.Fatalf("cidrs = %#v, want [1.2.3.4/32]", cidrs)
	}
}

func TestBuildTunConfigSocksOutbound(t *testing.T) {
	old := lookupHostIPv4
	defer func() { lookupHostIPv4 = old }()
	lookupHostIPv4 = func(host string) []net.IP {
		return []net.IP{net.ParseIP("93.184.216.34")}
	}

	data, err := BuildTunConfig(&BuildConfigOptions{
		LocalSocksPort: 11080,
		ProxyHost:      "proxy.example.com",
	})
	if err != nil {
		t.Fatalf("BuildTunConfig: %v", err)
	}
	if !strings.Contains(string(data), `"type": "socks"`) {
		t.Fatalf("expected socks outbound: %s", data)
	}
	if strings.Contains(string(data), `"type": "vmess"`) {
		t.Fatalf("vmess outbound should not appear: %s", data)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	dns, _ := cfg["dns"].(map[string]any)
	servers, _ := dns["servers"].([]any)
	for _, s := range servers {
		server, _ := s.(map[string]any)
		if server["tag"] == "remote" {
			t.Fatalf("SOCKS config should not include remote DNS server: %v", servers)
		}
	}
	route, _ := cfg["route"].(map[string]any)
	if route["default_domain_resolver"] != "bootstrap" {
		t.Fatalf("default_domain_resolver = %v, want bootstrap", route["default_domain_resolver"])
	}
	tun := findTunInbound(cfg)
	exclude, _ := tun["route_exclude_address"].([]any)
	foundProxyIP := false
	for _, e := range exclude {
		if e == "93.184.216.34/32" {
			foundProxyIP = true
		}
	}
	if !foundProxyIP {
		t.Fatalf("route_exclude_address missing proxy IP: %v", exclude)
	}
}

func TestBuildTunConfigRequiresLocalSocksPort(t *testing.T) {
	_, err := BuildTunConfig(&BuildConfigOptions{ProxyHost: "proxy.example.com"})
	if err == nil {
		t.Fatal("expected error when LocalSocksPort missing")
	}
}

func TestResolveHostIPv4CIDRsLookupFailure(t *testing.T) {
	old := lookupHostIPv4
	defer func() { lookupHostIPv4 = old }()

	lookupHostIPv4 = func(host string) []net.IP {
		return nil
	}

	if cidrs := resolveHostIPv4CIDRs("missing.example.com"); len(cidrs) != 0 {
		t.Fatalf("cidrs = %#v, want nil", cidrs)
	}
}

package backend

import "testing"

func TestParseIngressHostOrigins(t *testing.T) {
	yml := `
tunnel: abc
ingress:
  - hostname: a.example.com
    service: http://10.0.2.2:23712
  - hostname: b.example.com
    service: http://10.0.2.2:18080
  - service: http_status:404
`
	got := parseIngressHostOrigins(yml)
	if got["a.example.com"] != "http://10.0.2.2:23712" {
		t.Fatalf("a = %q", got["a.example.com"])
	}
	if got["b.example.com"] != "http://10.0.2.2:18080" {
		t.Fatalf("b = %q", got["b.example.com"])
	}
	if _, ok := got[""]; ok {
		t.Fatal("unexpected empty host")
	}
}

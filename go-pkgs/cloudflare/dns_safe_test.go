package cloudflare

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func safeDNSClient(t *testing.T, h http.HandlerFunc) *OriginCertDNSClient {
	t.Helper()
	server := httptest.NewServer(h)
	t.Cleanup(server.Close)
	cert := filepath.Join(t.TempDir(), "cert.pem")
	data := base64.StdEncoding.EncodeToString([]byte(`{"zoneID":"zone-1","apiToken":"fake-secret","accountID":"account-1"}`))
	if err := os.WriteFile(cert, []byte("-----BEGIN ARGO TUNNEL TOKEN-----\n"+data+"\n-----END ARGO TUNNEL TOKEN-----"), 0600); err != nil {
		t.Fatal(err)
	}
	return &OriginCertDNSClient{CertPath: cert, APIBase: server.URL, HTTPClient: server.Client()}
}
func TestSafeDNSNeverOverwritesExisting(t *testing.T) {
	var writes atomic.Int32
	client := safeDNSClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-secret" {
			t.Error("missing credentials")
		}
		if r.Method != "GET" {
			writes.Add(1)
		}
		fmt.Fprint(w, `{"success":true,"result":[{"id":"owned-by-other","type":"A","name":"p.example.com","content":"192.0.2.2"}]}`)
	})
	if err := client.EnsureTunnelCNAME(context.Background(), "p.example.com", "tunnel-1"); err == nil {
		t.Fatal("expected hostname conflict")
	}
	if writes.Load() != 0 {
		t.Fatal("overwrote another service")
	}
}
func TestSafeDNSCreatesAndReusesOnlyOwnedTunnel(t *testing.T) {
	var creates atomic.Int32
	var exists atomic.Bool
	client := safeDNSClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			creates.Add(1)
			exists.Store(true)
			fmt.Fprint(w, `{"success":true,"result":{}}`)
			return
		}
		if exists.Load() {
			fmt.Fprint(w, `{"success":true,"result":[{"type":"CNAME","name":"p.example.com","content":"tunnel-1.cfargotunnel.com","proxied":true}]}`)
		} else {
			fmt.Fprint(w, `{"success":true,"result":[]}`)
		}
	})
	for i := 0; i < 2; i++ {
		if err := client.EnsureTunnelCNAME(context.Background(), "p.example.com", "tunnel-1"); err != nil {
			t.Fatal(err)
		}
	}
	if creates.Load() != 1 {
		t.Fatal(creates.Load())
	}
	if err := client.EnsureTunnelCNAME(context.Background(), "p.example.com", "different-tunnel"); err == nil {
		t.Fatal("accepted wrong tunnel")
	}
}
func TestSafeDNSVerifiesZoneAccess(t *testing.T) {
	client := safeDNSClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success":true,"result":{"name":"example.com"}}`)
	})
	if err := client.VerifyDomain(context.Background(), "example.com"); err != nil {
		t.Fatal(err)
	}
	if err := client.VerifyDomain(context.Background(), "unowned.example.org"); err == nil {
		t.Fatal("accepted unowned domain")
	}
}

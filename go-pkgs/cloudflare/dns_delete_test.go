package cloudflare

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgoTunnelToken(t *testing.T) {
	t.Parallel()
	payload, _ := json.Marshal(map[string]string{
		"accountID": "acc",
		"zoneID":    "zone123",
		"apiToken":  "tok456",
	})
	b64 := base64.StdEncoding.EncodeToString(payload)
	pem := "-----BEGIN ARGO TUNNEL TOKEN-----\n" + b64 + "\n-----END ARGO TUNNEL TOKEN-----\n"
	tok, err := parseArgoTunnelToken(pem)
	if err != nil {
		t.Fatal(err)
	}
	if tok.ZoneID != "zone123" || tok.APIToken != "tok456" {
		t.Fatalf("got %+v", tok)
	}
}

func TestOriginCertDNSDeleter_DeleteHostname(t *testing.T) {
	t.Parallel()
	var deleted []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "no auth", 401)
			return
		}
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/dns_records"):
			_, _ = io.WriteString(w, `{"success":true,"result":[{"id":"rec1","name":"a.example.com"},{"id":"rec2","name":"a.example.com"}]}`)
		case r.Method == http.MethodDelete:
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			deleted = append(deleted, parts[len(parts)-1])
			_, _ = io.WriteString(w, `{"success":true,"result":{}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	payload, _ := json.Marshal(map[string]string{
		"accountID": "acc",
		"zoneID":    "zone123",
		"apiToken":  "tok456",
	})
	pem := "-----BEGIN ARGO TUNNEL TOKEN-----\n" + base64.StdEncoding.EncodeToString(payload) + "\n-----END ARGO TUNNEL TOKEN-----\n"
	certPath := filepath.Join(dir, "cert.pem")
	if err := os.WriteFile(certPath, []byte(pem), 0o644); err != nil {
		t.Fatal(err)
	}

	d := &OriginCertDNSDeleter{
		CertPath:   certPath,
		HTTPClient: srv.Client(),
		APIBase:    srv.URL,
	}
	if err := d.DeleteHostname("a.example.com"); err != nil {
		t.Fatal(err)
	}
	if len(deleted) != 2 || deleted[0] != "rec1" || deleted[1] != "rec2" {
		t.Fatalf("deleted=%v want [rec1 rec2]", deleted)
	}
}

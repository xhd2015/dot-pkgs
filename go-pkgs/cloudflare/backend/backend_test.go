package backend

import (
	"testing"
)

func TestNewSelectsKind(t *testing.T) {
	h := New(Select{Qemu: false})
	if h.Kind() != KindHost {
		t.Fatalf("Kind = %s, want host", h.Kind())
	}
	if got := h.OriginForHostPort(18080); got != "http://127.0.0.1:18080" {
		t.Fatalf("host origin = %q", got)
	}

	q := New(Select{Qemu: true})
	if q.Kind() != KindQemu {
		t.Fatalf("Kind = %s, want qemu", q.Kind())
	}
	if got := q.OriginForHostPort(18080); got != "http://10.0.2.2:18080" {
		t.Fatalf("qemu origin = %q", got)
	}
}

func TestFromQemuEnabled(t *testing.T) {
	if FromQemuEnabled(true).Kind() != KindQemu {
		t.Fatal("expected qemu")
	}
	if FromQemuEnabled(false).Kind() != KindHost {
		t.Fatal("expected host")
	}
}

func TestValidateRoute(t *testing.T) {
	if err := validateRoute(Route{}); err == nil {
		t.Fatal("expected error")
	}
	if err := validateRoute(Route{Hostname: "a.example.com"}); err == nil {
		t.Fatal("expected origin error")
	}
	if err := validateRoute(Route{Hostname: "a.example.com", Origin: "http://127.0.0.1:1"}); err != nil {
		t.Fatal(err)
	}
}

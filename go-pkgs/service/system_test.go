package service

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

type statusOnlyController struct{ runtime SystemRuntime }

func (c *statusOnlyController) SystemStatus() SystemRuntime { return c.runtime }

type lifecycleController struct {
	statusOnlyController
	started int
	stopped int
}

func (c *lifecycleController) StartSystem() error { c.started++; return nil }
func (c *lifecycleController) StopSystem() error  { c.stopped++; return nil }

type restarterController struct {
	lifecycleController
	restarted int
}

func (c *restarterController) RestartSystem() error { c.restarted++; return nil }

type autoStartController struct {
	lifecycleController
	enabled *bool
}

func (c *autoStartController) SetSystemAutoStart(enabled bool) error {
	c.enabled = Bool(enabled)
	return nil
}

type streamController struct{ lifecycleController }

func (c *streamController) StreamSystemOperation(w http.ResponseWriter, action string) error {
	return nil
}

func TestNormalizeServiceKind(t *testing.T) {
	if got := NormalizeServiceKind(""); got != ServiceKindUser {
		t.Fatalf("NormalizeServiceKind(\"\") = %q", got)
	}
	if got := NormalizeServiceKind("bogus"); got != ServiceKindUser {
		t.Fatalf("NormalizeServiceKind(bogus) = %q", got)
	}
	if got := NormalizeServiceKind(ServiceKindSystem); got != ServiceKindSystem {
		t.Fatalf("NormalizeServiceKind(system) = %q", got)
	}
}

func TestSystemActionsFollowControllerCapabilities(t *testing.T) {
	cases := []struct {
		name string
		ctrl SystemController
		want []string
	}{
		{"status only", &statusOnlyController{}, nil},
		{"lifecycle", &lifecycleController{}, []string{"start", "stop", "restart"}},
		{"restarter", &restarterController{}, []string{"start", "stop", "restart"}},
		{"streamer", &streamController{}, []string{"start", "stop", "restart"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SystemActions(tc.ctrl)
			if len(got) != len(tc.want) {
				t.Fatalf("SystemActions() = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("SystemActions() = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestSystemSupportsAction(t *testing.T) {
	statusOnly := &statusOnlyController{}
	if SystemSupportsAction(statusOnly, SystemActionStart) {
		t.Fatalf("status-only controller advertises start")
	}
	lifecycle := &lifecycleController{}
	for _, action := range []string{SystemActionStart, SystemActionStop, SystemActionRestart} {
		if !SystemSupportsAction(lifecycle, action) {
			t.Fatalf("lifecycle controller does not advertise %s", action)
		}
	}
	if SystemSupportsAction(lifecycle, "explode") {
		t.Fatalf("lifecycle controller advertises an unknown action")
	}
}

func TestSystemSupportsOperationStream(t *testing.T) {
	if SystemSupportsOperationStream(&lifecycleController{}) {
		t.Fatalf("plain lifecycle controller advertises a stream")
	}
	if !SystemSupportsOperationStream(&streamController{}) {
		t.Fatalf("stream controller does not advertise a stream")
	}
}

func TestSystemActionHint(t *testing.T) {
	for _, action := range []string{SystemActionStart, SystemActionStop} {
		if hint := SystemActionHint(action); !strings.Contains(hint, "Port Forwarding") {
			t.Fatalf("SystemActionHint(%q) = %q", action, hint)
		}
	}
	if hint := SystemActionHint(SystemActionRestart); !strings.Contains(hint, "nothing to reconcile") {
		t.Fatalf("SystemActionHint(restart) = %q", hint)
	}
	if hint := SystemActionHint("explode"); !strings.Contains(hint, "owned by the server") {
		t.Fatalf("SystemActionHint(explode) = %q", hint)
	}
}

func TestErrSystemActionUnsupportedNamesServiceAndHint(t *testing.T) {
	svc := SystemService{ID: "tunnel", Name: "Domain tunnel"}
	err := ErrSystemActionUnsupported(svc, SystemActionRestart)
	if !strings.Contains(err.Error(), `"Domain tunnel"`) || !strings.Contains(err.Error(), "restart") {
		t.Fatalf("error = %v", err)
	}
}

func TestReadOnlySystemServiceErrorWrapsSentinel(t *testing.T) {
	err := ReadOnlySystemServiceError(SystemService{Name: "Tunnel"}, "renamed")
	if !errors.Is(err, ErrSystemServiceImmutable) {
		t.Fatalf("error = %v, want ErrSystemServiceImmutable", err)
	}
	if !strings.Contains(err.Error(), `"Tunnel"`) || !strings.Contains(err.Error(), "renamed") {
		t.Fatalf("error = %v", err)
	}
}

func TestSystemRegistryRegisterValidates(t *testing.T) {
	registry := NewSystemRegistry()
	cases := []struct {
		name string
		svc  SystemService
		want string
	}{
		{"empty id", SystemService{Name: "n", Controller: &statusOnlyController{}}, "id is required"},
		{"empty name", SystemService{ID: "a", Controller: &statusOnlyController{}}, "name is required"},
		{"no controller", SystemService{ID: "a", Name: "n"}, "controller is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := registry.Register(tc.svc); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Register() error = %v, want %q", err, tc.want)
			}
		})
	}
	if ids := registry.IDs(); len(ids) != 0 {
		t.Fatalf("IDs() = %v, want empty", ids)
	}
}

func TestSystemRegistryTrimsAndRejectsDuplicates(t *testing.T) {
	registry := NewSystemRegistry()
	svc := SystemService{ID: "  llm-proxy ", Name: "  llm-proxy  ", Description: "  proxy  ", Controller: &statusOnlyController{}}
	if err := registry.Register(svc); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	got, ok := registry.Get("llm-proxy")
	if !ok {
		t.Fatalf("Get() = not found")
	}
	if got.ID != "llm-proxy" || got.Name != "llm-proxy" || got.Description != "proxy" {
		t.Fatalf("registered service = %+v, want trimmed fields", got)
	}
	if err := registry.Register(svc); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("duplicate Register() error = %v", err)
	}
	if ids := registry.IDs(); len(ids) != 1 || ids[0] != "llm-proxy" {
		t.Fatalf("IDs() = %v", ids)
	}
}

func TestSystemRegistryListIsACopy(t *testing.T) {
	registry := NewSystemRegistry()
	if err := registry.Register(SystemService{ID: "a", Name: "A", Controller: &statusOnlyController{}}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	list := registry.List()
	list[0].ID = "mutated"
	if got, _ := registry.Get("a"); got.ID != "a" {
		t.Fatalf("List() aliases the registry: %+v", got)
	}
}

func TestNilSystemRegistryIsUsable(t *testing.T) {
	var registry *SystemRegistry
	if list := registry.List(); list != nil {
		t.Fatalf("List() = %v", list)
	}
	if _, ok := registry.Get("a"); ok {
		t.Fatalf("Get() = found on a nil registry")
	}
	if ids := registry.IDs(); len(ids) != 0 {
		t.Fatalf("IDs() = %v", ids)
	}
}

package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// ServiceKind separates services a user defines from services the host owns.
// User services are child processes; system services are in-process
// subsystems that cannot be edited, renamed or removed.
type ServiceKind string

const (
	// ServiceKindUser is a definition the host stored; the supervisor runs its
	// command as a child process.
	ServiceKindUser ServiceKind = "user"
	// ServiceKindSystem is a host-owned in-process subsystem. It never appears
	// in a definitions file.
	ServiceKindSystem ServiceKind = "system"
)

// NormalizeServiceKind defaults an absent or unknown kind to user, so payloads
// written before kinds existed stay readable.
func NormalizeServiceKind(kind ServiceKind) ServiceKind {
	if kind == ServiceKindSystem {
		return ServiceKindSystem
	}
	return ServiceKindUser
}

// States a published hostname can be in. "starting" means the dial pool has not
// connected within its warm-up grace window; "unknown" means the edge could not
// be reached, so its state could not be established either way.
const (
	SystemHostLive     = "live"
	SystemHostStarting = "starting"
	SystemHostDead     = "dead"
	SystemHostMissing  = "missing"
	SystemHostUnknown  = "unknown"
)

// SystemHostStatus is one public hostname a system service publishes.
type SystemHostStatus struct {
	Host  string `json:"host"`
	Dials int    `json:"dials"`
	State string `json:"state"`
}

// SystemRuntime is the live state a system service reports. Only Running is
// required; every other field renders as extra detail when present.
type SystemRuntime struct {
	// Running reports whether the subsystem is currently up.
	Running bool
	// Detail is a short human line, e.g. "listening :21000".
	Detail string
	// PublicURL is the public address the subsystem serves, when it has one.
	PublicURL string
	// Port is the local listen port, when the subsystem listens on one.
	Port int
	// LogPath is a log the GUI can stream, when the subsystem keeps one.
	LogPath string
	// Mocked marks a subsystem whose work is simulated, so the GUI can label it
	// instead of implying a real integration.
	Mocked bool
	// AutoStart mirrors the subsystem's own auto-start switch. Nil means the
	// subsystem has no such switch, and enable/disable is hidden.
	AutoStart *bool
	// Edge is the upstream a proxying subsystem publishes through, when it has
	// one.
	Edge string
	// Hosts lists the public hostnames the subsystem is responsible for.
	Hosts []SystemHostStatus
	// Status overrides the running/stopped derivation when set. Used for
	// "error" (a host is dead) and "unknown" (the edge could not be reached).
	Status string
}

// SystemController reports a subsystem's state. It is the only required half:
// a controller that cannot start, stop, or reconcile anything is still a valid
// system service, and is rendered as status-only.
type SystemController interface {
	SystemStatus() SystemRuntime
}

// SystemLifecycle is implemented by controllers that can start and stop their
// subsystem. Start and Stop must be safe when already in the target state.
type SystemLifecycle interface {
	StartSystem() error
	StopSystem() error
}

// SystemRestarter is implemented by controllers that can reconcile a running
// subsystem, e.g. by republishing hostnames whose dial pool died.
type SystemRestarter interface {
	RestartSystem() error
}

// SystemHost is one hostname a system service is responsible for publishing.
// ServiceID is the owning user service, and is empty for a domain tunnel.
type SystemHost struct {
	Host      string
	ServiceID string
}

// SystemHostProvider is implemented by controllers that need to know which
// hostnames the host considers theirs. The host pushes the set before asking
// for status, because it already holds its lock there: a controller that called
// back into the host would deadlock.
type SystemHostProvider interface {
	SetSystemHosts(hosts []SystemHost)
}

// SystemAutoStarter is implemented by controllers that can toggle the
// subsystem's own auto-start flag. A controller that omits it offers no
// enable/disable action.
type SystemAutoStarter interface {
	SetSystemAutoStart(enabled bool) error
}

// SystemOperationStreamer is implemented by controllers that can run a
// start/stop/restart operation while streaming its logs over an SSE response.
type SystemOperationStreamer interface {
	StreamSystemOperation(w http.ResponseWriter, action string) error
}

// Actions a system service can advertise. Hosts only offer the actions a
// controller actually implements.
const (
	SystemActionStart   = "start"
	SystemActionStop    = "stop"
	SystemActionRestart = "restart"
)

// Messages reported for system-service enable/disable. They mirror the user
// service wording: the toggle controls boot auto-start, not the live state.
const (
	SystemMessageDisable = "The service won't stop immediately unless you manually stop it"
	SystemMessageEnable  = "The service won't start immediately until the next server boot"
)

// ErrSystemServiceImmutable marks a write aimed at a system service. Hosts map
// it to a 400 rather than 404: the service exists, it just cannot be changed.
var ErrSystemServiceImmutable = errors.New("system service is owned by the server")

// SystemService is one host-owned service in a registry.
type SystemService struct {
	ID          string
	Name        string
	Description string
	Controller  SystemController
}

// SystemActions lists the actions a controller supports, in display order.
func SystemActions(ctrl SystemController) []string {
	actions := make([]string, 0, 3)
	_, lifecycle := ctrl.(SystemLifecycle)
	_, restarter := ctrl.(SystemRestarter)
	if lifecycle {
		actions = append(actions, SystemActionStart, SystemActionStop)
	}
	if restarter || lifecycle {
		actions = append(actions, SystemActionRestart)
	}
	return actions
}

// SystemSupportsAction reports whether ctrl advertises action.
func SystemSupportsAction(ctrl SystemController, action string) bool {
	for _, supported := range SystemActions(ctrl) {
		if supported == action {
			return true
		}
	}
	return false
}

// SystemSupportsOperationStream reports whether the service can run an action
// through an SSE operation stream instead of the plain POST endpoint.
func SystemSupportsOperationStream(ctrl SystemController) bool {
	_, ok := ctrl.(SystemOperationStreamer)
	return ok
}

// SystemActionHint explains what to do instead when an action is unsupported,
// rather than just "not supported".
func SystemActionHint(action string) string {
	switch action {
	case SystemActionStart, SystemActionStop:
		return "control the individual hostnames in Port Forwarding instead"
	case SystemActionRestart:
		return "it has nothing to reconcile"
	default:
		return "it is owned by the server"
	}
}

// ErrSystemActionUnsupported explains a refused action in terms of what the
// caller should do instead.
func ErrSystemActionUnsupported(svc SystemService, action string) error {
	return fmt.Errorf("system service %q does not support %s; %s", svc.Name, action, SystemActionHint(action))
}

// ReadOnlySystemServiceError marks a write aimed at a system service.
func ReadOnlySystemServiceError(svc SystemService, action string) error {
	return fmt.Errorf("system service %q cannot be %s: %w", svc.Name, action, ErrSystemServiceImmutable)
}

// SystemRegistry holds the system services a host exposes alongside its child
// process services. It is safe for concurrent use.
type SystemRegistry struct {
	mu       sync.RWMutex
	services []SystemService
}

// NewSystemRegistry returns an empty registry.
func NewSystemRegistry() *SystemRegistry {
	return &SystemRegistry{}
}

// Register adds svc. IDs are trimmed; an empty id, empty name, missing
// controller, or duplicate id is an error.
func (r *SystemRegistry) Register(svc SystemService) error {
	svc.ID = strings.TrimSpace(svc.ID)
	svc.Name = strings.TrimSpace(svc.Name)
	svc.Description = strings.TrimSpace(svc.Description)
	switch {
	case svc.ID == "":
		return fmt.Errorf("system service id is required")
	case svc.Name == "":
		return fmt.Errorf("system service %s: name is required", svc.ID)
	case svc.Controller == nil:
		return fmt.Errorf("system service %s: controller is required", svc.ID)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.services {
		if existing.ID == svc.ID {
			return fmt.Errorf("system service %s is already registered", svc.ID)
		}
	}
	r.services = append(r.services, svc)
	return nil
}

// List returns the registered services in registration order.
func (r *SystemRegistry) List() []SystemService {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]SystemService(nil), r.services...)
}

// Get resolves id to a registered service.
func (r *SystemRegistry) Get(id string) (SystemService, bool) {
	if r == nil {
		return SystemService{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, svc := range r.services {
		if svc.ID == id {
			return svc, true
		}
	}
	return SystemService{}, false
}

// IDs returns the registered ids in registration order.
func (r *SystemRegistry) IDs() []string {
	services := r.List()
	ids := make([]string, 0, len(services))
	for _, svc := range services {
		ids = append(ids, svc.ID)
	}
	return ids
}

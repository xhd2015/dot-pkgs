# eventbus — shared event contract + HTTP publish client

L2 in-process doctests for `github.com/xhd2015/dot-pkgs/go-pkgs/eventbus`
(PHASE 1 foundation only).

## Version

0.0.2

# DSN (Domain Specific Notion)

Shared wire contract and thin clients so seatalk / agent-run publishers and
later hubs agree on one Event JSON envelope without depending on spl or ai-critic.

**Participants**

- **Event** — JSON envelope: `id`, `ts`, `source`, `type`, optional `host`, `payload`.
- **Constants** — locked port (`DefaultPublishPort`), v1 type strings, optional source strings.
- **Publisher** — HTTP client: `POST {baseURL}/publish` with JSON body; optional Bearer token;
  injectable `http.Client` / timeout; empty base URL is a no-op success.
- **HTTP mock** — `httptest.Server` in tests capturing method, path, headers, body.
- **ListenWS** — thin helper: dial WebSocket URL, decode JSON Event frames until ctx cancel.

**Behaviors**

- Event marshals/unmarshals with the locked field names; empty `host` is omitempty.
- `Publisher.Publish(ctx, Event)` posts JSON; non-2xx and transport failures return error
  (callers treat publish as best-effort).
- Token empty → no `Authorization` header; token set → `Authorization: Bearer <token>`.
- Empty base URL → `Publish` returns nil without performing HTTP.
- Default publish path is `/publish`; default port constant is `23891`.
- `ListenWS` delivers decoded events to a callback until context cancellation.

## Decision Tree

```
eventbus/tests/
├── constants/
│   └── locked-values/                    (LEAF) port + type + source string constants
├── event-json/
│   ├── round-trip-all-fields/            (LEAF) marshal/unmarshal preserves fields
│   └── host-omitted-when-empty/          (LEAF) empty host omitted from JSON
├── publish/
│   ├── empty-base-url-noop/              (LEAF) empty URL → nil, no HTTP
│   ├── success/
│   │   ├── posts-json-to-publish/        (LEAF) POST /publish + body; no Authorization
│   │   └── with-bearer-token/            (LEAF) Authorization: Bearer <token>
│   └── errors/
│       ├── non-2xx-status/               (LEAF) HTTP 500 → error
│       └── transport-error/              (LEAF) server closed → error
└── listen-ws/
    └── receives-one-event/               (LEAF) WS server sends one Event; callback gets it
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `constants/locked-values` | `DefaultPublishPort==23891`; type/source strings match locked wire names |
| 2 | `event-json/round-trip-all-fields` | `json.Marshal`/`Unmarshal` of `Event` preserves all fields including host |
| 3 | `event-json/host-omitted-when-empty` | Empty `Host` is absent from marshaled JSON (`omitempty`) |
| 4 | `publish/empty-base-url-noop` | Empty base URL: `Publish` returns nil; no HTTP request |
| 5 | `publish/success/posts-json-to-publish` | `POST /publish`, JSON body matches Event; no Authorization when token empty |
| 6 | `publish/success/with-bearer-token` | Token set → `Authorization: Bearer <token>` |
| 7 | `publish/errors/non-2xx-status` | Non-2xx response → non-nil error |
| 8 | `publish/errors/transport-error` | Unreachable server / closed listener → non-nil error |
| 9 | `listen-ws/receives-one-event` | `ListenWS` receives one JSON Event from a test WS server |

## How to Run

```sh
doctest vet ./external/dot-pkgs-master-2026-08-10-1/go-pkgs/eventbus/tests/
doctest test ./external/dot-pkgs-master-2026-08-10-1/go-pkgs/eventbus/tests/
```

```go
import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/eventbus"
)

// HTTPCapture records publish requests observed by the mock server.
type HTTPCapture struct {
	mu       sync.Mutex
	Requests []CapturedRequest
}

// CapturedRequest is one observed HTTP request.
type CapturedRequest struct {
	Method        string
	Path          string
	Authorization string
	ContentType   string
	Body          []byte
}

func (c *HTTPCapture) add(r CapturedRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Requests = append(c.Requests, r)
}

func (c *HTTPCapture) Len() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.Requests)
}

func (c *HTTPCapture) Last() (CapturedRequest, bool) {
	if c == nil {
		return CapturedRequest{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.Requests) == 0 {
		return CapturedRequest{}, false
	}
	return c.Requests[len(c.Requests)-1], true
}

// Request drives a single L2 scenario against the public eventbus API.
// Op selects the surface: constants | event-json | publish | listen-ws.
type Request struct {
	Op string

	// Event under test (event-json / publish / listen-ws fixtures).
	Event eventbus.Event

	// Publish configuration (product API inputs).
	BaseURL string
	Token   string

	// Test harness: optional HTTP mock for publish leaves.
	UseHTTPMock    bool
	MockStatusCode int  // 0 means 200
	CloseMockEarly bool // close listener before Publish (transport error)
	Capture        *HTTPCapture

	// Test harness: optional WS mock for listen-ws leaves.
	WSURL string
}

// Response holds observed package outputs for Assert.
type Response struct {
	// constants
	DefaultPublishPort         int
	TypeSeatalkMessageReceived string
	TypeSeatalkSessionOpened   string
	TypeAgentTTYStarted        string
	SourceSeatalkLocalBot      string
	SourceAgentRun             string

	// event-json
	JSONBytes []byte
	RoundTrip eventbus.Event

	// listen-ws
	ReceivedEvents []eventbus.Event
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch req.Op {
	case "constants":
		resp.DefaultPublishPort = eventbus.DefaultPublishPort
		resp.TypeSeatalkMessageReceived = eventbus.TypeSeatalkMessageReceived
		resp.TypeSeatalkSessionOpened = eventbus.TypeSeatalkSessionOpened
		resp.TypeAgentTTYStarted = eventbus.TypeAgentTTYStarted
		resp.SourceSeatalkLocalBot = eventbus.SourceSeatalkLocalBot
		resp.SourceAgentRun = eventbus.SourceAgentRun
		return resp, nil

	case "event-json":
		b, err := json.Marshal(req.Event)
		if err != nil {
			return resp, err
		}
		resp.JSONBytes = b
		var rt eventbus.Event
		if err := json.Unmarshal(b, &rt); err != nil {
			return resp, err
		}
		resp.RoundTrip = rt
		return resp, nil

	case "publish":
		pub := eventbus.NewPublisher(req.BaseURL,
			eventbus.WithToken(req.Token),
			eventbus.WithTimeout(2*time.Second),
		)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := pub.Publish(ctx, req.Event)
		return resp, err

	case "listen-ws":
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var got []eventbus.Event
		err := eventbus.ListenWS(ctx, req.WSURL, func(ev eventbus.Event) {
			got = append(got, ev)
			cancel() // one event is enough for this leaf
		})
		resp.ReceivedEvents = got
		// context.Canceled after successful receive is acceptable; Assert decides
		return resp, err

	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}
```

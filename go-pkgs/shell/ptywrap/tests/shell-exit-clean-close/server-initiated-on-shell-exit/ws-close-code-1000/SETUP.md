# Scenario

**Feature**: after shell exit, attach WebSocket must close with code 1000, not 1006

```
create sh -c sleep 1
  -> dial writer WS
  -> wait for server close after child exit
  -> CloseCode == 1000 (Normal Closure)
```

## Preconditions

- Phase `shell-exit-ws-close-code` records `resp.CloseCode` from the gorilla
  close error (or 1006 for abnormal / unexpected EOF).

## Steps

1. `Phase=shell-exit-ws-close-code`.
2. Default short-lived command from parent unless overridden.

## Context

Server half of the dual contract: on `<-s.done`, send close control **1000**
before closing the connection. Complements client marker-first hang-proofing
(see `marker-without-close`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "shell-exit-ws-close-code"
	return nil
}
```

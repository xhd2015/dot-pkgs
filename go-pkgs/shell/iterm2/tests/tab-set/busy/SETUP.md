# Scenario

**Feature**: classify session busy state from foreground process name

```
fg process name + probe ok
  -> ClassifyBusyFromComm
  -> BusyStateIdle | BusyStateBusy | BusyStateUnknown
```

## Preconditions

- Product exports `BusyState` constants and `ClassifyBusyFromComm(fgComm string, ok bool) BusyState`.

## Steps

1. Set `req.Phase` to `classify-busy`.
2. Leaf fills `FgComm` and `FgOK`.

## Context

- Pure function — no TTY device access in these leaves.
- Shell names: basename only (`zsh`) or with path (`/bin/zsh`) should count as idle.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "classify-busy"
	return nil
}
```

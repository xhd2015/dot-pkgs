# Scenario

**Feature**: discover tab-set sessions via marker scan and pure parse

```
# build find script
setName -> BuildTabSetFindScript -> AppleScript (scan user.koolTabSet / user.koolTabSetTab)

# parse dump
osascript stdout -> ParseTabSetFindOutput -> []TabSessionRef
```

## Preconditions

- Product exports `BuildTabSetFindScript`, `ParseTabSetFindOutput`, `TabSessionRef`.
- Marker names match create stamps: `user.koolTabSet`, `user.koolTabSetTab`.

## Steps

1. Leaves set Phase to `build-find-script` or `parse-find`.
2. Parse leaves load fixture text into `req.FindOutput`.

## Context

- No live iTerm: script leaves assert substrings; parse leaves use fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Clear create-only fields; leaves set Phase (build-find-script | parse-find).
	req.Tabs = nil
	req.WindowName = ""
	return nil
}
```

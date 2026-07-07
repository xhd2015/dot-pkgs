# Scenario

**Feature**: OnEntry streams each entry in sorted order and aborts on callback error

```
Scan -> OnEntry(entry) per child (sorted) -> abort when callback returns error
```

## Steps

1. Seed `on-entry` profile: `aaa-first`, `mmm-mid`, `zzz-last`.
2. Enable `CollectOnEntry` and `AbortAfterEntry = "mmm-mid"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	home := t.TempDir()
	req.Home = home
	req.SeedProfile = "on-entry"
	seedHome(t, home, req.SeedProfile)
	req.CollectOnEntry = true
	req.AbortAfterEntry = "mmm-mid"
	return nil
}
```
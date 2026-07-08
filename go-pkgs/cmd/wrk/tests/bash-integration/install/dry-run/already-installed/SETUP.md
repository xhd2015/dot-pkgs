# Scenario

**Feature**: install dry-run when fully installed reports already installed

```
pre-seeded bash.sh + dual profile markers
wrk --bash-integration --install --dry-run -> already installed, no changes
```

## Steps

1. Pre-seed bash.sh and both profile markers with unrelated profile content.
2. Run install dry-run.

```go
func Setup(t *testing.T, req *Request) error {
	req.PreExistingBashSh = preInstalledBashShContent()
	req.PreExistingBashProfile = preInstalledProfileContent()
	req.PreExistingBashRC = preInstalledProfileContent()
	return nil
}
```
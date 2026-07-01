# Scenario

**Feature**: domain gate skips or allows the hook based on origin domain

```
# hook with --origin-domain or --exclude-origin-domain -> origin check -> skip or run
hook binary --origin-domain X -> origin URL check -> skip (exit 0) | run (scan go.mod)
```

## Preconditions

- The repository has a configured `origin` remote.

## Steps

1. The leaf case sets the domain filter flag.
2. Run the hook binary with the domain filter applied.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = nil // leaf cases set domain filter flags
	return nil
}

```

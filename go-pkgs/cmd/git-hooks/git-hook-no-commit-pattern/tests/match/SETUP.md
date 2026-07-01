# Scenario

**Feature**: staged files are checked against pattern(s)

```
# staged files -> glob match against patterns -> match? -> print + exit 1 | exit 0
staged files -> pattern match -> matched? -> (yes: print file + exit 1) | (no: exit 0)
```

## Preconditions

- No domain filter is set (hook runs unconditionally).
- The leaf case creates and stages specific files.

## Steps

1. Create and stage files as specified by the leaf case.
2. Run the hook binary with patterns provided by the leaf case.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = nil // leaf cases set args and stage files
	return nil
}
```

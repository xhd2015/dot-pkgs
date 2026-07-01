# Scenario

**Feature**: unknown flag returns an error

```
# hook --unknown -> parse error -> exit non-zero
hook binary --unknown -> "unknown flag: --unknown"
```

## Preconditions

- The hook binary exists.

## Steps

1. Run the command with `--unknown`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--unknown"}
	return nil
}

```

# Scenario

**Feature**: help flag prints usage text

```
# hook --help -> print usage -> exit 0
hook binary --help -> help text
```

## Preconditions

- The hook binary exists.

## Steps

1. Run the command with `--help`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```

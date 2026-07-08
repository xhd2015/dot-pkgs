# Scenario

**Feature**: wrk skill is mutually exclusive with other wrk modes

```
wrk skill <subcommand> + another mode flag -> non-zero, mutually exclusive
```

## Steps

- Descendants combine `skill` subcommands with another wrk mode flag.

```go
func Setup(t *testing.T, req *Request) error {
	ensureSkillHelpersUsed()
	return nil
}
```
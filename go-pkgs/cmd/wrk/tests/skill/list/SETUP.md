# Scenario

**Feature**: wrk skill list prints the single wrk skill name

```
embedded SKILL.md in wrk binary
wrk skill list -> stdout wrk\n
```

## Steps

- Descendants run `wrk skill list`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureSkillHelpersUsed()
	return nil
}
```
## Steps
- Log the rebase mode.

## Steps
- All tests in this mode exercise `mvd --rebase DIR NEW-DIR` to change the root key of a tracked entry.
- The new base directory becomes the first entry in the history chain, with the original entries preserved after it.

```go
func Setup(t *testing.T, req *Request) error {
    t.Logf("mode: rebase")
    return nil
}
```

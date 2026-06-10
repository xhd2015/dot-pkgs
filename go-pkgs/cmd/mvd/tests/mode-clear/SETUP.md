## Steps
- Log the clear mode.

## Steps
- All tests in this mode exercise `mvd --clear SRC` to clear the movement history for a tracked entry.
- After clearing, the history file contains no record of the project.

```go
func Setup(t *testing.T, req *Request) error {
    t.Logf("mode: clear")
    return nil
}
```

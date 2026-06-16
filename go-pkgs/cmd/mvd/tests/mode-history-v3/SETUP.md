## Steps
- Tests in this mode verify history.json v3.0 moves schema (`from`, `to`, `from_type`, `to_type`) and round-trip via `DeriveMoves` / `LocationsFromMoves`.

```go
func Setup(t *testing.T, req *Request) error {
	t.Logf("mode: history-v3")
	return nil
}
```
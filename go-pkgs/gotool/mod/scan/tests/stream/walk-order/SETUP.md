# Scenario

**Feature**: ScanStream emits modules in walk order (unsorted); Scan sorts — the contrast proves stream is unsorted

```
# walk order is lexical DFS; sorted-by-Dir is lexical string compare. The two differ
# because '-' (0x2D) < '/' (0x2F), so a-c < a/b as strings, but DFS visits a/ before a-c.
root + a/b/go.mod (a/ has no go.mod) + a-c/go.mod
  -> ScanStream -> [., a/b, a-c]   (walk order: DFS visits a/b before a-c)
  -> Scan        -> [., a-c, a/b]  (sorted by Dir: a-c < a/b)
```

`ScanStream` must NOT sort — it emits each module as the walker discovers it. `Scan` must
sort by `Dir`. To prove the two differ we need a fixture where **lexical DFS walk order**
diverges from **lexical string sort by Dir** — independent of creation order, since
`filepath.WalkDir` (via `os.ReadDir`) always visits sibling entries in lexical order.

The fixture has a plain intermediate dir `a/` (no `go.mod`) containing `a/b/go.mod`, plus a
sibling `a-c/go.mod`. Walk order (DFS, lexical at each level): root `.`; root's children
sorted by name — `a` is a prefix of `a-c` so `a` < `a-c`, descend into `a/` → emit `a/b`;
then visit `a-c` → emit `a-c`. Result: `[., a/b, a-c]`. Sorted by Dir (lexical string
compare): `.` < `a-c` < `a/b` because at the 2nd character `-` (0x2D) < `/` (0x2F).
Result: `[., a-c, a/b]`. The two orders differ, proving the stream path does not
buffer-and-sort. Robust: no dependence on creation order or filesystem directory listing
order.

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`), git-init'd.
2. Add `a/b/go.mod` (`example.com/root/a/b`) — `a/` is a plain dir with NO `go.mod`.
3. Add `a-c/go.mod` (`example.com/root/a-c`).
4. Set `req.RootDir` (operation `stream` is set by the `stream/` grouping Setup).

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := initStreamWorkspace(t, "example.com/root")

	// a/ is a plain intermediate dir (no go.mod); the module lives at a/b.
	writeModule(t, filepath.Join(ws, "a", "b"), "example.com/root/a/b")
	// a-c is a sibling of a/ at the root level.
	writeModule(t, filepath.Join(ws, "a-c"), "example.com/root/a-c")

	req.RootDir = ws
	return nil
}
```

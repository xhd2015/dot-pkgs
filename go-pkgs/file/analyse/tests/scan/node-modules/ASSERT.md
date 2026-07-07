## Expected

- `nm-entry` children include `node_modules`.
- `nm-entry.Aggregates.NodeModulesDirs == 2`.
- Summary `NodeModulesDirs >= 2`.

## Errors

- `err` is nil.
- Missing child or wrong recursive count.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	nm := findEntry(t, resp.Entries, "nm-entry")

	foundChild := false
	for _, c := range nm.Children {
		if c.Name == "node_modules" {
			foundChild = true
			break
		}
	}
	if !foundChild {
		t.Fatalf("nm-entry children = %v, missing node_modules", childNames(nm))
	}

	if nm.Aggregates.NodeModulesDirs != 2 {
		t.Fatalf("nm-entry NodeModulesDirs = %d, want 2", nm.Aggregates.NodeModulesDirs)
	}

	if resp.Summary.NodeModulesDirs < 2 {
		t.Fatalf("summary.NodeModulesDirs = %d, want >= 2", resp.Summary.NodeModulesDirs)
	}
}
```
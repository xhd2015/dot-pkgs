## Expected

- `err` is nil.
- `resp.IsBinary == false` (truncation-safe UTF-8 prefix of sniff buffer).
- `resp.Desc` is not `"binary file"`.

## Errors

- Incomplete multi-byte sequence only at end of 512-byte sniff window treated as binary.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("DetectFileType(%q): %v", req.Path, err)
	}
	if resp.IsBinary {
		t.Fatalf("DetectFileType(%q) isBinary=true desc=%q, want false for UTF-8 truncated only at sniff boundary",
			req.Path, resp.Desc)
	}
	if resp.Desc == "binary file" {
		t.Fatalf("DetectFileType(%q) desc=%q, want text-like", req.Path, resp.Desc)
	}
}
```

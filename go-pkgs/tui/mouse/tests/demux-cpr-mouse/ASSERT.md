## Expected

- `err` is nil.
- Exactly one CPR event with Row1=41, Col1=1.
- Forward equals the SGR mouse sequence `\x1b[<0;67;25M`.
- Rest is empty (no incomplete hold).

## Errors

- Swallowing mouse bytes, dropping CPR, or holding complete sequences as rest.

```go
import (
	"bytes"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("DemuxCPR: %v", err)
	}
	if len(resp.Rest) != 0 {
		t.Fatalf("rest=%q, want empty", resp.Rest)
	}
	if len(resp.Events) != 1 || resp.Events[0].Row1 != 41 || resp.Events[0].Col1 != 1 {
		t.Fatalf("events=%v, want one CPR {41,1}", resp.Events)
	}
	wantFwd := []byte("\x1b[<0;67;25M")
	if !bytes.Equal(resp.Forward, wantFwd) {
		t.Fatalf("forward=%q, want %q", resp.Forward, wantFwd)
	}
}
```

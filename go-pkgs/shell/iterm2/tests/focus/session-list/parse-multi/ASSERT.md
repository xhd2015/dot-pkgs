## Expected

- Three `SessionRef` values, nil error; order preserved.
- First: WindowID `win-1`, WindowName `Main`, TabIndex `1`, SessionID `sess-aaa`,
  TTY `/dev/ttys148`, Name `Shell`.
- Second: TabIndex `2`, TTY `/dev/ttys149`, Name `App`.
- Third: WindowID `win-2`, WindowName `Other`, TabIndex `1`, TTY `/dev/ttys150`,
  empty Name allowed.

## Exit Code

- N/A (library)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Refs) != 3 {
		t.Fatalf("len(refs)=%d, want 3; refs=%+v", len(resp.Refs), resp.Refs)
	}
	r0 := resp.Refs[0]
	if r0.WindowID != "win-1" || r0.WindowName != "Main" {
		t.Fatalf("refs[0] window = %q/%q, want win-1/Main", r0.WindowID, r0.WindowName)
	}
	if r0.TabIndex != 1 {
		t.Fatalf("refs[0].TabIndex = %d, want 1", r0.TabIndex)
	}
	if r0.SessionID != "sess-aaa" {
		t.Fatalf("refs[0].SessionID = %q, want sess-aaa", r0.SessionID)
	}
	if r0.TTY != "/dev/ttys148" {
		t.Fatalf("refs[0].TTY = %q, want /dev/ttys148", r0.TTY)
	}
	if r0.Name != "Shell" {
		t.Fatalf("refs[0].Name = %q, want Shell", r0.Name)
	}
	r1 := resp.Refs[1]
	if r1.TabIndex != 2 || r1.TTY != "/dev/ttys149" || r1.Name != "App" {
		t.Fatalf("refs[1] = %+v, want TabIndex=2 TTY=/dev/ttys149 Name=App", r1)
	}
	r2 := resp.Refs[2]
	if r2.WindowID != "win-2" || r2.WindowName != "Other" || r2.TabIndex != 1 {
		t.Fatalf("refs[2] = %+v, want win-2/Other tab 1", r2)
	}
	if r2.SessionID != "sess-ccc" || r2.TTY != "/dev/ttys150" {
		t.Fatalf("refs[2] session/tty = %q/%q", r2.SessionID, r2.TTY)
	}
}
```

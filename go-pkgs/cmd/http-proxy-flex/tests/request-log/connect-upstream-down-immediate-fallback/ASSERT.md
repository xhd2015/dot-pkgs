## Expected

When upstream stops listening, CONNECT should fall back to direct access immediately —
not attempt the dead upstream and return 502.

- CONNECT logs "via direct" (not "via upstream proxy")
- No "connect to upstream proxy ... connection refused" error for the request
- CONNECT succeeds with "200 Connection Established"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := resp.Output

	want := "CONNECT " + req.ConnectTarget + " via direct"
	if strings.Contains(output, "CONNECT "+req.ConnectTarget+" via upstream proxy") {
		t.Fatalf("CONNECT should not route via dead upstream proxy, got:\n%s", output)
	}
	if strings.Contains(output, "connect to upstream proxy") && strings.Contains(output, "connection refused") {
		t.Fatalf("CONNECT should fall back to direct when upstream dial fails, got:\n%s", output)
	}
	if !strings.Contains(output, want) {
		t.Fatalf("expected %q after upstream stopped listening, got:\n%s", want, output)
	}
	if !strings.Contains(req.ConnectResponse, "200 Connection Established") {
		t.Fatalf("expected 200 Connection Established, got:\n%s", req.ConnectResponse)
	}
}
```
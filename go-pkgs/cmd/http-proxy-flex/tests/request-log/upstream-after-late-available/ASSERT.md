## Expected

After upstream becomes available, proxied HTTP GET must route via upstream proxy.

- Request log contains "via upstream proxy"
- Request log does not stay on "via direct" only

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := resp.Output

	if strings.Contains(output, "via upstream proxy") {
		return
	}
	t.Fatalf("expected GET via upstream proxy after upstream became available, got:\n%s", output)
}
```
## Expected

- The command exits with code 0 (success).
- Intra-repo replaces are allowed by default.

## Exit Code

- Exit code is 0.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for intra-repo replace in lenient mode, got %d\n%s", resp.ExitCode, resp.Output)
	}
}
```
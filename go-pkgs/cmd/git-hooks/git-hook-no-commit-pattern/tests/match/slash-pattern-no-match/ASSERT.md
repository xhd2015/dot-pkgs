## Expected

- The command exits successfully.
- The output is empty.

## Exit Code

- Exit code is `0`.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for slash-pattern non-match, got %d:\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "" {
		t.Fatalf("expected empty output, got:\n%s", resp.Output)
	}
}
```
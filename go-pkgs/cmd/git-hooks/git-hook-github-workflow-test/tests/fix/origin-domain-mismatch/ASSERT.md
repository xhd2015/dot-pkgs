## Expected

- The command exits successfully.
- The command prints no output because the origin gate does not match.

## Side Effects

- No workflow file is created.

## Exit Code

- Exit code is `0`.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "" {
		t.Fatalf("expected no output for origin-domain mismatch, got:\n%s", resp.Output)
	}
	if resp.WorkflowExists {
		t.Fatalf("origin mismatch must not create workflow at %s", resp.WorkflowPath)
	}
}
```

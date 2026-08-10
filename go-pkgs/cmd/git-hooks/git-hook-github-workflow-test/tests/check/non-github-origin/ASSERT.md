## Expected

- The command exits successfully.
- The command prints no output because non-GitHub origins are skipped in check mode.

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
		t.Fatalf("expected no output for non-GitHub origin, got:\n%s", resp.Output)
	}
	if resp.WorkflowExists {
		t.Fatalf("non-GitHub check mode must not create workflow at %s", resp.WorkflowPath)
	}
}
```

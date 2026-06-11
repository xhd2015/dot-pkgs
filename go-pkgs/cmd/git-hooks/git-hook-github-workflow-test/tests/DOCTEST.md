## How to Run

```sh
doctest test -v ./
```

## Test Tree

- `check/github-missing-workflow`: GitHub origin, no workflow, check mode warns and recommends `--fix`.
- `check/github-existing-workflow`: GitHub origin, existing workflow, check mode exits silently.
- `check/non-github-origin`: Non-GitHub origin, check mode skips silently.
- `check/origin-domain-mismatch`: GitHub origin but `--origin-domain` mismatch, check mode skips silently.
- `fix/github-create-workflow`: GitHub origin, no workflow, `--fix` creates a Go test workflow.
- `fix/github-existing-workflow`: GitHub origin, existing workflow, `--fix` reports that nothing changed.
- `fix/non-github-origin`: Non-GitHub origin, `--fix` errors and does not create a workflow.
- `fix/origin-domain-mismatch`: GitHub origin but `--origin-domain` mismatch, `--fix` skips silently.
- `args/help`: `--help` prints usage.
- `args/unknown-flag`: unknown flags return an error.

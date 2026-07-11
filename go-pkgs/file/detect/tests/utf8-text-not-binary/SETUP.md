# Scenario

**Feature**: UTF-8 / ASCII text must not be classified as binary; NUL must

```
path -> DetectFileType -> isBinary false for text; true when content has NUL
```

## Preconditions

- Every leaf under this branch exercises `DetectFileType` only (no magic-catalog
  expansion beyond existing unit coverage).
- Fixture bytes for the real snapshot are internalized under the leaf `testdata/`.

## Steps

1. Leaf Setup writes or points at the input file and sets `req.Path`.
2. Root Run calls `detect.DetectFileType`.
3. Assert checks `IsBinary` and that text cases are not described as `"binary file"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Isolate branch: clear Path so each leaf must set its own input.
	req.Path = ""
	return nil
}
```

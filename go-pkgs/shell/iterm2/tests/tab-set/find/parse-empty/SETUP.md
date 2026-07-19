# Scenario

**Feature**: empty find output parses to an empty session list

```
"" / whitespace -> ParseTabSetFindOutput -> [] (len 0), nil error
```

## Steps

1. Phase `parse-find`.
2. `FindOutput` is empty string.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "parse-find"
	req.FindOutput = ""
	return nil
}
```

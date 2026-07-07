# Scenario

**Feature**: empty porcelain yields zero counts

```
"" -> ParsePorcelain -> all counts zero
```

## Steps

1. Set `req.Porcelain` to empty string.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Porcelain = ""
	return nil
}
```

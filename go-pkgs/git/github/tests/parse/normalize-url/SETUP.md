# Scenario

**Feature**: NormalizeRepoURL canonicalizes SSH, git, and https URLs

```
# URL normalization
NormalizeRepoURL(o,r,raw) -> https://github.com/o/r
```

## Steps

1. Set `req.NormalizeOwner` to `o`, `req.NormalizeName` to `r`.
2. Set `req.NormalizeInput` from first row of `testdata/cases.tsv` for `Run`.
3. `Assert` validates all rows in the fixture file.

```go
import (
	"bufio"
	"os"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.NormalizeOwner = "o"
	req.NormalizeName = "r"
	f, err := os.Open(fixtureFile(d, "testdata/cases.tsv"))
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return scanner.Err()
	} // header
	if !scanner.Scan() {
		return scanner.Err()
	}
	fields := strings.Split(scanner.Text(), "\t")
	if len(fields) < 3 {
		t.Fatal("invalid cases.tsv")
	}
	req.NormalizeInput = fields[2]
	return scanner.Err()
}```
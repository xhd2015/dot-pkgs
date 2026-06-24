## Expected

- `resp.Normalized` for first case is `https://github.com/o/r`.
- All rows in `testdata/cases.tsv` normalize to `https://github.com/o/r`.

## Side Effects

- None (pure function).

## Errors

- `err` is nil.

## Exit Code

- N/A (library call).

```go
import (
	"bufio"
	"os"
	"strings"
	"testing"

	ghrepos "github.com/xhd2015/dot-pkgs-github/go-pkgs/git/github"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Normalized != "https://github.com/o/r" {
		t.Fatalf("expected first normalized URL https://github.com/o/r, got %q", resp.Normalized)
	}

	f, err := os.Open("testdata/cases.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("missing header")
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "owner") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("invalid row: %q", line)
		}
		owner, name, raw, expected := fields[0], fields[1], fields[2], fields[3]
		got := ghrepos.NormalizeRepoURL(owner, name, raw)
		if got != expected {
			t.Fatalf("NormalizeRepoURL(%q,%q,%q) = %q, want %q", owner, name, raw, got, expected)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}```
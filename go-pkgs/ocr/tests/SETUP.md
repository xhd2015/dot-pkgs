# Scenario

**Feature**: `ocr.RecognizeFile` / cache-compiled Vision helper

```
caller image -> RecognizeFile(opts) -> Result{Text, Engine:vision}
first use -> swiftc -O -> ~/Library/Caches/dot-pkgs/ocr/vision-ocr-<hash>
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/ocr` is importable.
- Leaves are L2 in-process unless labeled `e2e`.
- Parallel-safe: use `t.TempDir()` for cache/helpers; inject `HelperPath` / `Swiftc`.
- No process env / cwd mutation.

## Steps

1. Branch/leaf `Setup` sets `req.Op` and paths.
2. Root `Run` (in DOCTEST.md) calls the public API.
3. Leaf `Assert` checks text, engine, or error substrings.

## Context

- Trailing newline from the helper is trimmed in `Result.Text`.
- `HelperPath` skips compile (unit / most L2 leaves).
- Real Vision e2e uses `/tmp/2026-08-28-10-54-44-clipboard-d9d053dd.png` when present,
  else a copied fixture under the leaf.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = ""
	req.ImagePath = ""
	req.HelperPath = ""
	req.CacheDir = ""
	req.Swiftc = ""
	return nil
}

func writeFakeHelper(t *testing.T, stdout string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-helper")
	script := "#!/bin/sh\nprintf '%s' '" + stdout + "'\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeFakeSwiftc(t *testing.T) (swiftcPath, counterPath string) {
	t.Helper()
	dir := t.TempDir()
	counterPath = filepath.Join(dir, "count")
	swiftcPath = filepath.Join(dir, "swiftc")
	payload := filepath.Join(dir, "payload.sh")
	if err := os.WriteFile(payload, []byte("#!/bin/sh\necho fake-vision\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(counterPath, []byte("0"), 0644); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
set -e
count_file="` + counterPath + `"
n=$(cat "$count_file"); n=$((n+1)); echo "$n" > "$count_file"
out=""
while [ $# -gt 0 ]; do
  if [ "$1" = "-o" ]; then shift; out="$1"; fi
  shift
done
cp "` + payload + `" "$out"
chmod +x "$out"
`
	if err := os.WriteFile(swiftcPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command(swiftcPath, "-O", "-o", filepath.Join(dir, "x"), "y.swift").Run(); err != nil {
		t.Fatalf("fake swiftc smoke: %v", err)
	}
	_ = os.Remove(filepath.Join(dir, "x"))
	if err := os.WriteFile(counterPath, []byte("0"), 0644); err != nil {
		t.Fatal(err)
	}
	return swiftcPath, counterPath
}
```

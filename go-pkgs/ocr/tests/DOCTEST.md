# ocr — macOS Vision OCR (embed Swift → cache compile → recognize)

L2 in-process doctests for `github.com/xhd2015/dot-pkgs/go-pkgs/ocr`.
Default engine is Apple Vision via an embedded Swift helper compiled on first
use into a cache directory. Tests inject `HelperPath` / `Swiftc` so CI does not
require a real Vision compile except the sparse e2e leaf.

## Version

0.0.1

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — CLI (`ocr`, `get-clipboard --ocr`) or library user.
- **RecognizeFile / RecognizeBytes** — public OCR entrypoints.
- **Vision helper** — cached binary built from embedded `vision/RecognizeText.swift`.
- **Cache** — `UserCacheDir()/dot-pkgs/ocr/vision-ocr-<hash8>` (or `Options.CacheDir`).

**Behaviors**

- Empty path / empty bytes → error (`empty image path` / `empty image data`).
- Missing file → wrapped `os.Stat` error.
- With `HelperPath`, run that binary with the image path; stdout text → `Result.Text`
  (trailing newline trimmed); `Engine` is `vision`.
- First `ensureVisionHelper` compiles via `swiftc -O`; second call reuses the binary.

## Decision Tree

```text
ocr/tests/
├── recognize/
│   ├── success-helper/          HelperPath returns fixed text
│   ├── missing-file/            path does not exist
│   └── empty-path/              "" → empty image path
├── helper/
│   └── compile-once/            fake swiftc runs once across two ensures
└── live/                        label: e2e — real Vision on fixture PNG
    └── clipboard-sample/        contains How can I help / hihelo
```

### Parameter significance (high → low)

1. **Outcome** — success vs validation error vs helper compile.
2. **Input surface** — file path vs (e2e) real screenshot fixture.
3. **Engine wiring** — injected helper vs real Vision (e2e only).

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `recognize/success-helper` | `RecognizeFile` with HelperPath → fixed text, engine vision |
| 2 | `recognize/missing-file` | missing path → error containing `no such file` |
| 3 | `recognize/empty-path` | `""` → `empty image path` |
| 4 | `helper/compile-once` | fake swiftc invoked once for two ensure calls |
| 5 | `live/clipboard-sample` | e2e: real Vision on sample PNG (darwin) |

## How to Run

```sh
cd go-pkgs
doctest vet ./ocr/tests
doctest test ./ocr/tests
doctest test --label e2e ./ocr/tests
```

```go
import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/ocr"
)

type Request struct {
	Op string // recognize_file | ensure_twice | live_file

	ImagePath  string
	HelperPath string
	CacheDir   string
	Swiftc     string
}

type Response struct {
	Text    string
	Engine  string
	Text2   string
	SwiftcN string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	ctx := context.Background()
	opts := ocr.Options{
		HelperPath: req.HelperPath,
		CacheDir:   req.CacheDir,
		Swiftc:     req.Swiftc,
	}
	switch req.Op {
	case "recognize_file":
		res, err := ocr.RecognizeFile(ctx, req.ImagePath, opts)
		if err != nil {
			return nil, err
		}
		return &Response{Text: res.Text, Engine: res.Engine}, nil
	case "ensure_twice":
		if runtime.GOOS != "darwin" {
			t.Skip("Vision helper compile requires macOS")
		}
		img := filepath.Join(req.CacheDir, "probe.png")
		if err := os.WriteFile(img, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		res1, err := ocr.RecognizeFile(ctx, img, opts)
		if err != nil {
			return nil, err
		}
		res2, err := ocr.RecognizeFile(ctx, img, opts)
		if err != nil {
			return nil, err
		}
		countPath := filepath.Join(filepath.Dir(req.Swiftc), "count")
		b, err := os.ReadFile(countPath)
		if err != nil {
			t.Fatal(err)
		}
		return &Response{
			Text:    res1.Text,
			Engine:  res1.Engine,
			Text2:   res2.Text,
			SwiftcN: strings.TrimSpace(string(b)),
		}, nil
	case "live_file":
		if runtime.GOOS != "darwin" {
			t.Skip("Vision OCR requires macOS")
		}
		res, err := ocr.RecognizeFile(ctx, req.ImagePath, ocr.Options{})
		if err != nil {
			return nil, err
		}
		return &Response{Text: res.Text, Engine: res.Engine}, nil
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}
```

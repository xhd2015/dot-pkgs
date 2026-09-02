# shell/rg — BurntSushi ripgrep discover / install / ensure / search

## Version

0.0.1

L1 Go tests live beside the package (`*_test.go`). This note indexes significance.

## DSN

| API | Behavior |
| --- | --- |
| `Found` / `Newest` | RG_BIN + well-known + login PATH; `binaryversion.Newest` |
| `FormatUsingNotice` | `using rg VER (path); also found …` |
| `InstallLatest` | GitHub precompiled only; unsupported GOOS/GOARCH → error |
| `Ensure` | missing → install; present → newest; no auto-upgrade |
| `Search` / `SearchStream` | `-i -F --json` + globs; emit each match as parsed; exit 1 → empty |

## Decision tree (covered by Go tests)

```
shell/rg/
├── platform           supported triples / unsupported error / asset URL
├── format-using       also-found clause
├── ensure-noop        newest + notice with also found
├── ensure-missing     unsupported platform after not-found notice
├── install-tarball    extract rg from fake tar.gz via httptest
└── search             json parse + literal CI flag argv
```

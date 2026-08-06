# measure-write-text-limit

Re-run the empirical iTerm **write text** FollowUp length scan used to set
`applescript.WriteTextSafeMaxBytes` (900) and `WriteTextSoftMaxBytes` (1024).

## Requirements

- macOS
- iTerm2 installed
- Opens **many** new windows (ForceNew) — close them when done

## Run

From the `go-pkgs` module root:

```bash
go run ./shell/applescript/tests/scripts/measure-write-text-limit
go run ./shell/applescript/tests/scripts/measure-write-text-limit -out /tmp/my-scan -gap-ms 800
```

## Output

- Stdout progress lines: `pad=… follow=… PASS|EMPTY|MISMATCH`
- `$out/REPORT.txt` — full log + summary
- `$out/*.got` — captured payloads when non-empty

## How to read results

| Status | Meaning |
|--------|---------|
| PASS | File matches payload exactly |
| EMPTY | No file (command lost / never ran) |
| MISMATCH | Truncated or corrupted (often `diff@~1000`) |

**Control phase** must PASS: short `bash script.sh` + multi-KB 中文 body.

If single-line ASCII shows a clean cliff (PASS ≤ ~950 follow bytes, fail ≥ ~1050),
the package constants remain well-founded. Re-tune only with a new REPORT and a
code review of `check.go` constants.

## Related

- `applescript.CheckWriteText` / `DocumentWriteTextLimitation`
- Live doctests: `tests/live/` (`label: e2e`)

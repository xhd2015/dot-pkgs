// measure-write-text-limit re-measures the empirical iTerm write-text FollowUp
// length limit via the same ForceNew path as production local-bot open inject.
//
// Not run by default doctest CI (opens many iTerm windows). Re-run manually:
//
//	cd go-pkgs
//	go run ./shell/applescript/tests/scripts/measure-write-text-limit
//
// Writes REPORT.txt under -out (default /tmp/iterm2-limit-scan-<timestamp>).
//
// Interpreting results:
//
//	PASS     — got file byte-identical to payload
//	EMPTY    — no output file (command never ran / lost)
//	MISMATCH — partial or corrupted payload (often first_diff ~1KB)
//
// Control: short `bash script.sh` + large body must stay PASS if the limit is
// still on write-text command length, not on payload size.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func tryWriteText(follow, gotPath, wantExact string, timeout time.Duration) (status string, gotLen, firstDiff int) {
	_ = os.Remove(gotPath)
	cwd, err := os.UserHomeDir()
	if err != nil {
		return "NO_HOME", 0, -1
	}
	if err := iterm2.OpenConfig(cwd, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{follow},
	}); err != nil {
		return "OPEN_ERR", 0, -1
	}
	deadline := time.Now().Add(timeout)
	var got []byte
	stable, prevLen := 0, -1
	for time.Now().Before(deadline) {
		time.Sleep(150 * time.Millisecond)
		b, err := os.ReadFile(gotPath)
		if err != nil || len(b) == 0 {
			continue
		}
		if len(b) == prevLen {
			stable++
		} else {
			stable = 0
			prevLen = len(b)
		}
		got = b
		if stable >= 3 {
			break
		}
	}
	gotLen = len(got)
	if gotLen == 0 {
		return "EMPTY", 0, -1
	}
	if string(got) == wantExact {
		return "PASS", gotLen, -1
	}
	exp := []byte(wantExact)
	d := -1
	for i := 0; i < len(exp) && i < len(got); i++ {
		if exp[i] != got[i] {
			d = i
			break
		}
	}
	if d < 0 {
		d = len(got)
		if len(exp) < d {
			d = len(exp)
		}
	}
	return "MISMATCH", gotLen, d
}

func followPrintf(payload, gotPath string) string {
	return fmt.Sprintf("printf %%s %s > %s", shellQuote(payload), shellQuote(gotPath))
}

func main() {
	outDir := flag.String("out", "", "output directory (default /tmp/iterm2-limit-scan-<ts>)")
	gapMs := flag.Int("gap-ms", 700, "pause between ForceNew opens")
	flag.Parse()

	dir := *outDir
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "iterm2-limit-scan-"+time.Now().Format("150405"))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var rep strings.Builder
	logf := func(format string, args ...interface{}) {
		line := fmt.Sprintf(format, args...)
		fmt.Println(line)
		rep.WriteString(line + "\n")
	}

	logf("out=%s SafeMax=%d SoftMax=%d", dir, applescript.WriteTextSafeMaxBytes, applescript.WriteTextSoftMaxBytes)
	logf("=== Phase 1: ASCII multi-line (vary pad) ===")
	maxPass, minFail := 0, int(^uint(0)>>1)
	for _, pad := range []int{200, 400, 600, 800, 850, 900, 950, 1000, 1050, 1100, 1200, 1500, 2000, 2500} {
		p := "HDR use <<'EOF'\n" + strings.Repeat("X", pad) + "\n[image] /tmp/MID__seq_1.png\nEND\n"
		gotPath := filepath.Join(dir, fmt.Sprintf("ascii_pad%d.got", pad))
		f := followPrintf(p, gotPath)
		time.Sleep(time.Duration(*gapMs) * time.Millisecond)
		st, gl, d := tryWriteText(f, gotPath, p, 7*time.Second)
		logf("pad=%4d payload=%4d follow=%4d %s got=%4d diff@%d checkOK=%v soft=%v",
			pad, len(p), len(f), st, gl, d,
			applescript.CheckWriteText(f).OK, applescript.CheckWriteText(f).SoftExceeded)
		if st == "PASS" && len(f) > maxPass {
			maxPass = len(f)
		}
		if st != "PASS" && len(f) < minFail {
			minFail = len(f)
		}
	}
	logf("ASCII: maxPassFollow=%d minNonPassFollow=%d", maxPass, minFail)

	logf("\n=== Phase 2: Chinese multi-line ===")
	maxPassZh, minFailZh := 0, int(^uint(0)>>1)
	for _, n := range []int{50, 100, 200, 300, 350, 400, 500, 800} {
		p := "HDR use <<'EOF'\n" + strings.Repeat("字", n) + "\n[image] /tmp/MID__seq_1.png\n尾END\n"
		gotPath := filepath.Join(dir, fmt.Sprintf("zh_%d.got", n))
		f := followPrintf(p, gotPath)
		time.Sleep(time.Duration(*gapMs) * time.Millisecond)
		st, gl, d := tryWriteText(f, gotPath, p, 7*time.Second)
		logf("zh_runes=%4d payload=%4d follow=%4d %s got=%4d diff@%d", n, len(p), len(f), st, gl, d)
		if st == "PASS" && len(f) > maxPassZh {
			maxPassZh = len(f)
		}
		if st != "PASS" && len(f) < minFailZh {
			minFailZh = len(f)
		}
	}
	logf("ZH: maxPassFollow=%d minNonPassFollow=%d", maxPassZh, minFailZh)

	logf("\n=== Phase 3: single-line ASCII ===")
	maxPassSL, minFailSL := 0, int(^uint(0)>>1)
	for _, n := range []int{500, 800, 900, 950, 1000, 1100, 1500, 2000, 2500} {
		p := strings.Repeat("A", n)
		gotPath := filepath.Join(dir, fmt.Sprintf("sl_%d.got", n))
		f := followPrintf(p, gotPath)
		time.Sleep(time.Duration(*gapMs) * time.Millisecond)
		st, gl, d := tryWriteText(f, gotPath, p, 7*time.Second)
		logf("sl_n=%4d payload=%4d follow=%4d %s got=%4d diff@%d", n, len(p), len(f), st, gl, d)
		if st == "PASS" && len(f) > maxPassSL {
			maxPassSL = len(f)
		}
		if st != "PASS" && len(f) < minFailSL {
			minFailSL = len(f)
		}
	}
	logf("SL: maxPassFollow=%d minNonPassFollow=%d", maxPassSL, minFailSL)

	logf("\n=== Phase 4: control short write-text + large body ===")
	big := "HDR use <<'EOF'\n" + strings.Repeat("字", 800) + "\n[image] /tmp/MID__seq_1.png\nEND\n"
	script := filepath.Join(dir, "big.sh")
	gotPath := filepath.Join(dir, "control.got")
	_ = os.Remove(gotPath)
	_ = os.WriteFile(script, []byte("#!/bin/bash\nprintf %s "+shellQuote(big)+" > "+shellQuote(gotPath)+"\n"), 0o755)
	f := "bash " + shellQuote(script)
	time.Sleep(time.Duration(*gapMs) * time.Millisecond)
	st, gl, d := tryWriteText(f, gotPath, big, 8*time.Second)
	logf("control follow=%d payload=%d %s got=%d diff@%d", len(f), len(big), st, gl, d)

	logf("\n=== SUMMARY (compare to package constants SafeMax=%d SoftMax=%d) ===",
		applescript.WriteTextSafeMaxBytes, applescript.WriteTextSoftMaxBytes)
	logf("ASCII multi-line: maxPassFollow=%d minNonPassFollow=%d", maxPass, minFail)
	logf("Chinese multi-line: maxPassFollow=%d minNonPassFollow=%d", maxPassZh, minFailZh)
	logf("ASCII single-line: maxPassFollow=%d minNonPassFollow=%d", maxPassSL, minFailSL)
	logf("Control (short follow): %s", st)
	logf("\n%s", applescript.DocumentWriteTextLimitation())

	report := filepath.Join(dir, "REPORT.txt")
	_ = os.WriteFile(report, []byte(rep.String()), 0o644)
	fmt.Println("wrote", report)
}

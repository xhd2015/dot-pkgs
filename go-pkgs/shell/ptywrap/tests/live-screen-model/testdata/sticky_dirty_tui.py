#!/usr/bin/env python3
"""ANSI fixture TUI for ptywrap live-screen-model doctests.

Paints sticky chrome (prompt + footer) once on the bottom rows, then emits
dirty-region-only frames that rewrite only the top/mid rows. Optionally waits
for a target winsize (resize scenario) and/or emits enough dirty bytes to exceed
ptywrap's scrollback ring (pressure scenario).

Markers (ASCII, CUP-addressable):
  STICKY_PROMPT  — second-to-last row
  STICKY_FOOTER  — last row
  DIRTY_<n>      — top row during dirty updates
  DIRTY_DONE     — written once on the top row when dirty loop finishes
"""

from __future__ import annotations

import argparse
import fcntl
import os
import struct
import sys
import termios
import time


def winsize() -> tuple[int, int]:
    try:
        raw = fcntl.ioctl(sys.stdout.fileno(), termios.TIOCGWINSZ, b"\x00" * 8)
        rows, cols, _, _ = struct.unpack("HHHH", raw)
        cols = cols or 80
        rows = rows or 24
        return max(int(cols), 20), max(int(rows), 5)
    except Exception:
        return 80, 24


def write(s: str) -> None:
    sys.stdout.write(s)
    sys.stdout.flush()


def paint_sticky(cols: int, rows: int, sticky: str, prompt: str) -> None:
    # Hide cursor, clear screen, home.
    write("\x1b[?25l\x1b[2J\x1b[H")
    # Static top content (will be overwritten by dirty frames).
    write("\x1b[1;1HCONTENT_TOP")
    # Sticky chrome on the last two rows — never rewritten by dirty frames.
    prompt_text = ("> " + prompt)[: max(cols, 1)]
    sticky_text = sticky[: max(cols, 1)]
    write(f"\x1b[{rows - 1};1H{prompt_text}")
    write(f"\x1b[{rows};1H{sticky_text}")
    write("\x1b[H")
    sys.stdout.flush()


def dirty_frame(i: int, pad_len: int) -> str:
    # Synchronized update: rewrite top (and optional pad row only). Never touch
    # bottom sticky rows. Use CUP + EL so cells update without full repaint.
    body = f"\x1b[?2026h\x1b[1;1HDIRTY_{i}\x1b[K"
    if pad_len > 0:
        # Pad on row 2 with printable bytes so scrollback fills quickly.
        pad = ("#" * pad_len)[:pad_len]
        body += f"\x1b[2;1H{pad}\x1b[K"
    body += "\x1b[?2026l"
    return body


def mark_done(token: str) -> None:
    path = os.path.join(os.environ.get("TMPDIR", "/tmp"), f"ptywrap-lsm-done-{token}")
    try:
        with open(path, "w", encoding="utf-8") as f:
            f.write("done\n")
    except OSError:
        pass
    write("\x1b[?2026h\x1b[1;1HDIRTY_DONE\x1b[K\x1b[?2026l")


def wait_size(want_cols: int, want_rows: int, timeout_ms: int) -> tuple[int, int]:
    deadline = time.time() + (timeout_ms / 1000.0)
    cols, rows = winsize()
    while time.time() < deadline:
        cols, rows = winsize()
        if cols == want_cols and rows == want_rows:
            return cols, rows
        time.sleep(0.05)
    return cols, rows


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--token", required=True, help="unique token for PID / DONE file")
    ap.add_argument("--sticky", default="STICKY_FOOTER")
    ap.add_argument("--prompt", default="STICKY_PROMPT")
    ap.add_argument("--dirty-iters", type=int, default=30)
    ap.add_argument("--pressure-bytes", type=int, default=0)
    ap.add_argument("--wait-cols", type=int, default=0)
    ap.add_argument("--wait-rows", type=int, default=0)
    ap.add_argument("--wait-timeout-ms", type=int, default=5000)
    ap.add_argument("--pad", type=int, default=200, help="pad chars per dirty frame under pressure")
    args = ap.parse_args()

    if args.wait_cols > 0 and args.wait_rows > 0:
        cols, rows = wait_size(args.wait_cols, args.wait_rows, args.wait_timeout_ms)
    else:
        cols, rows = winsize()

    paint_sticky(cols, rows, args.sticky, args.prompt)

    emitted = 0
    i = 0
    if args.pressure_bytes > 0:
        pad = max(args.pad, 64)
        while emitted < args.pressure_bytes:
            i += 1
            frame = dirty_frame(i, pad)
            write(frame)
            emitted += len(frame.encode("utf-8", errors="replace"))
            # Avoid sleeping every frame under pressure (CI time).
            if i % 50 == 0:
                time.sleep(0.001)
    else:
        n = max(args.dirty_iters, 1)
        for i in range(1, n + 1):
            write(dirty_frame(i, 0))
            time.sleep(0.005)

    mark_done(args.token)
    # Stay alive so multi-snapshot / liveness checks can observe the child.
    time.sleep(3600)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except KeyboardInterrupt:
        raise SystemExit(0)

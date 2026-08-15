//go:build darwin

package chrome

import (
	"context"
	"strings"
	"time"
)

// removeOlderExtensions best-effort removes extension cards named VerifyName
// that do not match the just-loaded VerifyVersion (when known).
//
// Port of swift/ChromeLoadUnpacked.swift: AX cards (name → version → Remove),
// then confirm the in-page Cancel/Remove overlay via the top-right Remove
// button. Never walks System Events "entire contents".
func removeOlderExtensions(ctx context.Context, opts LoadUnpackedOpts) (int, error) {
	keepVer := strings.TrimSpace(opts.VerifyVersion)
	name := strings.TrimSpace(opts.VerifyName)
	if name == "" {
		return 0, nil
	}
	removed := 0
	const maxRounds = 8
	for round := 0; round < maxRounds; round++ {
		if err := ctx.Err(); err != nil {
			return removed, err
		}
		cards := axCollectNamedCards(opts.AppName, name)
		stepf(opts.Stdout, "  step      remove_older         cards=%d", len(cards))
		// Empty keep-version: remove every matching card (including the last).
		// Keep-version set: never delete the sole remaining card.
		if len(cards) == 0 {
			return removed, nil
		}
		if keepVer != "" && len(cards) <= 1 {
			return removed, nil
		}
		var victim *axExtCard
		for i := range cards {
			if keepVer == "" || cards[i].Version != keepVer {
				victim = &cards[i]
				break
			}
		}
		if victim == nil {
			return removed, nil
		}
		axQuartzClick(victim.CX, victim.CY)
		stepf(opts.Stdout, "  step      remove_older         clicked card Remove")
		if !confirmRemoveDialog(ctx, opts) {
			stepf(opts.Stdout, "  step      remove_older         confirm dialog not clicked — stop")
			return removed, nil
		}
		sleepCtx(ctx, 500*time.Millisecond)
		after := axCollectNamedCards(opts.AppName, name)
		stepf(opts.Stdout, "  step      remove_older         cards after=%d", len(after))
		if len(after) >= len(cards) {
			stepf(opts.Stdout, "  step      remove_older         card count did not drop — stop")
			return removed, nil
		}
		removed += len(cards) - len(after)
	}
	return removed, nil
}

func confirmRemoveDialog(ctx context.Context, opts LoadUnpackedOpts) bool {
	sleepCtx(ctx, 450*time.Millisecond)
	if axClickTopRightRemove(opts.AppName) {
		stepf(opts.Stdout, "  step      remove_older         confirmed Remove (top-right)")
		sleepCtx(ctx, 500*time.Millisecond)
		return true
	}
	_, _, err := runOSAscript(ctx, `
tell application "System Events" to keystroke return
`, 5*time.Second)
	if err != nil {
		stepf(opts.Stdout, "  step      remove_older         confirm dialog not clicked — stop")
		return false
	}
	stepf(opts.Stdout, "  step      remove_older         confirmed Remove (return)")
	sleepCtx(ctx, 500*time.Millisecond)
	return true
}

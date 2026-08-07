//go:build !darwin

package chrome

import (
	"context"
	"fmt"
)

func loadUnpacked(ctx context.Context, opts LoadUnpackedOpts) (LoadUnpackedResult, error) {
	_ = ctx
	opts, err := normalizeOpts(opts)
	if err != nil {
		return LoadUnpackedResult{}, err
	}
	stepf(opts.Stderr, "warning: chrome.LoadUnpacked is only supported on macOS (got non-darwin build)")
	return LoadUnpackedResult{}, fmt.Errorf("%w", ErrUnsupported)
}

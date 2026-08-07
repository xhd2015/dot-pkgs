//go:build !darwin

package chrome

import "context"

func removeOlderExtensions(ctx context.Context, opts LoadUnpackedOpts) (int, error) {
	_ = ctx
	_ = opts
	return 0, ErrUnsupported
}

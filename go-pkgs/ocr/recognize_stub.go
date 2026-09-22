//go:build !darwin

package ocr

import (
	"context"
	"fmt"
)

func ensureVisionHelper(ctx context.Context, opts Options) (string, error) {
	_ = ctx
	_ = opts
	return "", fmt.Errorf("Vision OCR requires macOS")
}

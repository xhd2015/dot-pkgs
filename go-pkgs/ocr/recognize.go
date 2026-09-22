package ocr

import (
	"context"
	"fmt"
	"strings"
)

func recognizePath(ctx context.Context, path string, opts Options) (*Result, error) {
	helper, err := resolveHelper(ctx, opts)
	if err != nil {
		return nil, err
	}
	stdout, stderr, err := runCmd(ctx, opts, helper, path)
	if err != nil {
		msg := strings.TrimSpace(string(stderr))
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("vision: %s", msg)
	}
	text := string(stdout)
	text = strings.TrimSuffix(text, "\n")
	if strings.HasSuffix(text, "\r") {
		text = strings.TrimSuffix(text, "\r")
	}
	return &Result{Text: text, Engine: EngineVision}, nil
}

func resolveHelper(ctx context.Context, opts Options) (string, error) {
	if opts.HelperPath != "" {
		return opts.HelperPath, nil
	}
	return ensureVisionHelper(ctx, opts)
}

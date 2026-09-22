package ocr

import (
	"bytes"
	"context"
	"os/exec"
)

func lookPath(opts Options, name string) (string, error) {
	if opts.LookPath != nil {
		return opts.LookPath(name)
	}
	return exec.LookPath(name)
}

func runCmd(ctx context.Context, opts Options, name string, args ...string) (stdout, stderr []byte, err error) {
	if opts.RunCmd != nil {
		return opts.RunCmd(ctx, name, args)
	}
	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

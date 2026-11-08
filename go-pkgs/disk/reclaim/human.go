package reclaim

import "github.com/xhd2015/dot-pkgs/go-pkgs/disk/human"

// humanSize formats n bytes in a short human-readable form.
func humanSize(n int64) string { return human.Size(n) }

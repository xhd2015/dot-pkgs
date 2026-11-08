package reclaim

import "testing"

func TestIsITermBakName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want bool
	}{
		{"iTerm.app.bak-1786114337", true},
		{"iTerm.app.back-1", true},
		{"iTerm2.app.bak-seed", true},
		{"iTerm2.app.back-seed2", true},
		{"ITERM.APP.BAK-X", true},
		{"iTerm.app", false},
		{"iTerm2.app", false},
		{"Something.app.bak-1", false},
		{"iTerm.app.bakup-1", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isITermBakName(tc.name); got != tc.want {
			t.Errorf("isITermBakName(%q) = %v want %v", tc.name, got, tc.want)
		}
	}
}

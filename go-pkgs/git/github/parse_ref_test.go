package github

import "testing"

func TestParseRef(t *testing.T) {
	tests := []struct {
		ref        string
		wantOwner  string
		wantName   string
		wantErr    bool
	}{
		{"xhd2015/os-bar", "xhd2015", "os-bar", false},
		{"https://github.com/xhd2015/os-bar", "xhd2015", "os-bar", false},
		{"git@github.com:xhd2015/os-bar.git", "xhd2015", "os-bar", false},
		{"", "", "", true},
		{"not-a-ref", "", "", true},
	}
	for _, tc := range tests {
		owner, name, err := ParseRef(tc.ref)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("ParseRef(%q) expected error", tc.ref)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseRef(%q): %v", tc.ref, err)
		}
		if owner != tc.wantOwner || name != tc.wantName {
			t.Fatalf("ParseRef(%q) = %q/%q, want %q/%q", tc.ref, owner, name, tc.wantOwner, tc.wantName)
		}
	}
}
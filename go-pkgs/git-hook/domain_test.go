package githook

import "testing"

func TestOriginHost(t *testing.T) {
	tests := map[string]string{
		"https://git.xxx.com/team/repo.git":      "git.xxx.com",
		"ssh://git@git.xxx.com:2222/team/repo":   "git.xxx.com",
		"git@git.xxx.com:team/repo.git":          "git.xxx.com",
		"git.xxx.com:team/repo.git":              "git.xxx.com",
		"/Users/me/src/repo":                     "",
		"https://git.xxx.com:8443/team/repo.git": "git.xxx.com",
	}
	for remote, want := range tests {
		if got := OriginHost(remote); got != want {
			t.Fatalf("OriginHost(%q) = %q, want %q", remote, got, want)
		}
	}
}

func TestDomainFilterNormalize(t *testing.T) {
	filter := DomainFilter{
		OriginDomain:        "https://Git.xxx.com/team/repo",
		ExcludeOriginDomain: "Other.example.com:2222",
	}
	if err := filter.Normalize(); err != nil {
		t.Fatal(err)
	}
	if filter.OriginDomain != "git.xxx.com" {
		t.Fatalf("origin domain = %q", filter.OriginDomain)
	}
	if filter.ExcludeOriginDomain != "other.example.com" {
		t.Fatalf("exclude origin domain = %q", filter.ExcludeOriginDomain)
	}
}
